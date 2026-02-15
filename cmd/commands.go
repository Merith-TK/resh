package cmd

import (
	"fmt"
	"strings"

	"github.com/Merith-TK/resonite-sh/pkg/resolink"
)

// listSlotContents displays the contents of a slot (children and components)
func listSlotContents(client *resolink.Client, slotID string, rootSlotID string) {
	// Get the current slot details (with components)
	slotResp, err := client.GetSlot(slotID, true, 0)
	if err != nil {
		fmt.Printf("Error getting slot: %v\n", err)
		return
	}

	slot := slotResp.Data

	// Display slot name header
	slotName := slotID
	if slot.Name != nil {
		slotName = slot.Name.Value
	}
	displaySlotHeader(slotName)

	// Check if this is the root slot (blacklist components)
	isRootSlot := (slotID == rootSlotID || slotID == "Root")

	// Display child slots first
	for _, child := range slot.Children {
		displaySlot(&child)
	}

	// Display components (or root protection message)
	if isRootSlot {
		if len(slot.Children) > 0 {
			fmt.Println()
		}
		displayRootProtection()
	} else if len(slot.Components) > 0 {
		for _, comp := range slot.Components {
			displayComponent(&comp)
		}
		fmt.Println()
	} else if len(slot.Children) == 0 {
		displayEmptySlot()
	} else {
		fmt.Println()
	}
}

// testConnection tests the connection by fetching the Root slot
func testConnection(client *resolink.Client) {
	rootSlot, err := client.GetSlot("Root", false, 0)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("✓ Root slot: %s\n", rootSlot.Data.ID)
	if rootSlot.Data.Name != nil {
		fmt.Printf("  Name: %s\n", rootSlot.Data.Name.Value)
	}
}

// printHelp displays the help message
func printHelp() {
	fmt.Println("Resonite Shell (RESH) - Commands:")
	fmt.Println()
	fmt.Println("Basic:")
	fmt.Println("  help             - Show this help")
	fmt.Println("  test             - Test connection (get Root slot)")
	fmt.Println("  exit, quit       - Exit shell")
	fmt.Println()
	fmt.Println("Navigation:")
	fmt.Println("  ls               - List slots and components")
	fmt.Println("  cd <slot>        - Navigate to child slot")
	fmt.Println("  cd ..            - Go to parent slot")
	fmt.Println("  cd /             - Go to root slot")
	fmt.Println()
	fmt.Println("More commands coming soon...")
}

// changeDirectory navigates to a different slot
func changeDirectory(client *resolink.Client, target string, state *replState) error {
	// Handle special cases
	switch target {
	case "/":
		// Go to root
		state.currentSlot = "Root"
		state.currentPath = "/Root"
		return nil

	case "..":
		// Go to parent
		return navigateToParent(client, state)

	default:
		// Navigate to child slot by name
		return navigateToChild(client, target, state)
	}
}

// navigateToParent moves to the parent slot
func navigateToParent(client *resolink.Client, state *replState) error {
	// Can't go above root
	if state.currentSlot == "Root" || state.currentSlot == state.rootSlotID {
		return fmt.Errorf("already at root")
	}

	// Get current slot to find parent
	slotResp, err := client.GetSlot(state.currentSlot, false, 0)
	if err != nil {
		return fmt.Errorf("failed to get current slot: %w", err)
	}

	// Check if parent exists
	if slotResp.Data.Parent == nil || slotResp.Data.Parent.TargetID == "" {
		return fmt.Errorf("no parent slot")
	}

	parentID := slotResp.Data.Parent.TargetID

	// Get parent slot to update path
	parentResp, err := client.GetSlot(parentID, false, 0)
	if err != nil {
		return fmt.Errorf("failed to get parent slot: %w", err)
	}

	// Update state
	state.currentSlot = parentID

	// Update path
	parentName := parentID
	if parentResp.Data.Name != nil {
		parentName = parentResp.Data.Name.Value
	}

	// Rebuild path by removing last component
	if lastSlash := strings.LastIndex(state.currentPath, "/"); lastSlash > 0 {
		state.currentPath = state.currentPath[:lastSlash]
	} else {
		state.currentPath = "/" + parentName
	}

	return nil
}

// navigateToChild moves to a child slot by name
func navigateToChild(client *resolink.Client, targetName string, state *replState) error {
	// Get current slot with children
	slotResp, err := client.GetSlot(state.currentSlot, false, 0)
	if err != nil {
		return fmt.Errorf("failed to get current slot: %w", err)
	}

	// Find child by name
	var targetSlot *resolink.SlotReference
	for _, child := range slotResp.Data.Children {
		if child.Name != nil && child.Name.Value == targetName {
			targetSlot = &child
			break
		}
	}

	if targetSlot == nil {
		return fmt.Errorf("slot not found: %s", targetName)
	}

	// Update state
	state.currentSlot = targetSlot.ID
	state.currentPath = state.currentPath + "/" + targetName

	return nil
}
