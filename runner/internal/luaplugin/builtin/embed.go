// Package builtin contains embedded Lua plugins for built-in agent support.
package builtin

import "embed"

//go:embed *.lua
var BuiltinPlugins embed.FS

//go:embed skills/*.md
var BuiltinSkills embed.FS
