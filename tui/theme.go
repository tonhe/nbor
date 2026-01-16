package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Theme represents a Base16 color theme
type Theme struct {
	Name string

	// Base16 colors
	Base00 lipgloss.Color // Background
	Base01 lipgloss.Color // Lighter background
	Base02 lipgloss.Color // Selection background
	Base03 lipgloss.Color // Comments, invisibles
	Base04 lipgloss.Color // Dark foreground
	Base05 lipgloss.Color // Default foreground
	Base06 lipgloss.Color // Light foreground
	Base07 lipgloss.Color // Lightest foreground
	Base08 lipgloss.Color // Red
	Base09 lipgloss.Color // Orange
	Base0A lipgloss.Color // Yellow
	Base0B lipgloss.Color // Green
	Base0C lipgloss.Color // Cyan
	Base0D lipgloss.Color // Blue
	Base0E lipgloss.Color // Magenta
	Base0F lipgloss.Color // Brown
}

// SolarizedDark is the default Solarized Dark theme
var SolarizedDark = Theme{
	Name:   "Solarized Dark",
	Base00: lipgloss.Color("#002b36"), // Background
	Base01: lipgloss.Color("#073642"), // Lighter background
	Base02: lipgloss.Color("#586e75"), // Selection background
	Base03: lipgloss.Color("#657b83"), // Comments
	Base04: lipgloss.Color("#839496"), // Dark foreground
	Base05: lipgloss.Color("#93a1a1"), // Default foreground
	Base06: lipgloss.Color("#eee8d5"), // Light foreground
	Base07: lipgloss.Color("#fdf6e3"), // Lightest foreground
	Base08: lipgloss.Color("#dc322f"), // Red
	Base09: lipgloss.Color("#cb4b16"), // Orange
	Base0A: lipgloss.Color("#b58900"), // Yellow
	Base0B: lipgloss.Color("#859900"), // Green
	Base0C: lipgloss.Color("#2aa198"), // Cyan
	Base0D: lipgloss.Color("#268bd2"), // Blue
	Base0E: lipgloss.Color("#6c71c4"), // Violet
	Base0F: lipgloss.Color("#d33682"), // Magenta
}

// Gruvbox is an alternative theme
var Gruvbox = Theme{
	Name:   "Gruvbox Dark",
	Base00: lipgloss.Color("#282828"),
	Base01: lipgloss.Color("#3c3836"),
	Base02: lipgloss.Color("#504945"),
	Base03: lipgloss.Color("#665c54"),
	Base04: lipgloss.Color("#bdae93"),
	Base05: lipgloss.Color("#d5c4a1"),
	Base06: lipgloss.Color("#ebdbb2"),
	Base07: lipgloss.Color("#fbf1c7"),
	Base08: lipgloss.Color("#fb4934"),
	Base09: lipgloss.Color("#fe8019"),
	Base0A: lipgloss.Color("#fabd2f"),
	Base0B: lipgloss.Color("#b8bb26"),
	Base0C: lipgloss.Color("#8ec07c"),
	Base0D: lipgloss.Color("#83a598"),
	Base0E: lipgloss.Color("#d3869b"),
	Base0F: lipgloss.Color("#d65d0e"),
}

// DefaultTheme is the currently active theme
var DefaultTheme = SolarizedDark

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
	TableHeader     lipgloss.Style
	TableRow        lipgloss.Style
	TableRowStale   lipgloss.Style
	TableRowNew     lipgloss.Style
	TableCell       lipgloss.Style
	TableCellStale  lipgloss.Style
	TableSelected   lipgloss.Style

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
			Foreground(theme.Base05).
			PaddingRight(2),

		TableCellStale: lipgloss.NewStyle().
			Foreground(theme.Base03).
			PaddingRight(2),

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
