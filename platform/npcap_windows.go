//go:build windows

package platform

import (
	"github.com/google/gopacket/pcap"
)

// CheckNpcap verifies that Npcap is installed and working
func CheckNpcap() error {
	_, err := pcap.FindAllDevs()
	if err != nil {
		return ErrNpcapNotFound
	}
	return nil
}
