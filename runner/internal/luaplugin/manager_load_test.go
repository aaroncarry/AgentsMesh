package luaplugin

import (
	"testing"
)

func TestPluginManagerLoadPlugins(t *testing.T) {
	m := NewPluginManager("")
	if err := m.LoadPlugins(); err != nil {
		t.Fatalf("LoadPlugins() error = %v", err)
	}

	plugins := m.GetPlugins()
	if len(plugins) == 0 {
		t.Fatal("Expected at least one plugin to be loaded")
	}

	// Verify expected plugins are loaded
	expectedPlugins := map[string]bool{
		"env":        false,
		"claude-code": false,
		"gemini-cli":  false,
		"codex-cli":   false,
		"opencode":    false,
	}

	for _, p := range plugins {
		if _, ok := expectedPlugins[p.Name]; ok {
			expectedPlugins[p.Name] = true
		}
	}

	for name, found := range expectedPlugins {
		if !found {
			t.Errorf("Expected plugin %s not found", name)
		}
	}
}

func TestPluginManagerGetCapabilities(t *testing.T) {
	m := NewPluginManager("")
	if err := m.LoadPlugins(); err != nil {
		t.Fatalf("LoadPlugins() error = %v", err)
	}

	caps := m.GetCapabilities()
	if len(caps) == 0 {
		t.Fatal("Expected at least one capability")
	}

	// Verify claude-code plugin has UI config
	for _, cap := range caps {
		if cap.Name == "claude-code" {
			if cap.UI == nil {
				t.Error("claude-code plugin should have UI config")
			}
			if !cap.UI.Configurable {
				t.Error("claude-code plugin should be configurable")
			}
			if len(cap.UI.Fields) == 0 {
				t.Error("claude-code plugin should have UI fields")
			}
			return
		}
	}
	t.Error("claude-code plugin not found in capabilities")
}

func TestPluginManagerFilterPlugins(t *testing.T) {
	m := NewPluginManager("")
	if err := m.LoadPlugins(); err != nil {
		t.Fatalf("LoadPlugins() error = %v", err)
	}

	// Test filtering for claude-code using executor
	executor := NewPluginExecutor()
	claudePlugins := executor.filterPlugins(m.GetPlugins(), "claude-code")

	hasEnv := false
	hasClaudeCode := false
	hasGemini := false

	for _, p := range claudePlugins {
		switch p.Name {
		case "env":
			hasEnv = true
		case "claude-code":
			hasClaudeCode = true
		case "gemini-cli":
			hasGemini = true
		}
	}

	if !hasEnv {
		t.Error("env plugin should be included for claude-code (supports all agents)")
	}
	if !hasClaudeCode {
		t.Error("claude-code plugin should be included for claude-code")
	}
	if hasGemini {
		t.Error("gemini-cli plugin should NOT be included for claude-code")
	}
}

func TestPluginIsBuiltinMethod(t *testing.T) {
	m := NewPluginManager("")
	if err := m.LoadPlugins(); err != nil {
		t.Fatalf("LoadPlugins() error = %v", err)
	}

	// All loaded plugins should be builtin when no user plugins dir is provided
	for _, p := range m.GetPlugins() {
		if !p.IsBuiltin() {
			t.Errorf("Plugin %s should be marked as builtin", p.Name)
		}
	}
}

func TestPluginManagerLoadPlugins_WithLoaderError(t *testing.T) {
	// Test that LoadPlugins handles loader errors appropriately
	errorLoader := &mockLoader{
		builtinPlugins: []*LuaPlugin{
			{Name: "test-plugin", Version: "1.0.0"},
		},
	}

	manager := NewPluginManager("", WithLoader(errorLoader))
	err := manager.LoadPlugins()
	if err != nil {
		t.Fatalf("LoadPlugins() error = %v", err)
	}

	// Should have loaded the mock plugin
	plugins := manager.GetPlugins()
	if len(plugins) != 1 {
		t.Errorf("Expected 1 plugin, got %d", len(plugins))
	}
}

func TestPluginManagerLoadPlugins_EmptyLoader(t *testing.T) {
	loader := &errorMockLoader{}
	manager := NewPluginManager("", WithLoader(loader))

	err := manager.LoadPlugins()
	if err != nil {
		t.Fatalf("LoadPlugins() error = %v", err)
	}

	plugins := manager.GetPlugins()
	if len(plugins) != 0 {
		t.Errorf("Expected 0 plugins from empty loader, got %d", len(plugins))
	}
}
