# Design Decisions & Requirements

Based on discussions, here are the finalized design decisions for resonite-sh.

## Core Concept

A dual-mode interface (TUI + REPL) for Resonite that treats the scene graph as a filesystem, with full Lua scripting support and persistent variable storage in the world itself.

## 1. UI Architecture

### Dual Mode System
- **TUI Mode**: Bubble Tea-based visual interface (Midnight Commander style)
- **REPL Mode**: Shell-like command interface
- **Switching**: Press `Tab` to toggle between modes
- Both modes are equally powerful, just different UX

### TUI Layout
```
┌─────────────────────┬──────────────────────────────────┐
│ Slot Tree           │ Inspector Panel                  │
│ /World              │ Selected: BoxMesh (comp-abc123)  │
│ ├─ Environment      │ ────────────────────────────────│
│ │  ├─ Box1 (s-123)  │ Type: FrooxEngine.BoxMesh       │
│ │  └─ Box2 (s-124)  │ Size: [2.0, 2.0, 2.0] ←         │
│ └─ UI               │ Persistent: true                 │
│ └─ RESH (s-999)     │ Enabled: true                    │
├─────────────────────┼──────────────────────────────────┤
│ Command: :lua test.lua | Status: Connected            │
└─────────────────────┴──────────────────────────────────┘
```

**Key Features**:
- RefIDs shown in parentheses: `SlotName (s-refid)`, `ComponentName (c-refid)`
- Arrow keys for navigation
- Mouse support in Inspector panel for clicking/editing
- Tree panel shows hierarchical structure
- Inspector shows selected slot/component properties

### REPL Mode
```
resonite:/World/Environment> ls
Box1 (s-123)/
Box2 (s-124)/
BoxMesh (c-abc)
MeshRenderer (c-def)

resonite:/World/Environment> edit Box1
[Opens editor for Box1 slot properties]

resonite:/World/Environment> cd Box1
resonite:/World/Environment/Box1> cat BoxMesh
Type: FrooxEngine.BoxMesh
Size: [2.0, 2.0, 2.0]
...
```

## 2. Filesystem Metaphor

### Structure
- **Slots** = Directories/Folders
- **Components** = Files
- **Editing a Slot** = Edit slot properties (Name, Position, Rotation, Scale, etc.)
- **Editing a Component** = Edit component members

### REPL Commands
- Navigation: `cd`, `pwd`, `ls`, `tree`
- Inspection: `cat`, `stat`, `inspect`
- Modification: `edit`, `set`, `mv`, `cp`, `rm`, `mkdir`, `touch`
- Search: `find`, `grep`
- Scripting: `:lua <path>` or `:lua VAR.<name>`

### TUI Navigation
- Arrow keys to navigate tree
- Mouse to click and edit in inspector
- Search/filter available
- Breadcrumb path at top

## 3. Variable Storage System (RESH Slot)

### RESH Slot Structure
On connection, create a persistent slot in the world:

```
Slot: RESH
├─ Parent: Root
├─ Name: "RESH"
├─ Persistent: true
├─ Position: [0, 0, 0]
├─ Rotation: [0, 0, 0, 1]
├─ Scale: [1, 1, 1]
├─ OrderOffset: 999
│
├─ Component: DynamicReferenceVariable<Slot>
│  ├─ VariableName: "World/RESH.DATA"
│  └─ Reference: THIS.SLOT
│
└─ Component: DynamicVariableSpace
   ├─ SpaceName: "RESH"
   └─ OnlyDirectBinding: true
```

### Variable Types
All Resonite dynamic variable types supported:
- `DynamicValueVariable<T>` - Primitives (int, float, bool, string, float3, floatQ, etc.)
- `DynamicObjectVariable<T>` - Object references
- `DynamicReferenceVariable<T>` - Component/Slot references

### Variable Access in Lua
```lua
-- Get variable
local myValue = RESH.VARS.MyVariable

-- Set variable
RESH.VARS.MyVariable = newValue

-- Create new variable
RESH.VARS.NewVar = {type = "float", value = 3.14}

-- Store script
RESH.VARS.MyScript = [[
  -- Lua code here
]]
```

### Variable Scopes
1. **Session Variables**: In-memory only, lost on exit (Lua tables)
2. **Local Variables**: Stored in local config file `~/.resonite-sh/vars.yaml`
3. **World Variables**: Stored in RESH slot (persistent in Resonite)

## 4. Lua Scripting

### Integration
- **Engine**: gopher-lua (pure Go Lua VM)
- **Version**: Lua 5.1 compatible
- **Access**: Full API access to ResoLink, VFS, variables

### Lua API Surface
```lua
-- Navigation
RESH.cd("/World/Environment")
RESH.pwd()
local items = RESH.ls("/World")

-- Slot operations
local slotId = RESH.createSlot("MySlot", parentId)
RESH.removeSlot(slotId)
local slot = RESH.getSlot(slotId)
RESH.updateSlot(slotId, {Name = "NewName", Position = {0, 1, 0}})

-- Component operations
local compId = RESH.addComponent(slotId, "FrooxEngine.BoxMesh")
RESH.removeComponent(compId)
local comp = RESH.getComponent(compId)
RESH.updateComponent(compId, {Size = {2, 2, 2}})

-- Variables
RESH.VARS.MyVar = 123
print(RESH.VARS.MyVar)

-- UI control
RESH.refresh()
RESH.focus(slotId)
RESH.message("Hello, World!")

-- Utilities
RESH.find("*Box*")
RESH.filter(function(slot) return slot.Active end)
```

### Script Execution
1. **File-based**: `:lua scripts/myScript.lua` (relative to working directory)
2. **Variable-based**: `:lua VAR.StoredScript` (from RESH variables)
3. **Inline**: `:lua print("Hello")` (one-liners)
4. **Auto-run**: Scripts can be marked as autorun on connect

### Script Storage
Scripts can be stored as:
- Local files: `scripts/*.lua`
- World variables: `RESH.VARS.ScriptName = "...code..."`
- Config presets: `~/.resonite-sh/scripts/*.lua`

### Component Type Discovery

ResoniteLink provides reflection APIs for component discovery:

```lua
-- Get list of all component types
local types = RESH.getComponentTypes()

-- Get detailed component definition (members, etc.)
local def = RESH.getComponentDefinition("FrooxEngine.BoxMesh")

-- User can define shortcuts
RESH.ALIASES.Box = "FrooxEngine.BoxMesh"
RESH.ALIASES.Grab = "FrooxEngine.Grabbable"

-- Use shortcuts in commands
touch Box  -- Creates FrooxEngine.BoxMesh
```

**Component Type Cache**:
- Cache component list on first connect
- Allow user-defined aliases (stored in local config)
- Fuzzy matching for partial names
- Show suggestions when ambiguous

## 5. Inspector Panel Details

### Slot Properties (when slot selected)
```
┌─ Slot: Box1 (s-123) ─────────────┐
│ Name:        Box1                 │
│ Parent:      Environment (s-100)  │← Click to jump
│ Active:      ☑                    │← Toggle
│ Persistent:  ☑                    │← Toggle
│ Position:    [0.0, 1.0, 0.0]      │← Edit inline
│ Rotation:    [0.0, 0.0, 0.0, 1.0] │← Edit inline
│ Scale:       [1.0, 1.0, 1.0]      │← Edit inline
│ OrderOffset: 0                    │← Edit inline
│ Tag:         myTag                │← Edit inline
│                                   │
│ Components (3):                   │
│  ├─ BoxMesh (c-abc)              │← Click to view
│  ├─ MeshRenderer (c-def)         │← Click to view
│  └─ PBS_Metallic (c-ghi)         │← Click to view
│                                   │
│ Children (2):                     │
│  ├─ ChildSlot1 (s-456)           │← Click to jump
│  └─ ChildSlot2 (s-789)           │← Click to jump
└───────────────────────────────────┘
```

### Component Properties (when component selected)
```
┌─ Component: BoxMesh (c-abc) ──────┐
│ Type: FrooxEngine.BoxMesh         │
│ Slot: Box1 (s-123)               │← Click to jump
│                                   │
│ Persistent:  ☑                    │
│ Enabled:     ☑                    │
│                                   │
│ Size:        [2.0, 2.0, 2.0]     │← Edit
│                                   │
│ [Driven by Transform (c-xyz)]    │← Shows drivers
│                                   │
│ [Apply] [Reset] [Remove]          │
└───────────────────────────────────┘
```

### Property Editing
- **Inline Edit**: Arrow down to property, press Enter, type value, press Enter
- **Boolean Toggle**: Space bar
- **References**: Click to jump, or type RefID
- **Lists**: Expand/collapse with arrow, add/remove items
- **Validation**: Type checking, range validation, format validation

## 6. Navigation & Controls

### TUI Mode Controls
- **Arrow Keys**: Navigate tree/inspector
- **Tab**: Switch to REPL mode
- **Enter**: Select/edit item
- **Space**: Toggle boolean
- **Esc**: Cancel edit/go back
- **Mouse Click**: Select item, focus field
- **Ctrl+S**: Save changes
- **Ctrl+R**: Refresh from Resonite
- **Ctrl+F**: Search/filter
- **/**: Start search
- **:**: Command mode (for `:lua`, etc.)

### REPL Mode Controls
- **Tab**: Switch to TUI mode
- **Arrow Up/Down**: Command history
- **Ctrl+R**: Reverse search history
- **Tab**: Auto-complete paths/commands

## 7. Performance & Loading

### Lazy Loading Strategy
1. Load root slot on connect
2. Load children only when expanded in tree or `cd` into
3. Load components only when inspecting slot
4. Cache loaded data with TTL
5. Background prefetch for visible nodes

### Virtual Scrolling
- Only render visible tree nodes (10-50 at a time)
- Smooth scrolling with overflow indicators
- Jump to search results

### Update Mechanism
1. **If ResoLink supports push updates**: Subscribe to changes on focused slot
2. **Otherwise**: Poll every 0.1s for focused slot/component
3. **Manual refresh**: Ctrl+R or `refresh` command
4. **Smart polling**: Slow down when idle, speed up during editing

## 8. Data Flow

### Connection Flow
```
Start App
  ↓
Connect to ResoLink (ws://localhost:29551)
  ↓
Check for RESH slot
  ↓
Create RESH slot if missing
  ↓
Set up DynamicVariableSpace
  ↓
Load session variables from local config
  ↓
Initialize TUI or REPL
  ↓
Ready for user interaction
```

### Edit Flow (TUI)
```
Select slot in tree
  ↓
Load slot details + components
  ↓
Display in inspector
  ↓
User edits property
  ↓
Validate input
  ↓
Apply to ResoLink (updateSlot/updateComponent)
  ↓
Update local cache
  ↓
Refresh inspector display
```

### Edit Flow (REPL)
```
User: edit BoxMesh
  ↓
Resolve path (current context)
  ↓
Get component data
  ↓
Launch external editor (YAML/JSON)
  ↓
User saves and exits
  ↓
Parse edited file
  ↓
Validate changes
  ↓
Apply via ResoLink
  ↓
Confirm success
```

## 9. RefID Display & Usage

### RefID Format
Resonite uses RefIDs in the format: `ID` followed by alphanumeric characters (e.g., `ID2300`, `IDabc123`, `ID0f8e2a`)
- Root slot typically has ID: `ID2300` (can also be identified by name "Root" with null parent)
- Custom IDs can be assigned when creating slots/components
- If no ID provided, Resonite auto-generates one

### Display Format
- Slots: `SlotName (s-refid)` or `SlotName (refid)`
- Components: `ComponentName (c-refid)` or `ComponentName (refid)`
- Configurable prefix or no prefix
- Special case: Root slot can be referenced as `"Root"` string or `"ID2300"`

### RefID as Links
When a property references another slot/component:
```
Parent: Environment (s-100)  [Click to jump]
         ↑               ↑
      Display name     RefID
```

### Manual RefID Entry
In inspector, when editing a reference field:
```
Target: [s-123________]
        ↑ Type RefID here
```

Validates RefID exists and is correct type.

## 10. Error Handling

### Connection Errors
- Auto-reconnect with exponential backoff
- Show connection status in UI
- Queue commands during disconnect, replay on reconnect

### Validation Errors
- Inline error messages in inspector
- Detailed error in status bar
- Revert to last valid value on cancel

### Script Errors
- Catch Lua errors, display in console
- Stack trace available
- Don't crash main app

## 11. Configuration

### Config File: `~/.resonite-sh/config.yaml`
```yaml
connection:
  url: ws://localhost:29551
  auto_reconnect: true
  
resh:
  slot_name: RESH
  order_offset: 999
  auto_create: true

tui:
  default_mode: tui  # or repl
  show_refids: true
  refid_prefix: true  # s- and c- prefixes
  mouse_enabled: true
  colors: true
  
repl:
  prompt: "resonite:%p> "
  history_size: 1000
  
editor:
  command: $EDITOR
  format: yaml  # or json
  
performance:
  update_interval: 0.1s  # 100ms
  cache_ttl: 5m
  lazy_load: true
  prefetch: true

lua:
  scripts_dir: ./scripts
  autorun: []  # Scripts to run on connect
```

## 12. Project Structure (Updated)

```
resonite-sh/
├─ cmd/
│  ├─ root.go          # CLI entry point
│  └─ repl.go          # Launch REPL/TUI
├─ pkg/
│  ├─ resolink/        # ResoLink WebSocket client
│  │  ├─ client.go
│  │  ├─ slots.go
│  │  └─ components.go
│  ├─ objects/         # Data models
│  │  ├─ slot.go
│  │  ├─ component.go
│  │  └─ property.go
│  ├─ vfs/             # Virtual filesystem
│  │  └─ vfs.go
│  ├─ resh/            # RESH slot management
│  │  ├─ init.go       # Create RESH slot
│  │  └─ variables.go  # Variable storage
│  ├─ tui/             # Bubble Tea TUI
│  │  ├─ app.go        # Main TUI app
│  │  ├─ tree.go       # Tree panel
│  │  ├─ inspector.go  # Inspector panel
│  │  └─ statusbar.go  # Status/command bar
│  ├─ repl/            # REPL shell
│  │  ├─ shell.go
│  │  └─ commands.go
│  ├─ lua/             # Lua scripting
│  │  ├─ engine.go     # Lua VM
│  │  └─ api.go        # Exposed API
│  └─ session/         # Session management
│     ├─ session.go    # Current state
│     └─ variables.go  # Variable management
├─ scripts/            # Default Lua scripts
│  └─ examples/
├─ main.go
├─ go.mod
└─ go.sum
```

## Implementation Priority

### Phase 1: Foundation (Weeks 1-2)
1. ✓ ResoLink client basics
2. ✓ Object models
3. ✓ VFS skeleton
4. Complete ResoLink protocol implementation
5. Add RESH slot creation logic

### Phase 2: REPL Mode (Week 3)
1. Complete REPL commands (cd, ls, cat, edit, etc.)
2. Path resolution and navigation
3. Component editing with external editor
4. History and completion

### Phase 3: TUI Mode (Weeks 4-5)
1. Bubble Tea app structure
2. Tree panel with lazy loading
3. Inspector panel with property display
4. Mode switching (Tab key)
5. Mouse support

### Phase 4: Variables & RESH (Week 6)
1. RESH slot initialization
2. DynamicVariable component creation
3. Variable CRUD operations
4. Session vs local vs world variables

### Phase 5: Lua Integration (Week 7)
1. gopher-lua integration
2. API bindings (RESH.* functions)
3. Script execution (file, var, inline)
4. Error handling

### Phase 6: Polish & Features (Week 8+)
1. Auto-update/polling
2. Search and filtering
3. RefID display and linking
4. Property validation
5. Bulk operations support
6. Documentation and examples

## Open Technical Questions

1. **ResoLink Update Events**: ResoniteLink does NOT support push updates - polling is required ✓
2. **RefID Format**: Format is `ID` + alphanumeric (e.g., `ID2300`, `IDabc123`) ✓
3. **Component Type Discovery**: ResoniteLink provides `getComponentTypeList` message to query available types ✓
4. **Component Definitions**: `getComponentDefinition` provides full member information for components ✓
5. **DynamicVariable Limits**: Need to test - may have limits on variable count or size
6. **Root Slot**: Can be accessed via special string `"Root"` or typically `"ID2300"`
7. **Undo/Redo**: Should we track edit history for undo/redo? (Future feature)
8. **Batch Operations**: ResoniteLink supports `dataModelOperationBatch` for multiple operations in one message
