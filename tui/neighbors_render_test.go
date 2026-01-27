package tui

import (
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"nbor/config"
	"nbor/types"
)

func TestRenderDetailView(t *testing.T) {
	// Create a minimal model for testing
	store := types.NewNeighborStore()
	cfg := config.DefaultConfig()

	mac, _ := net.ParseMAC("00:11:22:33:44:55")
	mgmtIP := net.ParseIP("192.168.1.1")

	neighbor := &types.Neighbor{
		ID:            "switch01",
		Hostname:      "switch01.local",
		PortID:        "Gi0/1",
		ManagementIP:  mgmtIP,
		Platform:      "Cisco IOS",
		Description:   "Test switch",
		Protocol:      types.ProtocolCDP,
		SourceMAC:     mac,
		Interface:     "eth0",
		FirstSeen:     time.Now(),
		LastSeen:      time.Now(),
		Capabilities:  []types.Capability{types.CapSwitch},
	}
	store.Update(neighbor)

	m := NewNeighborTable(store, types.InterfaceInfo{Name: "eth0"}, "", &cfg)
	m.width = 80
	m.height = 30
	m.showDetail = true
	m.selectedIndex = 0

	// Get the selected neighbor
	n := m.getSelectedNeighbor()
	if n == nil {
		t.Fatal("getSelectedNeighbor returned nil")
	}

	// Test renderDetailView
	output := m.renderDetailView(n)

	// Count lines
	lines := strings.Split(output, "\n")
	lineCount := len(lines)

	// Debug output
	fmt.Printf("=== Debug Output ===\n")
	fmt.Printf("Terminal height (m.height): %d\n", m.height)
	fmt.Printf("Actual line count: %d\n", lineCount)
	fmt.Printf("Expected line count: %d\n", m.height)
	fmt.Printf("\n=== Line breakdown ===\n")

	// Show first few and last few lines
	for i, line := range lines {
		if i < 3 || i >= len(lines)-3 {
			// Truncate long lines for display
			displayLine := line
			if len(displayLine) > 60 {
				displayLine = displayLine[:60] + "..."
			}
			fmt.Printf("Line %2d: %q\n", i+1, displayLine)
		} else if i == 3 {
			fmt.Printf("... (%d lines omitted) ...\n", len(lines)-6)
		}
	}

	// Check individual components
	fmt.Printf("\n=== Component Analysis ===\n")

	header := m.renderHeader()
	headerLines := strings.Count(header, "\n") + 1
	fmt.Printf("Header lines: %d\n", headerLines)

	contentHeight := m.height - 2
	popup := m.renderDetailPopup(n, contentHeight)
	popupNewlines := strings.Count(popup, "\n")
	fmt.Printf("Popup newlines: %d\n", popupNewlines)
	fmt.Printf("Popup ends with newline: %v\n", strings.HasSuffix(popup, "\n"))

	footer := m.renderFooter()
	footerLines := strings.Count(footer, "\n") + 1
	fmt.Printf("Footer lines: %d\n", footerLines)

	// The test
	if lineCount != m.height {
		t.Errorf("Line count mismatch: got %d, want %d", lineCount, m.height)
	}
}

func TestRenderDetailViewVariousHeights(t *testing.T) {
	store := types.NewNeighborStore()
	cfg := config.DefaultConfig()

	mac, _ := net.ParseMAC("00:11:22:33:44:55")
	neighbor := &types.Neighbor{
		ID:        "switch01",
		Hostname:  "switch01.local",
		SourceMAC: mac,
		Interface: "eth0",
		FirstSeen: time.Now(),
		LastSeen:  time.Now(),
	}
	store.Update(neighbor)

	// Test various terminal heights
	heights := []int{20, 24, 30, 40, 50}

	for _, h := range heights {
		m := NewNeighborTable(store, types.InterfaceInfo{Name: "eth0"}, "", &cfg)
		m.width = 80
		m.height = h
		m.showDetail = true
		m.selectedIndex = 0

		n := m.getSelectedNeighbor()
		if n == nil {
			t.Fatalf("height=%d: getSelectedNeighbor returned nil", h)
		}

		output := m.renderDetailView(n)
		lines := strings.Split(output, "\n")
		lineCount := len(lines)

		// Check if footer is on last line
		lastLine := lines[len(lines)-1]
		hasFooter := strings.Contains(lastLine, "refresh") || strings.Contains(lastLine, "quit")

		fmt.Printf("Height %d: lines=%d, hasFooter=%v, lastLine=%q\n",
			h, lineCount, hasFooter, truncateStr(lastLine, 50))

		if lineCount != h {
			t.Errorf("height=%d: got %d lines, want %d", h, lineCount, h)
		}
		if !hasFooter {
			t.Errorf("height=%d: footer not found on last line", h)
		}
	}
}

func truncateStr(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}

func TestRenderDetailViewTooSmall(t *testing.T) {
	store := types.NewNeighborStore()
	cfg := config.DefaultConfig()

	mac, _ := net.ParseMAC("00:11:22:33:44:55")
	neighbor := &types.Neighbor{
		ID:        "switch01",
		Hostname:  "switch01.local",
		SourceMAC: mac,
		Interface: "eth0",
		FirstSeen: time.Now(),
		LastSeen:  time.Now(),
	}
	store.Update(neighbor)

	// Test with very small terminal (below minDetailPopupHeight)
	m := NewNeighborTable(store, types.InterfaceInfo{Name: "eth0"}, "", &cfg)
	m.width = 80
	m.height = 15 // Below minimum of 20
	m.showDetail = true
	m.selectedIndex = 0

	n := m.getSelectedNeighbor()
	if n == nil {
		t.Fatal("getSelectedNeighbor returned nil")
	}

	output := m.renderDetailView(n)
	lines := strings.Split(output, "\n")
	lineCount := len(lines)

	// Should still be correct height
	if lineCount != m.height {
		t.Errorf("Line count mismatch: got %d, want %d", lineCount, m.height)
	}

	// Should show "too small" message
	if !strings.Contains(output, "too small") {
		t.Error("Expected 'too small' message in output")
	}

	// Should still have footer
	lastLine := lines[len(lines)-1]
	if !strings.Contains(lastLine, "refresh") {
		t.Error("Footer not found on last line")
	}

	fmt.Printf("Small terminal (height=%d): lines=%d, hasMessage=%v\n",
		m.height, lineCount, strings.Contains(output, "too small"))
}

func TestLipglossPlaceOutput(t *testing.T) {
	// Test what lipgloss.Place actually produces
	store := types.NewNeighborStore()
	cfg := config.DefaultConfig()

	mac, _ := net.ParseMAC("00:11:22:33:44:55")
	neighbor := &types.Neighbor{
		ID:        "switch01",
		Hostname:  "switch01.local",
		SourceMAC: mac,
		Interface: "eth0",
		FirstSeen: time.Now(),
		LastSeen:  time.Now(),
	}
	store.Update(neighbor)

	m := NewNeighborTable(store, types.InterfaceInfo{Name: "eth0"}, "", &cfg)
	m.width = 80
	m.height = 30

	contentHeight := m.height - 2 // 28

	n := m.getSelectedNeighbor()
	if n == nil {
		t.Fatal("getSelectedNeighbor returned nil")
	}

	popup := m.renderDetailPopup(n, contentHeight)

	newlineCount := strings.Count(popup, "\n")
	endsWithNewline := strings.HasSuffix(popup, "\n")

	fmt.Printf("=== lipgloss.Place Analysis ===\n")
	fmt.Printf("Requested height: %d\n", contentHeight)
	fmt.Printf("Newline count in output: %d\n", newlineCount)
	fmt.Printf("Ends with newline: %v\n", endsWithNewline)
	fmt.Printf("Implied line count: %d\n", newlineCount+1)

	// If Place returns contentHeight lines, it should have contentHeight-1 internal newlines
	// OR contentHeight lines with contentHeight newlines if there's a trailing one
	expectedNewlines := contentHeight - 1
	if endsWithNewline {
		expectedNewlines = contentHeight
	}

	if newlineCount != expectedNewlines && newlineCount != contentHeight-1 && newlineCount != contentHeight {
		t.Errorf("Unexpected newline count: got %d, expected around %d", newlineCount, contentHeight-1)
	}
}
