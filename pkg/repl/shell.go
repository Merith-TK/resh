package repl

import (
	"fmt"
	"strings"
	"time"

	"github.com/Merith-TK/resonite-sh/pkg/resolink"
	"github.com/Merith-TK/resonite-sh/pkg/vfs"
	"github.com/chzyer/readline"
)

// Shell represents the REPL shell
type Shell struct {
	client   *resolink.Client
	vfs      *vfs.VFS
	readline *readline.Instance
	running  bool
}

// NewShell creates a new shell instance
func NewShell(url string) (*Shell, error) {
	// Create client
	client := resolink.NewClient(url, 30*time.Second)

	// Connect to Resonite
	if err := client.Connect(); err != nil {
		return nil, err
	}

	// Create VFS
	filesystem := vfs.NewVFS(client)
	if err := filesystem.Initialize(); err != nil {
		client.Disconnect()
		return nil, err
	}

	// Create readline instance
	rl, err := readline.New("resonite> ")
	if err != nil {
		client.Disconnect()
		return nil, err
	}

	return &Shell{
		client:   client,
		vfs:      filesystem,
		readline: rl,
		running:  false,
	}, nil
}

// Run starts the REPL loop
func (s *Shell) Run() error {
	s.running = true
	defer s.cleanup()

	fmt.Println("Welcome to resonite-sh!")
	fmt.Println("Type 'help' for available commands or 'exit' to quit.")
	fmt.Println()

	for s.running {
		// Update prompt with current path
		s.updatePrompt()

		// Read line
		line, err := s.readline.Readline()
		if err != nil {
			// EOF or error
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Execute command
		if err := s.executeCommand(line); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}

	return nil
}

// executeCommand parses and executes a command
func (s *Shell) executeCommand(line string) error {
	parts := strings.Fields(line)
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

	case "pwd":
		fmt.Println(s.vfs.GetCurrentPath())
		return nil

	case "cd":
		if len(args) == 0 {
			return fmt.Errorf("usage: cd <path>")
		}
		return s.vfs.ChangeDirectory(args[0])

	case "ls":
		path := s.vfs.GetCurrentPath()
		if len(args) > 0 {
			path = args[0]
		}
		return s.listDirectory(path)

	default:
		return fmt.Errorf("unknown command: %s (type 'help' for available commands)", cmd)
	}
}

// listDirectory lists directory contents
func (s *Shell) listDirectory(path string) error {
	nodes, err := s.vfs.ListDirectory(path)
	if err != nil {
		return err
	}

	if len(nodes) == 0 {
		fmt.Println("(empty)")
		return nil
	}

	for _, node := range nodes {
		fmt.Printf("%s/\n", node.Slot.Name)
	}

	return nil
}

// updatePrompt updates the readline prompt
func (s *Shell) updatePrompt() {
	prompt := fmt.Sprintf("resonite:%s> ", s.vfs.GetCurrentPath())
	s.readline.SetPrompt(prompt)
}

// printHelp displays available commands
func (s *Shell) printHelp() {
	help := `Available commands:

Navigation:
  cd <path>     Change directory
  pwd           Print working directory
  ls [path]     List directory contents

System:
  help          Show this help message
  exit          Exit the shell
  quit          Exit the shell

More commands coming soon!
`
	fmt.Println(help)
}

// cleanup performs cleanup on shell exit
func (s *Shell) cleanup() {
	fmt.Println("\nGoodbye!")
	s.readline.Close()
	s.client.Disconnect()
}
