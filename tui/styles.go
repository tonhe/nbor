package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"nbor/version"
)

// Styles holds all the styled components for the TUI
type Styles struct {
	// App container
	App lipgloss.Style

	// Header styles
	Header      lipgloss.Style
	HeaderTitle lipgloss.Style
	HeaderInfo  lipgloss.Style

	// Footer styles
	Footer    lipgloss.Style
	FooterKey lipgloss.Style

	// Table styles
	TableHeader    lipgloss.Style
	TableRow       lipgloss.Style
	TableRowStale  lipgloss.Style
	TableRowNew    lipgloss.Style
	TableCell      lipgloss.Style
	TableCellStale lipgloss.Style
	TableSelected  lipgloss.Style

	// Interface picker styles
	PickerTitle    lipgloss.Style
	PickerItem     lipgloss.Style
	PickerSelected lipgloss.Style
	PickerUp       lipgloss.Style
	PickerDown     lipgloss.Style

	// Status styles
	StatusListening lipgloss.Style
	StatusError     lipgloss.Style
	StatusInfo      lipgloss.Style

	// Protocol badges
	BadgeCDP  lipgloss.Style
	BadgeLLDP lipgloss.Style
}

// NewStyles creates styled components based on the theme
func NewStyles(theme Theme) Styles {
	return Styles{
		// App container
		App: lipgloss.NewStyle().
			Background(theme.Base00),

		// Header
		Header: lipgloss.NewStyle().
			Background(theme.Base01).
			Foreground(theme.Base05).
			Padding(0, 1).
			Bold(true),

		HeaderTitle: lipgloss.NewStyle().
			Foreground(theme.Base0D).
			Bold(true),

		HeaderInfo: lipgloss.NewStyle().
			Foreground(theme.Base04),

		// Footer
		Footer: lipgloss.NewStyle().
			Background(theme.Base01).
			Foreground(theme.Base04).
			Padding(0, 1),

		FooterKey: lipgloss.NewStyle().
			Foreground(theme.Base0C).
			Bold(true),

		// Table
		TableHeader: lipgloss.NewStyle().
			Foreground(theme.Base0D).
			Bold(true).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(theme.Base02),

		TableRow: lipgloss.NewStyle().
			Foreground(theme.Base05),

		TableRowStale: lipgloss.NewStyle().
			Foreground(theme.Base03),

		TableRowNew: lipgloss.NewStyle().
			Foreground(theme.Base0B).
			Bold(true),

		TableCell: lipgloss.NewStyle().
			Foreground(theme.Base06), // Light foreground for better visibility

		TableCellStale: lipgloss.NewStyle().
			Foreground(theme.Base03),

		TableSelected: lipgloss.NewStyle().
			Background(theme.Base02).
			Foreground(theme.Base06),

		// Interface picker
		PickerTitle: lipgloss.NewStyle().
			Foreground(theme.Base0D).
			Bold(true).
			MarginBottom(1),

		PickerItem: lipgloss.NewStyle().
			Foreground(theme.Base05).
			PaddingLeft(2),

		PickerSelected: lipgloss.NewStyle().
			Foreground(theme.Base0B).
			Bold(true).
			PaddingLeft(0),

		PickerUp: lipgloss.NewStyle().
			Foreground(theme.Base0B), // Green

		PickerDown: lipgloss.NewStyle().
			Foreground(theme.Base08), // Red

		// Status
		StatusListening: lipgloss.NewStyle().
			Foreground(theme.Base0A).
			Italic(true),

		StatusError: lipgloss.NewStyle().
			Foreground(theme.Base08).
			Bold(true),

		StatusInfo: lipgloss.NewStyle().
			Foreground(theme.Base04),

		// Protocol badges
		BadgeCDP: lipgloss.NewStyle().
			Background(theme.Base0D).
			Foreground(theme.Base00).
			Padding(0, 1).
			Bold(true),

		BadgeLLDP: lipgloss.NewStyle().
			Background(theme.Base0B).
			Foreground(theme.Base00).
			Padding(0, 1).
			Bold(true),
	}
}

// DefaultStyles uses the default theme
var DefaultStyles = NewStyles(DefaultTheme)

// RenderHeader renders a consistent header bar across all views
// leftContent: app name/version, rightContent: screen title
func RenderHeader(leftContent, rightContent string, width int) string {
	theme := DefaultTheme
	bg := theme.Base01

	leftLen := lipgloss.Width(leftContent)
	rightLen := lipgloss.Width(rightContent)

	// Account for padding (1 on each side)
	availableWidth := width - 2
	gap := availableWidth - leftLen - rightLen
	if gap < 1 {
		gap = 1
	}

	// Build header content with background-colored spaces
	spaceStyle := lipgloss.NewStyle().Background(bg)
	headerContent := leftContent + spaceStyle.Render(strings.Repeat(" ", gap)) + rightContent

	// Apply background style
	headerStyle := lipgloss.NewStyle().
		Background(bg).
		Padding(0, 1).
		Width(width)

	return headerStyle.Render(headerContent)
}

// RenderFooter renders a consistent footer bar across all views
func RenderFooter(content string, width int) string {
	theme := DefaultTheme
	bg := theme.Base01

	contentWidth := lipgloss.Width(content)

	// Account for padding (1 on each side)
	availableWidth := width - 2
	padWidth := availableWidth - contentWidth
	if padWidth < 0 {
		padWidth = 0
	}

	// Build footer content with background-colored spaces
	spaceStyle := lipgloss.NewStyle().Background(bg)
	footerContent := content + spaceStyle.Render(strings.Repeat(" ", padWidth))

	// Apply background style
	footerStyle := lipgloss.NewStyle().
		Background(bg).
		Padding(0, 1).
		Width(width)

	return footerStyle.Render(footerContent)
}

// HeaderLeft returns styled app name and version for headers
func HeaderLeft() string {
	theme := DefaultTheme
	bg := theme.Base01

	nameStyle := lipgloss.NewStyle().
		Foreground(theme.Base0C).
		Background(bg).
		Bold(true)
	versionStyle := lipgloss.NewStyle().
		Foreground(theme.Base03).
		Background(bg)
	spaceStyle := lipgloss.NewStyle().Background(bg)

	return nameStyle.Render("nbor") + spaceStyle.Render(" ") + versionStyle.Render("v"+version.Version)
}

// HeaderTitle returns a styled title for the right side of headers
func HeaderTitle(title string) string {
	theme := DefaultTheme
	bg := theme.Base01

	titleStyle := lipgloss.NewStyle().
		Foreground(theme.Base0D).
		Background(bg).
		Bold(true)

	return titleStyle.Render(title)
}
