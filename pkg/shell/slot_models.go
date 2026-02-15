package shell

// SlotData represents detailed slot information for inspection
type SlotData struct {
	ID         string
	Properties []SlotProperty
}

// SlotProperty represents a single slot property
type SlotProperty struct {
	Name  string
	Type  string      // "string", "bool", "float3", "floatQ", "reference", "long"
	Value interface{} // Actual value
}
