package shell

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Merith-TK/resh/pkg/resolink"
)

// InspectComponent retrieves full component data for inspection
func InspectComponent(client *resolink.Client, componentID string) (*ComponentData, error) {
	// Convert display ID format (ID_xxx) to actual format (Reso_xxx) if needed
	actualID := componentID
	if strings.HasPrefix(componentID, "ID_") {
		actualID = strings.Replace(componentID, "ID_", "Reso_", 1)
	}

	resp, err := client.GetComponent(actualID)
	if err != nil {
		return nil, fmt.Errorf("failed to get component: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("component error: %s", resp.ErrorInfo)
	}

	data := &ComponentData{
		ID:            resp.Data.ID,
		ComponentType: resp.Data.ComponentType,
		TypeName:      parseComponentTypeName(resp.Data.ComponentType),
	}

	// Parse members
	for name, memberData := range resp.Data.Members {
		// Member data is a map with $type, id, value
		memberMap, ok := memberData.(map[string]interface{})
		if !ok {
			continue
		}

		member := MemberData{
			Name: name,
		}

		if typeVal, ok := memberMap["$type"].(string); ok {
			member.Type = typeVal
		}
		if idVal, ok := memberMap["id"].(string); ok {
			member.ID = idVal
		}

		// For reference types, targetId is at the same level, not under "value"
		if member.Type == "reference" {
			if targetID, ok := memberMap["targetId"].(string); ok && targetID != "" {
				member.Value = map[string]interface{}{
					"targetId": targetID,
				}
			} else {
				member.Value = map[string]interface{}{
					"targetId": nil,
				}
			}
		} else if valVal, ok := memberMap["value"]; ok {
			member.Value = valVal
		}

		data.Members = append(data.Members, member)
	}

	// Sort members alphabetically by name for consistent display
	sortMembersByName(data.Members)

	return data, nil
}

// sortMembersByName sorts component members alphabetically by name
func sortMembersByName(members []MemberData) {
	sort.Slice(members, func(i, j int) bool {
		return members[i].Name < members[j].Name
	})
}

// parseComponentTypeName extracts the readable name from component type
// Example: [FrooxEngine]FrooxEngine.StaticLocaleProvider -> StaticLocaleProvider
func parseComponentTypeName(fullType string) string {
	if idx := strings.LastIndex(fullType, "."); idx != -1 {
		return fullType[idx+1:]
	}
	return fullType
}

// SetComponentMember updates a component member value by ID or name
func SetComponentMember(client *resolink.Client, componentID string, memberIDOrName string, newValue interface{}) error {
	// Convert display ID format if needed
	actualComponentID := componentID
	if strings.HasPrefix(componentID, "ID_") {
		actualComponentID = strings.Replace(componentID, "ID_", "Reso_", 1)
	}

	// Get current component data to find member
	data, err := InspectComponent(client, actualComponentID)
	if err != nil {
		return fmt.Errorf("failed to get component: %w", err)
	}

	// Check if memberIDOrName is an ID or a name
	var actualMemberID string
	var memberType string
	var memberName string

	// If it starts with ID_ or Reso_, treat as ID
	if strings.HasPrefix(memberIDOrName, "ID_") || strings.HasPrefix(memberIDOrName, "Reso_") {
		actualMemberID = memberIDOrName
		if strings.HasPrefix(actualMemberID, "ID_") {
			actualMemberID = strings.Replace(actualMemberID, "ID_", "Reso_", 1)
		}

		// Find by ID
		for _, member := range data.Members {
			if member.ID == actualMemberID {
				memberType = member.Type
				memberName = member.Name
				break
			}
		}
	} else {
		// Treat as name, find by name
		for _, member := range data.Members {
			if member.Name == memberIDOrName {
				actualMemberID = member.ID
				memberType = member.Type
				memberName = member.Name
				break
			}
		}
	}

	if memberType == "" {
		return fmt.Errorf("member %s not found in component", memberIDOrName)
	}

	// Convert value to appropriate type
	convertedValue, err := convertValueByType(newValue, memberType)
	if err != nil {
		return fmt.Errorf("value conversion failed: %w", err)
	}

	// Build update data with the member to update
	// IMPORTANT: The key must be the member NAME, not the ID
	var memberUpdate map[string]interface{}

	if memberType == "reference" {
		// For references, targetId goes at the same level as $type and id
		refMap := map[string]interface{}{
			"$type": memberType,
			"id":    actualMemberID,
		}
		// Extract targetId from converted value
		if refData, ok := convertedValue.(map[string]interface{}); ok {
			if targetID := refData["targetId"]; targetID != nil {
				refMap["targetId"] = targetID
			}
		}
		memberUpdate = refMap
	} else {
		// For other types, value is nested
		memberUpdate = map[string]interface{}{
			"$type": memberType,
			"id":    actualMemberID,
			"value": convertedValue,
		}
	}

	updateData := &resolink.ComponentDefinition{
		ID: actualComponentID,
		Members: map[string]interface{}{
			memberName: memberUpdate,
		},
	}

	// Update the component member
	err = client.UpdateComponent(updateData)
	if err != nil {
		return fmt.Errorf("failed to update component: %w", err)
	}

	return nil
}

// convertValueByType converts a value string to the appropriate type
func convertValueByType(value interface{}, targetType string) (interface{}, error) {
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

	case "int":
		var intVal int
		if _, err := fmt.Sscanf(valueStr, "%d", &intVal); err != nil {
			return nil, fmt.Errorf("invalid int value: %s", valueStr)
		}
		return intVal, nil

	case "float":
		var floatVal float64
		if _, err := fmt.Sscanf(valueStr, "%f", &floatVal); err != nil {
			return nil, fmt.Errorf("invalid float value: %s", valueStr)
		}
		return floatVal, nil

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

	case "colorX", "color":
		// Parse format: r,g,b,a or [r,g,b,a]
		clean := strings.Trim(valueStr, "[]")
		var r, g, b, a float64
		if _, err := fmt.Sscanf(clean, "%f,%f,%f,%f", &r, &g, &b, &a); err != nil {
			return nil, fmt.Errorf("invalid colorX value: %s (use r,g,b,a)", valueStr)
		}
		return map[string]interface{}{
			"r": r,
			"g": g,
			"b": b,
			"a": a,
		}, nil

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

	case "double":
		var doubleVal float64
		if _, err := fmt.Sscanf(valueStr, "%f", &doubleVal); err != nil {
			return nil, fmt.Errorf("invalid double value: %s", valueStr)
		}
		return doubleVal, nil

	case "long":
		var longVal int64
		if _, err := fmt.Sscanf(valueStr, "%d", &longVal); err != nil {
			return nil, fmt.Errorf("invalid long value: %s", valueStr)
		}
		return longVal, nil

	case "int2":
		// Parse format: x,y or [x,y]
		clean := strings.Trim(valueStr, "[]")
		var x, y int
		if _, err := fmt.Sscanf(clean, "%d,%d", &x, &y); err != nil {
			return nil, fmt.Errorf("invalid int2 value: %s (use x,y)", valueStr)
		}
		return map[string]interface{}{
			"x": x,
			"y": y,
		}, nil

	case "string", "Uri":
		return valueStr, nil

	case "reference":
		// Handle reference type - convert ID format if needed
		refID := valueStr
		if refID == "" || refID == "null" || refID == "<null>" {
			return map[string]interface{}{
				"targetId": nil,
			}, nil
		}
		if strings.HasPrefix(refID, "ID_") {
			refID = strings.Replace(refID, "ID_", "Reso_", 1)
		}
		return map[string]interface{}{
			"targetId": refID,
		}, nil

	default:
		// For unknown types, pass as string and let the API handle it
		return valueStr, nil
	}
}
