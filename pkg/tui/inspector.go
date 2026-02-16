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
		b.WriteString(helpStyle.Render("  ↑/↓:nav Enter:edit Tab:tree"))
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
	b.WriteString("\n")

	// If we have inspected data, show it
	if m.InspectedData != nil {
		if slotData, ok := m.InspectedData.(*shell.SlotData); ok {
			// Display all properties
			for _, prop := range slotData.Properties {
				b.WriteString(fieldNameStyle.Render(fmt.Sprintf("  %s: ", prop.Name)))

				// Format value based on type
				valueStr := formatPropertyValue(prop)
				b.WriteString(fieldValueStyle.Render(valueStr))
				b.WriteString("\n")
			}
		}
	}
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
	b.WriteString(fieldValueStyle.Render(m.InspectedItem.Type))
	b.WriteString("\n\n")

	// If we have inspected data, show component members
	if m.InspectedData != nil {
		if compData, ok := m.InspectedData.(*shell.ComponentData); ok {
			b.WriteString(fieldNameStyle.Render("  Members:\n"))

			// Show members (Members is an array)
			maxDisplay := 10
			if len(compData.Members) < maxDisplay {
				maxDisplay = len(compData.Members)
			}

			for i := 0; i < maxDisplay; i++ {
				member := compData.Members[i]
				b.WriteString(fmt.Sprintf("    %s: ", member.Name))

				// Format value based on type
				valueStr := fmt.Sprintf("%v", member.Value)
				if len(valueStr) > 50 {
					valueStr = valueStr[:47] + "..."
				}
				b.WriteString(fieldValueStyle.Render(valueStr))
				b.WriteString("\n")
			}

			if len(compData.Members) > maxDisplay {
				remaining := len(compData.Members) - maxDisplay
				b.WriteString(helpStyle.Render(fmt.Sprintf("    ... and %d more", remaining)))
				b.WriteString("\n")
			}
		}
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
	// TODO: Calculate actual number of editable fields
	// For now, just prevent going too far
	m.FieldCursor++
}

// StartEditingField starts editing the currently selected field
func (m *Model) StartEditingField() {
	// TODO: Implement field editing
	m.EditingField = true
	m.EditBuffer = ""
}

// StopEditingField stops editing and saves/cancels
func (m *Model) StopEditingField(save bool) {
	if save {
		// TODO: Implement save logic
	}
	m.EditingField = false
	m.EditBuffer = ""
}
