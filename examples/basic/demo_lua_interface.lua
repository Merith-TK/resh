-- Demo script showing improved Lua data access
print("=== Lua Data Interface Demo ===")
print("")

-- Navigate to root
RESH.cd("/")
print("Current location: " .. RESH.pwd())
print("")

-- List current directory
print("Listing root:")
local listing = RESH.ls()

-- Access children array
print("Children (" .. #listing.children .. " total):")
for i, child in ipairs(listing.children) do
    -- Direct access to properties
    local status = child.active and "active" or "inactive"
    print("  " .. child.name .. " [" .. status .. "]")
    if i >= 5 then
        print("  ... (showing first 5)")
        break
    end
end
print("")

-- Access components array
print("Components (" .. #listing.components .. " total):")
for i, comp in ipairs(listing.components) do
    print("  " .. comp.type)
    if i >= 5 then
        print("  ... (showing first 5)")
        break
    end
end
print("")

-- Inspect a component (if any exist)
if #listing.components > 0 then
    local first_comp = listing.components[1]
    print("Inspecting first component: " .. first_comp.type)
    print("")
    
    local comp_data = RESH.inspect(first_comp.id)
    
    -- Component metadata
    print("Type: " .. comp_data.TypeName)
    print("Full Type: " .. comp_data.ComponentType)
    print("")
    
    -- Members accessed by name (not index!)
    print("Members:")
    local count = 0
    for name, member in pairs(comp_data.Members) do
        count = count + 1
        
        -- Type-aware value access
        local value_str
        if member.Type == "reference" then
            value_str = member.TargetId or "<null>"
        elseif type(member.Value) == "table" then
            value_str = "<table>"
        else
            value_str = tostring(member.Value)
        end
        
        print("  " .. name .. ":")
        print("    Type: " .. member.Type)
        print("    Value: " .. value_str)
        print("    ID: " .. member.ID)
        
        if count >= 3 then
            print("  ... (showing first 3)")
            break
        end
    end
    print("")
end

-- Inspect a slot
if #listing.children > 0 then
    local first_child = listing.children[1]
    print("Inspecting first child slot: " .. first_child.name)
    print("")
    
    local slot_data = RESH.inspect(first_child.id)
    
    -- Slot properties accessed directly by name
    print("Slot Properties:")
    print("  Name: " .. (slot_data.Name or "N/A"))
    print("  Active: " .. tostring(slot_data.Active))
    print("  Persistent: " .. tostring(slot_data.Persistent))
    print("  Tag: " .. (slot_data.Tag or "<none>"))
    
    -- Position, Rotation, Scale are structured values
    if slot_data.Position then
        print("  Position: " .. tostring(slot_data.Position))
    end
    print("")
end

print("=== Demo Complete ===")
print("")
print("Key improvements:")
print("- Component members indexed by NAME (not array)")
print("- Slot properties directly accessible")
print("- Type-aware value conversion (bool, number, string)")
print("- References have TargetId field")
print("- Complex values preserved as tables")
