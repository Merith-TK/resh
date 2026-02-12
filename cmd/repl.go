package cmd

import (
	"fmt"

	"github.com/Merith-TK/resonite-sh/pkg/repl"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// replCmd represents the repl command
var replCmd = &cobra.Command{
	Use:   "repl",
	Short: "Start interactive REPL shell",
	Long: `Start an interactive REPL (Read-Eval-Print Loop) shell for
navigating and manipulating Resonite worlds.

The REPL provides familiar shell commands like cd, ls, pwd, cat, etc.
for working with slots and components.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		url := viper.GetString("connection.url")

		fmt.Printf("Connecting to Resonite at %s...\n", url)

		// Create and start REPL
		shell, err := repl.NewShell(url)
		if err != nil {
			return fmt.Errorf("failed to create shell: %w", err)
		}

		return shell.Run()
	},
}

func init() {
	rootCmd.AddCommand(replCmd)
}
