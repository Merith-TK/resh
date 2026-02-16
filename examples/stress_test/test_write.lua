-- Test write operations by updating object position
print("=== Write Operation Test ===")
print("")

local TAG = "EXAMPLE_RESH_TEST"

-- Find the test object
print("Finding object with tag: " .. TAG)
RESH.cd("/")

local function find_by_tag(tag)
    local listing = RESH.ls()
    for i, child in ipairs(listing.children) do
        local slot = RESH.inspect(child.id)
        if slot.Tag == tag then
            print("Found: " .. slot.Name)
            return child.id
        end
    end
    return nil
end

local object_id = find_by_tag(TAG)

if not object_id then
    print("ERROR: Object not found!")
    return
end

print("Object ID: " .. object_id)
print("")

-- Navigate to the object and list its components
RESH.cd("/")
local listing = RESH.ls()
for i, child in ipairs(listing.children) do
    if child.id == object_id then
        RESH.cd(child.name)
        break
    end
end

print("Listing components on object:")
local comp_listing = RESH.ls()
for i, comp in ipairs(comp_listing.components) do
    print("  [" .. i .. "] " .. comp.type .. " (" .. comp.id .. ")")
end
print("")

-- Try to find and update the Slot component which should have Position
local slot_comp = nil
for i, comp in ipairs(comp_listing.components) do
    if comp.type:match("Slot") or comp.type:match("Transform") then
        print("Found potential position component: " .. comp.type)
        slot_comp = comp
        break
    end
end

if not slot_comp then
    print("No Slot/Transform component found, trying to update first component...")
    if #comp_listing.components > 0 then
        slot_comp = comp_listing.components[1]
    else
        print("ERROR: No components found on object!")
        return
    end
end

print("Using component: " .. slot_comp.type .. " (" .. slot_comp.id .. ")")
print("")

-- Inspect the component to see its members
local comp_data = RESH.inspect(slot_comp.id)
print("Component members:")
for name, member in pairs(comp_data.Members) do
    print("  " .. name .. " (" .. member.Type .. ")")
end
print("")

-- Try to update the Position member
if comp_data.Members["Position"] then
    print("Attempting to update Position...")
    
    local new_position = {
        x = 0.0,
        y = 1.5,
        z = 2.0
    }
    
    local success, err = RESH.update_component(slot_comp.id, {
        Position = new_position
    })
    
    if success then
        print("✓ Position updated successfully!")
        print("  New position: (" .. new_position.x .. ", " .. new_position.y .. ", " .. new_position.z .. ")")
    else
        print("✗ Update failed: " .. err)
    end
else
    print("WARNING: No Position member found in component")
    print("")
    print("Available members to update:")
    for name, member in pairs(comp_data.Members) do
        print("  - " .. name)
    end
end

print("")
print("=== Test Complete ===")
