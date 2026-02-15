package main

import (
	"fmt"
	"log"
	"time"

	"github.com/Merith-TK/resonite-sh/pkg/resolink"
	"github.com/Merith-TK/resonite-sh/pkg/shell"
)

func main() {
	// Connect to ResoLink
	client := resolink.NewClient("ws://localhost:39015", 30*time.Second)
	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	fmt.Println("Connected to ResoLink")
	fmt.Println()

	// Initialize state
	state, err := shell.InitializeState(client)
	if err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}

	fmt.Printf("Root slot: %s\n", state.RootSlotID)
	fmt.Println()

	// Get Root slot info with components
	resp, err := client.GetSlot(state.RootSlotID, true, 0)
	if err != nil {
		log.Fatalf("Failed to get Root: %v", err)
	}

	if len(resp.Data.Components) == 0 {
		fmt.Println("No components found on Root")
		return
	}

	// Take first component
	componentRef := resp.Data.Components[0]
	fmt.Printf("Inspecting component: %s (%s)\n", componentRef.ID, componentRef.ComponentType)
	fmt.Println()

	// Inspect it
	data, err := shell.InspectComponent(client, componentRef.ID)
	if err != nil {
		log.Fatalf("Failed to inspect: %v", err)
	}

	// Display
	fmt.Printf("Component: %s (%s)\n", data.TypeName, data.ID)
	fmt.Println()
	fmt.Println("Members:")
	for _, member := range data.Members {
		fmt.Printf("  %s [%s] %s = %v\n", member.ID, member.Type, member.Name, member.Value)
	}
}
