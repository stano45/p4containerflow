package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
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
	log "k8s.io/klog/v2"
)

type BMv2NetConf struct {
	types.NetConf
	ThriftPort     string      `json:"thriftPort"`
	LogFile        string      `json:"logFile"`
	ControllerAddr string      `json:"controllerAddr"`
	IPAM           *types.IPAM `json:"ipam"`
}

func init() {
	log.InitFlags(nil)
	err := flag.Set("logtostderr", "false")
	if err != nil {
		log.Error("Can't reset the logtostderr flag:", err)
		os.Exit(1)
	}
	err = flag.Set("log_file", "/var/log/bmv2-cni.log")
	if err != nil {
		log.Error("Can't set the log file:", err)
		os.Exit(1)
	}
}

func main() {
	defer log.Flush()
	log.Infof("Starting BMv2 CNI Plugin")
	cniFuncs := skel.CNIFuncs{
		Add:   cmdAdd,
		Check: cmdCheck,
		Del:   cmdDel,
	}
	skel.PluginMainFuncs(cniFuncs, version.All, "BMv2 CNI Plugin v0.1")
}

func cmdAdd(args *skel.CmdArgs) error {
	defer log.Flush()
	log.Infof("CNI Add operation called with args: %v", args)

	podNamespace, podName, err := parseArgs(args.Args)
	if err != nil {
		log.Errorf("Error parsing CNI Args: %v", err)
		return err
	}

	isClient, isServer := isClientOrServer(podName)
	log.Infof("Parsed Namespace: %s, Pod Name: %s, isClient: %t, isServer: %t", podNamespace, podName, isClient, isServer)

	conf, err := parseNetConf(args.StdinData)
	if err != nil {
		log.Errorf("Error loading CNI config: %v", err)
		return err
	}
	log.Infof("Parsed NetConf: %+v", conf)

	// TODO: do this dynamically or via config
	defaultGateway := net.ParseIP("10.1.1.1")
	if defaultGateway == nil {
		log.Errorf("Failed to parse default gateway IP")
		return fmt.Errorf("failed to parse default gateway IP")
	}
	log.Infof("Using default gateway: %s", defaultGateway.String())

	if err := checkBMv2Switch(conf.ThriftPort); err != nil {
		log.Errorf("Error checking BMv2 switch: %v", err)
		return err
	}
	log.Infof("BMv2 switch is reachable on Thrift port %s", conf.ThriftPort)

	hostNs, err := ns.GetCurrentNS()
	if err != nil {
		log.Errorf("Failed to get current netns: %v", err)
		return fmt.Errorf("failed to get current netns: %v", err)
	}
	defer hostNs.Close()
	log.Infof("Host netns: %s", hostNs.Path())

	containerNs, err := ns.GetNS(args.Netns)
	if err != nil {
		log.Errorf("Failed to open netns %q: %v", args.Netns, err)
		return fmt.Errorf("failed to open netns %q: %v", args.Netns, err)
	}
	defer containerNs.Close()
	log.Infof("Container netns: %s", containerNs.Path())

	var hostInterface, containerInterface net.Interface
	var result *current.Result
	var ipConfig *current.IPConfig

	err = containerNs.Do(func(_ ns.NetNS) error {
		// Create a veth pair and move one end to the host namespace
		hostVethName := buildHostVethName(args.ContainerID)
		hostInterface, containerInterface, err = ip.SetupVethWithName(args.IfName, hostVethName, 1500, "", hostNs)
		if err != nil {
			log.Errorf("Error creating veth pair: %v", err)
			return err
		}
		log.Infof("Created veth pair: container interface %s, host interface %s", containerInterface.Name, hostInterface.Name)

		// IPAM: Allocate an IP address using the IPAM plugin
		r, err := ipam.ExecAdd(conf.IPAM.Type, args.StdinData)
		if err != nil {
			log.Errorf("Error from IPAM: %v", err)
			return err
		}
		log.Infof("IPAM ExecAdd returned result: %v", r)

		result, err = current.NewResultFromResult(r)
		if err != nil {
			log.Errorf("Error converting IPAM result: %v", err)
			return err
		}

		if len(result.IPs) == 0 {
			log.Errorf("IPAM plugin returned no IPs")
			return fmt.Errorf("IPAM plugin returned no IPs")
		}
		ipConfig = result.IPs[0]
		log.Infof("Allocated IP: %s", ipConfig.Address.String())

		if err := deleteAllRoutes(); err != nil {
			log.Errorf("Error deleting all routes: %v", err)
			return err
		}
		log.Infof("Deleted all existing routes in the container namespace")

		containerLink, err := netlink.LinkByName(containerInterface.Name)
		if err != nil {
			log.Errorf("Failed to lookup %q: %v", containerInterface.Name, err)
			return fmt.Errorf("failed to lookup %q: %v", containerInterface.Name, err)
		}

		addr := &netlink.Addr{IPNet: &ipConfig.Address, Label: ""}
		log.Infof("Adding IP address %s to %q", addr, containerInterface.Name)

		if err := netlink.AddrAdd(containerLink, addr); err != nil {
			log.Errorf("Failed to add IP addr to %q: %v", containerInterface.Name, err)
			return fmt.Errorf("failed to add IP addr to %q: %v", containerInterface.Name, err)
		}

		if err := netlink.LinkSetUp(containerLink); err != nil {
			log.Errorf("Failed to set %q UP: %v", containerInterface.Name, err)
			return fmt.Errorf("failed to set %q UP: %v", containerInterface.Name, err)
		}
		log.Infof("Set interface %q up", containerInterface.Name)

		staticRoute := &netlink.Route{
			LinkIndex: containerLink.Attrs().Index,
			Dst:       &net.IPNet{IP: net.IPv4(0, 0, 0, 0), Mask: net.CIDRMask(0, 32)},
			Gw:        defaultGateway,
		}

		if err := netlink.RouteAdd(staticRoute); err != nil {
			log.Errorf("Failed to add default route via %s: %v", defaultGateway, err)
			return fmt.Errorf("failed to add default route via %s: %v", defaultGateway, err)
		}
		log.Infof("Added default route via %s", defaultGateway)
		return nil
	})
	if err != nil {
		log.Errorf("Error configuring container namespace: %v", err)
		return err
	}

	// Add static ARP entry for the gateway on the host interface
	if err := addStaticARPEntry(hostInterface.Name, defaultGateway, hostInterface.HardwareAddr); err != nil {
		log.Errorf("Error adding static ARP entry: %v", err)
		return err
	}
	log.Infof("Added static ARP entry for gateway %s on interface %s", defaultGateway.String(), hostInterface.Name)

	eport := 1
	if !isClient {
		// Generate a random egress port number for the server
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		eport = rng.Intn(255) + 1
	}
	log.Infof("Selected egress port %d for BMv2 switch", eport)

	if err := addPortToBMv2Switch(hostInterface.Name, conf.ThriftPort, eport); err != nil {
		log.Errorf("Error connecting veth to BMv2 switch: %v", err)
		return err
	}
	log.Infof("Veth %s connected to BMv2 switch on port %d", hostInterface.Name, eport)

	if isClient || isServer {
		ipv4 := ipConfig.Address.IP.String()
		dmac := containerInterface.HardwareAddr.String()
		smac := hostInterface.HardwareAddr.String()

		err = addNodeToController(conf.ControllerAddr, ipv4, smac, dmac, eport, isClient)
		if err != nil {
			log.Errorf("Error adding node to controller: %v", err)
			return err
		}
		log.Infof("Added node to controller: IPv4 %s, SMAC %s, DMAC %s, eport %d, isClient %t", ipv4, smac, dmac, eport, isClient)
	}

	result.Interfaces = []*current.Interface{{
		Name:    hostInterface.Name,
		Sandbox: args.Netns,
	}}
	result.IPs = []*current.IPConfig{ipConfig}

	log.Infof("CNI Add operation completed successfully")
	return types.PrintResult(result, conf.CNIVersion)
}

func cmdCheck(args *skel.CmdArgs) error {
	defer log.Flush()
	log.Warningf("CNI Check operation called with args: %v", args)
	// TODO: implement check
	log.Warningf("CNI Check operation not implemented")
	return nil
}

func cmdDel(args *skel.CmdArgs) error {
	defer log.Flush()
	log.Infof("CNI Del operation called with args: %v", args)

	conf, err := parseNetConf(args.StdinData)
	if err != nil {
		log.Errorf("Error loading CNI config for deletion: %v", err)
		return err
	}
	log.Infof("Parsed NetConf: %+v", conf)

	if err := checkBMv2Switch(conf.ThriftPort); err != nil {
		log.Errorf("Error checking BMv2 switch: %v", err)
		return err
	}
	log.Infof("BMv2 switch is reachable on Thrift port %s", conf.ThriftPort)

	if err := ipam.ExecDel(conf.IPAM.Type, args.StdinData); err != nil {
		log.Errorf("Error from IPAM on DEL: %v", err)
		return err
	}
	log.Infof("IPAM ExecDel successful")

	hostVethName := buildHostVethName(args.ContainerID)
	log.Infof("Deleting host veth interface %s", hostVethName)

	port, err := getPortNumberByIfaceName(conf.ThriftPort, hostVethName)
	if err != nil {
		log.Errorf("Error getting port number for iface %s: %v", hostVethName, err)
		return err
	}
	portStr := strconv.Itoa(port)
	log.Infof("Port number for interface %s is %s", hostVethName, portStr)

	cmd := exec.Command("simple_switch_CLI", "--thrift-port", conf.ThriftPort)
	cmd.Stdin = strings.NewReader(fmt.Sprintf("port_remove %s\n", portStr))
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf("Error detaching veth from BMv2 switch: %v, output: %s", err, string(output))
		return fmt.Errorf("error detaching veth from BMv2 switch: %v, output: %s", err, string(output))
	}
	log.Infof("Detached veth %s from BMv2 switch port %s", hostVethName, portStr)

	link, err := netlink.LinkByName(hostVethName)
	if err == nil {
		if err := netlink.LinkDel(link); err != nil {
			log.Errorf("Error deleting link %q: %v", hostVethName, err)
			return fmt.Errorf("failed to delete link %q: %v", hostVethName, err)
		}
		log.Infof("Deleted link %q", hostVethName)
	} else {
		log.Warningf("Link %q not found, might have been already deleted", hostVethName)
	}

	log.Infof("CNI Del operation completed successfully")
	return nil
}
