package resolink

import (
	"encoding/json"
	"fmt"
)

// GetComponentTypeList retrieves list of all available component types
func (c *Client) GetComponentTypeList() ([]string, error) {
	msg := &GetComponentTypeListMessage{
		BaseMessage: BaseMessage{
			Type: "getComponentTypeList",
		},
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
	var resp ComponentTypeListResponse
	if err := json.Unmarshal(rawResp, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return resp.Types, nil
}

// GetComponentDefinition retrieves detailed information about a component type
func (c *Client) GetComponentDefinition(componentType string) (json.RawMessage, error) {
	msg := &GetComponentDefinitionMessage{
		BaseMessage: BaseMessage{
			Type: "getComponentDefinition",
		},
		ComponentType: componentType,
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

	// Return raw response for now - definition structure is complex
	// Can be parsed later as needed
	return rawResp, nil
}

// GetTypeDefinition retrieves information about a type
func (c *Client) GetTypeDefinition(typeName string) (json.RawMessage, error) {
	msg := &GetTypeDefinitionMessage{
		BaseMessage: BaseMessage{
			Type: "getTypeDefinition",
		},
		TypeName: typeName,
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

	return rawResp, nil
}

// GetEnumDefinition retrieves enum definition
func (c *Client) GetEnumDefinition(enumType string) (json.RawMessage, error) {
	msg := &GetEnumDefinitionMessage{
		BaseMessage: BaseMessage{
			Type: "getEnumDefinition",
		},
		EnumType: enumType,
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

	return rawResp, nil
}
