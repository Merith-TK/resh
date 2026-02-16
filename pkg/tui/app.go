package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Init initializes the Bubble Tea model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil
	}

	return m, nil
}

// handleKeyPress handles keyboard input
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global keys (work regardless of focus)
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	}

	// Command mode
	if m.Focus == FocusCommand {
		return m.handleCommandMode(msg)
	}

	// Field editing mode
	if m.EditingField {
		return m.handleFieldEditing(msg)
	}

	// Handle keys based on focus
	switch msg.String() {
	case "tab":
		// Toggle focus between tree and inspector
		if m.Focus == FocusTree {
			m.Focus = FocusInspector
		} else {
			m.Focus = FocusTree
		}
		return m, nil

	case ":":
		// Enter command mode (if not editing)
		if !m.EditingField {
			m.Focus = FocusCommand
			m.CommandBuffer = ""
		}
		return m, nil
	}

	// Tree-specific keys
	if m.Focus == FocusTree {
		return m.handleTreeKeys(msg)
	}

	// Inspector-specific keys
	if m.Focus == FocusInspector {
		return m.handleInspectorKeys(msg)
	}

	return m, nil
}

// handleTreeKeys handles keys when tree has focus
func (m Model) handleTreeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		m.MoveCursorUp()

	case "down", "j":
		m.MoveCursorDown()

	case "enter":
		m.SelectCurrentItem()

	case "alt+enter":
		if err := m.FocusOnSlot(); err != nil {
			// Error message already set
		}

	case "alt+backspace":
		if err := m.FocusOnParent(); err != nil {
			// Error message already set
		}
	}

	return m, nil
}

// handleInspectorKeys handles keys when inspector has focus
func (m Model) handleInspectorKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		m.MoveFieldCursorUp()

	case "down", "j":
		m.MoveFieldCursorDown()

	case "enter":
		m.StartEditingField()

	case "esc":
		// Go back to tree
		m.Focus = FocusTree
	}

	return m, nil
}

// handleFieldEditing handles field editing mode
func (m Model) handleFieldEditing(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.StopEditingField(true) // Save
		return m, nil

	case "esc":
		m.StopEditingField(false) // Cancel
		return m, nil

	case "backspace":
		if len(m.EditBuffer) > 0 {
			m.EditBuffer = m.EditBuffer[:len(m.EditBuffer)-1]
		}

	default:
		// Add character to buffer
		m.EditBuffer += msg.String()
	}

	return m, nil
}

// handleCommandMode handles command entry mode
func (m Model) handleCommandMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Execute command
		m.executeCommand(m.CommandBuffer)
		m.Focus = FocusTree
		m.CommandBuffer = ""
		return m, nil

	case "esc":
		// Cancel command mode
		m.Focus = FocusTree
		m.CommandBuffer = ""
		return m, nil

	case "backspace":
		if len(m.CommandBuffer) > 0 {
			m.CommandBuffer = m.CommandBuffer[:len(m.CommandBuffer)-1]
		}

	default:
		// Add character to command buffer
		m.CommandBuffer += msg.String()
	}

	return m, nil
}

// executeCommand executes a REPL command
func (m *Model) executeCommand(cmd string) {
	// TODO: Execute REPL commands
	// For now, just show it was received
	m.StatusMessage = fmt.Sprintf("Command: %s (not yet implemented)", cmd)
}

// View renders the UI
func (m Model) View() string {
	if m.Width == 0 || m.Height == 0 {
		return "Loading..."
	}

	// Calculate panel dimensions (50/50 split)
	panelWidth := (m.Width / 2) - 2
	panelHeight := m.Height - 4 // Leave room for status bar

	// Render panels
	leftPanel := m.RenderTree(panelWidth, panelHeight)
	rightPanel := m.RenderInspector(panelWidth, panelHeight)

	// Join panels side by side
	panels := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)

	// Render status/command bar
	statusBar := m.renderStatusBar()

	// Join everything vertically
	return lipgloss.JoinVertical(lipgloss.Left, panels, statusBar)
}

// renderStatusBar renders the bottom status/command bar
func (m Model) renderStatusBar() string {
	width := m.Width - 2

	if m.Focus == FocusCommand {
		// Command mode
		prompt := commandBarStyle.Render(fmt.Sprintf(" :%s", m.CommandBuffer))
		return lipgloss.NewStyle().Width(width).Render(prompt)
	}

	// Status mode
	var left, right string

	if m.ErrorMessage != "" {
		left = errorStyle.Render(fmt.Sprintf(" ERROR: %s", m.ErrorMessage))
	} else {
		left = statusBarStyle.Render(fmt.Sprintf(" %s", m.StatusMessage))
	}

	// Show connection status and shortcuts
	right = statusBarStyle.Render(" Connected | :cmd Tab:switch q:quit ")

	// Calculate spacing
	spacing := width - lipgloss.Width(left) - lipgloss.Width(right)
	if spacing < 0 {
		spacing = 0
	}

	statusLine := left + strings.Repeat(" ", spacing) + right
	return lipgloss.NewStyle().Width(width).Render(statusLine)
}

// StartTUI starts the TUI application
func StartTUI(model Model) error {
	// Load initial tree
	if err := model.LoadTreeItems(); err != nil {
		return fmt.Errorf("failed to load tree: %w", err)
	}

	// Create and run Bubble Tea program
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	return nil
}
