package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	log "k8s.io/klog/v2"
)

type addNodeRequest struct {
	IPv4     string `json:"ipv4"`
	SMAC     string `json:"smac"`
	DMAC     string `json:"dmac"`
	Eport    int    `json:"eport"`
	IsClient bool   `json:"isClient"`
}

func addPortToBMv2Switch(ifName, thriftPort string, portNum int) error {
	log.Infof("Connecting veth %s to BMv2 switch on thrift port %s", ifName, thriftPort)

	cmd := exec.Command("simple_switch_CLI", "--thrift-port", thriftPort)
	cmd.Stdin = strings.NewReader(fmt.Sprintf("port_add %s %d\n", ifName, portNum))
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf("Error adding port to BMv2 switch: %v. Output: %s", err, output)
		return fmt.Errorf("failed to add port to BMv2 switch: %s", output)
	}
	log.Infof("Port %d added to BMv2 switch successfully", portNum)
	return nil
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
	log.Infof("Sending request to controller: %v", reqBody)
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

	log.Infof("Node with IPv4 %s and MAC %s added to controller on egress port %d", ipv4, dmac, eport)
	return nil
}

func checkBMv2Switch(thriftPort string) error {
	log.Infof("Checking BMv2 switch on thrift port %s", thriftPort)

	cmd := exec.Command("simple_switch_CLI", "--thrift-port", thriftPort)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf("Error connecting to BMv2 switch: %v. Output: %s", err, output)
		return fmt.Errorf("failed to add port to BMv2 switch: %s", output)
	}
	return nil
}

// Helper function to parse the show_ports output and determine the next available port number
func getPortNumberByIfaceName(thriftPort, ifaceName string) (int, error) {
	log.Infof("Getting port number for iface %s on BMv2 switch on thrift port %s", ifaceName, thriftPort)
	cmd := exec.Command("simple_switch_CLI", "--thrift-port", thriftPort)
	cmd.Stdin = strings.NewReader("show_ports")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf("Error getting port number for iface %s: %v. Output: %s", ifaceName, err, output)
		return 0, fmt.Errorf("failed to get port number for iface %s: %s", ifaceName, output)
	}
	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")
	// for i, line := range lines {
	// 	log.Infof("Line %d: %s", i, line)
	// }

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
