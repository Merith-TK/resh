package repl

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Merith-TK/resh/pkg/repl/commands"
	"github.com/Merith-TK/resh/pkg/resh"
	"github.com/Merith-TK/resh/pkg/resolink"
	"github.com/chzyer/readline"
)

// Shell represents the REPL shell
type Shell struct {
	client     *resolink.Client
	readline   *readline.Instance
	navigator  *commands.Navigator
	inspector  *commands.Inspector
	modifier   *commands.Modifier
	reshMgr    *resh.Manager
	running    bool
	modeSwitch func() // Callback to switch to TUI mode
}

// NewShell creates a new REPL shell
func NewShell(client *resolink.Client, reshMgr *resh.Manager) (*Shell, error) {
	// Create command handlers
	nav := commands.NewNavigator(client)
	insp := commands.NewInspector(client, nav)
	mod := commands.NewModifier(client, nav)

	rl, err := readline.New("> ")
	if err != nil {
		return nil, fmt.Errorf("failed to create readline: %w", err)
	}

	return &Shell{
		client:    client,
		readline:  rl,
		navigator: nav,
		inspector: insp,
		modifier:  mod,
		reshMgr:   reshMgr,
		running:   false,
	}, nil
}

// SetModeSwitchCallback sets the callback for switching to TUI mode
func (s *Shell) SetModeSwitchCallback(callback func()) {
	s.modeSwitch = callback
}

// Start starts the REPL shell
func (s *Shell) Start() error {
	s.running = true
	defer s.Stop()

	fmt.Println("Welcome to Resonite Shell (RESH)")
	fmt.Println("Type 'help' for commands, 'exit' to quit")
	fmt.Println("Press Tab to switch to TUI mode")
	fmt.Println()

	for s.running {
		// Update prompt with current path
		prompt := fmt.Sprintf("%s $ ", s.navigator.Pwd())
		s.readline.SetPrompt(prompt)

		line, err := s.readline.Readline()
		if err != nil {
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if err := s.executeCommand(line); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}

	return nil
}

// Stop stops the REPL shell
func (s *Shell) Stop() {
	s.running = false
	s.readline.Close()
}

// executeCommand executes a command
func (s *Shell) executeCommand(line string) error {
	parts := s.parseCommandLine(line)
	if len(parts) == 0 {
		return nil
	}

	cmd := parts[0]
	args := parts[1:]

	switch cmd {
	case "exit", "quit":
		s.running = false
		return nil

	case "help":
		s.printHelp()
		return nil

	// Navigation commands
	case "cd":
		if len(args) == 0 {
			return s.navigator.Cd("~")
		}
		return s.navigator.Cd(args[0])

	case "pwd":
		fmt.Println(s.navigator.Pwd())
		return nil

	case "ls":
		path := ""
		if len(args) > 0 {
			path = args[0]
		}
		slots, err := s.navigator.Ls(path)
		if err != nil {
			return err
		}
		s.printSlotList(slots)
		return nil

	case "tree":
		depth := 3
		if len(args) > 0 {
			d, err := strconv.Atoi(args[0])
			if err == nil {
				depth = d
			}
		}
		tree, err := s.navigator.Tree(depth)
		if err != nil {
			return err
		}
		fmt.Print(tree)
		return nil

	// Inspection commands
	case "cat":
		if len(args) == 0 {
			return fmt.Errorf("usage: cat <path>")
		}
		output, err := s.inspector.Cat(args[0])
		if err != nil {
			return err
		}
		fmt.Println(output)
		return nil

	case "stat":
		if len(args) == 0 {
			return fmt.Errorf("usage: stat <path>")
		}
		output, err := s.inspector.Stat(args[0])
		if err != nil {
			return err
		}
		fmt.Println(output)
		return nil

	case "find":
		if len(args) == 0 {
			return fmt.Errorf("usage: find <pattern> [root]")
		}
		root := ""
		if len(args) > 1 {
			root = args[1]
		}
		results, err := s.inspector.Find(args[0], root)
		if err != nil {
			return err
		}
		for _, r := range results {
			fmt.Printf("%s (s-%s) - %s\n", r.Name, r.RefID, r.Path)
		}
		return nil

	case "inspect":
		if len(args) < 2 {
			return fmt.Errorf("usage: inspect <path> <component-index>")
		}
		idx, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid component index: %v", err)
		}
		output, err := s.inspector.Inspect(args[0], idx)
		if err != nil {
			return err
		}
		fmt.Println(output)
		return nil

	// Modification commands
	case "mkdir":
		if len(args) == 0 {
			return fmt.Errorf("usage: mkdir <name> [parent-path]")
		}
		parent := ""
		if len(args) > 1 {
			parent = args[1]
		}
		return s.modifier.Mkdir(args[0], parent)

	case "touch":
		if len(args) == 0 {
			return fmt.Errorf("usage: touch <name>[.<component-type>] [parent-path]")
		}
		parent := ""
		if len(args) > 1 {
			parent = args[1]
		}
		return s.modifier.Touch(args[0], parent)

	case "rm":
		if len(args) == 0 {
			return fmt.Errorf("usage: rm [-r] <path>")
		}
		recursive := false
		path := args[0]
		if args[0] == "-r" {
			recursive = true
			if len(args) < 2 {
				return fmt.Errorf("usage: rm -r <path>")
			}
			path = args[1]
		}
		return s.modifier.Rm(path, recursive)

	case "edit":
		if len(args) < 3 {
			return fmt.Errorf("usage: edit <path> <property> <value> [type]")
		}
		valueType := "string"
		if len(args) > 3 {
			valueType = args[3]
		}
		value := s.parseValue(args[2], valueType)
		return s.modifier.Edit(args[0], args[1], value, valueType)

	case "set":
		if len(args) < 4 {
			return fmt.Errorf("usage: set <path> <comp-index> <field> <value> [type]")
		}
		idx, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid component index: %v", err)
		}
		valueType := "string"
		if len(args) > 4 {
			valueType = args[4]
		}
		value := s.parseValue(args[3], valueType)
		return s.modifier.Set(args[0], idx, args[2], value, valueType)

	// RESH variable commands
	case "var":
		return s.handleVarCommand(args)

	default:
		return fmt.Errorf("unknown command: %s (type 'help' for commands)", cmd)
	}
}

// handleVarCommand handles variable-related commands
func (s *Shell) handleVarCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: var <set|get|delete|list> ...")
	}

	subcmd := args[0]
	subargs := args[1:]

	switch subcmd {
	case "set":
		if len(subargs) < 3 {
			return fmt.Errorf("usage: var set <scope> <name> <value> [type]")
		}
		scope := resh.VariableScope(subargs[0])
		name := subargs[1]
		value := subargs[2]
		valueType := "string"
		if len(subargs) > 3 {
			valueType = subargs[3]
		}
		parsedValue := s.parseValue(value, valueType)
		return s.reshMgr.SetVariable(name, parsedValue, valueType, scope)

	case "get":
		if len(subargs) < 2 {
			return fmt.Errorf("usage: var get <scope> <name>")
		}
		scope := resh.VariableScope(subargs[0])
		name := subargs[1]
		variable, err := s.reshMgr.GetVariable(name, scope)
		if err != nil {
			return err
		}
		fmt.Printf("%s = %v (type: %s)\n", variable.Name, variable.Value, variable.Type)
		return nil

	case "delete", "del", "rm":
		if len(subargs) < 2 {
			return fmt.Errorf("usage: var delete <scope> <name>")
		}
		scope := resh.VariableScope(subargs[0])
		name := subargs[1]
		return s.reshMgr.DeleteVariable(name, scope)

	case "list", "ls":
		if len(subargs) < 1 {
			return fmt.Errorf("usage: var list <scope>")
		}
		scope := resh.VariableScope(subargs[0])
		variables, err := s.reshMgr.ListVariables(scope)
		if err != nil {
			return err
		}
		fmt.Printf("Variables in %s scope:\n", scope)
		for _, v := range variables {
			fmt.Printf("  %s = %v (type: %s)\n", v.Name, v.Value, v.Type)
		}
		return nil

	default:
		return fmt.Errorf("unknown var subcommand: %s", subcmd)
	}
}

// parseCommandLine splits a command line into tokens, respecting quotes
func (s *Shell) parseCommandLine(line string) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false

	for i := 0; i < len(line); i++ {
		ch := line[i]

		if ch == '"' {
			inQuotes = !inQuotes
		} else if ch == ' ' && !inQuotes {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		} else {
			current.WriteByte(ch)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// parseValue converts a string value to the appropriate type
func (s *Shell) parseValue(value string, valueType string) interface{} {
	switch valueType {
	case "bool":
		return value == "true" || value == "1"
	case "int":
		i, _ := strconv.Atoi(value)
		return i
	case "long":
		i, _ := strconv.ParseInt(value, 10, 64)
		return i
	case "float":
		f, _ := strconv.ParseFloat(value, 32)
		return float32(f)
	case "double":
		f, _ := strconv.ParseFloat(value, 64)
		return f
	default:
		return value
	}
}

// printSlotList prints a list of slots
func (s *Shell) printSlotList(slots []commands.SlotInfo) {
	if len(slots) == 0 {
		fmt.Println("(empty)")
		return
	}

	for _, slot := range slots {
		activeMarker := " "
		if slot.Active {
			activeMarker = "✓"
		}
		fmt.Printf("%s  %s (s-%s)\n", activeMarker, slot.Name, slot.RefID)
	}
}

// printHelp prints help text
func (s *Shell) printHelp() {
	fmt.Println("Resonite Shell (RESH) - Commands:")
	fmt.Println()
	fmt.Println("Navigation:")
	fmt.Println("  cd <path>        - Change directory (supports /, .., ~, RefIDs)")
	fmt.Println("  pwd              - Print working directory")
	fmt.Println("  ls [path]        - List slots in directory")
	fmt.Println("  tree [depth]     - Show hierarchy tree (default depth: 3)")
	fmt.Println()
	fmt.Println("Inspection:")
	fmt.Println("  cat <path>       - Show slot components")
	fmt.Println("  stat <path>      - Show detailed slot information")
	fmt.Println("  find <pattern>   - Search for slots by name")
	fmt.Println("  inspect <path> <idx> - Inspect component at index")
	fmt.Println()
	fmt.Println("Modification:")
	fmt.Println("  mkdir <name>     - Create new slot")
	fmt.Println("  touch <name>.<type> - Create slot with component")
	fmt.Println("  rm [-r] <path>   - Remove slot")
	fmt.Println("  edit <path> <property> <value> - Edit slot property")
	fmt.Println("  set <path> <idx> <field> <value> - Set component field")
	fmt.Println()
	fmt.Println("Variables:")
	fmt.Println("  var set <scope> <name> <value> - Set variable")
	fmt.Println("  var get <scope> <name>         - Get variable")
	fmt.Println("  var delete <scope> <name>      - Delete variable")
	fmt.Println("  var list <scope>               - List variables")
	fmt.Println("  Scopes: session, local, world")
	fmt.Println()
	fmt.Println("System:")
	fmt.Println("  help             - Show this help")
	fmt.Println("  exit, quit       - Exit shell")
	fmt.Println("  Tab              - Switch to TUI mode")
}
