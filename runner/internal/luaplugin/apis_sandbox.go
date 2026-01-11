package luaplugin

import (
	lua "github.com/yuin/gopher-lua"
)

// addArgsAPI creates a Lua function that adds launch arguments.
// Usage: ctx.add_args("--flag", "value") or ctx.add_args("--flag")
// Thread-safe: uses atomic AppendLaunchArgs method.
func addArgsAPI(sb SandboxAdapter) lua.LGFunction {
	return func(L *lua.LState) int {
		n := L.GetTop()
		args := make([]string, 0, n)
		for i := 1; i <= n; i++ {
			arg := L.CheckString(i)
			args = append(args, arg)
		}
		// Use atomic append instead of get/modify/set pattern
		sb.AppendLaunchArgs(args...)
		return 0
	}
}

// addEnvAPI creates a Lua function that adds an environment variable.
// Usage: ctx.add_env("KEY", "value")
func addEnvAPI(sb SandboxAdapter) lua.LGFunction {
	return func(L *lua.LState) int {
		key := L.CheckString(1)
		value := L.CheckString(2)
		sb.SetEnvVar(key, value)
		return 0
	}
}

// setMetadataAPI creates a Lua function that sets metadata.
// Usage: ctx.set_metadata("key", value)
func setMetadataAPI(sb SandboxAdapter) lua.LGFunction {
	return func(L *lua.LState) int {
		key := L.CheckString(1)
		value := L.CheckAny(2)
		sb.SetMetadata(key, luaValueToGo(value))
		return 0
	}
}

// getMetadataAPI creates a Lua function that gets metadata.
// Usage: value = ctx.get_metadata("key")
func getMetadataAPI(sb SandboxAdapter) lua.LGFunction {
	return func(L *lua.LState) int {
		key := L.CheckString(1)

		metadata := sb.GetMetadata()
		if metadata == nil {
			L.Push(lua.LNil)
			return 1
		}

		value, ok := metadata[key]
		if !ok {
			L.Push(lua.LNil)
			return 1
		}

		L.Push(goValueToLua(L, value))
		return 1
	}
}

// appendPromptAPI creates a Lua function that appends text to the initial prompt.
// Usage: ctx.append_prompt("\n\nultrathink")
func appendPromptAPI(sb SandboxAdapter) lua.LGFunction {
	return func(L *lua.LState) int {
		text := L.CheckString(1)

		// Store in metadata for later use by the runner
		metadata := sb.GetMetadata()
		existing, _ := metadata["prompt_suffix"].(string)
		sb.SetMetadata("prompt_suffix", existing+text)
		return 0
	}
}
