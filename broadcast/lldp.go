package broadcast

import (
	"encoding/binary"
	"net"

	"nbor/config"
	"nbor/protocol"
	"nbor/types"
)

// BuildLLDPFrame builds a complete LLDP frame ready for transmission
func BuildLLDPFrame(cfg *config.Config, iface *types.InterfaceInfo, systemName string) ([]byte, error) {
	// Build LLDP payload (TLVs)
	lldpPayload := buildLLDPPayload(cfg, iface, systemName)

	// Build complete frame
	// Ethernet header (14 bytes) + LLDP payload
	frameLen := 14 + len(lldpPayload)
	frame := make([]byte, frameLen)

	offset := 0

	// Ethernet header
	copy(frame[offset:offset+6], protocol.LLDPMulticastMAC) // Destination MAC
	offset += 6
	copy(frame[offset:offset+6], iface.MAC) // Source MAC
	offset += 6
	binary.BigEndian.PutUint16(frame[offset:offset+2], protocol.LLDPEtherType) // EtherType
	offset += 2

	// LLDP payload
	copy(frame[offset:], lldpPayload)

	return frame, nil
}

// buildLLDPPayload builds the LLDP TLVs
func buildLLDPPayload(cfg *config.Config, iface *types.InterfaceInfo, systemName string) []byte {
	var payload []byte

	// Mandatory TLV: Chassis ID (using MAC address)
	chassisIDData := make([]byte, 1+6)
	chassisIDData[0] = protocol.LLDPChassisIDSubtypeMAC
	copy(chassisIDData[1:], iface.MAC)
	payload = append(payload, encodeLLDPTLV(protocol.LLDPTLVChassisID, chassisIDData)...)

	// Mandatory TLV: Port ID (using interface name)
	portIDData := make([]byte, 1+len(iface.Name))
	portIDData[0] = protocol.LLDPPortIDSubtypeIfaceName
	copy(portIDData[1:], iface.Name)
	payload = append(payload, encodeLLDPTLV(protocol.LLDPTLVPortID, portIDData)...)

	// Mandatory TLV: TTL
	ttlData := make([]byte, 2)
	binary.BigEndian.PutUint16(ttlData, uint16(cfg.TTL))
	payload = append(payload, encodeLLDPTLV(protocol.LLDPTLVTTL, ttlData)...)

	// Optional TLV: Port Description
	payload = append(payload, encodeLLDPTLV(protocol.LLDPTLVPortDesc, []byte(iface.Name))...)

	// Optional TLV: System Name
	payload = append(payload, encodeLLDPTLV(protocol.LLDPTLVSystemName, []byte(systemName))...)

	// Optional TLV: System Description
	description := cfg.SystemDescription
	if description == "" {
		description = "nbor network neighbor discovery tool"
	}
	payload = append(payload, encodeLLDPTLV(protocol.LLDPTLVSystemDesc, []byte(description))...)

	// Optional TLV: System Capabilities
	capBits := protocol.BuildLLDPCapabilities(cfg.Capabilities)
	capData := make([]byte, 4)
	binary.BigEndian.PutUint16(capData[0:2], capBits) // System capabilities
	binary.BigEndian.PutUint16(capData[2:4], capBits) // Enabled capabilities
	payload = append(payload, encodeLLDPTLV(protocol.LLDPTLVSystemCap, capData)...)

	// Optional TLV: Management Address (if interface has IP)
	if len(iface.IPv4Addrs) > 0 {
		mgmtData := encodeLLDPMgmtAddress(iface.IPv4Addrs[0], iface.Name)
		payload = append(payload, encodeLLDPTLV(protocol.LLDPTLVMgmtAddress, mgmtData)...)
	}

	// End TLV (type 0, length 0)
	payload = append(payload, 0x00, 0x00)

	return payload
}

// encodeLLDPTLV encodes an LLDP TLV
// LLDP TLV format: Type (7 bits) + Length (9 bits) = 2 bytes header + Value
func encodeLLDPTLV(tlvType uint8, value []byte) []byte {
	length := len(value)
	if length > 511 {
		length = 511 // Max length is 9 bits
	}

	// Pack type (7 bits) and length (9 bits) into 2 bytes
	header := (uint16(tlvType) << 9) | uint16(length)

	tlv := make([]byte, 2+length)
	binary.BigEndian.PutUint16(tlv[0:2], header)
	copy(tlv[2:], value[:length])
	return tlv
}

// encodeLLDPMgmtAddress encodes the management address TLV data
func encodeLLDPMgmtAddress(ip net.IP, ifaceName string) []byte {
	ipv4 := ip.To4()
	if ipv4 == nil {
		return nil
	}

	// Management address TLV format:
	// Address string length (1 byte) = 1 + IP length
	// Address subtype (1 byte): 1 = IPv4
	// Address (4 bytes for IPv4)
	// Interface numbering subtype (1 byte): 2 = ifIndex
	// Interface number (4 bytes)
	// OID string length (1 byte): 0

	data := make([]byte, 12)
	data[0] = 5                   // Address string length (1 subtype + 4 IP bytes)
	data[1] = 1                   // Address subtype (IPv4)
	copy(data[2:6], ipv4)         // IP address
	data[6] = 2                   // Interface numbering subtype (ifIndex)
	binary.BigEndian.PutUint32(data[7:11], 1) // Interface number (use 1)
	data[11] = 0                  // OID string length

	return data
}
