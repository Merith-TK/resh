package shell

import (
	"fmt"
	"strings"

	"github.com/Merith-TK/resonite-sh/pkg/resolink"
)

// ListSlot retrieves the contents of a slot
func ListSlot(client *resolink.Client, slotID string, rootSlotID string) (*SlotListing, error) {
	// Get the slot details with components
	slotResp, err := client.GetSlot(slotID, true, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get slot: %w", err)
	}

	slot := slotResp.Data

	// Build the listing
	listing := &SlotListing{
		SlotID:     slotID,
		SlotName:   slotID,
		IsRootSlot: (slotID == rootSlotID || slotID == "Root"),
	}

	if slot.Name != nil {
		listing.SlotName = slot.Name.Value
	}

	// Convert child slots
	for _, child := range slot.Children {
		listing.Children = append(listing.Children, SlotInfoFromResoLink(&child))
	}

	// Convert components
	for _, comp := range slot.Components {
		listing.Components = append(listing.Components, ComponentInfoFromResoLink(&comp))
	}

	return listing, nil
}

// NavigateToChild navigates to a child slot by name
func NavigateToChild(client *resolink.Client, state *State, targetName string) error {
	// Get current slot with children
	slotResp, err := client.GetSlot(state.CurrentSlot, false, 0)
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
	state.CurrentSlot = targetSlot.ID
	state.CurrentPath = state.CurrentPath + "/" + targetName

	return nil
}

// NavigateToParent navigates to the parent slot
func NavigateToParent(client *resolink.Client, state *State) error {
	// Can't go above root
	if state.CurrentSlot == "Root" || state.CurrentSlot == state.RootSlotID {
		return fmt.Errorf("already at root")
	}

	// Get current slot to find parent
	slotResp, err := client.GetSlot(state.CurrentSlot, false, 0)
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
	state.CurrentSlot = parentID

	// Update path
	parentName := parentID
	if parentResp.Data.Name != nil {
		parentName = parentResp.Data.Name.Value
	}

	// Rebuild path by removing last component
	if lastSlash := strings.LastIndex(state.CurrentPath, "/"); lastSlash > 0 {
		state.CurrentPath = state.CurrentPath[:lastSlash]
	} else {
		state.CurrentPath = "/" + parentName
	}

	return nil
}

// NavigateToRoot navigates to the root slot
func NavigateToRoot(state *State) {
	state.CurrentSlot = "Root"
	state.CurrentPath = "/Root"
}

// TestConnection tests the connection by fetching the Root slot
func TestConnection(client *resolink.Client) (*ConnectionInfo, error) {
	rootSlot, err := client.GetSlot("Root", false, 0)
	if err != nil {
		return nil, err
	}

	info := &ConnectionInfo{
		SlotID: rootSlot.Data.ID,
	}

	if rootSlot.Data.Name != nil {
		info.SlotName = rootSlot.Data.Name.Value
	}

	return info, nil
}

// InitializeState creates the initial state by fetching the root slot
func InitializeState(client *resolink.Client) (*State, error) {
	rootSlotResp, err := client.GetSlot("Root", false, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get root slot: %w", err)
	}

	return NewStateFromResoLink(rootSlotResp), nil
}
