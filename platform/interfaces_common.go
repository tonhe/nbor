// Package platform provides OS-specific network interface operations.
package platform

import (
	"strings"

	"github.com/google/gopacket/pcap"
)

// canOpenInterface checks if pcap can open the interface by name
// This verifies the interface is available for packet capture
func canOpenInterface(name string) bool {
	devices, err := pcap.FindAllDevs()
	if err != nil {
		return false
	}

	for _, dev := range devices {
		if dev.Name == name {
			return true
		}
	}

	return false
}

// hasExcludedPrefix checks if name starts with any of the given prefixes
func hasExcludedPrefix(name string, prefixes []string) bool {
	nameLower := strings.ToLower(name)
	for _, prefix := range prefixes {
		if strings.HasPrefix(nameLower, prefix) {
			return true
		}
	}
	return false
}

// hasExcludedKeyword checks if name contains any of the given keywords
func hasExcludedKeyword(name string, keywords []string) bool {
	nameLower := strings.ToLower(name)
	for _, keyword := range keywords {
		if strings.Contains(nameLower, keyword) {
			return true
		}
	}
	return false
}

// findPrefixReason returns the reason string for the first matching prefix, or empty string
func findPrefixReason(name string, prefixReasons map[string]string) string {
	nameLower := strings.ToLower(name)
	for prefix, reason := range prefixReasons {
		if strings.HasPrefix(nameLower, prefix) {
			return reason
		}
	}
	return ""
}

// findKeywordReason returns the reason string for the first matching keyword, or empty string
func findKeywordReason(name string, keywordReasons map[string]string) string {
	nameLower := strings.ToLower(name)
	for keyword, reason := range keywordReasons {
		if strings.Contains(nameLower, keyword) {
			return reason
		}
	}
	return ""
}
