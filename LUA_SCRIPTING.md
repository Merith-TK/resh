# Lua Scripting Guide

RESH provides a Lua scripting API for automating Resonite world interactions. All RESH functions are in the `RESH` namespace (e.g., `RESH.cd()`, `RESH.ls()`).

## Quick Start

```lua
-- Navigate and list
RESH.cd("/")
local listing = RESH.ls()
for i, child in ipairs(listing.children) do
    print(child.name)
end

-- Inspect a slot
local slot = RESH.inspect("slot-id")
print(slot.Name)
```

## Running Scripts

```bash
# Interactive REPL
resh repl
> :script examples/basic/test.lua

# Standalone execution
resh script examples/basic/test.lua --url ws://localhost:55083
```

---

## API Functions

### Navigation

#### RESH.cd(path)
Navigate to a slot in the hierarchy.

```lua
RESH.cd("/")                    -- Go to root
RESH.cd("/RESH.DATA")          -- Absolute path
RESH.cd("ChildSlot")           -- Relative path
RESH.cd("..")                  -- Parent slot

-- Returns: success (boolean), error (string|nil)
local ok, err = RESH.cd("SomeSlot")
if not ok then
    print("Failed: " .. err)
end
```

#### RESH.pwd()
Get current path.

```lua
local path = RESH.pwd()  -- Returns: "/RESH.DATA/Bookmarks"
```

#### RESH.get_current_slot()
Get current slot ID.

```lua
local id = RESH.get_current_slot()  -- Returns: "slot-id-string"
```

#### RESH.find_slot(name)
Find child slot by name.

```lua
local id = RESH.find_slot("RESH.DATA")  -- Returns: id or nil
if id then
    RESH.cd("RESH.DATA")
end
```

---

### Inspection

#### RESH.ls()
List children and components of current slot.

```lua
local listing = RESH.ls()

-- Returns:
-- {
--   children = { {id="...", name="...", active=true, persistent=false}, ... },
--   components = { {id="...", type="FrooxEngine.Type", persistent=false}, ... }
-- }

-- Usage:
for i, child in ipairs(listing.children) do
    print(child.name .. " [" .. (child.active and "ON" or "OFF") .. "]")
end

for i, comp in ipairs(listing.components) do
    print(comp.type)
end
```

#### RESH.inspect(id)
Get detailed data about a slot or component.

```lua
local data = RESH.inspect("some-id")

-- For slots, returns properties directly:
-- { Name="...", Active=true, Tag="...", Position={x,y,z}, ... }

-- For components, returns members by name:
-- { TypeName="...", Members={MemberName={Value=..., Type="...", ID="..."}, ...} }
```

---

## Reading Variables

Component members and slot properties are **type-aware** - booleans are `true`/`false`, numbers are numbers, not strings.

### Reading Slot Properties

```lua
local slot = RESH.inspect("slot-id")

-- Access properties directly
print(slot.Name)                    -- string
print(slot.Active)                  -- boolean (true/false)
print(slot.Tag)                     -- string or nil
print(slot.Position.x)              -- number

-- Check for nil
if slot.Tag then
    print("Tagged: " .. slot.Tag)
end
```

### Reading Component Members

Component members are indexed **by name** (not array index).

```lua
local comp = RESH.inspect("component-id")

print(comp.TypeName)                -- "DynamicVariableSpace"
print(comp.ComponentType)           -- "FrooxEngine.DynamicVariableSpace"

-- Access members by name
local spaceName = comp.Members["SpaceName"]
print(spaceName.Value)              -- e.g., "RESH"
print(spaceName.Type)               -- e.g., "string"

-- Iterate all members
for name, member in pairs(comp.Members) do
    print(name .. " = " .. tostring(member.Value))
end
```

### Type-Aware Values

Values are automatically converted to proper Lua types:

```lua
-- Booleans (NOT strings!)
if comp.Members["Enabled"].Value then           -- ✅ Correct
    print("Enabled")
end

-- Numbers (NOT strings!)
local count = comp.Members["Count"].Value
print("Double: " .. (count * 2))               -- ✅ Works with arithmetic

-- Strings
local name = comp.Members["SpaceName"].Value
print("Space: " .. name)

-- References (tables with TargetId)
local ref = comp.Members["Reference"]
if ref.TargetId then
    print("References: " .. ref.TargetId)
    local target = RESH.inspect(ref.TargetId)
end

-- Complex types (nested tables)
local pos = comp.Members["Position"].Value
print(string.format("Position: (%.2f, %.2f, %.2f)", pos.x, pos.y, pos.z))
```

### Common Mistakes

```lua
-- ❌ WRONG: Comparing boolean to string
if comp.Members["Enabled"].Value == "true" then  -- Won't work!

-- ✅ CORRECT: Direct boolean check
if comp.Members["Enabled"].Value then            -- Works!

-- ❌ WRONG: Accessing members by index
local first = comp.Members[1]                    -- Won't work!

-- ✅ CORRECT: Access by name
local enabled = comp.Members["Enabled"]          -- Works!
```

---

## Common Patterns

### Find and Navigate

```lua
-- Safe navigation
RESH.cd("/")
local resh_id = RESH.find_slot("RESH.DATA")
if not resh_id then
    error("RESH.DATA not found!")
end
RESH.cd("RESH.DATA")
```

### Find Components by Type

```lua
local listing = RESH.ls()

for i, comp in ipairs(listing.components) do
    if comp.type:match("DynamicVariable") then
        local data = RESH.inspect(comp.id)
        print("Found: " .. data.TypeName)
    end
end
```

### Traverse Children

```lua
local listing = RESH.ls()

for i, child in ipairs(listing.children) do
    print("Child: " .. child.name)
    
    RESH.cd(child.name)
    -- Do work here
    RESH.cd("..")
end
```

### Filter and Map Children

```lua
-- Get all active children
local listing = RESH.ls()
local active_names = {}

for i, child in ipairs(listing.children) do
    if child.active then
        table.insert(active_names, child.name)
    end
end

print("Active children: " .. table.concat(active_names, ", "))
```

### Reusable Helper Functions

```lua
-- Helper: Find component by type pattern
local function find_component(type_pattern)
    local listing = RESH.ls()
    for i, comp in ipairs(listing.components) do
        if comp.type:match(type_pattern) then
            return comp
        end
    end
    return nil
end

-- Helper: Get all components of a type
local function find_all_components(type_pattern)
    local results = {}
    local listing = RESH.ls()
    for i, comp in ipairs(listing.components) do
        if comp.type:match(type_pattern) then
            table.insert(results, comp)
        end
    end
    return results
end

-- Usage
local space = find_component("DynamicVariableSpace")
if space then
    local data = RESH.inspect(space.id)
    print("Space: " .. data.Members["SpaceName"].Value)
end

local all_vars = find_all_components("DynamicVariable")
print("Found " .. #all_vars .. " variables")
```

---

## Complex Examples

### Example 1: List All Bookmarks

```lua
RESH.cd("/RESH.DATA/Bookmarks")
local listing = RESH.ls()

for i, child in ipairs(listing.children) do
    print("\nBookmark: " .. child.name)
    
    RESH.cd(child.name)
    local child_listing = RESH.ls()
    
    -- Find the reference variable
    for j, comp in ipairs(child_listing.components) do
        if comp.type:match("DynamicReferenceVariable") then
            local var = RESH.inspect(comp.id)
            local var_name = var.Members["VariableName"].Value
            local ref_id = var.Members["Reference"].TargetId
            
            print("  Variable: " .. var_name)
            if ref_id then
                local target = RESH.inspect(ref_id)
                print("  Target: " .. target.Name)
            else
                print("  Target: <null>")
            end
        end
    end
    
    RESH.cd("..")
end
```

### Example 2: Recursive Traversal

```lua
local function traverse(depth, max_depth)
    depth = depth or 0
    max_depth = max_depth or 5
    
    if depth > max_depth then return end
    
    local indent = string.rep("  ", depth)
    local slot = RESH.inspect(RESH.get_current_slot())
    
    print(indent .. slot.Name .. " [" .. (slot.Active and "ON" or "OFF") .. "]")
    
    local listing = RESH.ls()
    for i, child in ipairs(listing.children) do
        RESH.cd(child.name)
        traverse(depth + 1, max_depth)
        RESH.cd("..")
    end
end

RESH.cd("/")
traverse()
```

### Example 3: Collect Statistics

```lua
local function collect_stats(max_depth)
    max_depth = max_depth or 5
    local stats = {
        total_slots = 0,
        active_slots = 0,
        component_count = {}
    }
    
    local function collect(depth)
        if depth > max_depth then return end
        
        stats.total_slots = stats.total_slots + 1
        
        local slot = RESH.inspect(RESH.get_current_slot())
        if slot.Active then
            stats.active_slots = stats.active_slots + 1
        end
        
        local listing = RESH.ls()
        for i, comp in ipairs(listing.components) do
            stats.component_count[comp.type] = 
                (stats.component_count[comp.type] or 0) + 1
        end
        
        for i, child in ipairs(listing.children) do
            RESH.cd(child.name)
            collect(depth + 1)
            RESH.cd("..")
        end
    end
    
    collect(0)
    return stats
end

RESH.cd("/RESH.DATA")
local stats = collect_stats(3)
print("Total Slots: " .. stats.total_slots)
print("Active Slots: " .. stats.active_slots)
```

### Example 4: Find Slots by Tag

```lua
local function find_by_tag(tag, max_depth)
    max_depth = max_depth or 10
    local results = {}
    
    local function search(depth)
        if depth > max_depth then return end
        
        local id = RESH.get_current_slot()
        local slot = RESH.inspect(id)
        
        if slot.Tag == tag then
            table.insert(results, {
                id = id,
                name = slot.Name,
                path = RESH.pwd()
            })
        end
        
        local listing = RESH.ls()
        for i, child in ipairs(listing.children) do
            RESH.cd(child.name)
            search(depth + 1)
            RESH.cd("..")
        end
    end
    
    search(0)
    return results
end

RESH.cd("/")
local tagged = find_by_tag("RESHData", 5)
for i, slot in ipairs(tagged) do
    print(slot.path .. " (" .. slot.name .. ")")
end
```

### Example 5: Examine DynamicVariableSpace

```lua
RESH.cd("/RESH.DATA")
local listing = RESH.ls()

for i, comp in ipairs(listing.components) do
    if comp.type:match("DynamicVariableSpace") then
        local space = RESH.inspect(comp.id)
        
        print("Found DynamicVariableSpace:")
        print("  SpaceName: " .. space.Members["SpaceName"].Value)
        print("  OnlyDirectBinding: " .. tostring(space.Members["OnlyDirectBinding"].Value))
        
        -- List all members
        for name, member in pairs(space.Members) do
            print("  " .. name .. " (" .. member.Type .. ") = " .. tostring(member.Value))
        end
    end
end
```

### Example 6: Export Hierarchy to JSON-like String

```lua
local function export_slot()
    local id = RESH.get_current_slot()
    local slot = RESH.inspect(id)
    local listing = RESH.ls()
    
    local data = {
        name = slot.Name,
        active = slot.Active,
        tag = slot.Tag,
        children_count = #listing.children,
        components = {}
    }
    
    for i, comp in ipairs(listing.components) do
        table.insert(data.components, comp.type)
    end
    
    return data
end

-- Usage
RESH.cd("/RESH.DATA")
local data = export_slot()
print("Slot: " .. data.name)
print("Children: " .. data.children_count)
print("Components: " .. #data.components)
```

### Example 7: Check Variable Existence

```lua
local function has_component_type(type_pattern)
    local listing = RESH.ls()
    for i, comp in ipairs(listing.components) do
        if comp.type:match(type_pattern) then
            return true
        end
    end
    return false
end

-- Check if slot has DynamicVariableSpace
RESH.cd("/RESH.DATA")
if has_component_type("DynamicVariableSpace") then
    print("Has dynamic variable space")
else
    print("No dynamic variable space")
end
```

### Example 8: Build Slot Path Map

```lua
local function build_path_map(max_depth)
    max_depth = max_depth or 3
    local map = {}
    
    local function build(depth)
        if depth > max_depth then return end
        
        local path = RESH.pwd()
        local id = RESH.get_current_slot()
        local slot = RESH.inspect(id)
        
        map[slot.Name] = {
            path = path,
            id = id,
            active = slot.Active
        }
        
        local listing = RESH.ls()
        for i, child in ipairs(listing.children) do
            RESH.cd(child.name)
            build(depth + 1)
            RESH.cd("..")
        end
    end
    
    build(0)
    return map
end

-- Build map of all slots under RESH.DATA
RESH.cd("/RESH.DATA")
local map = build_path_map(2)

for name, info in pairs(map) do
    print(name .. " -> " .. info.path)
end
```

### Example 9: Validate RESH.DATA Structure

```lua
local function validate_resh_data()
    -- Navigate to RESH.DATA
    RESH.cd("/")
    if not RESH.find_slot("RESH.DATA") then
        return false, "RESH.DATA not found"
    end
    
    RESH.cd("RESH.DATA")
    
    -- Check for required components
    local listing = RESH.ls()
    local has_space = false
    local has_self_ref = false
    
    for i, comp in ipairs(listing.components) do
        if comp.type:match("DynamicVariableSpace") then
            has_space = true
        end
        if comp.type:match("DynamicReferenceVariable") then
            has_self_ref = true
        end
    end
    
    if not has_space then
        return false, "Missing DynamicVariableSpace"
    end
    if not has_self_ref then
        return false, "Missing self-reference variable"
    end
    
    -- Check for required folders
    if not RESH.find_slot("Bookmarks") then
        return false, "Missing Bookmarks folder"
    end
    if not RESH.find_slot("Variables") then
        return false, "Missing Variables folder"
    end
    
    return true, "RESH.DATA structure valid"
end

-- Validate
local ok, msg = validate_resh_data()
if ok then
    print("✓ " .. msg)
else
    print("✗ " .. msg)
end
```

### Example 10: Compare Two Slots

```lua
local function compare_slots(id1, id2)
    local slot1 = RESH.inspect(id1)
    local slot2 = RESH.inspect(id2)
    
    print("Comparing:")
    print("  " .. slot1.Name .. " vs " .. slot2.Name)
    print("")
    
    print("Active:")
    print("  " .. tostring(slot1.Active) .. " vs " .. tostring(slot2.Active))
    
    print("Persistent:")
    print("  " .. tostring(slot1.Persistent) .. " vs " .. tostring(slot2.Persistent))
    
    if slot1.Tag or slot2.Tag then
        print("Tag:")
        print("  " .. (slot1.Tag or "<none>") .. " vs " .. (slot2.Tag or "<none>"))
    end
end

-- Usage: Compare two child slots
local listing = RESH.ls()
if #listing.children >= 2 then
    compare_slots(listing.children[1].id, listing.children[2].id)
end
```

---

## Best Practices

### 1. Always Check for Nil

```lua
-- ❌ BAD: No error checking
local id = RESH.find_slot("SomeSlot")
RESH.cd("SomeSlot")  -- Might fail silently

-- ✅ GOOD: Check before using
local id = RESH.find_slot("SomeSlot")
if not id then
    error("Slot not found: SomeSlot")
end
RESH.cd("SomeSlot")

-- ✅ BETTER: Use a helper function
local function navigate_safe(name)
    if not RESH.find_slot(name) then
        error("Cannot navigate to " .. name)
    end
    RESH.cd(name)
end
```

### 2. Use Type-Aware Comparisons

```lua
-- ❌ WRONG: Comparing boolean to string
if comp.Members["Enabled"].Value == "true" then
    -- This will never work!
end

-- ✅ CORRECT: Direct boolean check
if comp.Members["Enabled"].Value then
    -- This works correctly
end

-- For numbers
local count = comp.Members["Count"].Value
if count > 10 then  -- ✅ Works with numbers
    print("Count is high")
end
```

### 3. Access Members by Name

```lua
-- ❌ WRONG: Trying to access by index
local members = comp.Members
local first = members[1]  -- Returns nil!

-- ✅ CORRECT: Access by name
local enabled = comp.Members["Enabled"]
local value = comp.Members["SpaceName"]

-- ✅ CORRECT: Iterate with pairs()
for name, member in pairs(comp.Members) do
    print(name)  -- Prints member names
end
```

### 4. Use Absolute Paths

```lua
-- ❌ RISKY: Multiple relative navigations
RESH.cd("/")
RESH.cd("RESH.DATA")
RESH.cd("Bookmarks")
RESH.cd("MyBookmark")

-- ✅ BETTER: Single absolute path
RESH.cd("/RESH.DATA/Bookmarks/MyBookmark")

-- ✅ BEST: With error checking
local success, err = RESH.cd("/RESH.DATA/Bookmarks/MyBookmark")
if not success then
    print("Navigation failed: " .. err)
end
```

### 5. Cache Inspection Results

```lua
-- ❌ SLOW: Re-inspecting in loop
for i = 1, 100 do
    local slot = RESH.inspect(some_id)  -- Network call every time!
    print(slot.Name)
end

-- ✅ FAST: Cache the result
local slot = RESH.inspect(some_id)  -- Single network call
for i = 1, 100 do
    print(slot.Name)
end
```

### 6. Build Reusable Functions

```lua
-- Create utility library
local Utils = {}

function Utils.find_component(type_pattern)
    local listing = RESH.ls()
    for i, comp in ipairs(listing.components) do
        if comp.type:match(type_pattern) then
            return comp
        end
    end
    return nil
end

function Utils.navigate_to_resh_data()
    RESH.cd("/")
    if not RESH.find_slot("RESH.DATA") then
        error("RESH.DATA not found")
    end
    RESH.cd("RESH.DATA")
end

-- Use in scripts
Utils.navigate_to_resh_data()
local space = Utils.find_component("DynamicVariableSpace")
```

### 7. Use pcall for Production

```lua
-- Wrap risky operations
local function safe_inspect(id)
    local success, result = pcall(function()
        return RESH.inspect(id)
    end)
    
    if success then
        return result
    else
        print("Inspection failed: " .. result)
        return nil
    end
end

-- Use it
local slot = safe_inspect("some-id")
if slot then
    print(slot.Name)
end
```

### 8. Document Your Functions

```lua
--- Navigate to a slot and return its inspection data
-- @param path string Absolute or relative path
-- @return table|nil Slot data, or nil on failure
local function get_slot_at(path)
    local success = RESH.cd(path)
    if not success then
        return nil
    end
    
    return RESH.inspect(RESH.get_current_slot())
end
```

---

## Quick Reference

```lua
-- Navigation
RESH.cd(path)               -- Navigate to path
RESH.pwd()                  -- Get current path
RESH.get_current_slot()     -- Get current slot ID
RESH.find_slot(name)        -- Find child by name

-- Inspection
RESH.ls()                   -- List children & components
RESH.inspect(id)            -- Get detailed data

-- Data Access
listing.children[i].name    -- Child name
listing.components[i].type  -- Component type
comp.Members["Name"].Value  -- Member value
slot.Name                   -- Slot property

-- Standard Lua
print(...)                  -- Output
string.format(fmt, ...)     -- Format string
table.insert(t, val)        -- Insert to table
```

---

## Example Scripts

- **[examples/basic/test.lua](examples/basic/test.lua)** - Basic usage
- **[examples/basic/demo_lua_interface.lua](examples/basic/demo_lua_interface.lua)** - Data structures
- **[examples/resh_data/inspect_resh_data.lua](examples/resh_data/inspect_resh_data.lua)** - RESH.DATA inspection
- **[examples/resh_data/explore_structure.lua](examples/resh_data/explore_structure.lua)** - Comprehensive explorer
- **[examples/resh_data/examine_bookmarks.lua](examples/resh_data/examine_bookmarks.lua)** - Bookmark system

---

## Limitations & Future

**Current Limitations:**
- Read-only (no create/modify/delete yet)
- Synchronous execution
- One script at a time

**Coming Soon:**
- `RESH.get_var(name)`, `RESH.set_var(name, value)`
- `RESH.create_var(name, type, value)`, `RESH.delete_var(name)`
- `RESH.add_slot(name)`, `RESH.delete_slot(id)`
- `RESH.add_component(slot_id, type)`, `RESH.update_component(id, members)`

See [docs/development/VARIABLE_SYSTEM.md](docs/development/VARIABLE_SYSTEM.md) for details.
