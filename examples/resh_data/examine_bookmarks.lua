-- Deep dive into bookmark structure
print("=" .. string.rep("=", 60))
print("Bookmark Structure Deep Dive")
print("=" .. string.rep("=", 60))
print("")

RESH.cd("/")
RESH.cd("RESH.DATA")
RESH.cd("Bookmarks")

print("📂 Inside RESH.DATA/Bookmarks")
print("")

local listing = RESH.ls()
print("Children: " .. #listing.children)
print("")

-- Examine each bookmark child
for i, child in ipairs(listing.children) do
    print("=" .. string.rep("-", 58))
    print("Bookmark [" .. i .. "]: " .. child.name)
    print("=" .. string.rep("-", 58))
    
    RESH.cd(child.name)
    
    -- Get slot properties
    local slot = RESH.inspect(child.id)
    print("Slot ID: " .. child.id)
    if slot.Tag and slot.Tag ~= "" then
        print("Tag: " .. slot.Tag)
    end
    print("Active: " .. tostring(child.active))
    print("Persistent: " .. tostring(child.persistent))
    print("")
    
    -- List components
    local child_listing = RESH.ls()
    print("Components (" .. #child_listing.components .. " total):")
    print("")
    
    for j, comp in ipairs(child_listing.components) do
        print("  [" .. j .. "] " .. comp.type)
        print("      ID: " .. comp.id)
        
        local comp_data = RESH.inspect(comp.id)
        if comp_data.Members then
            print("      Members:")
            for name, member in pairs(comp_data.Members) do
                local value_str
                
                if member.Type == "reference" then
                    if member.TargetId then
                        value_str = member.TargetId
                    else
                        value_str = "<null>"
                    end
                elseif type(member.Value) == "table" then
                    -- Check for common structures
                    if member.Value.x and member.Value.y then
                        if member.Value.w then
                            value_str = string.format("(%.2f, %.2f, %.2f, %.2f)", 
                                member.Value.x, member.Value.y, 
                                member.Value.z, member.Value.w)
                        else
                            value_str = string.format("(%.2f, %.2f, %.2f)", 
                                member.Value.x, member.Value.y, 
                                member.Value.z or 0)
                        end
                    else
                        -- Generic table
                        local parts = {}
                        for k, v in pairs(member.Value) do
                            table.insert(parts, k .. "=" .. tostring(v))
                        end
                        value_str = "{ " .. table.concat(parts, ", ") .. " }"
                    end
                else
                    value_str = tostring(member.Value)
                end
                
                print("        • " .. name .. " (" .. member.Type .. ")")
                print("          = " .. value_str)
            end
        end
        print("")
    end
    
    -- Check for sub-children
    if #child_listing.children > 0 then
        print("  Children (" .. #child_listing.children .. "):")
        for k, subchild in ipairs(child_listing.children) do
            print("    [" .. k .. "] " .. subchild.name)
        end
        print("")
    end
    
    RESH.cd("..")
end

print("")
print("=" .. string.rep("=", 60))
print("Key Findings:")
print("=" .. string.rep("=", 60))
print("")

-- Summary
RESH.cd("WorldObjects")
local wobj_listing = RESH.ls()
RESH.cd("..")

print("Bookmark Pattern:")
print("  • Location: RESH.DATA/Bookmarks/<bookmark_name>")
print("  • Example: 'WorldObjects'")
print("  • Components: " .. #wobj_listing.components)

-- Check for DynamicReferenceVariable pattern
for _, comp in ipairs(wobj_listing.components) do
    if string.match(comp.type, "DynamicReferenceVariable") then
        local comp_data = RESH.inspect(comp.id)
        if comp_data.Members then
            print("")
            print("Variable Configuration:")
            if comp_data.Members.VariableName then
                print("  • VariableName: " .. comp_data.Members.VariableName.Value)
            end
            if comp_data.Members.Reference then
                print("  • Reference Target: " .. (comp_data.Members.Reference.TargetId or "<null>"))
            end
        end
    end
end

print("")
print("To create a new bookmark:")
print("  1. Create child slot under RESH.DATA/Bookmarks/")
print("  2. Add DynamicReferenceVariable<Slot> component")
print("  3. Set VariableName to 'bookmark/<name>'")
print("  4. Set Reference to target slot ID")
print("")
