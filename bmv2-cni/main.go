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
	IPv4     string `json:"ipv4"`
	SMAC     string `json:"smac"`
	DMAC     string `json:"dmac"`
	Eport    int    `json:"eport"`
	IsClient bool   `json:"isClient"`
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

// func getKubeClient() (*kubernetes.Clientset, error) {
// 	kubeconfig := os.Getenv("KUBECONFIG")
// 	if kubeconfig == "" {
// 		kubeconfig = filepath.Join("/home", "stanley", ".kube", "config")
// 	}
// 	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
// 	if err != nil {
// 		return nil, err
// 	}

// 	clientset, err := kubernetes.NewForConfig(config)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return clientset, nil
// }

// func getPodAnnotations(namespace, podName string) (map[string]string, error) {
// 	clientset, err := getKubeClient()
// 	if err != nil {
// 		return nil, err
// 	}

// 	pod, err := clientset.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
// 	if err != nil {
// 		logger.Printf("Error getting pod %s in namespace %s: %v", podName, namespace, err)
// 		return nil, err
// 	}

// 	return pod.Annotations, nil
// }

// func getCustomCNIArg(namespace, podName string) (string, error) {
// 	annotations, err := getPodAnnotations(namespace, podName)
// 	if err != nil {
// 		return "", err
// 	}

// 	customArg := annotations["bmv2-cni/role"]
// 	return customArg, nil
// }

func loadConf(bytes []byte) (*BMv2NetConf, error) {
	n := &BMv2NetConf{}
	if err := json.Unmarshal(bytes, n); err != nil {
		logger.Printf("Error loading config: %v", err)
		return nil, err
	}
	logger.Println("Config loaded successfully")
	return n, nil
}

func addNodeToController(controllerAddr, ipv4, smac, dmac string, eport int, isClient bool) error {
	url := fmt.Sprintf("http://%s/add_node", controllerAddr)

	reqBody := addNodeRequest{
		IPv4:     ipv4,
		SMAC:     smac,
		DMAC:     dmac,
		Eport:    eport,
		IsClient: isClient,
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

func parseCNIArgs(cniArgs string) (string, string, error) {
	var podNamespace, podName string

	// Split the string by semicolon to get the individual key-value pairs
	argsArray := strings.Split(cniArgs, ";")
	for _, arg := range argsArray {
		kv := strings.Split(arg, "=")
		if len(kv) != 2 {
			continue // Invalid pair, skip
		}

		key := kv[0]
		value := kv[1]

		// Extract pod name and namespace
		switch key {
		case "K8S_POD_NAMESPACE":
			podNamespace = value
		case "K8S_POD_NAME":
			podName = value
		}
	}

	// Check if both values are extracted successfully
	if podNamespace == "" || podName == "" {
		return "", "", fmt.Errorf("failed to extract pod namespace or pod name from CNI_ARGS")
	}

	return podNamespace, podName, nil
}

func cmdAdd(args *skel.CmdArgs) error {
	logger.Printf("Args in CNI Add: %v", args)

	// Parse the CNI Args to get pod namespace and name
	podNamespace, podName, err := parseCNIArgs(args.Args)
	if err != nil {
		logger.Printf("Error parsing CNI Args: %v", err)
		return err
	}

	// Fetch custom CNI argument from pod annotations
	// customArg, err := getCustomCNIArg(podNamespace, podName)
	// if err != nil {
	// 	logger.Printf("Error retrieving custom CNI argument: %v\n", err)
	// 	os.Exit(1)
	// }
	// logger.Printf("Custom CNI Argument: %s\n", customArg)

	// TODO - Implement logic to determine if the pod is a server or client,
	// based on annotations or other metadata
	isClient := false
	isServer := false
	if strings.Contains(podName, "client") {
		isClient = true
	}
	if strings.Contains(podName, "server") {
		isServer = true
	}

	logger.Printf("Parsed Namespace: %s, Pod Name: %s, isClient: %t", podNamespace, podName, isClient)

	// Load the CNI config
	conf, err := loadConf(args.StdinData)
	if err != nil {
		logger.Printf("Error loading CNI config: %v", err)
		return err
	}

	// Now proceed with the rest of your networking setup, such as BMv2 switch connection, IPAM, etc.
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

	eport := 0
	if !isClient {
		// Generate a random egress port number for the server
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		eport = rng.Intn(255) + 1
	}

	if err := addPortToBMv2Switch(hostInterface.Name, conf.ThriftPort, eport); err != nil {
		logger.Printf("Error connecting veth to BMv2 switch: %v", err)
		return err
	}
	logger.Printf("Veth %s connected to BMv2 switch on port %d", hostInterface.Name, eport)

	// Only send node information to the controller if the node is a client or server
	// Otherwise the switch will load-balance traffic to kubernetes system nodes like CoreDNS, etc.
	if isClient || isServer {
		// Send the node information to the controller
		ipv4 := ipConfig.Address.IP.String()
		dmac := containerInterface.HardwareAddr.String()
		smac := hostInterface.HardwareAddr.String()

		err = addNodeToController(conf.ControllerAddr, ipv4, smac, dmac, eport, isClient)
		if err != nil {
			logger.Printf("Error adding node to controller: %v", err)
			return err
		}
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
