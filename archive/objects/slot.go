package objects

// Slot represents a Resonite slot (analogous to a directory)
type Slot struct {
	ID         string
	Name       string
	ParentID   string
	Position   [3]float64
	Rotation   [4]float64
	Scale      [3]float64
	Active     bool
	Components []*Component
	Children   []string // Child slot IDs
}

// NewSlot creates a new slot
func NewSlot(id, name, parentID string) *Slot {
	return &Slot{
		ID:         id,
		Name:       name,
		ParentID:   parentID,
		Position:   [3]float64{0, 0, 0},
		Rotation:   [4]float64{0, 0, 0, 1},
		Scale:      [3]float64{1, 1, 1},
		Active:     true,
		Components: []*Component{},
		Children:   []string{},
	}
}

// IsRoot returns true if this is the root slot
func (s *Slot) IsRoot() bool {
	return s.ParentID == ""
}
