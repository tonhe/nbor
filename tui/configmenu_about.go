package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"nbor/version"
)

// updateAbout handles key events for the About screen
func (m ConfigMenuModel) updateAbout(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Any key returns to main menu
	if key.Matches(msg, configMenuKeys.Back) || key.Matches(msg, configMenuKeys.Select) {
		m.subState = SubStateMain
	}
	return m, nil
}

// renderAbout renders the About screen
func (m ConfigMenuModel) renderAbout() string {
	theme := DefaultTheme
	var b strings.Builder

	// Styles for the ASCII art logo - gradient effect using theme colors
	logoStyle1 := lipgloss.NewStyle().Foreground(theme.Base0D).Bold(true) // Blue
	logoStyle2 := lipgloss.NewStyle().Foreground(theme.Base0C).Bold(true) // Cyan
	logoStyle3 := lipgloss.NewStyle().Foreground(theme.Base0B).Bold(true) // Green

	labelStyle := lipgloss.NewStyle().Foreground(theme.Base05)
	valueStyle := lipgloss.NewStyle().Foreground(theme.Base0B)
	dimStyle := lipgloss.NewStyle().Foreground(theme.Base03)
	linkStyle := lipgloss.NewStyle().Foreground(theme.Base0C)
	authorStyle := lipgloss.NewStyle().Foreground(theme.Base0E) // Purple/Magenta

	b.WriteString("\n")

	// ASCII art logo with gradient coloring
	logoLines := []string{
		"███╗   ██╗",
		"████╗  ██║",
		"██╔██╗ ██║",
		"██║╚██╗██║",
		"██║ ╚████║",
		"╚═╝  ╚═══╝",
	}
	logoLines2 := []string{
		"██████╗ ",
		"██╔══██╗",
		"██████╔╝",
		"██╔══██╗",
		"██████╔╝",
		"╚═════╝ ",
	}
	logoLines3 := []string{
		" ██████╗ ",
		"██╔═══██╗",
		"██║   ██║",
		"██║   ██║",
		"╚██████╔╝",
		" ╚═════╝ ",
	}
	logoLines4 := []string{
		"██████╗ ",
		"██╔══██╗",
		"██████╔╝",
		"██╔══██╗",
		"██║  ██║",
		"╚═╝  ╚═╝",
	}

	for i := 0; i < 6; i++ {
		b.WriteString("  ")
		b.WriteString(logoStyle1.Render(logoLines[i]))
		b.WriteString(logoStyle2.Render(logoLines2[i]))
		b.WriteString(logoStyle2.Render(logoLines3[i]))
		b.WriteString(logoStyle3.Render(logoLines4[i]))
		b.WriteString("\n")
	}

	// Version under logo
	b.WriteString("  ")
	b.WriteString(dimStyle.Render("              v" + version.Version))
	b.WriteString("\n\n")

	// Description
	b.WriteString("  ")
	b.WriteString(labelStyle.Render("Network neighbor discovery for CDP and LLDP"))
	b.WriteString("\n\n")

	// Author
	b.WriteString("  ")
	b.WriteString(dimStyle.Render("Author:"))
	b.WriteString(" ")
	b.WriteString(authorStyle.Render("Tony Mattke"))
	b.WriteString("\n")

	// GitHub
	b.WriteString("  ")
	b.WriteString(dimStyle.Render("GitHub:"))
	b.WriteString(" ")
	b.WriteString(linkStyle.Render("github.com/tonhe/nbor"))
	b.WriteString("\n\n")

	// Current theme
	b.WriteString("  ")
	b.WriteString(dimStyle.Render("Theme:"))
	b.WriteString("  ")
	b.WriteString(valueStyle.Render(DefaultTheme.Name))
	b.WriteString("\n\n")

	// Press any key
	b.WriteString("  ")
	b.WriteString(dimStyle.Render("Press Esc or Enter to return"))
	b.WriteString("\n")

	return b.String()
}
