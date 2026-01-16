package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"nbor/types"
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

	var b strings.Builder

	// Title
	title := m.styles.PickerTitle.Render("Select a network interface:")
	b.WriteString(title)
	b.WriteString("\n\n")

	if len(m.interfaces) == 0 {
		b.WriteString(m.styles.StatusError.Render("No suitable Ethernet interfaces found."))
		b.WriteString("\n\n")
		b.WriteString(m.styles.StatusInfo.Render("Make sure you have wired network adapters available."))
		return b.String()
	}

	// Interface list
	for i, iface := range m.interfaces {
		cursor := "  "
		style := m.styles.PickerItem

		if i == m.cursor {
			cursor = "> "
			style = m.styles.PickerSelected
		}

		// Format interface line with colored status dot
		var status string
		if iface.IsUp {
			status = m.styles.PickerUp.Render("●") // Green
		} else {
			status = m.styles.PickerDown.Render("●") // Red/gray
		}

		// Format MAC
		mac := ""
		if iface.MAC != nil {
			mac = iface.MAC.String()
		}

		// Format speed
		speed := ""
		if iface.Speed != "" {
			speed = fmt.Sprintf(" [%s]", iface.Speed)
		}

		// Format IP addresses
		ips := iface.FormatIPs()
		ipDisplay := ""
		if ips != "" {
			ipDisplay = fmt.Sprintf(" (%s)", ips)
		}

		line := fmt.Sprintf("%s%s %s  %s%s%s",
			cursor,
			status,
			iface.Name,
			mac,
			speed,
			ipDisplay,
		)

		b.WriteString(style.Render(line))
		b.WriteString("\n")
	}

	// Help text
	b.WriteString("\n")
	helpStyle := lipgloss.NewStyle().Foreground(m.styles.Footer.GetForeground())
	b.WriteString(helpStyle.Render("↑/↓ or j/k to navigate • enter to select • q to quit"))

	return b.String()
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
