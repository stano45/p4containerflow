package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/containernetworking/plugins/pkg/ip"
	"github.com/containernetworking/plugins/pkg/ipam"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/vishvananda/netlink"
)

type BMv2NetConf struct {
	types.NetConf
	ThriftPort     string      `json:"thriftPort"`
	LogFile        string      `json:"logFile"`
	ControllerAddr string      `json:"controllerAddr"`
	IPAM           *types.IPAM `json:"ipam"`
}

type addNodeRequest struct {
	IPv4  string `json:"ipv4"`
	SMAC  string `json:"smac"`
	DMAC  string `json:"dmac"`
	Eport int    `json:"eport"`
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

func addNodeToController(controllerAddr, ipv4, smac, dmac string, eport int) error {
	url := fmt.Sprintf("http://%s/add_node", controllerAddr)

	reqBody := addNodeRequest{
		IPv4:  ipv4,
		SMAC:  smac,
		DMAC:  dmac,
		Eport: eport,
	}
	logger.Printf("Sending request to controller: %v", reqBody)
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send HTTP request to controller: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("controller returned non-OK status: %v, failed to read response body: %v", resp.Status, err)
		}
		bodyString := string(bodyBytes)
		return fmt.Errorf("controller returned non-OK status: %v, response: %v", resp.Status, bodyString)
	}

	logger.Printf("Node with IPv4 %s and MAC %s added to controller on egress port %d", ipv4, dmac, eport)
	return nil
}

func cmdAdd(args *skel.CmdArgs) error {
	logger.Printf("Args in CNI Add: %v", args)
	conf, err := loadConf(args.StdinData)
	if err != nil {
		logger.Printf("Error loading CNI config: %v", err)
		return err
	}

	if err := checkBMv2Switch(conf.ThriftPort); err != nil {
		logger.Printf("Error checking BMv2 switch: %v", err)
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

	hostVethName := getHostVethName(args.ContainerID)

	var hostInterface, containerInterface net.Interface

	err = containerNs.Do(func(_ ns.NetNS) error {
		hostInterface, containerInterface, err = ip.SetupVethWithName(args.IfName, hostVethName, 1500, "", hostNs)
		if err != nil {
			logger.Printf("Error creating veth pair: %v", err)
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	// IPAM: Allocate an IP address using the IPAM plugin
	r, err := ipam.ExecAdd(conf.IPAM.Type, args.StdinData)
	if err != nil {
		logger.Printf("Error from IPAM: %v", err)
		return err
	}

	result, err := current.NewResultFromResult(r)
	if err != nil {
		logger.Printf("Error converting IPAM result: %v", err)
		return err
	}

	if len(result.IPs) == 0 {
		return fmt.Errorf("IPAM plugin returned no IPs")
	}

	ipConfig := result.IPs[0]

	err = containerNs.Do(func(_ ns.NetNS) error {
		// Apply the IP address to the container's interface
		link, err := netlink.LinkByName(containerInterface.Name)
		if err != nil {
			return fmt.Errorf("failed to lookup %q: %v", containerInterface.Name, err)
		}

		addr := &netlink.Addr{IPNet: &ipConfig.Address, Label: ""}
		logger.Printf("Adding IP address %s to %q", addr, containerInterface.Name)
		if err := netlink.AddrAdd(link, addr); err != nil {
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

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomPortNum := rng.Intn(255) + 1

	if err := addPortToBMv2Switch(hostInterface.Name, conf.ThriftPort, randomPortNum); err != nil {
		logger.Printf("Error connecting veth to BMv2 switch: %v", err)
		return err
	}
	logger.Printf("Veth %s connected to BMv2 switch on port %d", hostInterface.Name, randomPortNum)

	// Send the node information to the controller
	ipv4 := ipConfig.Address.IP.String()
	dmac := containerInterface.HardwareAddr.String()
	smac := hostInterface.HardwareAddr.String()

	err = addNodeToController(conf.ControllerAddr, ipv4, smac, dmac, randomPortNum)
	if err != nil {
		logger.Printf("Error adding node to controller: %v", err)
		return err
	}

	// Add IPs to the result
	result.Interfaces = []*current.Interface{{
		Name:    hostInterface.Name,
		Sandbox: args.Netns,
	}}
	result.IPs = []*current.IPConfig{ipConfig}

	logger.Println("CNI Add operation completed successfully")
	return types.PrintResult(result, conf.CNIVersion)
}

func cmdCheck(args *skel.CmdArgs) error {
	// TODO: implement check
	logger.Println("CNI Check operation called")
	return nil
}

func cmdDel(args *skel.CmdArgs) error {
	logger.Printf("Args in CNI Del: %v", args)
	conf, err := loadConf(args.StdinData)
	if err != nil {
		logger.Printf("Error loading CNI config for deletion: %v", err)
		return err
	}

	if err := checkBMv2Switch(conf.ThriftPort); err != nil {
		logger.Printf("Error checking BMv2 switch: %v", err)
		return err
	}

	if err := ipam.ExecDel(conf.IPAM.Type, args.StdinData); err != nil {
		logger.Printf("Error from IPAM on DEL: %v", err)
		return err
	}

	hostVethName := getHostVethName(args.ContainerID)
	port, err := getPortNumberByIfaceName(conf.ThriftPort, hostVethName)
	if err != nil {
		logger.Printf("Error getting port number for iface %s: %v", hostVethName, err)
		return err
	}
	portStr := strconv.Itoa(port)

	cmd := exec.Command("simple_switch_CLI", "--thrift-port", conf.ThriftPort)
	cmd.Stdin = strings.NewReader(fmt.Sprintf("port_remove %s\n", portStr))
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Printf("error detaching veth from BMv2 switch: %v, output: %s", err, string(output))
		return fmt.Errorf("error detaching veth from BMv2 switch: %v, output: %s", err, string(output))
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

func addPortToBMv2Switch(ifName, thriftPort string, portNum int) error {
	logger.Printf("Connecting veth %s to BMv2 switch on thrift port %s", ifName, thriftPort)

	cmd := exec.Command("simple_switch_CLI", "--thrift-port", thriftPort)
	cmd.Stdin = strings.NewReader(fmt.Sprintf("port_add %s %d\n", ifName, portNum))
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Printf("Error adding port to BMv2 switch: %v. Output: %s", err, output)
		return fmt.Errorf("failed to add port to BMv2 switch: %s", output)
	}
	logger.Printf("Port %d added to BMv2 switch successfully", portNum)
	return nil
}

func checkBMv2Switch(thriftPort string) error {
	logger.Printf("Checking BMv2 switch on thrift port %s", thriftPort)

	cmd := exec.Command("simple_switch_CLI", "--thrift-port", thriftPort)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Printf("Error connecting to BMv2 switch: %v. Output: %s", err, output)
		return fmt.Errorf("failed to add port to BMv2 switch: %s", output)
	}
	return nil
}

// Helper function to parse the show_ports output and determine the next available port number
func getPortNumberByIfaceName(thriftPort, ifaceName string) (int, error) {
	logger.Printf("Getting port number for iface %s on BMv2 switch on thrift port %s", ifaceName, thriftPort)
	cmd := exec.Command("simple_switch_CLI", "--thrift-port", thriftPort)
	cmd.Stdin = strings.NewReader("show_ports")
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Printf("Error getting port number for iface %s: %v. Output: %s", ifaceName, err, output)
		return 0, fmt.Errorf("failed to get port number for iface %s: %s", ifaceName, output)
	}
	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")
	for i, line := range lines {
		logger.Printf("Line %d: %s", i, line)
	}

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[1] == ifaceName {
			portNum, err := strconv.Atoi(fields[0])
			if err == nil {
				return portNum, nil
			}
			return 0, fmt.Errorf("failed to parse port number for iface %s", ifaceName)
		}
	}

	return 0, fmt.Errorf("iface %s not found in output", ifaceName)
}

func getHostVethName(containerID string) string {
	return fmt.Sprintf("veth%s", containerID[:5])
}
