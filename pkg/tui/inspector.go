package tui

import (
	"fmt"
	"strings"

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
			fieldIndex := 0
			for _, prop := range slotData.Properties {
				// Skip read-only fields
				if prop.Name == "Parent" || prop.Name == "ID" {
					continue
				}

				isSelected := (m.Focus == FocusInspector && fieldIndex == m.FieldCursor)
				isEditing := (isSelected && m.EditingField)

				// Render field name
				fieldLine := fmt.Sprintf("  %s: ", prop.Name)

				// Format and render value
				var valueStr string
				if isEditing {
					valueStr = m.EditBuffer
				} else {
					valueStr = formatPropertyValue(prop)
				}

				// Apply styling based on state
				if isEditing {
					b.WriteString(fieldNameStyle.Render(fieldLine))
					b.WriteString(editingFieldStyle.Render(valueStr + "_")) // Show cursor
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

// getEditableFieldCount returns number of editable fields for current item
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
				count++
			}
			return count
		}
	} else {
		if compData, ok := m.InspectedData.(*shell.ComponentData); ok {
			return len(compData.Members)
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
	case "float3", "floatQ", "reference":
		// Format complex types as strings
		return fmt.Sprintf("%v", prop.Value)
	}
	return fmt.Sprintf("%v", prop.Value)
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
				b.WriteString(fieldNameStyle.Render("  Members:\n"))

				// Show members with selection highlighting
				maxDisplay := 20
				if len(compData.Members) < maxDisplay {
					maxDisplay = len(compData.Members)
				}

				for i := 0; i < maxDisplay; i++ {
					member := compData.Members[i]

					isSelected := (m.Focus == FocusInspector && i == m.FieldCursor)
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
				}

				if len(compData.Members) > maxDisplay {
					remaining := len(compData.Members) - maxDisplay
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
				if fieldIndex == m.FieldCursor {
					return formatPropertyValue(prop)
				}
				fieldIndex++
			}
		}
	} else {
		if compData, ok := m.InspectedData.(*shell.ComponentData); ok {
			if m.FieldCursor >= 0 && m.FieldCursor < len(compData.Members) {
				member := compData.Members[m.FieldCursor]
				return fmt.Sprintf("%v", member.Value)
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
				if fieldIndex == m.FieldCursor {
					return prop.Type
				}
				fieldIndex++
			}
		}
	} else {
		if compData, ok := m.InspectedData.(*shell.ComponentData); ok {
			if m.FieldCursor >= 0 && m.FieldCursor < len(compData.Members) {
				member := compData.Members[m.FieldCursor]
				return member.Type
			}
		}
	}

	return ""
}

// toggleBoolField toggles a boolean field value
func (m *Model) toggleBoolField() {
	currentValue := m.getCurrentFieldValue()
	var newValue string
	if currentValue == "true" {
		newValue = "false"
	} else {
		newValue = "true"
	}

	if err := m.saveFieldValue(newValue); err != nil {
		m.ErrorMessage = fmt.Sprintf("Toggle failed: %v", err)
	} else {
		m.StatusMessage = "Field toggled"
		// Reload the item to show updated values
		if m.InspectedItem != nil {
			data, isSlot, err := shell.InspectItem(m.Client, m.InspectedItem.ID)
			if err == nil && isSlot == m.InspectedItem.IsSlot {
				m.InspectedData = data
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

	// Find the field being edited
	fieldIndex := 0
	var targetProp *shell.SlotProperty
	for i := range slotData.Properties {
		prop := &slotData.Properties[i]
		if prop.Name == "Parent" || prop.Name == "ID" {
			continue
		}
		if fieldIndex == m.FieldCursor {
			targetProp = prop
			break
		}
		fieldIndex++
	}

	if targetProp == nil {
		return fmt.Errorf("field not found")
	}

	// TODO: Implement UpdateSlot call with new value
	// For now, just return a placeholder error
	m.StatusMessage = fmt.Sprintf("Editing %s fields not yet implemented", targetProp.Name)
	return fmt.Errorf("slot updating not implemented yet")
}

// saveComponentField saves a component member
func (m *Model) saveComponentField(newValue string) error {
	compData, ok := m.InspectedData.(*shell.ComponentData)
	if !ok {
		return fmt.Errorf("invalid component data")
	}

	if m.FieldCursor < 0 || m.FieldCursor >= len(compData.Members) {
		return fmt.Errorf("field index out of range")
	}

	member := compData.Members[m.FieldCursor]

	// TODO: Implement UpdateComponent call with new value
	// For now, just return a placeholder error
	m.StatusMessage = fmt.Sprintf("Editing %s fields not yet implemented", member.Name)
	return fmt.Errorf("component updating not implemented yet")
}
