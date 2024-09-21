package main

import (
	"encoding/json"
	"fmt"
	"strings"

	log "k8s.io/klog/v2"
)

func parseNetConf(bytes []byte) (*BMv2NetConf, error) {
	n := &BMv2NetConf{}
	if err := json.Unmarshal(bytes, n); err != nil {
		log.Errorf("Error loading config: %v", err)
		return nil, err
	}
	log.Errorf("Config loaded successfully")
	return n, nil
}

func parseArgs(cniArgs string) (string, string, error) {
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

// TODO: configure this via annotations in pod manifest
func isClientOrServer(podName string) (bool, bool) {
	isClient, isServer := false, false
	if strings.Contains(podName, "client") {
		isClient = true
	}
	if strings.Contains(podName, "server") {
		isServer = true
	}
	return isClient, isServer
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
// 		log.Printf("Error getting pod %s in namespace %s: %v", podName, namespace, err)
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
