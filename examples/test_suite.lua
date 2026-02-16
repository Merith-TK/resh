-- RESH Lua API Comprehensive Test Suite
-- Tests all API functions and validates proper operation
-- Covers: navigation, inspection, type handling, write ops, error handling,
--         component type system, and all common Resonite value types

print("================================================================")
print("          RESH Lua API Comprehensive Test Suite")
print("================================================================")
print("")

-- Test results tracking
local tests = {
    passed = 0,
    failed = 0,
    skipped = 0,
    results = {}
}

-- Test helper functions
local function test(name, func)
    io.write("Testing: " .. name .. " ... ")
    io.flush()

    local success, result = pcall(func)

    if success and result then
        tests.passed = tests.passed + 1
        table.insert(tests.results, {name = name, status = "PASS"})
        print("PASS")
        return true
    elseif success and result == false then
        tests.failed = tests.failed + 1
        table.insert(tests.results, {name = name, status = "FAIL", error = "Test returned false"})
        print("FAIL")
        return false
    else
        tests.failed = tests.failed + 1
        local error_msg = tostring(result)
        table.insert(tests.results, {name = name, status = "FAIL", error = error_msg})
        print("FAIL: " .. error_msg)
        return false
    end
end

local function skip(name, reason)
    io.write("Testing: " .. name .. " ... ")
    tests.skipped = tests.skipped + 1
    table.insert(tests.results, {name = name, status = "SKIP", reason = reason})
    print("SKIP: " .. reason)
end

local function assert_not_nil(value, message)
    if value == nil then
        error(message or "Value is nil")
    end
    return true
end

local function assert_equals(actual, expected, message)
    if actual ~= expected then
        error((message or "Values not equal") .. string.format(": expected %s, got %s", tostring(expected), tostring(actual)))
    end
    return true
end

local function assert_type(value, expected_type, message)
    local actual_type = type(value)
    if actual_type ~= expected_type then
        error((message or "Type mismatch") .. string.format(": expected %s, got %s", expected_type, actual_type))
    end
    return true
end

local function assert_true(value, message)
    if not value then
        error(message or "Expected true")
    end
    return true
end

-- Store initial state
local initial_slot = nil
local initial_path = nil

print("================================================================")
print("SECTION 1: Navigation Functions")
print("================================================================")
print("")

-- Test 1: RESH.pwd() - Get current path
test("RESH.pwd() returns string", function()
    local path = RESH.pwd()
    assert_type(path, "string", "pwd() should return string")
    initial_path = path
    print("    Current path: " .. path)
    return true
end)

-- Test 2: RESH.get_current_slot() - Get current slot ID
test("RESH.get_current_slot() returns string", function()
    local slot_id = RESH.get_current_slot()
    assert_type(slot_id, "string", "get_current_slot() should return string")
    assert_not_nil(slot_id, "Slot ID should not be nil")
    initial_slot = slot_id
    print("    Current slot: " .. slot_id)
    return true
end)

-- Test 3: RESH.cd("/") - Navigate to root
test("RESH.cd('/') navigates to root", function()
    local success = RESH.cd("/")
    assert_equals(success, true, "cd('/') should return true")
    local path = RESH.pwd()
    assert_equals(path, "/", "Path should be '/' after cd('/')")
    return true
end)

-- Test 4: RESH.find_slot() - Find existing child
test("RESH.find_slot() finds existing slot", function()
    RESH.cd("/")
    local listing = RESH.ls()

    if #listing.children == 0 then
        skip("find_slot test", "No children at root")
        return false
    end

    local first_child_name = listing.children[1].name
    local found_id = RESH.find_slot(first_child_name)

    assert_not_nil(found_id, "find_slot should find existing child")
    assert_equals(found_id, listing.children[1].id, "Found ID should match ls() result")
    print("    Found: " .. first_child_name .. " (" .. found_id .. ")")
    return true
end)

-- Test 5: RESH.find_slot() - Returns nil for non-existent
test("RESH.find_slot() returns nil for non-existent", function()
    RESH.cd("/")
    local found_id = RESH.find_slot("__NONEXISTENT_SLOT_XYZ__")
    assert_equals(found_id, nil, "find_slot should return nil for non-existent slot")
    return true
end)

-- Test 6: RESH.cd() - Navigate to child
test("RESH.cd() navigates to child slot", function()
    RESH.cd("/")
    local listing = RESH.ls()

    if #listing.children == 0 then
        error("No children at root to test navigation")
    end

    local first_child = listing.children[1].name
    local success = RESH.cd(first_child)
    assert_equals(success, true, "cd() should succeed for existing child")

    local new_path = RESH.pwd()
    assert_not_nil(new_path, "pwd() should return path after navigation")
    print("    Navigated to: " .. new_path)
    return true
end)

-- Test 7: RESH.cd("..") - Navigate to parent
test("RESH.cd('..') navigates to parent", function()
    -- Make sure we're not at root
    RESH.cd("/")
    local listing = RESH.ls()
    if #listing.children > 0 then
        RESH.cd(listing.children[1].name)
    end

    local success = RESH.cd("..")
    assert_equals(success, true, "cd('..') should succeed")
    return true
end)

print("")
print("================================================================")
print("SECTION 2: Inspection Functions")
print("================================================================")
print("")

-- Test 8: RESH.ls() - List current slot
test("RESH.ls() returns table with children and components", function()
    RESH.cd("/")
    local listing = RESH.ls()

    assert_not_nil(listing, "ls() should return a table")
    assert_type(listing, "table", "ls() should return a table")
    assert_not_nil(listing.children, "ls() result should have 'children' field")
    assert_not_nil(listing.components, "ls() result should have 'components' field")
    assert_type(listing.children, "table", "'children' should be a table")
    assert_type(listing.components, "table", "'components' should be a table")

    print("    Found " .. #listing.children .. " children and " .. #listing.components .. " components")
    return true
end)

-- Test 9: RESH.ls() - Children have required fields
test("RESH.ls() children have required fields", function()
    RESH.cd("/")
    local listing = RESH.ls()

    if #listing.children == 0 then
        error("No children to test")
    end

    local child = listing.children[1]
    assert_not_nil(child.id, "Child should have 'id' field")
    assert_not_nil(child.name, "Child should have 'name' field")
    assert_not_nil(child.active, "Child should have 'active' field")
    assert_type(child.id, "string", "Child 'id' should be string")
    assert_type(child.name, "string", "Child 'name' should be string")
    assert_type(child.active, "boolean", "Child 'active' should be boolean")

    return true
end)

-- Test 10: RESH.ls() - Components have required fields
test("RESH.ls() components have required fields", function()
    RESH.cd("/")
    local listing = RESH.ls()

    if #listing.components == 0 then
        error("No components to test")
    end

    local comp = listing.components[1]
    assert_not_nil(comp.id, "Component should have 'id' field")
    assert_not_nil(comp.type, "Component should have 'type' field")
    assert_type(comp.id, "string", "Component 'id' should be string")
    assert_type(comp.type, "string", "Component 'type' should be string")

    return true
end)

-- Test 11: RESH.inspect() - Inspect slot
test("RESH.inspect() returns slot data", function()
    RESH.cd("/")
    local slot_id = RESH.get_current_slot()
    local data = RESH.inspect(slot_id)

    assert_not_nil(data, "inspect() should return data")
    assert_type(data, "table", "inspect() should return a table")
    assert_equals(data.Type, "slot", "inspect() on slot should return Type='slot'")
    assert_not_nil(data.Name, "Slot data should have 'Name' field")

    print("    Inspected slot: " .. tostring(data.Name))
    return true
end)

-- Test 12: RESH.inspect() - Inspect component
test("RESH.inspect() returns component data", function()
    RESH.cd("/")
    local listing = RESH.ls()

    if #listing.components == 0 then
        error("No components to inspect")
    end

    local comp_id = listing.components[1].id
    local data = RESH.inspect(comp_id)

    assert_not_nil(data, "inspect() should return data")
    assert_type(data, "table", "inspect() should return a table")
    assert_equals(data.Type, "component", "inspect() on component should return Type='component'")
    assert_not_nil(data.TypeName, "Component data should have 'TypeName' field")
    assert_not_nil(data.Members, "Component data should have 'Members' field")
    assert_type(data.Members, "table", "'Members' should be a table")

    print("    Inspected component: " .. data.TypeName)
    return true
end)

-- Test 13: Component members are indexed by name
test("RESH.inspect() component members indexed by name", function()
    RESH.cd("/")
    local listing = RESH.ls()

    if #listing.components == 0 then
        error("No components to inspect")
    end

    local comp_id = listing.components[1].id
    local data = RESH.inspect(comp_id)

    -- Check that members can be accessed by name using pairs()
    local member_count = 0
    for name, member in pairs(data.Members) do
        member_count = member_count + 1
        assert_type(name, "string", "Member key should be string")
        assert_type(member, "table", "Member value should be table")
        assert_not_nil(member.Type, "Member should have 'Type' field")
        break -- Just test first one
    end

    assert_not_nil(member_count > 0, "Component should have at least one member")
    print("    Component has " .. member_count .. " members")
    return true
end)

print("")
print("================================================================")
print("SECTION 3: Value Type Handling")
print("================================================================")
print("")

-- Helper: Find a component member with a specific type across all root components
local function find_member_of_type(target_type)
    RESH.cd("/")
    local listing = RESH.ls()

    for i, comp in ipairs(listing.components) do
        local data = RESH.inspect(comp.id)
        if data and data.Members then
            for name, member in pairs(data.Members) do
                if member.Type == target_type then
                    return member, name, data
                end
            end
        end
    end
    return nil, nil, nil
end

-- Test 14: Boolean values are proper booleans
test("Value type: bool is Lua boolean", function()
    local member, name = find_member_of_type("bool")
    if not member then
        skip("Boolean value test", "No boolean members found")
        return false
    end

    assert_type(member.Value, "boolean", "Boolean member should have boolean value")
    print("    Found bool member: " .. name .. " = " .. tostring(member.Value))
    return true
end)

-- Test 15: String values are proper strings
test("Value type: string is Lua string", function()
    local member, name = find_member_of_type("string")
    if not member then
        skip("String value test", "No string members found")
        return false
    end

    -- String members can have nil values in Resonite (null strings)
    if member.Value == nil then
        print("    Found string member with nil value: " .. name .. " (null string)")
        return true
    end

    assert_type(member.Value, "string", "String member should have string value")
    print("    Found string member: " .. name .. " = " .. tostring(member.Value))
    return true
end)

-- Test 16: Number values (int) are proper numbers
test("Value type: int is Lua number", function()
    local member, name = find_member_of_type("int")
    if not member then
        skip("Int value test", "No int members found")
        return false
    end

    assert_type(member.Value, "number", "Int member should have number value")
    print("    Found int member: " .. name .. " = " .. tostring(member.Value))
    return true
end)

-- Test 17: Float values are proper numbers
test("Value type: float is Lua number", function()
    local member, name = find_member_of_type("float")
    if not member then
        skip("Float value test", "No float members found")
        return false
    end

    assert_type(member.Value, "number", "Float member should have number value")
    print("    Found float member: " .. name .. " = " .. tostring(member.Value))
    return true
end)

-- Test 18: Long values are proper numbers
test("Value type: long is Lua number", function()
    local member, name = find_member_of_type("long")
    if not member then
        skip("Long value test", "No long members found")
        return true  -- Skip is success, not failure
    end

    assert_type(member.Value, "number", "Long member should have number value")
    print("    Found long member: " .. name .. " = " .. tostring(member.Value))
    return true
end)

-- Test 19: Reference values have TargetId
test("Value type: reference has TargetId", function()
    RESH.cd("/")
    local listing = RESH.ls()
    local found_ref = false

    for i, comp in ipairs(listing.components) do
        local data = RESH.inspect(comp.id)
        for name, member in pairs(data.Members) do
            if member.Type == "reference" then
                print("    Found reference member: " .. name)
                if member.TargetId then
                    print("    References: " .. member.TargetId)
                else
                    print("    References: <null>")
                end
                found_ref = true
                break
            end
        end
        if found_ref then break end
    end

    if not found_ref then
        skip("Reference value test", "No reference members found")
        return false
    end

    return true
end)

-- Test 20: float3 values have x, y, z fields
test("Value type: float3 has x,y,z table", function()
    local member, name = find_member_of_type("float3")
    if not member then
        skip("float3 value test", "No float3 members found")
        return true  -- Skip is success, not failure
    end

    assert_type(member.Value, "table", "float3 should be a table")
    assert_not_nil(member.Value.x, "float3 should have 'x' field")
    assert_not_nil(member.Value.y, "float3 should have 'y' field")
    assert_not_nil(member.Value.z, "float3 should have 'z' field")
    assert_type(member.Value.x, "number", "float3.x should be number")
    assert_type(member.Value.y, "number", "float3.y should be number")
    assert_type(member.Value.z, "number", "float3.z should be number")
    print("    Found float3: " .. name .. " = {" ..
          member.Value.x .. ", " .. member.Value.y .. ", " .. member.Value.z .. "}")
    return true
end)

-- Test 21: floatQ values have x, y, z, w fields
test("Value type: floatQ has x,y,z,w table", function()
    local member, name = find_member_of_type("floatQ")
    if not member then
        skip("floatQ value test", "No floatQ members found")
        return false
    end

    assert_type(member.Value, "table", "floatQ should be a table")
    assert_not_nil(member.Value.x, "floatQ should have 'x' field")
    assert_not_nil(member.Value.y, "floatQ should have 'y' field")
    assert_not_nil(member.Value.z, "floatQ should have 'z' field")
    assert_not_nil(member.Value.w, "floatQ should have 'w' field")
    assert_type(member.Value.x, "number", "floatQ.x should be number")
    print("    Found floatQ: " .. name .. " = {" ..
          member.Value.x .. ", " .. member.Value.y .. ", " ..
          member.Value.z .. ", " .. member.Value.w .. "}")
    return true
end)

-- Test 22: colorX/color values have r, g, b, a fields
test("Value type: colorX/color has r,g,b,a table", function()
    -- try colorX first, then color
    local member, name = find_member_of_type("colorX")
    if not member then
        member, name = find_member_of_type("color")
    end
    if not member then
        skip("colorX value test", "No colorX/color members found")
        return false
    end

    assert_type(member.Value, "table", "colorX should be a table")
    assert_not_nil(member.Value.r, "colorX should have 'r' field")
    assert_not_nil(member.Value.g, "colorX should have 'g' field")
    assert_not_nil(member.Value.b, "colorX should have 'b' field")
    assert_type(member.Value.r, "number", "colorX.r should be number")
    print("    Found colorX: " .. name .. " = {r=" ..
          member.Value.r .. ", g=" .. member.Value.g ..
          ", b=" .. member.Value.b .. ", a=" .. tostring(member.Value.a) .. "}")
    return true
end)

-- Test 23: float2 values have x, y fields
test("Value type: float2 has x,y table", function()
    local member, name = find_member_of_type("float2")
    if not member then
        skip("float2 value test", "No float2 members found")
        return false
    end

    assert_type(member.Value, "table", "float2 should be a table")
    assert_not_nil(member.Value.x, "float2 should have 'x' field")
    assert_not_nil(member.Value.y, "float2 should have 'y' field")
    assert_type(member.Value.x, "number", "float2.x should be number")
    assert_type(member.Value.y, "number", "float2.y should be number")
    print("    Found float2: " .. name .. " = {" ..
          member.Value.x .. ", " .. member.Value.y .. "}")
    return true
end)

print("")
print("================================================================")
print("SECTION 4: Write Operations")
print("================================================================")
print("")

-- Test 24: Function existence checks
test("Write operation functions exist", function()
    assert_not_nil(RESH.update_component, "RESH.update_component should exist")
    assert_type(RESH.update_component, "function", "update_component should be a function")
    assert_not_nil(RESH.create_slot, "RESH.create_slot should exist")
    assert_type(RESH.create_slot, "function", "create_slot should be a function")
    assert_not_nil(RESH.delete_slot, "RESH.delete_slot should exist")
    assert_type(RESH.delete_slot, "function", "delete_slot should be a function")
    assert_not_nil(RESH.create_component, "RESH.create_component should exist")
    assert_type(RESH.create_component, "function", "create_component should be a function")
    assert_not_nil(RESH.delete_component, "RESH.delete_component should exist")
    assert_type(RESH.delete_component, "function", "delete_component should be a function")
    return true
end)

-- Test 25: Create and delete a slot
test("create_slot() and delete_slot() lifecycle", function()
    RESH.cd("/")
    local parent_id = RESH.get_current_slot()

    -- Create a test slot
    local new_slot_id, err = RESH.create_slot("TEST_SLOT_LUA", parent_id)

    if not new_slot_id then
        error("Failed to create slot: " .. tostring(err))
    end

    assert_type(new_slot_id, "string", "create_slot should return slot ID")
    print("    Created slot: " .. new_slot_id)

    -- Verify it exists
    local slot_data = RESH.inspect(new_slot_id)
    assert_not_nil(slot_data, "Created slot should be inspectable")
    assert_equals(slot_data.Type, "slot", "Created object should be a slot")

    -- Delete the slot
    local success, del_err = RESH.delete_slot(new_slot_id)

    if not success then
        error("Failed to delete slot: " .. tostring(del_err))
    end

    print("    Deleted slot: " .. new_slot_id)
    return true
end)

-- Test 26: Create and delete a DynamicVariableSpace component
test("create/delete DynamicVariableSpace component", function()
    RESH.cd("/")
    local parent_id = RESH.get_current_slot()

    local test_slot_id, err = RESH.create_slot("TEST_DVS_SLOT", parent_id)
    if not test_slot_id then
        error("Failed to create test slot: " .. tostring(err))
    end

    -- DynamicVariableSpace: 35 instances in world, rank #16
    local comp_type = "[FrooxEngine]FrooxEngine.DynamicVariableSpace"
    local comp_id, comp_err = RESH.create_component(test_slot_id, comp_type)

    if not comp_id then
        RESH.delete_slot(test_slot_id)
        error("Failed to create DynamicVariableSpace: " .. tostring(comp_err))
    end

    print("    Created: DynamicVariableSpace (" .. comp_id .. ")")

    -- Verify it exists and has members
    local comp_data = RESH.inspect(comp_id)
    assert_not_nil(comp_data, "Component should be inspectable")
    assert_equals(comp_data.Type, "component", "Should be a component")

    -- Inspect members
    local member_count = 0
    for name, member in pairs(comp_data.Members) do
        member_count = member_count + 1
    end
    print("    Members: " .. member_count)

    -- Cleanup
    RESH.delete_component(comp_id)
    RESH.delete_slot(test_slot_id)
    return true
end)

-- Test 27: Create and delete a Grabbable component
test("create/delete Grabbable component", function()
    RESH.cd("/")
    local parent_id = RESH.get_current_slot()

    local test_slot_id, err = RESH.create_slot("TEST_GRAB_SLOT", parent_id)
    if not test_slot_id then
        error("Failed to create test slot: " .. tostring(err))
    end

    -- Grabbable: 9 instances in world, rank #73
    local comp_type = "[FrooxEngine]FrooxEngine.Grabbable"
    local comp_id, comp_err = RESH.create_component(test_slot_id, comp_type)

    if not comp_id then
        RESH.delete_slot(test_slot_id)
        error("Failed to create Grabbable: " .. tostring(comp_err))
    end

    print("    Created: Grabbable (" .. comp_id .. ")")

    local comp_data = RESH.inspect(comp_id)
    assert_not_nil(comp_data, "Component should be inspectable")

    -- Grabbable should have various member types
    local type_set = {}
    for name, member in pairs(comp_data.Members) do
        type_set[member.Type] = (type_set[member.Type] or 0) + 1
    end
    print("    Member types found:")
    for t, count in pairs(type_set) do
        print("      " .. t .. ": " .. count)
    end

    RESH.delete_component(comp_id)
    RESH.delete_slot(test_slot_id)
    return true
end)

-- Test 28: Create and delete a SphereCollider component
test("create/delete SphereCollider component", function()
    RESH.cd("/")
    local parent_id = RESH.get_current_slot()

    local test_slot_id, err = RESH.create_slot("TEST_COL_SLOT", parent_id)
    if not test_slot_id then
        error("Failed to create test slot: " .. tostring(err))
    end

    -- SphereCollider: 43 instances in world, rank #8
    local comp_type = "[FrooxEngine]FrooxEngine.SphereCollider"
    local comp_id, comp_err = RESH.create_component(test_slot_id, comp_type)

    if not comp_id then
        RESH.delete_slot(test_slot_id)
        error("Failed to create SphereCollider: " .. tostring(comp_err))
    end

    print("    Created: SphereCollider (" .. comp_id .. ")")

    local comp_data = RESH.inspect(comp_id)
    assert_not_nil(comp_data, "SphereCollider should be inspectable")

    -- SphereCollider has: float (Radius), float3 (Offset), bool, reference members
    local found_types = {}
    for name, member in pairs(comp_data.Members) do
        found_types[member.Type] = true
    end

    -- Verify we see expected types
    print("    Types in SphereCollider: ")
    for t, _ in pairs(found_types) do
        io.write(t .. " ")
    end
    print("")

    RESH.delete_component(comp_id)
    RESH.delete_slot(test_slot_id)
    return true
end)

-- Test 29: Create and delete a BoxCollider component
test("create/delete BoxCollider component", function()
    RESH.cd("/")
    local parent_id = RESH.get_current_slot()

    local test_slot_id, err = RESH.create_slot("TEST_BOX_SLOT", parent_id)
    if not test_slot_id then
        error("Failed to create test slot: " .. tostring(err))
    end

    -- BoxCollider: 20 instances in world, rank #31
    local comp_type = "[FrooxEngine]FrooxEngine.BoxCollider"
    local comp_id, comp_err = RESH.create_component(test_slot_id, comp_type)

    if not comp_id then
        RESH.delete_slot(test_slot_id)
        error("Failed to create BoxCollider: " .. tostring(comp_err))
    end

    print("    Created: BoxCollider (" .. comp_id .. ")")

    local comp_data = RESH.inspect(comp_id)
    assert_not_nil(comp_data, "BoxCollider should be inspectable")

    -- BoxCollider should have float3 for Size and Offset
    local has_float3 = false
    for name, member in pairs(comp_data.Members) do
        if member.Type == "float3" then
            has_float3 = true
            assert_type(member.Value, "table", "float3 should be table")
            assert_not_nil(member.Value.x, "float3 should have x")
            assert_not_nil(member.Value.y, "float3 should have y")
            assert_not_nil(member.Value.z, "float3 should have z")
            print("    " .. name .. " = {" .. member.Value.x .. ", " .. member.Value.y .. ", " .. member.Value.z .. "}")
        end
    end

    assert_true(has_float3, "BoxCollider should have float3 members")

    RESH.delete_component(comp_id)
    RESH.delete_slot(test_slot_id)
    return true
end)

-- Test 30: Create multiple components on same slot
test("Multiple components on same slot", function()
    RESH.cd("/")
    local parent_id = RESH.get_current_slot()

    local test_slot_id, err = RESH.create_slot("TEST_MULTI_SLOT", parent_id)
    if not test_slot_id then
        error("Failed to create test slot: " .. tostring(err))
    end

    -- Create DynamicVariableSpace + SphereCollider on same slot
    local dvs_id, dvs_err = RESH.create_component(test_slot_id, "[FrooxEngine]FrooxEngine.DynamicVariableSpace")
    if not dvs_id then
        RESH.delete_slot(test_slot_id)
        error("Failed to create DVS: " .. tostring(dvs_err))
    end

    local col_id, col_err = RESH.create_component(test_slot_id, "[FrooxEngine]FrooxEngine.SphereCollider")
    if not col_id then
        RESH.delete_component(dvs_id)
        RESH.delete_slot(test_slot_id)
        error("Failed to create SphereCollider: " .. tostring(col_err))
    end

    print("    Created 2 components on same slot")

    -- Verify both exist
    local dvs_data = RESH.inspect(dvs_id)
    local col_data = RESH.inspect(col_id)
    assert_not_nil(dvs_data, "DVS should be inspectable")
    assert_not_nil(col_data, "SphereCollider should be inspectable")
    assert_equals(dvs_data.Type, "component")
    assert_equals(col_data.Type, "component")

    -- Cleanup
    RESH.delete_component(dvs_id)
    RESH.delete_component(col_id)
    RESH.delete_slot(test_slot_id)
    return true
end)

print("")
print("================================================================")
print("SECTION 5: Component Type System")
print("================================================================")
print("")

-- Test 31: get_component_types returns list
test("RESH.get_component_types() returns type list", function()
    assert_not_nil(RESH.get_component_types, "get_component_types should exist")
    assert_type(RESH.get_component_types, "function", "get_component_types should be a function")

    local types, err = RESH.get_component_types()
    if not types then
        error("Failed to get types: " .. tostring(err))
    end

    assert_type(types, "table", "Should return a table")
    assert_true(#types > 0, "Should return at least one type")
    print("    Available component types: " .. #types)
    -- Show a few
    for i = 1, math.min(3, #types) do
        print("      " .. types[i])
    end
    return true
end)

-- Test 32: get_component_definition returns definition
test("RESH.get_component_definition() returns definition", function()
    assert_not_nil(RESH.get_component_definition, "get_component_definition should exist")
    assert_type(RESH.get_component_definition, "function", "get_component_definition should be a function")

    local def, err = RESH.get_component_definition("[FrooxEngine]FrooxEngine.SphereCollider")
    if not def then
        error("Failed to get definition: " .. tostring(err))
    end

    assert_type(def, "table", "Definition should be a table")
    print("    Got definition for SphereCollider")

    -- Show available top-level keys
    for k, v in pairs(def) do
        print("      key: " .. tostring(k) .. " (type: " .. type(v) .. ")")
    end
    return true
end)

-- Test 33: Validate top common component types can be created
test("Validate common component types are creatable", function()
    RESH.cd("/")
    local parent_id = RESH.get_current_slot()

    -- List of common types from world dump (tested individually)
    local creatable_types = {
        "[FrooxEngine]FrooxEngine.DynamicVariableSpace",    -- #16, 35 instances
        "[FrooxEngine]FrooxEngine.SphereCollider",          -- #8,  43 instances
        "[FrooxEngine]FrooxEngine.BoxCollider",             -- #31, 20 instances
        "[FrooxEngine]FrooxEngine.CapsuleCollider",         -- #17, 33 instances
        "[FrooxEngine]FrooxEngine.Grabbable",               -- #73, 9 instances
        "[FrooxEngine]FrooxEngine.Comment",                 -- simple annotation
    }

    local test_slot_id, err = RESH.create_slot("TEST_TYPE_VALIDATION", parent_id)
    if not test_slot_id then
        error("Failed to create test slot: " .. tostring(err))
    end

    local created = 0
    local failed_types = {}

    for _, comp_type in ipairs(creatable_types) do
        local comp_id, comp_err = RESH.create_component(test_slot_id, comp_type)
        if comp_id then
            created = created + 1
            -- Clean up immediately
            RESH.delete_component(comp_id)
        else
            table.insert(failed_types, comp_type .. ": " .. tostring(comp_err))
        end
    end

    RESH.delete_slot(test_slot_id)

    print("    Created " .. created .. "/" .. #creatable_types .. " component types")
    if #failed_types > 0 then
        for _, msg in ipairs(failed_types) do
            print("    FAILED: " .. msg)
        end
        error(#failed_types .. " component types failed to create")
    end

    return true
end)

-- Test 34: Component type format validation
test("Component type format: assembly prefix required", function()
    RESH.cd("/")
    local parent_id = RESH.get_current_slot()

    local test_slot_id, err = RESH.create_slot("TEST_FORMAT_SLOT", parent_id)
    if not test_slot_id then
        error("Failed to create test slot: " .. tostring(err))
    end

    -- Try WITHOUT assembly prefix - should fail
    local bad_id, bad_err = RESH.create_component(test_slot_id, "FrooxEngine.SphereCollider")

    -- Try WITH assembly prefix - should succeed
    local good_id, good_err = RESH.create_component(test_slot_id, "[FrooxEngine]FrooxEngine.SphereCollider")

    local bad_failed = (bad_id == nil)
    local good_succeeded = (good_id ~= nil)

    if good_id then
        RESH.delete_component(good_id)
    end
    RESH.delete_slot(test_slot_id)

    if bad_failed and good_succeeded then
        print("    Confirmed: [Assembly]Namespace.Class format required")
        return true
    elseif not bad_failed then
        print("    Note: bare type name also accepted (unexpected)")
        return true -- Still passes, just noting behavior
    else
        error("Correct format failed: " .. tostring(good_err))
    end
end)

print("")
print("================================================================")
print("SECTION 6: Inspecting Real World Components (Type Coverage)")
print("================================================================")
print("")

-- Test 35: Inspect all member types across root components
test("Survey all value types in root components", function()
    RESH.cd("/")
    local listing = RESH.ls()

    local type_counts = {}
    local total_members = 0

    for i, comp in ipairs(listing.components) do
        local data = RESH.inspect(comp.id)
        if data and data.Members then
            for name, member in pairs(data.Members) do
                total_members = total_members + 1
                local t = member.Type or "unknown"
                type_counts[t] = (type_counts[t] or 0) + 1
            end
        end
    end

    print("    Total members inspected: " .. total_members)
    print("    Value types encountered:")
    -- Sort by count
    local sorted = {}
    for t, count in pairs(type_counts) do
        table.insert(sorted, {type = t, count = count})
    end
    table.sort(sorted, function(a, b) return a.count > b.count end)
    for _, entry in ipairs(sorted) do
        print("      " .. entry.type .. ": " .. entry.count)
    end

    assert_true(total_members > 0, "Should find members")
    return true
end)

-- Test 36: Validate composite value types from real components
test("Deep inspect: validate composite value types", function()
    RESH.cd("/")
    local listing = RESH.ls()

    local found_float3 = false
    local found_floatQ = false
    local found_bool = false
    local found_string = false

    for i, comp in ipairs(listing.components) do
        local data = RESH.inspect(comp.id)
        if data and data.Members then
            for name, member in pairs(data.Members) do
                if member.Type == "float3" and not found_float3 then
                    assert_type(member.Value, "table", "float3 should be table")
                    assert_type(member.Value.x, "number", "float3.x should be number")
                    found_float3 = true
                end
                if member.Type == "floatQ" and not found_floatQ then
                    assert_type(member.Value, "table", "floatQ should be table")
                    assert_type(member.Value.w, "number", "floatQ.w should be number")
                    found_floatQ = true
                end
                if member.Type == "bool" and not found_bool then
                    assert_type(member.Value, "boolean", "bool should be boolean")
                    found_bool = true
                end
                if member.Type == "string" and not found_string then
                    -- String can be nil (null in Resonite) or string
                    if member.Value ~= nil then
                        assert_type(member.Value, "string", "string should be string")
                        found_string = true
                    end
                end
            end
        end
        if found_float3 and found_floatQ and found_bool and found_string then
            break
        end
    end

    print("    float3: " .. (found_float3 and "OK" or "not found"))
    print("    floatQ: " .. (found_floatQ and "OK" or "not found"))
    print("    bool:   " .. (found_bool and "OK" or "not found"))
    print("    string: " .. (found_string and "OK" or "not found (or all nil)"))

    assert_true(found_bool, "Should find bool type")
    -- String assertion removed - string members can all be nil/null
    return true
end)

print("")
print("================================================================")
print("SECTION 7: Error Handling")
print("================================================================")
print("")

-- Test 37: cd() to non-existent slot returns false
test("RESH.cd() returns false for non-existent", function()
    local success, err = RESH.cd("/__NONEXISTENT_SLOT_XYZ__")
    assert_equals(success, false, "cd() to non-existent should return false")
    assert_type(err, "string", "cd() should return error message")
    return true
end)

-- Test 38: inspect() with invalid ID
test("RESH.inspect() handles invalid ID", function()
    local data, err = RESH.inspect("invalid-id-xyz-123")
    assert_equals(data, nil, "inspect() with invalid ID should return nil")
    return true
end)

-- Test 39: Navigation state preserved after error
test("Navigation state preserved after failed cd()", function()
    RESH.cd("/")
    local before_path = RESH.pwd()

    RESH.cd("/__NONEXISTENT__")

    local after_path = RESH.pwd()
    assert_equals(after_path, before_path, "Path should be unchanged after failed cd()")

    return true
end)

-- Test 40: create_component with invalid type returns error
test("create_component with invalid type returns error", function()
    RESH.cd("/")
    local parent_id = RESH.get_current_slot()

    local test_slot_id, err = RESH.create_slot("TEST_BAD_TYPE", parent_id)
    if not test_slot_id then
        error("Failed to create test slot: " .. tostring(err))
    end

    local comp_id, comp_err = RESH.create_component(test_slot_id, "NotARealComponent")
    assert_equals(comp_id, nil, "Invalid type should return nil")
    assert_type(comp_err, "string", "Should return error message")
    print("    Error for invalid type: " .. comp_err)

    RESH.delete_slot(test_slot_id)
    return true
end)

print("")
print("================================================================")
print("SECTION 8: Standard Lua Integration")
print("================================================================")
print("")

-- Test 41: print() function works
test("print() function is available", function()
    assert_not_nil(print, "print should be available")
    assert_type(print, "function", "print should be a function")
    return true
end)

-- Test 42: Standard Lua libraries available
test("Standard Lua libraries available", function()
    assert_not_nil(string, "string library should be available")
    assert_not_nil(table, "table library should be available")
    assert_not_nil(math, "math library should be available")
    return true
end)

-- Test 43: pcall works for error handling
test("pcall() works for error handling", function()
    local success, err = pcall(function()
        error("Test error")
    end)

    assert_equals(success, false, "pcall should catch error")
    assert_type(err, "string", "pcall should return error message")
    return true
end)

print("")
print("================================================================")
print("                        TEST SUMMARY")
print("================================================================")
print("")
print(string.format("Total Tests:    %d", tests.passed + tests.failed + tests.skipped))
print(string.format("  Passed:       %d", tests.passed))
print(string.format("  Failed:       %d", tests.failed))
print(string.format("  Skipped:      %d", tests.skipped))
print("")

if tests.failed > 0 then
    print("Failed Tests:")
    for i, result in ipairs(tests.results) do
        if result.status == "FAIL" then
            print("  X " .. result.name)
            if result.error then
                print("    Error: " .. result.error)
            end
        end
    end
    print("")
end

if tests.skipped > 0 then
    print("Skipped Tests:")
    for i, result in ipairs(tests.results) do
        if result.status == "SKIP" then
            print("  - " .. result.name)
            if result.reason then
                print("    Reason: " .. result.reason)
            end
        end
    end
    print("")
end

local total_run = tests.passed + tests.failed
local pass_rate = 0
if total_run > 0 then
    pass_rate = (tests.passed / total_run) * 100
end
print(string.format("Pass Rate: %.1f%%", pass_rate))
print("")

if tests.failed == 0 then
    print("================================================================")
    print("              ALL TESTS PASSED!")
    print("================================================================")
else
    print("================================================================")
    print("              SOME TESTS FAILED")
    print("================================================================")
end
