# RESH Lua Script Examples

This directory contains example Lua scripts demonstrating various RESH capabilities.

## Quick Start

Run any script using:
```bash
# From REPL
/ $ script examples/basic/hello.lua

# Standalone mode
resh script examples/basic/hello.lua --url ws://localhost:39015
```

## Examples by Category

### Basic (`basic/`)

#### `test.lua`
Basic functionality test demonstrating:
- Navigation (`cd`, `pwd`)
- Listing (`ls`)
- Getting current location
- Finding slots

**Usage:**
```bash
resh script examples/basic/test.lua --url ws://localhost:39015
```

#### `demo_lua_interface.lua`
Comprehensive demo of the Lua data interface showing:
- Component inspection with type-aware values
- Slot property access
- Member access by name (not index)
- Reference handling
- Complex value structures

**Usage:**
```bash
resh script examples/basic/demo_lua_interface.lua --url ws://localhost:39015
```

**Good for:** Understanding how data is structured in Lua scripts

### RESH.DATA Management (`resh_data/`)

#### `inspect_resh_data.lua`
Inspects RESH.DATA structure and reports:
- Slot properties (Active, Persistent, Tag)
- All components and their members
- Bookmarks child slot structure
- DynamicVariableSpace and DynamicReferenceVariable components

**Usage:**
```bash
resh script examples/resh_data/inspect_resh_data.lua --url ws://localhost:39015
```

**Requirements:** Expects `/RESH.DATA` slot to exist in world

## Creating Your Own Scripts

### Script Template

```lua
-- My script description
print("Starting my script...")

-- Navigate to target
cd("/World/MySlot")

-- List contents
local listing = ls()
print("Found " .. #listing.children .. " children")

-- Inspect components
for _, comp in ipairs(listing.components) do
    print("Component: " .. comp.type)
    
    local data = inspect(comp.id)
    -- Access members by name
    if data.Members.Enabled then
        print("  Enabled: " .. tostring(data.Members.Enabled.Value))
    end
end

print("Script complete!")
```

### Available Functions

See [LUA_SCRIPTING.md](../LUA_SCRIPTING.md) for complete API documentation:

- **Navigation:** `cd()`, `pwd()`, `get_current_slot()`
- **Discovery:** `ls()`, `find_slot()`
- **Inspection:** `inspect()`
- **Output:** `print()`

### Data Structures

**Component Data:**
```lua
{
    Type = "component",
    ID = "Reso_123",
    ComponentType = "[FrooxEngine]FrooxEngine.BoxCollider",
    TypeName = "BoxCollider",
    Members = {
        Enabled = {
            ID = "Reso_124",
            Name = "Enabled",
            Type = "bool",
            Value = true
        },
        Size = {
            Type = "float3",
            Value = {x = 1.0, y = 1.0, z = 1.0}
        }
    }
}
```

**Slot Data:**
```lua
{
    Type = "slot",
    ID = "Reso_100",
    Name = "MySlot",
    Active = true,
    Persistent = false,
    Tag = "MyTag",
    Position = {x = 0, y = 1, z = 0}
}
```

## Tips

### Debugging

Add verbose output:
```lua
print("Current location: " .. pwd())
print("Current slot: " .. get_current_slot())
```

### Error Handling

Check for nil values:
```lua
local slot_id = find_slot("MySlot")
if not slot_id then
    print("ERROR: Slot not found!")
    return
end
```

### Iterating Components

Use `pairs()` for members (map), `ipairs()` for arrays:
```lua
-- Iterate members by name
for name, member in pairs(comp_data.Members) do
    print(name .. " = " .. tostring(member.Value))
end

-- Iterate component list
for i, comp in ipairs(listing.components) do
    print("[" .. i .. "] " .. comp.type)
end
```

### Working with References

References have a special `TargetId` field:
```lua
if member.Type == "reference" then
    if member.TargetId then
        print("Points to: " .. member.TargetId)
    else
        print("Null reference")
    end
end
```

## Contributing Examples

Have a useful script? Add it to the appropriate category:

1. Create your script file
2. Add description to this README
3. Test it works with test server
4. Submit a pull request

Categories for organization:
- `basic/` - Simple, single-feature demonstrations
- `advanced/` - Complex multi-step operations (coming soon)
- `resh_data/` - RESH.DATA management scripts
- `world_analysis/` - Scripts for analyzing world structure (coming soon)
- `automation/` - Automated tasks and workflows (coming soon)

## More Information

- [Lua Scripting API Documentation](../LUA_SCRIPTING.md)
- [Contributing Guide](../CONTRIBUTING.md)
- [Quick Start Guide](../QUICKSTART.md)
