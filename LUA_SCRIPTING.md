# Lua Scripting API

RESH provides a comprehensive Lua scripting API for automating Resonite world interactions. All RESH-specific functions are organized under the `RESH` namespace to keep the global scope clean and enable library development.

## Quick Start

```lua
-- Navigate to root
RESH.cd("/")

-- List contents
local listing = RESH.ls()
for i, child in ipairs(listing.children) do
    print(child.name)
end

-- Inspect a slot
local data = RESH.inspect("slot-id")
print(data.Name)
```

## Running Scripts

### Interactive REPL
```bash
resh
> :script examples/basic/test.lua
```

### Standalone Execution
```bash
resh script examples/basic/test.lua
```

## The RESH Namespace

All RESH-specific functions are under the `RESH` table to:
- Keep global scope clean
- Prevent naming conflicts
- Enable library development
- Provide clear API boundaries

Standard Lua functions (like `print`) remain global.

## Available Functions

### RESH.cd(target)

Navigate to a different slot in the hierarchy.

**Parameters:**
- `target` (string): Path to navigate to
  - Absolute: `"/RESH.DATA/Bookmarks"`
  - Relative: `"child_slot"`
  - Parent: `".."`
  - Root: `"/"`

**Returns:** boolean - true on success, false on failure

**Example:**
```lua
-- Absolute path
RESH.cd("/RESH.DATA")

-- Relative navigation
RESH.cd("Bookmarks")
RESH.cd("..")

-- Root
RESH.cd("/")
```

### RESH.ls()

List children and components of the current slot.

**Returns:** table with:
- `children` (array): Child slots
  - `id` (string): Slot ID
  - `name` (string): Slot name
  - `active` (boolean): Active state
- `components` (array): Components on slot
  - `id` (string): Component ID
  - `type` (string): Component type name

**Example:**
```lua
local listing = RESH.ls()

-- List all children
for i, child in ipairs(listing.children) do
    local status = child.active and "active" or "inactive"
    print(child.name .. " [" .. status .. "]")
end

-- List all components
for i, comp in ipairs(listing.components) do
    print(comp.type)
end
```

### RESH.inspect(id)

Get detailed information about a slot or component.

**Parameters:**
- `id` (string): ID of slot or component to inspect

**Returns:** table with slot or component data

#### Slot Inspection

Returns a table with direct property access:
- `Name` (string): Slot name
- `Active` (boolean): Active state
- `Persistent` (boolean): Persistence state
- `Tag` (string|nil): Tag value
- `Position` (table): Position vector
- `Rotation` (table): Rotation quaternion
- `Scale` (table): Scale vector
- `OrderOffset` (number): Hierarchy ordering
- `ActiveSelf` (boolean): Self active state

**Example:**
```lua
local slot = RESH.inspect("slot-id")
print("Name: " .. slot.Name)
print("Active: " .. tostring(slot.Active))
print("Tag: " .. (slot.Tag or "<none>"))
```

#### Component Inspection

Returns a table with:
- `TypeName` (string): Short component type
- `ComponentType` (string): Full component type
- `Members` (table): Component members indexed by name

Each member in `Members` has:
- `Type` (string): Member type
- `Value` (any): Type-aware value (bool, number, string, table)
- `ID` (string): Member ID
- `TargetId` (string|nil): For reference types, the referenced ID

**Example:**
```lua
local comp = RESH.inspect("comp-id")
print("Type: " .. comp.TypeName)

-- Access members by name
for name, member in pairs(comp.Members) do
    print(name .. " = " .. tostring(member.Value))
    
    -- Check type-specific data
    if member.Type == "reference" then
        print("  References: " .. (member.TargetId or "<null>"))
    end
end
```

### RESH.pwd()

Get the current path in the slot hierarchy.

**Returns:** string - current path

**Example:**
```lua
local path = RESH.pwd()
print("Current location: " .. path)
```

### RESH.get_current_slot()

Get the ID of the current slot.

**Returns:** string - current slot ID

**Example:**
```lua
local id = RESH.get_current_slot()
print("Current slot ID: " .. id)

-- Inspect current slot
local data = RESH.inspect(id)
```

### RESH.find_slot(name)

Find a child slot by name in the current directory.

**Parameters:**
- `name` (string): Name of slot to find

**Returns:** string|nil - slot ID if found, nil otherwise

**Example:**
```lua
local resh_id = RESH.find_slot("RESH.DATA")
if resh_id then
    print("Found RESH.DATA: " .. resh_id)
    RESH.cd("RESH.DATA")
else
    print("RESH.DATA not found")
end
```

## Data Structures

### Type-Aware Values

Component member values are automatically converted to appropriate Lua types:

- **Booleans**: `true`/`false` (not strings)
- **Numbers**: Lua numbers (integers and floats)
- **Strings**: Lua strings
- **References**: Tables with `TargetId` field
- **Complex types**: Nested tables

**Example:**
```lua
local comp = RESH.inspect("comp-id")
local member = comp.Members["Enabled"]

-- Direct boolean comparison (not string!)
if member.Value then
    print("Component is enabled")
end

-- Number arithmetic
local count = comp.Members["Count"].Value
print("Double: " .. (count * 2))
```

### Member Access by Name

Component members are indexed by name for easy access:

```lua
local comp = RESH.inspect("comp-id")

-- Access specific member directly
local enabled = comp.Members["Enabled"]
local space_name = comp.Members["SpaceName"]

-- Iterate all members
for name, member in pairs(comp.Members) do
    print(name .. ": " .. tostring(member.Value))
end
```

## Common Patterns

### Navigate to RESH.DATA

```lua
RESH.cd("/")
local resh_id = RESH.find_slot("RESH.DATA")
if resh_id then
    RESH.cd("RESH.DATA")
else
    error("RESH.DATA not found!")
end
```

### List All Bookmarks

```lua
RESH.cd("/RESH.DATA/Bookmarks")
local listing = RESH.ls()

for i, child in ipairs(listing.children) do
    print("Bookmark: " .. child.name)
    
    -- Find the reference variable
    RESH.cd(child.name)
    local child_listing = RESH.ls()
    
    for j, comp in ipairs(child_listing.components) do
        if comp.type:match("DynamicReferenceVariable") then
            local var = RESH.inspect(comp.id)
            local ref_id = var.Members["Reference"].TargetId
            print("  -> Points to: " .. (ref_id or "<null>"))
        end
    end
    
    RESH.cd("..")
end
```

### Find Component by Type

```lua
local listing = RESH.ls()

for i, comp in ipairs(listing.components) do
    if comp.type:match("DynamicVariableSpace") then
        local data = RESH.inspect(comp.id)
        local space_name = data.Members["SpaceName"].Value
        print("Found space: " .. space_name)
    end
end
```

### Recursive Slot Traversal

```lua
function traverse(depth)
    depth = depth or 0
    local indent = string.rep("  ", depth)
    
    local listing = RESH.ls()
    local path = RESH.pwd()
    
    print(indent .. path)
    
    for i, child in ipairs(listing.children) do
        RESH.cd(child.name)
        traverse(depth + 1)
        RESH.cd("..")
    end
end

RESH.cd("/")
traverse()
```

## Standard Lua Functions

Standard Lua functions remain in the global scope:

- `print(...)`: Output text
- `string.*`: String manipulation
- `table.*`: Table operations
- `math.*`: Mathematical functions
- `io.*`: File I/O (limited in sandboxed environments)
- And all other standard Lua libraries

## Examples

See the `examples/` directory for complete working examples:

- `examples/basic/test.lua` - Basic API usage
- `examples/basic/demo_lua_interface.lua` - Data structure demonstrations
- `examples/resh_data/inspect_resh_data.lua` - RESH.DATA structure inspection
- `examples/resh_data/explore_structure.lua` - Comprehensive structure explorer
- `examples/resh_data/examine_bookmarks.lua` - Bookmark system deep dive

## Best Practices

1. **Check for nil**: Always check `find_slot()` returns before using
2. **Type awareness**: Remember values are properly typed (booleans, numbers)
3. **Member names**: Use `pairs()` to iterate members, access by name
4. **Navigation**: Use absolute paths when possible for reliability
5. **Error handling**: Wrap operations in `pcall()` for production scripts
6. **Namespace**: Always use `RESH.*` for RESH functions

## Limitations

- Read-only access (write operations coming in future updates)
- Synchronous execution (blocking)
- Limited to current world connection
- No concurrent script execution

## Future API Additions

Planned additions to the RESH namespace:

- `RESH.get_var(name)` - Get variable value
- `RESH.set_var(name, value)` - Set variable value
- `RESH.create_var(name, type, value)` - Create variable
- `RESH.delete_var(name)` - Delete variable
- `RESH.list_vars()` - List all variables
- `RESH.add_slot(name, parent)` - Create slot
- `RESH.delete_slot(id)` - Delete slot
- `RESH.add_component(slot_id, type)` - Add component
- `RESH.update_component(id, members)` - Update component
- `RESH.delete_component(id)` - Delete component

See [docs/development/VARIABLE_SYSTEM.md](docs/development/VARIABLE_SYSTEM.md) for variable system design.
