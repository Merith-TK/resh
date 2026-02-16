package shell

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Merith-TK/resh/pkg/resolink"
	lua "github.com/yuin/gopher-lua"
)

// ScriptContext holds the context for script execution
type ScriptContext struct {
	Client *resolink.Client
	State  *State
	Output []string
}

// RunScript executes a Lua script with access to shell functions
func RunScript(client *resolink.Client, state *State, scriptPath string) error {
	L := lua.NewState()
	defer L.Close()

	ctx := &ScriptContext{
		Client: client,
		State:  state,
		Output: make([]string, 0),
	}

	// Register shell functions
	registerShellFunctions(L, ctx)

	// Execute script
	if err := L.DoFile(scriptPath); err != nil {
		return fmt.Errorf("script error: %w", err)
	}

	// Print any captured output
	for _, line := range ctx.Output {
		fmt.Println(line)
	}

	return nil
}

// registerShellFunctions registers shell operations as Lua functions under RESH table
func registerShellFunctions(L *lua.LState, ctx *ScriptContext) {
	// Create RESH table
	resh := L.NewTable()

	// Keep print as global (standard Lua function)
	L.SetGlobal("print", L.NewFunction(func(L *lua.LState) int {
		n := L.GetTop()
		parts := make([]string, n)
		for i := 1; i <= n; i++ {
			parts[i-1] = L.CheckAny(i).String()
		}
		output := strings.Join(parts, "\t")
		fmt.Println(output)
		ctx.Output = append(ctx.Output, output)
		return 0
	}))

	// cd function -> RESH.cd
	L.SetField(resh, "cd", L.NewFunction(func(L *lua.LState) int {
		target := L.CheckString(1)

		var err error
		switch target {
		case "/":
			NavigateToRoot(ctx.State)
		case "..":
			err = NavigateToParent(ctx.Client, ctx.State)
		default:
			err = NavigateToChild(ctx.Client, ctx.State, target)
		}

		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	// ls function -> RESH.ls - returns table of slot/component info
	L.SetField(resh, "ls", L.NewFunction(func(L *lua.LState) int {
		listing, err := ListSlot(ctx.Client, ctx.State.CurrentSlot, ctx.State.RootSlotID)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		result := L.NewTable()

		// Add children
		children := L.NewTable()
		for i, child := range listing.Children {
			childTable := L.NewTable()
			L.SetField(childTable, "id", lua.LString(child.ID))
			L.SetField(childTable, "name", lua.LString(child.Name))
			L.SetField(childTable, "active", lua.LBool(child.IsActive))
			L.SetField(childTable, "persistent", lua.LBool(child.IsPersistent))
			children.Append(childTable)
			_ = i
		}
		L.SetField(result, "children", children)

		// Add components
		components := L.NewTable()
		for i, comp := range listing.Components {
			compTable := L.NewTable()
			L.SetField(compTable, "id", lua.LString(comp.ID))
			L.SetField(compTable, "type", lua.LString(comp.Type))
			L.SetField(compTable, "persistent", lua.LBool(comp.IsPersistent))
			components.Append(compTable)
			_ = i
		}
		L.SetField(result, "components", components)

		L.Push(result)
		return 1
	}))

	// inspect function -> RESH.inspect - returns component/slot data
	L.SetField(resh, "inspect", L.NewFunction(func(L *lua.LState) int {
		targetID := L.CheckString(1)

		// Try component first
		compData, compErr := InspectComponent(ctx.Client, targetID)
		if compErr == nil {
			result := L.NewTable()
			L.SetField(result, "Type", lua.LString("component"))
			L.SetField(result, "ID", lua.LString(compData.ID))
			L.SetField(result, "ComponentType", lua.LString(compData.ComponentType))
			L.SetField(result, "TypeName", lua.LString(compData.TypeName))

			// Members as a map indexed by name for easy access
			members := L.NewTable()
			for _, member := range compData.Members {
				memberTable := L.NewTable()
				L.SetField(memberTable, "ID", lua.LString(member.ID))
				L.SetField(memberTable, "Name", lua.LString(member.Name))
				L.SetField(memberTable, "Type", lua.LString(member.Type))

				// Convert value to appropriate Lua type
				luaValue := convertToLuaValue(L, member.Value)
				L.SetField(memberTable, "Value", luaValue)

				// Handle references specially
				if member.Type == "reference" {
					if refMap, ok := member.Value.(map[string]interface{}); ok {
						if targetID, ok := refMap["targetId"].(string); ok {
							L.SetField(memberTable, "TargetId", lua.LString(targetID))
						} else {
							L.SetField(memberTable, "TargetId", lua.LNil)
						}
					}
				}

				// Index by member name for easy access
				L.SetField(members, member.Name, memberTable)
			}
			L.SetField(result, "Members", members)

			L.Push(result)
			return 1
		}

		// Try slot
		slotData, slotErr := InspectSlot(ctx.Client, targetID)
		if slotErr == nil {
			result := L.NewTable()
			L.SetField(result, "Type", lua.LString("slot"))
			L.SetField(result, "ID", lua.LString(slotData.ID))

			// Properties as a map for easy access
			for _, prop := range slotData.Properties {
				luaValue := convertToLuaValue(L, prop.Value)
				L.SetField(result, prop.Name, luaValue)
			}

			L.Push(result)
			return 1
		}

		L.Push(lua.LNil)
		L.Push(lua.LString(fmt.Sprintf("not found: %v, %v", compErr, slotErr)))
		return 2
	}))

	// pwd function -> RESH.pwd - returns current path
	L.SetField(resh, "pwd", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LString(ctx.State.CurrentPath))
		return 1
	}))

	// get_current_slot function -> RESH.get_current_slot - returns current slot ID
	L.SetField(resh, "get_current_slot", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LString(ctx.State.CurrentSlot))
		return 1
	}))

	// find_slot function -> RESH.find_slot - searches for slots by name
	L.SetField(resh, "find_slot", L.NewFunction(func(L *lua.LState) int {
		name := L.CheckString(1)

		listing, err := ListSlot(ctx.Client, ctx.State.CurrentSlot, ctx.State.RootSlotID)
		if err != nil {
			L.Push(lua.LNil)
			return 1
		}

		for _, child := range listing.Children {
			if child.Name == name {
				L.Push(lua.LString(child.ID))
				return 1
			}
		}

		L.Push(lua.LNil)
		return 1
	}))

	// update_component function -> RESH.update_component - updates component members
	L.SetField(resh, "update_component", L.NewFunction(func(L *lua.LState) int {
		componentID := L.CheckString(1)
		membersTable := L.CheckTable(2)

		// Convert Lua table to Go map
		members := make(map[string]interface{})
		membersTable.ForEach(func(key, value lua.LValue) {
			keyStr := key.String()
			members[keyStr] = convertFromLuaValue(value)
		})

		// Create component definition for update
		compDef := &resolink.ComponentDefinition{
			ID:      componentID,
			Members: members,
		}

		// Update the component
		err := ctx.Client.UpdateComponent(compDef)
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LBool(true))
		return 1
	}))

	// create_slot function -> RESH.create_slot - creates a new slot
	L.SetField(resh, "create_slot", L.NewFunction(func(L *lua.LState) int {
		name := L.CheckString(1)
		parentID := L.CheckString(2)

		// Generate a unique ID for the slot if not provided
		// We'll use a timestamp-based approach similar to how Resonite does it
		slotID := fmt.Sprintf("LuaSlot_%d", time.Now().UnixNano())

		// Create slot definition using helper functions
		slotDef := &resolink.SlotDefinition{
			ID:     slotID,
			Name:   resolink.NewValueString(name),
			Parent: resolink.NewValueReference(parentID),
		}

		// Create the slot
		resp, err := ctx.Client.AddSlot(slotDef)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		if resp == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("response is nil"))
			return 2
		}

		if !resp.Success {
			L.Push(lua.LNil)
			L.Push(lua.LString(resp.ErrorInfo))
			return 2
		}

		// If Data is present and has an ID, use that, otherwise use our provided ID
		if resp.Data != nil && resp.Data.ID != "" {
			L.Push(lua.LString(resp.Data.ID))
		} else {
			// Response doesn't include the slot data, so return the ID we provided
			L.Push(lua.LString(slotID))
		}
		return 1
	}))

	// delete_slot function -> RESH.delete_slot - deletes a slot
	L.SetField(resh, "delete_slot", L.NewFunction(func(L *lua.LState) int {
		slotID := L.CheckString(1)

		err := ctx.Client.RemoveSlot(slotID)
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LBool(true))
		return 1
	}))

	// create_component function -> RESH.create_component - creates a new component
	L.SetField(resh, "create_component", L.NewFunction(func(L *lua.LState) int {
		slotID := L.CheckString(1)
		componentType := L.CheckString(2)

		// Generate a truly unique ID using crypto/rand
		randomBytes := make([]byte, 8)
		rand.Read(randomBytes)
		compID := fmt.Sprintf("LuaComp_%d_%s", time.Now().UnixNano(), hex.EncodeToString(randomBytes)[:12])

		// Create component definition
		compDef := &resolink.ComponentDefinition{
			ID:            compID,
			ComponentType: componentType,
		}

		// Create the component
		resp, err := ctx.Client.AddComponent(slotID, compDef)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		if resp == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("response is nil"))
			return 2
		}

		if !resp.Success {
			L.Push(lua.LNil)
			L.Push(lua.LString(resp.ErrorInfo))
			return 2
		}

		// If Data is present and has an ID, use that, otherwise use our provided ID
		if resp.Data != nil && resp.Data.ID != "" {
			L.Push(lua.LString(resp.Data.ID))
		} else {
			// Response doesn't include the component data, so return the ID we provided
			L.Push(lua.LString(compID))
		}
		return 1
	}))

	// delete_component function -> RESH.delete_component - deletes a component
	L.SetField(resh, "delete_component", L.NewFunction(func(L *lua.LState) int {
		componentID := L.CheckString(1)

		err := ctx.Client.RemoveComponent(componentID)
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LBool(true))
		return 1
	}))

	// get_component_types function -> RESH.get_component_types - returns list of all component types
	L.SetField(resh, "get_component_types", L.NewFunction(func(L *lua.LState) int {
		types, err := ctx.Client.GetComponentTypeList()
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		result := L.NewTable()
		for _, t := range types {
			result.Append(lua.LString(t))
		}
		L.Push(result)
		return 1
	}))

	// get_component_definition function -> RESH.get_component_definition - returns component type definition
	L.SetField(resh, "get_component_definition", L.NewFunction(func(L *lua.LState) int {
		compType := L.CheckString(1)

		rawDef, err := ctx.Client.GetComponentDefinition(compType)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		// Parse the raw JSON into a map
		var defMap map[string]interface{}
		if jsonErr := json.Unmarshal(rawDef, &defMap); jsonErr != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(jsonErr.Error()))
			return 2
		}

		L.Push(convertToLuaValue(L, defMap))
		return 1
	}))

	// Set RESH table as global
	L.SetGlobal("RESH", resh)
}

// convertFromLuaValue converts a Lua value to a Go value
func convertFromLuaValue(lv lua.LValue) interface{} {
	switch v := lv.(type) {
	case *lua.LNilType:
		return nil
	case lua.LBool:
		return bool(v)
	case lua.LNumber:
		return float64(v)
	case lua.LString:
		return string(v)
	case *lua.LTable:
		// Check if it's an array or map
		isArray := true
		length := 0
		v.ForEach(func(key, val lua.LValue) {
			if num, ok := key.(lua.LNumber); ok {
				if int(num) != length+1 {
					isArray = false
				}
				length++
			} else {
				isArray = false
			}
		})

		if isArray {
			// Convert to array
			arr := make([]interface{}, 0, length)
			for i := 1; i <= length; i++ {
				val := v.RawGetInt(i)
				arr = append(arr, convertFromLuaValue(val))
			}
			return arr
		} else {
			// Convert to map
			m := make(map[string]interface{})
			v.ForEach(func(key, val lua.LValue) {
				m[key.String()] = convertFromLuaValue(val)
			})
			return m
		}
	default:
		return v.String()
	}
}

// convertToLuaValue converts a Go value to an appropriate Lua value
func convertToLuaValue(L *lua.LState, value interface{}) lua.LValue {
	if value == nil {
		return lua.LNil
	}

	switch v := value.(type) {
	case bool:
		return lua.LBool(v)
	case int:
		return lua.LNumber(v)
	case int64:
		return lua.LNumber(v)
	case float32:
		return lua.LNumber(v)
	case float64:
		return lua.LNumber(v)
	case string:
		return lua.LString(v)
	case map[string]interface{}:
		// Convert maps to Lua tables
		table := L.NewTable()
		for k, val := range v {
			L.SetField(table, k, convertToLuaValue(L, val))
		}
		return table
	case []interface{}:
		// Convert arrays to Lua tables
		table := L.NewTable()
		for _, val := range v {
			table.Append(convertToLuaValue(L, val))
		}
		return table
	default:
		// Fallback to string representation
		return lua.LString(fmt.Sprintf("%v", v))
	}
}
