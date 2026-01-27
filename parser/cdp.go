// Package parser provides CDP and LLDP packet parsing.
package parser

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"nbor/protocol"
	"nbor/types"
)

// ParseCDP parses a CDP packet and returns a Neighbor struct
func ParseCDP(packet gopacket.Packet, ifaceName string) (*types.Neighbor, error) {
	// Get the CDP layer
	cdpLayer := packet.Layer(layers.LayerTypeCiscoDiscovery)
	if cdpLayer == nil {
		return nil, fmt.Errorf("not a CDP packet")
	}

	cdp := cdpLayer.(*layers.CiscoDiscovery)

	neighbor := &types.Neighbor{
		Protocol:  types.ProtocolCDP,
		LastSeen:  time.Now(),
		Interface: ifaceName,
	}

	// Get source MAC from ethernet layer
	if ethLayer := packet.Layer(layers.LayerTypeEthernet); ethLayer != nil {
		eth := ethLayer.(*layers.Ethernet)
		neighbor.SourceMAC = eth.SrcMAC
	}

	// Parse TLVs
	for _, tlv := range cdp.Values {
		switch tlv.Type {
		case layers.CDPTLVDevID:
			neighbor.ID = string(tlv.Value)
			neighbor.Hostname = string(tlv.Value)

		case layers.CDPTLVPortID:
			neighbor.PortID = string(tlv.Value)

		case layers.CDPTLVPlatform:
			neighbor.Platform = string(tlv.Value)

		case layers.CDPTLVVersion:
			neighbor.Description = string(tlv.Value)

		case layers.CDPTLVCapabilities:
			neighbor.Capabilities = parseCDPCapabilities(tlv.Value)

		case layers.CDPTLVAddress:
			if ip := parseCDPAddresses(tlv.Value); ip != nil {
				neighbor.ManagementIP = ip
			}

		case layers.CDPTLVMgmtAddresses:
			if ip := parseCDPAddresses(tlv.Value); ip != nil {
				neighbor.ManagementIP = ip
			}

		case layers.CDPTLVLocation:
			neighbor.Location = parseCDPLocation(tlv.Value)
		}
	}

	// Use source MAC as ID if device ID is empty
	if neighbor.ID == "" && neighbor.SourceMAC != nil {
		neighbor.ID = neighbor.SourceMAC.String()
	}

	return neighbor, nil
}

// parseCDPCapabilities parses the CDP capabilities field
func parseCDPCapabilities(data []byte) []types.Capability {
	return protocol.ParseCDPCapabilities(data)
}

// parseCDPAddresses parses the CDP address TLV
func parseCDPAddresses(data []byte) net.IP {
	if len(data) < 4 {
		return nil
	}

	// Number of addresses
	numAddrs := binary.BigEndian.Uint32(data[:4])
	if numAddrs == 0 {
		return nil
	}

	offset := 4

	// Parse first address
	// Protocol type (1 byte) + Protocol length (1 byte)
	if offset+2 > len(data) {
		return nil
	}

	protoType := data[offset]
	protoLen := int(data[offset+1])
	offset += 2

	// Skip protocol identifier
	if offset+protoLen > len(data) {
		return nil
	}
	offset += protoLen

	// Address length (2 bytes)
	if offset+2 > len(data) {
		return nil
	}
	addrLen := int(binary.BigEndian.Uint16(data[offset : offset+2]))
	offset += 2

	// Address
	if offset+addrLen > len(data) {
		return nil
	}

	// Check if this is an IP address (protocol type 1 = NLPID, 0xCC = IPv4)
	if protoType == 1 && addrLen == 4 {
		return net.IP(data[offset : offset+4])
	}

	// Could also be IPv6
	if addrLen == 16 {
		return net.IP(data[offset : offset+16])
	}

	return nil
}

// parseCDPLocation parses the CDP location TLV
func parseCDPLocation(data []byte) string {
	if len(data) < 1 {
		return ""
	}

	// Location format: 1 byte type + location string
	// Type 1 = ASCII string
	if data[0] == 1 && len(data) > 1 {
		return string(data[1:])
	}

	// Some implementations don't include the type byte
	return string(data)
}
