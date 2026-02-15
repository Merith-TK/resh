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
		return startREPL()
	},
}

func init() {
	rootCmd.AddCommand(replCmd)
}

// startREPL initializes and runs the REPL loop
func startREPL() error {
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
	fmt.Println("Type 'help' for commands, 'exit' to quit")
	fmt.Println()

	// Initialize REPL state
	state, err := initializeREPLState(client)
	if err != nil {
		return err
	}

	// Create readline instance with autocomplete
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          state.currentPath + " $ ",
		AutoComplete:    newCompleter(client, state),
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		return fmt.Errorf("failed to create readline: %w", err)
	}
	defer rl.Close()

	// Run REPL loop
	return runREPLLoop(rl, client, state)
}

// connectClient creates and connects a ResoLink client
func connectClient(url string) (*resolink.Client, error) {
	timeout := 30 * time.Second
	client := resolink.NewClient(url, timeout)

	if err := client.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return client, nil
}

// replState holds the current state of the REPL session
type replState struct {
	currentSlot string
	currentPath string
	rootSlotID  string
}

// initializeREPLState sets up the initial REPL state
func initializeREPLState(client *resolink.Client) (*replState, error) {
	// Get root slot ID for blacklisting
	rootSlotResp, err := client.GetSlot("Root", false, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get root slot: %w", err)
	}

	return &replState{
		currentSlot: "Root",
		currentPath: "/Root",
		rootSlotID:  rootSlotResp.Data.ID,
	}, nil
}

// runREPLLoop runs the main REPL loop
func runREPLLoop(rl *readline.Instance, client *resolink.Client, state *replState) error {
	for {
		line, err := rl.Readline()
		if err != nil {
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := parseCommandLine(line)
		if len(parts) == 0 {
			continue
		}

		cmdName := parts[0]
		args := parts[1:]

		// Handle command
		if shouldExit := handleCommand(cmdName, args, client, state); shouldExit {
			fmt.Println("Goodbye!")
			return nil
		}

		// Update prompt and autocomplete
		rl.SetPrompt(state.currentPath + " $ ")
		rl.Config.AutoComplete = newCompleter(client, state)
	}

	return nil
}

// handleCommand processes a single command and returns whether to exit
func handleCommand(cmdName string, args []string, client *resolink.Client, state *replState) bool {
	switch cmdName {
	case "exit", "quit":
		return true

	case "help":
		printHelp()

	case "test":
		testConnection(client)

	case "ls":
		listSlotContents(client, state.currentSlot, state.rootSlotID)

	case "cd":
		if len(args) == 0 {
			fmt.Println("cd: missing argument")
			fmt.Println("Usage: cd <slot>, cd .., or cd /")
			fmt.Println("Tip: Use quotes for names with spaces: cd \"Slot Name\"")
		} else {
			if err := changeDirectory(client, args[0], state); err != nil {
				fmt.Printf("cd: %v\n", err)
				// If the error is "not found" and there are multiple args, suggest quoting
				if strings.Contains(err.Error(), "not found") && len(args) > 1 {
					fmt.Println("Tip: Use quotes for names with spaces: cd \"" + strings.Join(args, " ") + "\"")
				}
			}
		}

	default:
		fmt.Printf("Unknown command: %s (type 'help' for commands)\n", cmdName)
	}

	return false
}
