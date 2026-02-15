package cmd

import (
	"strings"

	"github.com/Merith-TK/resonite-sh/pkg/resolink"
	"github.com/Merith-TK/resonite-sh/pkg/shell"
)

// completer implements readline.AutoCompleter for command and slot name completion
type completer struct {
	client *resolink.Client
	state  *shell.State
}

// newCompleter creates a new autocompleter
func newCompleter(client *resolink.Client, state *shell.State) *completer {
	return &completer{
		client: client,
		state:  state,
	}
}

// Do implements readline.AutoCompleter interface
func (c *completer) Do(line []rune, pos int) (newLine [][]rune, length int) {
	lineStr := string(line[:pos])

	// Handle edge case: if line starts with special char, try to find where the actual command starts
	if len(lineStr) > 0 && (lineStr[0] == '<' || lineStr[0] == '>') {
		// User is likely trying to complete a slot name that starts with <
		// We need to find the command part first
		// For now, just return no completions if we can't parse properly
		parts := parseCommandLine(lineStr)
		if len(parts) == 0 {
			return nil, 0
		}
	}

	parts := parseCommandLine(lineStr)

	// No input yet - show available commands
	if len(parts) == 0 {
		return c.completeCommand("")
	}

	// First word - complete command
	if len(parts) == 1 && !strings.HasSuffix(lineStr, " ") {
		return c.completeCommand(parts[0])
	}

	// After command - complete based on command type
	if len(parts) >= 1 {
		cmd := parts[0]
		switch cmd {
		case "cd":
			// Complete slot names
			var prefix string
			if len(parts) > 1 {
				prefix = parts[len(parts)-1]
			}
			return c.completeSlotName(prefix)
		}
	}

	return nil, 0
}

// completeCommand completes command names
func (c *completer) completeCommand(prefix string) ([][]rune, int) {
	commands := []string{"help", "test", "ls", "cd", "exit", "quit"}

	var matches [][]rune
	for _, cmd := range commands {
		if strings.HasPrefix(cmd, prefix) {
			matches = append(matches, []rune(cmd[len(prefix):]))
		}
	}

	return matches, len(prefix)
}

// completeSlotName completes slot names in current directory
func (c *completer) completeSlotName(prefix string) ([][]rune, int) {
	// Get current slot with children
	slotResp, err := c.client.GetSlot(c.state.CurrentSlot, false, 0)
	if err != nil {
		return nil, 0
	}

	// Special completions
	specialNames := []string{"..", "/"}
	var matches [][]rune

	// Add special names
	for _, name := range specialNames {
		if strings.HasPrefix(name, prefix) {
			matches = append(matches, []rune(name[len(prefix):]))
		}
	}

	// Add child slot names
	for _, child := range slotResp.Data.Children {
		if child.Name != nil {
			name := child.Name.Value

			// If name contains spaces and prefix doesn't start with a quote,
			// we need to suggest the quoted version
			if strings.Contains(name, " ") {
				quotedName := `"` + name + `"`
				// Check if prefix matches either quoted or unquoted version
				if strings.HasPrefix(quotedName, prefix) {
					matches = append(matches, []rune(quotedName[len(prefix):]))
				} else if strings.HasPrefix(name, prefix) && prefix != "" && prefix[0] != '"' {
					// User typed part of name without quote - suggest with quote
					matches = append(matches, []rune(`"`+name[len(prefix):]+`"`))
				}
			} else {
				// No spaces - simple match
				if strings.HasPrefix(name, prefix) {
					matches = append(matches, []rune(name[len(prefix):]))
				}
			}
		}
	}

	return matches, len(prefix)
}
