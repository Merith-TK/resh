# RESH Lua Scripting Interface

## Overview

RESH supports Lua scripting for automation and testing. Scripts can navigate, inspect, and manipulate Resonite worlds programmatically.

## Running Scripts

### From REPL
```bash
/ $ script my_script.lua
```

### Standalone Mode (NEW!)
```bash
resh script my_script.lua --url ws://localhost:39015
```

The standalone mode connects, runs the script, and exits - perfect for CI/CD, automation, and tooling integration.

## Available Functions

### Navigation

#### `cd(target)`
Navigate to a different slot.
- `cd("/")` - Go to root
- `cd("..")` - Go to parent
- `cd("SlotName")` - Go to child by name
- Returns: `true` on success, `false, error_message` on failure

#### `pwd()`
Get current path.
- Returns: `string` - Current path (e.g., "/World/Avatar")

#### `get_current_slot()`
Get current slot ID.
- Returns: `string` - Current slot ID (e.g., "Reso_123")

### Listing & Discovery

#### `ls()`
List current slot contents.
- Returns: `table` with structure:
```lua
{
    children = {
        {id = "Reso_1", name = "Child1", active = true, persistent = false},
        {id = "Reso_2", name = "Child2", active = false, persistent = true},
        ...
    },
    components = {
        {id = "Reso_10", type = "FrooxEngine.BoxCollider", persistent = false},
        {id = "Reso_11", type = "FrooxEngine.MeshRenderer", persistent = true},
        ...
    }
}
```

#### `find_slot(name)`
Search for child slot by name.
- Parameters: `name` (string) - Slot name to find
- Returns: `string` - Slot ID if found, `nil` if not found

### Inspection

#### `inspect(id)`
Get detailed information about a component or slot.

**For Components:**
```lua
{
    Type = "component",
    ID = "Reso_123",
    ComponentType = "[FrooxEngine]FrooxEngine.BoxCollider",
    TypeName = "BoxCollider",
    Members = {
        -- Members indexed by NAME for easy access
        Enabled = {
            ID = "Reso_124",
            Name = "Enabled",
            Type = "bool",
            Value = true
        },
        Size = {
            ID = "Reso_125",
            Name = "Size",
            Type = "float3",
            Value = {x = 1.0, y = 1.0, z = 1.0}
        },
        Slot = {
            ID = "Reso_126",
            Name = "Slot",
            Type = "reference",
            Value = {targetId = "Reso_100"},
            TargetId = "Reso_100"  -- Convenience field for references
        }
    }
}
```

**For Slots:**
```lua
{
    Type = "slot",
    ID = "Reso_100",
    Name = "MySlot",
    Active = true,
    Persistent = false,
    Tag = "MyTag",
    Position = {x = 0, y = 1, z = 0},
    Rotation = {x = 0, y = 0, z = 0, w = 1},
    Scale = {x = 1, y = 1, z = 1},
    OrderOffset = 0,
    Parent = "Reso_99"
}
```

### Output

#### `print(...)`
Print output (captured and displayed after script completion).
- Parameters: Any number of values
- Returns: nothing

## Data Type Mapping

Go values are automatically converted to appropriate Lua types:

| Go Type | Lua Type | Notes |
|---------|----------|-------|
| `bool` | `boolean` | Direct conversion |
| `int`, `int64`, `float32`, `float64` | `number` | All numeric types become Lua numbers |
| `string` | `string` | Direct conversion |
| `map[string]interface{}` | `table` | Keys become table keys |
| `[]interface{}` | `table` | Array-like table (1-indexed) |
| `nil` | `nil` | Null values |

## Key Improvements (vs Old Interface)

### 1. Members Indexed by Name
**OLD (array-based):**
```lua
for i, member in ipairs(comp_data.members) do
    if member.name == "Enabled" then
        -- found it
    end
end
```

**NEW (map-based):**
```lua
local enabled = comp_data.Members.Enabled
print(enabled.Value)  -- Direct access!
```

### 2. Slot Properties Direct Access
**OLD (array of properties):**
```lua
for i, prop in ipairs(slot_data.properties) do
    if prop.name == "Name" then
        print(prop.value)
    end
end
```

**NEW (direct fields):**
```lua
print(slot_data.Name)
print(slot_data.Active)
print(slot_data.Position.x)
```

### 3. Type-Aware Values
Values preserve their types instead of being stringified:
- Booleans are `true`/`false` (not `"true"`/`"false"`)
- Numbers are numbers (not strings)
- Complex structures (float3, floatQ) are tables with fields
- References have convenient `TargetId` field

### 4. Consistent Naming
All table fields use PascalCase to match Resonite conventions:
- `comp_data.Members` (not `members`)
- `slot_data.Name` (not `name`)
- `member.TargetId` (not `targetId`)

## Example Scripts

### Navigate and Inspect
```lua
cd("/")
local listing = ls()

for _, child in ipairs(listing.children) do
    print("Child: " .. child.name)
    cd(child.name)
    
    local components = ls().components
    print("  Has " .. #components .. " components")
    
    cd("..")
end
```

### Find and Inspect Component
```lua
cd("/World/Avatar")
local listing = ls()

for _, comp in ipairs(listing.components) do
    if string.match(comp.type, "BoxCollider") then
        local data = inspect(comp.id)
        print("Found BoxCollider:")
        print("  Enabled: " .. tostring(data.Members.Enabled.Value))
        print("  Size: " .. tostring(data.Members.Size.Value))
    end
end
```

### Check References
```lua
local comp_data = inspect("Reso_123")

for name, member in pairs(comp_data.Members) do
    if member.Type == "reference" then
        if member.TargetId then
            print(name .. " points to " .. member.TargetId)
        else
            print(name .. " is null")
        end
    end
end
```

## Error Handling

Functions return multiple values on error:
```lua
local success, error = cd("NonExistent")
if not success then
    print("Error: " .. error)
end

local data, error = inspect("invalid_id")
if not data then
    print("Error: " .. error)
end
```

## Best Practices

1. **Check for nil before accessing nested fields:**
```lua
if slot_data.Position then
    print(slot_data.Position.x)
end
```

2. **Use pairs() for member tables:**
```lua
for name, member in pairs(comp_data.Members) do
    -- name is the member name
end
```

3. **Use ipairs() for arrays:**
```lua
for i, child in ipairs(listing.children) do
    -- i is 1-indexed
end
```

4. **Always handle reference null values:**
```lua
local target = member.TargetId
if target then
    -- reference is valid
else
    -- reference is null
end
```
