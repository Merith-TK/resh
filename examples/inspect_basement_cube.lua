-- Comprehensive inspection of TheBasementCube slot
print("╔════════════════════════════════════════════════════════════════╗")
print("║         TheBasementCube Detailed Inspection                   ║")
print("╚════════════════════════════════════════════════════════════════╝")
print("")

-- Helper function to print tables recursively
local function print_table(tbl, indent, max_depth, current_depth)
    indent = indent or ""
    max_depth = max_depth or 3
    current_depth = current_depth or 0
    
    if current_depth >= max_depth then
        print(indent .. "  <max depth reached>")
        return
    end
    
    if type(tbl) ~= "table" then
        print(indent .. tostring(tbl))
        return
    end
    
    for key, value in pairs(tbl) do
        if type(value) == "table" then
            print(indent .. tostring(key) .. ":")
            print_table(value, indent .. "  ", max_depth, current_depth + 1)
        else
            print(indent .. tostring(key) .. " = " .. tostring(value))
        end
    end
end

-- Navigate to root
RESH.cd("/")
local listing = RESH.ls()

-- Find TheBasementCube
local cube_id = nil
local cube_name = nil

print("═══════════════════════════════════════════════════════════════")
print("STEP 1: Finding TheBasementCube")
print("═══════════════════════════════════════════════════════════════")
print("")

for i, child in ipairs(listing.children) do
    if child.name == "TheBasementCube" then
        cube_id = child.id
        cube_name = child.name
        print("✓ Found: " .. child.name)
        print("  ID: " .. child.id)
        print("  Active: " .. tostring(child.active))
        break
    end
end

if not cube_id then
    print("✗ TheBasementCube not found at root")
    print("")
    print("Available slots at root:")
    for i, child in ipairs(listing.children) do
        print("  - " .. child.name .. " (ID: " .. child.id .. ")")
    end
    return
end

print("")
print("═══════════════════════════════════════════════════════════════")
print("STEP 2: Inspecting Slot Structure")
print("═══════════════════════════════════════════════════════════════")
print("")

local slot_data = RESH.inspect(cube_id)

if slot_data then
    print("Slot Properties:")
    print("  Type: " .. tostring(slot_data.Type))
    print("  ID: " .. tostring(slot_data.ID))
    print("  Name: " .. tostring(slot_data.Name))
    
    -- Print all available slot properties
    print("")
    print("All Slot Data Fields:")
    for key, value in pairs(slot_data) do
        if key ~= "Type" and key ~= "ID" and key ~= "Name" then
            if type(value) == "table" then
                print("  " .. key .. ": (table)")
                print_table(value, "    ", 2)
            else
                print("  " .. key .. ": " .. tostring(value))
            end
        end
    end
else
    print("✗ Failed to inspect slot")
end

print("")
print("═══════════════════════════════════════════════════════════════")
print("STEP 3: Listing All Components")
print("═══════════════════════════════════════════════════════════════")
print("")

-- Navigate to the slot
RESH.cd(cube_name)
local components_list = RESH.ls()

print("Found " .. #components_list.components .. " components:")
print("")

for i, comp in ipairs(components_list.components) do
    print("Component #" .. i .. ":")
    print("  Type: " .. comp.type)
    print("  ID: " .. comp.id)
end

print("")
print("═══════════════════════════════════════════════════════════════")
print("STEP 4: Detailed Component Inspection")
print("═══════════════════════════════════════════════════════════════")
print("")

for i, comp in ipairs(components_list.components) do
    print("───────────────────────────────────────────────────────────────")
    print("Component #" .. i .. ": " .. comp.type)
    print("───────────────────────────────────────────────────────────────")
    print("")
    
    local comp_data = RESH.inspect(comp.id)
    
    if comp_data then
        print("  Component Type: " .. tostring(comp_data.TypeName))
        print("  ID: " .. tostring(comp_data.ID))
        print("")
        
        if comp_data.Members then
            local member_count = 0
            for name, member in pairs(comp_data.Members) do
                member_count = member_count + 1
            end
            
            print("  Members (" .. member_count .. " total):")
            print("")
            
            -- Print each member with details
            for name, member in pairs(comp_data.Members) do
                print("    " .. name .. ":")
                print("      Type: " .. tostring(member.Type))
                
                -- Handle different member types
                if member.Type == "bool" or member.Type == "int" or member.Type == "float" or 
                   member.Type == "double" or member.Type == "string" then
                    -- Simple types - show value directly
                    print("      Value: " .. tostring(member.Value))
                    
                elseif member.Type == "reference" then
                    -- References show TargetId
                    print("      TargetId: " .. tostring(member.TargetId))
                    
                elseif member.Type == "float3" or member.Type == "floatQ" or member.Type == "color" then
                    -- Complex types - show structure
                    if type(member.Value) == "table" then
                        print("      Value:")
                        for key, val in pairs(member.Value) do
                            print("        " .. key .. " = " .. tostring(val))
                        end
                    else
                        print("      Value: " .. tostring(member.Value))
                    end
                    
                else
                    -- Unknown type - dump what we have
                    print("      Data:")
                    print_table(member, "        ", 2)
                end
                print("")
            end
        else
            print("  No members found")
        end
    else
        print("  ✗ Failed to inspect component")
    end
    
    print("")
end

print("═══════════════════════════════════════════════════════════════")
print("STEP 5: Component Type Analysis")
print("═══════════════════════════════════════════════════════════════")
print("")

print("Component types found on TheBasementCube:")
for i, comp in ipairs(components_list.components) do
    print("  " .. i .. ". " .. comp.type)
end

print("")
print("These component types should be valid for creating new components!")

print("")
print("═══════════════════════════════════════════════════════════════")
print("STEP 6: Testing Component Creation")
print("═══════════════════════════════════════════════════════════════")
print("")

-- Create a test slot
RESH.cd("/")
local test_slot_id = RESH.create_slot("COMPONENT_TEST_SLOT", "Root")

if test_slot_id then
    print("✓ Created test slot: " .. test_slot_id)
    print("")
    
    -- Try to create a component using one of the types we found
    if #components_list.components > 0 then
        local test_comp_type = components_list.components[1].type
        print("Attempting to create component: " .. test_comp_type)
        
        local new_comp_id, err = RESH.create_component(test_slot_id, test_comp_type)
        
        if new_comp_id then
            print("✓ SUCCESS! Created component: " .. new_comp_id)
            print("")
            
            -- Inspect it
            local new_comp_data = RESH.inspect(new_comp_id)
            if new_comp_data then
                print("New component details:")
                print("  TypeName: " .. tostring(new_comp_data.TypeName))
                print("  ID: " .. tostring(new_comp_data.ID))
                
                if new_comp_data.Members then
                    local count = 0
                    for _ in pairs(new_comp_data.Members) do count = count + 1 end
                    print("  Members: " .. count)
                end
            end
            
            -- Clean up the component
            print("")
            print("Cleaning up test component...")
            RESH.delete_component(new_comp_id)
        else
            print("✗ Failed: " .. tostring(err))
        end
    end
    
    -- Clean up the test slot
    print("")
    print("Cleaning up test slot...")
    RESH.delete_slot(test_slot_id)
    print("✓ Test complete")
else
    print("✗ Failed to create test slot")
end

print("")
print("╔════════════════════════════════════════════════════════════════╗")
print("║                  Inspection Complete                           ║")
print("╚════════════════════════════════════════════════════════════════╝")
