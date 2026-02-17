package tui

import (
	"fmt"
	"strings"

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
	if fieldType == "" {
		m.ErrorMessage = "Cannot edit this field"
		return
	}

	m.EditFieldType = fieldType

	// For bools, toggle immediately instead of entering edit mode
	if fieldType == "bool" {
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
						return formatPropertyValue(prop)
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
	currentValue := m.getCurrentFieldValue()
	if currentValue == "" {
		m.ErrorMessage = "Could not get current field value"
		return
	}

	var newValue string
	if currentValue == "true" {
		newValue = "false"
	} else {
		newValue = "true"
	}

	if err := m.saveFieldValue(newValue); err != nil {
		m.ErrorMessage = fmt.Sprintf("Toggle failed: %v", err)
	} else {
		m.StatusMessage = fmt.Sprintf("Field toggled to %s", newValue)
		// Reload the item to show updated values
		if m.InspectedItem != nil {
			data, isSlot, err := shell.InspectItem(m.Client, m.InspectedItem.ID)
			if err == nil && isSlot == m.InspectedItem.IsSlot {
				m.InspectedData = data
			} else if err != nil {
				m.ErrorMessage = fmt.Sprintf("Reload failed: %v", err)
			}
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

	// Find the property being edited using EditFieldName
	var targetProp *shell.SlotProperty
	for i := range slotData.Properties {
		if slotData.Properties[i].Name == m.EditFieldName {
			targetProp = &slotData.Properties[i]
			break
		}
	}

	if targetProp == nil {
		return fmt.Errorf("field not found")
	}

	// Handle subfield editing for composite types
	if m.EditSubField != "" {
		// Editing a component of a composite type (float3, floatQ)
		parsedFloat, err := fmt.Sscanf(newValue, "%f", new(float64))
		if parsedFloat != 1 || err != nil {
			return fmt.Errorf("invalid number: %s", newValue)
		}

		var floatVal float64
		fmt.Sscanf(newValue, "%f", &floatVal)

		// Get current composite value and update the specific component
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
			}
			// Create new value with updated component
			parsedValue := resolink.NewValueFloat3(currentVal.X, currentVal.Y, currentVal.Z)

			slotDef := &resolink.SlotDefinition{ID: slotData.ID}
			switch targetProp.Name {
			case "Position":
				slotDef.Position = parsedValue
			case "Scale":
				slotDef.Scale = parsedValue
			default:
				return fmt.Errorf("unknown float3 field: %s", targetProp.Name)
			}

			return m.Client.UpdateSlot(slotDef)

		} else if targetProp.Type == "floatQ" {
			currentVal, ok := targetProp.Value.(*resolink.FloatQ)
			if !ok {
				return fmt.Errorf("invalid floatQ value")
			}
			// Update the appropriate component
			switch m.EditSubField {
			case "x":
				currentVal.X = floatVal
			case "y":
				currentVal.Y = floatVal
			case "z":
				currentVal.Z = floatVal
				// w is read-only per user request
			}
			// Create new value with updated component
			parsedValue := resolink.NewValueFloatQ(currentVal.X, currentVal.Y, currentVal.Z, currentVal.W)

			slotDef := &resolink.SlotDefinition{ID: slotData.ID}
			switch targetProp.Name {
			case "Rotation":
				slotDef.Rotation = parsedValue
			default:
				return fmt.Errorf("unknown floatQ field: %s", targetProp.Name)
			}

			return m.Client.UpdateSlot(slotDef)
		}

		return fmt.Errorf("unsupported composite type: %s", targetProp.Type)
	}

	// Non-composite field editing
	parsedValue, err := m.parseSlotValue(newValue, targetProp.Type)
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	// Build SlotDefinition with just the field being updated
	slotDef := &resolink.SlotDefinition{
		ID: slotData.ID,
	}

	// Set the appropriate field
	switch targetProp.Name {
	case "Name":
		slotDef.Name = parsedValue.(*resolink.ValueString)
	case "Active":
		slotDef.Active = parsedValue.(*resolink.ValueBool)
	case "Tag":
		slotDef.Tag = parsedValue.(*resolink.ValueString)
	case "Persistent":
		slotDef.Persistent = parsedValue.(*resolink.ValueBool)
	case "Position":
		slotDef.Position = parsedValue.(*resolink.ValueFloat3)
	case "Rotation":
		slotDef.Rotation = parsedValue.(*resolink.ValueFloatQ)
	case "Scale":
		slotDef.Scale = parsedValue.(*resolink.ValueFloat3)
	case "OrderOffset":
		slotDef.OrderOffset = parsedValue.(*resolink.ValueLong)
	default:
		return fmt.Errorf("field %s cannot be edited", targetProp.Name)
	}

	// Call UpdateSlot
	if err := m.Client.UpdateSlot(slotDef); err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	return nil
}

// saveComponentField saves a component member
func (m *Model) saveComponentField(newValue string) error {
	compData, ok := m.InspectedData.(*shell.ComponentData)
	if !ok {
		return fmt.Errorf("invalid component data")
	}

	// Find the member being edited using EditFieldName
	var targetMember *shell.MemberData
	for i := range compData.Members {
		if compData.Members[i].Name == m.EditFieldName {
			targetMember = &compData.Members[i]
			break
		}
	}

	if targetMember == nil {
		return fmt.Errorf("member not found")
	}

	// Handle subfield editing for composite types
	if m.EditSubField != "" {
		// Editing a component of a composite type (float2/3/4, floatQ)
		parsedFloat, err := fmt.Sscanf(newValue, "%f", new(float64))
		if parsedFloat != 1 || err != nil {
			return fmt.Errorf("invalid number: %s", newValue)
		}

		var floatVal float64
		fmt.Sscanf(newValue, "%f", &floatVal)

		// Get current composite value and update the specific component
		if valueMap, ok := targetMember.Value.(map[string]interface{}); ok {
			// Create a copy of the map to update
			newMap := make(map[string]interface{})
			for k, v := range valueMap {
				newMap[k] = v
			}
			// Update the specific component
			newMap[m.EditSubField] = floatVal

			// Build ComponentDefinition with the updated member
			compDef := &resolink.ComponentDefinition{
				ID:            compData.ID,
				ComponentType: compData.ComponentType,
				Members: map[string]interface{}{
					targetMember.Name: newMap,
				},
			}

			return m.Client.UpdateComponent(compDef)
		}

		return fmt.Errorf("invalid composite value")
	}

	// Non-composite field editing
	parsedValue, err := m.parseComponentValue(newValue, targetMember.Type)
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	// Build ComponentDefinition with the updated member
	compDef := &resolink.ComponentDefinition{
		ID:            compData.ID,
		ComponentType: compData.ComponentType,
		Members: map[string]interface{}{
			targetMember.Name: parsedValue,
		},
	}

	// Call UpdateComponent
	if err := m.Client.UpdateComponent(compDef); err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	return nil
}

// parseSlotValue parses a string value into the appropriate type for slot properties
func (m *Model) parseSlotValue(valueStr string, fieldType string) (interface{}, error) {
	switch fieldType {
	case "string":
		return resolink.NewValueString(valueStr), nil

	case "bool":
		if valueStr == "true" {
			return resolink.NewValueBool(true), nil
		} else if valueStr == "false" {
			return resolink.NewValueBool(false), nil
		}
		return nil, fmt.Errorf("invalid bool value: %s (must be true or false)", valueStr)

	case "long":
		var longVal int64
		if _, err := fmt.Sscanf(valueStr, "%d", &longVal); err != nil {
			return nil, fmt.Errorf("invalid long value: %s", valueStr)
		}
		return resolink.NewValueLong(longVal), nil

	case "float3":
		// Parse format: x,y,z or [x,y,z]
		clean := strings.Trim(valueStr, "[]")
		var x, y, z float64
		if _, err := fmt.Sscanf(clean, "%f,%f,%f", &x, &y, &z); err != nil {
			return nil, fmt.Errorf("invalid float3 value: %s (use x,y,z)", valueStr)
		}
		return resolink.NewValueFloat3(x, y, z), nil

	case "floatQ":
		// Parse format: x,y,z,w or [x,y,z,w]
		clean := strings.Trim(valueStr, "[]")
		var x, y, z, w float64
		if _, err := fmt.Sscanf(clean, "%f,%f,%f,%f", &x, &y, &z, &w); err != nil {
			return nil, fmt.Errorf("invalid floatQ value: %s (use x,y,z,w)", valueStr)
		}
		return resolink.NewValueFloatQ(x, y, z, w), nil

	case "reference":
		// Convert display format if needed
		if strings.HasPrefix(valueStr, "ID_") {
			valueStr = strings.Replace(valueStr, "ID_", "Reso_", 1)
		}
		return resolink.NewValueReference(valueStr), nil

	default:
		return nil, fmt.Errorf("unsupported field type: %s", fieldType)
	}
}

// parseComponentValue parses a string value into the appropriate type for component members
func (m *Model) parseComponentValue(valueStr string, fieldType string) (interface{}, error) {
	switch fieldType {
	case "bool":
		if valueStr == "true" || valueStr == "1" {
			return true, nil
		} else if valueStr == "false" || valueStr == "0" {
			return false, nil
		}
		return nil, fmt.Errorf("invalid bool value: %s (use true/false)", valueStr)

	case "int":
		var intVal int
		if _, err := fmt.Sscanf(valueStr, "%d", &intVal); err != nil {
			return nil, fmt.Errorf("invalid int value: %s", valueStr)
		}
		return intVal, nil

	case "long":
		var longVal int64
		if _, err := fmt.Sscanf(valueStr, "%d", &longVal); err != nil {
			return nil, fmt.Errorf("invalid long value: %s", valueStr)
		}
		return longVal, nil

	case "float", "double":
		var floatVal float64
		if _, err := fmt.Sscanf(valueStr, "%f", &floatVal); err != nil {
			return nil, fmt.Errorf("invalid %s value: %s", fieldType, valueStr)
		}
		return floatVal, nil

	case "string":
		return valueStr, nil

	case "float2":
		// Parse format: x,y or [x,y]
		clean := strings.Trim(valueStr, "[]")
		var x, y float64
		if _, err := fmt.Sscanf(clean, "%f,%f", &x, &y); err != nil {
			return nil, fmt.Errorf("invalid float2 value: %s (use x,y)", valueStr)
		}
		return map[string]interface{}{
			"x": x,
			"y": y,
		}, nil

	case "float3":
		// Parse format: x,y,z or [x,y,z]
		clean := strings.Trim(valueStr, "[]")
		var x, y, z float64
		if _, err := fmt.Sscanf(clean, "%f,%f,%f", &x, &y, &z); err != nil {
			return nil, fmt.Errorf("invalid float3 value: %s (use x,y,z)", valueStr)
		}
		return map[string]interface{}{
			"x": x,
			"y": y,
			"z": z,
		}, nil

	case "floatQ":
		// Parse format: x,y,z,w or [x,y,z,w]
		clean := strings.Trim(valueStr, "[]")
		var x, y, z, w float64
		if _, err := fmt.Sscanf(clean, "%f,%f,%f,%f", &x, &y, &z, &w); err != nil {
			return nil, fmt.Errorf("invalid floatQ value: %s (use x,y,z,w)", valueStr)
		}
		return map[string]interface{}{
			"x": x,
			"y": y,
			"z": z,
			"w": w,
		}, nil

	case "color", "colorX":
		// Parse format: r,g,b,a or [r,g,b,a]
		clean := strings.Trim(valueStr, "[]")
		var r, g, b, a float64
		if _, err := fmt.Sscanf(clean, "%f,%f,%f,%f", &r, &g, &b, &a); err != nil {
			return nil, fmt.Errorf("invalid color value: %s (use r,g,b,a)", valueStr)
		}
		return map[string]interface{}{
			"r": r,
			"g": g,
			"b": b,
			"a": a,
		}, nil

	default:
		// For unknown types, try to return as-is
		// This allows for more complex types to potentially work
		return valueStr, nil
	}
}
