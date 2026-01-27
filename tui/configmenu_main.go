package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// updateMain handles key events for the main menu
func (m ConfigMenuModel) updateMain(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, configMenuKeys.Up):
		m.mainCursor--
		if m.mainCursor < 0 {
			m.mainCursor = int(ConfigMenuItemCount) - 1
		}

	case key.Matches(msg, configMenuKeys.Down), key.Matches(msg, configMenuKeys.Tab):
		m.mainCursor++
		if m.mainCursor >= int(ConfigMenuItemCount) {
			m.mainCursor = 0
		}

	case key.Matches(msg, configMenuKeys.Select):
		switch ConfigMenuItem(m.mainCursor) {
		case ConfigMenuChangeInterface:
			return m, func() tea.Msg { return ChangeInterfaceMsg{} }
		case ConfigMenuListening:
			m.subState = SubStateListening
			m.subCursor = 0
		case ConfigMenuBroadcast:
			m.subState = SubStateBroadcast
			m.subCursor = 0
			m.systemNameInput.Focus()
		case ConfigMenuLogging:
			m.subState = SubStateLogging
			m.subCursor = 0
		case ConfigMenuTheme:
			m.subState = SubStateTheme
			m.subCursor = m.themeIndex
			m.previousTheme = DefaultTheme
		case ConfigMenuAbout:
			m.subState = SubStateAbout
		case ConfigMenuSaveExit:
			return m.saveConfig()
		case ConfigMenuCancel:
			// Revert theme if it was changed
			if m.themePreviewDirty {
				SetTheme(m.previousTheme)
			}
			return m, func() tea.Msg { return ConfigCancelledMsg{} }
		}

	// ESC does nothing on main menu - must select Cancel
	case key.Matches(msg, configMenuKeys.Back):
		// Do nothing
	}

	return m, nil
}

// renderMainMenu renders the main configuration menu
func (m ConfigMenuModel) renderMainMenu() string {
	theme := DefaultTheme
	var b strings.Builder

	labelStyle := lipgloss.NewStyle().Foreground(theme.Base05)
	focusedStyle := lipgloss.NewStyle().Foreground(theme.Base0B).Bold(true)
	cursorStyle := lipgloss.NewStyle().Foreground(theme.Base0C).Bold(true)

	b.WriteString("\n")

	for i, label := range mainMenuLabels {
		focused := i == m.mainCursor
		b.WriteString("  ")

		if focused {
			b.WriteString(cursorStyle.Render(">"))
			b.WriteString(" ")
			b.WriteString(focusedStyle.Render("[" + label + "]"))
		} else {
			b.WriteString("  ")
			b.WriteString(labelStyle.Render("[" + label + "]"))
		}
		b.WriteString("\n")

		// Add spacing after groups
		if i == 0 || i == 5 {
			b.WriteString("\n")
		}
	}

	return b.String()
}
