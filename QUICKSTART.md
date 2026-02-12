# Quick Reference Guide

## For Developers

### Project Overview
**resonite-sh**: TUI/REPL shell for Resonite VR via ResoniteLink WebSocket API

**Key Concept**: Treat Resonite world as a filesystem
- Slots = Folders
- Components = Files
- Navigate with cd/ls or visual tree
- Edit properties inline or with external editor

### Running the Project

```bash
# Get dependencies
go mod download

# Run REPL mode
go run . repl

# Run with custom URL
go run . repl --url ws://localhost:29551

# Build
go build -o resonite-sh
```

### Key Architectural Decisions

1. **Dual Mode System**
   - TUI (Bubble Tea) - Visual tree + inspector
   - REPL - Shell commands
   - Tab key switches modes

2. **Variable Storage**
   - Session: In-memory (lost on exit)
   - Local: `~/.resonite-sh/vars.yaml`
   - World: RESH slot with DynamicVariableSpace

3. **Updates**
   - No push notifications from ResoniteLink
   - Poll focused slot at 0.1s interval
   - Smart polling (slow when idle)

4. **Component Types**
   - Query once on connect
   - Cache locally
   - User aliases in config
   - Fuzzy matching

### Important Files

- `PROTOCOL.md` - ResoniteLink protocol reference
- `DESIGN_DECISIONS.md` - Full design specification
- `ARCHITECTURE.md` - System architecture
- `ROADMAP.md` - Implementation timeline
- `submodules/ResoniteLink/` - Official protocol repo

### Development Workflow

```bash
# 1. Make changes
vim pkg/resolink/client.go

# 2. Test
go test ./pkg/resolink/...

# 3. Run
go run . repl

# 4. Format
go fmt ./...

# 5. Lint (optional)
golangci-lint run
```

### Key Protocol Messages

```go
// Get slot
{
    "$type": "getSlot",
    "slotId": "Root",
    "includeComponentData": false,
    "depth": 0
}

// Add slot
{
    "$type": "addSlot",
    "data": {
        "parent": {"$type": "reference", "targetId": "Root"},
        "name": {"$type": "string", "value": "MySlot"}
    }
}

// Add component
{
    "$type": "addComponent",
    "containerSlotId": "MySlot",
    "data": {
        "componentType": "[FrooxEngine]FrooxEngine.BoxMesh",
        "members": {
            "Size": {"$type": "float3", "value": {"x": 1, "y": 1, "z": 1}}
        }
    }
}
```

### Common Tasks

#### Adding a new command

1. Add command struct to `pkg/repl/commands.go`
2. Register in `executeCommand()` switch
3. Implement command logic
4. Add help text
5. Test

#### Adding a new TUI panel

1. Create file in `pkg/tui/`
2. Implement Bubble Tea Model interface
3. Add to main app composition
4. Wire up messages

#### Adding Lua API function

1. Add Go function in `pkg/lua/api.go`
2. Register in Lua state
3. Document in design docs
4. Add example script

### Testing Strategy

```bash
# Unit tests
go test ./pkg/resolink/...
go test ./pkg/vfs/...

# Integration tests (needs mock WebSocket)
go test ./pkg/... -tags=integration

# Manual testing
# 1. Start Resonite with ResoniteLink enabled
# 2. Run: go run . repl
# 3. Test commands
```

### Common Issues

**Can't connect**:
- Check Resonite is running
- Check ResoniteLink is enabled in settings
- Verify URL: `ws://localhost:29551`

**Protocol errors**:
- Check PROTOCOL.md for correct message format
- Verify messageId is unique
- Check component type format: `[Assembly]Namespace.Class`

**RefID issues**:
- RefIDs are strings starting with "ID"
- Root slot: use "Root" string or "ID2300"
- Can provide custom IDs or let Resonite generate

### Useful Commands (when REPL works)

```bash
# Navigation
pwd                    # Show current path
ls                     # List current directory
cd /World/Environment  # Change directory
tree                   # Show hierarchy

# Inspection
cat BoxMesh            # Show component properties
stat MySlot            # Show slot details

# Editing
edit BoxMesh           # Edit component (external editor)
set BoxMesh.Size 2,2,2 # Set property inline
mkdir NewSlot          # Create slot
touch BoxMesh          # Add component
rm BoxMesh             # Remove component

# Scripting
:lua scripts/test.lua     # Run script file
:lua VAR.MyScript         # Run script from variable
:lua print("test")        # Inline Lua

# Variables
# In Lua:
RESH.VARS.MyVar = 123
print(RESH.VARS.MyVar)
```

### Code Style

- Use `gofmt` for formatting
- Follow standard Go conventions
- Comment exported functions
- Use meaningful variable names
- Keep functions small and focused

### Git Workflow

```bash
# Create feature branch
git checkout -b feature/new-command

# Make changes, commit
git add .
git commit -m "Add ls command with RefID display"

# Push and create PR
git push origin feature/new-command
```

### Performance Tips

- Use lazy loading for tree (load on expand)
- Cache slot/component data
- Use batch operations when possible
- Virtual scrolling for long lists
- Debounce rapid updates

### Debugging

```go
// Add debug prints
fmt.Fprintf(os.Stderr, "Debug: %+v\n", data)

// Use delve debugger
dlv debug
(dlv) break pkg/resolink/client.go:45
(dlv) continue
```

### Resources

- [Go Documentation](https://go.dev/doc/)
- [Bubble Tea Tutorial](https://github.com/charmbracelet/bubbletea/tree/master/tutorials)
- [gopher-lua Examples](https://github.com/yuin/gopher-lua#examples)
- [Resonite Documentation](https://wiki.resonite.com/)
- [ResoniteLink Docs](https://yellow-dog-man.github.io/ResoniteLink/)

## Quick Start Checklist

- [ ] Clone repo
- [ ] Run `go mod download`
- [ ] Start Resonite with ResoniteLink enabled
- [ ] Run `go run . repl`
- [ ] Try basic commands (when implemented)
- [ ] Read PROTOCOL.md
- [ ] Check TODO.md for tasks
- [ ] Pick a task and start coding!
