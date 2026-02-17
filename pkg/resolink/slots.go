package resolink

import (
	"encoding/json"
	"fmt"

	"github.com/Merith-TK/resh/pkg/logger"
)

// GetSlot retrieves slot information
func (c *Client) GetSlot(slotID string, includeComponents bool, depth int) (*SlotDataResponse, error) {
	msg := &GetSlotMessage{
		BaseMessage: BaseMessage{
			Type: "getSlot",
		},
		SlotID:               slotID,
		IncludeComponentData: includeComponents,
		Depth:                depth,
	}

	rawResp, err := c.sendMessage(msg)
	if err != nil {
		logger.Error("GetSlot sendMessage error: %v", err)
		return nil, err
	}

	logger.Debug("GetSlot raw response for %s: %s", slotID, string(rawResp))

	// Parse as slot data response
	var resp SlotDataResponse
	if err := json.Unmarshal(rawResp, &resp); err != nil {
		logger.Error("GetSlot unmarshal error: %v", err)
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for error
	if !resp.Success {
		logger.Error("GetSlot operation failed: %s", resp.ErrorInfo)
		return nil, fmt.Errorf("slot operation failed: %s", resp.ErrorInfo)
	}

	return &resp, nil
}

// AddSlot creates a new slot under the specified parent
func (c *Client) AddSlot(data *SlotDefinition) (*SlotDataResponse, error) {
	msg := &AddSlotMessage{
		BaseMessage: BaseMessage{
			Type: "addSlot",
		},
		Data: data,
	}

	rawResp, err := c.sendMessage(msg)
	if err != nil {
		return nil, err
	}

	// Check for error response
	var errResp ErrorResponse
	if err := json.Unmarshal(rawResp, &errResp); err == nil && errResp.Error != "" {
		return nil, fmt.Errorf("server error: %s", errResp.Error)
	}

	// Parse response
	var resp SlotDataResponse
	if err := json.Unmarshal(rawResp, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &resp, nil
}

// UpdateSlot updates slot properties
func (c *Client) UpdateSlot(data *SlotDefinition) error {
	if data.ID == "" {
		return fmt.Errorf("slot ID is required for update")
	}

	msg := &UpdateSlotMessage{
		BaseMessage: BaseMessage{
			Type: "updateSlot",
		},
		Data: data,
	}

	// Log what we're about to send
	msgJSON, _ := json.MarshalIndent(msg, "", "  ")
	logger.Debug("UpdateSlot sending message:\n%s", string(msgJSON))

	// Also log the struct details
	logger.JSON("SlotDefinition", data)

	rawResp, err := c.sendMessage(msg)
	if err != nil {
		logger.Error("UpdateSlot sendMessage error: %v", err)
		return err
	}

	logger.Debug("UpdateSlot response: %s", string(rawResp))

	// Check for error response
	var errResp ErrorResponse
	if err := json.Unmarshal(rawResp, &errResp); err == nil && errResp.Error != "" {
		logger.Error("UpdateSlot server error: %s", errResp.Error)
		return fmt.Errorf("server error: %s", errResp.Error)
	}

	logger.Debug("UpdateSlot completed successfully")
	return nil
}

// RemoveSlot deletes a slot
func (c *Client) RemoveSlot(slotID string) error {
	msg := &RemoveSlotMessage{
		BaseMessage: BaseMessage{
			Type: "removeSlot",
		},
		SlotID: slotID,
	}

	rawResp, err := c.sendMessage(msg)
	if err != nil {
		return err
	}

	// Check for error response
	var errResp ErrorResponse
	if err := json.Unmarshal(rawResp, &errResp); err == nil && errResp.Error != "" {
		return fmt.Errorf("server error: %s", errResp.Error)
	}

	return nil
}

// FindSlotByName searches for a slot by name (helper method)
// This is a convenience method that gets a slot and searches its children
func (c *Client) FindSlotByName(parentID, name string) (*SlotResponse, error) {
	// Get parent slot with children (depth 1)
	resp, err := c.GetSlot(parentID, false, 1)
	if err != nil {
		return nil, err
	}

	// Search children for matching name
	for _, child := range resp.Data.Children {
		if child.Name != nil && child.Name.Value == name {
			// Get full slot data
			childResp, err := c.GetSlot(child.ID, false, 0)
			if err != nil {
				return nil, err
			}
			return childResp.Data, nil
		}
	}

	return nil, fmt.Errorf("slot with name '%s' not found", name)
}

// ListChildren lists all child slots of a parent (helper method)
func (c *Client) ListChildren(parentID string) ([]SlotReference, error) {
	resp, err := c.GetSlot(parentID, false, 1)
	if err != nil {
		return nil, err
	}

	return resp.Data.Children, nil
}
