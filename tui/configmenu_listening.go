package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// updateListening handles key events for the Listening Options sub-menu
func (m ConfigMenuModel) updateListening(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Listening sub-menu fields organized by row:
	// Row 0: CDP Listen (0), LLDP Listen (1)
	// Row 1: Filter Router (2), Filter Bridge (3), Filter Station (4)
	// Row 2: Staleness Timeout (5)
	// Row 3: Stale Removal (6)
	// Row 4: Back button (7)

	// Define row groupings for left/right navigation
	listeningRows := [][]int{
		{0, 1},    // CDP, LLDP
		{2, 3, 4}, // Router, Bridge, Station
		{5},       // Staleness
		{6},       // Stale Removal
		{7},       // Back
	}

	switch {
	case key.Matches(msg, configMenuKeys.Back):
		m.subState = SubStateMain
		m.stalenessInput.Blur()
		m.staleRemovalInput.Blur()

	case key.Matches(msg, configMenuKeys.Left):
		// Move left within the current row
		row, col := m.findListeningPosition(listeningRows)
		if col > 0 {
			m.stalenessInput.Blur()
			m.staleRemovalInput.Blur()
			m.subCursor = listeningRows[row][col-1]
			m.focusListeningInput()
		}

	case key.Matches(msg, configMenuKeys.Right):
		// Move right within the current row
		row, col := m.findListeningPosition(listeningRows)
		if col < len(listeningRows[row])-1 {
			m.stalenessInput.Blur()
			m.staleRemovalInput.Blur()
			m.subCursor = listeningRows[row][col+1]
			m.focusListeningInput()
		}

	case key.Matches(msg, configMenuKeys.Up):
		m.stalenessInput.Blur()
		m.staleRemovalInput.Blur()
		row, col := m.findListeningPosition(listeningRows)
		row--
		if row < 0 {
			row = len(listeningRows) - 1
		}
		// Keep column position if possible, otherwise go to last item in row
		if col >= len(listeningRows[row]) {
			col = len(listeningRows[row]) - 1
		}
		m.subCursor = listeningRows[row][col]
		m.focusListeningInput()

	case key.Matches(msg, configMenuKeys.Down), key.Matches(msg, configMenuKeys.Tab):
		m.stalenessInput.Blur()
		m.staleRemovalInput.Blur()
		row, col := m.findListeningPosition(listeningRows)
		row++
		if row >= len(listeningRows) {
			row = 0
		}
		// Keep column position if possible, otherwise go to last item in row
		if col >= len(listeningRows[row]) {
			col = len(listeningRows[row]) - 1
		}
		m.subCursor = listeningRows[row][col]
		m.focusListeningInput()

	case key.Matches(msg, configMenuKeys.Select):
		switch m.subCursor {
		case 0:
			m.cdpListen = !m.cdpListen
		case 1:
			m.lldpListen = !m.lldpListen
		case 2:
			m.filterRouter = !m.filterRouter
		case 3:
			m.filterBridge = !m.filterBridge
		case 4:
			m.filterStation = !m.filterStation
		case 7: // Back
			m.subState = SubStateMain
			m.stalenessInput.Blur()
			m.staleRemovalInput.Blur()
		}

	default:
		// Pass to text input if focused
		if m.subCursor == 5 {
			var cmd tea.Cmd
			m.stalenessInput, cmd = m.stalenessInput.Update(msg)
			return m, cmd
		} else if m.subCursor == 6 {
			var cmd tea.Cmd
			m.staleRemovalInput, cmd = m.staleRemovalInput.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

// findListeningPosition returns the row and column position for the current cursor
func (m *ConfigMenuModel) findListeningPosition(rows [][]int) (row, col int) {
	return findRowPosition(m.subCursor, rows)
}

func (m *ConfigMenuModel) focusListeningInput() {
	m.stalenessInput.Blur()
	m.staleRemovalInput.Blur()
	if m.subCursor == 5 {
		m.stalenessInput.Focus()
	} else if m.subCursor == 6 {
		m.staleRemovalInput.Focus()
	}
}

// renderListening renders the Listening Options sub-menu
func (m ConfigMenuModel) renderListening() string {
	theme := DefaultTheme
	var b strings.Builder

	sectionStyle := lipgloss.NewStyle().Foreground(theme.Base0D).Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(theme.Base03)

	b.WriteString("\n")

	// Protocol Listening
	b.WriteString("  ")
	b.WriteString(sectionStyle.Render("Protocol Listening"))
	b.WriteString("\n\n")

	// CDP Listen / LLDP Listen (same row)
	b.WriteString("  ")
	b.WriteString(renderCursor(m.subCursor == 0, theme))
	b.WriteString(renderCheckbox(m.cdpListen, m.subCursor == 0, theme))
	b.WriteString(" ")
	b.WriteString(renderLabel("CDP", m.subCursor == 0, theme))
	b.WriteString("     ")

	// LLDP Listen
	b.WriteString(renderCursor(m.subCursor == 1, theme))
	b.WriteString(renderCheckbox(m.lldpListen, m.subCursor == 1, theme))
	b.WriteString(" ")
	b.WriteString(renderLabel("LLDP", m.subCursor == 1, theme))
	b.WriteString("\n\n")

	// Filter by Capabilities
	b.WriteString("  ")
	b.WriteString(sectionStyle.Render("Filter by Capabilities"))
	b.WriteString(" ")
	b.WriteString(dimStyle.Render("(empty = show all)"))
	b.WriteString("\n\n")

	// Filter Router / Bridge / Station (same row)
	b.WriteString("  ")
	b.WriteString(renderCursor(m.subCursor == 2, theme))
	b.WriteString(renderCheckbox(m.filterRouter, m.subCursor == 2, theme))
	b.WriteString(" ")
	b.WriteString(renderLabel("Router", m.subCursor == 2, theme))
	b.WriteString("  ")

	// Filter Bridge
	b.WriteString(renderCursor(m.subCursor == 3, theme))
	b.WriteString(renderCheckbox(m.filterBridge, m.subCursor == 3, theme))
	b.WriteString(" ")
	b.WriteString(renderLabel("Bridge", m.subCursor == 3, theme))
	b.WriteString("  ")

	// Filter Station
	b.WriteString(renderCursor(m.subCursor == 4, theme))
	b.WriteString(renderCheckbox(m.filterStation, m.subCursor == 4, theme))
	b.WriteString(" ")
	b.WriteString(renderLabel("Station", m.subCursor == 4, theme))
	b.WriteString("\n\n")

	// Display Settings
	b.WriteString("  ")
	b.WriteString(sectionStyle.Render("Display Settings"))
	b.WriteString("\n\n")

	// Staleness Timeout
	b.WriteString("  ")
	b.WriteString(renderCursor(m.subCursor == 5, theme))
	b.WriteString(renderLabel("Staleness Timeout", m.subCursor == 5, theme))
	b.WriteString("  ")
	b.WriteString(m.stalenessInput.View())
	b.WriteString(dimStyle.Render(" seconds (gray out)"))
	b.WriteString("\n")

	// Stale Removal
	b.WriteString("  ")
	b.WriteString(renderCursor(m.subCursor == 6, theme))
	b.WriteString(renderLabel("Stale Removal", m.subCursor == 6, theme))
	b.WriteString("      ")
	b.WriteString(m.staleRemovalInput.View())
	b.WriteString(dimStyle.Render(" seconds (0 = never)"))
	b.WriteString("\n\n")

	// Back button
	b.WriteString("  ")
	b.WriteString(renderCursor(m.subCursor == 7, theme))
	b.WriteString(renderLabel("[Back]", m.subCursor == 7, theme))
	b.WriteString("\n")

	return b.String()
}
