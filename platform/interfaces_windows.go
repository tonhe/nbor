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

		// Use description as display name, but store GUID mapping
		displayName := dev.Description
		if displayName == "" {
			displayName = dev.Name
		}

		interfaceMapping[displayName] = dev.Name

		// Extract IP addresses from pcap device info
		// This is more reliable on Windows than matching to net.Interface
		var ipv4Addrs, ipv6Addrs []net.IP
		for _, addr := range dev.Addresses {
			if addr.IP == nil {
				continue
			}
			if ip4 := addr.IP.To4(); ip4 != nil {
				if !ip4.IsLoopback() {
					ipv4Addrs = append(ipv4Addrs, ip4)
				}
			} else {
				if !addr.IP.IsLinkLocalUnicast() && !addr.IP.IsLoopback() {
					ipv6Addrs = append(ipv6Addrs, addr.IP)
				}
			}
		}

		// Try to find matching net.Interface for MAC and status
		// Use multiple matching strategies
		iface := findNetInterfaceByPcap(dev)

		var mac net.HardwareAddr
		var isUp bool
		var mtu int

		if iface != nil {
			mac = iface.HardwareAddr
			isUp = iface.Flags&net.FlagUp != 0
			mtu = iface.MTU
		} else {
			// If we couldn't match, assume it's up if it has addresses
			isUp = len(ipv4Addrs) > 0 || len(ipv6Addrs) > 0
			mtu = 1500 // Default MTU
		}

		// Skip interfaces without MAC (unless we have IPs, which means it's valid)
		if len(mac) == 0 && len(ipv4Addrs) == 0 && len(ipv6Addrs) == 0 {
			continue
		}

		info := types.InterfaceInfo{
			Name:      displayName,
			MAC:       mac,
			IsUp:      isUp,
			MTU:       mtu,
			Speed:     "", // Speed detection is complex on Windows
			IPv4Addrs: ipv4Addrs,
			IPv6Addrs: ipv6Addrs,
		}

		result = append(result, info)
	}

	return result, nil
}

// windowsExcludedKeywords lists keywords to exclude on Windows
var windowsExcludedKeywords = []string{
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

// windowsEthernetKeywords lists keywords that identify Ethernet adapters
var windowsEthernetKeywords = []string{
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

// isExcludedInterface checks if the interface description indicates a non-physical interface
func isExcludedInterface(desc string) bool {
	return hasExcludedKeyword(desc, windowsExcludedKeywords)
}

// isEthernetInterface checks if the interface description indicates an Ethernet adapter
func isEthernetInterface(desc string) bool {
	return hasExcludedKeyword(desc, windowsEthernetKeywords)
}

// findNetInterfaceByPcap finds the net.Interface matching a pcap device
// Uses multiple strategies: GUID matching, IP address matching
func findNetInterfaceByPcap(dev pcap.Interface) *net.Interface {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}

	// Strategy 1: Match by GUID
	// Windows pcap names look like: \Device\NPF_{GUID}
	guid := extractGUID(dev.Name)
	if guid != "" {
		for i := range ifaces {
			ifaceGUID := extractGUID(ifaces[i].Name)
			if ifaceGUID != "" && strings.EqualFold(guid, ifaceGUID) {
				return &ifaces[i]
			}
		}
	}

	// Strategy 2: Match by IP address
	// If pcap device has addresses, find net.Interface with same IP
	for _, pcapAddr := range dev.Addresses {
		if pcapAddr.IP == nil {
			continue
		}
		for i := range ifaces {
			addrs, err := ifaces[i].Addrs()
			if err != nil {
				continue
			}
			for _, addr := range addrs {
				var ip net.IP
				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}
				if ip != nil && ip.Equal(pcapAddr.IP) {
					return &ifaces[i]
				}
			}
		}
	}

	// Strategy 3: Try substring match on name
	for i := range ifaces {
		if strings.Contains(dev.Name, ifaces[i].Name) || strings.Contains(ifaces[i].Name, dev.Name) {
			return &ifaces[i]
		}
	}

	return nil
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

// GetAllInterfaces returns all pcap-visible interfaces without filtering
func GetAllInterfaces() ([]types.InterfaceInfo, error) {
	devices, err := pcap.FindAllDevs()
	if err != nil {
		return nil, err
	}

	var result []types.InterfaceInfo

	for _, dev := range devices {
		// Use description as display name, but store GUID mapping
		displayName := dev.Description
		if displayName == "" {
			displayName = dev.Name
		}

		interfaceMapping[displayName] = dev.Name

		// Extract IP addresses from pcap device info
		var ipv4Addrs, ipv6Addrs []net.IP
		for _, addr := range dev.Addresses {
			if addr.IP == nil {
				continue
			}
			if ip4 := addr.IP.To4(); ip4 != nil {
				if !ip4.IsLoopback() {
					ipv4Addrs = append(ipv4Addrs, ip4)
				}
			} else {
				if !addr.IP.IsLinkLocalUnicast() && !addr.IP.IsLoopback() {
					ipv6Addrs = append(ipv6Addrs, addr.IP)
				}
			}
		}

		// Try to find matching net.Interface for MAC and status
		iface := findNetInterfaceByPcap(dev)

		var mac net.HardwareAddr
		var isUp bool
		var mtu int

		if iface != nil {
			mac = iface.HardwareAddr
			isUp = iface.Flags&net.FlagUp != 0
			mtu = iface.MTU
		} else {
			isUp = len(ipv4Addrs) > 0 || len(ipv6Addrs) > 0
			mtu = 1500
		}

		// Skip interfaces without MAC and without IPs (not useful)
		if len(mac) == 0 && len(ipv4Addrs) == 0 && len(ipv6Addrs) == 0 {
			continue
		}

		info := types.InterfaceInfo{
			Name:      displayName,
			MAC:       mac,
			IsUp:      isUp,
			MTU:       mtu,
			Speed:     "",
			IPv4Addrs: ipv4Addrs,
			IPv6Addrs: ipv6Addrs,
		}

		result = append(result, info)
	}

	return result, nil
}

// windowsKeywordReasons maps exclusion keywords to filter reasons
var windowsKeywordReasons = map[string]string{
	"wireless":     "WiFi interface",
	"wi-fi":        "WiFi interface",
	"wifi":         "WiFi interface",
	"bluetooth":    "Bluetooth interface",
	"virtual":      "virtual interface",
	"vpn":          "VPN interface",
	"loopback":     "loopback interface",
	"tunnel":       "tunnel interface",
	"miniport":     "WAN miniport",
	"wan miniport": "WAN miniport",
	"microsoft":    "Microsoft virtual adapter",
	"hyper-v":      "virtual interface (Hyper-V)",
	"vmware":       "virtual interface (VMware)",
	"virtualbox":   "virtual interface (VirtualBox)",
	"npcap":        "Npcap loopback adapter",
	"filter":       "filter driver",
	"pseudo":       "pseudo interface",
}

// GetFilterReason returns why an interface was filtered, or empty string if not filtered
func GetFilterReason(name string) string {
	// Check exclusion keywords
	if reason := findKeywordReason(name, windowsKeywordReasons); reason != "" {
		return reason
	}

	// Check if it's not an Ethernet-like interface
	if !hasExcludedKeyword(name, windowsEthernetKeywords) {
		return "not recognized as Ethernet adapter"
	}

	return ""
}
