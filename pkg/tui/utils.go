package tui

import (
	"fmt"

	"github.com/Merith-TK/resh/pkg/resolink"
	"github.com/Merith-TK/resh/pkg/shell"
)

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
}</content>
<parameter name="filePath">d:\Workspace\github.com\Merith-TK\resonite-sh\pkg\tui\utils.go