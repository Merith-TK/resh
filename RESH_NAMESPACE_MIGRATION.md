# RESH Namespace Migration

## Overview

All RESH-specific Lua functions have been moved from the global namespace to the `RESH` table to keep the global scope clean and enable library development.

## Changes Made

### 1. Core API (pkg/shell/script.go)

Refactored `registerShellFunctions()` to organize all RESH functions under the `RESH` namespace:

**Before:**
```lua
cd("/")
ls()
inspect(id)
pwd()
get_current_slot()
find_slot(name)
```

**After:**
```lua
RESH.cd("/")
RESH.ls()
RESH.inspect(id)
RESH.pwd()
RESH.get_current_slot()
RESH.find_slot(name)
```

**Implementation Details:**
- Created `RESH` table using `L.NewTable()`
- Changed all `L.SetGlobal(name, func)` to `L.SetField(resh, name, func)`
- Set RESH table as global with `L.SetGlobal("RESH", resh)`
- Kept standard Lua functions (like `print`) in global scope

### 2. Example Scripts

Updated all example scripts to use RESH namespace:

- ✅ `examples/basic/test.lua`
- ✅ `examples/basic/demo_lua_interface.lua`
- ✅ `examples/resh_data/inspect_resh_data.lua`
- ✅ `examples/resh_data/explore_structure.lua`
- ✅ `examples/resh_data/examine_bookmarks.lua`

### 3. Documentation

Created comprehensive `LUA_SCRIPTING.md` with:
- Quick start guide
- Namespace rationale
- Complete API reference for all `RESH.*` functions
- Data structure documentation
- Common patterns and examples
- Best practices

## Benefits

1. **Clean Global Scope**: Only standard Lua functions remain global
2. **No Name Conflicts**: User scripts can define `cd`, `ls`, etc. without issues
3. **Library Development**: Clear API boundary enables reusable Lua libraries
4. **Professional API**: Follows Lua best practices for module organization
5. **Future Expansion**: Easy to add new functions to RESH namespace

## Backward Compatibility

⚠️ **Breaking Change**: Existing scripts using `cd()`, `ls()`, etc. will need to update to `RESH.cd()`, `RESH.ls()`, etc.

This is a necessary breaking change for a cleaner, more maintainable API going forward.

## Standard Lua Functions

These remain in global scope (unchanged):
- `print(...)`
- `string.*`
- `table.*`
- `math.*`
- `io.*`
- All other standard Lua libraries

## Future Additions

The RESH namespace will be extended with:

**Variable Management:**
```lua
RESH.get_var(name)
RESH.set_var(name, value)
RESH.create_var(name, type, value)
RESH.delete_var(name)
RESH.list_vars()
```

**World Manipulation:**
```lua
RESH.add_slot(name, parent)
RESH.delete_slot(id)
RESH.add_component(slot_id, type)
RESH.update_component(id, members)
RESH.delete_component(id)
```

**Bookmark Management:**
```lua
RESH.list_bookmarks()
RESH.get_bookmark(name)
RESH.create_bookmark(name, slot_id)
RESH.delete_bookmark(name)
```

See [docs/development/VARIABLE_SYSTEM.md](docs/development/VARIABLE_SYSTEM.md) for detailed variable system design.

## Migration Guide

For existing scripts, update function calls:

```lua
-- Before
cd("/")
local listing = ls()
local data = inspect(id)
local path = pwd()
local current = get_current_slot()
local found = find_slot("name")

-- After
RESH.cd("/")
local listing = RESH.ls()
local data = RESH.inspect(id)
local path = RESH.pwd()
local current = RESH.get_current_slot()
local found = RESH.find_slot("name")
```

Simple find-and-replace patterns:
- `cd(` → `RESH.cd(`
- `ls()` → `RESH.ls()`
- `inspect(` → `RESH.inspect(`
- `pwd()` → `RESH.pwd()`
- `get_current_slot()` → `RESH.get_current_slot()`
- `find_slot(` → `RESH.find_slot(`

## Testing

Build verification:
```bash
go build -o resh.exe .
# ✅ Success
```

All example scripts have been updated and tested with the new namespace.

## References

- [LUA_SCRIPTING.md](LUA_SCRIPTING.md) - Complete API documentation
- [examples/](examples/) - Updated example scripts
- [pkg/shell/script.go](pkg/shell/script.go) - Implementation
