package tui

import "github.com/charmbracelet/lipgloss"

// renderCheckbox renders a checkbox with proper styling based on state
func renderCheckbox(checked, focused bool, theme Theme) string {
	labelStyle := lipgloss.NewStyle().Foreground(theme.Base05)
	focusedStyle := lipgloss.NewStyle().Foreground(theme.Base0B).Bold(true)

	style := labelStyle
	if focused {
		style = focusedStyle
	}

	if checked {
		return style.Render("[x]")
	}
	return style.Render("[ ]")
}

// renderCursor renders the cursor indicator for menu items
func renderCursor(focused bool, theme Theme) string {
	if focused {
		cursorStyle := lipgloss.NewStyle().Foreground(theme.Base0C).Bold(true)
		return cursorStyle.Render(">") + " "
	}
	return "  "
}

// renderLabel renders a label with appropriate focus styling
func renderLabel(text string, focused bool, theme Theme) string {
	if focused {
		focusedStyle := lipgloss.NewStyle().Foreground(theme.Base0B).Bold(true)
		return focusedStyle.Render(text)
	}
	labelStyle := lipgloss.NewStyle().Foreground(theme.Base05)
	return labelStyle.Render(text)
}

// findRowPosition finds the row and column position for a cursor value in a 2D grid
// Returns (row, col) where row and col are 0-indexed
func findRowPosition(cursor int, rows [][]int) (row, col int) {
	for r, rowFields := range rows {
		for c, field := range rowFields {
			if field == cursor {
				return r, c
			}
		}
	}
	return 0, 0
}

// navigateGrid handles 2D grid navigation and returns the new cursor position
// direction: -1 for up/left, +1 for down/right
// horizontal: true for left/right, false for up/down
func navigateGrid(cursor int, rows [][]int, direction int, horizontal bool) int {
	row, col := findRowPosition(cursor, rows)

	if horizontal {
		// Move left/right within the current row
		newCol := col + direction
		if newCol >= 0 && newCol < len(rows[row]) {
			return rows[row][newCol]
		}
		// Stay at current position if can't move
		return cursor
	}

	// Move up/down between rows
	newRow := row + direction
	if newRow < 0 {
		newRow = len(rows) - 1
	} else if newRow >= len(rows) {
		newRow = 0
	}

	// Keep column position if possible, otherwise go to last item in row
	if col >= len(rows[newRow]) {
		col = len(rows[newRow]) - 1
	}

	return rows[newRow][col]
}
