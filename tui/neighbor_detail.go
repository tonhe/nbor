package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"nbor/types"
)

// renderDetailPopupOverlay renders a centered popup overlaid on the base view
func (m NeighborTableModel) renderDetailPopupOverlay(n *types.Neighbor, baseView string) string {
	theme := DefaultTheme
	bg := theme.Base00

	// Popup dimensions
	popupWidth := 50
	if m.width > 0 && m.width < popupWidth+4 {
		popupWidth = m.width - 4
	}
	contentWidth := popupWidth - 4 // Account for border and padding

	// All styles include background for consistent appearance
	titleStyle := lipgloss.NewStyle().
		Foreground(theme.Base0D).
		Background(bg).
		Bold(true).
		Width(contentWidth).
		Align(lipgloss.Center)

	labelStyle := lipgloss.NewStyle().
		Foreground(theme.Base04).
		Background(bg).
		Width(14)

	valueStyle := lipgloss.NewStyle().
		Foreground(theme.Base0B).
		Background(bg)

	dimValueStyle := lipgloss.NewStyle().
		Foreground(theme.Base03).
		Background(bg)

	staleStyle := lipgloss.NewStyle().
		Foreground(theme.Base08).
		Background(bg).
		Bold(true)

	hintStyle := lipgloss.NewStyle().
		Foreground(theme.Base03).
		Background(bg).
		Width(contentWidth).
		Align(lipgloss.Center)

	separatorStyle := lipgloss.NewStyle().
		Foreground(theme.Base02).
		Background(bg)

	blankLineStyle := lipgloss.NewStyle().
		Background(bg).
		Width(contentWidth)

	// Build content
	var b strings.Builder

	// Title
	title := n.Hostname
	if title == "" {
		title = "Unknown Device"
	}
	if n.IsStale {
		title += " " + staleStyle.Render("(stale)")
	}
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n")
	b.WriteString(separatorStyle.Render(strings.Repeat("─", contentWidth)))
	b.WriteString("\n")
	b.WriteString(blankLineStyle.Render(""))
	b.WriteString("\n")

	// Helper to render a row with full-width background
	renderRow := func(label, value string) {
		labelRendered := labelStyle.Render(label)
		var valueRendered string
		if value == "" {
			valueRendered = dimValueStyle.Render("—")
		} else {
			valueRendered = valueStyle.Render(value)
		}
		// Calculate padding to fill the row
		usedWidth := lipgloss.Width(labelRendered) + lipgloss.Width(valueRendered)
		padding := ""
		if usedWidth < contentWidth {
			paddingStyle := lipgloss.NewStyle().Background(bg)
			padding = paddingStyle.Render(strings.Repeat(" ", contentWidth-usedWidth))
		}
		b.WriteString(labelRendered)
		b.WriteString(valueRendered)
		b.WriteString(padding)
		b.WriteString("\n")
	}

	// Helper for blank line with background
	blankLine := func() {
		b.WriteString(blankLineStyle.Render(""))
		b.WriteString("\n")
	}

	// Device Identity
	renderRow("Device ID:", n.ID)
	renderRow("Port:", formatPortInfo(n))
	renderRow("Protocol:", string(n.Protocol))
	blankLine()

	// Network Info
	mgmtIP := ""
	if n.ManagementIP != nil {
		mgmtIP = n.ManagementIP.String()
	}
	renderRow("Mgmt IP:", mgmtIP)

	srcMAC := ""
	if n.SourceMAC != nil {
		srcMAC = n.SourceMAC.String()
	}
	renderRow("Source MAC:", srcMAC)
	blankLine()

	// Platform Info
	renderRow("Platform:", truncateValue(n.Platform, contentWidth-15))
	renderRow("Description:", truncateValue(n.Description, contentWidth-15))
	renderRow("Location:", truncateValue(n.Location, contentWidth-15))
	blankLine()

	// Capabilities
	caps := formatCapabilitiesList(n.Capabilities)
	renderRow("Capabilities:", caps)
	blankLine()

	// Timing Info
	renderRow("First Seen:", formatTime(n.FirstSeen))
	renderRow("Last Seen:", formatLastSeen(n.LastSeen))
	renderRow("Interface:", n.Interface)

	blankLine()
	b.WriteString(hintStyle.Render("Press ESC or Enter to close"))

	// Apply border style
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Base0D).
		BorderBackground(bg).
		Background(bg).
		Padding(0, 1).
		Width(popupWidth)

	popup := borderStyle.Render(b.String())

	// Overlay popup on top of the base view
	return overlayPopup(baseView, popup, m.width, m.height)
}

// overlayPopup places a popup in the center of the base view
func overlayPopup(base, popup string, width, height int) string {
	baseLines := strings.Split(base, "\n")
	popupLines := strings.Split(popup, "\n")

	popupHeight := len(popupLines)
	popupWidth := lipgloss.Width(popupLines[0])

	// Calculate position to center the popup
	startRow := (height - popupHeight) / 2
	startCol := (width - popupWidth) / 2

	if startRow < 0 {
		startRow = 0
	}
	if startCol < 0 {
		startCol = 0
	}

	// Ensure we have enough lines
	for len(baseLines) < height {
		baseLines = append(baseLines, "")
	}

	// Overlay the popup onto the base
	for i, popupLine := range popupLines {
		row := startRow + i
		if row >= len(baseLines) {
			break
		}

		baseLine := baseLines[row]
		// Pad base line if needed
		baseRunes := []rune(baseLine)
		for len(baseRunes) < startCol+popupWidth {
			baseRunes = append(baseRunes, ' ')
		}

		// Insert popup line
		newLine := string(baseRunes[:startCol]) + popupLine
		if startCol+popupWidth < len(baseRunes) {
			newLine += string(baseRunes[startCol+popupWidth:])
		}
		baseLines[row] = newLine
	}

	return strings.Join(baseLines[:height], "\n")
}

// formatPortInfo formats port ID and description
func formatPortInfo(n *types.Neighbor) string {
	if n.PortDescription != "" && n.PortDescription != n.PortID {
		return n.PortID + " (" + n.PortDescription + ")"
	}
	return n.PortID
}

// formatCapabilitiesList formats capabilities as a comma-separated string
func formatCapabilitiesList(caps []types.Capability) string {
	if len(caps) == 0 {
		return ""
	}
	var strs []string
	for _, c := range caps {
		strs = append(strs, string(c))
	}
	return strings.Join(strs, ", ")
}

// formatTime formats a time for display
func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

// formatLastSeen formats the last seen time as relative duration
func formatLastSeen(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	duration := time.Since(t)
	if duration < time.Minute {
		return fmt.Sprintf("%ds ago", int(duration.Seconds()))
	} else if duration < time.Hour {
		return fmt.Sprintf("%dm ago", int(duration.Minutes()))
	} else if duration < 24*time.Hour {
		return fmt.Sprintf("%dh %dm ago", int(duration.Hours()), int(duration.Minutes())%60)
	}
	return t.Format("2006-01-02 15:04")
}

// truncateValue truncates a string to fit within maxWidth
func truncateValue(s string, maxWidth int) string {
	if maxWidth <= 3 {
		return s
	}
	if lipgloss.Width(s) <= maxWidth {
		return s
	}
	// Truncate by runes
	runes := []rune(s)
	result := ""
	for _, r := range runes {
		if lipgloss.Width(result+string(r)) > maxWidth-3 {
			break
		}
		result += string(r)
	}
	return result + "..."
}
