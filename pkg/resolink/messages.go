package resolink

import "encoding/json"

// Message types for ResoniteLink protocol
// Based on official C# implementation

// BaseMessage is the base for all messages
type BaseMessage struct {
	MessageID string `json:"messageId"`
	Type      string `json:"$type"` // Protocol uses $type for discriminator
}

// ============================================================================
// Slot Operations
// ============================================================================

// GetSlotMessage requests slot data
type GetSlotMessage struct {
	BaseMessage
	SlotID               string `json:"slotId"`
	IncludeComponentData bool   `json:"includeComponentData,omitempty"`
	Depth                int    `json:"depth,omitempty"`
}

// AddSlotMessage creates a new slot
type AddSlotMessage struct {
	BaseMessage
	Data *SlotDefinition `json:"data"`
}

// UpdateSlotMessage updates an existing slot
type UpdateSlotMessage struct {
	BaseMessage
	Data *SlotDefinition `json:"data"`
}

// RemoveSlotMessage deletes a slot
type RemoveSlotMessage struct {
	BaseMessage
	SlotID string `json:"slotId"`
}

// ============================================================================
// Component Operations
// ============================================================================

// GetComponentMessage requests component data
type GetComponentMessage struct {
	BaseMessage
	ComponentID string `json:"componentId"`
}

// AddComponentMessage creates a new component
type AddComponentMessage struct {
	BaseMessage
	ContainerSlotID string               `json:"containerSlotId"`
	Data            *ComponentDefinition `json:"data"`
}

// UpdateComponentMessage updates an existing component
type UpdateComponentMessage struct {
	BaseMessage
	Data *ComponentDefinition `json:"data"`
}

// RemoveComponentMessage deletes a component
type RemoveComponentMessage struct {
	BaseMessage
	ComponentID string `json:"componentId"`
}

// ============================================================================
// Reflection / Type Discovery
// ============================================================================

// GetComponentTypeListMessage requests list of all component types
type GetComponentTypeListMessage struct {
	BaseMessage
}

// GetComponentDefinitionMessage requests component type definition
type GetComponentDefinitionMessage struct {
	BaseMessage
	ComponentType string `json:"componentType"`
}

// GetTypeDefinitionMessage requests type information
type GetTypeDefinitionMessage struct {
	BaseMessage
	TypeName string `json:"typeName"`
}

// GetEnumDefinitionMessage requests enum definition
type GetEnumDefinitionMessage struct {
	BaseMessage
	EnumType string `json:"enumType"`
}

// ============================================================================
// Session Data
// ============================================================================

// RequestSessionDataMessage requests session information
type RequestSessionDataMessage struct {
	BaseMessage
}

// ============================================================================
// Batch Operations
// ============================================================================

// DataModelOperationBatchMessage executes multiple operations
type DataModelOperationBatchMessage struct {
	BaseMessage
	Operations []interface{} `json:"operations"`
}

// ============================================================================
// Data Definitions
// ============================================================================

// SlotDefinition represents a slot's data
type SlotDefinition struct {
	ID          string               `json:"id,omitempty"`
	Name        *ValueString         `json:"name,omitempty"`
	Active      *ValueBool           `json:"active,omitempty"`
	Persistent  *ValueBool           `json:"persistent,omitempty"`
	Position    *ValueFloat3         `json:"position,omitempty"`
	Rotation    *ValueFloatQ         `json:"rotation,omitempty"`
	Scale       *ValueFloat3         `json:"scale,omitempty"`
	OrderOffset *ValueLong           `json:"orderOffset,omitempty"`
	Tag         *ValueString         `json:"tag,omitempty"`
	Parent      *ValueReference      `json:"parent,omitempty"`
	Components  []ComponentReference `json:"components,omitempty"`
	Children    []SlotReference      `json:"children,omitempty"`
}

// ComponentDefinition represents a component's data
type ComponentDefinition struct {
	ID            string                 `json:"id,omitempty"`
	ComponentType string                 `json:"componentType,omitempty"`
	Members       map[string]interface{} `json:"members,omitempty"`
}

// ComponentReference is a reference to a component
type ComponentReference struct {
	ID            string `json:"id"`
	ComponentType string `json:"componentType"`
}

// SlotReference is a reference to a slot in responses (with full data structure)
type SlotReference struct {
	ID              string          `json:"id"`
	Parent          *ValueReference `json:"parent,omitempty"`
	Position        *ValueFloat3    `json:"position,omitempty"`
	Rotation        *ValueFloatQ    `json:"rotation,omitempty"`
	Scale           *ValueFloat3    `json:"scale,omitempty"`
	IsActive        *ValueBool      `json:"isActive,omitempty"`
	IsPersistent    *ValueBool      `json:"isPersistent,omitempty"`
	Name            *ValueString    `json:"name,omitempty"`
	Tag             *ValueString    `json:"tag,omitempty"`
	OrderOffset     *ValueLong      `json:"orderOffset,omitempty"`
	Components      json.RawMessage `json:"components,omitempty"` // Can be null or array
	Children        json.RawMessage `json:"children,omitempty"`   // Can be null or array
	IsReferenceOnly bool            `json:"isReferenceOnly"`
}

// ============================================================================
// Value Types (all have $type field)
// ============================================================================

// ValueString represents a string value
type ValueString struct {
	Type  string `json:"$type"`
	Value string `json:"value"`
}

// NewValueString creates a new string value
func NewValueString(value string) *ValueString {
	return &ValueString{Type: "string", Value: value}
}

// ValueBool represents a boolean value
type ValueBool struct {
	Type  string `json:"$type"`
	Value bool   `json:"value"`
}

// NewValueBool creates a new bool value
func NewValueBool(value bool) *ValueBool {
	return &ValueBool{Type: "bool", Value: value}
}

// ValueInt represents an integer value
type ValueInt struct {
	Type  string `json:"$type"`
	Value int    `json:"value"`
}

// NewValueInt creates a new int value
func NewValueInt(value int) *ValueInt {
	return &ValueInt{Type: "int", Value: value}
}

// ValueLong represents a long integer value
type ValueLong struct {
	Type  string `json:"$type"`
	Value int64  `json:"value"`
}

// NewValueLong creates a new long value
func NewValueLong(value int64) *ValueLong {
	return &ValueLong{Type: "long", Value: value}
}

// ValueFloat represents a float value
type ValueFloat struct {
	Type  string  `json:"$type"`
	Value float64 `json:"value"`
}

// NewValueFloat creates a new float value
func NewValueFloat(value float64) *ValueFloat {
	return &ValueFloat{Type: "float", Value: value}
}

// Float3 represents a 3D vector
type Float3 struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

// ValueFloat3 represents a float3 value
type ValueFloat3 struct {
	Type  string  `json:"$type"`
	Value *Float3 `json:"value"`
}

// NewValueFloat3 creates a new float3 value
func NewValueFloat3(x, y, z float64) *ValueFloat3 {
	return &ValueFloat3{
		Type:  "float3",
		Value: &Float3{X: x, Y: y, Z: z},
	}
}

// FloatQ represents a quaternion
type FloatQ struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
	W float64 `json:"w"`
}

// ValueFloatQ represents a floatQ value (quaternion)
type ValueFloatQ struct {
	Type  string  `json:"$type"`
	Value *FloatQ `json:"value"`
}

// NewValueFloatQ creates a new floatQ value
func NewValueFloatQ(x, y, z, w float64) *ValueFloatQ {
	return &ValueFloatQ{
		Type:  "floatQ",
		Value: &FloatQ{X: x, Y: y, Z: z, W: w},
	}
}

// ValueReference represents a reference to another object
type ValueReference struct {
	Type     string `json:"$type"`
	TargetID string `json:"targetId,omitempty"`
}

// NewValueReference creates a new reference value
func NewValueReference(targetID string) *ValueReference {
	return &ValueReference{Type: "reference", TargetID: targetID}
}

// ValueList represents a list of values
type ValueList struct {
	Type     string        `json:"$type"`
	Elements []interface{} `json:"elements"`
}

// NewValueList creates a new list value
func NewValueList(elements []interface{}) *ValueList {
	return &ValueList{Type: "list", Elements: elements}
}

// Color represents an RGBA color
type Color struct {
	R float64 `json:"r"`
	G float64 `json:"g"`
	B float64 `json:"b"`
	A float64 `json:"a"`
}

// ValueColor represents a color value
type ValueColor struct {
	Type  string `json:"$type"`
	Value *Color `json:"value"`
}

// NewValueColor creates a new color value
func NewValueColor(r, g, b, a float64) *ValueColor {
	return &ValueColor{
		Type:  "color",
		Value: &Color{R: r, G: g, B: b, A: a},
	}
}

// ============================================================================
// Response Types
// ============================================================================

// SlotDataResponse is the response for getSlot
type SlotDataResponse struct {
	SourceMessageID string          `json:"sourceMessageId"`
	MessageID       string          `json:"messageId"`
	Type            string          `json:"$type"`
	Depth           int             `json:"depth"`
	Data            *SlotDefinition `json:"data"`
	Success         bool            `json:"success"`
	ErrorInfo       string          `json:"errorInfo,omitempty"`
}

// ComponentDataResponse is the response for getComponent
type ComponentDataResponse struct {
	SourceMessageID string               `json:"sourceMessageId"`
	MessageID       string               `json:"messageId"`
	Type            string               `json:"$type"`
	Data            *ComponentDefinition `json:"data"`
	Success         bool                 `json:"success"`
	ErrorInfo       string               `json:"errorInfo,omitempty"`
}

// ComponentTypeListResponse is the response for getComponentTypeList
type ComponentTypeListResponse struct {
	MessageID string   `json:"messageId"`
	Type      string   `json:"$type"`
	Types     []string `json:"types"`
}

// ErrorResponse represents an error from Resonite
type ErrorResponse struct {
	MessageID string `json:"messageId"`
	Type      string `json:"$type"`
	Error     string `json:"error"`
}
