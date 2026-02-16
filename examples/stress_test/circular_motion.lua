-- Circular Motion Stress Test
-- Moves an object tagged "EXAMPLE_RESH_TEST" in a circle around world origin
-- NOTE: This requires write operations which are not yet implemented!
-- This script demonstrates the pattern and will work once RESH.update_slot() is added

print("=== Circular Motion Stress Test ===")
print("")

-- Configuration
local TAG = "EXAMPLE_RESH_TEST"
local RADIUS = 2.0          -- Circle radius
local HEIGHT = 1.5          -- Y height
local SPEED = 1.0           -- Rotations per second
local DURATION = 10         -- Run for 10 seconds
local UPDATE_RATE = 60      -- Updates per second (60 FPS)

-- Find the test object
print("Finding object with tag: " .. TAG)
RESH.cd("/")

local function find_by_tag(tag)
    local listing = RESH.ls()
    for i, child in ipairs(listing.children) do
        RESH.cd(child.name)
        local slot = RESH.inspect(RESH.get_current_slot())
        
        if slot.Tag == tag then
            print("Found: " .. slot.Name .. " at " .. RESH.pwd())
            RESH.cd("..")
            return child.id, slot
        end
        
        RESH.cd("..")
    end
    return nil, nil
end

local object_id, object_slot = find_by_tag(TAG)

if not object_id then
    print("")
    print("ERROR: Object with tag '" .. TAG .. "' not found!")
    print("")
    print("Please create an object at world root with:")
    print("  1. Tag: " .. TAG)
    print("  2. Any name (e.g., 'TestSphere')")
    print("  3. Position at origin initially")
    return
end

print("Object ID: " .. object_id)

-- Check if Position exists and print it
if object_slot.Position then
    if type(object_slot.Position) == "table" then
        print("Current Position: (" .. 
            tostring(object_slot.Position.x or "nil") .. ", " .. 
            tostring(object_slot.Position.y or "nil") .. ", " .. 
            tostring(object_slot.Position.z or "nil") .. ")")
    else
        print("Current Position: " .. tostring(object_slot.Position))
    end
else
    print("Current Position: <not available>")
    print("Note: Position data might not be available in slot inspection")
end
print("")

-- Calculate circular motion
local function calculate_position(time)
    local angle = time * SPEED * 2 * math.pi
    local x = math.cos(angle) * RADIUS
    local z = math.sin(angle) * RADIUS
    local y = HEIGHT
    
    return {x = x, y = y, z = z}
end

-- Stress test loop
print("Starting circular motion...")
print("  Radius: " .. RADIUS)
print("  Height: " .. HEIGHT)
print("  Speed: " .. SPEED .. " rotations/sec")
print("  Duration: " .. DURATION .. " seconds")
print("  Update Rate: " .. UPDATE_RATE .. " Hz")
print("")

local start_time = os.clock()
local updates = 0

while true do
    local elapsed = os.clock() - start_time
    if elapsed >= DURATION then
        break
    end
    
    -- Calculate new position
    local pos = calculate_position(elapsed)
    
    -- TODO: Update the object position
    -- RESH.update_slot(object_id, {Position = pos})
    -- ^ This function doesn't exist yet!
    
    -- For now, just print periodic updates
    if updates % UPDATE_RATE == 0 then
        print(string.format("%.2fs: Position (%.2f, %.2f, %.2f)", 
            elapsed, pos.x, pos.y, pos.z))
    end
    
    updates = updates + 1
    
    -- Sleep to maintain update rate
    -- Note: Lua doesn't have built-in sleep, would need os.execute("timeout /t 0 /nobreak")
    -- For now, just calculate positions as fast as possible
end

local end_time = os.clock()
local actual_duration = end_time - start_time

print("")
print("=== Stress Test Complete ===")
print("Updates: " .. updates)
print("Duration: " .. string.format("%.2f", actual_duration) .. " seconds")
print("Average Rate: " .. string.format("%.2f", updates / actual_duration) .. " updates/sec")
print("")
print("NOTE: Position updates are calculated but not applied.")
print("Waiting for RESH.update_slot() to be implemented!")
