-- Inspect RESH.DATA structure
-- This script navigates to /RESH.DATA and inspects its components and children

print("=== RESH.DATA Structure Inspector ===")
print("")

-- Navigate to root
print("Navigating to root...")
RESH.cd("/")

-- Find RESH.DATA
print("Looking for RESH.DATA slot...")
local resh_id = RESH.find_slot("RESH.DATA")

if not resh_id then
    print("ERROR: RESH.DATA slot not found!")
    print("Please create a RESH.DATA slot in your world first.")
    return
end

print("Found RESH.DATA: " .. resh_id)
print("")

-- Navigate to RESH.DATA
RESH.cd("RESH.DATA")
print("Inspecting RESH.DATA slot...")
print("")

-- Get slot data
local slot_data = RESH.inspect(resh_id)
print("Slot Properties:")
print("  Name: " .. (slot_data.Name or "N/A"))
print("  Tag: " .. (slot_data.Tag or "N/A"))
print("  Active: " .. tostring(slot_data.Active))
print("  Persistent: " .. tostring(slot_data.Persistent))
print("")

-- List components
local listing = RESH.ls()
print("Components on RESH.DATA:")
for i, comp in ipairs(listing.components) do
    print("  [" .. i .. "] " .. comp.type .. " (" .. comp.id .. ")")
    
    -- Inspect each component
    local comp_data = RESH.inspect(comp.id)
    if comp_data.Members then
        for name, member in pairs(comp_data.Members) do
            local value_str = "N/A"
            if member.Type == "reference" then
                value_str = member.TargetId or "<null>"
            else
                value_str = tostring(member.Value)
            end
            print("      " .. name .. " (" .. member.Type .. ") = " .. value_str)
        end
    end
    print("")
end

-- Look for Bookmarks child
print("Looking for Bookmarks child slot...")
local bookmarks_id = RESH.find_slot("Bookmarks")

if bookmarks_id then
    print("Found Bookmarks: " .. bookmarks_id)
    RESH.cd("Bookmarks")
    
    -- List Bookmarks components
    listing = RESH.ls()
    print("")
    print("Components on Bookmarks:")
    for i, comp in ipairs(listing.components) do
        print("  [" .. i .. "] " .. comp.type .. " (" .. comp.id .. ")")
        
        local comp_data = RESH.inspect(comp.id)
        if comp_data.Members then
            for name, member in pairs(comp_data.Members) do
                local value_str = "N/A"
                if member.Type == "reference" then
                    value_str = member.TargetId or "<null>"
                else
                    value_str = tostring(member.Value)
                end
                print("      " .. name .. " (" .. member.Type .. ") = " .. value_str)
            end
        end
        print("")
    end
    
    -- List bookmark slots
    print("Bookmark slots:")
    for i, child in ipairs(listing.children) do
        print("  [" .. i .. "] " .. child.name .. " (" .. child.id .. ")")
    end
else
    print("Bookmarks child slot not found")
end

print("")
print("=== Inspection Complete ===")
