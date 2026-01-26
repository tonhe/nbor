package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"nbor/types"
	"nbor/version"
)

// InterfacePickerModel is the model for the interface selection screen
type InterfacePickerModel struct {
	interfaces []types.InterfaceInfo
	cursor     int
	width      int
	height     int
	styles     Styles
	err        error
}

// NewInterfacePicker creates a new interface picker model
func NewInterfacePicker(interfaces []types.InterfaceInfo) InterfacePickerModel {
	// Sort interfaces: up with IPs first, then up without IPs, then down
	sortInterfaces(interfaces)

	return InterfacePickerModel{
		interfaces: interfaces,
		cursor:     0,
		styles:     DefaultStyles,
	}
}

// sortInterfaces sorts interfaces by priority:
// 1. Up with IPv4 address
// 2. Up with IPv6 (non-link-local) address
// 3. Up without IP
// 4. Down
func sortInterfaces(interfaces []types.InterfaceInfo) {
	sort.Slice(interfaces, func(i, j int) bool {
		// Calculate priority score (lower is better)
		scoreI := interfacePriority(interfaces[i])
		scoreJ := interfacePriority(interfaces[j])

		if scoreI != scoreJ {
			return scoreI < scoreJ
		}

		// Same priority, sort by name
		return interfaces[i].Name < interfaces[j].Name
	})
}

// interfacePriority returns a priority score for sorting (lower = higher priority)
func interfacePriority(iface types.InterfaceInfo) int {
	if !iface.IsUp {
		return 100 // Down interfaces last
	}

	if len(iface.IPv4Addrs) > 0 {
		return 0 // Up with IPv4 = highest priority
	}

	if len(iface.IPv6Addrs) > 0 {
		return 10 // Up with IPv6 = second priority
	}

	return 50 // Up without IP = third priority
}

// Init initializes the interface picker
func (m InterfacePickerModel) Init() tea.Cmd {
	return nil
}

// InterfaceSelectedMsg is sent when an interface is selected
type InterfaceSelectedMsg struct {
	Interface types.InterfaceInfo
}

// interfacePickerKeyMap defines the key bindings for the interface picker
type interfacePickerKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Select key.Binding
	Quit   key.Binding
}

var interfaceKeys = interfacePickerKeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Select: key.NewBinding(
		key.WithKeys("enter", " "),
		key.WithHelp("enter", "select"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c", "q"),
		key.WithHelp("ctrl+c/q", "quit"),
	),
}

// Update handles messages for the interface picker
func (m InterfacePickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, interfaceKeys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, interfaceKeys.Down):
			if m.cursor < len(m.interfaces)-1 {
				m.cursor++
			}
		case key.Matches(msg, interfaceKeys.Select):
			if len(m.interfaces) > 0 {
				return m, func() tea.Msg {
					return InterfaceSelectedMsg{Interface: m.interfaces[m.cursor]}
				}
			}
		case key.Matches(msg, interfaceKeys.Quit):
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// View renders the interface picker
func (m InterfacePickerModel) View() string {
	if m.err != nil {
		return m.styles.StatusError.Render(fmt.Sprintf("Error: %v", m.err))
	}

	header := m.renderHeader()
	content := m.renderContent()
	footer := m.renderFooter()

	// Calculate spacing to push footer to bottom
	headerLines := strings.Count(header, "\n") + 1
	contentLines := strings.Count(content, "\n") + 1
	footerLines := 1

	usedLines := headerLines + contentLines + footerLines
	remainingLines := m.height - usedLines
	if remainingLines < 0 {
		remainingLines = 0
	}

	var b strings.Builder
	b.WriteString(header)
	b.WriteString("\n")
	b.WriteString(content)
	b.WriteString(strings.Repeat("\n", remainingLines))
	b.WriteString(footer)

	return b.String()
}

// renderHeader renders the header bar
func (m InterfacePickerModel) renderHeader() string {
	theme := DefaultTheme
	bg := theme.Base01

	sp := lipgloss.NewStyle().Background(bg).Render(" ")

	nameStyle := lipgloss.NewStyle().
		Foreground(theme.Base0C).
		Background(bg).
		Bold(true)
	versionStyle := lipgloss.NewStyle().
		Foreground(theme.Base03).
		Background(bg)
	leftPart := nameStyle.Render("nbor") + sp + versionStyle.Render("v"+version.Version)

	titleStyle := lipgloss.NewStyle().
		Foreground(theme.Base0D).
		Background(bg).
		Bold(true)
	rightPart := titleStyle.Render("Select Interface")

	leftLen := lipgloss.Width(leftPart)
	rightLen := lipgloss.Width(rightPart)
	availableWidth := m.width - 2
	gap := availableWidth - leftLen - rightLen
	if gap < 1 {
		gap = 1
	}

	spaceStyle := lipgloss.NewStyle().Background(bg)
	headerContent := leftPart + spaceStyle.Render(strings.Repeat(" ", gap)) + rightPart

	headerStyle := lipgloss.NewStyle().
		Background(bg).
		Padding(0, 1).
		Width(m.width)

	return headerStyle.Render(headerContent)
}

// renderContent renders the interface list
func (m InterfacePickerModel) renderContent() string {
	theme := DefaultTheme
	var b strings.Builder

	b.WriteString("\n")

	if len(m.interfaces) == 0 {
		errorStyle := lipgloss.NewStyle().Foreground(theme.Base08)
		infoStyle := lipgloss.NewStyle().Foreground(theme.Base03)
		b.WriteString("  ")
		b.WriteString(errorStyle.Render("No suitable Ethernet interfaces found."))
		b.WriteString("\n\n  ")
		b.WriteString(infoStyle.Render("Make sure you have wired network adapters available."))
		return b.String()
	}

	selectedStyle := lipgloss.NewStyle().
		Foreground(theme.Base0B).
		Bold(true)
	normalStyle := lipgloss.NewStyle().
		Foreground(theme.Base05)
	dimStyle := lipgloss.NewStyle().
		Foreground(theme.Base03)
	cursorStyle := lipgloss.NewStyle().
		Foreground(theme.Base0C).
		Bold(true)
	upStyle := lipgloss.NewStyle().
		Foreground(theme.Base0B)
	downStyle := lipgloss.NewStyle().
		Foreground(theme.Base03)

	for i, iface := range m.interfaces {
		// Status dot
		var status string
		if iface.IsUp {
			status = upStyle.Render("●")
		} else {
			status = downStyle.Render("●")
		}

		// Format MAC
		mac := ""
		if iface.MAC != nil {
			mac = iface.MAC.String()
		}

		// Format speed
		speed := ""
		if iface.Speed != "" {
			speed = fmt.Sprintf("[%s]", iface.Speed)
		}

		// Format IP addresses
		ips := iface.FormatIPs()
		ipDisplay := ""
		if ips != "" {
			ipDisplay = fmt.Sprintf("(%s)", ips)
		}

		if i == m.cursor {
			b.WriteString("  ")
			b.WriteString(cursorStyle.Render(">"))
			b.WriteString(" ")
			b.WriteString(status)
			b.WriteString(" ")
			b.WriteString(selectedStyle.Render(iface.Name))
			b.WriteString("  ")
			b.WriteString(dimStyle.Render(mac))
			if speed != "" {
				b.WriteString(" ")
				b.WriteString(dimStyle.Render(speed))
			}
			if ipDisplay != "" {
				b.WriteString(" ")
				b.WriteString(dimStyle.Render(ipDisplay))
			}
		} else {
			b.WriteString("    ")
			b.WriteString(status)
			b.WriteString(" ")
			b.WriteString(normalStyle.Render(iface.Name))
			b.WriteString("  ")
			b.WriteString(dimStyle.Render(mac))
			if speed != "" {
				b.WriteString(" ")
				b.WriteString(dimStyle.Render(speed))
			}
			if ipDisplay != "" {
				b.WriteString(" ")
				b.WriteString(dimStyle.Render(ipDisplay))
			}
		}
		b.WriteString("\n")
	}

	return b.String()
}

// renderFooter renders the footer bar
func (m InterfacePickerModel) renderFooter() string {
	theme := DefaultTheme
	bg := theme.Base01

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

	sep := sepStyle.Render(" │ ")

	footerContent := keyStyle.Render("↑/↓") + textStyle.Render(" navigate") + sep +
		keyStyle.Render("enter") + textStyle.Render(" select") + sep +
		keyStyle.Render("q") + textStyle.Render(" quit")

	contentLen := lipgloss.Width(footerContent)
	availableWidth := m.width - 2
	gap := availableWidth - contentLen
	if gap < 0 {
		gap = 0
	}

	spaceStyle := lipgloss.NewStyle().Background(bg)
	footerContent = footerContent + spaceStyle.Render(strings.Repeat(" ", gap))

	footerStyle := lipgloss.NewStyle().
		Background(bg).
		Padding(0, 1).
		Width(m.width)

	return footerStyle.Render(footerContent)
}

// SetError sets an error to display
func (m *InterfacePickerModel) SetError(err error) {
	m.err = err
}

// SelectedInterface returns the currently highlighted interface
func (m InterfacePickerModel) SelectedInterface() *types.InterfaceInfo {
	if len(m.interfaces) == 0 {
		return nil
	}
	return &m.interfaces[m.cursor]
}
