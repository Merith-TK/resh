-- Test script for RESH Lua scripting
print("=== RESH Test Script ===")
print("")

-- Get current location
local path = RESH.pwd()
print("Current path: " .. path)

local slot_id = RESH.get_current_slot()
print("Current slot: " .. slot_id)
print("")

-- List contents
print("Listing current directory:")
local listing = RESH.ls()

print("Children: " .. #listing.children)
for i, child in ipairs(listing.children) do
    print("  [" .. i .. "] " .. child.name .. " (" .. child.id .. ")")
end

print("")
print("Components: " .. #listing.components)
for i, comp in ipairs(listing.components) do
    print("  [" .. i .. "] " .. comp.type .. " (" .. comp.id .. ")")
end
print("")

-- Test navigation
print("Navigating to root...")
RESH.cd("/")
print("New path: " .. RESH.pwd())
print("")

-- Find a slot
print("Searching for 'RESH' slot...")
local found = RESH.find_slot("RESH")
if found then
    print("Found: " .. found)
else
    print("Not found")
end
print("")

print("=== Test Complete ===")
