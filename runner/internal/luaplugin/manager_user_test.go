package luaplugin

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestPluginManagerLoadUserPlugins(t *testing.T) {
	// Create a temporary user plugins directory
	userPluginsDir := t.TempDir()

	// Create a test user plugin
	testPluginContent := `
plugin = {
    name = "test-user-plugin",
    version = "1.0.0",
    description = "A test user plugin",
    supported_agents = {"test-agent"},
    order = 100,
    critical = false,
    ui = {
        configurable = true,
        fields = {
            { name = "test_option", type = "boolean", label = "Test Option", default = true },
        },
    },
}

function setup(ctx)
    ctx.add_env("TEST_USER_PLUGIN", "loaded")
    ctx.log("Test user plugin executed")
end

function teardown(ctx)
end
`
	pluginPath := filepath.Join(userPluginsDir, "test_plugin.lua")
	if err := os.WriteFile(pluginPath, []byte(testPluginContent), 0644); err != nil {
		t.Fatalf("Failed to write test plugin: %v", err)
	}

	// Create plugin manager with user plugins directory
	m := NewPluginManager(userPluginsDir)
	if err := m.LoadPlugins(); err != nil {
		t.Fatalf("LoadPlugins() error = %v", err)
	}

	plugins := m.GetPlugins()

	// Verify user plugin is loaded
	var userPlugin *LuaPlugin
	for _, p := range plugins {
		if p.Name == "test-user-plugin" {
			userPlugin = p
			break
		}
	}

	if userPlugin == nil {
		t.Fatal("User plugin 'test-user-plugin' was not loaded")
	}

	// Verify it's marked as not builtin
	if userPlugin.IsBuiltin() {
		t.Error("User plugin should not be marked as builtin")
	}

	// Verify UI config is parsed
	if userPlugin.UI == nil {
		t.Error("User plugin UI config is nil")
	} else if !userPlugin.UI.Configurable {
		t.Error("User plugin should be configurable")
	}

	// Verify the plugin executes correctly
	tmpDir := t.TempDir()
	sb := newMockSandbox("test-pod", tmpDir)

	config := map[string]interface{}{}
	err := m.Execute(context.Background(), "test-agent", sb, config)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify environment variable was set
	if sb.GetEnvVars()["TEST_USER_PLUGIN"] != "loaded" {
		t.Error("User plugin did not set expected environment variable")
	}
}

func TestPluginManagerUserPluginDoesNotOverrideBuiltin(t *testing.T) {
	// Create a temporary user plugins directory
	userPluginsDir := t.TempDir()

	// Create a user plugin with the same name as a builtin plugin
	conflictPluginContent := `
plugin = {
    name = "claude-code",  -- Same name as builtin
    version = "99.0.0",
    description = "Conflicting plugin",
    supported_agents = {"claude-code"},
    order = 100,
    critical = false,
}

function setup(ctx)
    ctx.add_env("CONFLICT_PLUGIN", "loaded")
end
`
	pluginPath := filepath.Join(userPluginsDir, "conflict_plugin.lua")
	if err := os.WriteFile(pluginPath, []byte(conflictPluginContent), 0644); err != nil {
		t.Fatalf("Failed to write conflict plugin: %v", err)
	}

	// Create plugin manager with user plugins directory
	m := NewPluginManager(userPluginsDir)
	if err := m.LoadPlugins(); err != nil {
		t.Fatalf("LoadPlugins() error = %v", err)
	}

	// Count claude-code plugins - should only be 1 (the builtin)
	claudeCodeCount := 0
	for _, p := range m.GetPlugins() {
		if p.Name == "claude-code" {
			claudeCodeCount++
			// Verify it's the builtin, not the user plugin
			if p.Version == "99.0.0" {
				t.Error("User plugin should not override builtin plugin")
			}
		}
	}

	if claudeCodeCount != 1 {
		t.Errorf("Expected exactly 1 claude-code plugin, got %d", claudeCodeCount)
	}
}

func TestPluginManagerEmptyUserPluginsDir(t *testing.T) {
	// Test with non-existent directory
	m := NewPluginManager("/non/existent/path")
	if err := m.LoadPlugins(); err != nil {
		t.Fatalf("LoadPlugins() should not fail for non-existent user plugins dir: %v", err)
	}

	// Should still have builtin plugins
	plugins := m.GetPlugins()
	if len(plugins) == 0 {
		t.Error("Should have loaded builtin plugins even with non-existent user plugins dir")
	}
}
