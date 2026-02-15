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

		// Get root slot ID for blacklisting
		rootSlotResp, err := client.GetSlot("Root", false, 0)
		if err != nil {
			return fmt.Errorf("failed to get root slot: %w", err)
		}
		rootSlotID := rootSlotResp.Data.ID

		// Track current slot
		currentSlot := "Root"
		currentPath := "/Root"

		// Create readline instance
		rl, err := readline.New(currentPath + " $ ")
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

			case "ls":
				listSlotContents(client, currentSlot, rootSlotID)

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
	fmt.Println("  ls               - List slots and components")
	fmt.Println("  exit, quit       - Exit shell")
	fmt.Println()
	fmt.Println("More navigation commands will be added in next stage...")
}

func listSlotContents(client *resolink.Client, slotID string, rootSlotID string) {
	// Get the current slot details (with components)
	slotResp, err := client.GetSlot(slotID, true, 0)
	if err != nil {
		fmt.Printf("Error getting slot: %v\n", err)
		return
	}

	slot := slotResp.Data

	// Check if this is the root slot (blacklist components)
	isRootSlot := (slotID == rootSlotID || slotID == "Root")

	// Display slot name
	fmt.Println()
	if slot.Name != nil {
		fmt.Printf("Contents of slot: %s\n", slot.Name.Value)
	} else {
		fmt.Printf("Contents of slot: %s\n", slotID)
	}
	fmt.Println()

	// List child slots
	if len(slot.Children) > 0 {
		fmt.Println("Slots:")
		for _, child := range slot.Children {
			name := "<unnamed>"
			if child.Name != nil {
				name = child.Name.Value
			}

			// Determine persistence indicator (text-based)
			persistenceIndicator := "P"
			if child.IsPersistent != nil {
				if child.IsPersistent.Value {
					persistenceIndicator = "P" // Persistent
				} else {
					persistenceIndicator = "T" // Temporary/Non-persistent
				}
			}

			// Determine color based on active state
			// Active slots: cyan (like symlinks)
			// Inactive slots: blue (like directories)
			color := "\033[0;36m" // Cyan for active (default)
			resetColor := "\033[0m"
			if child.IsActive != nil && !child.IsActive.Value {
				color = "\033[0;34m" // Blue for inactive
			}

			fmt.Printf("  %s[%s] %s%s%s\n", persistenceIndicator, "slot", color, name, resetColor)
		}
		fmt.Println()
	}

	// List components (blacklist Root slot components)
	if isRootSlot {
		fmt.Println("[Root slot - components hidden for safety]")
		fmt.Println()
	} else if len(slot.Components) > 0 {
		fmt.Println("Components:")
		for _, comp := range slot.Components {
			// Component type
			compType := comp.ComponentType
			if compType == "" {
				compType = "<unknown type>"
			}

			// White text for components
			color := "\033[0;37m"
			resetColor := "\033[0m"

			fmt.Printf("  [comp] %s%s%s\n", color, compType, resetColor)
		}
		fmt.Println()
	} else if len(slot.Children) == 0 {
		fmt.Println("(empty slot - no children or components)")
		fmt.Println()
	}
}

func init() {
	rootCmd.AddCommand(replCmd)
}
