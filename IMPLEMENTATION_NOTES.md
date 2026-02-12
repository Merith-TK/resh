# Implementation Notes & Order

## Documentation Organization

All planning documentation is complete and organized as follows:

### Core Documentation
- **README.md** - Project overview, goals, and high-level roadmap
- **QUICKSTART.md** - Developer quick start guide
- **PROTOCOL.md** - Complete ResoniteLink WebSocket protocol reference

### Design Documentation  
- **DESIGN_DECISIONS.md** - Complete design specification (TUI/REPL, variables, etc.)
- **ARCHITECTURE.md** - Original architecture design document
- **ROADMAP.md** - 8-week implementation timeline

### Planning Documentation
- **PLANNING_SUMMARY.md** - Current status and confirmed details
- **TODO.md** - Detailed task breakdown by phase

### Submodules
- **submodules/ResoniteLink/** - Official C# implementation for reference

---

## Implementation Order

### Phase 1: ResoniteLink Client (Current)
**Goal**: Complete WebSocket client with full protocol support

**Tasks**:
- [ ] Proper JSON serialization with `$type` discriminator
- [ ] All message type structs (slots, components, reflection)
- [ ] Complete request/response handling
- [ ] Error handling and timeouts
- [ ] Connection lifecycle (connect, disconnect, reconnect)
- [ ] Helper methods for common operations
- [ ] Unit tests

**Files to implement**:
- `pkg/resolink/client.go` - Core client (partially done)
- `pkg/resolink/messages.go` - Message type definitions
- `pkg/resolink/slots.go` - Slot operations (partially done)
- `pkg/resolink/components.go` - Component operations (partially done)
- `pkg/resolink/reflection.go` - Type discovery operations
- `pkg/resolink/types.go` - Data type helpers

**Success Criteria**:
- Can connect to Resonite ResoniteLink
- Can query root slot
- Can create/update/delete slots
- Can create/update/delete components
- Can query component types
- Error handling works
- Reconnection works

---

### Phase 2: RESH Slot System
**Goal**: Initialize persistent variable storage in Resonite world

**Tasks**:
- [ ] Detect existing RESH slot on connect
- [ ] Create RESH slot if missing (with proper structure)
- [ ] Add DynamicVariableSpace component
- [ ] Add DynamicReferenceVariable component  
- [ ] Variable CRUD operations
- [ ] Session variable management (in-memory)
- [ ] Local variable management (file storage)
- [ ] World variable management (via RESH slot)

**Files to implement**:
- `pkg/resh/init.go` - RESH slot detection and creation
- `pkg/resh/variables.go` - Variable CRUD operations
- `pkg/resh/manager.go` - Variable scope management (session/local/world)

**Success Criteria**:
- RESH slot created on first connect
- Existing RESH slot detected on subsequent connects
- Can create/read/update/delete variables
- Variables persist in world
- Session and local variables work

---

### Phase 3: REPL Mode
**Goal**: Functional shell interface for world navigation

**Tasks**:
- [ ] Complete navigation commands (cd, ls, pwd, tree)
- [ ] Inspection commands (cat, stat, find, grep)
- [ ] Modification commands (edit, set, mkdir, touch, rm, mv, cp)
- [ ] Command history and completion
- [ ] Path resolution (absolute/relative)
- [ ] RefID display in output
- [ ] Help system
- [ ] Error handling with helpful messages

**Files to implement**:
- `pkg/repl/shell.go` - Main REPL loop (partially done)
- `pkg/repl/commands/` - Individual command implementations
  - `navigation.go` - cd, ls, pwd, tree
  - `inspection.go` - cat, stat, find, grep
  - `modification.go` - edit, set, mkdir, touch, rm, mv, cp
  - `variables.go` - var commands
- `pkg/repl/completion.go` - Tab completion
- `pkg/repl/history.go` - Command history

**Success Criteria**:
- Can navigate world with cd/ls/pwd
- Can inspect slots and components with cat
- Can create/modify/delete slots and components
- RefIDs shown in output
- Tab completion works
- Command history works

---

### Phase 4: TUI Mode
**Goal**: Visual interface with tree panel and inspector

**Tasks**:
- [ ] Bubble Tea app structure
- [ ] Tree panel with hierarchical display
- [ ] Inspector panel for property editing
- [ ] Status bar with connection status
- [ ] Keyboard navigation (arrows, Tab, Enter, Esc)
- [ ] Mouse support in inspector
- [ ] Mode switching (Tab between TUI/REPL)
- [ ] Lazy loading for tree nodes
- [ ] Virtual scrolling
- [ ] RefID display with clickable links
- [ ] Property validation and inline editing

**Files to implement**:
- `pkg/tui/app.go` - Main Bubble Tea application
- `pkg/tui/tree.go` - Tree panel component
- `pkg/tui/inspector.go` - Inspector panel component
- `pkg/tui/statusbar.go` - Status bar component
- `pkg/tui/styles.go` - Lipgloss styles
- `pkg/tui/messages.go` - Bubble Tea messages
- `pkg/tui/keybinds.go` - Key bindings

**Success Criteria**:
- TUI launches and displays tree + inspector
- Can navigate tree with arrow keys
- Can expand/collapse slots
- Inspector shows selected slot/component
- Can edit properties inline
- Tab switches to REPL mode
- Mouse clicking works
- RefIDs are clickable

---

### Phase 5: Lua Scripting
**Goal**: Full Lua integration for automation

**Tasks**:
- [ ] gopher-lua VM initialization
- [ ] Lua API bindings (RESH.* functions)
- [ ] Script execution (file, variable, inline)
- [ ] Variable access (RESH.VARS)
- [ ] Error handling and display
- [ ] Script storage in variables
- [ ] Autorun scripts on connect
- [ ] Example scripts

**Files to implement**:
- `pkg/lua/engine.go` - Lua VM management
- `pkg/lua/api.go` - RESH API bindings
- `pkg/lua/bindings.go` - Helper functions
- `scripts/examples/` - Example Lua scripts

**Success Criteria**:
- Can execute Lua scripts from files
- Can execute Lua scripts from variables
- RESH API works (cd, ls, createSlot, etc.)
- RESH.VARS works for variables
- Scripts can manipulate world
- Error handling works
- Example scripts demonstrate capabilities

---

## Shared Components

These are used across all phases:

### Objects Package
- `pkg/objects/slot.go` ✓ (basic, needs expansion)
- `pkg/objects/component.go` ✓ (basic, needs expansion)
- `pkg/objects/property.go` (to be created)

### VFS Package
- `pkg/vfs/vfs.go` ✓ (skeleton, needs completion)

### Session Package
- `pkg/session/session.go` (to be created)
- `pkg/session/context.go` (to be created)

---

## Current Status

**Completed**:
- ✅ All planning and design documentation
- ✅ Project structure
- ✅ Dependencies in go.mod
- ✅ Basic skeleton code
- ✅ ResoniteLink submodule added

**In Progress**:
- 🔄 Phase 1: ResoniteLink Client (starting now)

**Next**:
- Phase 2: RESH Slot System
- Phase 3: REPL Mode
- Phase 4: TUI Mode
- Phase 5: Lua Scripting

---

## Testing Strategy

### Unit Tests (per phase)
- Test each package independently
- Mock WebSocket for client tests
- Mock client for VFS tests

### Integration Tests
- Mock Resonite server
- Test full command flow
- Test TUI interactions

### Manual Testing
- Connect to real Resonite instance
- Test all commands
- Test edge cases

---

## Phase 1 Detailed Plan (Next Steps)

### Step 1.1: Message Type Definitions
Create proper Go structs for all ResoniteLink message types with JSON tags.

### Step 1.2: JSON Serialization
Implement proper serialization with `$type` discriminator field.

### Step 1.3: Core Operations
Complete slot and component operations with proper error handling.

### Step 1.4: Reflection API
Implement type discovery for component types.

### Step 1.5: Testing
Write unit tests for client operations.

Let's begin! 🚀
