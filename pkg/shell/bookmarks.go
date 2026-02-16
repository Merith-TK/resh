package shell

import (
	"fmt"

	"github.com/Merith-TK/resonite-sh/pkg/resolink"
)

// SetBookmark saves a slot ID with a bookmark name
func SetBookmark(client *resolink.Client, state *State, name string, slotID string) error {
	if state.RESHDataID == "" {
		return fmt.Errorf("RESH.DATA not initialized - variable storage unavailable")
	}

	// Store in state
	state.Variables[name] = slotID

	// TODO: Persist to a DynamicValueVariable component in RESH.DATA slot
	// For now, just keep in memory

	return nil
}

// GetBookmark retrieves a slot ID by bookmark name
func GetBookmark(state *State, name string) (string, error) {
	if state.RESHDataID == "" {
		return "", fmt.Errorf("RESH.DATA not initialized - variable storage unavailable")
	}

	slotID, exists := state.Variables[name]
	if !exists {
		return "", fmt.Errorf("bookmark '%s' not found", name)
	}

	return slotID, nil
}

// ListBookmarks returns all bookmark names
func ListBookmarks(state *State) []string {
	bookmarks := make([]string, 0, len(state.Variables))
	for name := range state.Variables {
		bookmarks = append(bookmarks, name)
	}
	return bookmarks
}

// DeleteBookmark removes a bookmark
func DeleteBookmark(state *State, name string) error {
	if state.RESHDataID == "" {
		return fmt.Errorf("RESH.DATA not initialized - variable storage unavailable")
	}

	if _, exists := state.Variables[name]; !exists {
		return fmt.Errorf("bookmark '%s' not found", name)
	}

	delete(state.Variables, name)
	return nil
}
