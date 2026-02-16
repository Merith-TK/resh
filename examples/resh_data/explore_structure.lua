-- Comprehensive RESH.DATA Structure Explorer
-- This script thoroughly examines the RESH.DATA slot and reports everything

print("=" .. string.rep("=", 60))
print("RESH.DATA Structure Explorer")
print("=" .. string.rep("=", 60))
print("")

-- Navigate to root
RESH.cd("/")

-- Try to find RESH slot first (might be RESH or RESH.DATA)
print("🔍 Searching for RESH slot...")
local resh_names = {"RESH", "RESH.DATA", "Resh", "resh"}
local resh_id = nil
local resh_name = nil

for _, name in ipairs(resh_names) do
    local found = RESH.find_slot(name)
    if found then
        resh_id = found
        resh_name = name
        print("✓ Found: " .. name .. " (" .. found .. ")")
        break
    end
end

if not resh_id then
    print("✗ No RESH/RESH.DATA slot found!")
    print("")
    print("Available children in root:")
    local listing = RESH.ls()
    for i, child in ipairs(listing.children) do
        print("  " .. child.name)
    end
    print("")
    print("Please create a RESH or RESH.DATA slot first.")
    return
end

print("")
print("=" .. string.rep("=", 60))
print("RESH SLOT: " .. resh_name)
print("=" .. string.rep("=", 60))
print("")

-- Navigate to RESH
RESH.cd(resh_name)

-- Inspect the RESH slot itself
local resh_slot = RESH.inspect(resh_id)
print("📦 Slot Properties:")
print("  ID: " .. (resh_slot.ID or "N/A"))
print("  Name: " .. (resh_slot.Name or "N/A"))
print("  Tag: " .. (resh_slot.Tag or "<none>"))
print("  Active: " .. tostring(resh_slot.Active))
print("  Persistent: " .. tostring(resh_slot.Persistent))
if resh_slot.OrderOffset then
    print("  OrderOffset: " .. tostring(resh_slot.OrderOffset))
end
print("")

-- List all components on RESH slot
local resh_listing = RESH.ls()
print("🔧 Components on RESH slot (" .. #resh_listing.components .. " total):")
print("")

for i, comp in ipairs(resh_listing.components) do
    print("[" .. i .. "] " .. comp.type)
    print("    ID: " .. comp.id)
    print("    Persistent: " .. tostring(comp.persistent))
    
    -- Inspect this component
    local comp_data = RESH.inspect(comp.id)
    if comp_data.Members then
        print("    Members:")
        for name, member in pairs(comp_data.Members) do
            local value_str
            if member.Type == "reference" then
                value_str = member.TargetId or "<null>"
            elseif type(member.Value) == "table" then
                value_str = "{ "
                local parts = {}
                for k, v in pairs(member.Value) do
                    table.insert(parts, k .. "=" .. tostring(v))
                end
                value_str = value_str .. table.concat(parts, ", ") .. " }"
            else
                value_str = tostring(member.Value)
            end
            print("      • " .. name .. " (" .. member.Type .. ") = " .. value_str)
        end
    end
    print("")
end

-- List child slots
print("=" .. string.rep("=", 60))
print("📁 Child Slots (" .. #resh_listing.children .. " total):")
print("=" .. string.rep("=", 60))
print("")

if #resh_listing.children == 0 then
    print("  (no children)")
    print("")
else
    for i, child in ipairs(resh_listing.children) do
        print("[" .. i .. "] " .. child.name)
        print("    ID: " .. child.id)
        print("    Active: " .. tostring(child.active))
        print("    Persistent: " .. tostring(child.persistent))
        
        -- Navigate into child
        RESH.cd(child.name)
        
        -- Get child slot details
        local child_slot = RESH.inspect(child.id)
        if child_slot.Tag and child_slot.Tag ~= "" then
            print("    Tag: " .. child_slot.Tag)
        end
        
        -- List child's components
        local child_listing = RESH.ls()
        print("    Components: " .. #child_listing.components)
        
        for j, comp in ipairs(child_listing.components) do
            print("      [" .. j .. "] " .. comp.type)
            
            -- Inspect component
            local comp_data = RESH.inspect(comp.id)
            if comp_data.Members then
                for name, member in pairs(comp_data.Members) do
                    local value_str
                    if member.Type == "reference" then
                        value_str = member.TargetId or "<null>"
                    elseif type(member.Value) == "table" then
                        -- Format tables nicely
                        if member.Value.x and member.Value.y then
                            value_str = string.format("(%.2f, %.2f, %.2f)", 
                                member.Value.x or 0, 
                                member.Value.y or 0, 
                                member.Value.z or 0)
                        else
                            value_str = "<table>"
                        end
                    else
                        value_str = tostring(member.Value)
                    end
                    print("          • " .. name .. " = " .. value_str)
                end
            end
        end
        
        -- Check for sub-children
        if #child_listing.children > 0 then
            print("    Children: " .. #child_listing.children)
            for k, subchild in ipairs(child_listing.children) do
                print("      [" .. k .. "] " .. subchild.name)
            end
        end
        
        -- Go back up
        RESH.cd("..")
        print("")
    end
end

print("=" .. string.rep("=", 60))
print("📊 Summary")
print("=" .. string.rep("=", 60))
print("  RESH Slot: " .. resh_name)
print("  Total Components: " .. #resh_listing.components)
print("  Total Children: " .. #resh_listing.children)
print("")

-- Look for specific patterns
print("🔎 Pattern Analysis:")
print("")

-- Check for DynamicVariableSpace
local has_var_space = false
local var_space_name = nil
local direct_binding = nil

for _, comp in ipairs(resh_listing.components) do
    if string.match(comp.type, "DynamicVariableSpace") then
        has_var_space = true
        local comp_data = RESH.inspect(comp.id)
        if comp_data.Members then
            if comp_data.Members.SpaceName then
                var_space_name = comp_data.Members.SpaceName.Value
            end
            if comp_data.Members.OnlyDirectBinding then
                direct_binding = comp_data.Members.OnlyDirectBinding.Value
            end
        end
    end
end

if has_var_space then
    print("  ✓ Has DynamicVariableSpace")
    print("    SpaceName: " .. (var_space_name or "N/A"))
    print("    OnlyDirectBinding: " .. tostring(direct_binding))
else
    print("  ✗ No DynamicVariableSpace found")
end
print("")

-- Check for DynamicReferenceVariable
local ref_vars = {}
for _, comp in ipairs(resh_listing.components) do
    if string.match(comp.type, "DynamicReferenceVariable") then
        local comp_data = RESH.inspect(comp.id)
        if comp_data.Members and comp_data.Members.VariableName then
            table.insert(ref_vars, comp_data.Members.VariableName.Value)
        end
    end
end

if #ref_vars > 0 then
    print("  ✓ Has " .. #ref_vars .. " DynamicReferenceVariable(s):")
    for _, var_name in ipairs(ref_vars) do
        print("    • " .. var_name)
    end
else
    print("  ✗ No DynamicReferenceVariable found")
end
print("")

-- Analyze children for bookmark pattern
if #resh_listing.children > 0 then
    print("  📂 Analyzing child slots for bookmark pattern:")
    for _, child in ipairs(resh_listing.children) do
        RESH.cd(child.name)
        local child_listing = RESH.ls()
        
        local has_ref_var = false
        local var_name = nil
        local target_id = nil
        
        for _, comp in ipairs(child_listing.components) do
            if string.match(comp.type, "DynamicReferenceVariable") then
                has_ref_var = true
                local comp_data = RESH.inspect(comp.id)
                if comp_data.Members then
                    if comp_data.Members.VariableName then
                        var_name = comp_data.Members.VariableName.Value
                    end
                    if comp_data.Members.Reference then
                        target_id = comp_data.Members.Reference.TargetId
                    end
                end
            end
        end
        
        RESH.cd("..")
        
        if has_ref_var then
            print("    ✓ " .. child.name .. ":")
            print("      Variable: " .. (var_name or "N/A"))
            print("      Target: " .. (target_id or "<null>"))
        end
    end
end

print("")
print("=" .. string.rep("=", 60))
print("✅ Exploration Complete!")
print("=" .. string.rep("=", 60))
