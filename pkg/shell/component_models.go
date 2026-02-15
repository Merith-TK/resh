package shell

// ComponentData represents full component information
type ComponentData struct {
	ID            string
	ComponentType string
	TypeName      string // Parsed short name (after last .)
	Members       []MemberData
}

// MemberData represents a component member/property
type MemberData struct {
	Name  string
	ID    string
	Type  string      // $type field
	Value interface{} // Actual value
}
