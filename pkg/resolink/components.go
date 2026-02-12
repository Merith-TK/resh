package resolink

import (
	"encoding/json"
	"fmt"
)

// GetComponent retrieves component information
func (c *Client) GetComponent(componentID string) (*ComponentDataResponse, error) {
	msg := &GetComponentMessage{
		BaseMessage: BaseMessage{
			Type: "getComponent",
		},
		ComponentID: componentID,
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
	var resp ComponentDataResponse
	if err := json.Unmarshal(rawResp, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &resp, nil
}

// AddComponent adds a component to a slot
func (c *Client) AddComponent(slotID string, data *ComponentDefinition) (*ComponentDataResponse, error) {
	if data.ComponentType == "" {
		return nil, fmt.Errorf("component type is required")
	}

	msg := &AddComponentMessage{
		BaseMessage: BaseMessage{
			Type: "addComponent",
		},
		ContainerSlotID: slotID,
		Data:            data,
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
	var resp ComponentDataResponse
	if err := json.Unmarshal(rawResp, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &resp, nil
}

// UpdateComponent updates component members
func (c *Client) UpdateComponent(data *ComponentDefinition) error {
	if data.ID == "" {
		return fmt.Errorf("component ID is required for update")
	}

	msg := &UpdateComponentMessage{
		BaseMessage: BaseMessage{
			Type: "updateComponent",
		},
		Data: data,
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

// RemoveComponent removes a component from a slot
func (c *Client) RemoveComponent(componentID string) error {
	msg := &RemoveComponentMessage{
		BaseMessage: BaseMessage{
			Type: "removeComponent",
		},
		ComponentID: componentID,
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

// ListComponents lists all components in a slot (helper method)
func (c *Client) ListComponents(slotID string) ([]ComponentReference, error) {
	resp, err := c.GetSlot(slotID, true, 0)
	if err != nil {
		return nil, err
	}

	return resp.Data.Components, nil
}
