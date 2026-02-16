package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/Merith-TK/resonite-sh/pkg/resolink"
	"github.com/Merith-TK/resonite-sh/pkg/shell"
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
	state, err := shell.InitializeState(client)
	if err != nil {
		return err
	}

	// Create readline instance with autocomplete
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          state.CurrentPath + " $ ",
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

// runREPLLoop runs the main REPL loop
func runREPLLoop(rl *readline.Instance, client *resolink.Client, state *shell.State) error {
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
		rl.SetPrompt(state.CurrentPath + " $ ")
		rl.Config.AutoComplete = newCompleter(client, state)
	}

	return nil
}

// handleCommand processes a single command and returns whether to exit
func handleCommand(cmdName string, args []string, client *resolink.Client, state *shell.State) bool {
	switch cmdName {
	case "exit", "quit":
		return true

	case "help":
		printHelp()

	case "test":
		testConnection(client)

	case "ls":
		listSlotContents(client, state)

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

	case "inspect":
		if len(args) == 0 {
			fmt.Println("inspect: missing component ID")
			fmt.Println("Usage: inspect <component_id>")
		} else {
			inspectComponent(client, args[0])
		}

	case "set":
		if len(args) < 2 {
			fmt.Println("set: missing arguments")
			fmt.Println("Usage: set <id> <member_id>=<value>  (for components)")
			fmt.Println("       set <id> <property>=<value>   (for slots)")
			fmt.Println("Example: set ID_100 ID_812=2")
			fmt.Println("Example: set ID_101 Name=\"My Slot\"")
		} else {
			setComponentMember(client, args)
		}

	case "bookmark":
		if len(args) == 0 {
			fmt.Println("bookmark: missing name")
			fmt.Println("Usage: bookmark <name>")
		} else {
			saveBookmark(client, state, args[0])
		}

	case "goto":
		if len(args) == 0 {
			fmt.Println("goto: missing bookmark name")
			fmt.Println("Usage: goto <name>")
		} else {
			gotoBookmark(client, state, args[0])
		}

	case "bookmarks":
		listBookmarks(state)

	case "unbookmark":
		if len(args) == 0 {
			fmt.Println("unbookmark: missing name")
			fmt.Println("Usage: unbookmark <name>")
		} else {
			deleteBookmark(state, args[0])
		}

	case "script":
		if len(args) == 0 {
			fmt.Println("script: missing file path")
			fmt.Println("Usage: script <file.lua>")
		} else {
			runScript(client, state, args[0])
		}

	default:
		fmt.Printf("Unknown command: %s (type 'help' for commands)\n", cmdName)
	}

	return false
}
