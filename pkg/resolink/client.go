package resolink

import (
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
	pendingRequests map[string]chan *Response
	mu              sync.RWMutex
	closed          bool
}

// Message represents a ResoLink protocol message
type Message struct {
	MessageID string                 `json:"messageId"`
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// Response represents a ResoLink response
type Response struct {
	MessageID string                 `json:"messageId"`
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Error     string                 `json:"error,omitempty"`
}

// NewClient creates a new ResoLink client
func NewClient(url string, timeout time.Duration) *Client {
	return &Client{
		url:             url,
		timeout:         timeout,
		pendingRequests: make(map[string]chan *Response),
	}
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
	c.pendingRequests = make(map[string]chan *Response)

	if c.conn != nil {
		return c.conn.Close()
	}

	return nil
}

// SendRequest sends a request and waits for response
func (c *Client) SendRequest(msgType string, data map[string]interface{}) (*Response, error) {
	if c.closed {
		return nil, fmt.Errorf("client is closed")
	}

	// Generate unique message ID
	msgID := uuid.New().String()

	// Create response channel
	respChan := make(chan *Response, 1)
	c.mu.Lock()
	c.pendingRequests[msgID] = respChan
	c.mu.Unlock()

	// Clean up on return
	defer func() {
		c.mu.Lock()
		delete(c.pendingRequests, msgID)
		c.mu.Unlock()
	}()

	// Construct message
	msg := Message{
		MessageID: msgID,
		Type:      msgType,
		Data:      data,
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
		if resp.Error != "" {
			return nil, fmt.Errorf("server error: %s", resp.Error)
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
		var resp Response
		err := c.conn.ReadJSON(&resp)
		if err != nil {
			// Connection closed
			return
		}

		// Route response to waiting request
		c.mu.RLock()
		respChan, exists := c.pendingRequests[resp.MessageID]
		c.mu.RUnlock()

		if exists {
			select {
			case respChan <- &resp:
			default:
				// Channel full or closed
			}
		}
	}
}

// Helper method to convert response data to specific types
func (r *Response) GetString(key string) (string, error) {
	val, ok := r.Data[key]
	if !ok {
		return "", fmt.Errorf("key %s not found", key)
	}
	str, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("key %s is not a string", key)
	}
	return str, nil
}

func (r *Response) GetMap(key string) (map[string]interface{}, error) {
	val, ok := r.Data[key]
	if !ok {
		return nil, fmt.Errorf("key %s not found", key)
	}
	m, ok := val.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("key %s is not a map", key)
	}
	return m, nil
}
