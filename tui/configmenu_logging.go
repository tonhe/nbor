package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// updateLogging handles key events for the Logging Options sub-menu
func (m ConfigMenuModel) updateLogging(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Logging sub-menu fields:
	// 0: Enable Logging toggle
	// 1: Log Directory (text)
	// 2: Back button
	const maxLoggingFields = 3

	switch {
	case key.Matches(msg, configMenuKeys.Back):
		m.subState = SubStateMain
		m.logDirInput.Blur()

	case key.Matches(msg, configMenuKeys.Up):
		m.logDirInput.Blur()
		m.subCursor--
		if m.subCursor < 0 {
			m.subCursor = maxLoggingFields - 1
		}
		if m.subCursor == 1 {
			m.logDirInput.Focus()
		}

	case key.Matches(msg, configMenuKeys.Down), key.Matches(msg, configMenuKeys.Tab):
		m.logDirInput.Blur()
		m.subCursor++
		if m.subCursor >= maxLoggingFields {
			m.subCursor = 0
		}
		if m.subCursor == 1 {
			m.logDirInput.Focus()
		}

	case key.Matches(msg, configMenuKeys.Select):
		switch m.subCursor {
		case 0:
			m.loggingEnabled = !m.loggingEnabled
		case 2: // Back
			m.subState = SubStateMain
			m.logDirInput.Blur()
		}

	default:
		// Pass to text input if focused
		if m.subCursor == 1 {
			var cmd tea.Cmd
			m.logDirInput, cmd = m.logDirInput.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

// renderLogging renders the Logging Options sub-menu
func (m ConfigMenuModel) renderLogging() string {
	theme := DefaultTheme
	var b strings.Builder

	sectionStyle := lipgloss.NewStyle().Foreground(theme.Base0D).Bold(true)
	labelStyle := lipgloss.NewStyle().Foreground(theme.Base05)
	focusedStyle := lipgloss.NewStyle().Foreground(theme.Base0B).Bold(true)
	cursorStyle := lipgloss.NewStyle().Foreground(theme.Base0C).Bold(true)

	checkbox := func(checked, focused bool) string {
		style := labelStyle
		if focused {
			style = focusedStyle
		}
		if checked {
			return style.Render("[x]")
		}
		return style.Render("[ ]")
	}

	cursor := func(focused bool) string {
		if focused {
			return cursorStyle.Render(">") + " "
		}
		return "  "
	}

	b.WriteString("\n")

	// Logging Settings
	b.WriteString("  ")
	b.WriteString(sectionStyle.Render("Logging Settings"))
	b.WriteString("\n\n")

	// Enable Logging
	b.WriteString("  ")
	b.WriteString(cursor(m.subCursor == 0))
	b.WriteString(checkbox(m.loggingEnabled, m.subCursor == 0))
	b.WriteString(" ")
	if m.subCursor == 0 {
		b.WriteString(focusedStyle.Render("Enable Logging"))
	} else {
		b.WriteString(labelStyle.Render("Enable Logging"))
	}
	b.WriteString("\n\n")

	// Log Directory
	b.WriteString("  ")
	b.WriteString(cursor(m.subCursor == 1))
	if m.subCursor == 1 {
		b.WriteString(focusedStyle.Render("Log Directory"))
	} else {
		b.WriteString(labelStyle.Render("Log Directory"))
	}
	b.WriteString("\n  ")
	b.WriteString("  ")
	b.WriteString(m.logDirInput.View())
	b.WriteString("\n\n")

	// Back button
	b.WriteString("  ")
	b.WriteString(cursor(m.subCursor == 2))
	if m.subCursor == 2 {
		b.WriteString(focusedStyle.Render("[Back]"))
	} else {
		b.WriteString(labelStyle.Render("[Back]"))
	}
	b.WriteString("\n")

	return b.String()
}
