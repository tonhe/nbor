package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"nbor/version"
)

// MainMenuItem represents a menu option
type MainMenuItem int

const (
	MenuItemStartCapture MainMenuItem = iota
	MenuItemConfiguration
	MenuItemQuit
)

// MainMenuModel is the model for the main menu screen
type MainMenuModel struct {
	cursor int
	items  []MainMenuItem
	width  int
	height int
	styles Styles
}

// NewMainMenu creates a new main menu model
func NewMainMenu() MainMenuModel {
	return MainMenuModel{
		cursor: 0,
		items: []MainMenuItem{
			MenuItemStartCapture,
			MenuItemConfiguration,
			MenuItemQuit,
		},
		styles: DefaultStyles,
	}
}

// Init initializes the main menu
func (m MainMenuModel) Init() tea.Cmd {
	return nil
}

// mainMenuKeyMap defines the key bindings for the main menu
type mainMenuKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Select key.Binding
	Quit   key.Binding
}

var mainMenuKeys = mainMenuKeyMap{
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
		key.WithHelp("q", "quit"),
	),
}

// GoToInterfacePickerMsg signals to navigate to interface picker
type GoToInterfacePickerMsg struct{}

// GoToConfigMenuMsg signals to navigate to config menu
type GoToConfigMenuMsg struct{}

// Update handles messages for the main menu
func (m MainMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, mainMenuKeys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, mainMenuKeys.Down):
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case key.Matches(msg, mainMenuKeys.Select):
			return m.handleSelect()
		case key.Matches(msg, mainMenuKeys.Quit):
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// handleSelect handles menu item selection
func (m MainMenuModel) handleSelect() (tea.Model, tea.Cmd) {
	switch m.items[m.cursor] {
	case MenuItemStartCapture:
		return m, func() tea.Msg {
			return GoToInterfacePickerMsg{}
		}
	case MenuItemConfiguration:
		return m, func() tea.Msg {
			return GoToConfigMenuMsg{}
		}
	case MenuItemQuit:
		return m, tea.Quit
	}
	return m, nil
}

// View renders the main menu
func (m MainMenuModel) View() string {
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
func (m MainMenuModel) renderHeader() string {
	theme := DefaultTheme
	bg := theme.Base01

	// Single space with background
	sp := lipgloss.NewStyle().Background(bg).Render(" ")

	// Left side: app name and version
	nameStyle := lipgloss.NewStyle().
		Foreground(theme.Base0C).
		Background(bg).
		Bold(true)
	versionStyle := lipgloss.NewStyle().
		Foreground(theme.Base03).
		Background(bg)
	leftPart := nameStyle.Render("nbor") + sp + versionStyle.Render("v"+version.Version)

	// Right side: subtitle
	subtitleStyle := lipgloss.NewStyle().
		Foreground(theme.Base04).
		Background(bg)
	rightPart := subtitleStyle.Render("Network Neighbor Discovery")

	// Calculate spacing
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

// renderContent renders the menu content
func (m MainMenuModel) renderContent() string {
	theme := DefaultTheme
	var b strings.Builder

	b.WriteString("\n")

	// Menu items
	menuLabels := map[MainMenuItem]string{
		MenuItemStartCapture:  "Start Capturing",
		MenuItemConfiguration: "Configuration",
		MenuItemQuit:          "Quit",
	}

	menuDescriptions := map[MainMenuItem]string{
		MenuItemStartCapture:  "Select an interface and listen for neighbors",
		MenuItemConfiguration: "Configure listening, broadcasting, and identity",
		MenuItemQuit:          "Exit the application",
	}

	selectedStyle := lipgloss.NewStyle().
		Foreground(theme.Base0B).
		Bold(true)
	normalStyle := lipgloss.NewStyle().
		Foreground(theme.Base05)
	descStyle := lipgloss.NewStyle().
		Foreground(theme.Base03)
	cursorStyle := lipgloss.NewStyle().
		Foreground(theme.Base0C).
		Bold(true)

	for i, item := range m.items {
		label := menuLabels[item]
		desc := menuDescriptions[item]

		if i == m.cursor {
			b.WriteString("  ")
			b.WriteString(cursorStyle.Render(">"))
			b.WriteString(" ")
			b.WriteString(selectedStyle.Render(label))
			b.WriteString("\n")
			b.WriteString("    ")
			b.WriteString(descStyle.Render(desc))
		} else {
			b.WriteString("    ")
			b.WriteString(normalStyle.Render(label))
			b.WriteString("\n")
			b.WriteString("    ")
			b.WriteString(descStyle.Render(desc))
		}
		b.WriteString("\n\n")
	}

	return b.String()
}

// renderFooter renders the footer bar
func (m MainMenuModel) renderFooter() string {
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

	// Pad to full width
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
