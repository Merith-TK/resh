package shell

import (
	"fmt"
	"strings"

	"github.com/Merith-TK/resonite-sh/pkg/resolink"
)

// InitializeRESHData ensures the /RESH.DATA slot exists with proper components
func InitializeRESHData(client *resolink.Client, rootSlotID string) (string, error) {
	// Try to navigate to /RESH.DATA
	resp, err := client.GetSlot(rootSlotID, false, 1)
	if err != nil {
		return "", fmt.Errorf("failed to get root slot: %w", err)
	}

	// Look for RESH.DATA child
	var dataSlotID string
	for _, child := range resp.Data.Children {
		if child.Name != nil && child.Name.Value == "RESH.DATA" {
			dataSlotID = child.ID
			break
		}
	}

	// If not found, create it
	if dataSlotID == "" {
		dataSlotID, err = createRESHDataSlot(client, rootSlotID)
		if err != nil {
			return "", fmt.Errorf("failed to create RESH.DATA slot: %w", err)
		}
	}

	// Verify/create required components
	err = ensureRESHDataComponents(client, dataSlotID)
	if err != nil {
		return "", fmt.Errorf("failed to setup RESH.DATA components: %w", err)
	}

	return dataSlotID, nil
}

// createRESHDataSlot creates a new RESH.DATA slot under root
func createRESHDataSlot(client *resolink.Client, parentSlotID string) (string, error) {
	slotDef := &resolink.SlotDefinition{
		Name:       resolink.NewValueString("RESH.DATA"),
		Persistent: resolink.NewValueBool(true),
		Active:     resolink.NewValueBool(true),
		Tag:        resolink.NewValueString("RESHData"),
		Parent:     resolink.NewValueReference(parentSlotID),
	}

	resp, err := client.AddSlot(slotDef)
	if err != nil {
		return "", err
	}

	if resp == nil || resp.Data == nil {
		return "", fmt.Errorf("invalid response from AddSlot")
	}

	if !resp.Success {
		return "", fmt.Errorf("create slot failed: %s", resp.ErrorInfo)
	}

	return resp.Data.ID, nil
}

// ensureRESHDataComponents ensures the required components exist on RESH.DATA
func ensureRESHDataComponents(client *resolink.Client, slotID string) error {
	// Get current components
	resp, err := client.GetSlot(slotID, true, 0)
	if err != nil {
		return err
	}

	hasDynamicVarSpace := false
	hasSelfReference := false

	// Check existing components
	for _, comp := range resp.Data.Components {
		compType := comp.ComponentType
		if strings.Contains(compType, "DynamicVariableSpace") {
			hasDynamicVarSpace = true
		}
		if strings.Contains(compType, "DynamicReferenceVariable") {
			hasSelfReference = true
		}
	}

	// Create DynamicVariableSpace if needed
	if !hasDynamicVarSpace {
		err := createDynamicVariableSpace(client, slotID)
		if err != nil {
			return fmt.Errorf("failed to create DynamicVariableSpace: %w", err)
		}
	}

	// Create self-reference variable if needed
	if !hasSelfReference {
		err := createSelfReferenceVariable(client, slotID)
		if err != nil {
			return fmt.Errorf("failed to create self-reference: %w", err)
		}
	}

	return nil
}

// createDynamicVariableSpace creates a DynamicVariableSpace component with value "RESH"
func createDynamicVariableSpace(client *resolink.Client, slotID string) error {
	compDef := &resolink.ComponentDefinition{
		ComponentType: "[FrooxEngine]FrooxEngine.DynamicVariableSpace",
		Members: map[string]interface{}{
			"SpaceName": map[string]interface{}{
				"$type": "string",
				"value": "RESH",
			},
		},
	}

	resp, err := client.AddComponent(slotID, compDef)
	if err != nil {
		return err
	}

	if resp == nil {
		return fmt.Errorf("invalid response from AddComponent")
	}

	if !resp.Success {
		return fmt.Errorf("create component failed: %s", resp.ErrorInfo)
	}

	return nil
}

// createSelfReferenceVariable creates a DynamicReferenceVariable<Slot> pointing to the slot itself
func createSelfReferenceVariable(client *resolink.Client, slotID string) error {
	compDef := &resolink.ComponentDefinition{
		ComponentType: "[FrooxEngine]FrooxEngine.DynamicReferenceVariable`1[[FrooxEngine.Slot, FrooxEngine]]",
		Members: map[string]interface{}{
			"VariableName": map[string]interface{}{
				"$type": "string",
				"value": "RESH.DATA",
			},
			"Reference": map[string]interface{}{
				"$type":    "reference",
				"targetId": slotID,
			},
		},
	}

	resp, err := client.AddComponent(slotID, compDef)
	if err != nil {
		return err
	}

	if resp == nil {
		return fmt.Errorf("invalid response from AddComponent")
	}

	if !resp.Success {
		return fmt.Errorf("create component failed: %s", resp.ErrorInfo)
	}

	return nil
}
