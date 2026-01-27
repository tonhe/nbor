package protocol

import (
	"encoding/binary"

	"nbor/types"
)

// ParseCDPCapabilities converts CDP capability bits to a Capability slice
func ParseCDPCapabilities(data []byte) []types.Capability {
	if len(data) < 4 {
		return nil
	}

	caps := binary.BigEndian.Uint32(data)
	var result []types.Capability

	if caps&CDPCapRouter != 0 {
		result = append(result, types.CapRouter)
	}
	if caps&CDPCapTransBridge != 0 || caps&CDPCapSourceBridge != 0 {
		result = append(result, types.CapBridge)
	}
	if caps&CDPCapSwitch != 0 {
		result = append(result, types.CapSwitch)
	}
	if caps&CDPCapRepeater != 0 {
		result = append(result, types.CapRepeater)
	}
	if caps&CDPCapPhone != 0 {
		result = append(result, types.CapPhone)
	}
	if caps&CDPCapHost != 0 {
		result = append(result, types.CapStation)
	}

	return result
}

// BuildCDPCapabilities converts capability strings to CDP capability bits
func BuildCDPCapabilities(caps []string) uint32 {
	var bits uint32
	for _, cap := range caps {
		switch cap {
		case "router":
			bits |= CDPCapRouter
		case "bridge":
			bits |= CDPCapTransBridge
		case "switch":
			bits |= CDPCapSwitch
		case "station", "host":
			bits |= CDPCapStation
		case "phone":
			bits |= CDPCapPhone
		}
	}
	// Default to station if nothing set
	if bits == 0 {
		bits = CDPCapStation
	}
	return bits
}

// BuildLLDPCapabilities converts capability strings to LLDP capability bits
func BuildLLDPCapabilities(caps []string) uint16 {
	var bits uint16
	for _, cap := range caps {
		switch cap {
		case "router":
			bits |= LLDPCapRouter
		case "bridge":
			bits |= LLDPCapBridge
		case "switch":
			bits |= LLDPCapBridge // LLDP uses bridge for switches
		case "station", "host":
			bits |= LLDPCapStation
		case "phone":
			bits |= LLDPCapPhone
		case "ap", "wlan":
			bits |= LLDPCapWLANAP
		case "repeater":
			bits |= LLDPCapRepeater
		}
	}
	// Default to station if nothing set
	if bits == 0 {
		bits = LLDPCapStation
	}
	return bits
}
