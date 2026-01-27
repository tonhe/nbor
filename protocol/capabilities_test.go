package protocol

import (
	"encoding/binary"
	"testing"

	"nbor/types"
)

func TestParseCDPCapabilities(t *testing.T) {
	tests := []struct {
		name string
		bits uint32
		want []types.Capability
	}{
		{
			name: "router only",
			bits: CDPCapRouter,
			want: []types.Capability{types.CapRouter},
		},
		{
			name: "switch only",
			bits: CDPCapSwitch,
			want: []types.Capability{types.CapSwitch},
		},
		{
			name: "router and switch",
			bits: CDPCapRouter | CDPCapSwitch,
			want: []types.Capability{types.CapRouter, types.CapSwitch},
		},
		{
			name: "trans bridge gives bridge",
			bits: CDPCapTransBridge,
			want: []types.Capability{types.CapBridge},
		},
		{
			name: "source bridge gives bridge",
			bits: CDPCapSourceBridge,
			want: []types.Capability{types.CapBridge},
		},
		{
			name: "both bridges give one bridge",
			bits: CDPCapTransBridge | CDPCapSourceBridge,
			want: []types.Capability{types.CapBridge},
		},
		{
			name: "station/host",
			bits: CDPCapHost,
			want: []types.Capability{types.CapStation},
		},
		{
			name: "phone",
			bits: CDPCapPhone,
			want: []types.Capability{types.CapPhone},
		},
		{
			name: "repeater",
			bits: CDPCapRepeater,
			want: []types.Capability{types.CapRepeater},
		},
		{
			name: "all capabilities",
			bits: CDPCapRouter | CDPCapSwitch | CDPCapTransBridge | CDPCapHost | CDPCapPhone | CDPCapRepeater,
			want: []types.Capability{types.CapRouter, types.CapBridge, types.CapSwitch, types.CapRepeater, types.CapPhone, types.CapStation},
		},
		{
			name: "no capabilities",
			bits: 0,
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]byte, 4)
			binary.BigEndian.PutUint32(data, tt.bits)
			got := ParseCDPCapabilities(data)

			if len(got) != len(tt.want) {
				t.Errorf("ParseCDPCapabilities() returned %d caps, want %d: got %v, want %v",
					len(got), len(tt.want), got, tt.want)
				return
			}

			// Check each capability (order matters based on bit order)
			for i, want := range tt.want {
				if got[i] != want {
					t.Errorf("ParseCDPCapabilities()[%d] = %q, want %q", i, got[i], want)
				}
			}
		})
	}
}

func TestParseCDPCapabilities_ShortData(t *testing.T) {
	// Test with data shorter than 4 bytes
	tests := [][]byte{
		nil,
		{},
		{0x00},
		{0x00, 0x00},
		{0x00, 0x00, 0x00},
	}

	for _, data := range tests {
		got := ParseCDPCapabilities(data)
		if got != nil {
			t.Errorf("ParseCDPCapabilities(%v) = %v, want nil", data, got)
		}
	}
}

func TestBuildCDPCapabilities(t *testing.T) {
	tests := []struct {
		name string
		caps []string
		want uint32
	}{
		{
			name: "router",
			caps: []string{"router"},
			want: CDPCapRouter,
		},
		{
			name: "switch",
			caps: []string{"switch"},
			want: CDPCapSwitch,
		},
		{
			name: "bridge",
			caps: []string{"bridge"},
			want: CDPCapTransBridge,
		},
		{
			name: "station",
			caps: []string{"station"},
			want: CDPCapStation,
		},
		{
			name: "host (alias for station)",
			caps: []string{"host"},
			want: CDPCapStation,
		},
		{
			name: "phone",
			caps: []string{"phone"},
			want: CDPCapPhone,
		},
		{
			name: "multiple capabilities",
			caps: []string{"router", "switch"},
			want: CDPCapRouter | CDPCapSwitch,
		},
		{
			name: "empty defaults to station",
			caps: []string{},
			want: CDPCapStation,
		},
		{
			name: "nil defaults to station",
			caps: nil,
			want: CDPCapStation,
		},
		{
			name: "unknown capabilities default to station",
			caps: []string{"unknown", "invalid"},
			want: CDPCapStation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildCDPCapabilities(tt.caps)
			if got != tt.want {
				t.Errorf("BuildCDPCapabilities(%v) = 0x%02x, want 0x%02x", tt.caps, got, tt.want)
			}
		})
	}
}

func TestBuildLLDPCapabilities(t *testing.T) {
	tests := []struct {
		name string
		caps []string
		want uint16
	}{
		{
			name: "router",
			caps: []string{"router"},
			want: LLDPCapRouter,
		},
		{
			name: "bridge",
			caps: []string{"bridge"},
			want: LLDPCapBridge,
		},
		{
			name: "switch maps to bridge",
			caps: []string{"switch"},
			want: LLDPCapBridge,
		},
		{
			name: "station",
			caps: []string{"station"},
			want: LLDPCapStation,
		},
		{
			name: "host (alias for station)",
			caps: []string{"host"},
			want: LLDPCapStation,
		},
		{
			name: "phone",
			caps: []string{"phone"},
			want: LLDPCapPhone,
		},
		{
			name: "ap",
			caps: []string{"ap"},
			want: LLDPCapWLANAP,
		},
		{
			name: "wlan",
			caps: []string{"wlan"},
			want: LLDPCapWLANAP,
		},
		{
			name: "repeater",
			caps: []string{"repeater"},
			want: LLDPCapRepeater,
		},
		{
			name: "multiple capabilities",
			caps: []string{"router", "bridge"},
			want: LLDPCapRouter | LLDPCapBridge,
		},
		{
			name: "empty defaults to station",
			caps: []string{},
			want: LLDPCapStation,
		},
		{
			name: "nil defaults to station",
			caps: nil,
			want: LLDPCapStation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildLLDPCapabilities(tt.caps)
			if got != tt.want {
				t.Errorf("BuildLLDPCapabilities(%v) = 0x%04x, want 0x%04x", tt.caps, got, tt.want)
			}
		})
	}
}
