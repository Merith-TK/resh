package cmd

import (
	"fmt"

	"github.com/Merith-TK/resh/pkg/shell"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// scriptCmd represents the script command
var scriptCmd = &cobra.Command{
	Use:   "script <file.lua>",
	Short: "Run a Lua script without entering REPL",
	Long: `Execute a Lua script file against a connected Resonite world.
The script has access to all shell functions (cd, ls, inspect, etc.)
and runs in a non-interactive mode.

Example:
  resonite-sh script inspect_resh_data.lua --url ws://localhost:39015`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		scriptPath := args[0]
		return runScriptCommand(scriptPath)
	},
}

func init() {
	rootCmd.AddCommand(scriptCmd)
}

// runScriptCommand executes a script file
func runScriptCommand(scriptPath string) error {
	// Get connection URL
	url := viper.GetString("connection.url")
	if url == "" {
		url = "ws://localhost:39015"
	}

	fmt.Printf("Connecting to Resonite at %s...\n", url)

	// Create and connect client
	client, err := connectClient(url)
	if err != nil {
		return err
	}
	defer client.Disconnect()

	fmt.Println("✓ Connected!")
	fmt.Printf("Running script: %s\n\n", scriptPath)

	// Initialize state
	state, err := shell.InitializeState(client)
	if err != nil {
		return err
	}

	// Run the script
	if err := shell.RunScript(client, state, scriptPath); err != nil {
		return fmt.Errorf("script execution failed: %w", err)
	}

	return nil
}
