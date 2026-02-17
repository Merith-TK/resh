package tui

import (
	"github.com/Merith-TK/resh/pkg/resolink"
	"github.com/Merith-TK/resh/pkg/shell"
)

// FocusMode indicates which panel has focus
type FocusMode int

const (
	FocusTree FocusMode = iota
	FocusInspector
	FocusCommand
)

// TreeItem represents an item in the tree view
type TreeItem struct {
	ID         string
	Name       string
	IsSlot     bool       // true = slot, false = component
	Type       string     // Component type if IsSlot=false
	IsActive   bool       // For slots
	Children   []TreeItem // Cached children (lazy loaded)
	IsExpanded bool
	IsLoaded   bool // Have we loaded children yet?
}

// Model is the main Bubble Tea model
type Model struct {
	// Connection
	Client *resolink.Client
	State  *shell.State

	// Tree panel state
	TreeItems       []TreeItem // Current visible tree items
	TreeCursor      int        // Selected item index
	FocusedSlotID   string     // Current focused slot in tree
	FocusedSlotName string     // Name of focused slot
	PathBreadcrumb  []string   // Path breadcrumb for navigation

	// Inspector panel state
	InspectedItem  *TreeItem   // Currently inspected item
	InspectedData  interface{} // Full data from RESH.inspect()
	FieldCursor    int         // Which field is selected in inspector
	SubFieldCursor int         // Which subfield (x/y/z/w) is selected for composite types
	EditingField   bool        // Is user currently editing a field?
	EditBuffer     string      // Buffer for field editing
	EditFieldType  string      // Type of field being edited (bool, float, string, etc.)
	EditFieldName  string      // Name of field being edited
	EditSubField   string      // Subfield name (x/y/z/w) if editing composite type component

	// UI state
	Focus         FocusMode // Which panel has focus
	Width         int       // Terminal width
	Height        int       // Terminal height
	CommandBuffer string    // Command line input buffer
	StatusMessage string    // Status bar message
	ErrorMessage  string    // Error message to display

	// Settings
	UpdateOnNavigate bool // Update inspector as cursor moves (vs only on Enter)

	// Config
	HideRoot     bool // Hide root slot components
	HideRESHData bool // Hide RESH.DATA slot completely
}

// NewModel creates a new TUI model
func NewModel(client *resolink.Client, state *shell.State) Model {
	return Model{
		Client:           client,
		State:            state,
		Focus:            FocusTree,
		TreeItems:        []TreeItem{},
		TreeCursor:       0,
		FocusedSlotID:    state.RootSlotID,
		FocusedSlotName:  "Root",
		PathBreadcrumb:   []string{"Root"},
		UpdateOnNavigate: false, // Default: update only on Enter
		HideRoot:         true,  // Hide root components
		HideRESHData:     true,  // Hide RESH.DATA
		StatusMessage:    "Connected",
	}
}
