package resolink

import "fmt"

// AddComponent adds a component to a slot
func (c *Client) AddComponent(slotID, componentType string) (string, error) {
	data := map[string]interface{}{
		"slotId": slotID,
		"type":   componentType,
	}

	resp, err := c.SendRequest("addComponent", data)
	if err != nil {
		return "", err
	}

	componentID, err := resp.GetString("componentId")
	if err != nil {
		return "", fmt.Errorf("invalid response: %w", err)
	}

	return componentID, nil
}

// GetComponent retrieves component information
func (c *Client) GetComponent(componentID string) (*Response, error) {
	data := map[string]interface{}{
		"componentId": componentID,
	}

	return c.SendRequest("getComponent", data)
}

// UpdateComponent updates component members
func (c *Client) UpdateComponent(componentID string, members map[string]interface{}) error {
	data := map[string]interface{}{
		"componentId": componentID,
		"members":     members,
	}

	_, err := c.SendRequest("updateComponent", data)
	return err
}

// RemoveComponent removes a component from a slot
func (c *Client) RemoveComponent(componentID string) error {
	data := map[string]interface{}{
		"componentId": componentID,
	}

	_, err := c.SendRequest("removeComponent", data)
	return err
}

// ListComponents lists all components in a slot
func (c *Client) ListComponents(slotID string) (*Response, error) {
	data := map[string]interface{}{
		"slotId": slotID,
	}

	return c.SendRequest("listComponents", data)
}
