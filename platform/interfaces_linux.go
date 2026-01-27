//go:build linux

package platform

import (
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"nbor/types"
)

const sysClassNet = "/sys/class/net"

// GetEthernetInterfaces returns a list of wired Ethernet interfaces on Linux
func GetEthernetInterfaces() ([]types.InterfaceInfo, error) {
	entries, err := os.ReadDir(sysClassNet)
	if err != nil {
		return nil, err
	}

	var result []types.InterfaceInfo

	for _, entry := range entries {
		ifaceName := entry.Name()
		ifacePath := filepath.Join(sysClassNet, ifaceName)

		// Check interface type - type 1 = Ethernet
		typeFile := filepath.Join(ifacePath, "type")
		typeData, err := os.ReadFile(typeFile)
		if err != nil {
			continue
		}
		ifaceType, err := strconv.Atoi(strings.TrimSpace(string(typeData)))
		if err != nil || ifaceType != 1 {
			continue
		}

		// Check if wireless - wireless directory exists for wireless interfaces
		wirelessPath := filepath.Join(ifacePath, "wireless")
		if _, err := os.Stat(wirelessPath); err == nil {
			continue // Skip wireless
		}

		// Skip loopback and common virtual interfaces by name
		if isVirtualInterface(ifaceName) {
			continue
		}

		// Get interface from net package for additional info
		iface, err := net.InterfaceByName(ifaceName)
		if err != nil {
			continue
		}

		// Skip interfaces without MAC address
		if len(iface.HardwareAddr) == 0 {
			continue
		}

		// Verify pcap can open this interface
		if !canOpenInterface(ifaceName) {
			continue
		}

		// Get IP addresses
		ipv4Addrs, ipv6Addrs := types.GetInterfaceAddresses(iface)

		info := types.InterfaceInfo{
			Name:      ifaceName,
			MAC:       iface.HardwareAddr,
			IsUp:      iface.Flags&net.FlagUp != 0,
			MTU:       iface.MTU,
			Speed:     getInterfaceSpeed(ifaceName),
			IPv4Addrs: ipv4Addrs,
			IPv6Addrs: ipv6Addrs,
		}

		result = append(result, info)
	}

	return result, nil
}

// linuxExcludedPrefixes lists interface prefixes to exclude on Linux
var linuxExcludedPrefixes = []string{
	"lo",
	"veth",
	"docker",
	"br-",
	"virbr",
	"vnet",
	"tun",
	"tap",
	"bond",
	"dummy",
}

// isVirtualInterface checks if an interface name indicates a virtual interface
func isVirtualInterface(name string) bool {
	return hasExcludedPrefix(name, linuxExcludedPrefixes)
}

// getInterfaceSpeed reads the interface speed from sysfs
func getInterfaceSpeed(name string) string {
	speedFile := filepath.Join(sysClassNet, name, "speed")
	data, err := os.ReadFile(speedFile)
	if err != nil {
		return ""
	}

	speedMbps := strings.TrimSpace(string(data))
	if speedMbps == "" || speedMbps == "-1" {
		return ""
	}

	speed, err := strconv.Atoi(speedMbps)
	if err != nil {
		return ""
	}

	if speed >= 1000 {
		return strconv.Itoa(speed/1000) + " Gbps"
	}
	return speedMbps + " Mbps"
}

// GetInterfaceDisplayName returns the display name for an interface
// On Linux, the interface name is used directly
func GetInterfaceDisplayName(name string) string {
	return name
}

// GetInterfaceInternalName returns the internal name for pcap
// On Linux, this is the same as the interface name
func GetInterfaceInternalName(name string) string {
	return name
}

// GetAllInterfaces returns all network interfaces without filtering
func GetAllInterfaces() ([]types.InterfaceInfo, error) {
	entries, err := os.ReadDir(sysClassNet)
	if err != nil {
		return nil, err
	}

	var result []types.InterfaceInfo

	for _, entry := range entries {
		ifaceName := entry.Name()
		ifacePath := filepath.Join(sysClassNet, ifaceName)

		// Check interface type - type 1 = Ethernet
		typeFile := filepath.Join(ifacePath, "type")
		typeData, err := os.ReadFile(typeFile)
		if err != nil {
			continue
		}
		ifaceType, err := strconv.Atoi(strings.TrimSpace(string(typeData)))
		if err != nil || ifaceType != 1 {
			continue
		}

		// Get interface from net package for additional info
		iface, err := net.InterfaceByName(ifaceName)
		if err != nil {
			continue
		}

		// Skip interfaces without MAC address
		if len(iface.HardwareAddr) == 0 {
			continue
		}

		// Get IP addresses
		ipv4Addrs, ipv6Addrs := types.GetInterfaceAddresses(iface)

		info := types.InterfaceInfo{
			Name:      ifaceName,
			MAC:       iface.HardwareAddr,
			IsUp:      iface.Flags&net.FlagUp != 0,
			MTU:       iface.MTU,
			Speed:     getInterfaceSpeed(ifaceName),
			IPv4Addrs: ipv4Addrs,
			IPv6Addrs: ipv6Addrs,
		}

		result = append(result, info)
	}

	return result, nil
}

// linuxPrefixReasons maps interface prefixes to filter reasons
var linuxPrefixReasons = map[string]string{
	"lo":     "loopback interface",
	"veth":   "virtual Ethernet (container)",
	"docker": "Docker bridge interface",
	"br-":    "bridge interface",
	"virbr":  "virtual bridge (libvirt)",
	"vnet":   "virtual network interface",
	"tun":    "tunnel interface",
	"tap":    "TAP interface",
	"bond":   "bonding interface",
	"dummy":  "dummy interface",
}

// GetFilterReason returns why an interface was filtered, or empty string if not filtered
func GetFilterReason(name string) string {
	ifacePath := filepath.Join(sysClassNet, name)

	// Check if wireless
	wirelessPath := filepath.Join(ifacePath, "wireless")
	if _, err := os.Stat(wirelessPath); err == nil {
		return "WiFi interface"
	}

	// Check virtual interface patterns
	if reason := findPrefixReason(name, linuxPrefixReasons); reason != "" {
		return reason
	}

	// Check if pcap can't open it
	if !canOpenInterface(name) {
		return "not available for capture"
	}

	return ""
}
