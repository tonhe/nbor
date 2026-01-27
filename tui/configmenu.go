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
	"nbor/version"
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

// updateListening handles key events for the Listening Options sub-menu
func (m ConfigMenuModel) updateListening(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Listening sub-menu fields organized by row:
	// Row 0: CDP Listen (0), LLDP Listen (1)
	// Row 1: Filter Router (2), Filter Bridge (3), Filter Station (4)
	// Row 2: Staleness Timeout (5)
	// Row 3: Stale Removal (6)
	// Row 4: Back button (7)

	// Define row groupings for left/right navigation
	listeningRows := [][]int{
		{0, 1},       // CDP, LLDP
		{2, 3, 4},    // Router, Bridge, Station
		{5},          // Staleness
		{6},          // Stale Removal
		{7},          // Back
	}

	switch {
	case key.Matches(msg, configMenuKeys.Back):
		m.subState = SubStateMain
		m.stalenessInput.Blur()
		m.staleRemovalInput.Blur()

	case key.Matches(msg, configMenuKeys.Left):
		// Move left within the current row
		row, col := m.findListeningPosition(listeningRows)
		if col > 0 {
			m.stalenessInput.Blur()
			m.staleRemovalInput.Blur()
			m.subCursor = listeningRows[row][col-1]
			m.focusListeningInput()
		}

	case key.Matches(msg, configMenuKeys.Right):
		// Move right within the current row
		row, col := m.findListeningPosition(listeningRows)
		if col < len(listeningRows[row])-1 {
			m.stalenessInput.Blur()
			m.staleRemovalInput.Blur()
			m.subCursor = listeningRows[row][col+1]
			m.focusListeningInput()
		}

	case key.Matches(msg, configMenuKeys.Up):
		m.stalenessInput.Blur()
		m.staleRemovalInput.Blur()
		row, col := m.findListeningPosition(listeningRows)
		row--
		if row < 0 {
			row = len(listeningRows) - 1
		}
		// Keep column position if possible, otherwise go to last item in row
		if col >= len(listeningRows[row]) {
			col = len(listeningRows[row]) - 1
		}
		m.subCursor = listeningRows[row][col]
		m.focusListeningInput()

	case key.Matches(msg, configMenuKeys.Down), key.Matches(msg, configMenuKeys.Tab):
		m.stalenessInput.Blur()
		m.staleRemovalInput.Blur()
		row, col := m.findListeningPosition(listeningRows)
		row++
		if row >= len(listeningRows) {
			row = 0
		}
		// Keep column position if possible, otherwise go to last item in row
		if col >= len(listeningRows[row]) {
			col = len(listeningRows[row]) - 1
		}
		m.subCursor = listeningRows[row][col]
		m.focusListeningInput()

	case key.Matches(msg, configMenuKeys.Select):
		switch m.subCursor {
		case 0:
			m.cdpListen = !m.cdpListen
		case 1:
			m.lldpListen = !m.lldpListen
		case 2:
			m.filterRouter = !m.filterRouter
		case 3:
			m.filterBridge = !m.filterBridge
		case 4:
			m.filterStation = !m.filterStation
		case 7: // Back
			m.subState = SubStateMain
			m.stalenessInput.Blur()
			m.staleRemovalInput.Blur()
		}

	default:
		// Pass to text input if focused
		if m.subCursor == 5 {
			var cmd tea.Cmd
			m.stalenessInput, cmd = m.stalenessInput.Update(msg)
			return m, cmd
		} else if m.subCursor == 6 {
			var cmd tea.Cmd
			m.staleRemovalInput, cmd = m.staleRemovalInput.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

// findListeningPosition returns the row and column position for the current cursor
func (m *ConfigMenuModel) findListeningPosition(rows [][]int) (row, col int) {
	for r, rowFields := range rows {
		for c, field := range rowFields {
			if field == m.subCursor {
				return r, c
			}
		}
	}
	return 0, 0
}

func (m *ConfigMenuModel) focusListeningInput() {
	m.stalenessInput.Blur()
	m.staleRemovalInput.Blur()
	if m.subCursor == 5 {
		m.stalenessInput.Focus()
	} else if m.subCursor == 6 {
		m.staleRemovalInput.Focus()
	}
}

// updateBroadcast handles key events for the Broadcast Options sub-menu
func (m ConfigMenuModel) updateBroadcast(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Broadcast sub-menu fields organized by row:
	// Row 0: System Name (0)
	// Row 1: Description (1)
	// Row 2: CDP Broadcast (2), LLDP Broadcast (3)
	// Row 3: Start on Launch (4)
	// Row 4: Interval (5)
	// Row 5: TTL (6)
	// Row 6: Cap Router (7), Cap Bridge (8), Cap Station (9)
	// Row 7: Back button (10)

	// Define row groupings for left/right navigation
	broadcastRows := [][]int{
		{0},          // System Name
		{1},          // Description
		{2, 3},       // CDP, LLDP
		{4},          // Start on Launch
		{5},          // Interval
		{6},          // TTL
		{7, 8, 9},    // Router, Bridge, Station
		{10},         // Back
	}

	switch {
	case key.Matches(msg, configMenuKeys.Back):
		m.subState = SubStateMain
		m.blurAllBroadcastInputs()

	case key.Matches(msg, configMenuKeys.Left):
		// Move left within the current row
		row, col := m.findBroadcastPosition(broadcastRows)
		if col > 0 {
			m.blurAllBroadcastInputs()
			m.subCursor = broadcastRows[row][col-1]
			m.focusBroadcastInput()
		}

	case key.Matches(msg, configMenuKeys.Right):
		// Move right within the current row
		row, col := m.findBroadcastPosition(broadcastRows)
		if col < len(broadcastRows[row])-1 {
			m.blurAllBroadcastInputs()
			m.subCursor = broadcastRows[row][col+1]
			m.focusBroadcastInput()
		}

	case key.Matches(msg, configMenuKeys.Up):
		m.blurAllBroadcastInputs()
		row, col := m.findBroadcastPosition(broadcastRows)
		row--
		if row < 0 {
			row = len(broadcastRows) - 1
		}
		// Keep column position if possible, otherwise go to last item in row
		if col >= len(broadcastRows[row]) {
			col = len(broadcastRows[row]) - 1
		}
		m.subCursor = broadcastRows[row][col]
		m.focusBroadcastInput()

	case key.Matches(msg, configMenuKeys.Down), key.Matches(msg, configMenuKeys.Tab):
		m.blurAllBroadcastInputs()
		row, col := m.findBroadcastPosition(broadcastRows)
		row++
		if row >= len(broadcastRows) {
			row = 0
		}
		// Keep column position if possible, otherwise go to last item in row
		if col >= len(broadcastRows[row]) {
			col = len(broadcastRows[row]) - 1
		}
		m.subCursor = broadcastRows[row][col]
		m.focusBroadcastInput()

	case key.Matches(msg, configMenuKeys.Select):
		switch m.subCursor {
		case 2:
			m.cdpBroadcast = !m.cdpBroadcast
		case 3:
			m.lldpBroadcast = !m.lldpBroadcast
		case 4:
			m.broadcastOnStartup = !m.broadcastOnStartup
		case 7:
			m.capRouter = !m.capRouter
		case 8:
			m.capBridge = !m.capBridge
		case 9:
			m.capStation = !m.capStation
		case 10: // Back
			m.subState = SubStateMain
			m.blurAllBroadcastInputs()
		}

	default:
		// Pass to text input if focused
		var cmd tea.Cmd
		switch m.subCursor {
		case 0:
			m.systemNameInput, cmd = m.systemNameInput.Update(msg)
			return m, cmd
		case 1:
			m.systemDescInput, cmd = m.systemDescInput.Update(msg)
			return m, cmd
		case 5:
			m.intervalInput, cmd = m.intervalInput.Update(msg)
			return m, cmd
		case 6:
			m.ttlInput, cmd = m.ttlInput.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

// findBroadcastPosition returns the row and column position for the current cursor
func (m *ConfigMenuModel) findBroadcastPosition(rows [][]int) (row, col int) {
	for r, rowFields := range rows {
		for c, field := range rowFields {
			if field == m.subCursor {
				return r, c
			}
		}
	}
	return 0, 0
}

func (m *ConfigMenuModel) blurAllBroadcastInputs() {
	m.systemNameInput.Blur()
	m.systemDescInput.Blur()
	m.intervalInput.Blur()
	m.ttlInput.Blur()
}

func (m *ConfigMenuModel) focusBroadcastInput() {
	switch m.subCursor {
	case 0:
		m.systemNameInput.Focus()
	case 1:
		m.systemDescInput.Focus()
	case 5:
		m.intervalInput.Focus()
	case 6:
		m.ttlInput.Focus()
	}
}

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
		// Confirm theme selection
		m.themeIndex = m.subCursor
		slug, _, _ := GetThemeByIndex(m.themeIndex)
		m.config.Theme = slug
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

// updateAbout handles key events for the About screen
func (m ConfigMenuModel) updateAbout(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Any key returns to main menu
	if key.Matches(msg, configMenuKeys.Back) || key.Matches(msg, configMenuKeys.Select) {
		m.subState = SubStateMain
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

// renderListening renders the Listening Options sub-menu
func (m ConfigMenuModel) renderListening() string {
	theme := DefaultTheme
	var b strings.Builder

	sectionStyle := lipgloss.NewStyle().Foreground(theme.Base0D).Bold(true)
	labelStyle := lipgloss.NewStyle().Foreground(theme.Base05)
	focusedStyle := lipgloss.NewStyle().Foreground(theme.Base0B).Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(theme.Base03)
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

	// Protocol Listening
	b.WriteString("  ")
	b.WriteString(sectionStyle.Render("Protocol Listening"))
	b.WriteString("\n\n")

	// CDP Listen
	b.WriteString("  ")
	b.WriteString(cursor(m.subCursor == 0))
	b.WriteString(checkbox(m.cdpListen, m.subCursor == 0))
	b.WriteString(" ")
	if m.subCursor == 0 {
		b.WriteString(focusedStyle.Render("CDP"))
	} else {
		b.WriteString(labelStyle.Render("CDP"))
	}
	b.WriteString("        ")

	// LLDP Listen
	b.WriteString(checkbox(m.lldpListen, m.subCursor == 1))
	b.WriteString(" ")
	if m.subCursor == 1 {
		b.WriteString(focusedStyle.Render("LLDP"))
	} else {
		b.WriteString(labelStyle.Render("LLDP"))
	}
	b.WriteString("\n\n")

	// Filter by Capabilities
	b.WriteString("  ")
	b.WriteString(sectionStyle.Render("Filter by Capabilities"))
	b.WriteString(" ")
	b.WriteString(dimStyle.Render("(empty = show all)"))
	b.WriteString("\n\n")

	// Filter Router
	b.WriteString("  ")
	b.WriteString(cursor(m.subCursor == 2))
	b.WriteString(checkbox(m.filterRouter, m.subCursor == 2))
	b.WriteString(" ")
	if m.subCursor == 2 {
		b.WriteString(focusedStyle.Render("Router"))
	} else {
		b.WriteString(labelStyle.Render("Router"))
	}
	b.WriteString("     ")

	// Filter Bridge
	b.WriteString(checkbox(m.filterBridge, m.subCursor == 3))
	b.WriteString(" ")
	if m.subCursor == 3 {
		b.WriteString(focusedStyle.Render("Bridge"))
	} else {
		b.WriteString(labelStyle.Render("Bridge"))
	}
	b.WriteString("     ")

	// Filter Station
	b.WriteString(checkbox(m.filterStation, m.subCursor == 4))
	b.WriteString(" ")
	if m.subCursor == 4 {
		b.WriteString(focusedStyle.Render("Station"))
	} else {
		b.WriteString(labelStyle.Render("Station"))
	}
	b.WriteString("\n\n")

	// Display Settings
	b.WriteString("  ")
	b.WriteString(sectionStyle.Render("Display Settings"))
	b.WriteString("\n\n")

	// Staleness Timeout
	b.WriteString("  ")
	b.WriteString(cursor(m.subCursor == 5))
	if m.subCursor == 5 {
		b.WriteString(focusedStyle.Render("Staleness Timeout"))
	} else {
		b.WriteString(labelStyle.Render("Staleness Timeout"))
	}
	b.WriteString("  ")
	b.WriteString(m.stalenessInput.View())
	b.WriteString(dimStyle.Render(" seconds (gray out)"))
	b.WriteString("\n")

	// Stale Removal
	b.WriteString("  ")
	b.WriteString(cursor(m.subCursor == 6))
	if m.subCursor == 6 {
		b.WriteString(focusedStyle.Render("Stale Removal"))
	} else {
		b.WriteString(labelStyle.Render("Stale Removal"))
	}
	b.WriteString("      ")
	b.WriteString(m.staleRemovalInput.View())
	b.WriteString(dimStyle.Render(" seconds (0 = never)"))
	b.WriteString("\n\n")

	// Back button
	b.WriteString("  ")
	b.WriteString(cursor(m.subCursor == 7))
	if m.subCursor == 7 {
		b.WriteString(focusedStyle.Render("[Back]"))
	} else {
		b.WriteString(labelStyle.Render("[Back]"))
	}
	b.WriteString("\n")

	return b.String()
}

// renderBroadcast renders the Broadcast Options sub-menu
func (m ConfigMenuModel) renderBroadcast() string {
	theme := DefaultTheme
	var b strings.Builder

	sectionStyle := lipgloss.NewStyle().Foreground(theme.Base0D).Bold(true)
	labelStyle := lipgloss.NewStyle().Foreground(theme.Base05)
	focusedStyle := lipgloss.NewStyle().Foreground(theme.Base0B).Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(theme.Base03)
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

	// System Identity
	b.WriteString("  ")
	b.WriteString(sectionStyle.Render("System Identity"))
	b.WriteString("\n\n")

	// System Name
	b.WriteString("  ")
	b.WriteString(cursor(m.subCursor == 0))
	if m.subCursor == 0 {
		b.WriteString(focusedStyle.Render("System Name"))
	} else {
		b.WriteString(labelStyle.Render("System Name"))
	}
	b.WriteString("    ")
	b.WriteString(m.systemNameInput.View())
	b.WriteString("\n")

	// Description
	b.WriteString("  ")
	b.WriteString(cursor(m.subCursor == 1))
	if m.subCursor == 1 {
		b.WriteString(focusedStyle.Render("Description"))
	} else {
		b.WriteString(labelStyle.Render("Description"))
	}
	b.WriteString("    ")
	b.WriteString(m.systemDescInput.View())
	b.WriteString("\n\n")

	// Protocol Broadcasting
	b.WriteString("  ")
	b.WriteString(sectionStyle.Render("Protocol Broadcasting"))
	b.WriteString("\n\n")

	// CDP Broadcast
	b.WriteString("  ")
	b.WriteString(cursor(m.subCursor == 2))
	b.WriteString(checkbox(m.cdpBroadcast, m.subCursor == 2))
	b.WriteString(" ")
	if m.subCursor == 2 {
		b.WriteString(focusedStyle.Render("CDP"))
	} else {
		b.WriteString(labelStyle.Render("CDP"))
	}
	b.WriteString("        ")

	// LLDP Broadcast
	b.WriteString(checkbox(m.lldpBroadcast, m.subCursor == 3))
	b.WriteString(" ")
	if m.subCursor == 3 {
		b.WriteString(focusedStyle.Render("LLDP"))
	} else {
		b.WriteString(labelStyle.Render("LLDP"))
	}
	b.WriteString("\n")

	// Start on Launch
	b.WriteString("  ")
	b.WriteString(cursor(m.subCursor == 4))
	b.WriteString(checkbox(m.broadcastOnStartup, m.subCursor == 4))
	b.WriteString(" ")
	if m.subCursor == 4 {
		b.WriteString(focusedStyle.Render("Start on launch"))
	} else {
		b.WriteString(labelStyle.Render("Start on launch"))
	}
	b.WriteString("\n\n")

	// Timing
	b.WriteString("  ")
	b.WriteString(sectionStyle.Render("Timing"))
	b.WriteString("\n\n")

	// Interval
	b.WriteString("  ")
	b.WriteString(cursor(m.subCursor == 5))
	if m.subCursor == 5 {
		b.WriteString(focusedStyle.Render("Interval"))
	} else {
		b.WriteString(labelStyle.Render("Interval"))
	}
	b.WriteString("       ")
	b.WriteString(m.intervalInput.View())
	b.WriteString(dimStyle.Render(" seconds"))
	b.WriteString("\n")

	// TTL
	b.WriteString("  ")
	b.WriteString(cursor(m.subCursor == 6))
	if m.subCursor == 6 {
		b.WriteString(focusedStyle.Render("TTL"))
	} else {
		b.WriteString(labelStyle.Render("TTL"))
	}
	b.WriteString("            ")
	b.WriteString(m.ttlInput.View())
	b.WriteString(dimStyle.Render(" seconds"))
	b.WriteString("\n\n")

	// Capabilities
	b.WriteString("  ")
	b.WriteString(sectionStyle.Render("Capabilities (advertised)"))
	b.WriteString("\n\n")

	// Router
	b.WriteString("  ")
	b.WriteString(cursor(m.subCursor == 7))
	b.WriteString(checkbox(m.capRouter, m.subCursor == 7))
	b.WriteString(" ")
	if m.subCursor == 7 {
		b.WriteString(focusedStyle.Render("Router"))
	} else {
		b.WriteString(labelStyle.Render("Router"))
	}
	b.WriteString("     ")

	// Bridge
	b.WriteString(checkbox(m.capBridge, m.subCursor == 8))
	b.WriteString(" ")
	if m.subCursor == 8 {
		b.WriteString(focusedStyle.Render("Bridge"))
	} else {
		b.WriteString(labelStyle.Render("Bridge"))
	}
	b.WriteString("     ")

	// Station
	b.WriteString(checkbox(m.capStation, m.subCursor == 9))
	b.WriteString(" ")
	if m.subCursor == 9 {
		b.WriteString(focusedStyle.Render("Station"))
	} else {
		b.WriteString(labelStyle.Render("Station"))
	}
	b.WriteString("\n\n")

	// Back button
	b.WriteString("  ")
	b.WriteString(cursor(m.subCursor == 10))
	if m.subCursor == 10 {
		b.WriteString(focusedStyle.Render("[Back]"))
	} else {
		b.WriteString(labelStyle.Render("[Back]"))
	}
	b.WriteString("\n")

	return b.String()
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
	// ███╗   ██╗██████╗  ██████╗ ██████╗
	// ████╗  ██║██╔══██╗██╔═══██╗██╔══██╗
	// ██╔██╗ ██║██████╔╝██║   ██║██████╔╝
	// ██║╚██╗██║██╔══██╗██║   ██║██╔══██╗
	// ██║ ╚████║██████╔╝╚██████╔╝██║  ██║
	// ╚═╝  ╚═══╝╚═════╝  ╚═════╝ ╚═╝  ╚═╝

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
