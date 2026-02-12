# resonite-sh

A Golang-based REPL/shell interface for Resonite VR that treats everything as an object in a filesystem-like hierarchy.

## Concept

`resonite-sh` provides a command-line interface to interact with Resonite worlds through ResoLink (WebSocket API on `ws://localhost:29551`). The core philosophy is **"everything is an object"**:

- **Slots** are folders that can contain components
- **Components** are files within slot folders
- Edit operations work on both containers and contents
- Navigate the scene graph like a filesystem with `cd`, `ls`, `pwd`
- Create, modify, and inspect objects with familiar shell commands

## Design Philosophy

### Filesystem-Like Navigation
```bash
resonite> ls
/World
  /Players
  /Environment
  /UI
  
resonite> cd /World/Environment
resonite:/World/Environment> ls
BoxMesh [Component]
Transform [Component]
MeshRenderer [Component]

resonite:/World/Environment> cat BoxMesh
Type: FrooxEngine.BoxMesh
Size: [1, 1, 1]
```

### Everything is Editable
```bash
# Edit slot properties
resonite> edit Slot
Name: MySlot
Position: [0, 1, 0]
Active: true

# Edit component properties  
resonite> edit BoxMesh
Size: [2, 2, 2]
```

### Object-Oriented Operations
```bash
# Create new slot (folder)
resonite> mkdir MyNewObject

# Add component (file)
resonite> touch BoxMesh
resonite> touch Transform

# Copy entire hierarchies
resonite> cp -r /World/Environment/MyObject ./MyObject

# Remove slots/components
resonite> rm BoxMesh
resonite> rmdir MyObject
```

## Architecture

### Core Components

1. **ResoLink Client** (`pkg/resolink/`)
   - WebSocket connection to Resonite
   - Protocol implementation for slot/component operations
   - Request/response handling with timeout support

2. **Virtual Filesystem** (`pkg/vfs/`)
   - Abstraction layer treating slots as directories
   - Components as files within directories
   - Lazy loading and caching of scene graph
   - Path resolution and navigation

3. **REPL Shell** (`pkg/repl/`)
   - Interactive command-line interface
   - Command parsing and execution
   - Tab completion for paths and commands
   - History management

4. **Commands** (`pkg/commands/`)
   - Familiar Unix-like commands: `ls`, `cd`, `pwd`, `cat`, `edit`
   - Resonite-specific commands: `spawn`, `inspect`, `find`
   - Component manipulation: `add-component`, `list-components`
   - Batch operations and scripting support

5. **Object System** (`pkg/objects/`)
   - Slot representation
   - Component representation
   - Property serialization/deserialization
   - Type system integration

### Technology Stack

- **Language**: Go 1.21+
- **WebSocket**: gorilla/websocket
- **REPL**: chzyer/readline or peterh/liner
- **CLI Framework**: spf13/cobra (for command structure)
- **Config**: spf13/viper (for configuration management)

## Project Status

🚧 **In Planning Phase** 🚧

This project is currently in the design and planning stage. See [ARCHITECTURE.md](ARCHITECTURE.md) for detailed design decisions.

## Inspiration

- [resolink-mcp](https://github.com/rassi0429/resolink-mcp) - TypeScript MCP server and CLI for ResoLink
- Unix/Linux shell philosophy - treating everything as files/objects
- Plan9 - everything is a file system

## Roadmap

### Phase 1: Core Infrastructure
- [ ] ResoLink WebSocket client
- [ ] Basic connection management
- [ ] Protocol message handling
- [ ] Error handling and reconnection

### Phase 2: Virtual Filesystem
- [ ] Slot tree representation
- [ ] Component as file abstraction
- [ ] Path resolution
- [ ] Caching layer

### Phase 3: Basic Commands
- [ ] Navigation: `cd`, `ls`, `pwd`, `tree`
- [ ] Inspection: `cat`, `inspect`, `stat`
- [ ] Basic editing: `edit`, `set`

### Phase 4: Advanced Commands
- [ ] Creation: `mkdir`, `touch`, `spawn`
- [ ] Modification: `mv`, `cp`, `rm`
- [ ] Search: `find`, `grep`
- [ ] Batch operations

### Phase 5: REPL Features
- [ ] Interactive prompt
- [ ] Tab completion
- [ ] Command history
- [ ] Syntax highlighting
- [ ] Multi-line editing

### Phase 6: Scripting & Automation
- [ ] Script execution mode
- [ ] Batch command files
- [ ] Variables and substitution
- [ ] Control flow (if/loop)

## Similar Projects

- **resolink-mcp** (TypeScript): MCP server with 21 tools for Resonite
- **resonite-cli** (hypothetical): Focus on listing operations

## Contributing

This project is in early planning. Contributions and design feedback welcome!

## License

TBD (To Be Determined)

## References

- [ResoLink Protocol Documentation](https://deepwiki.com/rassi0429/resolink-mcp)
- [Resonite VR](https://resonite.com/)
