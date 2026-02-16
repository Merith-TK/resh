-- Test script for RESH Lua scripting
print("=== RESH Test Script ===")
print("")

-- Get current location
local path = pwd()
print("Current path: " .. path)

local slot_id = get_current_slot()
print("Current slot: " .. slot_id)
print("")

-- List contents
print("Listing current directory:")
local listing = ls()

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
cd("/")
print("New path: " .. pwd())
print("")

-- Find a slot
print("Searching for 'RESH' slot...")
local found = find_slot("RESH")
if found then
    print("Found: " .. found)
else
    print("Not found")
end
print("")

print("=== Test Complete ===")
