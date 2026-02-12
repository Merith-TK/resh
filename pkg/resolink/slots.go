package resolink

import "fmt"

// AddSlot creates a new slot under the specified parent
func (c *Client) AddSlot(parentID, name string) (string, error) {
	data := map[string]interface{}{
		"parentId": parentID,
		"name":     name,
	}

	resp, err := c.SendRequest("addSlot", data)
	if err != nil {
		return "", err
	}

	slotID, err := resp.GetString("slotId")
	if err != nil {
		return "", fmt.Errorf("invalid response: %w", err)
	}

	return slotID, nil
}

// GetSlot retrieves slot information
func (c *Client) GetSlot(slotID string, includeComponents bool) (*Response, error) {
	data := map[string]interface{}{
		"slotId": slotID,
	}

	if includeComponents {
		data["includeComponentData"] = true
	}

	return c.SendRequest("getSlot", data)
}

// UpdateSlot updates slot properties
func (c *Client) UpdateSlot(slotID string, properties map[string]interface{}) error {
	data := map[string]interface{}{
		"slotId": slotID,
	}

	// Merge properties into data
	for k, v := range properties {
		data[k] = v
	}

	_, err := c.SendRequest("updateSlot", data)
	return err
}

// RemoveSlot deletes a slot
func (c *Client) RemoveSlot(slotID string) error {
	data := map[string]interface{}{
		"slotId": slotID,
	}

	_, err := c.SendRequest("removeSlot", data)
	return err
}

// FindSlotByName searches for a slot by name
func (c *Client) FindSlotByName(parentID, name string) (*Response, error) {
	data := map[string]interface{}{
		"parentId": parentID,
		"name":     name,
	}

	return c.SendRequest("findSlotByName", data)
}

// ListChildren lists all child slots of a parent
func (c *Client) ListChildren(parentID string) (*Response, error) {
	data := map[string]interface{}{
		"slotId": parentID,
	}

	return c.SendRequest("listChildren", data)
}
