# Architecture Design Document

## Overview

`resonite-sh` implements a filesystem-like abstraction over Resonite's scene graph, accessible through ResoLink WebSocket API. The key innovation is treating the hierarchical slot/component structure as a filesystem where slots are directories and components are files.

## Core Concepts

### 1. Everything is an Object

The system models the Resonite world as a hierarchical object system:

```
Slot (Directory-like)
├── Properties (editable metadata)
│   ├── Name: string
│   ├── Position: [x, y, z]
│   ├── Rotation: [x, y, z, w]
│   ├── Scale: [x, y, z]
│   └── Active: bool
├── Components (File-like)
│   ├── Transform
│   ├── BoxMesh
│   ├── MeshRenderer
│   └── PBS_Metallic
└── Children (Sub-directories)
    ├── ChildSlot1/
    └── ChildSlot2/
```

### 2. Dual Nature of Slots

Slots have a **dual nature**:
- **As a Directory**: Contains components and child slots
- **As an Object**: Has editable properties (name, transform, etc.)

This is handled by:
- `ls` shows contents (components + children)
- `cat <SlotName>` shows slot properties
- `edit <SlotName>` opens slot properties for editing
- `cd <SlotName>` changes context into that slot

### 3. Components as Files

Components are treated as files with content:
- `cat BoxMesh` displays component properties
- `edit BoxMesh` allows editing component members
- `rm BoxMesh` removes the component
- `touch BoxMesh` adds a new component

## System Architecture

```
┌─────────────────────────────────────────────────────────┐
│                     CLI Interface                        │
│  (cobra commands, flags, argument parsing)              │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                   REPL Shell                             │
│  - Interactive prompt (readline)                         │
│  - Command history                                       │
│  - Tab completion                                        │
│  - Context management (current directory)                │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                 Command Layer                            │
│  - Command parsing and validation                        │
│  - Path resolution                                       │
│  - Output formatting                                     │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│              Virtual Filesystem (VFS)                    │
│  - Slot tree representation                              │
│  - Path-based navigation                                 │
│  - Cache management                                      │
│  - Lazy loading                                          │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                Object Model Layer                        │
│  - Slot abstraction                                      │
│  - Component abstraction                                 │
│  - Property system                                       │
│  - Type definitions                                      │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                ResoLink Client                           │
│  - WebSocket connection (ws://localhost:29551)           │
│  - Protocol message encoding/decoding                    │
│  - Request/response correlation                          │
│  - Connection management                                 │
└─────────────────────┬───────────────────────────────────┘
                      │
                      ▼
              Resonite VR Instance
```

## Module Breakdown

### pkg/resolink/

**Purpose**: Low-level ResoLink WebSocket client

**Responsibilities**:
- WebSocket connection lifecycle
- Message serialization/deserialization
- Request/response correlation via `messageId`
- Timeout handling
- Reconnection logic

**Key Types**:
```go
type Client struct {
    conn *websocket.Conn
    url  string
    timeout time.Duration
    pendingRequests map[string]chan *Response
}

type Message struct {
    MessageId string                 `json:"messageId"`
    Type      string                 `json:"type"`
    Data      map[string]interface{} `json:"data,omitempty"`
}
```

**Key Operations**:
- `Connect(url string) error`
- `Disconnect() error`
- `SendRequest(msgType string, data map[string]interface{}) (*Response, error)`
- `AddSlot(parentId, name string) (string, error)`
- `GetSlot(slotId string, includeComponents bool) (*Slot, error)`
- `UpdateSlot(slotId string, properties map[string]interface{}) error`
- `RemoveSlot(slotId string) error`
- `AddComponent(slotId, componentType string) (string, error)`
- `UpdateComponent(componentId string, members map[string]interface{}) error`

### pkg/objects/

**Purpose**: Object model for Slots and Components

**Responsibilities**:
- Slot structure representation
- Component structure representation
- Property type system
- Serialization/deserialization

**Key Types**:
```go
type Slot struct {
    ID         string
    Name       string
    ParentID   string
    Position   [3]float64
    Rotation   [4]float64
    Scale      [3]float64
    Active     bool
    Components []Component
    Children   []string  // Child slot IDs
}

type Component struct {
    ID         string
    SlotID     string
    Type       string  // e.g., "FrooxEngine.BoxMesh"
    Members    map[string]Member
}

type Member struct {
    Type  string      // "primitive", "reference", "list"
    Value interface{} // Actual value
}
```

### pkg/vfs/

**Purpose**: Virtual filesystem abstraction

**Responsibilities**:
- Maintain in-memory slot tree
- Path-based navigation
- Caching with invalidation
- Lazy loading of subtrees
- Path resolution (absolute/relative)

**Key Types**:
```go
type VFS struct {
    client      *resolink.Client
    root        *Node
    currentPath string
    cache       map[string]*Node
}

type Node struct {
    Slot     *objects.Slot
    Parent   *Node
    Children map[string]*Node
    Loaded   bool
}
```

**Key Operations**:
- `ResolvePath(path string) (*Node, error)`
- `ListDirectory(path string) ([]*Node, error)`
- `GetNode(path string) (*Node, error)`
- `CreateSlot(path, name string) error`
- `RemoveSlot(path string) error`
- `GetCurrentPath() string`
- `ChangeDirectory(path string) error`

### pkg/commands/

**Purpose**: Command implementations

**Responsibilities**:
- Parse command arguments
- Execute operations via VFS
- Format output
- Error handling

**Command Categories**:

1. **Navigation Commands**:
   - `cd <path>` - Change directory
   - `pwd` - Print working directory
   - `ls [path]` - List contents
   - `tree [path]` - Show tree structure

2. **Inspection Commands**:
   - `cat <path>` - Show object/component properties
   - `stat <path>` - Show detailed metadata
   - `inspect <path>` - Interactive property viewer
   - `find <pattern>` - Search for slots/components

3. **Modification Commands**:
   - `edit <path>` - Edit properties interactively
   - `set <path> <property> <value>` - Set single property
   - `mkdir <name>` - Create new slot
   - `touch <component>` - Add component
   - `rm <path>` - Remove component or slot
   - `mv <src> <dst>` - Move/rename
   - `cp <src> <dst>` - Copy slot hierarchy

4. **Resonite-Specific Commands**:
   - `spawn <template>` - Create from template
   - `attach <component>` - Add component to current slot
   - `components` - List available component types
   - `search-components <query>` - Search component database

### pkg/repl/

**Purpose**: Interactive REPL environment

**Responsibilities**:
- Interactive prompt
- Command history
- Tab completion
- Multi-line input
- Syntax highlighting

**Key Features**:
- Prompt shows current path: `resonite:/World/Environment>`
- Tab completion for:
  - Command names
  - Paths (slots/components)
  - Component types
- Command history with up/down arrows
- Multi-line editing for complex operations

## Data Flow Examples

### Example 1: Listing Directory Contents

```
User: ls
  ↓
REPL: Parse command "ls"
  ↓
Command Layer: Execute ListCommand with currentPath
  ↓
VFS: ResolvePath(currentPath) → Get Node
  ↓
VFS: Check if Node.Loaded
  ↓ (if not loaded)
ResoLink: GetSlot(node.Slot.ID, includeComponents=false)
  ↓
ResoLink: ListChildren(node.Slot.ID)
  ↓
VFS: Update cache with children
  ↓
Command Layer: Format output
  ↓
REPL: Display to user
```

### Example 2: Editing Component Properties

```
User: edit BoxMesh
  ↓
REPL: Parse command "edit BoxMesh"
  ↓
Command Layer: Resolve "BoxMesh" relative to currentPath
  ↓
VFS: Find component in current slot
  ↓
ResoLink: GetComponent(componentId)
  ↓
Command Layer: Launch interactive editor
  ↓
User: Modify properties in editor
  ↓
Command Layer: Validate changes
  ↓
ResoLink: UpdateComponent(componentId, newProperties)
  ↓
VFS: Invalidate cache for parent slot
  ↓
REPL: Confirm success
```

### Example 3: Creating New Object Hierarchy

```
User: mkdir MyObject
User: cd MyObject
User: touch BoxMesh
User: touch MeshRenderer
  ↓
Command Layer: Create slot "MyObject"
  ↓
VFS: ResolvePath(currentPath + "/MyObject")
  ↓
ResoLink: AddSlot(parentId, "MyObject")
  ↓
VFS: Add to cache, update current path
  ↓
Command Layer: Add BoxMesh component
  ↓
ResoLink: AddComponent(slotId, "FrooxEngine.BoxMesh")
  ↓
Command Layer: Add MeshRenderer component
  ↓
ResoLink: AddComponent(slotId, "FrooxEngine.MeshRenderer")
  ↓
VFS: Invalidate slot cache
```

## Design Decisions

### 1. Lazy Loading

**Decision**: Load slot data on-demand rather than loading entire scene graph upfront

**Rationale**:
- Resonite worlds can be massive (thousands of slots)
- Reduces initial connection time
- Minimizes memory usage
- Allows working in specific areas without full scan

**Implementation**:
- VFS maintains `Loaded` flag per node
- First access triggers load
- Cache invalidation on modifications

### 2. Caching Strategy

**Decision**: Cache loaded slot/component data with manual invalidation

**Rationale**:
- Reduces repeated WebSocket requests
- Improves responsiveness
- Allows offline inspection of loaded data

**Invalidation Rules**:
- Modification operations invalidate affected node
- Timer-based refresh option (configurable)
- Manual `refresh` command

### 3. Component Editing

**Decision**: Use external editor for complex properties, inline for simple values

**Rationale**:
- Complex components have many nested properties
- External editor provides better UX for structured data
- Inline editing for quick single-property changes

**Modes**:
- `edit <component>` - Launch external editor (JSON/YAML)
- `set <component>.<property> <value>` - Inline single property

### 4. Path Resolution

**Decision**: Support both absolute and relative paths with Unix-like syntax

**Rationale**:
- Familiar to developers
- Flexible navigation
- Supports scripting

**Path Syntax**:
- Absolute: `/World/Environment/MyObject`
- Relative: `../OtherObject` or `Child/Grandchild`
- Current: `.` (current slot)
- Parent: `..` (parent slot)

### 5. Error Handling

**Decision**: Graceful degradation with informative error messages

**Rationale**:
- WebSocket can disconnect
- Resonite might not respond
- Invalid operations should not crash shell

**Strategy**:
- Reconnection attempts with backoff
- Clear error messages with suggestions
- Continue shell operation even on failure

## Configuration

### Config File (`~/.resonite-sh/config.yaml`)

```yaml
connection:
  url: ws://localhost:29551
  timeout: 30s
  reconnect: true
  reconnect_interval: 5s

cache:
  enable: true
  ttl: 5m

editor:
  command: $EDITOR
  format: yaml  # or json

ui:
  prompt: "resonite:%p> "
  colors: true
  
history:
  size: 1000
  file: ~/.resonite-sh/history
```

### Environment Variables

- `RESONITE_WS_URL`: Override WebSocket URL
- `RESONITE_SH_CONFIG`: Config file location
- `EDITOR`: External editor for property editing

## Future Extensions

### 1. Scripting Language

Add scripting support for automation:

```bash
# script.rsh
cd /World
for slot in $(ls); do
    if [[ $(cat $slot/Transform.Position.y) -lt 0 ]]; then
        set $slot/Transform.Position.y 0
    fi
done
```

### 2. Component Templates

Pre-defined templates for common objects:

```bash
spawn cube --size 2 --color red
spawn button --text "Click Me" --callback my_function
```

### 3. ProtoFlux Graph Visualization

ASCII-art visualization of ProtoFlux connections:

```
[ValueInput<int>] ──> [ValueAdd<int>] ──> [ValueDisplay<int>]
       ↑                    ↑
       │                    │
   Input1              Input2
```

### 4. Diff and Merge

Track changes and merge operations:

```bash
diff /World/Before /World/After
merge /Template /World/MyObject
```

### 5. Remote Execution

Execute commands on different Resonite instances:

```bash
resonite-sh --remote ws://other-host:29551
```

## Testing Strategy

### Unit Tests
- ResoLink client message handling
- VFS path resolution
- Object model serialization
- Command parsing

### Integration Tests
- Mock WebSocket server for protocol testing
- Full command execution flow
- Cache invalidation scenarios

### E2E Tests
- Connect to actual Resonite instance (headless mode?)
- Execute common workflows
- Performance benchmarks

## Implementation Phases

### Phase 1: Foundation (Weeks 1-2)
- ResoLink WebSocket client
- Basic message protocol
- Connection management
- Unit tests

### Phase 2: Object Model (Week 3)
- Slot and Component types
- Property system
- Serialization
- Type definitions for common components

### Phase 3: Virtual Filesystem (Week 4)
- VFS implementation
- Path resolution
- Caching layer
- Lazy loading

### Phase 4: Core Commands (Weeks 5-6)
- Navigation commands (cd, ls, pwd, tree)
- Inspection commands (cat, stat)
- Basic modification (mkdir, touch, rm)

### Phase 5: REPL (Week 7)
- Interactive shell
- Command history
- Basic tab completion
- Prompt customization

### Phase 6: Advanced Features (Week 8+)
- Complex editing
- Component templates
- Scripting support
- Advanced commands

## Performance Considerations

### 1. WebSocket Connection
- Keep-alive pings
- Connection pooling (future)
- Request batching

### 2. Caching
- LRU eviction for memory limits
- Configurable TTL
- Selective invalidation

### 3. Large Hierarchies
- Pagination for list operations
- Streaming for tree views
- Depth limits for recursive operations

### 4. Concurrent Operations
- Goroutines for async operations
- Request queue management
- Timeout handling

## Security Considerations

### 1. Local-Only by Default
- Default connection to localhost:29551
- Warning when connecting to remote hosts

### 2. Input Validation
- Sanitize all user input
- Validate paths before operations
- Type checking for property values

### 3. Destructive Operations
- Confirmation prompts for rm/rmdir
- Option for safe mode (no destructive ops)
- Operation logging

## Open Questions

1. How to handle concurrent modifications from other sources (Resonite UI, other scripts)?
   - Potential solution: Event subscription via ResoLink (if supported)
   - Fallback: Cache TTL and manual refresh

2. Component type discovery - should we bundle type definitions or query dynamically?
   - Hybrid approach: Bundle common types, query for unknown

3. How to represent complex component relationships (references between components)?
   - Show as symbolic links or special notation?

4. Support for ProtoFlux graph manipulation?
   - Phase 2 feature - needs special handling for connections

5. Multi-user scenarios?
   - Out of scope for v1.0
   - Future: User tracking and conflict resolution

## References

- [ResoLink MCP Documentation](https://deepwiki.com/rassi0429/resolink-mcp)
- [resolink-mcp GitHub](https://github.com/rassi0429/resolink-mcp)
- [Resonite VR](https://resonite.com/)
