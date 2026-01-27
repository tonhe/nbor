package tui

import "github.com/charmbracelet/lipgloss"

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

// DefaultTheme is the currently active theme
var DefaultTheme = SolarizedDark

// SetTheme updates the default theme and regenerates all styles
func SetTheme(theme Theme) {
	DefaultTheme = theme
	DefaultStyles = NewStyles(theme)
}

// GetThemeByName returns a theme by its slug name, or nil if not found
func GetThemeByName(name string) *Theme {
	if theme, ok := Themes[name]; ok {
		return &theme
	}
	return nil
}

// ListThemes returns a sorted list of theme slugs and display names
func ListThemes() [][2]string {
	return [][2]string{
		{"solarized-dark", "Solarized Dark"},
		{"solarized-light", "Solarized Light"},
		{"gruvbox-dark", "Gruvbox Dark"},
		{"gruvbox-light", "Gruvbox Light"},
		{"dracula", "Dracula"},
		{"nord", "Nord"},
		{"one-dark", "One Dark"},
		{"monokai", "Monokai"},
		{"tokyo-night", "Tokyo Night"},
		{"catppuccin-mocha", "Catppuccin Mocha"},
		{"catppuccin-latte", "Catppuccin Latte"},
		{"everforest", "Everforest"},
		{"kanagawa", "Kanagawa"},
		{"rose-pine", "Rosé Pine"},
		{"tomorrow-night", "Tomorrow Night"},
		{"ayu-dark", "Ayu Dark"},
		{"horizon", "Horizon"},
		{"zenburn", "Zenburn"},
		{"palenight", "Palenight"},
		{"github-dark", "GitHub Dark"},
	}
}

// GetThemeCount returns the number of available themes
func GetThemeCount() int {
	return len(ListThemes())
}

// GetThemeByIndex returns the theme slug, display name, and Theme at the given index
// Returns empty strings and nil Theme if index is out of range
func GetThemeByIndex(idx int) (slug string, name string, theme *Theme) {
	themes := ListThemes()
	if idx < 0 || idx >= len(themes) {
		return "", "", nil
	}
	slug = themes[idx][0]
	name = themes[idx][1]
	theme = GetThemeByName(slug)
	return slug, name, theme
}

// GetThemeIndex returns the index of a theme by its slug name
// Returns -1 if the theme is not found
func GetThemeIndex(slug string) int {
	themes := ListThemes()
	for i, t := range themes {
		if t[0] == slug {
			return i
		}
	}
	return -1
}
