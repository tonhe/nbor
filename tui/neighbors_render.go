package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"nbor/logger"
	"nbor/types"
	"nbor/version"
)

// column defines a table column for responsive display
type column struct {
	name     string
	width    int
	priority int // Lower = higher priority (shown first)
	getter   func(*types.Neighbor) string
}

// View renders the neighbor table
func (m NeighborTableModel) View() string {
	// If detail popup is active, show header + popup + footer
	if m.showDetail {
		if n := m.getSelectedNeighbor(); n != nil {
			return m.renderDetailView(n)
		}
	}

	// Render normal table view
	return m.renderBaseView()
}

// Minimum height required to display the detail popup (popup ~17 lines + header + footer)
const minDetailPopupHeight = 20

// renderDetailView renders the detail popup with header and footer visible
func (m NeighborTableModel) renderDetailView(n *types.Neighbor) string {
	header := m.renderHeader()
	footer := m.renderFooter()

	// If terminal is too small, show a message instead of the popup
	if m.height < minDetailPopupHeight {
		contentHeight := m.height - 2
		return m.renderTooSmallMessage(header, footer, contentHeight)
	}

	// Render popup centered in content area
	contentHeight := m.height - 2
	popup := m.renderDetailPopup(n, contentHeight)

	// Remove any trailing newline from popup to ensure consistent formatting
	popup = strings.TrimSuffix(popup, "\n")

	// Count actual popup lines
	popupLines := strings.Count(popup, "\n") + 1

	// Truncate if popup is larger than contentHeight
	if popupLines > contentHeight {
		lines := strings.Split(popup, "\n")
		lines = lines[:contentHeight]
		popup = strings.Join(lines, "\n")
		popupLines = contentHeight
	}

	// Calculate padding needed to push footer to bottom
	// Total lines needed: header (1) + popup + padding + footer (1) = m.height
	headerLines := strings.Count(header, "\n") + 1
	footerLines := 1
	usedLines := headerLines + popupLines + footerLines
	paddingLines := m.height - usedLines
	if paddingLines < 0 {
		paddingLines = 0
	}

	var b strings.Builder
	// DEBUG: Add visible marker at very start to see what's happening
	b.WriteString(fmt.Sprintf("[h=%d,ph=%d,pad=%d] ", m.height, popupLines, paddingLines))
	b.WriteString(header)
	b.WriteString("\n")
	b.WriteString(popup)
	// Add padding to push footer to bottom (popup already ends without trailing newline)
	// We need paddingLines + 1 newlines: paddingLines for blank space, plus 1 to end the last popup line
	b.WriteString(strings.Repeat("\n", paddingLines+1))
	b.WriteString(footer)

	return b.String()
}

// renderTooSmallMessage renders a message when terminal is too small for the popup
func (m NeighborTableModel) renderTooSmallMessage(header, footer string, contentHeight int) string {
	theme := DefaultTheme
	bg := theme.Base00

	msg := lipgloss.NewStyle().
		Foreground(theme.Base03).
		Background(bg).
		Width(m.width).
		Align(lipgloss.Center).
		Render("Terminal too small for details. Press ESC to close.")

	// Center the message vertically
	content := lipgloss.Place(
		m.width,
		contentHeight,
		lipgloss.Center,
		lipgloss.Center,
		msg,
		lipgloss.WithWhitespaceBackground(bg),
	)
	content = strings.TrimSuffix(content, "\n")

	var b strings.Builder
	b.WriteString(header)
	b.WriteString("\n")
	b.WriteString(content)
	b.WriteString("\n")
	b.WriteString(footer)

	return b.String()
}

// renderBaseView renders the main table view (header + table + footer)
func (m NeighborTableModel) renderBaseView() string {
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

	// Calculate which columns fit (already sorted by priority in definition order 1-8)
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

	return visibleColumns
}

// renderTable renders the neighbor table
func (m NeighborTableModel) renderTable() string {
	var b strings.Builder

	neighbors := m.getFilteredNeighbors()
	columns := m.getVisibleColumns()

	// Blank line after header
	b.WriteString("\n")

	// Table header (with prefix space for alignment with row cursor)
	var headerCells []string
	for _, col := range columns {
		headerCells = append(headerCells, truncate(col.name, col.width))
	}

	headerRow := "  " + strings.Join(headerCells, "  ")
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

	// Determine visible range
	startIdx := m.scrollOffset
	endIdx := startIdx + m.visibleRows()
	if endIdx > len(neighbors) {
		endIdx = len(neighbors)
	}

	// Render visible rows
	for i := startIdx; i < endIdx; i++ {
		n := neighbors[i]
		isSelected := (i == m.selectedIndex)
		b.WriteString(m.renderNeighborRow(n, columns, isSelected))
		b.WriteString("\n")
	}

	// Scroll indicator
	if len(neighbors) > m.visibleRows() {
		scrollInfo := fmt.Sprintf("  [%d-%d of %d]", startIdx+1, endIdx, len(neighbors))
		b.WriteString(m.styles.StatusInfo.Render(scrollInfo))
	}

	return b.String()
}

// renderNeighborRow renders a single neighbor row
func (m NeighborTableModel) renderNeighborRow(n *types.Neighbor, columns []column, isSelected bool) string {
	theme := DefaultTheme

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

	// Subtle cursor indicator for selection
	var prefix string
	if isSelected {
		cursorStyle := lipgloss.NewStyle().
			Foreground(theme.Base0D).
			Bold(true)
		prefix = cursorStyle.Render("▸ ")
	} else {
		prefix = "  "
	}

	var cells []string
	for _, col := range columns {
		value := col.getter(n)
		cells = append(cells, cellStyle.Render(truncate(value, col.width)))
	}

	row := strings.Join(cells, "  ")

	return prefix + row
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
		keyStyle.Render("↑/↓") + textStyle.Render(" select") + sep +
		keyStyle.Render("enter") + textStyle.Render(" details") + sep +
		keyStyle.Render("q") + textStyle.Render(" quit")

	// Build right side: log file
	var rightPart string
	if m.logPath != "" {
		fileStyle := lipgloss.NewStyle().
			Foreground(theme.Base0A).
			Background(bg)
		rightPart = textStyle.Render("log: ") + fileStyle.Render(m.logPath)
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
