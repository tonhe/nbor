package broadcast

import (
	"encoding/binary"
	"net"

	"nbor/config"
	"nbor/protocol"
	"nbor/types"
)

// BuildCDPFrame builds a complete CDP frame ready for transmission
func BuildCDPFrame(cfg *config.Config, iface *types.InterfaceInfo, systemName string) ([]byte, error) {
	// Build CDP payload (header + TLVs)
	cdpPayload := buildCDPPayload(cfg, iface, systemName)

	// Calculate checksum over CDP payload
	checksum := calculateChecksum(cdpPayload)
	// Insert checksum into payload (bytes 2-3)
	binary.BigEndian.PutUint16(cdpPayload[2:4], checksum)

	// Build complete frame
	// Ethernet header (14 bytes) + LLC (3 bytes) + SNAP (5 bytes) + CDP payload
	frameLen := 14 + 3 + 5 + len(cdpPayload)
	frame := make([]byte, frameLen)

	offset := 0

	// Ethernet header
	copy(frame[offset:offset+6], protocol.CDPMulticastMAC) // Destination MAC
	offset += 6
	copy(frame[offset:offset+6], iface.MAC) // Source MAC
	offset += 6
	// Length field for 802.3 frame (not EtherType)
	binary.BigEndian.PutUint16(frame[offset:offset+2], uint16(3+5+len(cdpPayload)))
	offset += 2

	// LLC header (3 bytes)
	frame[offset] = 0xAA   // DSAP
	frame[offset+1] = 0xAA // SSAP
	frame[offset+2] = 0x03 // Control
	offset += 3

	// SNAP header (5 bytes)
	frame[offset] = 0x00   // OUI byte 1
	frame[offset+1] = 0x00 // OUI byte 2
	frame[offset+2] = 0x0C // OUI byte 3 (Cisco)
	frame[offset+3] = 0x20 // Protocol ID byte 1 (CDP)
	frame[offset+4] = 0x00 // Protocol ID byte 2
	offset += 5

	// CDP payload
	copy(frame[offset:], cdpPayload)

	return frame, nil
}

// buildCDPPayload builds the CDP header and TLVs
func buildCDPPayload(cfg *config.Config, iface *types.InterfaceInfo, systemName string) []byte {
	var payload []byte

	// CDP header (4 bytes)
	header := make([]byte, 4)
	header[0] = 0x02                                     // Version 2
	header[1] = byte(cfg.TTL)                            // TTL in seconds
	binary.BigEndian.PutUint16(header[2:4], 0x0000)      // Checksum placeholder
	payload = append(payload, header...)

	// TLV: Device ID
	payload = append(payload, encodeCDPTLV(protocol.CDPTLVDeviceID, []byte(systemName))...)

	// TLV: Port ID
	payload = append(payload, encodeCDPTLV(protocol.CDPTLVPortID, []byte(iface.Name))...)

	// TLV: Capabilities
	capBits := protocol.BuildCDPCapabilities(cfg.Capabilities)
	capData := make([]byte, 4)
	binary.BigEndian.PutUint32(capData, capBits)
	payload = append(payload, encodeCDPTLV(protocol.CDPTLVCapabilities, capData)...)

	// TLV: Platform
	platform := "nbor"
	payload = append(payload, encodeCDPTLV(protocol.CDPTLVPlatform, []byte(platform))...)

	// TLV: Software Version (Description)
	description := cfg.SystemDescription
	if description == "" {
		description = "nbor network neighbor discovery tool"
	}
	payload = append(payload, encodeCDPTLV(protocol.CDPTLVVersion, []byte(description))...)

	// TLV: Addresses (if interface has IP)
	if len(iface.IPv4Addrs) > 0 {
		addrData := encodeCDPAddresses(iface.IPv4Addrs)
		payload = append(payload, encodeCDPTLV(protocol.CDPTLVAddress, addrData)...)
	}

	return payload
}

// encodeCDPTLV encodes a CDP TLV
func encodeCDPTLV(tlvType uint16, value []byte) []byte {
	// TLV format: Type (2 bytes) + Length (2 bytes, includes type and length) + Value
	length := uint16(4 + len(value))
	tlv := make([]byte, length)
	binary.BigEndian.PutUint16(tlv[0:2], tlvType)
	binary.BigEndian.PutUint16(tlv[2:4], length)
	copy(tlv[4:], value)
	return tlv
}

// encodeCDPAddresses encodes IP addresses for the Address TLV
func encodeCDPAddresses(ips []net.IP) []byte {
	// Format: Number of addresses (4 bytes) + address entries
	numAddrs := uint32(len(ips))
	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, numAddrs)

	for _, ip := range ips {
		ipv4 := ip.To4()
		if ipv4 == nil {
			continue
		}
		// Address entry format:
		// Protocol type (1 byte): 1 = NLPID
		// Protocol length (1 byte): 1
		// Protocol: 0xCC (IPv4)
		// Address length (2 bytes): 4
		// Address (4 bytes)
		entry := []byte{
			0x01,       // Protocol type (NLPID)
			0x01,       // Protocol length
			0xCC,       // Protocol (IPv4)
			0x00, 0x04, // Address length (big endian)
		}
		entry = append(entry, ipv4...)
		data = append(data, entry...)
	}

	return data
}

// calculateChecksum calculates the CDP checksum (RFC 1071 Internet checksum)
func calculateChecksum(data []byte) uint16 {
	var sum uint32

	// Sum all 16-bit words
	for i := 0; i+1 < len(data); i += 2 {
		sum += uint32(binary.BigEndian.Uint16(data[i : i+2]))
	}

	// Add odd byte if present
	if len(data)%2 == 1 {
		sum += uint32(data[len(data)-1]) << 8
	}

	// Fold 32-bit sum to 16 bits
	for sum > 0xFFFF {
		sum = (sum & 0xFFFF) + (sum >> 16)
	}

	// One's complement
	return ^uint16(sum)
}
