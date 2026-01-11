// Package luaplugin provides Lua plugin API implementations.
//
// API functions are organized into files by category:
//   - apis_file.go: File system operations (write_file, read_file, read_json, mkdir, file_exists)
//   - apis_sandbox.go: Sandbox operations (add_args, add_env, set_metadata, get_metadata, append_prompt)
//   - apis_util.go: Utility operations (json_encode, log, read_builtin_resource)
//
// Note: API registration is done in createContextTable() via ctx methods.
// The ctx table provides all plugin APIs (write_file, add_args, etc.)
// There's no need for global API registration since plugins access APIs through ctx parameter.
package luaplugin
