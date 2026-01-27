package protocol

import (
	"bytes"
	"testing"
)

func TestCDPTLVTypes(t *testing.T) {
	// Verify CDP TLV type values match Cisco specification
	tests := []struct {
		name  string
		value uint16
		want  uint16
	}{
		{"CDPTLVDeviceID", CDPTLVDeviceID, 0x0001},
		{"CDPTLVAddress", CDPTLVAddress, 0x0002},
		{"CDPTLVPortID", CDPTLVPortID, 0x0003},
		{"CDPTLVCapabilities", CDPTLVCapabilities, 0x0004},
		{"CDPTLVVersion", CDPTLVVersion, 0x0005},
		{"CDPTLVPlatform", CDPTLVPlatform, 0x0006},
		{"CDPTLVNativeVLAN", CDPTLVNativeVLAN, 0x000a},
		{"CDPTLVDuplex", CDPTLVDuplex, 0x000b},
		{"CDPTLVLocation", CDPTLVLocation, 0x000c},
		{"CDPTLVMgmtAddress", CDPTLVMgmtAddress, 0x0016},
	}

	for _, tt := range tests {
		if tt.value != tt.want {
			t.Errorf("%s = 0x%04x, want 0x%04x", tt.name, tt.value, tt.want)
		}
	}
}

func TestCDPCapabilityBits(t *testing.T) {
	// Verify CDP capability bits match Cisco specification
	tests := []struct {
		name  string
		value uint32
		want  uint32
	}{
		{"CDPCapRouter", CDPCapRouter, 0x01},
		{"CDPCapTransBridge", CDPCapTransBridge, 0x02},
		{"CDPCapSourceBridge", CDPCapSourceBridge, 0x04},
		{"CDPCapSwitch", CDPCapSwitch, 0x08},
		{"CDPCapHost", CDPCapHost, 0x10},
		{"CDPCapStation", CDPCapStation, 0x10}, // alias
		{"CDPCapIGMP", CDPCapIGMP, 0x20},
		{"CDPCapRepeater", CDPCapRepeater, 0x40},
		{"CDPCapPhone", CDPCapPhone, 0x80},
	}

	for _, tt := range tests {
		if tt.value != tt.want {
			t.Errorf("%s = 0x%02x, want 0x%02x", tt.name, tt.value, tt.want)
		}
	}
}

func TestLLDPTLVTypes(t *testing.T) {
	// Verify LLDP TLV types match IEEE 802.1AB specification
	tests := []struct {
		name  string
		value uint8
		want  uint8
	}{
		{"LLDPTLVEnd", LLDPTLVEnd, 0},
		{"LLDPTLVChassisID", LLDPTLVChassisID, 1},
		{"LLDPTLVPortID", LLDPTLVPortID, 2},
		{"LLDPTLVTTL", LLDPTLVTTL, 3},
		{"LLDPTLVPortDesc", LLDPTLVPortDesc, 4},
		{"LLDPTLVSystemName", LLDPTLVSystemName, 5},
		{"LLDPTLVSystemDesc", LLDPTLVSystemDesc, 6},
		{"LLDPTLVSystemCap", LLDPTLVSystemCap, 7},
		{"LLDPTLVMgmtAddress", LLDPTLVMgmtAddress, 8},
	}

	for _, tt := range tests {
		if tt.value != tt.want {
			t.Errorf("%s = %d, want %d", tt.name, tt.value, tt.want)
		}
	}
}

func TestLLDPCapabilityBits(t *testing.T) {
	// Verify LLDP capability bits match IEEE 802.1AB specification
	tests := []struct {
		name  string
		value uint16
		want  uint16
	}{
		{"LLDPCapOther", LLDPCapOther, 0x0001},
		{"LLDPCapRepeater", LLDPCapRepeater, 0x0002},
		{"LLDPCapBridge", LLDPCapBridge, 0x0004},
		{"LLDPCapWLANAP", LLDPCapWLANAP, 0x0008},
		{"LLDPCapRouter", LLDPCapRouter, 0x0010},
		{"LLDPCapPhone", LLDPCapPhone, 0x0020},
		{"LLDPCapDocsis", LLDPCapDocsis, 0x0040},
		{"LLDPCapStation", LLDPCapStation, 0x0080},
	}

	for _, tt := range tests {
		if tt.value != tt.want {
			t.Errorf("%s = 0x%04x, want 0x%04x", tt.name, tt.value, tt.want)
		}
	}
}

func TestMulticastAddresses(t *testing.T) {
	// Verify multicast MAC addresses
	expectedCDP := []byte{0x01, 0x00, 0x0c, 0xcc, 0xcc, 0xcc}
	if !bytes.Equal(CDPMulticastMAC, expectedCDP) {
		t.Errorf("CDPMulticastMAC = %v, want %v", CDPMulticastMAC, expectedCDP)
	}

	expectedLLDP := []byte{0x01, 0x80, 0xc2, 0x00, 0x00, 0x0e}
	if !bytes.Equal(LLDPMulticastMAC, expectedLLDP) {
		t.Errorf("LLDPMulticastMAC = %v, want %v", LLDPMulticastMAC, expectedLLDP)
	}
}

func TestLLDPEtherType(t *testing.T) {
	if LLDPEtherType != 0x88CC {
		t.Errorf("LLDPEtherType = 0x%04x, want 0x88CC", LLDPEtherType)
	}
}
