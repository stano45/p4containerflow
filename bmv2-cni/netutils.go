package main

import (
	"fmt"
	"net"
	"os/exec"

	log "k8s.io/klog/v2"

	"github.com/vishvananda/netlink"
)

func buildHostVethName(containerID string) string {
	return fmt.Sprintf("veth%s", containerID[:5])
}

func addStaticARPEntry(ifaceName string, ip net.IP, mac net.HardwareAddr) error {
	log.Infof("Adding static ARP entry for IP %s and MAC %s on interface %s", ip, mac, ifaceName)
	// Find the link (interface) by its name
	link, err := netlink.LinkByName(ifaceName)
	if err != nil {
		return fmt.Errorf("failed to find interface %s: %v", ifaceName, err)
	}

	// Create the ARP entry
	arp := &netlink.Neigh{
		LinkIndex:    link.Attrs().Index,
		IP:           ip,
		HardwareAddr: mac,
		State:        netlink.NUD_PERMANENT, // Static ARP entry
	}

	// Add the ARP entry to the interface
	if err := netlink.NeighAdd(arp); err != nil {
		return fmt.Errorf("failed to add static ARP entry: %v", err)
	}

	return nil
}

func enableProxyARP(iface string) error {
	// Enable Proxy ARP for a specific interface
	cmd := exec.Command("sh", "-c", fmt.Sprintf("echo 1 > /proc/sys/net/ipv4/conf/%s/proxy_arp", iface))
	if err := cmd.Run(); err != nil {
		return err
	}

	// Enable Proxy ARP for all interfaces globally
	globalCmd := exec.Command("sh", "-c", "echo 1 > /proc/sys/net/ipv4/conf/all/proxy_arp")
	return globalCmd.Run()
}

// func disableIPForwarding(ifaceName string) error {
// 	forwardingPath := fmt.Sprintf("/proc/sys/net/ipv4/conf/%s/forwarding", ifaceName)
// 	err := os.WriteFile(forwardingPath, []byte("0"), 0644)
// 	if err != nil {
// 		return fmt.Errorf("failed to disable IP forwarding for interface %s: %v", ifaceName, err)
// 	}
// 	logger.Printf("IP forwarding disabled for interface %s", ifaceName)
// 	return nil
// }

// Helper function to delete all existing routes
func deleteAllRoutes() error {
	routes, err := netlink.RouteList(nil, netlink.FAMILY_V4)
	if err != nil {
		return fmt.Errorf("failed to list routes: %v", err)
	}

	for _, route := range routes {
		if err := netlink.RouteDel(&route); err != nil {
			return fmt.Errorf("failed to delete route %v: %v", route, err)
		}
	}

	return nil
}
