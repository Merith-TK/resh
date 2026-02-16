package objects

// Component represents a Resonite component (analogous to a file)
type Component struct {
	ID      string
	SlotID  string
	Type    string // e.g., "FrooxEngine.BoxMesh"
	Members map[string]*Member
}

// Member represents a component member/property
type Member struct {
	Type  string      // "primitive", "reference", "list"
	Value interface{} // Actual value
}

// NewComponent creates a new component
func NewComponent(id, slotID, compType string) *Component {
	return &Component{
		ID:      id,
		SlotID:  slotID,
		Type:    compType,
		Members: make(map[string]*Member),
	}
}

// GetName returns the simple component name (without namespace)
func (c *Component) GetName() string {
	// Extract name from type string
	// e.g., "FrooxEngine.BoxMesh" -> "BoxMesh"
	for i := len(c.Type) - 1; i >= 0; i-- {
		if c.Type[i] == '.' {
			return c.Type[i+1:]
		}
	}
	return c.Type
}
