package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"nbor/types"
)

// renderDetailPopup renders a centered popup with all neighbor information
func (m NeighborTableModel) renderDetailPopup(n *types.Neighbor) string {
	theme := DefaultTheme

	// Popup dimensions
	popupWidth := 50
	if m.width > 0 && m.width < popupWidth+4 {
		popupWidth = m.width - 4
	}
	contentWidth := popupWidth - 4 // Account for border and padding

	// Styles
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Base0D).
		Background(theme.Base00).
		Padding(0, 1).
		Width(popupWidth)

	titleStyle := lipgloss.NewStyle().
		Foreground(theme.Base0D).
		Bold(true).
		Width(contentWidth).
		Align(lipgloss.Center)

	labelStyle := lipgloss.NewStyle().
		Foreground(theme.Base04).
		Width(14)

	valueStyle := lipgloss.NewStyle().
		Foreground(theme.Base0B)

	dimValueStyle := lipgloss.NewStyle().
		Foreground(theme.Base03)

	staleStyle := lipgloss.NewStyle().
		Foreground(theme.Base08).
		Bold(true)

	hintStyle := lipgloss.NewStyle().
		Foreground(theme.Base03).
		Width(contentWidth).
		Align(lipgloss.Center)

	separatorStyle := lipgloss.NewStyle().
		Foreground(theme.Base02)

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
	b.WriteString("\n\n")

	// Helper to render a row
	renderRow := func(label, value string) {
		if value == "" {
			value = dimValueStyle.Render("—")
		} else {
			value = valueStyle.Render(value)
		}
		b.WriteString(labelStyle.Render(label))
		b.WriteString(value)
		b.WriteString("\n")
	}

	// Device Identity
	renderRow("Device ID:", n.ID)
	renderRow("Port:", formatPortInfo(n))
	renderRow("Protocol:", string(n.Protocol))
	b.WriteString("\n")

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
	b.WriteString("\n")

	// Platform Info
	renderRow("Platform:", truncateValue(n.Platform, contentWidth-15))
	renderRow("Description:", truncateValue(n.Description, contentWidth-15))
	renderRow("Location:", truncateValue(n.Location, contentWidth-15))
	b.WriteString("\n")

	// Capabilities
	caps := formatCapabilitiesList(n.Capabilities)
	renderRow("Capabilities:", caps)
	b.WriteString("\n")

	// Timing Info
	renderRow("First Seen:", formatTime(n.FirstSeen))
	renderRow("Last Seen:", formatLastSeen(n.LastSeen))
	renderRow("Interface:", n.Interface)

	b.WriteString("\n")
	b.WriteString(hintStyle.Render("Press ESC or Enter to close"))

	popup := borderStyle.Render(b.String())

	// Center the popup on screen
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		popup,
		lipgloss.WithWhitespaceBackground(theme.Base00),
	)
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
