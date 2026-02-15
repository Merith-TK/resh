package cmd

import (
	"fmt"

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
