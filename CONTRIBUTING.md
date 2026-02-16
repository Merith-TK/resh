# Contributing to resonite-sh

Thank you for your interest in contributing to resonite-sh! This document will help you understand the codebase structure and development workflow.

## Table of Contents

- [Code Organization](#code-organization)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Code Style](#code-style)
- [Adding New Features](#adding-new-features)
- [Testing](#testing)
- [Documentation](#documentation)

## Code Organization

```
resonite-sh/
├── main.go                    # Entry point
├── cmd/                       # CLI commands & UI layer
│   ├── root.go               # Cobra root command & config
│   ├── repl.go               # Interactive REPL
│   ├── script.go             # Script execution
│   ├── commands.go           # Command implementations
│   ├── display.go            # Output formatting
│   ├── autocomplete.go       # Tab completion
│   └── parser.go             # Command parsing
│
├── pkg/                       # Core libraries
│   ├── resolink/             # ResoLink WebSocket protocol
│   ├── shell/                # Shell business logic
│   └── resh/                 # RESH.DATA management
│
├── examples/                  # Example Lua scripts
├── docs/                      # Documentation
└── .test-server/             # Docker test environment
```

### Package Responsibilities

- **`cmd/`** - User interface layer. Handles CLI parsing, REPL loop, output formatting. Should NOT contain business logic.
- **`pkg/resolink/`** - Pure protocol layer. WebSocket communication with Resonite. No business logic, just message send/receive.
- **`pkg/shell/`** - Business logic layer. Orchestrates resolink calls, manages state, implements shell operations.
- **`pkg/resh/`** - RESH.DATA specific logic. Variable management and storage.

### Dependency Flow

```
cmd/ → pkg/shell/ → pkg/resolink/ → WebSocket
     → pkg/resh/  ↗
```

**Important:** Dependencies should only flow downward. Never import `cmd/` from `pkg/`.

## Development Setup

### Prerequisites

- Go 1.21 or later
- Docker (for test server)
- Resonite VR with ResoLink mod installed

### Initial Setup

```bash
# Clone repository
git clone https://github.com/Merith-TK/resh.git
cd resonite-sh

# Install dependencies
go mod download

# Build
go build -o resh.exe .

# Run tests
go test ./...
```

### Running Test Server

```bash
cd .test-server
docker compose up
```

This starts a mock ResoLink server on `ws://localhost:39015` for testing.

## Making Changes

### Workflow

1. **Create a branch** for your feature/fix
   ```bash
   git checkout -b feature/my-new-feature
   ```

2. **Make your changes** following the code style guide

3. **Test your changes**
   ```bash
   go build -o resh.exe .
   ./resh.exe repl --url ws://localhost:39015
   ```

4. **Update documentation** if needed

5. **Commit with clear messages**
   ```bash
   git commit -m "feat: add new command for X"
   ```

6. **Push and create pull request**

### Commit Message Format

Follow conventional commits:

```
<type>: <description>

[optional body]
[optional footer]
```

Types:
- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `refactor:` - Code refactoring
- `test:` - Test additions/changes
- `chore:` - Maintenance tasks

Examples:
```
feat: add set command for editing components
fix: handle null references in inspect output
docs: update LUA_SCRIPTING.md with new functions
refactor: split commands.go into separate files
```

## Code Style

### Go Conventions

Follow standard Go conventions:
- Use `gofmt` for formatting (automatic in most editors)
- Use meaningful variable names
- Add comments for exported functions
- Keep functions focused and small
- Handle errors explicitly

### Naming Conventions

- **Files:** `snake_case.go` (e.g., `component_operations.go`)
- **Packages:** `lowercase` (e.g., `resolink`, `shell`)
- **Functions:** `PascalCase` for exported, `camelCase` for private
- **Constants:** `PascalCase` (e.g., `RESHSlotName`)

### Error Handling

Always handle errors explicitly:

```go
// Good
data, err := shell.InspectComponent(client, id)
if err != nil {
    return fmt.Errorf("failed to inspect component: %w", err)
}

// Bad - never ignore errors
data, _ := shell.InspectComponent(client, id)
```

Use `%w` for error wrapping to maintain error chains.

### Comments

Add godoc comments for exported functions:

```go
// InspectComponent retrieves full component data for inspection.
// It handles both display ID format (ID_xxx) and internal format (Reso_xxx).
// Returns ComponentData with all members parsed, or error if component not found.
func InspectComponent(client *resolink.Client, componentID string) (*ComponentData, error) {
    // ...
}
```

## Adding New Features

### Adding a New Command

1. **Add command handler in `cmd/commands.go`:**

```go
func myNewCommand(client *resolink.Client, state *shell.State, args []string) {
    // Validate args
    if len(args) == 0 {
        fmt.Println("mycommand: missing argument")
        return
    }
    
    // Call business logic
    result, err := shell.DoSomething(client, state, args[0])
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    // Display result
    displayResult(result)
}
```

2. **Add business logic in `pkg/shell/`:**

```go
// DoSomething performs the operation
func DoSomething(client *resolink.Client, state *State, arg string) (*Result, error) {
    // Use resolink client to communicate with Resonite
    resp, err := client.GetComponent(arg)
    if err != nil {
        return nil, fmt.Errorf("failed to get component: %w", err)
    }
    
    // Process and return
    return &Result{Data: resp.Data}, nil
}
```

3. **Register in `cmd/repl.go`:**

```go
case "mycommand":
    myNewCommand(client, state, args)
```

4. **Add to help in `cmd/commands.go`:**

```go
fmt.Println("  mycommand <arg>   - Description of what it does")
```

5. **Add autocomplete in `cmd/autocomplete.go`** if needed

### Adding a Lua Function

1. **Add to `pkg/shell/script.go` in `registerShellFunctions`:**

```go
// my_function - description
L.SetGlobal("my_function", L.NewFunction(func(L *lua.LState) int {
    arg := L.CheckString(1)
    
    // Do the operation
    result, err := DoSomething(ctx.Client, ctx.State, arg)
    if err != nil {
        L.Push(lua.LNil)
        L.Push(lua.LString(err.Error()))
        return 2
    }
    
    // Convert result to Lua value
    luaResult := convertToLuaValue(L, result)
    L.Push(luaResult)
    return 1
}))
```

2. **Document in `LUA_SCRIPTING.md`:**

```markdown
#### `my_function(arg)`
Description of what it does.
- Parameters: `arg` (string) - What the argument is
- Returns: `table` - Structure of return value
```

3. **Add example script in `examples/`**

### Adding a New ResoLink Message Type

1. **Define in `pkg/resolink/messages.go`:**

```go
type MyNewMessage struct {
    BaseMessage
    SomeField string `json:"someField"`
}
```

2. **Add client method in `pkg/resolink/client.go` or appropriate file:**

```go
func (c *Client) DoNewThing(field string) (*MyResponse, error) {
    msg := &MyNewMessage{
        BaseMessage: BaseMessage{
            Type: "MyNewMessage",
        },
        SomeField: field,
    }
    
    rawResp, err := c.sendMessage(msg)
    if err != nil {
        return nil, err
    }
    
    var resp MyResponse
    if err := json.Unmarshal(rawResp, &resp); err != nil {
        return nil, fmt.Errorf("failed to parse response: %w", err)
    }
    
    return &resp, nil
}
```

3. **Update switch in `sendMessage` to include your message type**

## Testing

### Manual Testing

1. **Start test server:**
   ```bash
   cd .test-server
   docker compose up
   ```

2. **Run REPL:**
   ```bash
   go run . repl --url ws://localhost:39015
   ```

3. **Test your changes interactively**

### Automated Testing

Add unit tests for business logic:

```go
// component_operations_test.go
func TestInspectComponent(t *testing.T) {
    // Setup
    client := setupTestClient(t)
    
    // Test
    data, err := InspectComponent(client, "Reso_123")
    
    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if data.ID != "Reso_123" {
        t.Errorf("expected ID Reso_123, got %s", data.ID)
    }
}
```

Run tests:
```bash
go test ./...
go test -v ./pkg/shell/  # Verbose output for specific package
```

### Testing Lua Scripts

Create test scripts in `examples/basic/`:

```lua
-- test_my_feature.lua
print("Testing my feature...")

local result = my_function("test")
if result then
    print("✓ Success")
else
    print("✗ Failed")
end
```

Run:
```bash
./resh.exe script examples/basic/test_my_feature.lua --url ws://localhost:39015
```

## Documentation

### What to Document

- **User-facing features** → Update `README.md` and `QUICKSTART.md`
- **Lua functions** → Update `LUA_SCRIPTING.md`
- **Architecture changes** → Update `docs/development/ARCHITECTURE.md`
- **Breaking changes** → Add to `CHANGELOG.md`
- **Design decisions** → Document in `docs/development/DESIGN_DECISIONS.md`

### Documentation Style

- Use clear, concise language
- Include code examples
- Add usage examples for new commands
- Keep README.md up to date with current features

### Creating Examples

When adding new features, create example scripts:

```
examples/
├── basic/
│   └── using_my_feature.lua
└── advanced/
    └── complex_my_feature.lua
```

Add README.md in examples/ describing each script.

## Pull Request Process

1. **Update documentation** for your changes
2. **Add tests** if applicable
3. **Run `gofmt`** to format code
4. **Test manually** with real Resonite or test server
5. **Update CHANGELOG.md** with your changes
6. **Create PR** with clear description

### PR Description Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing Done
- [ ] Manual testing with test server
- [ ] Manual testing with Resonite
- [ ] Unit tests added/updated
- [ ] Example scripts tested

## Checklist
- [ ] Code follows project style
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] No new warnings or errors
```

## Questions?

- Check `docs/` for more detailed documentation
- Look at existing code for patterns
- Ask in issues or discussions

Thank you for contributing! 🎉
