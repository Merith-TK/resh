package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Merith-TK/resh/pkg/resolink"
)

// Modifier handles modification commands (mkdir, touch, rm, edit, set)
type Modifier struct {
	client    *resolink.Client
	navigator *Navigator
}

// NewModifier creates a new modifier
func NewModifier(client *resolink.Client, nav *Navigator) *Modifier {
	return &Modifier{
		client:    client,
		navigator: nav,
	}
}

// Mkdir creates a new slot (directory)
func (m *Modifier) Mkdir(name string, parentPath string) error {
	// Resolve parent path
	parentID := m.navigator.GetCurrentSlot()
	if parentPath != "" {
		refID, err := m.resolvePath(parentPath)
		if err != nil {
			return fmt.Errorf("invalid parent path: %w", err)
		}
		parentID = refID
	}

	// Create slot
	slotDef := &resolink.SlotDefinition{
		ID:       name,
		Name:     resolink.NewValueString(name),
		ParentID: parentID,
		Active:   resolink.NewValueBool(true),
	}

	if _, err := m.client.AddSlot(slotDef); err != nil {
		return fmt.Errorf("failed to create slot: %w", err)
	}

	return nil
}

// Touch creates an empty slot with a specific component (like creating a file with extension)
// Example: touch MyLight.PointLight creates a slot with PointLight component
func (m *Modifier) Touch(nameWithType string, parentPath string) error {
	parts := strings.Split(nameWithType, ".")
	if len(parts) < 2 {
		// No type specified, just create empty slot
		return m.Mkdir(nameWithType, parentPath)
	}

	name := strings.Join(parts[:len(parts)-1], ".")
	componentType := parts[len(parts)-1]

	// Resolve parent path
	parentID := m.navigator.GetCurrentSlot()
	if parentPath != "" {
		refID, err := m.resolvePath(parentPath)
		if err != nil {
			return fmt.Errorf("invalid parent path: %w", err)
		}
		parentID = refID
	}

	// Create slot
	slotDef := &resolink.SlotDefinition{
		ID:       name,
		Name:     resolink.NewValueString(name),
		ParentID: parentID,
		Active:   resolink.NewValueBool(true),
	}

	slotResp, err := m.client.AddSlot(slotDef)
	if err != nil {
		return fmt.Errorf("failed to create slot: %w", err)
	}

	var slotData resolink.SlotDataResponse
	if err := json.Unmarshal(slotResp, &slotData); err != nil {
		return fmt.Errorf("failed to parse slot: %w", err)
	}

	// Add component if type specified
	// Attempt to resolve short name to full type
	fullType, err := m.resolveComponentType(componentType)
	if err != nil {
		return fmt.Errorf("failed to resolve component type: %w", err)
	}

	compDef := &resolink.ComponentDefinition{
		Type:   fullType,
		SlotID: slotData.Data.ID,
		Fields: map[string]interface{}{},
	}

	if _, err := m.client.AddComponent(slotData.Data.ID, compDef); err != nil {
		return fmt.Errorf("failed to add component: %w", err)
	}

	return nil
}

// Rm removes a slot and all its children
func (m *Modifier) Rm(path string, recursive bool) error {
	// Resolve path
	refID, err := m.resolvePath(path)
	if err != nil {
		return fmt.Errorf("path not found: %w", err)
	}

	// Safety check: don't allow removing Root
	if refID == "Root" {
		return fmt.Errorf("cannot remove Root slot")
	}

	// Check if slot has children (if not recursive)
	if !recursive {
		children, err := m.client.ListChildren(refID)
		if err == nil && len(children) > 0 {
			return fmt.Errorf("slot has children, use -r flag to remove recursively")
		}
	}

	// Remove slot
	if _, err := m.client.RemoveSlot(refID); err != nil {
		return fmt.Errorf("failed to remove slot: %w", err)
	}

	return nil
}

// Edit modifies a slot's properties
func (m *Modifier) Edit(path string, property string, value interface{}, valueType string) error {
	// Resolve path
	refID, err := m.resolvePath(path)
	if err != nil {
		return fmt.Errorf("path not found: %w", err)
	}

	// Get current slot data
	slotData, err := m.client.GetSlot(refID, false, 0)
	if err != nil {
		return fmt.Errorf("failed to get slot: %w", err)
	}

	var slotResp resolink.SlotDataResponse
	if err := json.Unmarshal(slotData, &slotResp); err != nil {
		return fmt.Errorf("failed to parse slot: %w", err)
	}

	// Create update definition
	updateDef := &resolink.SlotDefinition{
		ID: refID,
	}

	// Set property based on name
	switch property {
	case "name":
		updateDef.Name = resolink.NewValueString(value.(string))
	case "active":
		updateDef.Active = resolink.NewValueBool(value.(bool))
	case "orderOffset":
		updateDef.OrderOffset = resolink.NewValueInt(value.(int))
	case "position":
		updateDef.Position = value.(map[string]interface{}) // Assume it's already a proper float3
	case "rotation":
		updateDef.Rotation = value.(map[string]interface{}) // Assume it's already a proper floatQ
	case "scale":
		updateDef.Scale = value.(map[string]interface{}) // Assume it's already a proper float3
	default:
		return fmt.Errorf("unknown property: %s", property)
	}

	// Update slot
	if _, err := m.client.UpdateSlot(updateDef); err != nil {
		return fmt.Errorf("failed to update slot: %w", err)
	}

	return nil
}

// Set modifies a component field value
func (m *Modifier) Set(slotPath string, componentIndex int, fieldName string, value interface{}, valueType string) error {
	// Resolve path
	refID, err := m.resolvePath(slotPath)
	if err != nil {
		return fmt.Errorf("path not found: %w", err)
	}

	// Get slot with components
	slotData, err := m.client.GetSlot(refID, true, 0)
	if err != nil {
		return fmt.Errorf("failed to get slot: %w", err)
	}

	var slotResp resolink.SlotDataResponse
	if err := json.Unmarshal(slotData, &slotResp); err != nil {
		return fmt.Errorf("failed to parse slot: %w", err)
	}

	// Check component index
	if componentIndex < 0 || componentIndex >= len(slotResp.Data.Components) {
		return fmt.Errorf("component index %d out of range (0-%d)", componentIndex, len(slotResp.Data.Components)-1)
	}

	comp := slotResp.Data.Components[componentIndex]

	// Create update definition
	updateDef := &resolink.ComponentDefinition{
		ID: comp.ID,
		Fields: map[string]interface{}{
			fieldName: m.convertValueType(value, valueType),
		},
	}

	// Update component
	if _, err := m.client.UpdateComponent(updateDef); err != nil {
		return fmt.Errorf("failed to update component: %w", err)
	}

	return nil
}

// AddComponent adds a component to a slot
func (m *Modifier) AddComponent(slotPath string, componentType string, fields map[string]interface{}) error {
	// Resolve path
	refID, err := m.resolvePath(slotPath)
	if err != nil {
		return fmt.Errorf("path not found: %w", err)
	}

	// Resolve component type
	fullType, err := m.resolveComponentType(componentType)
	if err != nil {
		return fmt.Errorf("failed to resolve component type: %w", err)
	}

	// Create component definition
	compDef := &resolink.ComponentDefinition{
		Type:   fullType,
		SlotID: refID,
		Fields: fields,
	}

	// Add component
	if _, err := m.client.AddComponent(refID, compDef); err != nil {
		return fmt.Errorf("failed to add component: %w", err)
	}

	return nil
}

// RemoveComponent removes a component from a slot
func (m *Modifier) RemoveComponent(slotPath string, componentIndex int) error {
	// Resolve path
	refID, err := m.resolvePath(slotPath)
	if err != nil {
		return fmt.Errorf("path not found: %w", err)
	}

	// Get slot with components
	slotData, err := m.client.GetSlot(refID, true, 0)
	if err != nil {
		return fmt.Errorf("failed to get slot: %w", err)
	}

	var slotResp resolink.SlotDataResponse
	if err := json.Unmarshal(slotData, &slotResp); err != nil {
		return fmt.Errorf("failed to parse slot: %w", err)
	}

	// Check component index
	if componentIndex < 0 || componentIndex >= len(slotResp.Data.Components) {
		return fmt.Errorf("component index %d out of range (0-%d)", componentIndex, len(slotResp.Data.Components)-1)
	}

	comp := slotResp.Data.Components[componentIndex]

	// Remove component
	if _, err := m.client.RemoveComponent(comp.ID); err != nil {
		return fmt.Errorf("failed to remove component: %w", err)
	}

	return nil
}

// resolvePath resolves a path string to a RefID
func (m *Modifier) resolvePath(path string) (string, error) {
	if path == "" || path == "." {
		return m.navigator.GetCurrentSlot(), nil
	}

	if strings.HasPrefix(path, "ID") {
		return path, nil
	}

	// Use navigator to resolve path
	oldSlot := m.navigator.GetCurrentSlot()
	oldPath := make([]string, len(m.navigator.GetCurrentPath()))
	copy(oldPath, m.navigator.GetCurrentPath())

	// Temporarily cd to resolve path
	if err := m.navigator.Cd(path); err != nil {
		return "", err
	}

	refID := m.navigator.GetCurrentSlot()

	// Restore original position
	m.navigator.currentSlot = oldSlot
	m.navigator.currentPath = oldPath

	return refID, nil
}

// resolveComponentType attempts to resolve a short component name to full type
// Example: "PointLight" -> "FrooxEngine.PointLight"
func (m *Modifier) resolveComponentType(shortName string) (string, error) {
	// If already fully qualified, return as-is
	if strings.Contains(shortName, ".") {
		return shortName, nil
	}

	// Try common prefixes
	prefixes := []string{
		"FrooxEngine.",
		"Elements.Core.",
		"FrooxEngine.UIX.",
	}

	// Get component type list to verify
	typeList, err := m.client.GetComponentTypeList()
	if err != nil {
		// Fallback to FrooxEngine prefix if can't get type list
		return "FrooxEngine." + shortName, nil
	}

	var types []string
	if err := json.Unmarshal(typeList, &types); err != nil {
		return "FrooxEngine." + shortName, nil
	}

	// Search for matching type
	for _, fullType := range types {
		typeParts := strings.Split(fullType, ".")
		typeName := typeParts[len(typeParts)-1]

		if typeName == shortName {
			return fullType, nil
		}
	}

	// Try prefixes
	for _, prefix := range prefixes {
		candidate := prefix + shortName
		for _, fullType := range types {
			if fullType == candidate {
				return fullType, nil
			}
		}
	}

	// Fallback to FrooxEngine prefix
	return "FrooxEngine." + shortName, nil
}

// convertValueType converts a Go value to a ResoniteLink value type
func (m *Modifier) convertValueType(value interface{}, valueType string) interface{} {
	switch valueType {
	case "string":
		return resolink.NewValueString(value.(string))
	case "bool":
		return resolink.NewValueBool(value.(bool))
	case "int":
		return resolink.NewValueInt(value.(int))
	case "long":
		return resolink.NewValueLong(value.(int64))
	case "float":
		return resolink.NewValueFloat(value.(float32))
	case "double":
		return resolink.NewValueDouble(value.(float64))
	case "float3":
		return value // Assume it's already a proper float3 struct
	case "floatQ":
		return value // Assume it's already a proper quaternion struct
	case "color":
		return value // Assume it's already a proper color struct
	default:
		// Default to string
		return resolink.NewValueString(fmt.Sprintf("%v", value))
	}
}
