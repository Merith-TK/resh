package shell

import "github.com/Merith-TK/resonite-sh/pkg/resolink"

// State represents the current session state
type State struct {
	CurrentSlot string
	CurrentPath string
	RootSlotID  string
	RESHDataID  string            // ID of /RESH.DATA slot for variable storage
	Variables   map[string]string // Bookmark variables (name -> slotID)
}

// SlotListing contains the contents of a slot
type SlotListing struct {
	SlotID     string
	SlotName   string
	IsRootSlot bool
	Children   []SlotInfo
	Components []ComponentInfo
}

// SlotInfo represents a child slot with relevant display information
type SlotInfo struct {
	ID           string
	Name         string
	IsActive     bool
	IsPersistent bool
}

// ComponentInfo represents a component with relevant display information
type ComponentInfo struct {
	ID           string
	Type         string
	IsPersistent bool
}

// ConnectionInfo contains information about the connection test
type ConnectionInfo struct {
	SlotID   string
	SlotName string
}

// NewStateFromResoLink creates a State from a ResoLink slot response
func NewStateFromResoLink(slotResp *resolink.SlotDataResponse) *State {
	return &State{
		CurrentSlot: "Root",
		CurrentPath: "/",
		RootSlotID:  slotResp.Data.ID,
	}
}

// SlotInfoFromResoLink converts a ResoLink SlotReference to SlotInfo
func SlotInfoFromResoLink(slot *resolink.SlotReference) SlotInfo {
	info := SlotInfo{
		ID:           slot.ID,
		Name:         "<unnamed>",
		IsActive:     true,
		IsPersistent: true,
	}

	if slot.Name != nil {
		info.Name = slot.Name.Value
	}
	if slot.IsActive != nil {
		info.IsActive = slot.IsActive.Value
	}
	if slot.IsPersistent != nil {
		info.IsPersistent = slot.IsPersistent.Value
	}

	return info
}

// ComponentInfoFromResoLink converts a ResoLink ComponentReference to ComponentInfo
func ComponentInfoFromResoLink(comp *resolink.ComponentReference) ComponentInfo {
	return ComponentInfo{
		ID:           comp.ID,
		Type:         comp.ComponentType,
		IsPersistent: true, // Components are persistent by default in Resonite
	}
}
