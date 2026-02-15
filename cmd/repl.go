package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/Merith-TK/resonite-sh/pkg/resolink"
	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// replCmd represents the repl command
var replCmd = &cobra.Command{
	Use:   "repl",
	Short: "Start interactive REPL shell",
	Long: `Start an interactive REPL (Read-Eval-Print Loop) shell for
navigating and manipulating Resonite worlds.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		url := viper.GetString("connection.url")
		if url == "" {
			url = "ws://localhost:39015"
		}

		fmt.Printf("Connecting to Resonite at %s...\n", url)

		// Create client
		timeout := 30 * time.Second
		client := resolink.NewClient(url, timeout)

		// Connect
		if err := client.Connect(); err != nil {
			return fmt.Errorf("failed to connect: %w", err)
		}
		defer client.Disconnect()

		fmt.Println("✓ Connected!")
		fmt.Println("Type 'help' for commands, 'exit' to quit")
		fmt.Println()

		// Create readline instance
		rl, err := readline.New("/Root $ ")
		if err != nil {
			return fmt.Errorf("failed to create readline: %w", err)
		}
		defer rl.Close()

		// REPL loop
		for {
			line, err := rl.Readline()
			if err != nil {
				break
			}

			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			parts := strings.Fields(line)
			if len(parts) == 0 {
				continue
			}

			cmdName := parts[0]

			switch cmdName {
			case "exit", "quit":
				fmt.Println("Goodbye!")
				return nil

			case "help":
				printHelp()

			case "test":
				// Test command - get Root slot
				rootSlot, err := client.GetSlot("Root", false, 0)
				if err != nil {
					fmt.Printf("Error: %v\n", err)
				} else {
					fmt.Printf("✓ Root slot: %s\n", rootSlot.Data.ID)
					if rootSlot.Data.Name != nil {
						fmt.Printf("  Name: %s\n", rootSlot.Data.Name.Value)
					}
				}

			default:
				fmt.Printf("Unknown command: %s (type 'help' for commands)\n", cmdName)
			}
		}

		return nil
	},
}

func printHelp() {
	fmt.Println("Resonite Shell (RESH) - Commands:")
	fmt.Println()
	fmt.Println("Basic:")
	fmt.Println("  help             - Show this help")
	fmt.Println("  test             - Test connection (get Root slot)")
	fmt.Println("  exit, quit       - Exit shell")
	fmt.Println()
	fmt.Println("Navigation commands will be added in next stage...")
}

func init() {
	rootCmd.AddCommand(replCmd)
}
