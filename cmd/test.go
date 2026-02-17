package cmd

import (
	"fmt"
	"time"

	"github.com/Merith-TK/resh/pkg/resolink"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// testCmd tests connection to ResoLink
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test connection to ResoLink WebSocket",
	Long:  `Connects to the ResoLink WebSocket and verifies the connection works.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get WebSocket URL from config or flag
		url := viper.GetString("connection.url")
		if url == "" {
			url = "ws://localhost:29551"
		}

		fmt.Printf("Connecting to ResoLink at %s...\n", url)

		// Create client
		timeout := 30 * time.Second
		client := resolink.NewClient(url, timeout)

		// Connect
		if err := client.Connect(); err != nil {
			return fmt.Errorf("failed to connect: %w", err)
		}
		defer client.Disconnect()

		fmt.Println("✓ Connected successfully!")

		// Test basic operation - get Root slot
		// Try "Root" first, then fallback to ID2300 if not found
		fmt.Println("\nTesting GetSlot(Root)...")
		rootSlot, err := client.GetSlot("Root", false, 0)
		if err != nil {
			fmt.Printf("  Trying with 'Root' failed: %v\n", err)
			fmt.Println("  Trying with 'ID2300'...")
			rootSlot, err = client.GetSlot("ID2300", false, 0)
			if err != nil {
				return fmt.Errorf("failed to get Root slot: %w", err)
			}
		}

		fmt.Printf("✓ Got Root slot: %s\n", rootSlot.Data.ID)
		if rootSlot.Data.Name != nil {
			fmt.Printf("  Name: %s\n", rootSlot.Data.Name.Value)
		}
		if rootSlot.Data.IsActive != nil {
			fmt.Printf("  IsActive: %v\n", rootSlot.Data.IsActive.Value)
		}

		// Test ListChildren
		fmt.Println("\nTesting ListChildren(Root)...")
		children, err := client.ListChildren("Root")
		if err != nil {
			return fmt.Errorf("failed to list children: %w", err)
		}

		fmt.Printf("✓ Found %d children under Root:\n", len(children))
		for i, child := range children {
			if i >= 5 {
				fmt.Printf("  ... and %d more\n", len(children)-5)
				break
			}
			childName := "(unnamed)"
			if child.Name != nil {
				childName = child.Name.Value
			}
			fmt.Printf("  - %s (ID: %s)\n", childName, child.ID)
		}

		fmt.Println("\n✓ All tests passed!")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
