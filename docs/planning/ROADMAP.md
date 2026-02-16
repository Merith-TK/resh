# Revised Implementation Roadmap

Based on the detailed design decisions, here's the updated implementation plan for resonite-sh.

## Architecture Summary

**Core Modes**: Dual-mode system (TUI + REPL) with Tab to switch
**Storage**: RESH slot in Resonite world for persistent variables
**Scripting**: Lua with full API access
**Navigation**: Filesystem metaphor (slots=folders, components=files)

## Phase 1: Core Infrastructure ✓ (Weeks 1-2)

### Week 1: ResoLink Client
- [x] Project structure
- [x] Basic WebSocket client
- [x] Message protocol structures
- [ ] Complete all ResoLink operations:
  - [x] Slot operations (add, get, update, remove, list, find)
  - [x] Component operations (add, get, update, remove, list)
  - [ ] Response parsing and error handling
  - [ ] Connection lifecycle (connect, disconnect, reconnect)
  - [ ] Request timeout handling
- [ ] Unit tests for client

### Week 2: Object Models & VFS
- [x] Slot and Component models
- [ ] Property type system (primitives, references, lists)
- [ ] Member serialization/deserialization
- [ ] VFS path resolution
- [ ] Lazy loading implementation
- [ ] Cache management
- [ ] Unit tests for models and VFS

## Phase 2: RESH Slot System (Week 3)

### RESH Initialization
- [ ] Detect existing RESH slot on connect
- [ ] Create RESH slot if missing:
  ```go
  - Create slot with name "RESH", OrderOffset 999
  - Add DynamicReferenceVariable<Slot> component
  - Add DynamicVariableSpace component with SpaceName "RESH"
  ```
- [ ] Verify RESH structure

### Variable Management
- [ ] Variable CRUD operations via DynamicVariable components
- [ ] Support for value/object/reference variable types
- [ ] Session variables (in-memory map)
- [ ] Local variables (file storage ~/.resonite-sh/vars.yaml)
- [ ] World variables (RESH slot storage)
- [ ] Variable scope resolution (session → local → world)

### Variable API
```go
pkg/resh/
├─ init.go          # RESH slot creation
├─ variables.go     # Variable CRUD
└─ manager.go       # Variable scope management
```

## Phase 3: REPL Mode (Week 4)

### Core REPL Commands
- [x] Basic shell structure
- [x] pwd, cd, ls (skeleton)
- [ ] Complete navigation commands:
  - [ ] `cd <path>` - Change directory
  - [ ] `pwd` - Print working directory  
  - [ ] `ls [path]` - List with RefIDs shown
  - [ ] `tree [path]` - Hierarchical view

### Inspection Commands
- [ ] `cat <path>` - Show slot/component properties
- [ ] `stat <path>` - Detailed metadata
- [ ] `find <pattern>` - Search for slots/components
- [ ] `grep <pattern>` - Search in property values

### Modification Commands
- [ ] `edit <path>` - External editor (YAML/JSON)
- [ ] `set <path>.<prop> <value>` - Inline property set
- [ ] `mkdir <name>` - Create new slot
- [ ] `touch <component>` - Add component
- [ ] `rm <path>` - Remove slot/component
- [ ] `mv <src> <dst>` - Move/rename
- [ ] `cp <src> <dst>` - Copy hierarchy

### REPL Features
- [ ] Command history (arrow keys)
- [ ] Tab completion (commands, paths, types)
- [ ] Prompt with current path
- [ ] Help system
- [ ] Alias support

## Phase 4: TUI Mode (Weeks 5-6)

### Week 5: Bubble Tea Foundation

#### App Structure
```go
pkg/tui/
├─ app.go           # Main Bubble Tea app
├─ tree.go          # Left panel: slot tree
├─ inspector.go     # Right panel: property inspector
├─ statusbar.go     # Bottom: status + command line
├─ models.go        # Bubble Tea models
└─ styles.go        # Lipgloss styles
```

#### Basic TUI
- [ ] Bubble Tea app initialization
- [ ] Layout with two panels + status bar
- [ ] Keyboard handling (arrow keys, Tab, Esc, Enter)
- [ ] Mode switching (Tab between TUI/REPL)
- [ ] Message passing between components

### Week 6: Tree & Inspector Panels

#### Tree Panel
- [ ] Hierarchical slot display
- [ ] Show RefIDs in parentheses: `SlotName (s-refid)`
- [ ] Expandable/collapsible nodes
- [ ] Lazy loading on expand
- [ ] Virtual scrolling (only render visible)
- [ ] Selection highlighting
- [ ] Arrow key navigation
- [ ] Search/filter overlay (/)

#### Inspector Panel
- [ ] Property display for selected slot/component
- [ ] Show all editable fields
- [ ] RefID display and clickable links
- [ ] Mouse support for clicking fields
- [ ] Inline editing (Enter to edit, Esc to cancel)
- [ ] Boolean toggles (Space bar)
- [ ] Field validation with error display
- [ ] Apply/Reset/Remove buttons
- [ ] Component list for slots
- [ ] Children list for slots

#### Status Bar
- [ ] Connection status indicator
- [ ] Current path breadcrumb
- [ ] Command mode (: for commands)
- [ ] Status messages
- [ ] Help hints

## Phase 5: Lua Scripting (Week 7)

### Lua Engine Integration
- [ ] gopher-lua initialization
- [ ] Script loading (file, variable, inline)
- [ ] Error handling and display
- [ ] Script execution context

### Lua API Implementation
```lua
-- Core API
RESH.cd(path)
RESH.pwd()
RESH.ls(path)
RESH.createSlot(name, parentId)
RESH.removeSlot(slotId)
RESH.getSlot(slotId)
RESH.updateSlot(slotId, props)
RESH.addComponent(slotId, type)
RESH.removeComponent(componentId)
RESH.getComponent(componentId)
RESH.updateComponent(componentId, members)

-- Variables
RESH.VARS.MyVar = value
value = RESH.VARS.MyVar

-- UI Control
RESH.refresh()
RESH.focus(slotId)
RESH.message(text)
RESH.switchMode("tui"|"repl")

-- Utilities
RESH.find(pattern)
RESH.filter(function)
```

### Script Execution
- [ ] `:lua filepath` - Execute file
- [ ] `:lua VAR.ScriptName` - Execute from variable
- [ ] `:lua code` - Execute inline
- [ ] Autorun scripts on connect
- [ ] Script output display in console

### Script Storage
- [ ] Store scripts in RESH.VARS
- [ ] Load scripts from local files
- [ ] Script presets in config
- [ ] Script library management

## Phase 6: Polish & Advanced Features (Week 8+)

### Real-time Updates
- [ ] Check if ResoLink supports push updates
- [ ] If yes: Subscribe to focused slot changes
- [ ] If no: Implement 0.1s polling for focused slot
- [ ] Smart polling (slow when idle, fast when editing)
- [ ] Manual refresh (Ctrl+R)
- [ ] Cache invalidation on updates

### RefID System
- [ ] RefID display everywhere (tree, inspector)
- [ ] Configurable prefix (s-, c-, or none)
- [ ] Clickable RefID links (jump to slot/component)
- [ ] Manual RefID entry in reference fields
- [ ] RefID validation
- [ ] RefID autocomplete

### Search & Filter
- [ ] Global search (Ctrl+F in TUI)
- [ ] Tree filtering
- [ ] Property search
- [ ] Search history
- [ ] Regex support
- [ ] Jump to search results

### Property Editing
- [ ] All property types supported:
  - Primitives (int, float, string, bool)
  - Vectors (float2, float3, float4, floatQ)
  - Colors (color, colorX)
  - References (slot, component)
  - Lists (Materials, etc.)
- [ ] Type-specific editors
- [ ] Validation with helpful errors
- [ ] Undo/Redo support (optional)
- [ ] Copy/paste property values
- [ ] Property history

### Field Drivers
- [ ] Show what's driving a field
- [ ] Click to jump to driver
- [ ] Remove driver option

### Bulk Operations (Script-based)
- [ ] Multi-select API for Lua
- [ ] Batch update examples
- [ ] Find-and-replace scripts
- [ ] Bulk component addition

### Configuration
- [ ] Complete config system
- [ ] Config validation
- [ ] Runtime config reload
- [ ] Profile switching
- [ ] Export/import settings

### Documentation
- [ ] User guide
- [ ] Command reference
- [ ] Lua API documentation
- [ ] Example scripts
- [ ] Video tutorials (optional)
- [ ] Troubleshooting guide

## Phase 7: Testing & Quality (Ongoing)

### Unit Tests
- [ ] ResoLink client tests
- [ ] VFS tests
- [ ] Object model tests
- [ ] Variable management tests
- [ ] Lua API tests

### Integration Tests
- [ ] Mock WebSocket server
- [ ] End-to-end command tests
- [ ] TUI interaction tests
- [ ] Script execution tests

### Performance Tests
- [ ] Large world loading (1000+ slots)
- [ ] Rapid property updates
- [ ] Memory usage profiling
- [ ] CPU usage profiling
- [ ] Network efficiency

### User Testing
- [ ] Alpha testers
- [ ] Feedback collection
- [ ] Bug reports and fixes
- [ ] Feature requests

## Deliverables Timeline

**Week 2**: Working ResoLink client, basic models
**Week 3**: RESH slot creation, variable system
**Week 4**: Functional REPL mode with all commands
**Week 6**: Functional TUI mode with tree and inspector
**Week 7**: Lua scripting fully integrated
**Week 8+**: Polished, tested, documented release

## Success Metrics

- [ ] Can connect to Resonite ResoLink
- [ ] RESH slot created and variables stored
- [ ] Navigate world in REPL mode
- [ ] Navigate world in TUI mode
- [ ] Edit slot and component properties
- [ ] Execute Lua scripts
- [ ] Scripts can manipulate world
- [ ] Variables persist across sessions
- [ ] Real-time or near-real-time updates
- [ ] Stable with large worlds (1000+ slots)
- [ ] Comprehensive documentation

## Dependencies

```go
// Core
github.com/gorilla/websocket          // WebSocket client
github.com/google/uuid                // Message IDs

// CLI
github.com/spf13/cobra               // Command structure
github.com/spf13/viper               // Configuration
github.com/chzyer/readline           // REPL line editing

// TUI
github.com/charmbracelet/bubbletea   // TUI framework
github.com/charmbracelet/lipgloss    // TUI styling
github.com/charmbracelet/bubbles     // TUI components (optional)

// Scripting
github.com/yuin/gopher-lua           // Lua VM

// Utilities
gopkg.in/yaml.v3                     // YAML parsing
```

## Risk Mitigation

**Risk**: ResoLink protocol changes
**Mitigation**: Version detection, backwards compatibility layer

**Risk**: Performance issues with large worlds
**Mitigation**: Aggressive lazy loading, virtual scrolling, caching

**Risk**: Lua scripts crashing app
**Mitigation**: Sandboxing, panic recovery, timeout limits

**Risk**: Complex property types
**Mitigation**: Start with common types, expand gradually

**Risk**: Real-time updates not available
**Mitigation**: Polling fallback with configurable interval

## Next Steps

1. Complete ResoLink client implementation
2. Implement RESH slot initialization
3. Build out REPL commands
4. Start TUI with Bubble Tea
5. Integrate Lua engine
6. Test, polish, document
