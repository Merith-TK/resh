package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Merith-TK/resonite-sh/pkg/resolink"
)

// Inspector handles inspection commands (cat, stat, find, inspect)
type Inspector struct {
	client    *resolink.Client
	navigator *Navigator
}

// NewInspector creates a new inspector
func NewInspector(client *resolink.Client, nav *Navigator) *Inspector {
	return &Inspector{
		client:    client,
		navigator: nav,
	}
}

// Cat displays the components of a slot (analogous to viewing file contents)
func (i *Inspector) Cat(path string) (string, error) {
	// Resolve path to RefID
	refID, err := i.resolvePath(path)
	if err != nil {
		return "", err
	}

	// Get slot with components
	slotData, err := i.client.GetSlot(refID, true, 0)
	if err != nil {
		return "", fmt.Errorf("failed to get slot: %w", err)
	}

	var slotResp resolink.SlotDataResponse
	if err := json.Unmarshal(slotData, &slotResp); err != nil {
		return "", fmt.Errorf("failed to parse slot: %w", err)
	}

	// Build output
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Slot: %s (s-%s)\n", slotResp.Data.Name.Value, refID))
	builder.WriteString(fmt.Sprintf("Active: %v\n", slotResp.Data.Active.Value))
	builder.WriteString(fmt.Sprintf("OrderOffset: %v\n", slotResp.Data.OrderOffset.Value))
	builder.WriteString(fmt.Sprintf("\nComponents (%d):\n", len(slotResp.Data.Components)))

	for idx, comp := range slotResp.Data.Components {
		builder.WriteString(fmt.Sprintf("\n[%d] %s (c-%s)\n", idx, comp.Type, comp.ID))

		// Show component fields
		if len(comp.Fields) > 0 {
			builder.WriteString("  Fields:\n")
			for fieldName, fieldValue := range comp.Fields {
				builder.WriteString(fmt.Sprintf("    %s: %v\n", fieldName, i.formatValue(fieldValue)))
			}
		}
	}

	return builder.String(), nil
}

// Stat displays detailed information about a slot (similar to stat command)
func (i *Inspector) Stat(path string) (string, error) {
	// Resolve path to RefID
	refID, err := i.resolvePath(path)
	if err != nil {
		return "", err
	}

	// Get slot with components
	slotData, err := i.client.GetSlot(refID, true, 0)
	if err != nil {
		return "", fmt.Errorf("failed to get slot: %w", err)
	}

	var slotResp resolink.SlotDataResponse
	if err := json.Unmarshal(slotData, &slotResp); err != nil {
		return "", fmt.Errorf("failed to parse slot: %w", err)
	}

	// Get children count
	children, err := i.client.ListChildren(refID)
	childCount := 0
	if err == nil {
		childCount = len(children)
	}

	// Build output
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("  Slot: %s\n", slotResp.Data.Name.Value))
	builder.WriteString(fmt.Sprintf("  RefID: %s\n", refID))
	builder.WriteString(fmt.Sprintf("  Active: %v\n", slotResp.Data.Active.Value))
	builder.WriteString(fmt.Sprintf("  OrderOffset: %v\n", slotResp.Data.OrderOffset.Value))
	builder.WriteString(fmt.Sprintf("  Parent: %s\n", slotResp.Data.ParentID))
	builder.WriteString(fmt.Sprintf("  Children: %d\n", childCount))
	builder.WriteString(fmt.Sprintf("  Components: %d\n", len(slotResp.Data.Components)))

	if len(slotResp.Data.Components) > 0 {
		builder.WriteString("\n  Component Types:\n")
		for _, comp := range slotResp.Data.Components {
			// Show short type name
			typeParts := strings.Split(comp.Type, ".")
			shortType := typeParts[len(typeParts)-1]
			builder.WriteString(fmt.Sprintf("    - %s (c-%s)\n", shortType, comp.ID))
		}
	}

	return builder.String(), nil
}

// Find searches for slots by name pattern
func (i *Inspector) Find(pattern string, searchRoot string) ([]FindResult, error) {
	// Determine search root
	rootSlot := i.navigator.GetCurrentSlot()
	if searchRoot != "" {
		refID, err := i.resolvePath(searchRoot)
		if err != nil {
			return nil, fmt.Errorf("invalid search root: %w", err)
		}
		rootSlot = refID
	}

	// Perform recursive search
	var results []FindResult
	if err := i.findRecursive(rootSlot, pattern, "", &results, 0, 10); err != nil {
		return nil, err
	}

	return results, nil
}

// findRecursive performs recursive slot search
func (i *Inspector) findRecursive(slotID string, pattern string, currentPath string, results *[]FindResult, depth int, maxDepth int) error {
	if depth >= maxDepth {
		return nil
	}

	// Get slot info
	slotData, err := i.client.GetSlot(slotID, false, 0)
	if err != nil {
		return nil // Skip slots we can't access
	}

	var slotResp resolink.SlotDataResponse
	if err := json.Unmarshal(slotData, &slotResp); err != nil {
		return nil
	}

	slotName := slotResp.Data.Name.Value.(string)
	fullPath := currentPath + "/" + slotName

	// Check if name matches pattern (simple contains match)
	if strings.Contains(strings.ToLower(slotName), strings.ToLower(pattern)) {
		*results = append(*results, FindResult{
			Name:  slotName,
			RefID: slotID,
			Path:  fullPath,
		})
	}

	// Search children
	children, err := i.client.ListChildren(slotID)
	if err != nil {
		return nil // Skip if can't list children
	}

	for _, child := range children {
		var childResp resolink.SlotDataResponse
		if err := json.Unmarshal(child, &childResp); err != nil {
			continue
		}

		i.findRecursive(childResp.Data.ID, pattern, fullPath, results, depth+1, maxDepth)
	}

	return nil
}

// Inspect shows detailed component information for a specific component
func (i *Inspector) Inspect(slotPath string, componentIndex int) (string, error) {
	// Resolve path to RefID
	refID, err := i.resolvePath(slotPath)
	if err != nil {
		return "", err
	}

	// Get slot with components
	slotData, err := i.client.GetSlot(refID, true, 0)
	if err != nil {
		return "", fmt.Errorf("failed to get slot: %w", err)
	}

	var slotResp resolink.SlotDataResponse
	if err := json.Unmarshal(slotData, &slotResp); err != nil {
		return "", fmt.Errorf("failed to parse slot: %w", err)
	}

	// Check component index
	if componentIndex < 0 || componentIndex >= len(slotResp.Data.Components) {
		return "", fmt.Errorf("component index %d out of range (0-%d)", componentIndex, len(slotResp.Data.Components)-1)
	}

	comp := slotResp.Data.Components[componentIndex]

	// Build detailed output
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Component: %s\n", comp.Type))
	builder.WriteString(fmt.Sprintf("RefID: %s\n", comp.ID))
	builder.WriteString(fmt.Sprintf("Slot: %s (s-%s)\n\n", slotResp.Data.Name.Value, refID))
	builder.WriteString("Fields:\n")

	if len(comp.Fields) == 0 {
		builder.WriteString("  (no fields)\n")
	} else {
		for fieldName, fieldValue := range comp.Fields {
			builder.WriteString(fmt.Sprintf("  %s:\n", fieldName))
			builder.WriteString(fmt.Sprintf("    %s\n", i.formatValueDetailed(fieldValue)))
		}
	}

	return builder.String(), nil
}

// resolvePath resolves a path string to a RefID
func (i *Inspector) resolvePath(path string) (string, error) {
	if path == "" || path == "." {
		return i.navigator.GetCurrentSlot(), nil
	}

	if strings.HasPrefix(path, "ID") {
		return path, nil
	}

	// Use navigator to resolve path
	oldSlot := i.navigator.GetCurrentSlot()
	oldPath := make([]string, len(i.navigator.GetCurrentPath()))
	copy(oldPath, i.navigator.GetCurrentPath())

	// Temporarily cd to resolve path
	if err := i.navigator.Cd(path); err != nil {
		return "", err
	}

	refID := i.navigator.GetCurrentSlot()

	// Restore original position
	i.navigator.currentSlot = oldSlot
	i.navigator.currentPath = oldPath

	return refID, nil
}

// formatValue formats a field value for display
func (i *Inspector) formatValue(value interface{}) string {
	valueMap, ok := value.(map[string]interface{})
	if !ok {
		return fmt.Sprintf("%v", value)
	}

	// Check for $type field
	typeField, hasType := valueMap["$type"]
	if !hasType {
		return fmt.Sprintf("%v", value)
	}

	switch typeField {
	case "string":
		return fmt.Sprintf("\"%v\"", valueMap["value"])
	case "bool", "int", "long", "float", "double":
		return fmt.Sprintf("%v", valueMap["value"])
	case "float3":
		v := valueMap["value"].(map[string]interface{})
		return fmt.Sprintf("(%v, %v, %v)", v["x"], v["y"], v["z"])
	case "floatQ":
		v := valueMap["value"].(map[string]interface{})
		return fmt.Sprintf("(%v, %v, %v, %v)", v["x"], v["y"], v["z"], v["w"])
	case "color":
		v := valueMap["value"].(map[string]interface{})
		return fmt.Sprintf("rgba(%v, %v, %v, %v)", v["r"], v["g"], v["b"], v["a"])
	case "reference":
		return fmt.Sprintf("-> %v", valueMap["targetId"])
	default:
		return fmt.Sprintf("%v", value)
	}
}

// formatValueDetailed formats a field value with type information
func (i *Inspector) formatValueDetailed(value interface{}) string {
	valueMap, ok := value.(map[string]interface{})
	if !ok {
		return fmt.Sprintf("Value: %v", value)
	}

	// Check for $type field
	typeField, hasType := valueMap["$type"]
	if !hasType {
		return fmt.Sprintf("Value: %v", value)
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Type: %s\n    ", typeField))
	builder.WriteString(fmt.Sprintf("Value: %s", i.formatValue(value)))

	return builder.String()
}

// FindResult represents a search result
type FindResult struct {
	Name  string
	RefID string
	Path  string
}
