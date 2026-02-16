package shell

import (
	"fmt"
	"strings"

	"github.com/Merith-TK/resh/pkg/resolink"
)

// InspectItem inspects either a slot or component by ID
// Returns (data, isSlot, error)
func InspectItem(client *resolink.Client, id string) (interface{}, bool, error) {
	// Try to determine if it's a slot or component by ID format
	// Slots typically have format like "Reso_xxx" or start with specific prefixes
	// Components have different ID formats

	// First, try as a slot
	slotData, slotErr := InspectSlot(client, id)
	if slotErr == nil {
		return slotData, true, nil
	}

	// If that failed, try as a component
	compData, compErr := InspectComponent(client, id)
	if compErr == nil {
		return compData, false, nil
	}

	// Both failed
	return nil, false, fmt.Errorf("failed to inspect item %s: slot error: %v, component error: %v",
		id, slotErr, compErr)
}

// IsSlotID attempts to determine if an ID belongs to a slot
func IsSlotID(id string) bool {
	// Heuristic: check if ID starts with common slot prefixes
	return strings.HasPrefix(id, "Reso_") ||
		strings.HasPrefix(id, "ID_") ||
		strings.Contains(id, "-slot-") // Some other pattern
}
