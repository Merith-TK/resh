package cmd

import (
	"strings"
	"unicode"
)

// parseCommandLine parses a command line into arguments, respecting quotes
func parseCommandLine(line string) []string {
	var args []string
	var current strings.Builder
	inQuotes := false
	quoteChar := rune(0)

	for i, r := range line {
		switch {
		case r == '"' || r == '\'':
			if !inQuotes {
				// Start quote
				inQuotes = true
				quoteChar = r
			} else if r == quoteChar {
				// End quote
				inQuotes = false
				quoteChar = 0
			} else {
				// Different quote inside quotes
				current.WriteRune(r)
			}

		case unicode.IsSpace(r) && !inQuotes:
			// Space outside quotes - end of argument
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}

		case r == '\\' && i+1 < len(line):
			// Escape character - take next character literally
			i++
			if i < len(line) {
				current.WriteRune(rune(line[i]))
			}

		default:
			// Regular character
			current.WriteRune(r)
		}
	}

	// Add final argument if any
	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}
