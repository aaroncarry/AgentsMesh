package luaplugin

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewPluginLoader(t *testing.T) {
	loader := NewPluginLoader()
	if loader == nil {
		t.Fatal("NewPluginLoader() returned nil")
	}
	if loader.parser == nil {
		t.Fatal("Loader parser is nil")
	}
}

func TestPluginLoader_LoadBuiltinPlugins(t *testing.T) {
	loader := NewPluginLoader()

	plugins, err := loader.LoadBuiltinPlugins()
	if err != nil {
		t.Fatalf("LoadBuiltinPlugins() error = %v", err)
	}

	// Should have loaded the builtin plugins
	if len(plugins) == 0 {
		t.Error("Expected at least some builtin plugins")
	}

	// Verify expected builtin plugins
	names := make(map[string]bool)
	for _, p := range plugins {
		names[p.Name] = true
	}

	expectedPlugins := []string{"env", "claude-code", "gemini-cli", "codex-cli", "opencode"}
	for _, expected := range expectedPlugins {
		if !names[expected] {
			t.Errorf("Expected builtin plugin %q not found", expected)
		}
	}
}

func TestPluginLoader_LoadUserPlugins(t *testing.T) {
	loader := NewPluginLoader()

	// Create a temp directory with a valid user plugin
	userPluginsDir := t.TempDir()

	pluginContent := `
plugin = {
    name = "user-test-plugin",
    version = "1.0.0",
    description = "A user plugin for testing",
    order = 100,
    supported_agents = {"test-agent"},
}

function setup(ctx)
    ctx.log("User plugin setup")
end
`
	pluginPath := filepath.Join(userPluginsDir, "user_plugin.lua")
	if err := os.WriteFile(pluginPath, []byte(pluginContent), 0644); err != nil {
		t.Fatalf("Failed to write user plugin: %v", err)
	}

	loadedNames := make(map[string]bool)
	plugins, err := loader.LoadUserPlugins(userPluginsDir, loadedNames)
	if err != nil {
		t.Fatalf("LoadUserPlugins() error = %v", err)
	}

	if len(plugins) != 1 {
		t.Errorf("Expected 1 plugin, got %d", len(plugins))
	}

	if plugins[0].Name != "user-test-plugin" {
		t.Errorf("Expected plugin name 'user-test-plugin', got %q", plugins[0].Name)
	}
}

func TestPluginLoader_LoadUserPlugins_NonExistentDir(t *testing.T) {
	loader := NewPluginLoader()

	plugins, err := loader.LoadUserPlugins("/nonexistent/path", make(map[string]bool))
	if err != nil {
		t.Fatalf("LoadUserPlugins() should not return error for non-existent dir: %v", err)
	}

	if plugins != nil && len(plugins) > 0 {
		t.Error("Expected no plugins for non-existent directory")
	}
}

func TestPluginLoader_LoadUserPlugins_NotADirectory(t *testing.T) {
	loader := NewPluginLoader()

	// Create a file, not a directory
	tmpFile := filepath.Join(t.TempDir(), "not_a_dir")
	os.WriteFile(tmpFile, []byte("content"), 0644)

	_, err := loader.LoadUserPlugins(tmpFile, make(map[string]bool))
	if err == nil {
		t.Fatal("Expected error for non-directory path")
	}
}

func TestPluginLoader_LoadUserPlugins_ConflictWithBuiltin(t *testing.T) {
	loader := NewPluginLoader()

	// Create a plugin with a conflicting name
	userPluginsDir := t.TempDir()

	pluginContent := `
plugin = {
    name = "claude-code",  -- Conflicts with builtin
    version = "99.0.0",
}

function setup(ctx)
end
`
	pluginPath := filepath.Join(userPluginsDir, "conflict.lua")
	os.WriteFile(pluginPath, []byte(pluginContent), 0644)

	// Simulate builtin plugins already loaded
	loadedNames := map[string]bool{
		"claude-code": true,
	}

	plugins, err := loader.LoadUserPlugins(userPluginsDir, loadedNames)
	if err != nil {
		t.Fatalf("LoadUserPlugins() error = %v", err)
	}

	// The conflicting plugin should be skipped
	for _, p := range plugins {
		if p.Name == "claude-code" {
			t.Error("Conflicting plugin should have been skipped")
		}
	}
}

func TestPluginLoader_LoadUserPlugins_InvalidPlugin(t *testing.T) {
	loader := NewPluginLoader()

	userPluginsDir := t.TempDir()

	// Create an invalid plugin (no plugin table)
	invalidContent := `
something_else = {
    name = "not-a-plugin"
}
`
	pluginPath := filepath.Join(userPluginsDir, "invalid.lua")
	os.WriteFile(pluginPath, []byte(invalidContent), 0644)

	loadedNames := make(map[string]bool)
	plugins, err := loader.LoadUserPlugins(userPluginsDir, loadedNames)
	if err != nil {
		t.Fatalf("LoadUserPlugins() error = %v", err)
	}

	// Invalid plugins should be skipped, not cause error
	if len(plugins) != 0 {
		t.Errorf("Expected 0 plugins (invalid should be skipped), got %d", len(plugins))
	}
}

func TestPluginLoader_LoadUserPlugins_MixedValidAndInvalid(t *testing.T) {
	loader := NewPluginLoader()

	userPluginsDir := t.TempDir()

	// Create a valid plugin
	validContent := `
plugin = {
    name = "valid-plugin",
    version = "1.0.0",
}

function setup(ctx)
end
`
	os.WriteFile(filepath.Join(userPluginsDir, "valid.lua"), []byte(validContent), 0644)

	// Create an invalid plugin
	invalidContent := `this is not valid lua {{{`
	os.WriteFile(filepath.Join(userPluginsDir, "invalid.lua"), []byte(invalidContent), 0644)

	loadedNames := make(map[string]bool)
	plugins, err := loader.LoadUserPlugins(userPluginsDir, loadedNames)
	if err != nil {
		t.Fatalf("LoadUserPlugins() error = %v", err)
	}

	// Only valid plugin should be loaded
	if len(plugins) != 1 {
		t.Errorf("Expected 1 plugin, got %d", len(plugins))
	}
	if plugins[0].Name != "valid-plugin" {
		t.Errorf("Expected 'valid-plugin', got %q", plugins[0].Name)
	}
}

func TestPluginLoader_LoadUserPlugins_PluginWithoutName(t *testing.T) {
	loader := NewPluginLoader()

	userPluginsDir := t.TempDir()

	// Plugin without required name field
	noNameContent := `
plugin = {
    version = "1.0.0",
}
`
	os.WriteFile(filepath.Join(userPluginsDir, "noname.lua"), []byte(noNameContent), 0644)

	loadedNames := make(map[string]bool)
	plugins, err := loader.LoadUserPlugins(userPluginsDir, loadedNames)
	if err != nil {
		t.Fatalf("LoadUserPlugins() error = %v", err)
	}

	// Plugin without name should be skipped (validation fails)
	if len(plugins) != 0 {
		t.Errorf("Expected 0 plugins (no name should be skipped), got %d", len(plugins))
	}
}

func TestPluginLoader_LoadFromContent(t *testing.T) {
	loader := NewPluginLoader()

	// Test valid content
	t.Run("valid content", func(t *testing.T) {
		content := []byte(`
plugin = {
    name = "test",
    version = "1.0.0",
}
`)
		plugin, err := loader.loadFromContent("test.lua", content, true)
		if err != nil {
			t.Fatalf("loadFromContent() error = %v", err)
		}
		if plugin.Name != "test" {
			t.Errorf("Expected name 'test', got %q", plugin.Name)
		}
	})

	// Test invalid content
	t.Run("invalid content", func(t *testing.T) {
		content := []byte(`this is not valid lua`)
		_, err := loader.loadFromContent("invalid.lua", content, true)
		if err == nil {
			t.Fatal("Expected error for invalid content")
		}
	})

	// Test content failing validation
	t.Run("validation failure", func(t *testing.T) {
		content := []byte(`
plugin = {
    -- missing name
    version = "1.0.0",
}
`)
		_, err := loader.loadFromContent("noname.lua", content, true)
		if err == nil {
			t.Fatal("Expected validation error")
		}
	})
}

func TestPluginLoader_EmptyUserPluginsDir(t *testing.T) {
	loader := NewPluginLoader()

	// Create empty directory
	emptyDir := t.TempDir()

	loadedNames := make(map[string]bool)
	plugins, err := loader.LoadUserPlugins(emptyDir, loadedNames)
	if err != nil {
		t.Fatalf("LoadUserPlugins() error = %v", err)
	}

	if len(plugins) != 0 {
		t.Errorf("Expected 0 plugins for empty dir, got %d", len(plugins))
	}
}
