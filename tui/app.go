package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"nbor/types"
)

// AppState represents the current state of the application
type AppState int

const (
	StateSelectInterface AppState = iota
	StateCapturing
)

// AppModel is the main application model
type AppModel struct {
	state     AppState
	picker    InterfacePickerModel
	neighbors NeighborTableModel
	store     *types.NeighborStore
	err       error
	width     int
	height    int

	// Channel for sending selected interface back to main
	selectChan chan<- types.InterfaceInfo
}

// NewApp creates a new application model
func NewApp(interfaces []types.InterfaceInfo, store *types.NeighborStore, selectChan chan<- types.InterfaceInfo) AppModel {
	return AppModel{
		state:      StateSelectInterface,
		picker:     NewInterfacePicker(interfaces),
		store:      store,
		selectChan: selectChan,
	}
}

// Init initializes the application
func (m AppModel) Init() tea.Cmd {
	return tea.Batch(
		m.picker.Init(),
		tea.EnterAltScreen,
	)
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
		if m.state == StateSelectInterface {
			var cmd tea.Cmd
			newPicker, cmd := m.picker.Update(msg)
			m.picker = newPicker.(InterfacePickerModel)
			return m, cmd
		} else {
			var cmd tea.Cmd
			m.neighbors, cmd = m.neighbors.Update(msg)
			return m, cmd
		}

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
		m.neighbors = NewNeighborTable(m.store, msg.Interface, msg.LogPath)
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
