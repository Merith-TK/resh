package cmd

import (
	"fmt"
	"strings"

	"github.com/Merith-TK/resonite-sh/pkg/resolink"
	"github.com/Merith-TK/resonite-sh/pkg/shell"
)

// listSlotContents displays the contents of a slot
func listSlotContents(client *resolink.Client, state *shell.State) {
	listing, err := shell.ListSlot(client, state.CurrentSlot, state.RootSlotID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	displaySlotListing(listing)
}

// testConnection tests the connection by fetching the Root slot
func testConnection(client *resolink.Client) {
	info, err := shell.TestConnection(client)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("✓ Root slot: %s\n", info.SlotID)
	if info.SlotName != "" {
		fmt.Printf("  Name: %s\n", info.SlotName)
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
	fmt.Println("Components:")
	fmt.Println("  inspect <id>              - Inspect component or slot")
	fmt.Println("  set <id> <key>=<val>      - Set component member or slot property")
	fmt.Println()
	fmt.Println("More commands coming soon...")
}

// changeDirectory navigates to a different slot
func changeDirectory(client *resolink.Client, target string, state *shell.State) error {
	// Handle special cases
	switch target {
	case "/":
		shell.NavigateToRoot(state)
		return nil

	case "..":
		return shell.NavigateToParent(client, state)

	default:
		return shell.NavigateToChild(client, state, target)
	}
}

// inspectComponent displays detailed component member information
func inspectComponent(client *resolink.Client, componentID string) {
	// Try component first
	compData, compErr := shell.InspectComponent(client, componentID)
	if compErr == nil {
		displayComponentData(compData)
		return
	}

	// If component failed, try slot
	slotData, slotErr := shell.InspectSlot(client, componentID)
	if slotErr == nil {
		displaySlotData(slotData)
		return
	}

	// Both failed
	fmt.Printf("Error: Not a valid component or slot ID\n")
	fmt.Printf("  Component error: %v\n", compErr)
	fmt.Printf("  Slot error: %v\n", slotErr)
}

// setComponentMember updates a component member value or slot property
// Expects format: componentID memberID=value OR slotID propertyName=value
func setComponentMember(client *resolink.Client, args []string) {
	if len(args) < 2 {
		fmt.Println("set: missing arguments")
		fmt.Println("Usage: set <id> <member_id>=<value>  (for components)")
		fmt.Println("       set <id> <property>=<value>   (for slots)")
		fmt.Println("Example: set ID_100 ID_812=2")
		fmt.Println("Example: set ID_101 Name=\"My Slot\"")
		return
	}

	targetID := args[0]
	assignment := args[1]

	// Parse member_id=value or property=value
	parts := strings.SplitN(assignment, "=", 2)
	if len(parts) != 2 {
		fmt.Println("set: invalid assignment format")
		fmt.Println("Usage: set <id> <key>=<value>")
		return
	}

	key := parts[0]
	value := parts[1]

	// Try to determine if target is a component or slot by attempting component first
	// Components are more common in typical workflows
	compErr := shell.SetComponentMember(client, targetID, key, value)
	if compErr == nil {
		fmt.Println("✓ Member updated")
		return
	}

	// If component failed, try slot property
	slotErr := shell.SetSlotProperty(client, targetID, key, value)
	if slotErr == nil {
		fmt.Println("✓ Property updated")
		return
	}

	// Both failed - show both errors
	fmt.Printf("Error: Could not set as component member or slot property\n")
	fmt.Printf("  Component: %v\n", compErr)
	fmt.Printf("  Slot: %v\n", slotErr)
}
