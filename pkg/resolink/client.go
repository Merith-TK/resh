package resolink

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Client represents a ResoLink WebSocket client
type Client struct {
	url             string
	conn            *websocket.Conn
	timeout         time.Duration
	pendingRequests map[string]chan json.RawMessage
	mu              sync.RWMutex
	closed          bool
	reconnect       bool
	reconnectDelay  time.Duration
}

// RawResponse is the minimal response structure for routing
type RawResponse struct {
	MessageID string          `json:"messageId"`
	Type      string          `json:"$type"`
	RawData   json.RawMessage `json:"-"`
}

// NewClient creates a new ResoLink client
func NewClient(url string, timeout time.Duration) *Client {
	return &Client{
		url:             url,
		timeout:         timeout,
		pendingRequests: make(map[string]chan json.RawMessage),
		reconnect:       true,
		reconnectDelay:  5 * time.Second,
	}
}

// SetReconnect enables or disables automatic reconnection
func (c *Client) SetReconnect(enabled bool, delay time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.reconnect = enabled
	c.reconnectDelay = delay
}

// Connect establishes WebSocket connection to Resonite
func (c *Client) Connect() error {
	conn, _, err := websocket.DefaultDialer.Dial(c.url, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", c.url, err)
	}

	c.conn = conn
	c.closed = false

	// Start response handler
	go c.handleResponses()

	return nil
}

// Disconnect closes the WebSocket connection
func (c *Client) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true

	// Close all pending requests
	for _, ch := range c.pendingRequests {
		close(ch)
	}
	c.pendingRequests = make(map[string]chan json.RawMessage)

	if c.conn != nil {
		return c.conn.Close()
	}

	return nil
}

// sendMessage sends a message struct and waits for raw response
func (c *Client) sendMessage(msg interface{}) (json.RawMessage, error) {
	if c.closed {
		return nil, fmt.Errorf("client is closed")
	}

	// Extract or generate message ID
	msgID := uuid.New().String()

	// Create response channel
	respChan := make(chan json.RawMessage, 1)
	c.mu.Lock()
	c.pendingRequests[msgID] = respChan
	c.mu.Unlock()

	// Clean up on return
	defer func() {
		c.mu.Lock()
		delete(c.pendingRequests, msgID)
		c.mu.Unlock()
	}()

	// Set message ID on the struct
	switch m := msg.(type) {
	case *GetSlotMessage:
		m.MessageID = msgID
	case *AddSlotMessage:
		m.MessageID = msgID
	case *UpdateSlotMessage:
		m.MessageID = msgID
	case *RemoveSlotMessage:
		m.MessageID = msgID
	case *GetComponentMessage:
		m.MessageID = msgID
	case *AddComponentMessage:
		m.MessageID = msgID
	case *UpdateComponentMessage:
		m.MessageID = msgID
	case *RemoveComponentMessage:
		m.MessageID = msgID
	case *GetComponentTypeListMessage:
		m.MessageID = msgID
	case *GetComponentDefinitionMessage:
		m.MessageID = msgID
	default:
		return nil, fmt.Errorf("unsupported message type")
	}

	// Send message
	if err := c.conn.WriteJSON(msg); err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	// Wait for response with timeout
	select {
	case resp := <-respChan:
		if resp == nil {
			return nil, fmt.Errorf("connection closed")
		}
		return resp, nil
	case <-time.After(c.timeout):
		return nil, fmt.Errorf("request timeout after %v", c.timeout)
	}
}

// handleResponses processes incoming WebSocket messages
func (c *Client) handleResponses() {
	defer func() {
		c.Disconnect()
	}()

	for {
		// Read raw message
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			// Connection closed
			return
		}

		// Parse minimal response to get messageId
		var rawResp RawResponse
		if err := json.Unmarshal(message, &rawResp); err != nil {
			continue // Skip malformed messages
		}

		// Store the raw data for later parsing
		rawResp.RawData = message

		// Route response to waiting request
		c.mu.RLock()
		respChan, exists := c.pendingRequests[rawResp.MessageID]
		c.mu.RUnlock()

		if exists {
			select {
			case respChan <- message:
			default:
				// Channel full or closed
			}
		}
	}
}

// IsConnected returns whether the client is connected
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn != nil && !c.closed
}
