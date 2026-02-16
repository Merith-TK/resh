package resh

import (
	"encoding/json"
	"fmt"

	"github.com/Merith-TK/resonite-sh/pkg/resolink"
)

const (
	RESHSlotName    = "RESH"
	OrderOffset     = -999 // Special offset to make RESH appear last
	VarsSlotName    = "VARS"
	SessionSlotName = "SESSION"
	LocalSlotName   = "LOCAL"
	WorldSlotName   = "WORLD"
)

// Manager handles RESH slot initialization and variable management
type Manager struct {
	client        *resolink.Client
	reshSlotID    string // RefID of the RESH slot
	varsSlotID    string // RefID of the VARS slot
	sessionSlotID string // RefID of the SESSION slot
	localSlotID   string // RefID of the LOCAL slot
	worldSlotID   string // RefID of the WORLD slot
}

// NewManager creates a new RESH manager
func NewManager(client *resolink.Client) *Manager {
	return &Manager{
		client: client,
	}
}

// Initialize finds or creates the RESH slot structure:
// Root
// └── RESH (OrderOffset: 999)
//
//	├── DynamicVariableSpace (SpaceName: "RESH")
//	└── VARS
//	    ├── SESSION
//	    ├── LOCAL
//	    └── WORLD
func (m *Manager) Initialize() error {
	// First, try to find existing RESH slot
	reshSlot, err := m.client.FindSlotByName("Root", RESHSlotName)
	if err == nil && reshSlot != nil {
		// RESH exists, extract RefID
		var slotData resolink.SlotDataResponse
		if err := json.Unmarshal(reshSlot, &slotData); err != nil {
			return fmt.Errorf("failed to parse RESH slot data: %w", err)
		}
		m.reshSlotID = slotData.Data.ID

		// Verify RESH structure
		if err := m.verifyRESHStructure(); err != nil {
			return fmt.Errorf("RESH slot exists but structure is invalid: %w", err)
		}

		return nil
	}

	// RESH doesn't exist, create it
	return m.createRESHStructure()
}

// verifyRESHStructure checks that RESH has the correct components and sub-slots
func (m *Manager) verifyRESHStructure() error {
	// Get RESH slot with components
	reshData, err := m.client.GetSlot(m.reshSlotID, true, 0)
	if err != nil {
		return fmt.Errorf("failed to get RESH slot: %w", err)
	}

	var slotResp resolink.SlotDataResponse
	if err := json.Unmarshal(reshData, &slotResp); err != nil {
		return fmt.Errorf("failed to parse RESH slot: %w", err)
	}

	// Check for DynamicVariableSpace component
	hasDynamicVarSpace := false
	for _, comp := range slotResp.Data.Components {
		if comp.Type == "FrooxEngine.DynamicVariableSpace" {
			hasDynamicVarSpace = true
			break
		}
	}

	if !hasDynamicVarSpace {
		return fmt.Errorf("RESH slot missing DynamicVariableSpace component")
	}

	// Find VARS sub-slot
	varsSlot, err := m.client.FindSlotByName(m.reshSlotID, VarsSlotName)
	if err != nil {
		return fmt.Errorf("VARS slot not found: %w", err)
	}

	var varsResp resolink.SlotDataResponse
	if err := json.Unmarshal(varsSlot, &varsResp); err != nil {
		return fmt.Errorf("failed to parse VARS slot: %w", err)
	}
	m.varsSlotID = varsResp.Data.ID

	// Find SESSION, LOCAL, WORLD slots
	sessionSlot, err := m.client.FindSlotByName(m.varsSlotID, SessionSlotName)
	if err != nil {
		return fmt.Errorf("SESSION slot not found: %w", err)
	}
	var sessionResp resolink.SlotDataResponse
	if err := json.Unmarshal(sessionSlot, &sessionResp); err != nil {
		return fmt.Errorf("failed to parse SESSION slot: %w", err)
	}
	m.sessionSlotID = sessionResp.Data.ID

	localSlot, err := m.client.FindSlotByName(m.varsSlotID, LocalSlotName)
	if err != nil {
		return fmt.Errorf("LOCAL slot not found: %w", err)
	}
	var localResp resolink.SlotDataResponse
	if err := json.Unmarshal(localSlot, &localResp); err != nil {
		return fmt.Errorf("failed to parse LOCAL slot: %w", err)
	}
	m.localSlotID = localResp.Data.ID

	worldSlot, err := m.client.FindSlotByName(m.varsSlotID, WorldSlotName)
	if err != nil {
		return fmt.Errorf("WORLD slot not found: %w", err)
	}
	var worldResp resolink.SlotDataResponse
	if err := json.Unmarshal(worldSlot, &worldResp); err != nil {
		return fmt.Errorf("failed to parse WORLD slot: %w", err)
	}
	m.worldSlotID = worldResp.Data.ID

	return nil
}

// createRESHStructure creates the full RESH slot hierarchy
func (m *Manager) createRESHStructure() error {
	// Create RESH slot under Root
	reshSlotDef := &resolink.SlotDefinition{
		ID:          RESHSlotName, // This will become the persistent ID
		Name:        resolink.NewValueString(RESHSlotName),
		ParentID:    "Root",
		Active:      resolink.NewValueBool(true),
		OrderOffset: resolink.NewValueInt(OrderOffset),
	}

	reshResp, err := m.client.AddSlot(reshSlotDef)
	if err != nil {
		return fmt.Errorf("failed to create RESH slot: %w", err)
	}

	var reshData resolink.SlotDataResponse
	if err := json.Unmarshal(reshResp, &reshData); err != nil {
		return fmt.Errorf("failed to parse RESH response: %w", err)
	}
	m.reshSlotID = reshData.Data.ID

	// Add DynamicVariableSpace component
	dynVarSpaceDef := &resolink.ComponentDefinition{
		Type:   "FrooxEngine.DynamicVariableSpace",
		SlotID: m.reshSlotID,
		Fields: map[string]interface{}{
			"SpaceName": resolink.NewValueString(RESHSlotName),
		},
	}

	if _, err := m.client.AddComponent(m.reshSlotID, dynVarSpaceDef); err != nil {
		return fmt.Errorf("failed to add DynamicVariableSpace: %w", err)
	}

	// Create VARS slot
	varsSlotDef := &resolink.SlotDefinition{
		ID:       VarsSlotName,
		Name:     resolink.NewValueString(VarsSlotName),
		ParentID: m.reshSlotID,
		Active:   resolink.NewValueBool(true),
	}

	varsResp, err := m.client.AddSlot(varsSlotDef)
	if err != nil {
		return fmt.Errorf("failed to create VARS slot: %w", err)
	}

	var varsData resolink.SlotDataResponse
	if err := json.Unmarshal(varsResp, &varsData); err != nil {
		return fmt.Errorf("failed to parse VARS response: %w", err)
	}
	m.varsSlotID = varsData.Data.ID

	// Create SESSION slot
	sessionSlotDef := &resolink.SlotDefinition{
		ID:       SessionSlotName,
		Name:     resolink.NewValueString(SessionSlotName),
		ParentID: m.varsSlotID,
		Active:   resolink.NewValueBool(true),
	}

	sessionResp, err := m.client.AddSlot(sessionSlotDef)
	if err != nil {
		return fmt.Errorf("failed to create SESSION slot: %w", err)
	}

	var sessionData resolink.SlotDataResponse
	if err := json.Unmarshal(sessionResp, &sessionData); err != nil {
		return fmt.Errorf("failed to parse SESSION response: %w", err)
	}
	m.sessionSlotID = sessionData.Data.ID

	// Create LOCAL slot
	localSlotDef := &resolink.SlotDefinition{
		ID:       LocalSlotName,
		Name:     resolink.NewValueString(LocalSlotName),
		ParentID: m.varsSlotID,
		Active:   resolink.NewValueBool(true),
	}

	localResp, err := m.client.AddSlot(localSlotDef)
	if err != nil {
		return fmt.Errorf("failed to create LOCAL slot: %w", err)
	}

	var localData resolink.SlotDataResponse
	if err := json.Unmarshal(localResp, &localData); err != nil {
		return fmt.Errorf("failed to parse LOCAL response: %w", err)
	}
	m.localSlotID = localData.Data.ID

	// Create WORLD slot
	worldSlotDef := &resolink.SlotDefinition{
		ID:       WorldSlotName,
		Name:     resolink.NewValueString(WorldSlotName),
		ParentID: m.varsSlotID,
		Active:   resolink.NewValueBool(true),
	}

	worldResp, err := m.client.AddSlot(worldSlotDef)
	if err != nil {
		return fmt.Errorf("failed to create WORLD slot: %w", err)
	}

	var worldData resolink.SlotDataResponse
	if err := json.Unmarshal(worldResp, &worldData); err != nil {
		return fmt.Errorf("failed to parse WORLD response: %w", err)
	}
	m.worldSlotID = worldData.Data.ID

	return nil
}

// GetRESHSlotID returns the RESH slot RefID
func (m *Manager) GetRESHSlotID() string {
	return m.reshSlotID
}

// GetVarsSlotID returns the VARS slot RefID
func (m *Manager) GetVarsSlotID() string {
	return m.varsSlotID
}

// GetSessionSlotID returns the SESSION slot RefID
func (m *Manager) GetSessionSlotID() string {
	return m.sessionSlotID
}

// GetLocalSlotID returns the LOCAL slot RefID
func (m *Manager) GetLocalSlotID() string {
	return m.localSlotID
}

// GetWorldSlotID returns the WORLD slot RefID
func (m *Manager) GetWorldSlotID() string {
	return m.worldSlotID
}
