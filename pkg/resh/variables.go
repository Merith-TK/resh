package resh

import (
	"encoding/json"
	"fmt"

	"github.com/Merith-TK/resh/pkg/resolink"
)

// VariableScope represents where a variable is stored
type VariableScope string

const (
	ScopeSession VariableScope = "session" // Temporary, cleared on disconnect
	ScopeLocal   VariableScope = "local"   // Persists in local DB
	ScopeWorld   VariableScope = "world"   // Saved with world
)

// Variable represents a RESH variable
type Variable struct {
	Name  string
	Value interface{}
	Type  string // e.g., "string", "int", "float", "float3", etc.
	Scope VariableScope
}

// SetVariable creates or updates a variable in the specified scope
// Variables are stored as DynamicReferenceVariable components with DynamicVariableSpace
func (m *Manager) SetVariable(name string, value interface{}, valueType string, scope VariableScope) error {
	// Determine parent slot based on scope
	var parentSlotID string
	switch scope {
	case ScopeSession:
		parentSlotID = m.sessionSlotID
	case ScopeLocal:
		parentSlotID = m.localSlotID
	case ScopeWorld:
		parentSlotID = m.worldSlotID
	default:
		return fmt.Errorf("invalid scope: %s", scope)
	}

	// Check if variable slot already exists
	varSlot, err := m.client.FindSlotByName(parentSlotID, name)
	if err == nil && varSlot != nil {
		// Variable exists, update it
		var slotData resolink.SlotDataResponse
		if err := json.Unmarshal(varSlot, &slotData); err != nil {
			return fmt.Errorf("failed to parse variable slot: %w", err)
		}
		return m.updateVariable(slotData.Data.ID, value, valueType)
	}

	// Variable doesn't exist, create it
	return m.createVariable(parentSlotID, name, value, valueType)
}

// createVariable creates a new variable slot with DynamicReferenceVariable component
func (m *Manager) createVariable(parentSlotID, name string, value interface{}, valueType string) error {
	// Create variable slot
	varSlotDef := &resolink.SlotDefinition{
		ID:       name,
		Name:     resolink.NewValueString(name),
		ParentID: parentSlotID,
		Active:   resolink.NewValueBool(true),
	}

	varResp, err := m.client.AddSlot(varSlotDef)
	if err != nil {
		return fmt.Errorf("failed to create variable slot: %w", err)
	}

	var varData resolink.SlotDataResponse
	if err := json.Unmarshal(varResp, &varData); err != nil {
		return fmt.Errorf("failed to parse variable response: %w", err)
	}
	varSlotID := varData.Data.ID

	// Add DynamicReferenceVariable component
	// This component type is generic and will hold our value
	dynRefVarDef := &resolink.ComponentDefinition{
		Type:   "FrooxEngine.DynamicReferenceVariable`1[[FrooxEngine.IWorldElement, FrooxEngine]]",
		SlotID: varSlotID,
		Fields: map[string]interface{}{
			"VariableName": resolink.NewValueString(name),
			"Value":        m.convertValueType(value, valueType),
		},
	}

	if _, err := m.client.AddComponent(varSlotID, dynRefVarDef); err != nil {
		return fmt.Errorf("failed to add DynamicReferenceVariable: %w", err)
	}

	return nil
}

// updateVariable updates an existing variable's value
func (m *Manager) updateVariable(varSlotID string, value interface{}, valueType string) error {
	// Get the variable slot with components
	slotData, err := m.client.GetSlot(varSlotID, true, 0)
	if err != nil {
		return fmt.Errorf("failed to get variable slot: %w", err)
	}

	var slotResp resolink.SlotDataResponse
	if err := json.Unmarshal(slotData, &slotResp); err != nil {
		return fmt.Errorf("failed to parse variable slot: %w", err)
	}

	// Find DynamicReferenceVariable component
	var dynRefVarID string
	for _, comp := range slotResp.Data.Components {
		if comp.Type == "FrooxEngine.DynamicReferenceVariable`1[[FrooxEngine.IWorldElement, FrooxEngine]]" {
			dynRefVarID = comp.ID
			break
		}
	}

	if dynRefVarID == "" {
		return fmt.Errorf("DynamicReferenceVariable component not found")
	}

	// Update the component's Value field
	updateDef := &resolink.ComponentDefinition{
		ID: dynRefVarID,
		Fields: map[string]interface{}{
			"Value": m.convertValueType(value, valueType),
		},
	}

	if _, err := m.client.UpdateComponent(updateDef); err != nil {
		return fmt.Errorf("failed to update variable value: %w", err)
	}

	return nil
}

// GetVariable retrieves a variable from the specified scope
func (m *Manager) GetVariable(name string, scope VariableScope) (*Variable, error) {
	// Determine parent slot based on scope
	var parentSlotID string
	switch scope {
	case ScopeSession:
		parentSlotID = m.sessionSlotID
	case ScopeLocal:
		parentSlotID = m.localSlotID
	case ScopeWorld:
		parentSlotID = m.worldSlotID
	default:
		return nil, fmt.Errorf("invalid scope: %s", scope)
	}

	// Find variable slot
	varSlot, err := m.client.FindSlotByName(parentSlotID, name)
	if err != nil {
		return nil, fmt.Errorf("variable not found: %w", err)
	}

	var slotData resolink.SlotDataResponse
	if err := json.Unmarshal(varSlot, &slotData); err != nil {
		return nil, fmt.Errorf("failed to parse variable slot: %w", err)
	}

	// Get the slot with components to read the value
	slotWithComps, err := m.client.GetSlot(slotData.Data.ID, true, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get variable components: %w", err)
	}

	var slotResp resolink.SlotDataResponse
	if err := json.Unmarshal(slotWithComps, &slotResp); err != nil {
		return nil, fmt.Errorf("failed to parse variable data: %w", err)
	}

	// Extract value from DynamicReferenceVariable component
	for _, comp := range slotResp.Data.Components {
		if comp.Type == "FrooxEngine.DynamicReferenceVariable`1[[FrooxEngine.IWorldElement, FrooxEngine]]" {
			// Parse the component fields to extract value
			valueField, ok := comp.Fields["Value"]
			if !ok {
				return nil, fmt.Errorf("variable has no value field")
			}

			// valueField should be a value type struct
			return &Variable{
				Name:  name,
				Value: valueField,
				Type:  m.detectValueType(valueField),
				Scope: scope,
			}, nil
		}
	}

	return nil, fmt.Errorf("DynamicReferenceVariable component not found")
}

// DeleteVariable removes a variable from the specified scope
func (m *Manager) DeleteVariable(name string, scope VariableScope) error {
	// Determine parent slot based on scope
	var parentSlotID string
	switch scope {
	case ScopeSession:
		parentSlotID = m.sessionSlotID
	case ScopeLocal:
		parentSlotID = m.localSlotID
	case ScopeWorld:
		parentSlotID = m.worldSlotID
	default:
		return fmt.Errorf("invalid scope: %s", scope)
	}

	// Find variable slot
	varSlot, err := m.client.FindSlotByName(parentSlotID, name)
	if err != nil {
		return fmt.Errorf("variable not found: %w", err)
	}

	var slotData resolink.SlotDataResponse
	if err := json.Unmarshal(varSlot, &slotData); err != nil {
		return fmt.Errorf("failed to parse variable slot: %w", err)
	}

	// Remove the variable slot (and all its components)
	if _, err := m.client.RemoveSlot(slotData.Data.ID); err != nil {
		return fmt.Errorf("failed to remove variable: %w", err)
	}

	return nil
}

// ListVariables returns all variables in the specified scope
func (m *Manager) ListVariables(scope VariableScope) ([]*Variable, error) {
	// Determine parent slot based on scope
	var parentSlotID string
	switch scope {
	case ScopeSession:
		parentSlotID = m.sessionSlotID
	case ScopeLocal:
		parentSlotID = m.localSlotID
	case ScopeWorld:
		parentSlotID = m.worldSlotID
	default:
		return nil, fmt.Errorf("invalid scope: %s", scope)
	}

	// Get all children of the scope slot
	children, err := m.client.ListChildren(parentSlotID)
	if err != nil {
		return nil, fmt.Errorf("failed to list variables: %w", err)
	}

	var variables []*Variable
	for _, child := range children {
		var slotData resolink.SlotDataResponse
		if err := json.Unmarshal(child, &slotData); err != nil {
			continue // Skip invalid slots
		}

		// Get variable details
		variable, err := m.GetVariable(slotData.Data.Name.Value.(string), scope)
		if err != nil {
			continue // Skip variables we can't read
		}

		variables = append(variables, variable)
	}

	return variables, nil
}

// convertValueType converts a Go value to a ResoniteLink value type
func (m *Manager) convertValueType(value interface{}, valueType string) interface{} {
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

// detectValueType determines the type of a value field
func (m *Manager) detectValueType(value interface{}) string {
	// This is a simplified type detection
	// In reality, we'd inspect the $type field of the value
	valueMap, ok := value.(map[string]interface{})
	if !ok {
		return "unknown"
	}

	typeField, ok := valueMap["$type"]
	if !ok {
		return "unknown"
	}

	return typeField.(string)
}
