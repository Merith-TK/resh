package shell

import (
	"fmt"

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