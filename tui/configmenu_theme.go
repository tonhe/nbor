package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// updateTheme handles key events for the Change Theme sub-menu
func (m ConfigMenuModel) updateTheme(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	themeCount := GetThemeCount()

	switch {
	case key.Matches(msg, configMenuKeys.Back):
		// Revert to previous theme
		SetTheme(m.previousTheme)
		m.themePreviewDirty = false
		m.subState = SubStateMain

	case key.Matches(msg, configMenuKeys.Up):
		m.subCursor--
		if m.subCursor < 0 {
			m.subCursor = themeCount - 1
		}
		m.previewTheme()

	case key.Matches(msg, configMenuKeys.Down), key.Matches(msg, configMenuKeys.Tab):
		m.subCursor++
		if m.subCursor >= themeCount {
			m.subCursor = 0
		}
		m.previewTheme()

	case key.Matches(msg, configMenuKeys.Select):
		// Confirm theme selection - just update the index, don't modify config yet
		// Config will be updated when Save & Exit or Ctrl+S is pressed
		m.themeIndex = m.subCursor
		m.themePreviewDirty = true
		m.subState = SubStateMain
	}

	return m, nil
}

func (m *ConfigMenuModel) previewTheme() {
	_, _, theme := GetThemeByIndex(m.subCursor)
	if theme != nil {
		SetTheme(*theme)
	}
}

// renderTheme renders the Change Theme sub-menu
func (m ConfigMenuModel) renderTheme() string {
	theme := DefaultTheme
	var b strings.Builder

	labelStyle := lipgloss.NewStyle().Foreground(theme.Base05)
	focusedStyle := lipgloss.NewStyle().Foreground(theme.Base0B).Bold(true)
	cursorStyle := lipgloss.NewStyle().Foreground(theme.Base0C).Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(theme.Base03)

	b.WriteString("\n")
	b.WriteString("  ")
	b.WriteString(dimStyle.Render("Use ↑/↓ to preview, Enter to select, Esc to cancel"))
	b.WriteString("\n\n")

	themes := ListThemes()

	// Calculate visible range (show ~15 themes at a time)
	visibleCount := 15
	if m.height > 0 {
		visibleCount = m.height - 8 // Account for header, footer, instructions
		if visibleCount < 5 {
			visibleCount = 5
		}
	}

	startIdx := m.subCursor - visibleCount/2
	if startIdx < 0 {
		startIdx = 0
	}
	if startIdx+visibleCount > len(themes) {
		startIdx = len(themes) - visibleCount
		if startIdx < 0 {
			startIdx = 0
		}
	}

	endIdx := startIdx + visibleCount
	if endIdx > len(themes) {
		endIdx = len(themes)
	}

	// Show scroll indicator if not at top
	if startIdx > 0 {
		b.WriteString("  ")
		b.WriteString(dimStyle.Render("  ↑ more themes above"))
		b.WriteString("\n")
	}

	for i := startIdx; i < endIdx; i++ {
		focused := i == m.subCursor
		_, name := themes[i][0], themes[i][1]

		b.WriteString("  ")
		if focused {
			b.WriteString(cursorStyle.Render(">"))
			b.WriteString(" ")
			b.WriteString(focusedStyle.Render(name))
			if i == m.themeIndex {
				b.WriteString(dimStyle.Render(" (current)"))
			}
		} else {
			b.WriteString("  ")
			if i == m.themeIndex {
				b.WriteString(labelStyle.Render(name))
				b.WriteString(dimStyle.Render(" (current)"))
			} else {
				b.WriteString(labelStyle.Render(name))
			}
		}
		b.WriteString("\n")
	}

	// Show scroll indicator if not at bottom
	if endIdx < len(themes) {
		b.WriteString("  ")
		b.WriteString(dimStyle.Render("  ↓ more themes below"))
		b.WriteString("\n")
	}

	return b.String()
}
