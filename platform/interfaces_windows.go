//go:build windows

package platform

import (
	"net"
	"strings"

	"github.com/google/gopacket/pcap"

	"nbor/types"
)

// interfaceMapping maps friendly names to internal GUID names
var interfaceMapping = make(map[string]string)

// GetEthernetInterfaces returns a list of wired Ethernet interfaces on Windows
func GetEthernetInterfaces() ([]types.InterfaceInfo, error) {
	devices, err := pcap.FindAllDevs()
	if err != nil {
		return nil, err
	}

	var result []types.InterfaceInfo

	for _, dev := range devices {
		// Get the description for filtering
		desc := strings.ToLower(dev.Description)

		// Skip wireless/virtual interfaces
		if isExcludedInterface(desc) {
			continue
		}

		// Include only Ethernet-like interfaces
		if !isEthernetInterface(desc) {
			continue
		}

		// Get interface from net package for MAC and status
		iface, err := findNetInterface(dev.Name)
		if err != nil {
			continue
		}

		// Skip interfaces without MAC
		if len(iface.HardwareAddr) == 0 {
			continue
		}

		// Use description as display name, but store GUID mapping
		displayName := dev.Description
		if displayName == "" {
			displayName = dev.Name
		}

		interfaceMapping[displayName] = dev.Name

		// Get IP addresses
		ipv4Addrs, ipv6Addrs := types.GetInterfaceAddresses(iface)

		info := types.InterfaceInfo{
			Name:      displayName,
			MAC:       iface.HardwareAddr,
			IsUp:      iface.Flags&net.FlagUp != 0,
			MTU:       iface.MTU,
			Speed:     "", // Speed detection is complex on Windows
			IPv4Addrs: ipv4Addrs,
			IPv6Addrs: ipv6Addrs,
		}

		result = append(result, info)
	}

	return result, nil
}

// isExcludedInterface checks if the interface description indicates a non-physical interface
func isExcludedInterface(desc string) bool {
	excludeKeywords := []string{
		"wireless",
		"wi-fi",
		"wifi",
		"bluetooth",
		"virtual",
		"vpn",
		"loopback",
		"tunnel",
		"miniport",
		"wan miniport",
		"microsoft",
		"hyper-v",
		"vmware",
		"virtualbox",
		"npcap",
		"filter",
		"pseudo",
	}

	for _, keyword := range excludeKeywords {
		if strings.Contains(desc, keyword) {
			return true
		}
	}

	return false
}

// isEthernetInterface checks if the interface description indicates an Ethernet adapter
func isEthernetInterface(desc string) bool {
	includeKeywords := []string{
		"ethernet",
		"gbe",
		"nic",
		"gigabit",
		"10/100",
		"realtek",
		"intel",
		"broadcom",
		"marvell",
		"network adapter",
		"pci",
		"usb ethernet",
		"usb3.0 to ethernet",
	}

	for _, keyword := range includeKeywords {
		if strings.Contains(desc, keyword) {
			return true
		}
	}

	return false
}

// findNetInterface finds the net.Interface matching a pcap device name
func findNetInterface(pcapName string) (*net.Interface, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	// Windows pcap names look like: \Device\NPF_{GUID}
	// Net interface names may have a different format
	// Try to match by extracting the GUID
	guid := extractGUID(pcapName)

	for i := range ifaces {
		ifaceGUID := extractGUID(ifaces[i].Name)
		if guid != "" && ifaceGUID != "" && strings.EqualFold(guid, ifaceGUID) {
			return &ifaces[i], nil
		}
	}

	// Fallback: try exact match
	for i := range ifaces {
		if strings.Contains(pcapName, ifaces[i].Name) || strings.Contains(ifaces[i].Name, pcapName) {
			return &ifaces[i], nil
		}
	}

	// Return first interface with a MAC as last resort
	for i := range ifaces {
		if len(ifaces[i].HardwareAddr) > 0 {
			return &ifaces[i], nil
		}
	}

	return nil, net.ErrClosed
}

// extractGUID extracts a GUID from a string
func extractGUID(s string) string {
	// Look for pattern like {xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx}
	start := strings.Index(s, "{")
	end := strings.Index(s, "}")

	if start >= 0 && end > start {
		return strings.ToLower(s[start : end+1])
	}

	return ""
}

// GetInterfaceDisplayName returns the display name for an interface
// On Windows, this is the friendly description
func GetInterfaceDisplayName(name string) string {
	return name
}

// GetInterfaceInternalName returns the internal GUID name for pcap
// On Windows, we need to look up the mapping
func GetInterfaceInternalName(displayName string) string {
	if internal, ok := interfaceMapping[displayName]; ok {
		return internal
	}
	return displayName
}
