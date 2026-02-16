-- Comprehensive World Dump Script
-- Recursively scans entire world hierarchy and saves to file

print("╔════════════════════════════════════════════════════════════════╗")
print("║              Resonite World Dump Utility                      ║")
print("╚════════════════════════════════════════════════════════════════╝")
print("")

-- Output buffer
local output = {}
local function write(text)
    table.insert(output, text)
end

-- Statistics
local stats = {
    total_slots = 0,
    total_components = 0,
    max_depth = 0,
    component_types = {}
}

-- Helper to format slot info
local function format_slot_info(slot_data)
    local info = {}
    
    if slot_data.Name then
        table.insert(info, "Name: " .. tostring(slot_data.Name))
    end
    if slot_data.Tag and slot_data.Tag ~= "" then
        table.insert(info, "Tag: " .. tostring(slot_data.Tag))
    end
    if slot_data.Position then
        local pos = tostring(slot_data.Position)
        table.insert(info, "Pos: " .. pos)
    end
    
    return table.concat(info, ", ")
end

-- Recursive function to scan slot and children
local function scan_slot(slot_id, depth, indent)
    stats.total_slots = stats.total_slots + 1
    if depth > stats.max_depth then
        stats.max_depth = depth
    end
    
    -- Get slot data
    local slot_data = RESH.inspect(slot_id)
    if not slot_data then
        write(indent .. "✗ Failed to inspect slot: " .. slot_id)
        return
    end
    
    -- Write slot header
    local slot_info = format_slot_info(slot_data)
    write(indent .. "├─ SLOT [" .. slot_id .. "] " .. slot_info)
    
    -- Navigate to this slot to list its contents
    local current_path = RESH.pwd()
    
    -- Try to find and navigate to this slot
    local listing = nil
    if slot_id == "Root" then
        RESH.cd("/")
        listing = RESH.ls()
    else
        -- We're already at parent, try to navigate by name
        listing = RESH.ls()
        local found = false
        for _, child in ipairs(listing.children) do
            if child.id == slot_id then
                if RESH.cd(child.name) then
                    listing = RESH.ls()
                    found = true
                    break
                end
            end
        end
        
        if not found then
            write(indent .. "│  └─ (navigation failed)")
            return
        end
    end
    
    -- List components
    if listing and #listing.components > 0 then
        write(indent .. "│  Components (" .. #listing.components .. "):")
        for i, comp in ipairs(listing.components) do
            stats.total_components = stats.total_components + 1
            
            -- Track component type
            if not stats.component_types[comp.type] then
                stats.component_types[comp.type] = 0
            end
            stats.component_types[comp.type] = stats.component_types[comp.type] + 1
            
            -- Get component details
            local comp_data = RESH.inspect(comp.id)
            local member_count = 0
            if comp_data and comp_data.Members then
                for _ in pairs(comp_data.Members) do
                    member_count = member_count + 1
                end
            end
            
            local prefix = "│  "
            if i == #listing.components then
                prefix = "│  "
            end
            
            write(indent .. prefix .. "├─ COMP: " .. comp.type)
            write(indent .. prefix .. "│  ID: " .. comp.id .. ", Members: " .. member_count)
        end
    end
    
    -- Recursively scan children
    if listing and #listing.children > 0 then
        write(indent .. "│  Children (" .. #listing.children .. "):")
        for i, child in ipairs(listing.children) do
            local child_indent = indent .. "│  "
            if i == #listing.children then
                child_indent = indent .. "   "
            end
            
            scan_slot(child.id, depth + 1, child_indent)
            
            -- Navigate back to parent
            RESH.cd("..")
        end
    end
end

-- Start the scan
print("Starting world scan from Root...")
print("")

write("════════════════════════════════════════════════════════════════")
write("RESONITE WORLD HIERARCHY DUMP")
write("════════════════════════════════════════════════════════════════")
write("")

-- Scan from root
RESH.cd("/")
scan_slot("Root", 0, "")

-- Write statistics
write("")
write("════════════════════════════════════════════════════════════════")
write("STATISTICS")
write("════════════════════════════════════════════════════════════════")
write("")
write("Total Slots: " .. stats.total_slots)
write("Total Components: " .. stats.total_components)
write("Max Depth: " .. stats.max_depth)
write("")
write("Component Types (" .. #stats.component_types .. " unique):")
write("")

-- Sort component types by usage
local sorted_types = {}
for comp_type, count in pairs(stats.component_types) do
    table.insert(sorted_types, {type = comp_type, count = count})
end

table.sort(sorted_types, function(a, b)
    return a.count > b.count
end)

for i, entry in ipairs(sorted_types) do
    write(string.format("  %3d | %-60s | Count: %d", i, entry.type, entry.count))
end

write("")
write("════════════════════════════════════════════════════════════════")
write("END OF DUMP")
write("════════════════════════════════════════════════════════════════")

-- Write output to a file
-- Note: Since Lua sandboxing might prevent file I/O, we'll print to console
-- and user can redirect output: go run ./ script dump_world.lua > world_dump.txt

print("")
print("═══════════════════════════════════════════════════════════════")
print("Scan complete!")
print("═══════════════════════════════════════════════════════════════")
print("")
print("Statistics:")
print("  Slots: " .. stats.total_slots)
print("  Components: " .. stats.total_components)
print("  Max Depth: " .. stats.max_depth)
print("  Unique Component Types: " .. #sorted_types)
print("")
print("Generating output...")
print("")
print("")

-- Print all output
for _, line in ipairs(output) do
    print(line)
end

print("")
print("")
print("╔════════════════════════════════════════════════════════════════╗")
print("║                    Dump Complete!                              ║")
print("║                                                                ║")
print("║  To save to file, run with output redirection:                ║")
print("║  go run ./ script examples/dump_world.lua > world_dump.txt    ║")
print("╚════════════════════════════════════════════════════════════════╝")
