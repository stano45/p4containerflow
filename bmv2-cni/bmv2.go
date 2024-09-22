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
	log.Infof("Connecting veth %s to BMv2 switch on Thrift port %s", ifName, thriftPort)

	cmd := exec.Command("simple_switch_CLI", "--thrift-port", thriftPort)
	cmd.Stdin = strings.NewReader(fmt.Sprintf("port_add %s %d\n", ifName, portNum))
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf("Error adding port to BMv2 switch: %v. Output: %s", err, string(output))
		return fmt.Errorf("failed to add port to BMv2 switch: %s", string(output))
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
	log.Infof("Preparing to send request to controller at %s", url)
	log.Infof("Request body: %+v", reqBody)

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		log.Errorf("Failed to marshal request body: %v", err)
		return fmt.Errorf("failed to marshal request body: %v", err)
	}
	log.Infof("Marshalled JSON data: %s", string(jsonData))

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Errorf("Failed to send HTTP request to controller: %v", err)
		return fmt.Errorf("failed to send HTTP request to controller: %v", err)
	}
	defer resp.Body.Close()
	log.Infof("Received response from controller: %s", resp.Status)

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Errorf("Controller returned non-OK status: %v, failed to read response body: %v", resp.Status, err)
			return fmt.Errorf("controller returned non-OK status: %v, failed to read response body: %v", resp.Status, err)
		}
		bodyString := string(bodyBytes)
		log.Errorf("Controller returned non-OK status: %v, response: %v", resp.Status, bodyString)
		return fmt.Errorf("controller returned non-OK status: %v, response: %v", resp.Status, bodyString)
	}

	log.Infof("Node with IPv4 %s and MAC %s added to controller on egress port %d", ipv4, dmac, eport)
	return nil
}

func checkBMv2Switch(thriftPort string) error {
	log.Infof("Checking BMv2 switch connectivity on Thrift port %s", thriftPort)

	cmd := exec.Command("simple_switch_CLI", "--thrift-port", thriftPort)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf("Error connecting to BMv2 switch: %v. Output: %s", err, string(output))
		return fmt.Errorf("failed to connect to BMv2 switch: %s", string(output))
	}
	log.Infof("Successfully connected to BMv2 switch on Thrift port %s", thriftPort)
	return nil
}

// Helper function to parse the show_ports output and determine the port number by interface name
func getPortNumberByIfaceName(thriftPort, ifaceName string) (int, error) {
	log.Infof("Getting port number for interface %s on BMv2 switch (Thrift port %s)", ifaceName, thriftPort)

	cmd := exec.Command("simple_switch_CLI", "--thrift-port", thriftPort)
	cmd.Stdin = strings.NewReader("show_ports\n")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf("Error executing 'show_ports' on BMv2 switch: %v. Output: %s", err, string(output))
		return 0, fmt.Errorf("failed to get port number for interface %s: %s", ifaceName, string(output))
	}
	log.Infof("Output from 'show_ports':\n%s", string(output))

	lines := strings.Split(string(output), "\n")
	for i, line := range lines {
		log.Infof("Parsing line %d: %s", i+1, line)
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[1] == ifaceName {
			portNum, err := strconv.Atoi(fields[0])
			if err == nil {
				log.Infof("Found port number %d for interface %s", portNum, ifaceName)
				return portNum, nil
			}
			log.Errorf("Failed to parse port number from line: %s", line)
			return 0, fmt.Errorf("failed to parse port number for interface %s", ifaceName)
		}
	}

	log.Errorf("Interface %s not found in BMv2 switch port list", ifaceName)
	return 0, fmt.Errorf("interface %s not found in BMv2 switch port list", ifaceName)
}
