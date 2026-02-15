package cmd

import (
	"fmt"
	"strings"

	"github.com/Merith-TK/resonite-sh/pkg/shell"
)

// ANSI color codes for terminal output
const (
	ColorReset  = "\033[0m"
	ColorCyan   = "\033[0;36m" // Active slots (like symlinks)
	ColorBlue   = "\033[0;34m" // Inactive slots (like directories)
	ColorWhite  = "\033[0;37m" // Components
	ColorGrey   = "\033[0;90m" // Greyed out (components)
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
	return fmt.Sprintf("%s[comp]%s", ColorGrey, ColorReset)
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

// displaySlotListing renders a complete slot listing
func displaySlotListing(listing *shell.SlotListing) {
	// Display header
	displaySlotHeader(listing.SlotName)

	// Display child slots
	for i := range listing.Children {
		displaySlotInfo(&listing.Children[i])
	}

	// Display components or protection message
	if listing.IsRootSlot {
		if len(listing.Children) > 0 {
			fmt.Println()
		}
		displayRootProtection()
	} else if len(listing.Components) > 0 {
		for i := range listing.Components {
			displayComponentInfo(&listing.Components[i])
		}
		fmt.Println()
	} else if len(listing.Children) == 0 {
		displayEmptySlot()
	} else {
		fmt.Println()
	}
}

// displaySlotInfo renders a SlotInfo
func displaySlotInfo(info *shell.SlotInfo) {
	persistIndicator := getPersistenceIndicator(info.IsPersistent)
	formattedType := formatSlotType(info.IsActive)
	fmt.Printf("%s %s %s\n", persistIndicator, formattedType, info.Name)
}

// displayComponentInfo renders a ComponentInfo
func displayComponentInfo(info *shell.ComponentInfo) {
	compType := info.Type
	if compType == "" {
		compType = "<unknown type>"
	}

	// Strip [FrooxEngine]FrooxEngine. or similar prefixes
	compType = stripComponentPrefix(compType)

	persistIndicator := getPersistenceIndicator(info.IsPersistent)
	formattedType := formatComponentType()
	fmt.Printf("%s %s %s\n", persistIndicator, formattedType, compType)
}

// stripComponentPrefix removes common component prefixes like [FrooxEngine]FrooxEngine.
func stripComponentPrefix(typeName string) string {
	// Strip [AssemblyName]Namespace. pattern
	// Example: [FrooxEngine]FrooxEngine.StaticFont -> StaticFont
	if idx := strings.LastIndex(typeName, "."); idx != -1 {
		return typeName[idx+1:]
	}
	return typeName
}
