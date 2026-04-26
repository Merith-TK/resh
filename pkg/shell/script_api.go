package shell

import (
	"fmt"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

// registerShellFunctions registers shell operations as Lua functions under RESH table
func registerShellFunctions(L *lua.LState, ctx *ScriptContext) {
	// Create RESH table
	resh := L.NewTable()

	// Register different categories of functions
	registerNavigationFunctions(L, resh, ctx)
	registerInspectionFunctions(L, resh, ctx)
	registerModificationFunctions(L, resh, ctx)
	registerReflectionFunctions(L, resh, ctx)

	// Set RESH table as global
	L.SetGlobal("RESH", resh)
}

// registerNavigationFunctions registers navigation-related functions
func registerNavigationFunctions(L *lua.LState, resh *lua.LTable, ctx *ScriptContext) {
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
}

// registerInspectionFunctions registers inspection-related functions
func registerInspectionFunctions(L *lua.LState, resh *lua.LTable, ctx *ScriptContext) {
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
}