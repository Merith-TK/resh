package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Merith-TK/resonite-sh/pkg/resolink"
)

// Navigator handles navigation commands (cd, pwd, ls, tree)
type Navigator struct {
	client      *resolink.Client
	currentPath []string // Stack of slot IDs representing current path
	currentSlot string   // Current slot RefID
}

// NewNavigator creates a new navigator starting at Root
func NewNavigator(client *resolink.Client) *Navigator {
	return &Navigator{
		client:      client,
		currentPath: []string{"Root"},
		currentSlot: "Root",
	}
}

// Cd changes directory to the specified slot
// Supports: absolute paths (/Root/Slot1), relative paths (Slot1, ../Slot2), RefIDs (ID2300)
func (n *Navigator) Cd(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Handle special cases
	if path == "~" || path == "/" {
		n.currentPath = []string{"Root"}
		n.currentSlot = "Root"
		return nil
	}

	if path == ".." {
		return n.cdUp()
	}

	if path == "." {
		return nil // Stay in current directory
	}

	// Check if it's a RefID (starts with "ID")
	if strings.HasPrefix(path, "ID") {
		return n.cdToRefID(path)
	}

	// Handle absolute path
	if strings.HasPrefix(path, "/") {
		return n.cdAbsolute(path)
	}

	// Handle relative path
	return n.cdRelative(path)
}

// cdUp moves up one level in the hierarchy
func (n *Navigator) cdUp() error {
	if len(n.currentPath) <= 1 {
		return fmt.Errorf("already at root")
	}

	// Remove last element from path
	n.currentPath = n.currentPath[:len(n.currentPath)-1]

	// Get the parent slot's RefID
	parentName := n.currentPath[len(n.currentPath)-1]

	// Build path from root to find the slot
	var slotID string
	var err error

	if parentName == "Root" {
		slotID = "Root"
	} else {
		// Navigate from Root to find this slot
		slotID, err = n.resolvePathToRefID(n.currentPath)
		if err != nil {
			return fmt.Errorf("failed to resolve parent: %w", err)
		}
	}

	n.currentSlot = slotID
	return nil
}

// cdToRefID changes to a slot by its RefID
func (n *Navigator) cdToRefID(refID string) error {
	// Verify the slot exists
	slotData, err := n.client.GetSlot(refID, false, 0)
	if err != nil {
		return fmt.Errorf("slot %s not found: %w", refID, err)
	}

	var slotResp resolink.SlotDataResponse
	if err := json.Unmarshal(slotData, &slotResp); err != nil {
		return fmt.Errorf("failed to parse slot: %w", err)
	}

	// Build path from root to this slot
	path, err := n.buildPathToSlot(refID)
	if err != nil {
		return fmt.Errorf("failed to build path: %w", err)
	}

	n.currentPath = path
	n.currentSlot = refID
	return nil
}

// cdAbsolute changes to an absolute path from root
func (n *Navigator) cdAbsolute(path string) error {
	// Remove leading /
	path = strings.TrimPrefix(path, "/")

	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] != "Root" {
		// Prepend Root if not specified
		parts = append([]string{"Root"}, parts...)
	}

	// Resolve path to RefID
	refID, err := n.resolvePathToRefID(parts)
	if err != nil {
		return fmt.Errorf("path not found: %w", err)
	}

	n.currentPath = parts
	n.currentSlot = refID
	return nil
}

// cdRelative changes to a relative path from current slot
func (n *Navigator) cdRelative(path string) error {
	parts := strings.Split(path, "/")

	// Start from current path
	newPath := make([]string, len(n.currentPath))
	copy(newPath, n.currentPath)

	// Process each part
	for _, part := range parts {
		if part == "" || part == "." {
			continue
		}
		if part == ".." {
			if len(newPath) > 1 {
				newPath = newPath[:len(newPath)-1]
			}
		} else {
			newPath = append(newPath, part)
		}
	}

	// Resolve final path to RefID
	refID, err := n.resolvePathToRefID(newPath)
	if err != nil {
		return fmt.Errorf("path not found: %w", err)
	}

	n.currentPath = newPath
	n.currentSlot = refID
	return nil
}

// resolvePathToRefID navigates from Root through each path element to find the final RefID
func (n *Navigator) resolvePathToRefID(path []string) (string, error) {
	if len(path) == 0 {
		return "", fmt.Errorf("empty path")
	}

	if path[0] != "Root" {
		return "", fmt.Errorf("path must start with Root")
	}

	currentID := "Root"

	// Navigate through each path segment
	for i := 1; i < len(path); i++ {
		childName := path[i]

		// Find child by name
		childData, err := n.client.FindSlotByName(currentID, childName)
		if err != nil {
			return "", fmt.Errorf("slot '%s' not found in %s", childName, strings.Join(path[:i], "/"))
		}

		var slotResp resolink.SlotDataResponse
		if err := json.Unmarshal(childData, &slotResp); err != nil {
			return "", fmt.Errorf("failed to parse slot: %w", err)
		}

		currentID = slotResp.Data.ID
	}

	return currentID, nil
}

// buildPathToSlot builds the path array from Root to the given slot
func (n *Navigator) buildPathToSlot(refID string) ([]string, error) {
	if refID == "Root" {
		return []string{"Root"}, nil
	}

	// Get slot data
	slotData, err := n.client.GetSlot(refID, false, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get slot: %w", err)
	}

	var slotResp resolink.SlotDataResponse
	if err := json.Unmarshal(slotData, &slotResp); err != nil {
		return nil, fmt.Errorf("failed to parse slot: %w", err)
	}

	// Recursively build path from parent
	if slotResp.Data.ParentID == "" {
		return []string{slotResp.Data.Name.Value.(string)}, nil
	}

	parentPath, err := n.buildPathToSlot(slotResp.Data.ParentID)
	if err != nil {
		return nil, err
	}

	return append(parentPath, slotResp.Data.Name.Value.(string)), nil
}

// Pwd returns the current working directory path
func (n *Navigator) Pwd() string {
	return "/" + strings.Join(n.currentPath, "/")
}

// Ls lists the children of the current slot or specified path
func (n *Navigator) Ls(path string) ([]SlotInfo, error) {
	targetSlot := n.currentSlot

	// If path specified, resolve it
	if path != "" {
		if strings.HasPrefix(path, "ID") {
			targetSlot = path
		} else if strings.HasPrefix(path, "/") {
			parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
			if parts[0] != "Root" {
				parts = append([]string{"Root"}, parts...)
			}
			refID, err := n.resolvePathToRefID(parts)
			if err != nil {
				return nil, fmt.Errorf("path not found: %w", err)
			}
			targetSlot = refID
		} else {
			// Relative path
			parts := strings.Split(path, "/")
			newPath := make([]string, len(n.currentPath))
			copy(newPath, n.currentPath)
			newPath = append(newPath, parts...)

			refID, err := n.resolvePathToRefID(newPath)
			if err != nil {
				return nil, fmt.Errorf("path not found: %w", err)
			}
			targetSlot = refID
		}
	}

	// Get children
	children, err := n.client.ListChildren(targetSlot)
	if err != nil {
		return nil, fmt.Errorf("failed to list children: %w", err)
	}

	var slots []SlotInfo
	for _, child := range children {
		var slotResp resolink.SlotDataResponse
		if err := json.Unmarshal(child, &slotResp); err != nil {
			continue
		}

		slots = append(slots, SlotInfo{
			Name:        slotResp.Data.Name.Value.(string),
			RefID:       slotResp.Data.ID,
			Active:      slotResp.Data.Active.Value.(bool),
			OrderOffset: int(slotResp.Data.OrderOffset.Value.(float64)),
		})
	}

	return slots, nil
}

// Tree displays a tree view of the hierarchy starting from current slot
func (n *Navigator) Tree(depth int) (string, error) {
	if depth <= 0 {
		depth = 3 // Default depth
	}

	var builder strings.Builder
	if err := n.buildTree(&builder, n.currentSlot, "", depth, 0); err != nil {
		return "", err
	}

	return builder.String(), nil
}

// buildTree recursively builds a tree representation
func (n *Navigator) buildTree(builder *strings.Builder, slotID string, prefix string, maxDepth int, currentDepth int) error {
	if currentDepth >= maxDepth {
		return nil
	}

	// Get slot info
	slotData, err := n.client.GetSlot(slotID, false, 0)
	if err != nil {
		return fmt.Errorf("failed to get slot: %w", err)
	}

	var slotResp resolink.SlotDataResponse
	if err := json.Unmarshal(slotData, &slotResp); err != nil {
		return fmt.Errorf("failed to parse slot: %w", err)
	}

	// Print current slot
	builder.WriteString(fmt.Sprintf("%s%s (s-%s)\n", prefix, slotResp.Data.Name.Value, slotID))

	// Get children
	children, err := n.client.ListChildren(slotID)
	if err != nil {
		return nil // Skip if can't list children
	}

	// Process children
	for i, child := range children {
		var childResp resolink.SlotDataResponse
		if err := json.Unmarshal(child, &childResp); err != nil {
			continue
		}

		// Determine prefix for child
		isLast := i == len(children)-1
		var childPrefix string
		if isLast {
			childPrefix = prefix + "└── "
		} else {
			childPrefix = prefix + "├── "
		}

		// Recursively process child
		var nextPrefix string
		if isLast {
			nextPrefix = prefix + "    "
		} else {
			nextPrefix = prefix + "│   "
		}

		builder.WriteString(childPrefix)
		n.buildTree(builder, childResp.Data.ID, nextPrefix, maxDepth, currentDepth+1)
	}

	return nil
}

// GetCurrentSlot returns the current slot RefID
func (n *Navigator) GetCurrentSlot() string {
	return n.currentSlot
}

// GetCurrentPath returns the current path
func (n *Navigator) GetCurrentPath() []string {
	return n.currentPath
}

// SlotInfo represents basic slot information for listing
type SlotInfo struct {
	Name        string
	RefID       string
	Active      bool
	OrderOffset int
}
