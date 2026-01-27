package types

import (
	"net"
	"testing"
	"time"
)

func TestNeighborKey(t *testing.T) {
	mac, _ := net.ParseMAC("00:11:22:33:44:55")

	tests := []struct {
		name     string
		neighbor *Neighbor
		want     string
	}{
		{
			name: "with source MAC",
			neighbor: &Neighbor{
				Interface: "eth0",
				SourceMAC: mac,
				ID:        "switch01",
			},
			want: "eth0:00:11:22:33:44:55",
		},
		{
			name: "without MAC, uses ID",
			neighbor: &Neighbor{
				Interface: "eth0",
				ID:        "Switch01", // uppercase to test lowercase conversion
			},
			want: "eth0:switch01",
		},
		{
			name: "without MAC or ID",
			neighbor: &Neighbor{
				Interface: "eth0",
			},
			want: "eth0:unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.neighbor.NeighborKey()
			if got != tt.want {
				t.Errorf("NeighborKey() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestUpdateProtocol(t *testing.T) {
	tests := []struct {
		name     string
		seenCDP  bool
		seenLLDP bool
		want     Protocol
	}{
		{
			name:     "CDP only",
			seenCDP:  true,
			seenLLDP: false,
			want:     ProtocolCDP,
		},
		{
			name:     "LLDP only",
			seenCDP:  false,
			seenLLDP: true,
			want:     ProtocolLLDP,
		},
		{
			name:     "both protocols",
			seenCDP:  true,
			seenLLDP: true,
			want:     ProtocolBoth,
		},
		{
			name:     "neither (should be empty)",
			seenCDP:  false,
			seenLLDP: false,
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &Neighbor{
				SeenCDP:  tt.seenCDP,
				SeenLLDP: tt.seenLLDP,
			}
			n.UpdateProtocol()
			if n.Protocol != tt.want {
				t.Errorf("UpdateProtocol() set Protocol = %q, want %q", n.Protocol, tt.want)
			}
		})
	}
}

func TestNeighborStoreUpdate(t *testing.T) {
	store := NewNeighborStore()
	mac, _ := net.ParseMAC("00:11:22:33:44:55")

	// Add new neighbor
	n1 := &Neighbor{
		Interface: "eth0",
		SourceMAC: mac,
		Hostname:  "switch01",
		Protocol:  ProtocolCDP,
		LastSeen:  time.Now(),
	}

	isNew := store.Update(n1)
	if !isNew {
		t.Error("Update() returned false for new neighbor, want true")
	}
	if store.Count() != 1 {
		t.Errorf("Count() = %d, want 1", store.Count())
	}

	// Update existing neighbor
	n2 := &Neighbor{
		Interface: "eth0",
		SourceMAC: mac,
		PortID:    "Gi0/1",
		Protocol:  ProtocolLLDP,
		LastSeen:  time.Now(),
	}

	isNew = store.Update(n2)
	if isNew {
		t.Error("Update() returned true for existing neighbor, want false")
	}
	if store.Count() != 1 {
		t.Errorf("Count() = %d after update, want 1", store.Count())
	}

	// Verify merged data
	neighbors := store.GetAll()
	if len(neighbors) != 1 {
		t.Fatalf("GetAll() returned %d neighbors, want 1", len(neighbors))
	}
	neighbor := neighbors[0]
	if neighbor.Hostname != "switch01" {
		t.Errorf("Hostname = %q, want %q", neighbor.Hostname, "switch01")
	}
	if neighbor.PortID != "Gi0/1" {
		t.Errorf("PortID = %q, want %q", neighbor.PortID, "Gi0/1")
	}
	if neighbor.Protocol != ProtocolBoth {
		t.Errorf("Protocol = %q, want %q", neighbor.Protocol, ProtocolBoth)
	}
}

func TestNeighborStoreMarkStale(t *testing.T) {
	store := NewNeighborStore()
	mac, _ := net.ParseMAC("00:11:22:33:44:55")

	n := &Neighbor{
		Interface: "eth0",
		SourceMAC: mac,
		LastSeen:  time.Now().Add(-2 * time.Minute),
	}
	store.Update(n)

	// Not stale yet (threshold 3 minutes)
	store.MarkStale(3 * time.Minute)
	neighbors := store.GetAll()
	if neighbors[0].IsStale {
		t.Error("Neighbor marked stale before threshold")
	}

	// Now stale (threshold 1 minute)
	store.MarkStale(1 * time.Minute)
	neighbors = store.GetAll()
	if !neighbors[0].IsStale {
		t.Error("Neighbor not marked stale after threshold")
	}
}

func TestNeighborStoreRemoveStale(t *testing.T) {
	store := NewNeighborStore()
	mac1, _ := net.ParseMAC("00:11:22:33:44:55")
	mac2, _ := net.ParseMAC("00:11:22:33:44:66")

	// Add two neighbors - one old, one recent
	n1 := &Neighbor{
		Interface: "eth0",
		SourceMAC: mac1,
		LastSeen:  time.Now().Add(-5 * time.Minute),
	}
	n2 := &Neighbor{
		Interface: "eth0",
		SourceMAC: mac2,
		LastSeen:  time.Now(),
	}
	store.Update(n1)
	store.Update(n2)

	// Mark old one as stale
	store.MarkStale(1 * time.Minute)

	// Remove stale neighbors older than 2 minutes
	removed := store.RemoveStale(2 * time.Minute)
	if removed != 1 {
		t.Errorf("RemoveStale() removed %d, want 1", removed)
	}
	if store.Count() != 1 {
		t.Errorf("Count() after RemoveStale() = %d, want 1", store.Count())
	}
}

func TestNeighborStoreClear(t *testing.T) {
	store := NewNeighborStore()
	mac, _ := net.ParseMAC("00:11:22:33:44:55")

	n := &Neighbor{
		Interface: "eth0",
		SourceMAC: mac,
		LastSeen:  time.Now(),
	}
	store.Update(n)

	if store.Count() != 1 {
		t.Fatalf("Count() before Clear() = %d, want 1", store.Count())
	}

	store.Clear()
	if store.Count() != 0 {
		t.Errorf("Count() after Clear() = %d, want 0", store.Count())
	}
}

func TestNeighborStoreGetByInterface(t *testing.T) {
	store := NewNeighborStore()
	mac1, _ := net.ParseMAC("00:11:22:33:44:55")
	mac2, _ := net.ParseMAC("00:11:22:33:44:66")

	n1 := &Neighbor{
		Interface: "eth0",
		SourceMAC: mac1,
		LastSeen:  time.Now(),
	}
	n2 := &Neighbor{
		Interface: "eth1",
		SourceMAC: mac2,
		LastSeen:  time.Now(),
	}
	store.Update(n1)
	store.Update(n2)

	eth0Neighbors := store.GetByInterface("eth0")
	if len(eth0Neighbors) != 1 {
		t.Errorf("GetByInterface(eth0) returned %d neighbors, want 1", len(eth0Neighbors))
	}

	eth1Neighbors := store.GetByInterface("eth1")
	if len(eth1Neighbors) != 1 {
		t.Errorf("GetByInterface(eth1) returned %d neighbors, want 1", len(eth1Neighbors))
	}

	eth2Neighbors := store.GetByInterface("eth2")
	if len(eth2Neighbors) != 0 {
		t.Errorf("GetByInterface(eth2) returned %d neighbors, want 0", len(eth2Neighbors))
	}
}

func TestInterfaceInfoString(t *testing.T) {
	tests := []struct {
		name  string
		info  InterfaceInfo
		want  string
	}{
		{
			name: "interface up",
			info: InterfaceInfo{Name: "eth0", IsUp: true},
			want: "eth0 (up)",
		},
		{
			name: "interface down",
			info: InterfaceInfo{Name: "eth1", IsUp: false},
			want: "eth1 (down)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.info.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestInterfaceInfoFormatIPs(t *testing.T) {
	tests := []struct {
		name string
		info InterfaceInfo
		want string
	}{
		{
			name: "no IPs",
			info: InterfaceInfo{},
			want: "",
		},
		{
			name: "IPv4 only",
			info: InterfaceInfo{
				IPv4Addrs: []net.IP{net.ParseIP("192.168.1.1")},
			},
			want: "192.168.1.1",
		},
		{
			name: "IPv4 and IPv6",
			info: InterfaceInfo{
				IPv4Addrs: []net.IP{net.ParseIP("192.168.1.1")},
				IPv6Addrs: []net.IP{net.ParseIP("2001:db8::1")},
			},
			want: "192.168.1.1, 2001:db8::1",
		},
		{
			name: "multiple IPs",
			info: InterfaceInfo{
				IPv4Addrs: []net.IP{net.ParseIP("192.168.1.1"), net.ParseIP("10.0.0.1")},
			},
			want: "192.168.1.1, 10.0.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.info.FormatIPs()
			if got != tt.want {
				t.Errorf("FormatIPs() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCapabilityConstants(t *testing.T) {
	// Verify capability constants have expected values
	tests := []struct {
		cap  Capability
		want string
	}{
		{CapRouter, "Router"},
		{CapSwitch, "Switch"},
		{CapBridge, "Bridge"},
		{CapAccessPoint, "AP"},
		{CapPhone, "Phone"},
		{CapDocsis, "DOCSIS"},
		{CapStation, "Station"},
		{CapRepeater, "Repeater"},
		{CapOther, "Other"},
	}

	for _, tt := range tests {
		if string(tt.cap) != tt.want {
			t.Errorf("Capability %v = %q, want %q", tt.cap, string(tt.cap), tt.want)
		}
	}
}

func TestProtocolConstants(t *testing.T) {
	// Verify protocol constants
	if ProtocolCDP != "CDP" {
		t.Errorf("ProtocolCDP = %q, want %q", ProtocolCDP, "CDP")
	}
	if ProtocolLLDP != "LLDP" {
		t.Errorf("ProtocolLLDP = %q, want %q", ProtocolLLDP, "LLDP")
	}
	if ProtocolBoth != "CDP+LLDP" {
		t.Errorf("ProtocolBoth = %q, want %q", ProtocolBoth, "CDP+LLDP")
	}
}
