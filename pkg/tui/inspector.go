package tui

import (
	"fmt"
	"strings"

	"github.com/Merith-TK/resh/pkg/logger"
	"github.com/Merith-TK/resh/pkg/resolink"
	"github.com/Merith-TK/resh/pkg/shell"
)

// RenderInspector renders the inspector panel
func (m Model) RenderInspector(width, height int) string {
	var b strings.Builder

	// Title
	title := titleStyle.Render("Inspector")
	b.WriteString(title)
	b.WriteString("\n\n")

	if m.InspectedItem == nil {
		b.WriteString(helpStyle.Render("  No item selected\n"))
		b.WriteString(helpStyle.Render("  Press Enter on a tree item to inspect"))
	} else {
		// Render based on item type
		if m.InspectedItem.IsSlot {
			m.renderSlotInspector(&b)
		} else {
			m.renderComponentInspector(&b)
		}
	}

	// Help text at bottom
	b.WriteString("\n")
	if m.Focus == FocusInspector {
		if m.EditingField {
			b.WriteString(helpStyle.Render("  Type to edit | Enter:save Esc:cancel"))
		} else {
			// Show field type hint
			fieldType := m.getCurrentFieldType()
			if fieldType == "bool" {
				b.WriteString(helpStyle.Render("  ↑/↓:nav Enter:toggle Esc/Tab:back"))
			} else {
				b.WriteString(helpStyle.Render("  ↑/↓:nav Enter:edit Esc/Tab:back"))
			}
		}
	}

	content := b.String()
	return ApplyBorder(content, m.Focus == FocusInspector, width-2, height-2)
}

// renderSlotInspector renders slot metadata
func (m Model) renderSlotInspector(b *strings.Builder) {
	b.WriteString(fieldNameStyle.Render("  Type: "))
	b.WriteString("Slot\n\n")

	b.WriteString(fieldNameStyle.Render("  ID: "))
	b.WriteString(fieldValueStyle.Render(m.InspectedItem.ID))
	b.WriteString("\n\n")

	// If we have inspected data, show it
	if m.InspectedData != nil {
		if slotData, ok := m.InspectedData.(*shell.SlotData); ok {
			// Display all properties with selection highlighting
			// For composite types (float3, floatQ), expand into subfields
			fieldIndex := 0
			for _, prop := range slotData.Properties {
				// Skip read-only fields
				if prop.Name == "Parent" || prop.Name == "ID" {
					continue
				}

				// Expand composite types into subfields
				if prop.Type == "float3" || prop.Type == "floatQ" {
					m.renderCompositeField(b, prop, &fieldIndex)
				} else {
					isSelected := (m.Focus == FocusInspector && fieldIndex == m.FieldCursor)
					isEditing := (isSelected && m.EditingField)

					fieldLine := fmt.Sprintf("  %s: ", prop.Name)

					var valueStr string
					if isEditing {
						valueStr = m.EditBuffer
					} else {
						valueStr = formatPropertyValue(prop)
					}

					if isEditing {
						b.WriteString(fieldNameStyle.Render(fieldLine))
						b.WriteString(editingFieldStyle.Render(valueStr + "_"))
					} else if isSelected {
						b.WriteString(fieldNameStyle.Render(fieldLine))
						b.WriteString(selectedFieldStyle.Render(valueStr))
					} else {
						b.WriteString(fieldNameStyle.Render(fieldLine))
						b.WriteString(fieldValueStyle.Render(valueStr))
					}

					b.WriteString("\n")
					fieldIndex++
				}
			}
		}
	}
}

// isCompositeType checks if a type needs subfield expansion
func isCompositeType(typeName string) bool {
	return typeName == "float2" || typeName == "float3" || typeName == "float4" || typeName == "floatQ"
}

// getCompositeSubfieldCount returns number of subfields for a composite type
func getCompositeSubfieldCount(typeName string) int {
	switch typeName {
	case "float2":
		return 2
	case "float3":
		return 3
	case "float4", "floatQ":
		return 4
	default:
		return 0
	}
}

// getEditableFieldCount returns number of editable fields for current item
// For composite types (float2/3/4, floatQ), counts each component as a separate field
func (m Model) getEditableFieldCount() int {
	if m.InspectedItem == nil || m.InspectedData == nil {
		return 0
	}

	if m.InspectedItem.IsSlot {
		if slotData, ok := m.InspectedData.(*shell.SlotData); ok {
			count := 0
			for _, prop := range slotData.Properties {
				// Skip read-only fields
				if prop.Name == "Parent" || prop.Name == "ID" {
					continue
				}
				// Composite types have multiple subfield components
				if prop.Type == "float3" || prop.Type == "floatQ" {
					count += 3 // x, y, z (w is readonly for floatQ)
				} else {
					count++
				}
			}
			return count
		}
	} else {
		if compData, ok := m.InspectedData.(*shell.ComponentData); ok {
			count := 0
			for _, member := range compData.Members {
				if isCompositeType(member.Type) {
					// For component composites, show all components including w
					count += getCompositeSubfieldCount(member.Type)
				} else {
					count++
				}
			}
			return count
		}
	}
	return 0
}

// formatPropertyValue formats a property value for display
func formatPropertyValue(prop shell.SlotProperty) string {
	switch prop.Type {
	case "bool":
		if b, ok := prop.Value.(bool); ok {
			return fmt.Sprintf("%t", b)
		}
	case "float3":
		if v, ok := prop.Value.(*resolink.Float3); ok {
			return fmt.Sprintf("%.8f, %.8f, %.8f", v.X, v.Y, v.Z)
		}
	case "floatQ":
		if v, ok := prop.Value.(*resolink.FloatQ); ok {
			return fmt.Sprintf("%.8f, %.8f, %.8f, %.8f", v.X, v.Y, v.Z, v.W)
		}
	case "reference":
		if ref, ok := prop.Value.(string); ok {
			if ref == "" {
				return "(null)"
			}
			return ref
		}
	}
	return fmt.Sprintf("%v", prop.Value)
}

// renderCompositeField renders a composite field (float3, floatQ) as separate subfields
func (m Model) renderCompositeField(b *strings.Builder, prop shell.SlotProperty, fieldIndex *int) {
	var components []string
	var values []float64

	if prop.Type == "float3" {
		if v, ok := prop.Value.(*resolink.Float3); ok {
			components = []string{"x", "y", "z"}
			values = []float64{v.X, v.Y, v.Z}
		}
	} else if prop.Type == "floatQ" {
		if v, ok := prop.Value.(*resolink.FloatQ); ok {
			// For floatQ, only x, y, z are editable (w is readonly per user request)
			components = []string{"x", "y", "z"}
			values = []float64{v.X, v.Y, v.Z}
		}
	}

	for i, comp := range components {
		isSelected := (m.Focus == FocusInspector && *fieldIndex == m.FieldCursor)
		isEditing := (isSelected && m.EditingField)

		fieldLine := fmt.Sprintf("  %s.%s: ", prop.Name, comp)

		var valueStr string
		if isEditing {
			valueStr = m.EditBuffer
		} else {
			valueStr = fmt.Sprintf("%.8f", values[i])
		}

		if isEditing {
			b.WriteString(fieldNameStyle.Render(fieldLine))
			b.WriteString(editingFieldStyle.Render(valueStr + "_"))
		} else if isSelected {
			b.WriteString(fieldNameStyle.Render(fieldLine))
			b.WriteString(selectedFieldStyle.Render(valueStr))
		} else {
			b.WriteString(fieldNameStyle.Render(fieldLine))
			b.WriteString(fieldValueStyle.Render(valueStr))
		}

		b.WriteString("\n")
		*fieldIndex++
	}
}

// renderComponentCompositeField renders a composite component member (float2/3/4, floatQ) as separate субfields
func (m Model) renderComponentCompositeField(b *strings.Builder, member shell.MemberData, fieldIndex *int) {
	var components []string
	var values []float64

	// Extract values from map[string]interface{}
	if valueMap, ok := member.Value.(map[string]interface{}); ok {
		switch member.Type {
		case "float2":
			if x, xOk := valueMap["x"].(float64); xOk {
				if y, yOk := valueMap["y"].(float64); yOk {
					components = []string{"x", "y"}
					values = []float64{x, y}
				}
			}
		case "float3":
			if x, xOk := valueMap["x"].(float64); xOk {
				if y, yOk := valueMap["y"].(float64); yOk {
					if z, zOk := valueMap["z"].(float64); zOk {
						components = []string{"x", "y", "z"}
						values = []float64{x, y, z}
					}
				}
			}
		case "float4", "floatQ":
			if x, xOk := valueMap["x"].(float64); xOk {
				if y, yOk := valueMap["y"].(float64); yOk {
					if z, zOk := valueMap["z"].(float64); zOk {
						if w, wOk := valueMap["w"].(float64); wOk {
							components = []string{"x", "y", "z", "w"}
							values = []float64{x, y, z, w}
						}
					}
				}
			}
		}
	}

	// Render each subfield
	for i, comp := range components {
		isSelected := (m.Focus == FocusInspector && *fieldIndex == m.FieldCursor)
		isEditing := (isSelected && m.EditingField)

		// Truncate member name if too long
		memberName := member.Name
		if len(memberName) > 20 {
			memberName = memberName[:17] + "..."
		}

		fieldLine := fmt.Sprintf("    %s.%s: ", memberName, comp)

		var valueStr string
		if isEditing {
			valueStr = m.EditBuffer
		} else {
			valueStr = fmt.Sprintf("%.8f", values[i])
		}

		if isEditing {
			b.WriteString(fieldLine)
			b.WriteString(editingFieldStyle.Render(valueStr + "_"))
		} else if isSelected {
			b.WriteString(fieldLine)
			b.WriteString(selectedFieldStyle.Render(valueStr))
		} else {
			b.WriteString(fieldLine)
			b.WriteString(fieldValueStyle.Render(valueStr))
		}

		b.WriteString("\n")
		*fieldIndex++
	}
}

// renderComponentInspector renders component metadata
func (m Model) renderComponentInspector(b *strings.Builder) {
	b.WriteString(fieldNameStyle.Render("  Type: "))
	b.WriteString("Component\n\n")

	b.WriteString(fieldNameStyle.Render("  ID: "))
	b.WriteString(fieldValueStyle.Render(m.InspectedItem.ID))
	b.WriteString("\n")

	b.WriteString(fieldNameStyle.Render("  TypeName: "))
	typeName := m.InspectedItem.Type
	if len(typeName) > 40 {
		typeName = typeName[:37] + "..."
	}
	b.WriteString(fieldValueStyle.Render(typeName))
	b.WriteString("\n\n")

	// If we have inspected data, show component members
	if m.InspectedData != nil {
		if compData, ok := m.InspectedData.(*shell.ComponentData); ok {
			if len(compData.Members) == 0 {
				b.WriteString(helpStyle.Render("  (No members)"))
			} else {
				b.WriteString(fieldNameStyle.Render("  Members:"))
				b.WriteString("\n")

				// Show members with selection highlighting
				// For composite types, expand into subfields
				fieldIndex := 0
				maxMembersDisplay := 20
				displayedFields := 0

				for i := 0; i < len(compData.Members) && i < maxMembersDisplay; i++ {
					member := compData.Members[i]

					// Expand composite types into subfields
					if isCompositeType(member.Type) {
						m.renderComponentCompositeField(b, member, &fieldIndex)
						displayedFields += getCompositeSubfieldCount(member.Type)
					} else {
						isSelected := (m.Focus == FocusInspector && fieldIndex == m.FieldCursor)
						isEditing := (isSelected && m.EditingField)

						// Truncate member name if too long
						memberName := member.Name
						if len(memberName) > 25 {
							memberName = memberName[:22] + "..."
						}

						fieldLine := fmt.Sprintf("    %s: ", memberName)

						// Format value
						var valueStr string
						if isEditing {
							valueStr = m.EditBuffer
						} else {
							valueStr = fmt.Sprintf("%v", member.Value)
							if len(valueStr) > 40 {
								valueStr = valueStr[:37] + "..."
							}
						}

						// Apply styling based on state
						if isEditing {
							b.WriteString(fieldLine)
							b.WriteString(editingFieldStyle.Render(valueStr + "_")) // Show cursor
						} else if isSelected {
							b.WriteString(fieldLine)
							b.WriteString(selectedFieldStyle.Render(valueStr))
						} else {
							b.WriteString(fieldLine)
							b.WriteString(fieldValueStyle.Render(valueStr))
						}
						b.WriteString("\n")
						fieldIndex++
						displayedFields++
					}
				}

				if len(compData.Members) > maxMembersDisplay {
					remaining := len(compData.Members) - maxMembersDisplay
					b.WriteString(helpStyle.Render(fmt.Sprintf("    ... and %d more members", remaining)))
					b.WriteString("\n")
				}
			}
		} else {
			b.WriteString(errorStyle.Render("  Failed to load component data"))
		}
	} else {
		b.WriteString(helpStyle.Render("  Loading..."))
	}
}

// MoveFieldCursorUp moves the inspector field cursor up
func (m *Model) MoveFieldCursorUp() {
	if m.FieldCursor > 0 {
		m.FieldCursor--
	}
}

// MoveFieldCursorDown moves the inspector field cursor down
func (m *Model) MoveFieldCursorDown() {
	maxFields := m.getEditableFieldCount()
	if m.FieldCursor < maxFields-1 {
		m.FieldCursor++
	}
}

// StartEditingField starts editing the currently selected field
func (m *Model) StartEditingField() {
	// Get current field type and value
	fieldType := m.getCurrentFieldType()
	fieldValue := m.getCurrentFieldValue()

	if fieldType == "" {
		m.ErrorMessage = "Cannot edit this field"
		return
	}

	m.EditFieldType = fieldType
	m.StatusMessage = fmt.Sprintf("Editing %s (type: %s, value: %s)", m.EditFieldName, fieldType, fieldValue)

	// For bools, toggle immediately instead of entering edit mode
	if fieldType == "bool" {
		m.ErrorMessage = "" // Clear any previous errors
		m.toggleBoolField()
		return
	}

	// For other types, enter edit mode
	currentValue := m.getCurrentFieldValue()
	if currentValue == "" {
		m.ErrorMessage = "Cannot edit this field"
		return
	}

	m.EditingField = true
	m.EditBuffer = currentValue
	m.ErrorMessage = ""
}

// StopEditingField stops editing and saves/cancels
func (m *Model) StopEditingField(save bool) {
	if !save {
		// Just cancel
		m.EditingField = false
		m.EditBuffer = ""
		return
	}

	// Save the edited value
	if err := m.saveFieldValue(m.EditBuffer); err != nil {
		m.ErrorMessage = fmt.Sprintf("Save failed: %v", err)
		m.EditingField = false
		m.EditBuffer = ""
		return
	}

	m.EditingField = false
	m.EditBuffer = ""
	m.StatusMessage = "Field saved"

	// Reload the item to show updated values
	if m.InspectedItem != nil {
		data, isSlot, err := shell.InspectItem(m.Client, m.InspectedItem.ID)
		if err == nil && isSlot == m.InspectedItem.IsSlot {
			m.InspectedData = data
		}
	}
}

// getCurrentFieldValue gets the value of the currently selected field
func (m *Model) getCurrentFieldValue() string {
	if m.InspectedItem == nil || m.InspectedData == nil {
		m.EditFieldName = ""
		m.EditSubField = ""
		return ""
	}

	if m.InspectedItem.IsSlot {
		if slotData, ok := m.InspectedData.(*shell.SlotData); ok {
			fieldIndex := 0
			for _, prop := range slotData.Properties {
				// Skip read-only fields
				if prop.Name == "Parent" || prop.Name == "ID" {
					continue
				}

				// Handle composite types with subfields
				if prop.Type == "float3" || prop.Type == "floatQ" {
					// Check if cursor is within this composite's 3 subfields (x, y, z)
					if m.FieldCursor >= fieldIndex && m.FieldCursor < fieldIndex+3 {
						subIndex := m.FieldCursor - fieldIndex
						if prop.Type == "float3" {
							if v, ok := prop.Value.(*resolink.Float3); ok {
								vals := []float64{v.X, v.Y, v.Z}
								m.EditFieldName = prop.Name
								m.EditSubField = []string{"x", "y", "z"}[subIndex]
								return fmt.Sprintf("%.8f", vals[subIndex])
							}
						} else if prop.Type == "floatQ" {
							if v, ok := prop.Value.(*resolink.FloatQ); ok {
								vals := []float64{v.X, v.Y, v.Z}
								m.EditFieldName = prop.Name
								m.EditSubField = []string{"x", "y", "z"}[subIndex]
								return fmt.Sprintf("%.8f", vals[subIndex])
							}
						}
					}
					fieldIndex += 3
				} else {
					if fieldIndex == m.FieldCursor {
						m.EditFieldName = prop.Name
						m.EditSubField = ""
						val := formatPropertyValue(prop)
						return val
					}
					fieldIndex++
				}
			}
		}
	} else {
		if compData, ok := m.InspectedData.(*shell.ComponentData); ok {
			fieldIndex := 0
			for _, member := range compData.Members {
				// Handle composite types with subfields
				if isCompositeType(member.Type) {
					subfieldCount := getCompositeSubfieldCount(member.Type)
					// Check if cursor is within this composite's subfields
					if m.FieldCursor >= fieldIndex && m.FieldCursor < fieldIndex+subfieldCount {
						subIndex := m.FieldCursor - fieldIndex
						if valueMap, ok := member.Value.(map[string]interface{}); ok {
							components := []string{"x", "y", "z", "w"}
							if subIndex < len(components) {
								m.EditFieldName = member.Name
								m.EditSubField = components[subIndex]
								if val, ok := valueMap[components[subIndex]].(float64); ok {
									return fmt.Sprintf("%.8f", val)
								}
							}
						}
					}
					fieldIndex += subfieldCount
				} else {
					if fieldIndex == m.FieldCursor {
						m.EditFieldName = member.Name
						m.EditSubField = ""
						return fmt.Sprintf("%v", member.Value)
					}
					fieldIndex++
				}
			}
		}
	}

	// No field found at cursor position
	m.EditFieldName = ""
	m.EditSubField = ""
	return ""
}

// getCurrentFieldType gets the type of the currently selected field
func (m *Model) getCurrentFieldType() string {
	if m.InspectedItem == nil || m.InspectedData == nil {
		return ""
	}

	if m.InspectedItem.IsSlot {
		if slotData, ok := m.InspectedData.(*shell.SlotData); ok {
			fieldIndex := 0
			for _, prop := range slotData.Properties {
				// Skip read-only fields
				if prop.Name == "Parent" || prop.Name == "ID" {
					continue
				}

				// Handle composite types - each component is a float
				if prop.Type == "float3" || prop.Type == "floatQ" {
					if m.FieldCursor >= fieldIndex && m.FieldCursor < fieldIndex+3 {
						return "float"
					}
					fieldIndex += 3
				} else {
					if fieldIndex == m.FieldCursor {
						return prop.Type
					}
					fieldIndex++
				}
			}
		}
	} else {
		if compData, ok := m.InspectedData.(*shell.ComponentData); ok {
			fieldIndex := 0
			for _, member := range compData.Members {
				// Handle composite types - each subcomponent is a float or double
				if isCompositeType(member.Type) {
					subfieldCount := getCompositeSubfieldCount(member.Type)
					if m.FieldCursor >= fieldIndex && m.FieldCursor < fieldIndex+subfieldCount {
						// Return the base type (float or double)
						if strings.HasPrefix(member.Type, "double") {
							return "double"
						}
						return "float"
					}
					fieldIndex += subfieldCount
				} else {
					if fieldIndex == m.FieldCursor {
						return member.Type
					}
					fieldIndex++
				}
			}
		}
	}

	return ""
}

// toggleBoolField toggles a boolean field value
func (m *Model) toggleBoolField() {
	// First ensure EditFieldName is set by calling getCurrentFieldValue
	currentValue := m.getCurrentFieldValue()

	if currentValue == "" {
		m.ErrorMessage = "Could not read bool field"
		return
	}

	var newValue string
	if currentValue == "true" {
		newValue = "false"
	} else if currentValue == "false" {
		newValue = "true"
	} else {
		m.ErrorMessage = fmt.Sprintf("Invalid bool value: %s", currentValue)
		return
	}

	// Log the attempted toggle
	logger.Debug("toggleBoolField: field=%s, current=%s, new=%s", m.EditFieldName, currentValue, newValue)
	if m.InspectedItem != nil {
		logger.Debug("toggleBoolField: InspectedItem ID=%s, Type=%s, IsSlot=%v", m.InspectedItem.ID, m.InspectedItem.Type, m.InspectedItem.IsSlot)
	}

	m.StatusMessage = fmt.Sprintf("Attempting toggle of %s to %s...", m.EditFieldName, newValue)

	if err := m.saveFieldValue(newValue); err != nil {
		logger.Error("toggleBoolField: saveFieldValue FAILED: %v", err)
		m.ErrorMessage = fmt.Sprintf("Bool toggle FAILED: %v", err)
		m.StatusMessage = ""
		return
	}

	logger.Debug("toggleBoolField: saveFieldValue succeeded, reloading...")

	m.ErrorMessage = "" // Clear any errors
	m.StatusMessage = fmt.Sprintf("Toggled %s to %s", m.EditFieldName, newValue)

	// Reload the item to show updated values
	if m.InspectedItem != nil {
		data, isSlot, err := shell.InspectItem(m.Client, m.InspectedItem.ID)
		if err != nil {
			logger.Error("toggleBoolField: InspectItem FAILED: %v", err)
			m.ErrorMessage = fmt.Sprintf("Reload FAILED after toggle: %v", err)
			return
		}
		logger.Debug("toggleBoolField: InspectItem succeeded, isSlot=%v, expected=%v", isSlot, m.InspectedItem.IsSlot)
		if isSlot == m.InspectedItem.IsSlot {
			m.InspectedData = data
			// Log the reloaded value
			if isSlot {
				slotData := data.(*shell.SlotData)
				for _, prop := range slotData.Properties {
					if prop.Name == m.EditFieldName {
						logger.Debug("toggleBoolField: Reloaded %s = %v (type: %T)", m.EditFieldName, prop.Value, prop.Value)
						break
					}
				}
			}
		} else {
			logger.Error("toggleBoolField: Type mismatch: got isSlot=%v, expected=%v", isSlot, m.InspectedItem.IsSlot)
			m.ErrorMessage = "Reload returned wrong type"
		}
	}
}

// saveFieldValue saves the edited field value back to Resonite
func (m *Model) saveFieldValue(newValue string) error {
	if m.InspectedItem == nil || m.InspectedData == nil {
		return fmt.Errorf("no item selected")
	}

	if m.InspectedItem.IsSlot {
		return m.saveSlotField(newValue)
	}

	return m.saveComponentField(newValue)
}

// saveSlotField saves a slot property
func (m *Model) saveSlotField(newValue string) error {
	slotData, ok := m.InspectedData.(*shell.SlotData)
	if !ok {
		return fmt.Errorf("invalid slot data")
	}

	// If editing a subfield (e.g., Position.x), handle it specially
	if m.EditSubField != "" {
		return m.saveSlotSubfield(slotData, newValue)
	}

	logger.Debug("saveSlotField: Using shared shell.SetSlotProperty for %s=%s", m.EditFieldName, newValue)

	// Use the shared function from shell package
	if err := shell.SetSlotProperty(m.Client, slotData.ID, m.EditFieldName, newValue); err != nil {
		logger.Error("saveSlotField: SetSlotProperty FAILED: %v", err)
		return fmt.Errorf("SetSlotProperty call failed: %w", err)
	}

	logger.Debug("saveSlotField: SetSlotProperty succeeded")
	return nil
}

// saveSlotSubfield handles editing individual components of composite types (e.g., Position.x)
func (m *Model) saveSlotSubfield(slotData *shell.SlotData, newValue string) error {
	// Find the property being edited
	var targetProp *shell.SlotProperty
	for i := range slotData.Properties {
		if slotData.Properties[i].Name == m.EditFieldName {
			targetProp = &slotData.Properties[i]
			break
		}
	}

	if targetProp == nil {
		return fmt.Errorf("field '%s' not found in slot", m.EditFieldName)
	}

	// Parse the new value as a float
	var floatVal float64
	if _, err := fmt.Sscanf(newValue, "%f", &floatVal); err != nil {
		return fmt.Errorf("invalid number: %s", newValue)
	}

	logger.Debug("saveSlotSubfield: Updating %s.%s to %.8f", m.EditFieldName, m.EditSubField, floatVal)

	// Handle composite types
	if targetProp.Type == "float3" {
		currentVal, ok := targetProp.Value.(*resolink.Float3)
		if !ok {
			return fmt.Errorf("invalid float3 value")
		}

		// Update the appropriate component
		switch m.EditSubField {
		case "x":
			currentVal.X = floatVal
		case "y":
			currentVal.Y = floatVal
		case "z":
			currentVal.Z = floatVal
		default:
			return fmt.Errorf("unknown subfield: %s", m.EditSubField)
		}

		// Format as string and use shared function
		newValueStr := fmt.Sprintf("%.8f,%.8f,%.8f", currentVal.X, currentVal.Y, currentVal.Z)
		return shell.SetSlotProperty(m.Client, slotData.ID, m.EditFieldName, newValueStr)

	} else if targetProp.Type == "floatQ" {
		currentVal, ok := targetProp.Value.(*resolink.FloatQ)
		if !ok {
			return fmt.Errorf("invalid floatQ value")
		}

		// Update the appropriate component (w is read-only)
		switch m.EditSubField {
		case "x":
			currentVal.X = floatVal
		case "y":
			currentVal.Y = floatVal
		case "z":
			currentVal.Z = floatVal
		case "w":
			return fmt.Errorf("w component is read-only")
		default:
			return fmt.Errorf("unknown subfield: %s", m.EditSubField)
		}

		// Format as string and use shared function
		newValueStr := fmt.Sprintf("%.8f,%.8f,%.8f,%.8f", currentVal.X, currentVal.Y, currentVal.Z, currentVal.W)
		return shell.SetSlotProperty(m.Client, slotData.ID, m.EditFieldName, newValueStr)
	}

	return fmt.Errorf("unsupported composite type: %s", targetProp.Type)
}

// saveComponentField saves a component member
func (m *Model) saveComponentField(newValue string) error {
	compData, ok := m.InspectedData.(*shell.ComponentData)
	if !ok {
		return fmt.Errorf("invalid component data")
	}

	// If editing a subfield (e.g., Position.x), handle it specially
	if m.EditSubField != "" {
		return m.saveComponentSubfield(compData, newValue)
	}

	logger.Debug("saveComponentField: Using shared shell.SetComponentMember for %s=%s", m.EditFieldName, newValue)

	// Use the shared function from shell package
	if err := shell.SetComponentMember(m.Client, compData.ID, m.EditFieldName, newValue); err != nil {
		logger.Error("saveComponentField: SetComponentMember FAILED: %v", err)
		return fmt.Errorf("SetComponentMember call failed: %w", err)
	}

	logger.Debug("saveComponentField: SetComponentMember succeeded")
	return nil
}

// saveComponentSubfield handles editing individual components of composite types
func (m *Model) saveComponentSubfield(compData *shell.ComponentData, newValue string) error {
	// Find the member being edited
	var targetMember *shell.MemberData
	for i := range compData.Members {
		if compData.Members[i].Name == m.EditFieldName {
			targetMember = &compData.Members[i]
			break
		}
	}

	if targetMember == nil {
		return fmt.Errorf("member '%s' not found in component", m.EditFieldName)
	}

	// Parse the new value as a float
	var floatVal float64
	if _, err := fmt.Sscanf(newValue, "%f", &floatVal); err != nil {
		return fmt.Errorf("invalid number: %s", newValue)
	}

	logger.Debug("saveComponentSubfield: Updating %s.%s to %.8f", m.EditFieldName, m.EditSubField, floatVal)

	// Get current composite value and update the specific component
	valueMap, ok := targetMember.Value.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid composite value")
	}

	// Create a copy and update the component
	newMap := make(map[string]interface{})
	for k, v := range valueMap {
		newMap[k] = v
	}
	newMap[m.EditSubField] = floatVal

	// Format as string for the shared function
	var newValueStr string
	if x, okX := newMap["x"].(float64); okX {
		if y, okY := newMap["y"].(float64); okY {
			if z, okZ := newMap["z"].(float64); okZ {
				if w, okW := newMap["w"].(float64); okW {
					// floatQ format
					newValueStr = fmt.Sprintf("%.8f,%.8f,%.8f,%.8f", x, y, z, w)
				} else {
					// float3 format
					newValueStr = fmt.Sprintf("%.8f,%.8f,%.8f", x, y, z)
				}
			}
		}
	}

	if newValueStr == "" {
		return fmt.Errorf("could not format composite value")
	}

	return shell.SetComponentMember(m.Client, compData.ID, m.EditFieldName, newValueStr)
}
