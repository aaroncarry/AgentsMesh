package luaplugin

import (
	"testing"
)

func TestPluginParser_Parse(t *testing.T) {
	parser := NewPluginParser()

	// Test parsing valid plugin
	t.Run("valid plugin", func(t *testing.T) {
		content := []byte(`
plugin = {
    name = "test-plugin",
    version = "1.0.0",
    description = "A test plugin",
    order = 50,
    critical = true,
    supported_agents = {"agent1", "agent2"},
    ui = {
        configurable = true,
        fields = {
            { name = "enabled", type = "boolean", label = "Enable", default = true },
        },
    },
}

function setup(ctx)
end
`)

		plugin, err := parser.Parse("test.lua", content, true)
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}

		if plugin.Name != "test-plugin" {
			t.Errorf("Expected name 'test-plugin', got %q", plugin.Name)
		}
		if plugin.Version != "1.0.0" {
			t.Errorf("Expected version '1.0.0', got %q", plugin.Version)
		}
		if plugin.Description != "A test plugin" {
			t.Errorf("Expected description 'A test plugin', got %q", plugin.Description)
		}
		if plugin.Order != 50 {
			t.Errorf("Expected order 50, got %d", plugin.Order)
		}
		if !plugin.Critical {
			t.Error("Expected Critical to be true")
		}
		if len(plugin.SupportedAgents) != 2 {
			t.Errorf("Expected 2 supported agents, got %d", len(plugin.SupportedAgents))
		}
		if !plugin.IsBuiltin() {
			t.Error("Expected plugin to be builtin")
		}
		if plugin.UI == nil {
			t.Error("Expected UI config to be present")
		}
	})

	// Test parsing plugin with syntax error
	t.Run("syntax error", func(t *testing.T) {
		content := []byte(`
plugin = {
    name = "broken"  -- missing comma
    version = "1.0.0"
}
`)
		_, err := parser.Parse("broken.lua", content, false)
		if err == nil {
			t.Fatal("Expected syntax error")
		}
	})

	// Test parsing without plugin table
	t.Run("no plugin table", func(t *testing.T) {
		content := []byte(`
something_else = {
    name = "not-a-plugin"
}
`)
		_, err := parser.Parse("noplugin.lua", content, false)
		if err == nil {
			t.Fatal("Expected error for missing plugin table")
		}
	})

	// Test parsing with plugin not being a table
	t.Run("plugin not a table", func(t *testing.T) {
		content := []byte(`plugin = "not a table"`)
		_, err := parser.Parse("notatable.lua", content, false)
		if err == nil {
			t.Fatal("Expected error for plugin not being a table")
		}
	})

	// Test parsing user plugin (not builtin)
	t.Run("user plugin", func(t *testing.T) {
		content := []byte(`
plugin = {
    name = "user-plugin",
    version = "1.0.0",
}
`)
		plugin, err := parser.Parse("user.lua", content, false)
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}
		if plugin.IsBuiltin() {
			t.Error("User plugin should not be marked as builtin")
		}
	})

	// Test plugin without UI config
	t.Run("no UI config", func(t *testing.T) {
		content := []byte(`
plugin = {
    name = "no-ui",
    version = "1.0.0",
}
`)
		plugin, err := parser.Parse("noui.lua", content, false)
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}
		if plugin.UI != nil {
			t.Error("Expected UI to be nil")
		}
	})
}

func TestPluginParser_Validate(t *testing.T) {
	parser := NewPluginParser()

	// Test valid plugin
	t.Run("valid plugin", func(t *testing.T) {
		plugin := &LuaPlugin{
			Name:    "valid-plugin",
			Version: "1.0.0",
		}
		err := parser.Validate(plugin, "valid.lua")
		if err != nil {
			t.Fatalf("Validate() error = %v", err)
		}
	})

	// Test plugin without name
	t.Run("missing name", func(t *testing.T) {
		plugin := &LuaPlugin{
			Name:    "",
			Version: "1.0.0",
		}
		err := parser.Validate(plugin, "noname.lua")
		if err == nil {
			t.Fatal("Expected validation error for missing name")
		}
	})
}

func TestNewPluginParser(t *testing.T) {
	parser := NewPluginParser()
	if parser == nil {
		t.Fatal("NewPluginParser() returned nil")
	}
}
