package cmd

import (
	"fmt"
	"time"

	"github.com/Merith-TK/resh/pkg/resolink"
	"github.com/Merith-TK/resh/pkg/shell"
	"github.com/Merith-TK/resh/pkg/tui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// tuiCmd represents the tui command
var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Start TUI (Terminal User Interface) mode",
	Long: `Start an interactive TUI for navigating and inspecting Resonite worlds.

The TUI provides a visual interface with two panels:
- Left: Tree view for navigation (slots and components)
- Right: Inspector for viewing and editing properties

Navigation:
  ↑/↓         Navigate tree or inspector fields
  Enter       Select item or start editing
  Tab         Switch between tree and inspector panels
  Alt+Enter   Focus on selected slot (change tree root)
  Alt+Bksp    Focus on parent slot
  :           Enter command mode
  q           Quit

The TUI is ideal for visual exploration and inspection of the world hierarchy.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return startTUI()
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}

// startTUI initializes and runs the TUI mode
func startTUI() error {
	// Get connection URL
	url := viper.GetString("connection.url")
	if url == "" {
		url = "ws://localhost:39015"
	}

	fmt.Printf("Connecting to Resonite at %s...\n", url)

	// Create and connect client
	timeout := 30 * time.Second
	client := resolink.NewClient(url, timeout)

	if err := client.Connect(); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Disconnect()

	fmt.Println("✓ Connected!")
	fmt.Println("Initializing TUI...")

	// Initialize state
	state, err := shell.InitializeState(client)
	if err != nil {
		return fmt.Errorf("failed to initialize state: %w", err)
	}

	// Create TUI model
	model := tui.NewModel(client, state)

	// Start TUI
	return tui.StartTUI(model)
}
