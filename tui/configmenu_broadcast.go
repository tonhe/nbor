package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// updateBroadcast handles key events for the Broadcast Options sub-menu
func (m ConfigMenuModel) updateBroadcast(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Broadcast sub-menu fields organized by row:
	// Row 0: System Name (0)
	// Row 1: Description (1)
	// Row 2: CDP Broadcast (2), LLDP Broadcast (3)
	// Row 3: Start on Launch (4)
	// Row 4: Interval (5)
	// Row 5: TTL (6)
	// Row 6: Cap Router (7), Cap Bridge (8), Cap Station (9)
	// Row 7: Back button (10)

	// Define row groupings for left/right navigation
	broadcastRows := [][]int{
		{0},       // System Name
		{1},       // Description
		{2, 3},    // CDP, LLDP
		{4},       // Start on Launch
		{5},       // Interval
		{6},       // TTL
		{7, 8, 9}, // Router, Bridge, Station
		{10},      // Back
	}

	switch {
	case key.Matches(msg, configMenuKeys.Back):
		m.subState = SubStateMain
		m.blurAllBroadcastInputs()

	case key.Matches(msg, configMenuKeys.Left):
		// Move left within the current row
		row, col := m.findBroadcastPosition(broadcastRows)
		if col > 0 {
			m.blurAllBroadcastInputs()
			m.subCursor = broadcastRows[row][col-1]
			m.focusBroadcastInput()
		}

	case key.Matches(msg, configMenuKeys.Right):
		// Move right within the current row
		row, col := m.findBroadcastPosition(broadcastRows)
		if col < len(broadcastRows[row])-1 {
			m.blurAllBroadcastInputs()
			m.subCursor = broadcastRows[row][col+1]
			m.focusBroadcastInput()
		}

	case key.Matches(msg, configMenuKeys.Up):
		m.blurAllBroadcastInputs()
		row, col := m.findBroadcastPosition(broadcastRows)
		row--
		if row < 0 {
			row = len(broadcastRows) - 1
		}
		// Keep column position if possible, otherwise go to last item in row
		if col >= len(broadcastRows[row]) {
			col = len(broadcastRows[row]) - 1
		}
		m.subCursor = broadcastRows[row][col]
		m.focusBroadcastInput()

	case key.Matches(msg, configMenuKeys.Down), key.Matches(msg, configMenuKeys.Tab):
		m.blurAllBroadcastInputs()
		row, col := m.findBroadcastPosition(broadcastRows)
		row++
		if row >= len(broadcastRows) {
			row = 0
		}
		// Keep column position if possible, otherwise go to last item in row
		if col >= len(broadcastRows[row]) {
			col = len(broadcastRows[row]) - 1
		}
		m.subCursor = broadcastRows[row][col]
		m.focusBroadcastInput()

	case key.Matches(msg, configMenuKeys.Select):
		switch m.subCursor {
		case 2:
			m.cdpBroadcast = !m.cdpBroadcast
		case 3:
			m.lldpBroadcast = !m.lldpBroadcast
		case 4:
			m.broadcastOnStartup = !m.broadcastOnStartup
		case 7:
			m.capRouter = !m.capRouter
		case 8:
			m.capBridge = !m.capBridge
		case 9:
			m.capStation = !m.capStation
		case 10: // Back
			m.subState = SubStateMain
			m.blurAllBroadcastInputs()
		}

	default:
		// Pass to text input if focused
		var cmd tea.Cmd
		switch m.subCursor {
		case 0:
			m.systemNameInput, cmd = m.systemNameInput.Update(msg)
			return m, cmd
		case 1:
			m.systemDescInput, cmd = m.systemDescInput.Update(msg)
			return m, cmd
		case 5:
			m.intervalInput, cmd = m.intervalInput.Update(msg)
			return m, cmd
		case 6:
			m.ttlInput, cmd = m.ttlInput.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

// findBroadcastPosition returns the row and column position for the current cursor
func (m *ConfigMenuModel) findBroadcastPosition(rows [][]int) (row, col int) {
	return findRowPosition(m.subCursor, rows)
}

func (m *ConfigMenuModel) blurAllBroadcastInputs() {
	m.systemNameInput.Blur()
	m.systemDescInput.Blur()
	m.intervalInput.Blur()
	m.ttlInput.Blur()
}

func (m *ConfigMenuModel) focusBroadcastInput() {
	switch m.subCursor {
	case 0:
		m.systemNameInput.Focus()
	case 1:
		m.systemDescInput.Focus()
	case 5:
		m.intervalInput.Focus()
	case 6:
		m.ttlInput.Focus()
	}
}

// renderBroadcast renders the Broadcast Options sub-menu
func (m ConfigMenuModel) renderBroadcast() string {
	theme := DefaultTheme
	var b strings.Builder

	sectionStyle := lipgloss.NewStyle().Foreground(theme.Base0D).Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(theme.Base03)

	b.WriteString("\n")

	// System Identity
	b.WriteString("  ")
	b.WriteString(sectionStyle.Render("System Identity"))
	b.WriteString("\n\n")

	// System Name
	b.WriteString("  ")
	b.WriteString(renderCursor(m.subCursor == 0, theme))
	b.WriteString(renderLabel("System Name", m.subCursor == 0, theme))
	b.WriteString("    ")
	b.WriteString(m.systemNameInput.View())
	b.WriteString("\n")

	// Description
	b.WriteString("  ")
	b.WriteString(renderCursor(m.subCursor == 1, theme))
	b.WriteString(renderLabel("Description", m.subCursor == 1, theme))
	b.WriteString("    ")
	b.WriteString(m.systemDescInput.View())
	b.WriteString("\n\n")

	// Protocol Broadcasting
	b.WriteString("  ")
	b.WriteString(sectionStyle.Render("Protocol Broadcasting"))
	b.WriteString("\n\n")

	// CDP Broadcast / LLDP Broadcast (same row)
	b.WriteString("  ")
	b.WriteString(renderCursor(m.subCursor == 2, theme))
	b.WriteString(renderCheckbox(m.cdpBroadcast, m.subCursor == 2, theme))
	b.WriteString(" ")
	b.WriteString(renderLabel("CDP", m.subCursor == 2, theme))
	b.WriteString("     ")

	// LLDP Broadcast
	b.WriteString(renderCursor(m.subCursor == 3, theme))
	b.WriteString(renderCheckbox(m.lldpBroadcast, m.subCursor == 3, theme))
	b.WriteString(" ")
	b.WriteString(renderLabel("LLDP", m.subCursor == 3, theme))
	b.WriteString("\n")

	// Start on Launch
	b.WriteString("  ")
	b.WriteString(renderCursor(m.subCursor == 4, theme))
	b.WriteString(renderCheckbox(m.broadcastOnStartup, m.subCursor == 4, theme))
	b.WriteString(" ")
	b.WriteString(renderLabel("Start on launch", m.subCursor == 4, theme))
	b.WriteString("\n\n")

	// Timing
	b.WriteString("  ")
	b.WriteString(sectionStyle.Render("Timing"))
	b.WriteString("\n\n")

	// Interval
	b.WriteString("  ")
	b.WriteString(renderCursor(m.subCursor == 5, theme))
	b.WriteString(renderLabel("Interval", m.subCursor == 5, theme))
	b.WriteString("       ")
	b.WriteString(m.intervalInput.View())
	b.WriteString(dimStyle.Render(" seconds"))
	b.WriteString("\n")

	// TTL
	b.WriteString("  ")
	b.WriteString(renderCursor(m.subCursor == 6, theme))
	b.WriteString(renderLabel("TTL", m.subCursor == 6, theme))
	b.WriteString("            ")
	b.WriteString(m.ttlInput.View())
	b.WriteString(dimStyle.Render(" seconds"))
	b.WriteString("\n\n")

	// Capabilities
	b.WriteString("  ")
	b.WriteString(sectionStyle.Render("Capabilities (advertised)"))
	b.WriteString("\n\n")

	// Router / Bridge / Station (same row)
	b.WriteString("  ")
	b.WriteString(renderCursor(m.subCursor == 7, theme))
	b.WriteString(renderCheckbox(m.capRouter, m.subCursor == 7, theme))
	b.WriteString(" ")
	b.WriteString(renderLabel("Router", m.subCursor == 7, theme))
	b.WriteString("  ")

	// Bridge
	b.WriteString(renderCursor(m.subCursor == 8, theme))
	b.WriteString(renderCheckbox(m.capBridge, m.subCursor == 8, theme))
	b.WriteString(" ")
	b.WriteString(renderLabel("Bridge", m.subCursor == 8, theme))
	b.WriteString("  ")

	// Station
	b.WriteString(renderCursor(m.subCursor == 9, theme))
	b.WriteString(renderCheckbox(m.capStation, m.subCursor == 9, theme))
	b.WriteString(" ")
	b.WriteString(renderLabel("Station", m.subCursor == 9, theme))
	b.WriteString("\n\n")

	// Back button
	b.WriteString("  ")
	b.WriteString(renderCursor(m.subCursor == 10, theme))
	b.WriteString(renderLabel("[Back]", m.subCursor == 10, theme))
	b.WriteString("\n")

	return b.String()
}
