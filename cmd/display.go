package cmd

import (
	"fmt"

	"github.com/Merith-TK/resonite-sh/pkg/resolink"
)

// ANSI color codes for terminal output
const (
	ColorReset  = "\033[0m"
	ColorCyan   = "\033[0;36m" // Active slots (like symlinks)
	ColorBlue   = "\033[0;34m" // Inactive slots (like directories)
	ColorWhite  = "\033[0;37m" // Components
	ColorYellow = "\033[0;33m" // Warnings/info
)

// Display utilities for REPL output

// formatSlotType returns a colored [slot] indicator based on active state
func formatSlotType(isActive bool) string {
	color := ColorCyan
	if !isActive {
		color = ColorBlue
	}
	return fmt.Sprintf("%s[slot]%s", color, ColorReset)
}

// formatComponentType returns a colored [comp] indicator
func formatComponentType() string {
	return fmt.Sprintf("%s[comp]%s", ColorWhite, ColorReset)
}

// getPersistenceIndicator returns a text indicator for persistence
func getPersistenceIndicator(isPersistent bool) string {
	if isPersistent {
		return "P" // Persistent
	}
	return "T" // Temporary
}

// displaySlotHeader prints the slot name header
func displaySlotHeader(slotName string) {
	fmt.Println()
	fmt.Printf("Contents of slot: %s\n", slotName)
	fmt.Println()
}

// displaySlot prints a single slot entry with unified format
func displaySlot(slot *resolink.SlotReference) {
	// Get slot name
	name := "<unnamed>"
	if slot.Name != nil {
		name = slot.Name.Value
	}

	// Determine if active
	isActive := true
	if slot.IsActive != nil {
		isActive = slot.IsActive.Value
	}

	// Determine persistence
	isPersistent := true
	if slot.IsPersistent != nil {
		isPersistent = slot.IsPersistent.Value
	}

	// Format and print with unified format: {P/T} [slot] name
	persistIndicator := getPersistenceIndicator(isPersistent)
	formattedType := formatSlotType(isActive)
	fmt.Printf("%s %s %s\n", persistIndicator, formattedType, name)
}

// displayComponent prints a single component entry with unified format
func displayComponent(comp *resolink.ComponentReference) {
	compType := comp.ComponentType
	if compType == "" {
		compType = "<unknown type>"
	}

	formattedType := formatComponentType()
	fmt.Printf("  %s %s\n", formattedType, compType)
}

// displayEmptySlot prints a message for empty slots
func displayEmptySlot() {
	fmt.Println("(empty slot - no children or components)")
	fmt.Println()
}

// displayRootProtection prints the root slot protection message
func displayRootProtection() {
	fmt.Printf("%s[Root slot - components hidden for safety]%s\n", ColorYellow, ColorReset)
	fmt.Println()
}
