// Package protocol provides shared CDP and LLDP protocol constants and utilities.
package protocol

import "net"

// CDP TLV types
const (
	CDPTLVDeviceID     uint16 = 0x0001
	CDPTLVAddress      uint16 = 0x0002
	CDPTLVPortID       uint16 = 0x0003
	CDPTLVCapabilities uint16 = 0x0004
	CDPTLVVersion      uint16 = 0x0005
	CDPTLVPlatform     uint16 = 0x0006
	CDPTLVIPPrefix     uint16 = 0x0007
	CDPTLVHello        uint16 = 0x0008
	CDPTLVVTPDomain    uint16 = 0x0009
	CDPTLVNativeVLAN   uint16 = 0x000a
	CDPTLVDuplex       uint16 = 0x000b
	CDPTLVLocation     uint16 = 0x000c
	CDPTLVMgmtAddress  uint16 = 0x0016
)

// CDP capability bits
const (
	CDPCapRouter       uint32 = 0x01
	CDPCapTransBridge  uint32 = 0x02
	CDPCapSourceBridge uint32 = 0x04
	CDPCapSwitch       uint32 = 0x08
	CDPCapHost         uint32 = 0x10
	CDPCapIGMP         uint32 = 0x20
	CDPCapRepeater     uint32 = 0x40
	CDPCapPhone        uint32 = 0x80
)

// CDPCapStation is an alias for CDPCapHost
const CDPCapStation = CDPCapHost

// LLDP TLV types
const (
	LLDPTLVEnd         uint8 = 0
	LLDPTLVChassisID   uint8 = 1
	LLDPTLVPortID      uint8 = 2
	LLDPTLVTTL         uint8 = 3
	LLDPTLVPortDesc    uint8 = 4
	LLDPTLVSystemName  uint8 = 5
	LLDPTLVSystemDesc  uint8 = 6
	LLDPTLVSystemCap   uint8 = 7
	LLDPTLVMgmtAddress uint8 = 8
)

// LLDP Chassis ID subtypes
const (
	LLDPChassisIDSubtypeMAC uint8 = 4
)

// LLDP Port ID subtypes
const (
	LLDPPortIDSubtypeIfaceName uint8 = 5
)

// LLDP capability bits
const (
	LLDPCapOther    uint16 = 0x0001
	LLDPCapRepeater uint16 = 0x0002
	LLDPCapBridge   uint16 = 0x0004
	LLDPCapWLANAP   uint16 = 0x0008
	LLDPCapRouter   uint16 = 0x0010
	LLDPCapPhone    uint16 = 0x0020
	LLDPCapDocsis   uint16 = 0x0040
	LLDPCapStation  uint16 = 0x0080
)

// LLDP EtherType
const LLDPEtherType uint16 = 0x88CC

// Multicast MAC addresses
var (
	CDPMulticastMAC  = net.HardwareAddr{0x01, 0x00, 0x0c, 0xcc, 0xcc, 0xcc}
	LLDPMulticastMAC = net.HardwareAddr{0x01, 0x80, 0xc2, 0x00, 0x00, 0x0e}
)
