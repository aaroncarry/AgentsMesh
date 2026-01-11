package luaplugin

import (
	"github.com/anthropics/agentmesh/runner/internal/luaplugin/builtin"
	lua "github.com/yuin/gopher-lua"
)

// createContextTable creates the ctx table passed to Lua setup/teardown functions.
// ctx contains:
//   - config: plugin configuration from CreatePodCommand
//   - sandbox: sandbox information (pod_key, root_path, work_dir)
//   - write_file(path, content): write file
//   - read_file(path): read file content
//   - read_json(path): read and parse JSON file
//   - mkdir(path): create directory
//   - file_exists(path): check if file exists
//   - add_args(...): add launch arguments
//   - add_env(key, value): add environment variable
//   - set_metadata(key, value): set metadata
//   - get_metadata(key): get metadata
//   - json_encode(table): encode table to JSON string
//   - append_prompt(text): append text to initial prompt
//   - log(message): log a message
//   - read_builtin_resource(path): read embedded builtin resource file
func createContextTable(L *lua.LState, sb SandboxAdapter, config map[string]interface{}) *lua.LTable {
	ctx := L.NewTable()

	// Set config table
	configTable := L.NewTable()
	if config != nil {
		for k, v := range config {
			configTable.RawSetString(k, goValueToLua(L, v))
		}
	}
	ctx.RawSetString("config", configTable)

	// Set sandbox table
	sandboxTable := L.NewTable()
	sandboxTable.RawSetString("pod_key", lua.LString(sb.GetPodKey()))
	sandboxTable.RawSetString("root_path", lua.LString(sb.GetRootPath()))
	sandboxTable.RawSetString("work_dir", lua.LString(sb.GetWorkDir()))
	ctx.RawSetString("sandbox", sandboxTable)

	// Set API functions
	ctx.RawSetString("write_file", L.NewFunction(writeFileAPI(sb)))
	ctx.RawSetString("read_file", L.NewFunction(readFileAPI(sb)))
	ctx.RawSetString("read_json", L.NewFunction(readJSONAPI(sb)))
	ctx.RawSetString("mkdir", L.NewFunction(mkdirAPI(sb)))
	ctx.RawSetString("file_exists", L.NewFunction(fileExistsAPI(sb)))
	ctx.RawSetString("add_args", L.NewFunction(addArgsAPI(sb)))
	ctx.RawSetString("add_env", L.NewFunction(addEnvAPI(sb)))
	ctx.RawSetString("set_metadata", L.NewFunction(setMetadataAPI(sb)))
	ctx.RawSetString("get_metadata", L.NewFunction(getMetadataAPI(sb)))
	ctx.RawSetString("json_encode", L.NewFunction(jsonEncodeAPI()))
	ctx.RawSetString("append_prompt", L.NewFunction(appendPromptAPI(sb)))
	ctx.RawSetString("log", L.NewFunction(logAPI()))
	ctx.RawSetString("read_builtin_resource", L.NewFunction(readBuiltinResourceAPI(builtin.BuiltinSkills.ReadFile)))

	return ctx
}
