package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/containernetworking/plugins/pkg/ip"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/vishvananda/netlink"
)

type BMv2NetConf struct {
	types.NetConf
	ThriftPort string `json:"thriftPort"`
	LogFile    string `json:"logFile"`
}

var (
	logger = log.New(os.Stdout, "", log.LstdFlags)
)

func main() {
	logFile, err := os.OpenFile("/var/log/bmv2-cni.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}
	defer logFile.Close()
	logger.SetOutput(logFile)

	cniFuncs := skel.CNIFuncs{
		Add:   cmdAdd,
		Check: cmdCheck,
		Del:   cmdDel,
	}
	skel.PluginMainFuncs(cniFuncs, version.All, "BMv2 CNI Plugin v0.1")
}

func loadConf(bytes []byte) (*BMv2NetConf, error) {
	n := &BMv2NetConf{}
	if err := json.Unmarshal(bytes, n); err != nil {
		logger.Printf("Error loading config: %v", err)
		return nil, err
	}
	logger.Println("Config loaded successfully")
	return n, nil
}

func cmdAdd(args *skel.CmdArgs) error {
	conf, err := loadConf(args.StdinData)
	if err != nil {
		logger.Printf("Error loading CNI config: %v", err)
		return err
	}

	hostNs, err := ns.GetCurrentNS()
	if err != nil {
		return fmt.Errorf("failed to get current netns: %v", err)
	}
	defer hostNs.Close()
	logger.Printf("Host netns: %s", hostNs.Path())

	containerNs, err := ns.GetNS(args.Netns)
	if err != nil {
		return fmt.Errorf("failed to open netns %q: %v", args.Netns, err)
	}
	defer containerNs.Close()
	logger.Printf("Container netns: %s", containerNs.Path())

	var hostInterface, containerInterface net.Interface
	// Configure the container interface
	err = containerNs.Do(func(_ ns.NetNS) error {
		// Create the veth pair
		hostInterface, containerInterface, err = ip.SetupVeth(args.IfName, 1500, "", hostNs)
		if err != nil {
			logger.Printf("Error creating veth pair: %v", err)
			return err
		}

		// Allocate an IP address (you might want to use a proper IPAM plugin here)
		ipv4Addr, ipv4Net, err := net.ParseCIDR("10.0.0.1/24")
		if err != nil {
			logger.Printf("Error parsing CIDR: %v", err)
			return err
		}
		link, err := netlink.LinkByName(containerInterface.Name)
		if err != nil {
			return fmt.Errorf("failed to lookup %q: %v", containerInterface.Name, err)
		}

		if err := netlink.AddrAdd(link, &netlink.Addr{IPNet: &net.IPNet{IP: ipv4Addr, Mask: ipv4Net.Mask}}); err != nil {
			return fmt.Errorf("failed to add IP addr to %q: %v", containerInterface.Name, err)
		}

		if err := netlink.LinkSetUp(link); err != nil {
			return fmt.Errorf("failed to set %q UP: %v", containerInterface.Name, err)
		}

		return nil
	})
	if err != nil {
		return err
	}
	// Connect the host veth to the BMv2 switch
	if err := connectToBMv2Switch(hostInterface.Name, conf.ThriftPort); err != nil {
		logger.Printf("Error connecting veth to BMv2 switch: %v", err)
		return err
	}

	// Prepare the result
	result := &current.Result{
		CNIVersion: conf.CNIVersion,
		Interfaces: []*current.Interface{{
			Name:    args.IfName,
			Sandbox: args.Netns,
		}},
		// IPs: []*current.IPConfig{{
		// 	Address: net.IPNet{IP: ipv4Addr, Mask: ipv4Net.Mask},
		// }},
	}

	logger.Println("CNI Add operation completed successfully")
	return types.PrintResult(result, conf.CNIVersion)
}

func cmdCheck(args *skel.CmdArgs) error {
	// Implement status checks here
	logger.Println("CNI Check operation called")
	return nil
}

func cmdDel(args *skel.CmdArgs) error {
	conf, err := loadConf(args.StdinData)
	if err != nil {
		logger.Printf("Error loading CNI config for deletion: %v", err)
		return err
	}

	hostVethName := fmt.Sprintf("cni-%s", args.ContainerID[:8])

	if err = exec.Command("simple_switch_CLI", "--thrift-port", conf.ThriftPort, "port_remove", "0").Run(); err != nil {
		logger.Printf("Error detaching veth from BMv2 switch: %v", err)
		return fmt.Errorf("failed to detach veth from BMv2 switch: %v", err)
	}

	link, err := netlink.LinkByName(hostVethName)
	if err == nil {
		if err := netlink.LinkDel(link); err != nil {
			logger.Printf("Error deleting link %q: %v", hostVethName, err)
			return fmt.Errorf("failed to delete link %q: %v", hostVethName, err)
		}
	}

	logger.Println("CNI Del operation completed successfully")
	return nil
}

func connectToBMv2Switch(ifName, thriftPort string) error {
	logger.Printf("Connecting veth %s to BMv2 switch on thrift port %s", ifName, thriftPort)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomPortNum := rng.Intn(65535) + 1 

	cmd := exec.Command("simple_switch_CLI", "--thrift-port", thriftPort)
	cmd.Stdin = strings.NewReader(fmt.Sprintf("port_add %s %d\n", ifName, randomPortNum))
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Printf("Error adding port to BMv2 switch: %v. Output: %s", err, output)
		return fmt.Errorf("failed to add port to BMv2 switch: %s", output)
	}
	logger.Printf("Port %d added to BMv2 switch successfully", randomPortNum)
	return nil
}

// Helper function to parse the show_ports output and determine the next available port number
// func getNextPortNumber(output string) (int, error) {
// 	lines := strings.Split(output, "\n")
// 	for i, line := range lines {
// 		logger.Printf("Line %d: %s", i, line)
// 	}
// 	numLines := len(lines)

// 	if numLines < 3 {
// 		// If there are fewer than 2 lines, default to port 0
// 		return 0, nil
// 	}

// 	// Get the second-to-last line
// 	secondToLastLine := lines[numLines-3]
// 	logger.Printf("Second-to-last line: %s", secondToLastLine)
// 	fields := strings.Fields(secondToLastLine)

// 	// Attempt to parse the first field as the port number
// 	if len(fields) > 0 {
// 		portNum, err := strconv.Atoi(fields[0])
// 		if err == nil {
// 			return portNum + 1, nil // Return the next available port number
// 		}
// 	}

// 	// Default to port 0 if parsing fails
// 	return 0, nil
// }
