package capture

import (
	"errors"
	"fmt"
	"net"
	"runtime"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

var (
	// CDP multicast address
	CDPMulticast = net.HardwareAddr{0x01, 0x00, 0x0c, 0xcc, 0xcc, 0xcc}
	// LLDP multicast address
	LLDPMulticast = net.HardwareAddr{0x01, 0x80, 0xc2, 0x00, 0x00, 0x0e}
)

// ErrInterfaceNotFound is returned when the specified interface doesn't exist
var ErrInterfaceNotFound = errors.New("interface not found")

// ErrInterfaceDown is returned when the interface is down
var ErrInterfaceDown = errors.New("interface is down")

// Capturer handles packet capture on an interface
type Capturer struct {
	handle      *pcap.Handle
	iface       string
	packets     chan gopacket.Packet
	stop        chan struct{}
	stopped     bool
	ownsHandle  bool // Whether this capturer owns the handle (should close it on stop)
}

// NewCapturer creates a new packet capturer for the given interface
func NewCapturer(ifaceName string) (*Capturer, error) {
	// On Windows, interface names are GUIDs that don't exist in net.Interfaces
	// So we skip the interface check on Windows and rely on pcap to validate
	if runtime.GOOS != "windows" {
		iface, err := net.InterfaceByName(ifaceName)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrInterfaceNotFound, ifaceName)
		}

		if iface.Flags&net.FlagUp == 0 {
			return nil, fmt.Errorf("%w: %s", ErrInterfaceDown, ifaceName)
		}
	}

	// Open pcap handle
	// Snapshot length of 65535 to capture full packets
	// Promiscuous mode to see all packets
	handle, err := pcap.OpenLive(ifaceName, 65535, true, pcap.BlockForever)
	if err != nil {
		return nil, fmt.Errorf("failed to open interface %s: %w", ifaceName, err)
	}

	// Set BPF filter to only capture CDP and LLDP packets
	filter := "ether dst 01:00:0c:cc:cc:cc or ether dst 01:80:c2:00:00:0e"
	if err := handle.SetBPFFilter(filter); err != nil {
		handle.Close()
		return nil, fmt.Errorf("failed to set BPF filter: %w", err)
	}

	return &Capturer{
		handle:     handle,
		iface:      ifaceName,
		packets:    make(chan gopacket.Packet, 100),
		stop:       make(chan struct{}),
		ownsHandle: true,
	}, nil
}

// NewCapturerWithHandle creates a new capturer using an existing pcap handle
// The handle should already have BPF filter set
// The caller is responsible for closing the handle
func NewCapturerWithHandle(handle *pcap.Handle, ifaceName string) *Capturer {
	return &Capturer{
		handle:     handle,
		iface:      ifaceName,
		packets:    make(chan gopacket.Packet, 100),
		stop:       make(chan struct{}),
		ownsHandle: false,
	}
}

// Start begins capturing packets
func (c *Capturer) Start() <-chan gopacket.Packet {
	go func() {
		packetSource := gopacket.NewPacketSource(c.handle, c.handle.LinkType())
		packetSource.NoCopy = true

		for {
			select {
			case <-c.stop:
				return
			default:
				packet, err := packetSource.NextPacket()
				if err != nil {
					// Check if we're stopping
					select {
					case <-c.stop:
						return
					default:
						continue
					}
				}

				select {
				case c.packets <- packet:
				case <-c.stop:
					return
				default:
					// Drop packet if channel is full
				}
			}
		}
	}()

	return c.packets
}

// Stop stops the packet capture
func (c *Capturer) Stop() {
	if c.stopped {
		return
	}
	c.stopped = true
	close(c.stop)
	if c.ownsHandle {
		c.handle.Close()
	}
}

// Interface returns the interface name
func (c *Capturer) Interface() string {
	return c.iface
}

// IsCDPPacket checks if a packet is destined for the CDP multicast address
func IsCDPPacket(packet gopacket.Packet) bool {
	ethLayer := packet.Layer(layers.LayerTypeEthernet)
	if ethLayer == nil {
		return false
	}
	eth := ethLayer.(*layers.Ethernet)
	return eth.DstMAC.String() == CDPMulticast.String()
}

// IsLLDPPacket checks if a packet is destined for the LLDP multicast address
func IsLLDPPacket(packet gopacket.Packet) bool {
	ethLayer := packet.Layer(layers.LayerTypeEthernet)
	if ethLayer == nil {
		return false
	}
	eth := ethLayer.(*layers.Ethernet)
	return eth.DstMAC.String() == LLDPMulticast.String()
}

// GetSourceMAC extracts the source MAC address from a packet
func GetSourceMAC(packet gopacket.Packet) net.HardwareAddr {
	ethLayer := packet.Layer(layers.LayerTypeEthernet)
	if ethLayer == nil {
		return nil
	}
	eth := ethLayer.(*layers.Ethernet)
	return eth.SrcMAC
}
