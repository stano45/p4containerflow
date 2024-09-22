package main

import (
	"fmt"
	"net"

	log "k8s.io/klog/v2"

	"github.com/vishvananda/netlink"
)

// buildHostVethName constructs the host veth interface name based on the container ID.
// It logs the input and the resulting veth name.
func buildHostVethName(containerID string) string {
	log.Infof("Building host veth name using container ID: %s", containerID)
	if len(containerID) < 5 {
		log.Warningf("Container ID %s is shorter than expected. Using full ID for veth name.", containerID)
		return fmt.Sprintf("veth%s", containerID)
	}
	vethName := fmt.Sprintf("veth%s", containerID[:5])
	log.Infof("Constructed host veth name: %s", vethName)
	return vethName
}

// addStaticARPEntry adds a static ARP entry to the specified interface.
// It logs each step, including finding the interface, creating the ARP entry, and adding it.
func addStaticARPEntry(ifaceName string, ip net.IP, mac net.HardwareAddr) error {
	log.Infof("Adding static ARP entry for IP %s and MAC %s on interface %s", ip, mac, ifaceName)

	// Find the link (interface) by its name
	link, err := netlink.LinkByName(ifaceName)
	if err != nil {
		log.Errorf("Failed to find interface %s: %v", ifaceName, err)
		return fmt.Errorf("failed to find interface %s: %v", ifaceName, err)
	}
	log.Infof("Found interface %s: %+v", ifaceName, link)

	// Create the ARP entry
	arp := &netlink.Neigh{
		LinkIndex:    link.Attrs().Index,
		IP:           ip,
		HardwareAddr: mac,
		State:        netlink.NUD_PERMANENT, // Static ARP entry
	}
	log.Infof("Created ARP entry struct: %+v", arp)

	// Add the ARP entry to the interface
	if err := netlink.NeighAdd(arp); err != nil {
		log.Errorf("Failed to add static ARP entry: %v", err)
		return fmt.Errorf("failed to add static ARP entry: %v", err)
	}
	log.Infof("Successfully added static ARP entry for IP %s on interface %s", ip, ifaceName)

	return nil
}

// enableProxyARP enables Proxy ARP on the specified interface and globally.
// It logs each step, including executing shell commands and any errors encountered.
// func enableProxyARP(iface string) error {
// 	log.Infof("Enabling Proxy ARP for interface %s", iface)

// 	// Enable Proxy ARP for the specific interface
// 	cmd := exec.Command("sh", "-c", fmt.Sprintf("echo 1 > /proc/sys/net/ipv4/conf/%s/proxy_arp", iface))
// 	if err := cmd.Run(); err != nil {
// 		log.Errorf("Failed to enable Proxy ARP for interface %s: %v", iface, err)
// 		return fmt.Errorf("failed to enable Proxy ARP for interface %s: %v", iface, err)
// 	}
// 	log.Infof("Proxy ARP enabled for interface %s", iface)

// 	// Enable Proxy ARP globally for all interfaces
// 	globalCmd := exec.Command("sh", "-c", "echo 1 > /proc/sys/net/ipv4/conf/all/proxy_arp")
// 	if err := globalCmd.Run(); err != nil {
// 		log.Errorf("Failed to enable global Proxy ARP: %v", err)
// 		return fmt.Errorf("failed to enable global Proxy ARP: %v", err)
// 	}
// 	log.Infof("Global Proxy ARP enabled successfully")

// 	return nil
// }

// disableIPForwarding disables IP forwarding on the specified interface.
// It logs each step, including writing to the sysctl interface and any errors encountered.
// func disableIPForwarding(ifaceName string) error {
// 	log.Infof("Disabling IP forwarding for interface %s", ifaceName)
// 	forwardingPath := fmt.Sprintf("/proc/sys/net/ipv4/conf/%s/forwarding", ifaceName)
// 	err := os.WriteFile(forwardingPath, []byte("0"), 0644)
// 	if err != nil {
// 		log.Errorf("Failed to disable IP forwarding for interface %s: %v", ifaceName, err)
// 		return fmt.Errorf("failed to disable IP forwarding for interface %s: %v", ifaceName, err)
// 	}
// 	log.Infof("IP forwarding disabled for interface %s", ifaceName)
// 	return nil
// }

// deleteAllRoutes removes all existing IPv4 routes from the system.
// It logs the number of routes found and each deletion attempt.
func deleteAllRoutes() error {
	log.Infof("Deleting all existing IPv4 routes")
	routes, err := netlink.RouteList(nil, netlink.FAMILY_V4)
	if err != nil {
		log.Errorf("Failed to list routes: %v", err)
		return fmt.Errorf("failed to list routes: %v", err)
	}
	log.Infof("Found %d IPv4 routes to delete", len(routes))

	for i, route := range routes {
		log.Infof("Attempting to delete route #%d: %+v", i+1, route)
		if err := netlink.RouteDel(&route); err != nil {
			log.Errorf("Failed to delete route %+v: %v", route, err)
			return fmt.Errorf("failed to delete route %+v: %v", route, err)
		}
		log.Infof("Successfully deleted route #%d: %+v", i+1, route)
	}

	log.Infof("All IPv4 routes deleted successfully")
	return nil
}
