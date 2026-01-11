package luaplugin

import (
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

// PluginParser handles parsing Lua plugin metadata.
type PluginParser struct{}

// NewPluginParser creates a new PluginParser.
func NewPluginParser() *PluginParser {
	return &PluginParser{}
}

// Parse extracts plugin metadata from Lua script content.
func (p *PluginParser) Parse(filename string, content []byte, isBuiltin bool) (*LuaPlugin, error) {
	L := lua.NewState()
	defer L.Close()

	// Execute the Lua script to get the plugin table
	if err := L.DoString(string(content)); err != nil {
		return nil, fmt.Errorf("lua syntax error: %w", err)
	}

	// Get the plugin table
	pluginTable := L.GetGlobal("plugin")
	if pluginTable == lua.LNil {
		return nil, fmt.Errorf("plugin table not found in %s", filename)
	}

	tbl, ok := pluginTable.(*lua.LTable)
	if !ok {
		return nil, fmt.Errorf("plugin is not a table in %s", filename)
	}

	plugin := &LuaPlugin{
		Name:            getStringField(tbl, "name"),
		Version:         getStringField(tbl, "version"),
		Description:     getStringField(tbl, "description"),
		Order:           getIntField(tbl, "order"),
		Critical:        getBoolField(tbl, "critical"),
		SupportedAgents: getStringArrayField(tbl, "supported_agents"),
		Executable:      getStringField(tbl, "executable"),
		content:         content,
		isBuiltin:       isBuiltin,
		isAvailable:     true, // Will be updated by loader after checking executable
	}

	// Parse UI config if present
	uiTable := tbl.RawGetString("ui")
	if uiTable != lua.LNil {
		if uiTbl, ok := uiTable.(*lua.LTable); ok {
			plugin.UI = parseUIConfig(uiTbl)
		}
	}

	return plugin, nil
}

// Validate validates plugin required fields.
func (p *PluginParser) Validate(plugin *LuaPlugin, filename string) error {
	if plugin.Name == "" {
		return fmt.Errorf("plugin name is required in %s", filename)
	}
	return nil
}
