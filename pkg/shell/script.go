package shell

import (
	"fmt"
	"strings"

	"github.com/Merith-TK/resonite-sh/pkg/resolink"
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

// registerShellFunctions registers shell operations as Lua functions
func registerShellFunctions(L *lua.LState, ctx *ScriptContext) {
	// print function
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

	// cd function
	L.SetGlobal("cd", L.NewFunction(func(L *lua.LState) int {
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

	// ls function - returns table of slot/component info
	L.SetGlobal("ls", L.NewFunction(func(L *lua.LState) int {
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

	// inspect function - returns component/slot data
	L.SetGlobal("inspect", L.NewFunction(func(L *lua.LState) int {
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

	// pwd function - returns current path
	L.SetGlobal("pwd", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LString(ctx.State.CurrentPath))
		return 1
	}))

	// get_current_slot function - returns current slot ID
	L.SetGlobal("get_current_slot", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LString(ctx.State.CurrentSlot))
		return 1
	}))

	// find_slot function - searches for slots by name
	L.SetGlobal("find_slot", L.NewFunction(func(L *lua.LState) int {
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
