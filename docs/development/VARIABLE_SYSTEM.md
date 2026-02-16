# RESH Variable System Design

## Overview

RESH.DATA will support a full variable system for storing persistent, LogiX-accessible data beyond just bookmarks. This includes primitive types, references, and potentially complex structures.

## Structure

```
/RESH.DATA/
├── Tag: "RESHData"
├── OrderOffset: -9999 (appears last in hierarchy)
├── Components:
│   ├── DynamicVariableSpace (SpaceName="RESH", OnlyDirectBinding=true)
│   └── DynamicReferenceVariable<Slot> (VariableName="World/RESH.DATA", Reference=self)
│
├── Bookmarks/
│   ├── Tag: "RESHData" (inherited)
│   ├── DynamicVariableSpace (SpaceName="bookmark", OnlyDirectBinding=true)
│   └── <BookmarkName>/
│       ├── Tag: "RESHData" (inherited)
│       ├── Persistent: true
│       └── DynamicReferenceVariable<Slot>
│           ├── VariableName: "bookmark/<name>"
│           └── Reference: <target slot>
│
└── Variables/
    ├── Tag: "RESHData" (inherited)
    ├── DynamicVariableSpace (SpaceName="var", OnlyDirectBinding=true)
    └── <VarName>/
        ├── Tag: "RESHData" (inherited)
        ├── Persistent: true
        └── DynamicVariable<T> (one of the types below)
            ├── VariableName: "var/<name>"
            └── Value: <value>
```

## Supported Variable Types

### 1. Primitive Types
- **bool** - `DynamicVariable<bool>`
- **int** - `DynamicVariable<int>`
- **long** - `DynamicVariable<long>`
- **float** - `DynamicVariable<float>`
- **double** - `DynamicVariable<double>`
- **string** - `DynamicVariable<string>`

### 2. Vector Types
- **float2** - `DynamicVariable<float2>`
- **float3** - `DynamicVariable<float3>` (positions, scales)
- **float4** - `DynamicVariable<float4>`
- **floatQ** - `DynamicVariable<floatQ>` (quaternion rotations)

### 3. Color Types
- **color** - `DynamicVariable<color>` (RGBA)
- **colorX** - `DynamicVariable<colorX>` (extended color)

### 4. Reference Types
- **Slot** - `DynamicReferenceVariable<Slot>` (like bookmarks)
- **User** - `DynamicReferenceVariable<User>`
- **Component** - `DynamicReferenceVariable<Component>` (generic)

### 5. URI Types
- **Uri** - `DynamicVariable<Uri>`

## Variable Naming Convention

**Format:** `var/<name>`

Examples:
- `var/playerCount` (int)
- `var/debugMode` (bool)
- `var/spawnPoint` (float3)
- `var/gameState` (string)

**For specialized categories, can use subcategories:**
- `var/game/score` (int)
- `var/config/maxPlayers` (int)
- `var/temp/lastUpdate` (string)

## REPL Commands

### Variable Management Commands

```bash
# List all variables
varlist
varlist <pattern>  # Filter by pattern

# Get variable value
varget <name>
get var/<name>     # Alternative syntax

# Set variable value
varset <name> <type> <value>
set var/<name> <value>         # Auto-detect or preserve type

# Create new variable
varcreate <name> <type> [value]

# Delete variable
vardelete <name>
vardel <name>

# Variable info
varinfo <name>     # Show type, value, component ID
```

### Examples

```bash
# Create variables
/ $ varcreate playerCount int 0
✓ Created variable 'playerCount' (int) = 0

/ $ varcreate debugMode bool true
✓ Created variable 'debugMode' (bool) = true

/ $ varcreate spawnPoint float3 "0,1,0"
✓ Created variable 'spawnPoint' (float3) = (0, 1, 0)

# Get values
/ $ varget playerCount
playerCount (int) = 0

/ $ varlist
Variables (3):
  playerCount (int) = 0
  debugMode (bool) = true
  spawnPoint (float3) = (0, 1, 0)

# Set values
/ $ varset playerCount 5
✓ Set playerCount = 5

/ $ varset debugMode false
✓ Set debugMode = false

# Delete
/ $ vardelete playerCount
✓ Deleted variable 'playerCount'
```

## Lua API

### Reading Variables

```lua
-- Get variable value
local value = get_var("playerCount")
local debug = get_var("debugMode")
local spawn = get_var("spawnPoint")

-- List all variables
local vars = list_vars()
for name, info in pairs(vars) do
    print(name .. " (" .. info.type .. ") = " .. tostring(info.value))
end

-- Check if variable exists
if has_var("debugMode") then
    print("Debug mode is enabled")
end
```

### Writing Variables

```lua
-- Set variable value
set_var("playerCount", 10)
set_var("debugMode", true)
set_var("spawnPoint", {x=0, y=1, z=0})

-- Create new variable
create_var("newVar", "int", 42)
create_var("message", "string", "Hello World")

-- Delete variable
delete_var("oldVar")
```

### Type Conversion

```lua
-- Automatic type detection for common types
set_var("count", 10)           -- int
set_var("ratio", 3.14)         -- float
set_var("enabled", true)       -- bool
set_var("name", "test")        -- string

-- Explicit types for vectors
set_var("position", {x=1, y=2, z=3}, "float3")
set_var("rotation", {x=0, y=0, z=0, w=1}, "floatQ")
```

## Implementation Plan

### Phase 1: Core Variable System
1. **Create variable infrastructure**
   - Functions to create/read/update/delete variable slots
   - Component type mapping (string -> FrooxEngine type)
   - Value serialization/deserialization

2. **REPL commands**
   - `varlist`, `varget`, `varset`, `varcreate`, `vardelete`, `varinfo`
   - Display formatting for different types
   - Tab completion for variable names

3. **Lua API**
   - `get_var()`, `set_var()`, `create_var()`, `delete_var()`
   - `list_vars()`, `has_var()`, `var_info()`
   - Type conversion helpers

### Phase 2: Advanced Features
1. **Variable categories/namespaces**
   - Support `var/category/name` pattern
   - List by category
   - Bulk operations on categories

2. **Variable metadata**
   - Description field
   - Last modified timestamp
   - Access tracking

3. **Import/Export**
   - Export variables to JSON
   - Import from JSON
   - Backup/restore functionality

### Phase 3: Integration
1. **Bookmark integration**
   - Bookmarks are special reference variables
   - Unified `goto` that works with both bookmarks and reference variables
   - Variable can store "last location"

2. **Script helpers**
   - Common patterns as Lua functions
   - State management utilities
   - Session variables vs persistent variables

## Technical Details

### Component Type Names

Full FrooxEngine type names for components:

```go
var DynamicVariableTypes = map[string]string{
    "bool":    "[FrooxEngine]FrooxEngine.DynamicVariable<System.Boolean>",
    "int":     "[FrooxEngine]FrooxEngine.DynamicVariable<System.Int32>",
    "long":    "[FrooxEngine]FrooxEngine.DynamicVariable<System.Int64>",
    "float":   "[FrooxEngine]FrooxEngine.DynamicVariable<System.Single>",
    "double":  "[FrooxEngine]FrooxEngine.DynamicVariable<System.Double>",
    "string":  "[FrooxEngine]FrooxEngine.DynamicVariable<System.String>",
    "float2":  "[FrooxEngine]FrooxEngine.DynamicVariable<[FrooxEngine]Elements.Core.float2>",
    "float3":  "[FrooxEngine]FrooxEngine.DynamicVariable<[FrooxEngine]Elements.Core.float3>",
    "float4":  "[FrooxEngine]FrooxEngine.DynamicVariable<[FrooxEngine]Elements.Core.float4>",
    "floatQ":  "[FrooxEngine]FrooxEngine.DynamicVariable<[FrooxEngine]Elements.Core.floatQ>",
    "color":   "[FrooxEngine]FrooxEngine.DynamicVariable<[FrooxEngine]Elements.Core.color>",
    "colorX":  "[FrooxEngine]FrooxEngine.DynamicVariable<[FrooxEngine]Elements.Core.colorX>",
    "Uri":     "[FrooxEngine]FrooxEngine.DynamicVariable<System.Uri>",
    
    // References
    "Slot":      "[FrooxEngine]FrooxEngine.DynamicReferenceVariable<[FrooxEngine]FrooxEngine.Slot>",
    "User":      "[FrooxEngine]FrooxEngine.DynamicReferenceVariable<[FrooxEngine]FrooxEngine.User>",
    "Component": "[FrooxEngine]FrooxEngine.DynamicReferenceVariable<[FrooxEngine]FrooxEngine.Component>",
}
```

### Value Member Names

- **Value types** use `Value` member
- **Reference types** use `Reference` member

### Creating a Variable

**Algorithm:**
1. Navigate to `/RESH.DATA/Variables/`
2. Create child slot with variable name
3. Add appropriate `DynamicVariable<T>` component
4. Set `VariableName` member to `"var/<name>"`
5. Set `Value` or `Reference` member to initial value
6. Set `Enabled` to `true`
7. Set `persistent` to `true`

### Reading a Variable

**Algorithm:**
1. Navigate to `/RESH.DATA/Variables/<name>/`
2. Find `DynamicVariable<T>` component
3. Read `Value` or `Reference` member
4. Return typed value

### Updating a Variable

**Algorithm:**
1. Navigate to `/RESH.DATA/Variables/<name>/`
2. Find `DynamicVariable<T>` component
3. Update `Value` or `Reference` member
4. Return success

## Usage Examples

### Session State

```bash
# Track session information
/ $ varcreate sessionStart string "2026-02-16T10:30:00Z"
/ $ varcreate visitCount int 0

# Increment on each visit
/ $ varget visitCount
visitCount (int) = 0
/ $ varset visitCount 1
✓ Set visitCount = 1
```

### Game State

```lua
-- Initialize game state
create_var("gameActive", "bool", false)
create_var("playerScore", "int", 0)
create_var("currentLevel", "string", "tutorial")

-- Game loop
if get_var("gameActive") then
    local score = get_var("playerScore")
    set_var("playerScore", score + 10)
end
```

### Configuration

```bash
# Store configuration
/ $ varcreate config/maxPlayers int 16
/ $ varcreate config/difficulty string "normal"
/ $ varcreate config/pvpEnabled bool false

# Query config
/ $ varlist config/*
Variables (3):
  config/maxPlayers (int) = 16
  config/difficulty (string) = normal
  config/pvpEnabled (bool) = false
```

## Security Considerations

1. **Variable naming**: Validate names (no special chars except `/`)
2. **Type safety**: Prevent type mismatches
3. **Access control**: Consider read-only variables (future)
4. **Size limits**: Prevent extremely large strings/data

## Migration Path

### Current Bookmarks (In-Memory)
- Phase 1: Keep in-memory system working
- Phase 2: Add persistent RESH.DATA storage
- Phase 3: Migrate on first use
- Phase 4: Remove in-memory system

### New Variables
- Start with persistent storage from day one
- No in-memory fallback needed

## Benefits

1. **LogiX Integration**: Variables accessible from ProtoFlux
2. **Persistence**: Survive session restarts
3. **Visibility**: Appear in world hierarchy
4. **Standard Pattern**: Follows Resonite conventions
5. **Type Safety**: Proper typed storage
6. **Flexibility**: Support many data types
7. **Scriptable**: Full Lua API access

## Future Enhancements

- Variable watchers (notify on change)
- Variable history/versioning
- Variable permissions/access control
- Variable sync across users
- Variable templates
- Variable validation rules
- Variable documentation/help text
