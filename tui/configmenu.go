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

// ConfigSubState represents the current sub-menu state
type ConfigSubState int

const (
	SubStateMain ConfigSubState = iota
	SubStateListening
	SubStateBroadcast
	SubStateLogging
	SubStateTheme
	SubStateAbout
)

// ConfigMenuItem represents an item in the config menu
type ConfigMenuItem int

const (
	ConfigMenuChangeInterface ConfigMenuItem = iota
	ConfigMenuListening
	ConfigMenuBroadcast
	ConfigMenuLogging
	ConfigMenuTheme
	ConfigMenuAbout
	ConfigMenuSaveExit
	ConfigMenuCancel
	ConfigMenuItemCount
)

// Main menu item labels
var mainMenuLabels = []string{
	"Change Interface",
	"Listening Options",
	"Broadcast Options",
	"Logging Options",
	"Change Theme",
	"About",
	"Save & Exit",
	"Cancel",
}

// ConfigMenuModel is the model for the configuration menu
type ConfigMenuModel struct {
	subState   ConfigSubState
	mainCursor int // Cursor position in main menu
	subCursor  int // Cursor position in current sub-menu

	config *config.Config

	// Theme preview
	previousTheme     Theme
	themeIndex        int  // Current theme index being previewed
	themePreviewDirty bool // True if theme has been changed

	// Text inputs for Broadcast Options
	systemNameInput textinput.Model
	systemDescInput textinput.Model
	intervalInput   textinput.Model
	ttlInput        textinput.Model

	// Text inputs for Listening Options
	stalenessInput   textinput.Model
	staleRemovalInput textinput.Model

	// Text inputs for Logging Options
	logDirInput textinput.Model

	// Listening Options state
	cdpListen     bool
	lldpListen    bool
	filterRouter  bool
	filterBridge  bool
	filterStation bool
	stalenessTimeout int
	staleRemovalTime int

	// Broadcast Options state
	cdpBroadcast       bool
	lldpBroadcast      bool
	broadcastOnStartup bool
	capRouter          bool
	capBridge          bool
	capStation         bool

	// Logging Options state
	loggingEnabled bool
	logDirectory   string

	// Track original settings for change detection
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

	// Create text inputs for Broadcast Options
	systemNameInput := textinput.New()
	systemNameInput.Placeholder = resolvedHostname
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

	// Create text inputs for Listening Options
	stalenessInput := textinput.New()
	stalenessInput.Placeholder = "180"
	stalenessInput.CharLimit = 5
	stalenessInput.Width = 6
	stalenessInput.SetValue(strconv.Itoa(cfg.StalenessTimeout))

	staleRemovalInput := textinput.New()
	staleRemovalInput.Placeholder = "0"
	staleRemovalInput.CharLimit = 5
	staleRemovalInput.Width = 6
	staleRemovalInput.SetValue(strconv.Itoa(cfg.StaleRemovalTime))

	// Create text inputs for Logging Options
	logDirInput := textinput.New()
	logDirInput.Placeholder = "(default location)"
	logDirInput.CharLimit = 256
	logDirInput.Width = 40
	logDirInput.SetValue(cfg.LogDirectory)

	// Parse broadcast capabilities
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

	// Parse filter capabilities
	filterRouter := false
	filterBridge := false
	filterStation := false
	for _, cap := range cfg.FilterCapabilities {
		switch strings.ToLower(cap) {
		case "router":
			filterRouter = true
		case "bridge":
			filterBridge = true
		case "station":
			filterStation = true
		}
	}

	// Get current theme index
	themeIndex := GetThemeIndex(cfg.Theme)
	if themeIndex < 0 {
		themeIndex = 0
	}

	return ConfigMenuModel{
		subState:           SubStateMain,
		mainCursor:         0,
		subCursor:          0,
		config:             cfg,
		previousTheme:      DefaultTheme,
		themeIndex:         themeIndex,
		systemNameInput:    systemNameInput,
		systemDescInput:    systemDescInput,
		intervalInput:      intervalInput,
		ttlInput:           ttlInput,
		stalenessInput:     stalenessInput,
		staleRemovalInput:  staleRemovalInput,
		logDirInput:        logDirInput,
		cdpListen:          cfg.CDPListen,
		lldpListen:         cfg.LLDPListen,
		filterRouter:       filterRouter,
		filterBridge:       filterBridge,
		filterStation:      filterStation,
		stalenessTimeout:   cfg.StalenessTimeout,
		staleRemovalTime:   cfg.StaleRemovalTime,
		cdpBroadcast:       cfg.CDPBroadcast,
		lldpBroadcast:      cfg.LLDPBroadcast,
		broadcastOnStartup: cfg.BroadcastOnStartup,
		capRouter:          capRouter,
		capBridge:          capBridge,
		capStation:         capStation,
		loggingEnabled:     cfg.LoggingEnabled,
		logDirectory:       cfg.LogDirectory,
		originalCDPListen:  cfg.CDPListen,
		originalLLDPListen: cfg.LLDPListen,
		resolvedHostname:   resolvedHostname,
		styles:             DefaultStyles,
	}
}

// Init initializes the config menu
func (m ConfigMenuModel) Init() tea.Cmd {
	return textinput.Blink
}

// Key bindings
type configMenuKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Left   key.Binding
	Right  key.Binding
	Select key.Binding
	Back   key.Binding
	Save   key.Binding
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
	Left: key.NewBinding(
		key.WithKeys("left"),
		key.WithHelp("←", "left"),
	),
	Right: key.NewBinding(
		key.WithKeys("right"),
		key.WithHelp("→", "right"),
	),
	Select: key.NewBinding(
		key.WithKeys(" ", "enter"),
		key.WithHelp("enter", "select"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Save: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "save"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next"),
	),
}

// Messages
type ConfigSavedMsg struct {
	Config                *config.Config
	ListenSettingsChanged bool
}

type ConfigCancelledMsg struct{}

type ChangeInterfaceMsg struct{}

// Update handles messages for the config menu
func (m ConfigMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Ctrl+S saves from anywhere
		if key.Matches(msg, configMenuKeys.Save) {
			return m.saveConfig()
		}

		// Route to appropriate handler based on sub-state
		switch m.subState {
		case SubStateMain:
			return m.updateMain(msg)
		case SubStateListening:
			return m.updateListening(msg)
		case SubStateBroadcast:
			return m.updateBroadcast(msg)
		case SubStateLogging:
			return m.updateLogging(msg)
		case SubStateTheme:
			return m.updateTheme(msg)
		case SubStateAbout:
			return m.updateAbout(msg)
		}
	}

	return m, nil
}

// saveConfig saves the configuration and returns a message
func (m ConfigMenuModel) saveConfig() (tea.Model, tea.Cmd) {
	// Parse staleness values
	staleness, err := strconv.Atoi(m.stalenessInput.Value())
	if err != nil || staleness < 0 {
		staleness = 180
	}
	staleRemoval, err := strconv.Atoi(m.staleRemovalInput.Value())
	if err != nil || staleRemoval < 0 {
		staleRemoval = 0
	}

	// Parse broadcast values
	interval, err := strconv.Atoi(m.intervalInput.Value())
	if err != nil || interval <= 0 {
		interval = 5
	}
	ttl, err := strconv.Atoi(m.ttlInput.Value())
	if err != nil || ttl <= 0 {
		ttl = 20
	}

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

	// Build filter capabilities list
	var filterCaps []string
	if m.filterRouter {
		filterCaps = append(filterCaps, "router")
	}
	if m.filterBridge {
		filterCaps = append(filterCaps, "bridge")
	}
	if m.filterStation {
		filterCaps = append(filterCaps, "station")
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
	m.config.FilterCapabilities = filterCaps
	m.config.StalenessTimeout = staleness
	m.config.StaleRemovalTime = staleRemoval
	m.config.LoggingEnabled = m.loggingEnabled
	m.config.LogDirectory = m.logDirInput.Value()

	// Update theme from the selected index
	themeSlug, _, _ := GetThemeByIndex(m.themeIndex)
	if themeSlug != "" {
		m.config.Theme = themeSlug
	}

	// Check if listen settings changed
	listenChanged := m.cdpListen != m.originalCDPListen || m.lldpListen != m.originalLLDPListen

	// Save to file
	_ = config.Save(*m.config)

	return m, func() tea.Msg {
		return ConfigSavedMsg{Config: m.config, ListenSettingsChanged: listenChanged}
	}
}

// View renders the config menu
func (m ConfigMenuModel) View() string {
	var content string

	switch m.subState {
	case SubStateMain:
		content = m.renderMainMenu()
	case SubStateListening:
		content = m.renderListening()
	case SubStateBroadcast:
		content = m.renderBroadcast()
	case SubStateLogging:
		content = m.renderLogging()
	case SubStateTheme:
		content = m.renderTheme()
	case SubStateAbout:
		content = m.renderAbout()
	}

	header := m.renderHeader()
	footer := m.renderFooter()

	// Calculate padding
	headerLines := strings.Count(header, "\n") + 1
	contentLines := strings.Count(content, "\n")
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
	b.WriteString(strings.Repeat("\n", padding))
	b.WriteString(footer)

	return b.String()
}

// renderHeader renders the header bar
func (m ConfigMenuModel) renderHeader() string {
	var title string
	switch m.subState {
	case SubStateMain:
		title = "Configuration"
	case SubStateListening:
		title = "Listening Options"
	case SubStateBroadcast:
		title = "Broadcast Options"
	case SubStateLogging:
		title = "Logging Options"
	case SubStateTheme:
		title = "Change Theme"
	case SubStateAbout:
		title = "About"
	}
	return RenderHeader(HeaderLeft(), HeaderTitle(title), m.width)
}

// renderFooter renders the footer bar
func (m ConfigMenuModel) renderFooter() string {
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

	var content string
	switch m.subState {
	case SubStateMain:
		content = keyStyle.Render("↑/↓") + textStyle.Render(" navigate") + sep +
			keyStyle.Render("enter") + textStyle.Render(" select") + sep +
			keyStyle.Render("ctrl+s") + textStyle.Render(" save")
	case SubStateTheme:
		content = keyStyle.Render("↑/↓") + textStyle.Render(" preview") + sep +
			keyStyle.Render("enter") + textStyle.Render(" select") + sep +
			keyStyle.Render("esc") + textStyle.Render(" cancel")
	case SubStateAbout:
		content = keyStyle.Render("esc") + textStyle.Render(" back") + sep +
			keyStyle.Render("enter") + textStyle.Render(" back")
	case SubStateListening, SubStateBroadcast:
		content = keyStyle.Render("↑/↓/←/→") + textStyle.Render(" navigate") + sep +
			keyStyle.Render("space") + textStyle.Render(" toggle") + sep +
			keyStyle.Render("esc") + textStyle.Render(" back") + sep +
			keyStyle.Render("ctrl+s") + textStyle.Render(" save")
	default:
		content = keyStyle.Render("↑/↓") + textStyle.Render(" navigate") + sep +
			keyStyle.Render("space") + textStyle.Render(" toggle") + sep +
			keyStyle.Render("esc") + textStyle.Render(" back") + sep +
			keyStyle.Render("ctrl+s") + textStyle.Render(" save")
	}

	return RenderFooter(content, m.width)
}

// GetConfig returns the current config
func (m *ConfigMenuModel) GetConfig() *config.Config {
	return m.config
}
