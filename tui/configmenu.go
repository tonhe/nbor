package tui

import (
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"nbor/config"
)

// ConfigField represents a field in the config menu
type ConfigField int

const (
	FieldChangeInterface ConfigField = iota
	FieldSystemName
	FieldSystemDescription
	FieldCDPListen
	FieldLLDPListen
	FieldCDPBroadcast
	FieldLLDPBroadcast
	FieldBroadcastOnStartup
	FieldInterval
	FieldTTL
	FieldCapRouter
	FieldCapBridge
	FieldCapStation
	FieldSave
	FieldCancel
	FieldCount // Used to track total number of fields
)

// ConfigMenuModel is the model for the configuration menu
type ConfigMenuModel struct {
	focusIndex int
	config     *config.Config

	// Text inputs
	systemNameInput textinput.Model
	systemDescInput textinput.Model
	intervalInput   textinput.Model
	ttlInput        textinput.Model

	// Toggle states (mirrors config but for local editing)
	cdpListen          bool
	lldpListen         bool
	cdpBroadcast       bool
	lldpBroadcast      bool
	broadcastOnStartup bool
	capRouter          bool
	capBridge          bool
	capStation         bool

	// Track original listen settings to detect changes
	originalCDPListen  bool
	originalLLDPListen bool

	// Resolved hostname (for display when SystemName is empty)
	resolvedHostname string

	width  int
	height int
	styles Styles
}

// NewConfigMenu creates a new config menu model
func NewConfigMenu(cfg *config.Config) ConfigMenuModel {
	// Resolve the actual hostname that will be used
	resolvedHostname := cfg.SystemName
	if resolvedHostname == "" {
		if hostname, err := os.Hostname(); err == nil {
			resolvedHostname = hostname
		} else {
			resolvedHostname = "nbor"
		}
	}

	// Create text inputs with minimal styling
	systemNameInput := textinput.New()
	systemNameInput.Placeholder = resolvedHostname // Show actual hostname as placeholder
	systemNameInput.CharLimit = 64
	systemNameInput.Width = 30
	systemNameInput.SetValue(cfg.SystemName)

	systemDescInput := textinput.New()
	systemDescInput.Placeholder = "nbor network tool"
	systemDescInput.CharLimit = 128
	systemDescInput.Width = 30
	systemDescInput.SetValue(cfg.SystemDescription)

	intervalInput := textinput.New()
	intervalInput.Placeholder = "5"
	intervalInput.CharLimit = 4
	intervalInput.Width = 6
	intervalInput.SetValue(strconv.Itoa(cfg.AdvertiseInterval))

	ttlInput := textinput.New()
	ttlInput.Placeholder = "20"
	ttlInput.CharLimit = 5
	ttlInput.Width = 6
	ttlInput.SetValue(strconv.Itoa(cfg.TTL))

	// Parse capabilities
	capRouter := false
	capBridge := false
	capStation := false
	for _, cap := range cfg.Capabilities {
		switch strings.ToLower(cap) {
		case "router":
			capRouter = true
		case "bridge":
			capBridge = true
		case "station":
			capStation = true
		}
	}

	m := ConfigMenuModel{
		focusIndex:         0,
		config:             cfg,
		systemNameInput:    systemNameInput,
		systemDescInput:    systemDescInput,
		intervalInput:      intervalInput,
		ttlInput:           ttlInput,
		cdpListen:          cfg.CDPListen,
		lldpListen:         cfg.LLDPListen,
		cdpBroadcast:       cfg.CDPBroadcast,
		lldpBroadcast:      cfg.LLDPBroadcast,
		broadcastOnStartup: cfg.BroadcastOnStartup,
		capRouter:          capRouter,
		capBridge:          capBridge,
		capStation:         capStation,
		originalCDPListen:  cfg.CDPListen,
		originalLLDPListen: cfg.LLDPListen,
		resolvedHostname:   resolvedHostname,
		styles:             DefaultStyles,
	}

	// Focus first field (change interface button)
	// No text input focused initially

	return m
}

// Init initializes the config menu
func (m ConfigMenuModel) Init() tea.Cmd {
	return textinput.Blink
}

// configMenuKeyMap defines the key bindings for the config menu
type configMenuKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Toggle key.Binding
	Save   key.Binding
	Cancel key.Binding
	Tab    key.Binding
}

var configMenuKeys = configMenuKeyMap{
	Up: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "down"),
	),
	Toggle: key.NewBinding(
		key.WithKeys(" ", "enter"),
		key.WithHelp("space/enter", "toggle"),
	),
	Save: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "save"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next field"),
	),
}

// ConfigSavedMsg is sent when config is saved
type ConfigSavedMsg struct {
	Config             *config.Config
	ListenSettingsChanged bool // True if CDP/LLDP listen settings changed (need new log file)
}

// ConfigCancelledMsg is sent when config editing is cancelled
type ConfigCancelledMsg struct{}

// ChangeInterfaceMsg is sent when user wants to change the interface
type ChangeInterfaceMsg struct{}

// Update handles messages for the config menu
func (m ConfigMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle escape/cancel first - always works regardless of field type
		if key.Matches(msg, configMenuKeys.Cancel) {
			return m, func() tea.Msg {
				return ConfigCancelledMsg{}
			}
		}

		// Handle save - always works
		if key.Matches(msg, configMenuKeys.Save) {
			return m.saveConfig()
		}

		// Handle navigation - up arrow always navigates
		if key.Matches(msg, configMenuKeys.Up) {
			m.focusIndex--
			if m.focusIndex < 0 {
				m.focusIndex = int(FieldCount) - 1
			}
			m.updateFocus()
			return m, nil
		}

		// Handle tab - always navigates to next field
		if key.Matches(msg, configMenuKeys.Tab) {
			m.focusIndex++
			if m.focusIndex >= int(FieldCount) {
				m.focusIndex = 0
			}
			m.updateFocus()
			return m, nil
		}

		// Handle down arrow - always navigates (single-line text inputs don't need down arrow)
		if key.Matches(msg, configMenuKeys.Down) {
			m.focusIndex++
			if m.focusIndex >= int(FieldCount) {
				m.focusIndex = 0
			}
			m.updateFocus()
			return m, nil
		}

		// Handle toggle for checkboxes and buttons
		if key.Matches(msg, configMenuKeys.Toggle) {
			if m.isToggleField() {
				m.toggleCurrentField()
				return m, nil
			} else if ConfigField(m.focusIndex) == FieldChangeInterface {
				return m, func() tea.Msg {
					return ChangeInterfaceMsg{}
				}
			} else if ConfigField(m.focusIndex) == FieldSave {
				return m.saveConfig()
			} else if ConfigField(m.focusIndex) == FieldCancel {
				return m, func() tea.Msg {
					return ConfigCancelledMsg{}
				}
			}
		}

		// Pass remaining keys to text inputs only
		if m.isTextInputField() {
			var cmd tea.Cmd
			switch ConfigField(m.focusIndex) {
			case FieldSystemName:
				m.systemNameInput, cmd = m.systemNameInput.Update(msg)
			case FieldSystemDescription:
				m.systemDescInput, cmd = m.systemDescInput.Update(msg)
			case FieldInterval:
				m.intervalInput, cmd = m.intervalInput.Update(msg)
			case FieldTTL:
				m.ttlInput, cmd = m.ttlInput.Update(msg)
			}
			return m, cmd
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// isTextInputField returns true if current field is a text input
func (m ConfigMenuModel) isTextInputField() bool {
	switch ConfigField(m.focusIndex) {
	case FieldSystemName, FieldSystemDescription, FieldInterval, FieldTTL:
		return true
	}
	return false
}

// isToggleField returns true if current field is a toggle
func (m ConfigMenuModel) isToggleField() bool {
	switch ConfigField(m.focusIndex) {
	case FieldCDPListen, FieldLLDPListen, FieldCDPBroadcast, FieldLLDPBroadcast,
		FieldBroadcastOnStartup, FieldCapRouter, FieldCapBridge, FieldCapStation:
		return true
	}
	return false
}

// toggleCurrentField toggles the current field's value
func (m *ConfigMenuModel) toggleCurrentField() {
	switch ConfigField(m.focusIndex) {
	case FieldCDPListen:
		m.cdpListen = !m.cdpListen
	case FieldLLDPListen:
		m.lldpListen = !m.lldpListen
	case FieldCDPBroadcast:
		m.cdpBroadcast = !m.cdpBroadcast
	case FieldLLDPBroadcast:
		m.lldpBroadcast = !m.lldpBroadcast
	case FieldBroadcastOnStartup:
		m.broadcastOnStartup = !m.broadcastOnStartup
	case FieldCapRouter:
		m.capRouter = !m.capRouter
	case FieldCapBridge:
		m.capBridge = !m.capBridge
	case FieldCapStation:
		m.capStation = !m.capStation
	}
}

// updateFocus updates which input is focused
func (m *ConfigMenuModel) updateFocus() {
	m.systemNameInput.Blur()
	m.systemDescInput.Blur()
	m.intervalInput.Blur()
	m.ttlInput.Blur()

	switch ConfigField(m.focusIndex) {
	case FieldSystemName:
		m.systemNameInput.Focus()
	case FieldSystemDescription:
		m.systemDescInput.Focus()
	case FieldInterval:
		m.intervalInput.Focus()
	case FieldTTL:
		m.ttlInput.Focus()
	}
}

// saveConfig saves the configuration and returns a message
func (m ConfigMenuModel) saveConfig() (tea.Model, tea.Cmd) {
	// Build capabilities list
	var caps []string
	if m.capRouter {
		caps = append(caps, "router")
	}
	if m.capBridge {
		caps = append(caps, "bridge")
	}
	if m.capStation {
		caps = append(caps, "station")
	}
	if len(caps) == 0 {
		caps = []string{"station"}
	}

	// Parse numeric values
	interval, err := strconv.Atoi(m.intervalInput.Value())
	if err != nil || interval <= 0 {
		interval = 5
	}

	ttl, err := strconv.Atoi(m.ttlInput.Value())
	if err != nil || ttl <= 0 {
		ttl = 20
	}

	// Update config
	m.config.SystemName = m.systemNameInput.Value()
	m.config.SystemDescription = m.systemDescInput.Value()
	m.config.CDPListen = m.cdpListen
	m.config.LLDPListen = m.lldpListen
	m.config.CDPBroadcast = m.cdpBroadcast
	m.config.LLDPBroadcast = m.lldpBroadcast
	m.config.BroadcastOnStartup = m.broadcastOnStartup
	m.config.AdvertiseInterval = interval
	m.config.TTL = ttl
	m.config.Capabilities = caps

	// Check if listen settings changed (need new log file)
	listenChanged := m.cdpListen != m.originalCDPListen || m.lldpListen != m.originalLLDPListen

	// Save to file
	_ = config.Save(*m.config)

	return m, func() tea.Msg {
		return ConfigSavedMsg{Config: m.config, ListenSettingsChanged: listenChanged}
	}
}

// View renders the config menu
func (m ConfigMenuModel) View() string {
	header := m.renderHeader()
	content := m.renderContent()
	footer := m.renderFooter()

	// Count lines used
	headerLines := strings.Count(header, "\n") + 1
	contentLines := strings.Count(content, "\n") + 1
	footerLines := 1

	usedLines := headerLines + contentLines + footerLines
	padding := m.height - usedLines
	if padding < 0 {
		padding = 0
	}

	var b strings.Builder
	b.WriteString(header)
	b.WriteString("\n")
	b.WriteString(content)
	b.WriteString("\n")
	b.WriteString(strings.Repeat("\n", padding))
	b.WriteString(footer)

	return b.String()
}

// renderHeader renders the header bar
func (m ConfigMenuModel) renderHeader() string {
	return RenderHeader(HeaderLeft(), HeaderTitle("Configuration"), m.width)
}

// renderContent renders the form content
func (m ConfigMenuModel) renderContent() string {
	theme := DefaultTheme
	var b strings.Builder

	// Styles
	sectionStyle := lipgloss.NewStyle().
		Foreground(theme.Base0D).
		Bold(true)
	labelStyle := lipgloss.NewStyle().
		Foreground(theme.Base05)
	focusedStyle := lipgloss.NewStyle().
		Foreground(theme.Base0B).
		Bold(true)
	dimStyle := lipgloss.NewStyle().
		Foreground(theme.Base03)
	cursorStyle := lipgloss.NewStyle().
		Foreground(theme.Base0C).
		Bold(true)

	// Checkbox helper
	checkbox := func(checked bool, focused bool) string {
		style := labelStyle
		if focused {
			style = focusedStyle
		}
		if checked {
			return style.Render("[x]")
		}
		return style.Render("[ ]")
	}

	// Cursor helper
	cursor := func(focused bool) string {
		if focused {
			return cursorStyle.Render(">") + " "
		}
		return "  "
	}

	b.WriteString("\n")

	// ═══════════════════════════════════════════════════════════
	// Interface Selection
	// ═══════════════════════════════════════════════════════════
	focused := ConfigField(m.focusIndex) == FieldChangeInterface
	b.WriteString("  ")
	b.WriteString(cursor(focused))
	if focused {
		b.WriteString(focusedStyle.Render("[Change Interface]"))
	} else {
		b.WriteString(labelStyle.Render("[Change Interface]"))
	}
	b.WriteString("\n\n")

	// ═══════════════════════════════════════════════════════════
	// System Identity
	// ═══════════════════════════════════════════════════════════
	b.WriteString("  ")
	b.WriteString(sectionStyle.Render("System Identity"))
	b.WriteString("\n\n")

	// System Name
	focused = ConfigField(m.focusIndex) == FieldSystemName
	b.WriteString("  ")
	b.WriteString(cursor(focused))
	if focused {
		b.WriteString(focusedStyle.Render("System Name"))
	} else {
		b.WriteString(labelStyle.Render("System Name"))
	}
	b.WriteString("       ")
	b.WriteString(m.systemNameInput.View())
	b.WriteString("\n")

	// Description
	focused = ConfigField(m.focusIndex) == FieldSystemDescription
	b.WriteString("  ")
	b.WriteString(cursor(focused))
	if focused {
		b.WriteString(focusedStyle.Render("Description"))
	} else {
		b.WriteString(labelStyle.Render("Description"))
	}
	b.WriteString("       ")
	b.WriteString(m.systemDescInput.View())
	b.WriteString("\n\n")

	// ═══════════════════════════════════════════════════════════
	// Listening
	// ═══════════════════════════════════════════════════════════
	b.WriteString("  ")
	b.WriteString(sectionStyle.Render("Listening"))
	b.WriteString("\n\n")

	// CDP Listen
	focused = ConfigField(m.focusIndex) == FieldCDPListen
	b.WriteString("  ")
	b.WriteString(cursor(focused))
	b.WriteString(checkbox(m.cdpListen, focused))
	b.WriteString(" ")
	if focused {
		b.WriteString(focusedStyle.Render("CDP"))
	} else {
		b.WriteString(labelStyle.Render("CDP"))
	}

	b.WriteString("        ")

	// LLDP Listen
	focused = ConfigField(m.focusIndex) == FieldLLDPListen
	b.WriteString(checkbox(m.lldpListen, focused))
	b.WriteString(" ")
	if focused {
		b.WriteString(focusedStyle.Render("LLDP"))
	} else {
		b.WriteString(labelStyle.Render("LLDP"))
	}
	b.WriteString("\n\n")

	// ═══════════════════════════════════════════════════════════
	// Broadcasting
	// ═══════════════════════════════════════════════════════════
	b.WriteString("  ")
	b.WriteString(sectionStyle.Render("Broadcasting"))
	b.WriteString("\n\n")

	// CDP Broadcast
	focused = ConfigField(m.focusIndex) == FieldCDPBroadcast
	b.WriteString("  ")
	b.WriteString(cursor(focused))
	b.WriteString(checkbox(m.cdpBroadcast, focused))
	b.WriteString(" ")
	if focused {
		b.WriteString(focusedStyle.Render("CDP"))
	} else {
		b.WriteString(labelStyle.Render("CDP"))
	}

	b.WriteString("        ")

	// LLDP Broadcast
	focused = ConfigField(m.focusIndex) == FieldLLDPBroadcast
	b.WriteString(checkbox(m.lldpBroadcast, focused))
	b.WriteString(" ")
	if focused {
		b.WriteString(focusedStyle.Render("LLDP"))
	} else {
		b.WriteString(labelStyle.Render("LLDP"))
	}
	b.WriteString("\n")

	// Broadcast on Startup
	focused = ConfigField(m.focusIndex) == FieldBroadcastOnStartup
	b.WriteString("  ")
	b.WriteString(cursor(focused))
	b.WriteString(checkbox(m.broadcastOnStartup, focused))
	b.WriteString(" ")
	if focused {
		b.WriteString(focusedStyle.Render("Start on launch"))
	} else {
		b.WriteString(labelStyle.Render("Start on launch"))
	}
	b.WriteString("\n\n")

	// Interval
	focused = ConfigField(m.focusIndex) == FieldInterval
	b.WriteString("  ")
	b.WriteString(cursor(focused))
	if focused {
		b.WriteString(focusedStyle.Render("Interval"))
	} else {
		b.WriteString(labelStyle.Render("Interval"))
	}
	b.WriteString("          ")
	b.WriteString(m.intervalInput.View())
	b.WriteString(dimStyle.Render(" seconds"))
	b.WriteString("\n")

	// TTL
	focused = ConfigField(m.focusIndex) == FieldTTL
	b.WriteString("  ")
	b.WriteString(cursor(focused))
	if focused {
		b.WriteString(focusedStyle.Render("TTL (Hold Time)"))
	} else {
		b.WriteString(labelStyle.Render("TTL (Hold Time)"))
	}
	b.WriteString("   ")
	b.WriteString(m.ttlInput.View())
	b.WriteString(dimStyle.Render(" seconds"))
	b.WriteString("\n\n")

	// ═══════════════════════════════════════════════════════════
	// Capabilities
	// ═══════════════════════════════════════════════════════════
	b.WriteString("  ")
	b.WriteString(sectionStyle.Render("Capabilities (advertised)"))
	b.WriteString("\n\n")

	// Router
	focused = ConfigField(m.focusIndex) == FieldCapRouter
	b.WriteString("  ")
	b.WriteString(cursor(focused))
	b.WriteString(checkbox(m.capRouter, focused))
	b.WriteString(" ")
	if focused {
		b.WriteString(focusedStyle.Render("Router"))
	} else {
		b.WriteString(labelStyle.Render("Router"))
	}

	b.WriteString("     ")

	// Bridge
	focused = ConfigField(m.focusIndex) == FieldCapBridge
	b.WriteString(checkbox(m.capBridge, focused))
	b.WriteString(" ")
	if focused {
		b.WriteString(focusedStyle.Render("Bridge"))
	} else {
		b.WriteString(labelStyle.Render("Bridge"))
	}

	b.WriteString("     ")

	// Station
	focused = ConfigField(m.focusIndex) == FieldCapStation
	b.WriteString(checkbox(m.capStation, focused))
	b.WriteString(" ")
	if focused {
		b.WriteString(focusedStyle.Render("Station"))
	} else {
		b.WriteString(labelStyle.Render("Station"))
	}
	b.WriteString("\n\n")

	// ═══════════════════════════════════════════════════════════
	// Buttons
	// ═══════════════════════════════════════════════════════════
	// Save
	focused = ConfigField(m.focusIndex) == FieldSave
	b.WriteString("  ")
	b.WriteString(cursor(focused))
	if focused {
		b.WriteString(focusedStyle.Render("[Save]"))
	} else {
		b.WriteString(labelStyle.Render("[Save]"))
	}

	b.WriteString("     ")

	// Cancel
	focused = ConfigField(m.focusIndex) == FieldCancel
	if focused {
		b.WriteString(focusedStyle.Render("[Cancel]"))
	} else {
		b.WriteString(labelStyle.Render("[Cancel]"))
	}

	return b.String()
}

// renderFooter renders the footer bar
func (m ConfigMenuModel) renderFooter() string {
	theme := DefaultTheme
	bg := theme.Base01

	// All text needs background color
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

	content := keyStyle.Render("↑/↓") + textStyle.Render(" navigate") + sep +
		keyStyle.Render("space") + textStyle.Render(" toggle") + sep +
		keyStyle.Render("tab") + textStyle.Render(" next") + sep +
		keyStyle.Render("ctrl+s") + textStyle.Render(" save") + sep +
		keyStyle.Render("esc") + textStyle.Render(" cancel")

	return RenderFooter(content, m.width)
}

// GetConfig returns the current config
func (m *ConfigMenuModel) GetConfig() *config.Config {
	return m.config
}
