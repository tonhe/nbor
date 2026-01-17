package types

import (
	"net"
	"strings"
	"sync"
	"time"
)

// Protocol represents the discovery protocol type
type Protocol string

const (
	ProtocolCDP  Protocol = "CDP"
	ProtocolLLDP Protocol = "LLDP"
	ProtocolBoth Protocol = "CDP+LLDP"
)

// Capability represents device capabilities
type Capability string

const (
	CapRouter      Capability = "Router"
	CapSwitch      Capability = "Switch"
	CapBridge      Capability = "Bridge"
	CapAccessPoint Capability = "AP"
	CapPhone       Capability = "Phone"
	CapDocsis      Capability = "DOCSIS"
	CapStation     Capability = "Station"
	CapRepeater    Capability = "Repeater"
	CapOther       Capability = "Other"
)

// Neighbor represents a discovered network neighbor
type Neighbor struct {
	// Unique identifier (typically chassis ID or device ID)
	ID string

	// Device hostname/system name
	Hostname string

	// Port ID - the port we're connected to on the neighbor
	PortID string

	// Port description
	PortDescription string

	// Management IP address
	ManagementIP net.IP

	// Platform/model information
	Platform string

	// System description
	Description string

	// SNMP Location
	Location string

	// Device capabilities
	Capabilities []Capability

	// Discovery protocol(s) used - can be CDP, LLDP, or CDP+LLDP
	Protocol Protocol

	// Track which protocols we've seen
	SeenCDP  bool
	SeenLLDP bool

	// First time this neighbor was seen
	FirstSeen time.Time

	// Last time this neighbor announced itself
	LastSeen time.Time

	// Whether this neighbor is considered stale
	IsStale bool

	// Whether this is a newly discovered neighbor (for highlighting)
	IsNew bool

	// Source MAC address of the neighbor
	SourceMAC net.HardwareAddr

	// The interface this neighbor was seen on
	Interface string
}

// NeighborKey generates a unique key for this neighbor
// We key by source MAC since that identifies the physical port sending to us
// CDP and LLDP from the same physical port will have the same source MAC
func (n *Neighbor) NeighborKey() string {
	// Source MAC is the most reliable key - it's the actual MAC sending the packet
	// Both CDP and LLDP from the same port should have the same source MAC
	if n.SourceMAC != nil {
		return n.Interface + ":" + n.SourceMAC.String()
	}
	// Fallback to device ID
	if n.ID != "" {
		return n.Interface + ":" + strings.ToLower(n.ID)
	}
	return n.Interface + ":unknown"
}

// UpdateProtocol updates the protocol field based on what we've seen
func (n *Neighbor) UpdateProtocol() {
	if n.SeenCDP && n.SeenLLDP {
		n.Protocol = ProtocolBoth
	} else if n.SeenCDP {
		n.Protocol = ProtocolCDP
	} else if n.SeenLLDP {
		n.Protocol = ProtocolLLDP
	}
}

// NeighborStore manages discovered neighbors with thread-safe access
type NeighborStore struct {
	mu        sync.RWMutex
	neighbors map[string]*Neighbor
	// Callback for when a new neighbor is discovered
	OnNewNeighbor func(*Neighbor)
	// Callback for when a neighbor is updated
	OnUpdate func(*Neighbor)
}

// NewNeighborStore creates a new neighbor store
func NewNeighborStore() *NeighborStore {
	return &NeighborStore{
		neighbors: make(map[string]*Neighbor),
	}
}

// Update adds or updates a neighbor in the store
// Returns true if this is a new neighbor
func (s *NeighborStore) Update(n *Neighbor) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := n.NeighborKey()
	existing, exists := s.neighbors[key]

	if exists {
		// Update existing neighbor - merge information
		// Prefer non-empty values (CDP often has more detail than LLDP or vice versa)
		if n.Hostname != "" {
			existing.Hostname = n.Hostname
		}
		if n.PortID != "" {
			existing.PortID = n.PortID
		}
		if n.PortDescription != "" {
			existing.PortDescription = n.PortDescription
		}
		if n.ManagementIP != nil {
			existing.ManagementIP = n.ManagementIP
		}
		if n.Platform != "" {
			existing.Platform = n.Platform
		}
		if n.Description != "" {
			existing.Description = n.Description
		}
		if n.Location != "" {
			existing.Location = n.Location
		}
		if len(n.Capabilities) > 0 {
			existing.Capabilities = mergeCapabilities(existing.Capabilities, n.Capabilities)
		}

		// Track which protocols we've seen
		if n.Protocol == ProtocolCDP {
			existing.SeenCDP = true
		} else if n.Protocol == ProtocolLLDP {
			existing.SeenLLDP = true
		}
		existing.UpdateProtocol()

		existing.LastSeen = n.LastSeen
		existing.IsStale = false
		existing.SourceMAC = n.SourceMAC

		if s.OnUpdate != nil {
			s.OnUpdate(existing)
		}
		return false
	}

	// New neighbor
	n.FirstSeen = n.LastSeen
	n.IsNew = true
	n.IsStale = false

	// Set initial protocol flags
	if n.Protocol == ProtocolCDP {
		n.SeenCDP = true
	} else if n.Protocol == ProtocolLLDP {
		n.SeenLLDP = true
	}

	s.neighbors[key] = n

	if s.OnNewNeighbor != nil {
		s.OnNewNeighbor(n)
	}
	return true
}

// mergeCapabilities merges two capability lists, removing duplicates
func mergeCapabilities(existing, new []Capability) []Capability {
	seen := make(map[Capability]bool)
	for _, c := range existing {
		seen[c] = true
	}
	for _, c := range new {
		seen[c] = true
	}

	result := make([]Capability, 0, len(seen))
	for c := range seen {
		result = append(result, c)
	}
	return result
}

// GetAll returns all neighbors
func (s *NeighborStore) GetAll() []*Neighbor {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Neighbor, 0, len(s.neighbors))
	for _, n := range s.neighbors {
		result = append(result, n)
	}
	return result
}

// GetByInterface returns neighbors for a specific interface
func (s *NeighborStore) GetByInterface(iface string) []*Neighbor {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Neighbor
	for _, n := range s.neighbors {
		if n.Interface == iface {
			result = append(result, n)
		}
	}
	return result
}

// MarkStale marks neighbors that haven't been seen recently as stale
func (s *NeighborStore) MarkStale(threshold time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for _, n := range s.neighbors {
		if now.Sub(n.LastSeen) > threshold {
			n.IsStale = true
		}
	}
}

// ClearNewFlags clears the IsNew flag on all neighbors
func (s *NeighborStore) ClearNewFlags() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, n := range s.neighbors {
		n.IsNew = false
	}
}

// Clear removes all neighbors
func (s *NeighborStore) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.neighbors = make(map[string]*Neighbor)
}

// Count returns the number of neighbors
func (s *NeighborStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.neighbors)
}

// InterfaceInfo holds information about a network interface
type InterfaceInfo struct {
	Name      string
	MAC       net.HardwareAddr
	IsUp      bool
	Speed     string // Link speed if available
	MTU       int
	IPv4Addrs []net.IP // IPv4 addresses assigned to this interface
	IPv6Addrs []net.IP // IPv6 addresses (excluding link-local fe80::)
}

// String returns a display string for the interface
func (i *InterfaceInfo) String() string {
	status := "down"
	if i.IsUp {
		status = "up"
	}
	return i.Name + " (" + status + ")"
}

// FormatIPs returns a formatted string of IP addresses
func (i *InterfaceInfo) FormatIPs() string {
	var ips []string
	for _, ip := range i.IPv4Addrs {
		ips = append(ips, ip.String())
	}
	for _, ip := range i.IPv6Addrs {
		ips = append(ips, ip.String())
	}
	if len(ips) == 0 {
		return ""
	}
	return strings.Join(ips, ", ")
}

// GetInterfaceAddresses returns non-link-local IP addresses for an interface
func GetInterfaceAddresses(iface *net.Interface) ([]net.IP, []net.IP) {
	var ipv4Addrs, ipv6Addrs []net.IP

	addrs, err := iface.Addrs()
	if err != nil {
		return nil, nil
	}

	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}

		if ip == nil {
			continue
		}

		// Check if IPv4
		if ip4 := ip.To4(); ip4 != nil {
			// Skip loopback
			if !ip4.IsLoopback() {
				ipv4Addrs = append(ipv4Addrs, ip4)
			}
		} else {
			// IPv6
			// Skip link-local (fe80::)
			if !ip.IsLinkLocalUnicast() && !ip.IsLoopback() {
				ipv6Addrs = append(ipv6Addrs, ip)
			}
		}
	}

	return ipv4Addrs, ipv6Addrs
}
