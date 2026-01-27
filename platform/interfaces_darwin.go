//go:build darwin

package platform

import (
	"net"
	"os/exec"
	"strings"

	"nbor/types"
)

// GetEthernetInterfaces returns a list of wired Ethernet interfaces on macOS
func GetEthernetInterfaces() ([]types.InterfaceInfo, error) {
	// Get list of WiFi interfaces from networksetup
	wifiInterfaces := getWiFiInterfaces()

	// Get interface status from ifconfig
	ifaceStatus := getInterfaceStatus()

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var result []types.InterfaceInfo

	for _, iface := range ifaces {
		// Skip loopback
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		// Skip interfaces without MAC
		if len(iface.HardwareAddr) == 0 {
			continue
		}

		// Skip WiFi interfaces detected via networksetup
		if wifiInterfaces[iface.Name] {
			continue
		}

		// Skip virtual/wireless interfaces by name pattern
		if isVirtualOrWirelessDarwin(iface.Name) {
			continue
		}

		// Verify pcap can open this interface
		if !canOpenInterface(iface.Name) {
			continue
		}

		// Get IP addresses
		ipv4Addrs, ipv6Addrs := types.GetInterfaceAddresses(&iface)

		// Check actual link status from ifconfig (not just net.FlagUp)
		isActive := ifaceStatus[iface.Name]

		info := types.InterfaceInfo{
			Name:      iface.Name,
			MAC:       iface.HardwareAddr,
			IsUp:      isActive,
			MTU:       iface.MTU,
			Speed:     getInterfaceSpeed(iface.Name),
			IPv4Addrs: ipv4Addrs,
			IPv6Addrs: ipv6Addrs,
		}

		result = append(result, info)
	}

	return result, nil
}

// getInterfaceStatus uses ifconfig to get actual link status
func getInterfaceStatus() map[string]bool {
	status := make(map[string]bool)

	cmd := exec.Command("ifconfig", "-a")
	output, err := cmd.Output()
	if err != nil {
		return status
	}

	lines := strings.Split(string(output), "\n")
	var currentIface string

	for _, line := range lines {
		// Interface line starts with interface name (no leading whitespace)
		if len(line) > 0 && line[0] != '\t' && line[0] != ' ' {
			// Extract interface name (before the colon)
			if idx := strings.Index(line, ":"); idx > 0 {
				currentIface = line[:idx]
				// Default to inactive
				status[currentIface] = false
			}
		}

		// Check for status line
		if currentIface != "" && strings.Contains(line, "status:") {
			if strings.Contains(line, "status: active") {
				status[currentIface] = true
			}
			// status: inactive means down
		}
	}

	return status
}

// getWiFiInterfaces uses networksetup to find WiFi interfaces
func getWiFiInterfaces() map[string]bool {
	wifiInterfaces := make(map[string]bool)

	// Get all hardware ports and find Wi-Fi ones
	cmd := exec.Command("networksetup", "-listallhardwareports")
	output, err := cmd.Output()
	if err != nil {
		return wifiInterfaces
	}

	lines := strings.Split(string(output), "\n")
	isWiFi := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "Hardware Port:") {
			portName := strings.TrimPrefix(line, "Hardware Port: ")
			// Check if this is a Wi-Fi port
			isWiFi = strings.Contains(strings.ToLower(portName), "wi-fi") ||
				strings.Contains(strings.ToLower(portName), "wifi") ||
				strings.Contains(strings.ToLower(portName), "airport")
		}

		if isWiFi && strings.HasPrefix(line, "Device:") {
			device := strings.TrimPrefix(line, "Device: ")
			device = strings.TrimSpace(device)
			if device != "" {
				wifiInterfaces[device] = true
			}
		}
	}

	return wifiInterfaces
}

// darwinExcludedPrefixes lists interface prefixes to exclude on macOS
var darwinExcludedPrefixes = []string{
	"lo",      // loopback
	"gif",     // generic tunnel
	"stf",     // 6to4 tunnel
	"awdl",    // Apple Wireless Direct Link
	"llw",     // Low Latency WLAN
	"utun",    // User tunnel
	"bridge",  // Bridge
	"ap",      // Access point
	"vmnet",   // VMware
	"vboxnet", // VirtualBox
	"anpi",    // Apple Network Plugin Interface
	"feth",    // Forwarding ethernet
	"ipsec",   // IPSec tunnel
	"ppp",     // PPP
}

// isVirtualOrWirelessDarwin checks if an interface is virtual or wireless on macOS
func isVirtualOrWirelessDarwin(name string) bool {
	return hasExcludedPrefix(name, darwinExcludedPrefixes)
}

// getInterfaceSpeed attempts to get link speed via system_profiler (expensive, so we skip)
func getInterfaceSpeed(name string) string {
	// system_profiler is too slow to call for each interface
	// Could use ioctl or CoreFoundation, but not worth the complexity
	return ""
}

// GetInterfaceDisplayName returns the display name for an interface
func GetInterfaceDisplayName(name string) string {
	return name
}

// GetInterfaceInternalName returns the internal name for pcap
func GetInterfaceInternalName(name string) string {
	return name
}

// GetAllInterfaces returns all network interfaces without filtering
func GetAllInterfaces() ([]types.InterfaceInfo, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	// Get interface status from ifconfig
	ifaceStatus := getInterfaceStatus()

	var result []types.InterfaceInfo

	for _, iface := range ifaces {
		// Skip loopback
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		// Skip interfaces without MAC
		if len(iface.HardwareAddr) == 0 {
			continue
		}

		// Get IP addresses
		ipv4Addrs, ipv6Addrs := types.GetInterfaceAddresses(&iface)

		// Check actual link status from ifconfig
		isActive := ifaceStatus[iface.Name]

		info := types.InterfaceInfo{
			Name:      iface.Name,
			MAC:       iface.HardwareAddr,
			IsUp:      isActive,
			MTU:       iface.MTU,
			Speed:     getInterfaceSpeed(iface.Name),
			IPv4Addrs: ipv4Addrs,
			IPv6Addrs: ipv6Addrs,
		}

		result = append(result, info)
	}

	return result, nil
}

// darwinPrefixReasons maps interface prefixes to filter reasons
var darwinPrefixReasons = map[string]string{
	"lo":      "loopback interface",
	"gif":     "tunnel interface",
	"stf":     "tunnel interface",
	"awdl":    "Apple Wireless Direct Link",
	"llw":     "Low Latency WLAN interface",
	"utun":    "tunnel interface",
	"bridge":  "bridge interface",
	"ap":      "access point interface",
	"vmnet":   "virtual interface (VMware)",
	"vboxnet": "virtual interface (VirtualBox)",
	"anpi":    "Apple Network Plugin Interface",
	"feth":    "forwarding ethernet interface",
	"ipsec":   "IPSec tunnel interface",
	"ppp":     "PPP interface",
}

// GetFilterReason returns why an interface was filtered, or empty string if not filtered
func GetFilterReason(name string) string {
	// Check WiFi interfaces
	wifiInterfaces := getWiFiInterfaces()
	if wifiInterfaces[name] {
		return "WiFi interface"
	}

	// Check virtual/wireless patterns
	if reason := findPrefixReason(name, darwinPrefixReasons); reason != "" {
		return reason
	}

	// Check if pcap can't open it
	if !canOpenInterface(name) {
		return "not available for capture"
	}

	return ""
}
