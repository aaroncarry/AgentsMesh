package luaplugin

import (
	"encoding/json"
	"log"

	lua "github.com/yuin/gopher-lua"
)

// jsonEncodeAPI creates a Lua function that encodes a table to JSON.
// Usage: json_str = ctx.json_encode(table)
func jsonEncodeAPI() lua.LGFunction {
	return func(L *lua.LState) int {
		tbl := L.CheckTable(1)

		goMap := luaTableToGoMap(tbl)
		data, err := json.MarshalIndent(goMap, "", "  ")
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LString(string(data)))
		return 1
	}
}

// logAPI creates a Lua function that logs a message.
// Usage: ctx.log("message")
func logAPI() lua.LGFunction {
	return func(L *lua.LState) int {
		msg := L.CheckString(1)
		log.Printf("[luaplugin] %s", msg)
		return 0
	}
}

// readBuiltinResourceAPI creates a Lua function that reads embedded builtin resources.
// Usage: content = ctx.read_builtin_resource("skills/am-delegate.md")
// This allows plugins to load content from embedded resources without hardcoding in Lua.
func readBuiltinResourceAPI(resourceFS func(string) ([]byte, error)) lua.LGFunction {
	return func(L *lua.LState) int {
		path := L.CheckString(1)

		content, err := resourceFS(path)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LString(string(content)))
		return 1
	}
}
