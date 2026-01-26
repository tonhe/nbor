package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"nbor/config"
	"nbor/types"
)

// AppState represents the current state of the application
type AppState int

const (
	StateSelectInterface AppState = iota
	StateConfigMenu
	StateCapturing
)

// AppModel is the main application model
type AppModel struct {
	state      AppState
	picker     InterfacePickerModel
	configMenu ConfigMenuModel
	neighbors  NeighborTableModel
	store      *types.NeighborStore
	config     *config.Config
	err        error
	width      int
	height     int

	// Channel for sending selected interface back to main
	selectChan chan<- types.InterfaceInfo
}

// NewApp creates a new application model (starts at interface picker)
func NewApp(interfaces []types.InterfaceInfo, store *types.NeighborStore, cfg *config.Config, selectChan chan<- types.InterfaceInfo) AppModel {
	return AppModel{
		state:      StateSelectInterface,
		picker:     NewInterfacePicker(interfaces),
		store:      store,
		config:     cfg,
		selectChan: selectChan,
	}
}

// NewAppAtInterfacePicker creates a new application model starting at interface picker
// Used when interface is specified via CLI
func NewAppAtInterfacePicker(interfaces []types.InterfaceInfo, store *types.NeighborStore, cfg *config.Config, selectChan chan<- types.InterfaceInfo) AppModel {
	return AppModel{
		state:      StateSelectInterface,
		picker:     NewInterfacePicker(interfaces),
		store:      store,
		config:     cfg,
		selectChan: selectChan,
	}
}

// Init initializes the application
func (m AppModel) Init() tea.Cmd {
	switch m.state {
	case StateSelectInterface:
		return tea.Batch(
			m.picker.Init(),
			tea.EnterAltScreen,
		)
	case StateConfigMenu:
		return tea.Batch(
			m.configMenu.Init(),
			tea.EnterAltScreen,
		)
	default:
		return tea.EnterAltScreen
	}
}

// ErrorMsg represents an error message
type ErrorMsg struct {
	Err error
}

// StartCaptureMsg signals to start capturing on the selected interface
type StartCaptureMsg struct {
	Interface types.InterfaceInfo
	LogPath   string
}

// Update handles messages for the application
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Forward to current view
		switch m.state {
		case StateSelectInterface:
			var cmd tea.Cmd
			newPicker, cmd := m.picker.Update(msg)
			m.picker = newPicker.(InterfacePickerModel)
			return m, cmd
		case StateConfigMenu:
			var cmd tea.Cmd
			newConfig, cmd := m.configMenu.Update(msg)
			m.configMenu = newConfig.(ConfigMenuModel)
			return m, cmd
		case StateCapturing:
			var cmd tea.Cmd
			m.neighbors, cmd = m.neighbors.Update(msg)
			return m, cmd
		}

	case GoToConfigMenuMsg:
		// Navigate to config menu from capture screen
		m.state = StateConfigMenu
		m.configMenu = NewConfigMenu(m.config)
		return m, m.configMenu.Init()

	case ConfigSavedMsg:
		// Config was saved, return to capturing
		m.config = msg.Config
		// Update the neighbors model with new config
		m.neighbors.config = m.config
		m.neighbors.broadcasting = m.config.CDPBroadcast || m.config.LLDPBroadcast
		m.state = StateCapturing
		return m, m.neighbors.Init()

	case ConfigCancelledMsg:
		// Config was cancelled, return to capturing
		m.state = StateCapturing
		return m, m.neighbors.Init()

	case InterfaceSelectedMsg:
		// Interface was selected, send to channel
		if m.selectChan != nil {
			// Non-blocking send
			select {
			case m.selectChan <- msg.Interface:
			default:
			}
		}
		return m, nil

	case StartCaptureMsg:
		// Transition to capturing state
		m.state = StateCapturing
		m.neighbors = NewNeighborTable(m.store, msg.Interface, msg.LogPath, m.config)
		m.neighbors.width = m.width
		m.neighbors.height = m.height
		return m, m.neighbors.Init()

	case ErrorMsg:
		m.err = msg.Err
		return m, tea.Quit

	case tea.KeyMsg:
		// Handle global quit
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	// Route messages to current view
	switch m.state {
	case StateSelectInterface:
		var cmd tea.Cmd
		newPicker, cmd := m.picker.Update(msg)
		m.picker = newPicker.(InterfacePickerModel)
		return m, cmd

	case StateConfigMenu:
		var cmd tea.Cmd
		newConfig, cmd := m.configMenu.Update(msg)
		m.configMenu = newConfig.(ConfigMenuModel)
		return m, cmd

	case StateCapturing:
		var cmd tea.Cmd
		m.neighbors, cmd = m.neighbors.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View renders the application
func (m AppModel) View() string {
	if m.err != nil {
		return DefaultStyles.StatusError.Render(fmt.Sprintf("Error: %v\n", m.err))
	}

	switch m.state {
	case StateSelectInterface:
		return m.picker.View()
	case StateConfigMenu:
		return m.configMenu.View()
	case StateCapturing:
		return m.neighbors.View()
	}

	return ""
}

// GetStore returns the neighbor store
func (m *AppModel) GetStore() *types.NeighborStore {
	return m.store
}

// SendNewNeighbor sends a new neighbor message to the TUI
func SendNewNeighbor(n *types.Neighbor) tea.Msg {
	return NewNeighborMsg{Neighbor: n}
}
