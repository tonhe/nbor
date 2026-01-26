package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"nbor/config"
	"nbor/logger"
	"nbor/types"
	"nbor/version"
)

// Column definition for responsive table
type column struct {
	name     string
	width    int
	priority int // Lower = higher priority (shown first)
	getter   func(*types.Neighbor) string
}

// NeighborTableModel is the model for the neighbor table view
type NeighborTableModel struct {
	store        *types.NeighborStore
	ifaceInfo    types.InterfaceInfo
	config       *config.Config
	width        int
	height       int
	styles       Styles
	scrollOffset int
	flashRows    map[string]time.Time // Track rows to flash
	logPath      string
	broadcasting bool // Whether broadcasting is currently active
}

// NewNeighborTable creates a new neighbor table model
func NewNeighborTable(store *types.NeighborStore, ifaceInfo types.InterfaceInfo, logPath string, cfg *config.Config) NeighborTableModel {
	// Determine initial broadcast state from config
	// Broadcasting only starts if BroadcastOnStartup is true AND a protocol is configured
	broadcasting := cfg.BroadcastOnStartup && (cfg.CDPBroadcast || cfg.LLDPBroadcast)

	return NeighborTableModel{
		store:        store,
		ifaceInfo:    ifaceInfo,
		config:       cfg,
		styles:       DefaultStyles,
		flashRows:    make(map[string]time.Time),
		logPath:      logPath,
		broadcasting: broadcasting,
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
		key.WithHelp("↑", "scroll up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓", "scroll down"),
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
		switch {
		case key.Matches(msg, neighborKeys.Refresh):
			// Clear stale entries and refresh
			m.store.ClearNewFlags()
			m.flashRows = make(map[string]time.Time)
			m.scrollOffset = 0
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
			if m.scrollOffset > 0 {
				m.scrollOffset--
			}
		case key.Matches(msg, neighborKeys.Down):
			neighbors := m.store.GetAll()
			maxScroll := len(neighbors) - m.visibleRows()
			if maxScroll < 0 {
				maxScroll = 0
			}
			if m.scrollOffset < maxScroll {
				m.scrollOffset++
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case TickMsg:
		// Mark stale neighbors (not seen in 3-4 minutes)
		m.store.MarkStale(4 * time.Minute)

		// Clear old flash entries
		now := time.Now()
		for k, t := range m.flashRows {
			if now.Sub(t) > 2*time.Second {
				delete(m.flashRows, k)
			}
		}

		return m, tickCmd()

	case NewNeighborMsg:
		// Mark this row for flashing
		m.flashRows[msg.Neighbor.NeighborKey()] = time.Now()
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

// View renders the neighbor table
func (m NeighborTableModel) View() string {
	// Calculate content heights
	header := m.renderHeader()
	table := m.renderTable()
	footer := m.renderFooter()

	// Calculate how many blank lines we need to push footer to bottom
	headerLines := strings.Count(header, "\n") + 1
	tableLines := strings.Count(table, "\n")
	footerLines := 1

	usedLines := headerLines + tableLines + footerLines
	remainingLines := m.height - usedLines
	if remainingLines < 0 {
		remainingLines = 0
	}

	// Build the view with padding to push footer to bottom
	var b strings.Builder
	b.WriteString(header)
	b.WriteString("\n")
	b.WriteString(table)
	b.WriteString(strings.Repeat("\n", remainingLines))
	b.WriteString(footer)

	return b.String()
}

// renderHeader renders the application header with colors spread across width
func (m NeighborTableModel) renderHeader() string {
	theme := DefaultTheme
	bg := theme.Base01

	// Single space with background for joining elements
	sp := lipgloss.NewStyle().Background(bg).Render(" ")

	// Left side: app name and version
	nameStyle := lipgloss.NewStyle().
		Foreground(theme.Base0C).
		Background(bg).
		Bold(true)
	versionStyle := lipgloss.NewStyle().
		Foreground(theme.Base03).
		Background(bg)
	leftPart := nameStyle.Render("nbor") + sp + versionStyle.Render("v"+version.Version)

	// Middle: interface info
	ifaceStyle := lipgloss.NewStyle().
		Foreground(theme.Base0D).
		Background(bg).
		Bold(true)
	macStyle := lipgloss.NewStyle().
		Foreground(theme.Base05).
		Background(bg)
	speedStyle := lipgloss.NewStyle().
		Foreground(theme.Base0A).
		Background(bg)

	mac := ""
	if m.ifaceInfo.MAC != nil {
		mac = m.ifaceInfo.MAC.String()
	}

	middlePart := ifaceStyle.Render(m.ifaceInfo.Name)
	if mac != "" {
		middlePart += sp + macStyle.Render(mac)
	}
	if m.ifaceInfo.Speed != "" {
		middlePart += sp + speedStyle.Render(m.ifaceInfo.Speed)
	}

	// Right side: neighbor count
	countStyle := lipgloss.NewStyle().
		Foreground(theme.Base0B).
		Background(bg).
		Bold(true)
	labelStyle := lipgloss.NewStyle().
		Foreground(theme.Base04).
		Background(bg)
	count := m.store.Count()
	rightPart := countStyle.Render(fmt.Sprintf("%d", count)) + sp + labelStyle.Render("neighbor(s)")

	// Calculate spacing to spread across width
	leftLen := lipgloss.Width(leftPart)
	middleLen := lipgloss.Width(middlePart)
	rightLen := lipgloss.Width(rightPart)

	// Account for padding (1 on each side)
	availableWidth := m.width - 2
	totalContentWidth := leftLen + middleLen + rightLen

	// Distribute remaining space
	remainingSpace := availableWidth - totalContentWidth
	if remainingSpace < 2 {
		remainingSpace = 2
	}

	leftGap := remainingSpace / 2
	rightGap := remainingSpace - leftGap

	// Build header content with background-colored spaces
	spaceStyle := lipgloss.NewStyle().Background(bg)
	headerContent := leftPart + spaceStyle.Render(strings.Repeat(" ", leftGap)) + middlePart + spaceStyle.Render(strings.Repeat(" ", rightGap)) + rightPart

	// Apply background style to container
	headerStyle := lipgloss.NewStyle().
		Background(bg).
		Padding(0, 1).
		Width(m.width)

	return headerStyle.Render(headerContent)
}

// getVisibleColumns returns columns that fit in the current width
func (m NeighborTableModel) getVisibleColumns() []column {
	// Define all columns with priorities (lower = shown first)
	// Priority order: hostname, port, last seen, mgmt IP, platform, location, protocol, capabilities
	allColumns := []column{
		{name: "Hostname", width: 20, priority: 1, getter: func(n *types.Neighbor) string { return n.Hostname }},
		{name: "Port", width: 12, priority: 2, getter: func(n *types.Neighbor) string { return n.PortID }},
		{name: "Last Seen", width: 10, priority: 3, getter: func(n *types.Neighbor) string { return logger.FormatDuration(n.LastSeen) }},
		{name: "Mgmt IP", width: 15, priority: 4, getter: func(n *types.Neighbor) string {
			if n.ManagementIP != nil {
				return n.ManagementIP.String()
			}
			return ""
		}},
		{name: "Platform", width: 18, priority: 5, getter: func(n *types.Neighbor) string { return n.Platform }},
		{name: "Location", width: 15, priority: 6, getter: func(n *types.Neighbor) string { return n.Location }},
		{name: "Proto", width: 9, priority: 7, getter: func(n *types.Neighbor) string { return string(n.Protocol) }},
		{name: "Capabilities", width: 15, priority: 8, getter: func(n *types.Neighbor) string { return logger.FormatCapabilities(n.Capabilities) }},
	}

	// Sort by priority
	sort.Slice(allColumns, func(i, j int) bool {
		return allColumns[i].priority < allColumns[j].priority
	})

	// Calculate which columns fit
	availableWidth := m.width - 2 // Padding
	usedWidth := 0
	var visibleColumns []column

	for _, col := range allColumns {
		colWidth := col.width + 2 // Add spacing between columns
		if usedWidth+colWidth <= availableWidth {
			visibleColumns = append(visibleColumns, col)
			usedWidth += colWidth
		}
	}

	// Sort visible columns back to a logical display order
	// Hostname, Port, Mgmt IP, Platform, Location, Capabilities, Proto, Last Seen
	displayOrder := map[string]int{
		"Hostname":     1,
		"Port":         2,
		"Mgmt IP":      3,
		"Platform":     4,
		"Location":     5,
		"Capabilities": 6,
		"Proto":        7,
		"Last Seen":    8,
	}

	sort.Slice(visibleColumns, func(i, j int) bool {
		return displayOrder[visibleColumns[i].name] < displayOrder[visibleColumns[j].name]
	})

	return visibleColumns
}

// renderTable renders the neighbor table
func (m NeighborTableModel) renderTable() string {
	var b strings.Builder

	neighbors := m.store.GetAll()
	columns := m.getVisibleColumns()

	// Sort neighbors by hostname
	sort.Slice(neighbors, func(i, j int) bool {
		return neighbors[i].Hostname < neighbors[j].Hostname
	})

	// Blank line after header
	b.WriteString("\n")

	// Table header
	var headerCells []string
	for _, col := range columns {
		headerCells = append(headerCells, truncate(col.name, col.width))
	}

	headerRow := strings.Join(headerCells, "  ")
	b.WriteString(m.styles.TableHeader.Render(headerRow))
	b.WriteString("\n")

	if len(neighbors) == 0 {
		// Show listening message
		b.WriteString("\n")
		listening := m.styles.StatusListening.Render("  Listening for CDP and LLDP packets...")
		b.WriteString(listening)
		b.WriteString("\n\n")
		hint := m.styles.StatusInfo.Render("  Neighbors will appear here as they announce themselves.")
		b.WriteString(hint)
		return b.String()
	}

	// Apply scroll offset
	visibleNeighbors := neighbors
	if m.scrollOffset > 0 && m.scrollOffset < len(neighbors) {
		visibleNeighbors = neighbors[m.scrollOffset:]
	}

	// Limit to visible rows
	maxRows := m.visibleRows()
	if len(visibleNeighbors) > maxRows {
		visibleNeighbors = visibleNeighbors[:maxRows]
	}

	// Render rows
	for _, n := range visibleNeighbors {
		b.WriteString(m.renderNeighborRow(n, columns))
		b.WriteString("\n")
	}

	// Scroll indicator
	if len(neighbors) > maxRows {
		scrollInfo := fmt.Sprintf("  [%d-%d of %d]", m.scrollOffset+1, m.scrollOffset+len(visibleNeighbors), len(neighbors))
		b.WriteString(m.styles.StatusInfo.Render(scrollInfo))
	}

	return b.String()
}

// renderNeighborRow renders a single neighbor row
func (m NeighborTableModel) renderNeighborRow(n *types.Neighbor, columns []column) string {
	// Determine style based on state:
	// - Stale (no updates for 3-4 min) = gray
	// - Active (getting updates) = green
	// - New/flashing = bold green
	var cellStyle lipgloss.Style

	if n.IsStale {
		cellStyle = m.styles.TableCellStale
	} else if _, flashing := m.flashRows[n.NeighborKey()]; flashing || n.IsNew {
		// Brand new or just updated - bold green
		cellStyle = lipgloss.NewStyle().
			Foreground(m.styles.TableRowNew.GetForeground()).
			Bold(true)
	} else {
		// Active neighbor - regular green (not bold)
		cellStyle = lipgloss.NewStyle().
			Foreground(m.styles.TableRowNew.GetForeground())
	}

	var cells []string
	for _, col := range columns {
		value := col.getter(n)
		cells = append(cells, cellStyle.Render(truncate(value, col.width)))
	}

	return strings.Join(cells, "  ")
}

// renderFooter renders the footer with hotkeys spread across width
func (m NeighborTableModel) renderFooter() string {
	theme := DefaultTheme
	bg := theme.Base01

	// Key styling - all with background
	keyStyle := lipgloss.NewStyle().
		Foreground(theme.Base0C).
		Background(bg).
		Bold(true)
	textStyle := lipgloss.NewStyle().
		Foreground(theme.Base04).
		Background(bg)
	sepStyle := lipgloss.NewStyle().
		Foreground(theme.Base02).
		Background(bg)
	onStyle := lipgloss.NewStyle().
		Foreground(theme.Base0B).
		Background(bg).
		Bold(true)
	offStyle := lipgloss.NewStyle().
		Foreground(theme.Base03).
		Background(bg)

	// Build left side: commands with broadcast status
	sep := sepStyle.Render(" │ ")

	// Broadcast status indicator
	var broadcastStatus string
	if m.broadcasting {
		broadcastStatus = onStyle.Render("TX")
	} else {
		broadcastStatus = offStyle.Render("--")
	}

	leftPart := keyStyle.Render("r") + textStyle.Render(" refresh") + sep +
		keyStyle.Render("b") + textStyle.Render(" broadcast:") + broadcastStatus + sep +
		keyStyle.Render("c") + textStyle.Render(" config") + sep +
		keyStyle.Render("↑/↓") + textStyle.Render(" scroll") + sep +
		keyStyle.Render("q") + textStyle.Render(" quit")

	// Build right side: log file
	var rightPart string
	if m.logPath != "" {
		fileStyle := lipgloss.NewStyle().
			Foreground(theme.Base0A).
			Background(bg)
		rightPart = textStyle.Render("logging: ") + fileStyle.Render(m.logPath)
	}

	// Calculate spacing to spread across width
	leftLen := lipgloss.Width(leftPart)
	rightLen := lipgloss.Width(rightPart)

	// Account for padding (1 on each side)
	availableWidth := m.width - 2
	totalContentWidth := leftLen + rightLen

	// Calculate gap
	gap := availableWidth - totalContentWidth
	if gap < 1 {
		gap = 1
	}

	// Build footer content with background-colored spaces
	spaceStyle := lipgloss.NewStyle().Background(bg)
	footerContent := leftPart + spaceStyle.Render(strings.Repeat(" ", gap)) + rightPart

	// Apply background style
	footerStyle := lipgloss.NewStyle().
		Background(bg).
		Padding(0, 1).
		Width(m.width)

	return footerStyle.Render(footerContent)
}

// truncate truncates a string to the given width and pads with spaces
func truncate(s string, width int) string {
	// Use lipgloss width to handle Unicode properly
	visWidth := lipgloss.Width(s)
	if visWidth <= width {
		return s + strings.Repeat(" ", width-visWidth)
	}
	if width <= 3 {
		// Truncate by runes, not bytes
		runes := []rune(s)
		if len(runes) > width {
			return string(runes[:width])
		}
		return s
	}
	// Truncate to width-3 and add ellipsis
	runes := []rune(s)
	targetLen := width - 3
	if targetLen < 0 {
		targetLen = 0
	}
	// Find how many runes fit in targetLen visual width
	result := ""
	for _, r := range runes {
		if lipgloss.Width(result+string(r)) > targetLen {
			break
		}
		result += string(r)
	}
	return result + "..."
}

// MarkNewNeighbor marks a neighbor for flashing
func (m *NeighborTableModel) MarkNewNeighbor(n *types.Neighbor) {
	m.flashRows[n.NeighborKey()] = time.Now()
}
