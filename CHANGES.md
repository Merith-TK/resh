# Changes Summary

## 1. Standalone Script Execution

Added `--run-script` functionality as a new command:

### Usage
```bash
# Run script without entering REPL
resh script my_script.lua --url ws://localhost:39015

# Or from within REPL
/ $ script my_script.lua
```

### Implementation
- New file: [cmd/script.go](cmd/script.go)
- Registered as `script` subcommand in cobra
- Connects to Resonite, initializes state, runs script, then exits
- Perfect for automation, CI/CD, and tooling integration

## 2. Improved Lua Data Interface

Completely revamped how data is presented to Lua scripts for better ergonomics and type safety.

### Key Improvements

#### A. Component Members Now Indexed by Name
**Before:** Array of members, had to iterate to find specific member
```lua
for i, member in ipairs(comp_data.members) do
    if member.name == "Enabled" then
        print(member.value)  -- string "true"
    end
end
```

**After:** Direct access by member name with proper types
```lua
print(comp_data.Members.Enabled.Value)  -- boolean true
```

#### B. Slot Properties as Direct Fields
**Before:** Array of properties
```lua
for i, prop in ipairs(slot_data.properties) do
    if prop.name == "Name" then
        print(prop.value)
    end
end
```

**After:** Direct field access
```lua
print(slot_data.Name)
print(slot_data.Active)  -- boolean, not string
print(slot_data.Position.x)  -- structured value
```

#### C. Type-Aware Value Conversion
**Before:** All values converted to strings via `fmt.Sprintf("%v", value)`
```lua
if comp_data.Members.Enabled.value == "true" then  -- string comparison
```

**After:** Proper type conversion
```lua
if comp_data.Members.Enabled.Value then  -- boolean
    local size = comp_data.Members.Size.Value
    print(size.x + size.y + size.z)  -- math operations work
end
```

Added `convertToLuaValue()` function that handles:
- `bool` → Lua boolean
- `int`, `int64`, `float32`, `float64` → Lua number
- `string` → Lua string
- `map[string]interface{}` → Lua table (key-value)
- `[]interface{}` → Lua table (array)
- `nil` → Lua nil

#### D. Reference Convenience Field
References now have a dedicated `TargetId` field:
```lua
if member.Type == "reference" then
    local target = member.TargetId  -- "Reso_123" or nil
    if target then
        print("Points to: " .. target)
    else
        print("Null reference")
    end
end
```

#### E. Consistent Naming Convention
All fields use PascalCase (matching Resonite):
- `Members` not `members`
- `Type` not `type`
- `TargetId` not `targetId`

### Files Changed
- [pkg/shell/script.go](pkg/shell/script.go)
  - Rewrote `inspect()` function to return member map instead of array
  - Rewrote slot inspection to return direct properties
  - Added `convertToLuaValue()` helper function
  - Updated all field names to PascalCase

### Updated Example Scripts
- [test.lua](test.lua) - Basic functionality test
- [inspect_resh_data.lua](inspect_resh_data.lua) - RESH.DATA structure inspector
- [demo_lua_interface.lua](demo_lua_interface.lua) - NEW: Demonstrates improved interface
- [LUA_SCRIPTING.md](LUA_SCRIPTING.md) - NEW: Complete documentation

## Benefits

1. **More Idiomatic Lua:** Direct property access instead of array iteration
2. **Type Safety:** Boolean is boolean, number is number
3. **Better Performance:** No need to iterate through all members to find one
4. **Easier Scripting:** More intuitive and less boilerplate code
5. **CI/CD Ready:** Standalone script mode enables automation workflows

## Testing

Build and test:
```bash
# Build
go build -o resh.exe .

# Test standalone mode
resh script test.lua --url ws://localhost:39015

# Test from REPL
resh repl --url ws://localhost:39015
/ $ script demo_lua_interface.lua
```

## Next Steps

With the improved Lua interface, you can now:
1. Create sample RESH.DATA structure in Resonite
2. Use `resh script inspect_resh_data.lua` to examine it
3. Implement bookmark persistence based on observed structure
4. Build more complex automation scripts
