package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Panel styles
	focusedBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")) // Blue

	unfocusedBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")) // Gray

	// Tree item styles
	selectedTreeItem = lipgloss.NewStyle().
				Foreground(lipgloss.Color("205")). // Pink/Magenta
				Bold(true)

	normalTreeItem = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")) // Light gray

	slotStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("87")) // Cyan

	componentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("229")) // Yellow

	inactiveStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")). // Dark gray
			Strikethrough(true)

	// Inspector styles
	fieldNameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("211")). // Pink
			Bold(true)

	fieldValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")) // Light gray

	selectedFieldStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("205")). // Pink/Magenta
				Background(lipgloss.Color("237")). // Dark gray bg
				Bold(true)

	editingFieldStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("46")).  // Green
				Background(lipgloss.Color("237")). // Dark gray bg
				Bold(true)

	// Status bar styles
	statusBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("63")).  // Blue
			Foreground(lipgloss.Color("255")). // White
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")). // Red
			Bold(true)

	// Command bar style
	commandBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("237")). // Dark gray
			Foreground(lipgloss.Color("252"))  // Light gray

	// Title styles
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("211")). // Pink
			Bold(true).
			Padding(0, 1)

	// Help text
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")) // Gray
)

// ApplyBorder applies border style based on focus
func ApplyBorder(content string, focused bool, width, height int) string {
	style := unfocusedBorder
	if focused {
		style = focusedBorder
	}
	return style.Width(width).Height(height).Render(content)
}
