package tui

import (
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"nbor/config"
	"nbor/types"
)

// NeighborTableModel is the model for the neighbor table view
type NeighborTableModel struct {
	store         *types.NeighborStore
	ifaceInfo     types.InterfaceInfo
	config        *config.Config
	width         int
	height        int
	styles        Styles
	scrollOffset  int
	selectedIndex int                   // Currently selected row index
	showDetail    bool                  // Whether detail popup is visible
	flashRows     map[string]time.Time  // Track rows to flash
	logPath       string
	broadcasting  bool // Whether broadcasting is currently active
}

// NewNeighborTable creates a new neighbor table model
func NewNeighborTable(store *types.NeighborStore, ifaceInfo types.InterfaceInfo, logPath string, cfg *config.Config) NeighborTableModel {
	// Determine initial broadcast state from config
	// Broadcasting only starts if BroadcastOnStartup is true AND a protocol is configured
	broadcasting := cfg.BroadcastOnStartup && (cfg.CDPBroadcast || cfg.LLDPBroadcast)

	return NeighborTableModel{
		store:         store,
		ifaceInfo:     ifaceInfo,
		config:        cfg,
		styles:        DefaultStyles,
		flashRows:     make(map[string]time.Time),
		logPath:       logPath,
		broadcasting:  broadcasting,
		selectedIndex: 0,
		showDetail:    false,
	}
}

// Init initializes the neighbor table
func (m NeighborTableModel) Init() tea.Cmd {
	return tickCmd()
}

// TickMsg triggers periodic updates
type TickMsg time.Time

// NewNeighborMsg indicates a new neighbor was discovered
type NewNeighborMsg struct {
	Neighbor *types.Neighbor
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// neighborTableKeyMap defines key bindings for the neighbor table
type neighborTableKeyMap struct {
	Refresh   key.Binding
	Broadcast key.Binding
	Config    key.Binding
	Quit      key.Binding
	Up        key.Binding
	Down      key.Binding
	Select    key.Binding
	Back      key.Binding
}

var neighborKeys = neighborTableKeyMap{
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh display"),
	),
	Broadcast: key.NewBinding(
		key.WithKeys("b"),
		key.WithHelp("b", "toggle broadcast"),
	),
	Config: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "configuration"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c", "q"),
		key.WithHelp("ctrl+c", "quit"),
	),
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Select: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "view details"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "close"),
	),
}

// ToggleBroadcastMsg is sent when broadcast is toggled
type ToggleBroadcastMsg struct {
	Enabled bool
}

// Update handles messages for the neighbor table
func (m NeighborTableModel) Update(msg tea.Msg) (NeighborTableModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle detail popup mode separately
		if m.showDetail {
			return m.updateDetailMode(msg)
		}
		return m.updateTableMode(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case TickMsg:
		// Mark stale neighbors based on config
		stalenessTimeout := time.Duration(m.config.StalenessTimeout) * time.Second
		m.store.MarkStale(stalenessTimeout)

		// Remove stale neighbors if configured (0 = never remove)
		if m.config.StaleRemovalTime > 0 {
			removalTimeout := time.Duration(m.config.StaleRemovalTime) * time.Second
			m.store.RemoveStale(removalTimeout)
		}

		// Clear old flash entries
		now := time.Now()
		for k, t := range m.flashRows {
			if now.Sub(t) > 2*time.Second {
				delete(m.flashRows, k)
			}
		}

		// Ensure selectedIndex stays valid if neighbors were removed
		neighbors := m.getFilteredNeighbors()
		if m.selectedIndex >= len(neighbors) && len(neighbors) > 0 {
			m.selectedIndex = len(neighbors) - 1
		}

		return m, tickCmd()

	case NewNeighborMsg:
		// Mark this row for flashing
		m.flashRows[msg.Neighbor.NeighborKey()] = time.Now()
	}

	return m, nil
}

// updateTableMode handles key events when viewing the table
func (m NeighborTableModel) updateTableMode(msg tea.KeyMsg) (NeighborTableModel, tea.Cmd) {
	neighbors := m.getFilteredNeighbors()
	neighborCount := len(neighbors)

	switch {
	case key.Matches(msg, neighborKeys.Refresh):
		// Clear stale entries and refresh
		m.store.ClearNewFlags()
		m.flashRows = make(map[string]time.Time)
		m.scrollOffset = 0
		m.selectedIndex = 0
		// Force a screen clear/redraw
		return m, tea.ClearScreen

	case key.Matches(msg, neighborKeys.Broadcast):
		// Toggle broadcasting on/off (runtime only, doesn't change protocol config)
		m.broadcasting = !m.broadcasting
		// Send message to main to start/stop broadcaster
		return m, func() tea.Msg {
			return ToggleBroadcastMsg{Enabled: m.broadcasting}
		}

	case key.Matches(msg, neighborKeys.Config):
		// Open configuration menu
		return m, func() tea.Msg {
			return GoToConfigMenuMsg{}
		}

	case key.Matches(msg, neighborKeys.Quit):
		return m, tea.Quit

	case key.Matches(msg, neighborKeys.Up):
		if neighborCount > 0 {
			m.selectedIndex--
			if m.selectedIndex < 0 {
				m.selectedIndex = neighborCount - 1
				// Scroll to show the last item
				maxScroll := neighborCount - m.visibleRows()
				if maxScroll > 0 {
					m.scrollOffset = maxScroll
				}
			}
			// Ensure selected item is visible (scroll up if needed)
			if m.selectedIndex < m.scrollOffset {
				m.scrollOffset = m.selectedIndex
			}
		}

	case key.Matches(msg, neighborKeys.Down):
		if neighborCount > 0 {
			m.selectedIndex++
			if m.selectedIndex >= neighborCount {
				m.selectedIndex = 0
				m.scrollOffset = 0
			}
			// Ensure selected item is visible (scroll down if needed)
			visibleEnd := m.scrollOffset + m.visibleRows() - 1
			if m.selectedIndex > visibleEnd {
				m.scrollOffset = m.selectedIndex - m.visibleRows() + 1
			}
		}

	case key.Matches(msg, neighborKeys.Select):
		// Open detail popup if we have a valid selection
		if neighborCount > 0 && m.selectedIndex < neighborCount {
			m.showDetail = true
		}
	}

	return m, nil
}

// updateDetailMode handles key events when viewing the detail popup
func (m NeighborTableModel) updateDetailMode(msg tea.KeyMsg) (NeighborTableModel, tea.Cmd) {
	switch {
	case key.Matches(msg, neighborKeys.Back), key.Matches(msg, neighborKeys.Select):
		// Close detail popup
		m.showDetail = false
	case key.Matches(msg, neighborKeys.Quit):
		return m, tea.Quit
	}
	return m, nil
}

// visibleRows returns the number of visible table rows
func (m NeighborTableModel) visibleRows() int {
	// Account for header (1 line) + blank line + table header (1 line) + footer (1 line) + padding
	available := m.height - 6
	if available < 1 {
		available = 1
	}
	return available
}

// MarkNewNeighbor marks a neighbor for flashing
func (m *NeighborTableModel) MarkNewNeighbor(n *types.Neighbor) {
	m.flashRows[n.NeighborKey()] = time.Now()
}

// matchesCapabilityFilter checks if a neighbor matches the capability filter
// If no filter is set (empty slice), all neighbors match
func (m *NeighborTableModel) matchesCapabilityFilter(n *types.Neighbor) bool {
	// Empty filter means show all
	if len(m.config.FilterCapabilities) == 0 {
		return true
	}

	// Check if any of the neighbor's capabilities match the filter
	for _, neighborCap := range n.Capabilities {
		for _, filterCap := range m.config.FilterCapabilities {
			if strings.EqualFold(string(neighborCap), filterCap) {
				return true
			}
		}
	}
	return false
}

// getFilteredNeighbors returns neighbors that match the capability filter, sorted by hostname
func (m *NeighborTableModel) getFilteredNeighbors() []*types.Neighbor {
	allNeighbors := m.store.GetAll()

	var filtered []*types.Neighbor
	// If no filter, use all
	if len(m.config.FilterCapabilities) == 0 {
		filtered = allNeighbors
	} else {
		// Filter neighbors
		for _, n := range allNeighbors {
			if m.matchesCapabilityFilter(n) {
				filtered = append(filtered, n)
			}
		}
	}

	// Sort by hostname for consistent ordering
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Hostname < filtered[j].Hostname
	})

	return filtered
}

// getSelectedNeighbor returns the currently selected neighbor or nil
func (m *NeighborTableModel) getSelectedNeighbor() *types.Neighbor {
	neighbors := m.getFilteredNeighbors()
	if m.selectedIndex < 0 || m.selectedIndex >= len(neighbors) {
		return nil
	}
	return neighbors[m.selectedIndex]
}
