package cmd

import (
	"fmt"
	"strings"

	"github.com/Merith-TK/resh/pkg/resolink"
	"github.com/Merith-TK/resh/pkg/shell"
)

// ANSI color codes for terminal output
const (
	ColorReset  = "\033[0m"
	ColorCyan   = "\033[0;36m" // Active slots (like symlinks)
	ColorBlue   = "\033[0;34m" // Inactive slots (like directories)
	ColorWhite  = "\033[0;37m" // Components
	ColorGrey   = "\033[0;90m" // Greyed out (components)
	ColorYellow = "\033[0;33m" // Warnings/info
)

// Display utilities for REPL output

// formatSlotType returns a colored [slot] indicator based on active state
func formatSlotType(isActive bool) string {
	color := ColorCyan
	if !isActive {
		color = ColorBlue
	}
	return fmt.Sprintf("%s[slot]%s", color, ColorReset)
}

// formatComponentType returns a colored [comp] indicator
func formatComponentType() string {
	return fmt.Sprintf("%s[comp]%s", ColorGrey, ColorReset)
}

// formatID returns a yellow-colored ID prefix
func formatID(id string) string {
	// Replace Reso_ with ID_ for shorter display
	displayID := strings.Replace(id, "Reso_", "ID_", 1)
	return fmt.Sprintf("%s%s%s", ColorYellow, displayID, ColorReset)
}

// getPersistenceIndicator returns a text indicator for persistence
func getPersistenceIndicator(isPersistent bool) string {
	if isPersistent {
		return "P" // Persistent
	}
	return "T" // Temporary
}

// displaySlotHeader prints the slot name header
func displaySlotHeader(slotName string) {
	fmt.Println()
	fmt.Printf("Contents of slot: %s\n", slotName)
	fmt.Println()
}

// displayEmptySlot prints a message for empty slots
func displayEmptySlot() {
	fmt.Println("(empty slot - no children or components)")
	fmt.Println()
}

// displayRootProtection prints the root slot protection message
func displayRootProtection() {
	fmt.Printf("%s[Root slot - components hidden for safety]%s\n", ColorYellow, ColorReset)
	fmt.Println()
}

// displaySlotListing renders a complete slot listing
func displaySlotListing(listing *shell.SlotListing) {
	// Display header
	displaySlotHeader(listing.SlotName)

	// Display child slots
	for i := range listing.Children {
		displaySlotInfo(&listing.Children[i])
	}

	// Display components or protection message
	if listing.IsRootSlot {
		if len(listing.Children) > 0 {
			fmt.Println()
		}
		displayRootProtection()
	} else if len(listing.Components) > 0 {
		for i := range listing.Components {
			displayComponentInfo(&listing.Components[i])
		}
		fmt.Println()
	} else if len(listing.Children) == 0 {
		displayEmptySlot()
	} else {
		fmt.Println()
	}
}

// displaySlotInfo renders a SlotInfo
func displaySlotInfo(info *shell.SlotInfo) {
	persistIndicator := getPersistenceIndicator(info.IsPersistent)
	formattedType := formatSlotType(info.IsActive)
	formattedID := formatID(info.ID)
	fmt.Printf("%s %s %s %s\n", persistIndicator, formattedType, formattedID, info.Name)
}

// displayComponentInfo renders a ComponentInfo
func displayComponentInfo(info *shell.ComponentInfo) {
	compType := info.Type
	if compType == "" {
		compType = "<unknown type>"
	}

	// Strip [FrooxEngine]FrooxEngine. or similar prefixes
	compType = stripComponentPrefix(compType)

	persistIndicator := getPersistenceIndicator(info.IsPersistent)
	formattedType := formatComponentType()
	formattedID := formatID(info.ID)
	fmt.Printf("%s %s %s %s\n", persistIndicator, formattedType, formattedID, compType)
}

// stripComponentPrefix removes common component prefixes like [FrooxEngine]FrooxEngine.
func stripComponentPrefix(typeName string) string {
	// Strip [AssemblyName]Namespace. pattern
	// Example: [FrooxEngine]FrooxEngine.StaticFont -> StaticFont
	if idx := strings.LastIndex(typeName, "."); idx != -1 {
		return typeName[idx+1:]
	}
	return typeName
}

// displayComponentData renders full component inspection
func displayComponentData(data *shell.ComponentData) {
	fmt.Println()
	fmt.Printf("Component: %s %s\n", data.TypeName, formatID(data.ID))
	fmt.Println()

	if len(data.Members) == 0 {
		fmt.Println("(no members)")
		fmt.Println()
		return
	}

	// Find max widths for alignment
	maxIDWidth := 0
	maxTypeWidth := 0
	maxNameWidth := 0
	for i := range data.Members {
		idLen := len(strings.Replace(data.Members[i].ID, "Reso_", "ID_", 1))
		if idLen > maxIDWidth {
			maxIDWidth = idLen
		}
		typeLen := len(data.Members[i].Type)
		if typeLen > maxTypeWidth {
			maxTypeWidth = typeLen
		}
		nameLen := len(data.Members[i].Name)
		if nameLen > maxNameWidth {
			maxNameWidth = nameLen
		}
	}

	// Display each member
	for i := range data.Members {
		member := &data.Members[i]
		displayMemberInfo(member, maxIDWidth, maxTypeWidth, maxNameWidth)
	}
	fmt.Println()
}

// displayMemberInfo renders a single component member
func displayMemberInfo(member *shell.MemberData, idWidth, typeWidth, nameWidth int) {
	formattedID := formatID(member.ID)
	// Account for ANSI color codes in width calculation
	idPadding := idWidth + len(ColorYellow) + len(ColorReset)

	// Format value
	valueStr := formatMemberValue(member.Value)

	// Print: ID [Type] Name = Value
	fmt.Printf("%-*s [%-*s] %-*s = %s\n",
		idPadding, formattedID,
		typeWidth, member.Type,
		nameWidth, member.Name,
		valueStr)
}

// formatMemberValue formats a member value for display
func formatMemberValue(value interface{}) string {
	if value == nil {
		return "<null>"
	}

	switch v := value.(type) {
	case bool:
		if v {
			return "true"
		}
		return "false"
	case string:
		if v == "" {
			return "<empty>"
		}
		return fmt.Sprintf("%q", v)
	case float64:
		// JSON numbers are always float64
		if v == float64(int64(v)) {
			return fmt.Sprintf("%d", int64(v))
		}
		return fmt.Sprintf("%g", v)
	case map[string]interface{}:
		// Handle structured types like float3, floatQ, references, etc.

		// Check for reference type (has targetId field)
		if targetID, hasTarget := v["targetId"]; hasTarget {
			if targetStr, ok := targetID.(string); ok && targetStr != "" {
				displayID := strings.Replace(targetStr, "Reso_", "ID_", 1)
				return fmt.Sprintf("%s%s%s", ColorYellow, displayID, ColorReset)
			}
			return "<null>"
		}

		// Check for float3/floatQ (has x, y, z fields)
		if x, hasX := v["x"]; hasX {
			if y, hasY := v["y"]; hasY {
				if z, hasZ := v["z"]; hasZ {
					if w, hasW := v["w"]; hasW {
						// floatQ (x,y,z,w)
						return fmt.Sprintf("[%v, %v, %v, %v]", x, y, z, w)
					}
					// float3 (x,y,z)
					return fmt.Sprintf("[%v, %v, %v]", x, y, z)
				}
			}
		}
		// Other map types - fallback
		return fmt.Sprintf("%v", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// displaySlotData renders full slot inspection
func displaySlotData(data *shell.SlotData) {
	fmt.Println()
	fmt.Printf("Slot: %s\n", formatID(data.ID))
	fmt.Println()

	if len(data.Properties) == 0 {
		fmt.Println("(no properties)")
		fmt.Println()
		return
	}

	// Find max widths for alignment
	maxNameWidth := 0
	maxTypeWidth := 0
	for i := range data.Properties {
		nameLen := len(data.Properties[i].Name)
		if nameLen > maxNameWidth {
			maxNameWidth = nameLen
		}
		typeLen := len(data.Properties[i].Type)
		if typeLen > maxTypeWidth {
			maxTypeWidth = typeLen
		}
	}

	// Display each property
	for i := range data.Properties {
		prop := &data.Properties[i]
		displaySlotProperty(prop, maxNameWidth, maxTypeWidth)
	}
	fmt.Println()
}

// displaySlotProperty renders a single slot property
func displaySlotProperty(prop *shell.SlotProperty, nameWidth, typeWidth int) {
	// Format value
	valueStr := formatSlotValue(prop.Value, prop.Type)

	// Print: Name [Type] = Value
	fmt.Printf("%-*s [%-*s] = %s\n",
		nameWidth, prop.Name,
		typeWidth, prop.Type,
		valueStr)
}

// formatSlotValue formats a slot property value for display
func formatSlotValue(value interface{}, valueType string) string {
	if value == nil {
		return "<null>"
	}

	switch valueType {
	case "float3":
		if v, ok := value.(*resolink.Float3); ok {
			return fmt.Sprintf("[%g, %g, %g]", v.X, v.Y, v.Z)
		}
	case "floatQ":
		if v, ok := value.(*resolink.FloatQ); ok {
			return fmt.Sprintf("[%g, %g, %g, %g]", v.X, v.Y, v.Z, v.W)
		}
	case "reference":
		if v, ok := value.(string); ok && v != "" {
			displayID := strings.Replace(v, "Reso_", "ID_", 1)
			return fmt.Sprintf("%s%s%s", ColorYellow, displayID, ColorReset)
		}
		return "<null>"
	case "bool":
		if v, ok := value.(bool); ok {
			if v {
				return "true"
			}
			return "false"
		}
	case "string":
		if v, ok := value.(string); ok {
			if v == "" {
				return "<empty>"
			}
			return fmt.Sprintf("%q", v)
		}
	case "long":
		if v, ok := value.(int64); ok {
			return fmt.Sprintf("%d", v)
		}
	}

	return fmt.Sprintf("%v", value)
}
