package parser

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"nbor/protocol"
	"nbor/types"
)

// ParseLLDP parses an LLDP packet and returns a Neighbor struct
func ParseLLDP(packet gopacket.Packet, ifaceName string) (*types.Neighbor, error) {
	// Try to get the LLDP layer from gopacket
	lldpLayer := packet.Layer(layers.LayerTypeLinkLayerDiscovery)
	if lldpLayer == nil {
		return nil, fmt.Errorf("not an LLDP packet")
	}

	lldp := lldpLayer.(*layers.LinkLayerDiscovery)

	neighbor := &types.Neighbor{
		Protocol:  types.ProtocolLLDP,
		LastSeen:  time.Now(),
		Interface: ifaceName,
	}

	// Get source MAC from ethernet layer
	if ethLayer := packet.Layer(layers.LayerTypeEthernet); ethLayer != nil {
		eth := ethLayer.(*layers.Ethernet)
		neighbor.SourceMAC = eth.SrcMAC
	}

	// Parse Chassis ID
	neighbor.ID = parseLLDPChassisID(lldp.ChassisID)

	// Parse Port ID
	neighbor.PortID = parseLLDPPortID(lldp.PortID)

	// Get LLDP info layer for additional TLVs
	lldpInfoLayer := packet.Layer(layers.LayerTypeLinkLayerDiscoveryInfo)
	if lldpInfoLayer != nil {
		lldpInfo := lldpInfoLayer.(*layers.LinkLayerDiscoveryInfo)

		neighbor.PortDescription = lldpInfo.PortDescription
		neighbor.Hostname = lldpInfo.SysName
		neighbor.Description = lldpInfo.SysDescription

		// Parse capabilities from the struct
		neighbor.Capabilities = parseLLDPCapabilitiesStruct(lldpInfo.SysCapabilities.EnabledCap)

		// Parse management address
		if len(lldpInfo.MgmtAddress.Address) > 0 {
			neighbor.ManagementIP = parseLLDPMgmtAddress(lldpInfo.MgmtAddress)
		}

		// Parse organization-specific TLVs for location
		for _, orgTLV := range lldpInfo.OrgTLVs {
			// Check for LLDP-MED location TLV
			if orgTLV.OUI == 0x0012bb && orgTLV.SubType == 3 {
				neighbor.Location = parseLLDPLocation(orgTLV.Info)
			}
		}
	}

	// Use source MAC as ID if chassis ID parsing failed
	if neighbor.ID == "" && neighbor.SourceMAC != nil {
		neighbor.ID = neighbor.SourceMAC.String()
	}

	return neighbor, nil
}

// parseLLDPChassisID parses the chassis ID TLV
func parseLLDPChassisID(chassisID layers.LLDPChassisID) string {
	switch chassisID.Subtype {
	case layers.LLDPChassisIDSubTypeMACAddr:
		if len(chassisID.ID) == 6 {
			mac := net.HardwareAddr(chassisID.ID)
			return mac.String()
		}
		return fmt.Sprintf("%x", chassisID.ID)

	case layers.LLDPChassisIDSubTypeNetworkAddr:
		// First byte is address family
		if len(chassisID.ID) >= 5 && chassisID.ID[0] == 1 {
			// IPv4
			return net.IP(chassisID.ID[1:5]).String()
		}
		if len(chassisID.ID) >= 17 && chassisID.ID[0] == 2 {
			// IPv6
			return net.IP(chassisID.ID[1:17]).String()
		}
		return fmt.Sprintf("%x", chassisID.ID)

	case layers.LLDPChassisIDSubTypeLocal,
		layers.LLDPChassisIDSubTypeChassisComp,
		layers.LLDPChassisIDSubtypeIfaceName,
		layers.LLDPChassisIDSubtypeIfaceAlias:
		return protocol.CleanString(string(chassisID.ID))

	default:
		return protocol.CleanString(string(chassisID.ID))
	}
}

// parseLLDPPortID parses the port ID TLV
func parseLLDPPortID(portID layers.LLDPPortID) string {
	switch portID.Subtype {
	case layers.LLDPPortIDSubtypeMACAddr:
		if len(portID.ID) == 6 {
			mac := net.HardwareAddr(portID.ID)
			return mac.String()
		}
		return fmt.Sprintf("%x", portID.ID)

	case layers.LLDPPortIDSubtypeNetworkAddr:
		// First byte is address family
		if len(portID.ID) >= 5 && portID.ID[0] == 1 {
			return net.IP(portID.ID[1:5]).String()
		}
		if len(portID.ID) >= 17 && portID.ID[0] == 2 {
			return net.IP(portID.ID[1:17]).String()
		}
		return fmt.Sprintf("%x", portID.ID)

	case layers.LLDPPortIDSubtypeLocal,
		layers.LLDPPortIDSubtypeIfaceName,
		layers.LLDPPortIDSubtypeIfaceAlias,
		layers.LLDPPortIDSubtypeAgentCircuitID:
		return protocol.CleanString(string(portID.ID))

	default:
		return protocol.CleanString(string(portID.ID))
	}
}

// parseLLDPCapabilitiesStruct parses the LLDP capabilities struct
func parseLLDPCapabilitiesStruct(caps layers.LLDPCapabilities) []types.Capability {
	var result []types.Capability

	if caps.Router {
		result = append(result, types.CapRouter)
	}
	if caps.Bridge {
		result = append(result, types.CapBridge)
	}
	if caps.WLANAP {
		result = append(result, types.CapAccessPoint)
	}
	if caps.Phone {
		result = append(result, types.CapPhone)
	}
	if caps.DocSis {
		result = append(result, types.CapDocsis)
	}
	if caps.StationOnly {
		result = append(result, types.CapStation)
	}
	if caps.Repeater {
		result = append(result, types.CapRepeater)
	}
	if caps.Other && len(result) == 0 {
		result = append(result, types.CapOther)
	}

	// If no capabilities were set but the device responded, assume it's a switch
	if len(result) == 0 {
		result = append(result, types.CapSwitch)
	}

	return result
}

// parseLLDPMgmtAddress parses the management address TLV
func parseLLDPMgmtAddress(mgmtAddr layers.LLDPMgmtAddress) net.IP {
	if len(mgmtAddr.Address) == 0 {
		return nil
	}

	switch mgmtAddr.Subtype {
	case layers.IANAAddressFamilyIPV4:
		if len(mgmtAddr.Address) >= 4 {
			return net.IP(mgmtAddr.Address[:4])
		}
	case layers.IANAAddressFamilyIPV6:
		if len(mgmtAddr.Address) >= 16 {
			return net.IP(mgmtAddr.Address[:16])
		}
	default:
		// Try to interpret as IPv4 if it looks like one
		if len(mgmtAddr.Address) == 4 {
			return net.IP(mgmtAddr.Address)
		}
	}

	return nil
}

// parseLLDPLocation parses LLDP-MED location TLV
func parseLLDPLocation(data []byte) string {
	if len(data) < 1 {
		return ""
	}

	// Location data format type
	locType := data[0]

	switch locType {
	case 1: // Coordinate-based
		return "Coordinate-based location"
	case 2: // Civic address
		return parseCivicAddress(data[1:])
	case 3: // ECS ELIN
		if len(data) > 1 {
			return "ELIN: " + string(data[1:])
		}
	}

	return ""
}

// parseCivicAddress parses LLDP-MED civic address
func parseCivicAddress(data []byte) string {
	if len(data) < 2 {
		return ""
	}

	// Skip country code (2 bytes)
	offset := 2
	var parts []string

	for offset < len(data) {
		if offset+2 > len(data) {
			break
		}

		// caType := data[offset]
		caLen := int(data[offset+1])
		offset += 2

		if offset+caLen > len(data) {
			break
		}

		value := protocol.CleanString(string(data[offset : offset+caLen]))
		if value != "" {
			parts = append(parts, value)
		}
		offset += caLen
	}

	return strings.Join(parts, ", ")
}

// Helper to convert big endian bytes to uint16
func beUint16(b []byte) uint16 {
	if len(b) < 2 {
		return 0
	}
	return binary.BigEndian.Uint16(b)
}
