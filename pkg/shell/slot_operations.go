package shell

import (
	"fmt"
	"strings"

	"github.com/Merith-TK/resonite-sh/pkg/resolink"
)

// InspectSlot retrieves full slot data for inspection
func InspectSlot(client *resolink.Client, slotID string) (*SlotData, error) {
	// Convert display ID format (ID_xxx) to actual format (Reso_xxx) if needed
	actualID := slotID
	if strings.HasPrefix(slotID, "ID_") {
		actualID = strings.Replace(slotID, "ID_", "Reso_", 1)
	}

	resp, err := client.GetSlot(actualID, false, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get slot: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("slot error: %s", resp.ErrorInfo)
	}

	data := &SlotData{
		ID:         resp.Data.ID,
		Properties: make([]SlotProperty, 0),
	}

	// Parse slot properties
	if resp.Data.Name != nil {
		data.Properties = append(data.Properties, SlotProperty{
			Name:  "Name",
			Type:  "string",
			Value: resp.Data.Name.Value,
		})
	}

	if resp.Data.Tag != nil {
		data.Properties = append(data.Properties, SlotProperty{
			Name:  "Tag",
			Type:  "string",
			Value: resp.Data.Tag.Value,
		})
	}

	if resp.Data.Active != nil {
		data.Properties = append(data.Properties, SlotProperty{
			Name:  "Active",
			Type:  "bool",
			Value: resp.Data.Active.Value,
		})
	}

	if resp.Data.Persistent != nil {
		data.Properties = append(data.Properties, SlotProperty{
			Name:  "Persistent",
			Type:  "bool",
			Value: resp.Data.Persistent.Value,
		})
	}

	if resp.Data.Position != nil && resp.Data.Position.Value != nil {
		data.Properties = append(data.Properties, SlotProperty{
			Name:  "Position",
			Type:  "float3",
			Value: resp.Data.Position.Value,
		})
	}

	if resp.Data.Rotation != nil && resp.Data.Rotation.Value != nil {
		data.Properties = append(data.Properties, SlotProperty{
			Name:  "Rotation",
			Type:  "floatQ",
			Value: resp.Data.Rotation.Value,
		})
	}

	if resp.Data.Scale != nil && resp.Data.Scale.Value != nil {
		data.Properties = append(data.Properties, SlotProperty{
			Name:  "Scale",
			Type:  "float3",
			Value: resp.Data.Scale.Value,
		})
	}

	if resp.Data.OrderOffset != nil {
		data.Properties = append(data.Properties, SlotProperty{
			Name:  "OrderOffset",
			Type:  "long",
			Value: resp.Data.OrderOffset.Value,
		})
	}

	if resp.Data.Parent != nil {
		data.Properties = append(data.Properties, SlotProperty{
			Name:  "Parent",
			Type:  "reference",
			Value: resp.Data.Parent.TargetID,
		})
	}

	return data, nil
}

// SetSlotProperty updates a slot property
func SetSlotProperty(client *resolink.Client, slotID string, propertyName string, newValue interface{}) error {
	// Convert display ID format if needed
	actualID := slotID
	if strings.HasPrefix(slotID, "ID_") {
		actualID = strings.Replace(slotID, "ID_", "Reso_", 1)
	}

	// Get current slot data to determine property type
	data, err := InspectSlot(client, actualID)
	if err != nil {
		return fmt.Errorf("failed to get slot: %w", err)
	}

	// Find the property
	var propType string
	for _, prop := range data.Properties {
		if strings.EqualFold(prop.Name, propertyName) {
			propType = prop.Type
			propertyName = prop.Name // Use canonical name
			break
		}
	}

	if propType == "" {
		return fmt.Errorf("property %s not found in slot", propertyName)
	}

	// Convert value to appropriate type
	convertedValue, err := convertSlotValueByType(newValue, propType)
	if err != nil {
		return fmt.Errorf("value conversion failed: %w", err)
	}

	// Build update data
	updateData := &resolink.SlotDefinition{
		ID: actualID,
	}

	// Set the appropriate field
	switch propertyName {
	case "Name":
		updateData.Name = resolink.NewValueString(convertedValue.(string))
	case "Tag":
		updateData.Tag = resolink.NewValueString(convertedValue.(string))
	case "Active":
		updateData.Active = resolink.NewValueBool(convertedValue.(bool))
	case "Persistent":
		updateData.Persistent = resolink.NewValueBool(convertedValue.(bool))
	case "Position":
		vec := convertedValue.(*resolink.Float3)
		updateData.Position = resolink.NewValueFloat3(vec.X, vec.Y, vec.Z)
	case "Rotation":
		quat := convertedValue.(*resolink.FloatQ)
		updateData.Rotation = resolink.NewValueFloatQ(quat.X, quat.Y, quat.Z, quat.W)
	case "Scale":
		vec := convertedValue.(*resolink.Float3)
		updateData.Scale = resolink.NewValueFloat3(vec.X, vec.Y, vec.Z)
	case "OrderOffset":
		updateData.OrderOffset = resolink.NewValueLong(convertedValue.(int64))
	case "Parent":
		updateData.Parent = resolink.NewValueReference(convertedValue.(string))
	default:
		return fmt.Errorf("unknown property: %s", propertyName)
	}

	// Update the slot
	err = client.UpdateSlot(updateData)
	if err != nil {
		return fmt.Errorf("failed to update slot: %w", err)
	}

	return nil
}

// convertSlotValueByType converts a value string to the appropriate type for slots
func convertSlotValueByType(value interface{}, targetType string) (interface{}, error) {
	valueStr, ok := value.(string)
	if !ok {
		return value, nil // Already converted
	}

	switch targetType {
	case "bool":
		if valueStr == "true" || valueStr == "1" {
			return true, nil
		} else if valueStr == "false" || valueStr == "0" {
			return false, nil
		}
		return nil, fmt.Errorf("invalid bool value: %s (use true/false)", valueStr)

	case "long":
		var longVal int64
		if _, err := fmt.Sscanf(valueStr, "%d", &longVal); err != nil {
			return nil, fmt.Errorf("invalid long value: %s", valueStr)
		}
		return longVal, nil

	case "string":
		return valueStr, nil

	case "reference":
		// Convert display format if needed
		if strings.HasPrefix(valueStr, "ID_") {
			valueStr = strings.Replace(valueStr, "ID_", "Reso_", 1)
		}
		return valueStr, nil

	case "float3":
		// Parse format: x,y,z or [x,y,z]
		clean := strings.Trim(valueStr, "[]")
		var x, y, z float64
		if _, err := fmt.Sscanf(clean, "%f,%f,%f", &x, &y, &z); err != nil {
			return nil, fmt.Errorf("invalid float3 value: %s (use x,y,z)", valueStr)
		}
		return &resolink.Float3{X: x, Y: y, Z: z}, nil

	case "floatQ":
		// Parse format: x,y,z,w or [x,y,z,w]
		clean := strings.Trim(valueStr, "[]")
		var x, y, z, w float64
		if _, err := fmt.Sscanf(clean, "%f,%f,%f,%f", &x, &y, &z, &w); err != nil {
			return nil, fmt.Errorf("invalid floatQ value: %s (use x,y,z,w)", valueStr)
		}
		return &resolink.FloatQ{X: x, Y: y, Z: z, W: w}, nil

	default:
		return nil, fmt.Errorf("unsupported type: %s", targetType)
	}
}
