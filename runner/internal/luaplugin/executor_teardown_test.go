package luaplugin

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestPluginExecutor_Teardown(t *testing.T) {
	executor := NewPluginExecutor()

	tmpDir := t.TempDir()
	sb := newMockSandbox("test-pod", tmpDir)

	// Create a plugin with teardown function
	plugin := &LuaPlugin{
		Name:            "teardown-plugin",
		Version:         "1.0.0",
		SupportedAgents: []string{"test-agent"},
		Order:           10,
		content: []byte(`
function setup(ctx)
    ctx.add_env("SETUP_RAN", "true")
end

function teardown(ctx)
    ctx.add_env("TEARDOWN_RAN", "true")
end
`),
	}

	plugins := []*LuaPlugin{plugin}

	// First run setup
	err := executor.Execute(context.Background(), plugins, "test-agent", sb, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Then run teardown
	err = executor.Teardown(context.Background(), plugins, "test-agent", sb)
	if err != nil {
		t.Fatalf("Teardown() error = %v", err)
	}

	if sb.GetEnvVars()["TEARDOWN_RAN"] != "true" {
		t.Error("Teardown function was not executed")
	}
}

func TestPluginExecutor_Teardown_NoTeardownFunction(t *testing.T) {
	executor := NewPluginExecutor()

	tmpDir := t.TempDir()
	sb := newMockSandbox("test-pod", tmpDir)

	// Plugin without teardown function
	plugin := &LuaPlugin{
		Name:            "no-teardown-plugin",
		Version:         "1.0.0",
		SupportedAgents: []string{"test-agent"},
		content: []byte(`
function setup(ctx)
    ctx.add_env("SETUP_RAN", "true")
end
-- No teardown function
`),
	}

	plugins := []*LuaPlugin{plugin}

	// Should not error when there's no teardown function
	err := executor.Teardown(context.Background(), plugins, "test-agent", sb)
	if err != nil {
		t.Fatalf("Teardown() should not fail when plugin has no teardown: %v", err)
	}
}

func TestPluginExecutor_Teardown_ReverseOrder(t *testing.T) {
	executor := NewPluginExecutor()

	tmpDir := t.TempDir()
	sb := newMockSandbox("test-pod", tmpDir)

	// Create plugins with different orders
	plugin1 := &LuaPlugin{
		Name:            "first-plugin",
		Version:         "1.0.0",
		SupportedAgents: []string{"test-agent"},
		Order:           10,
		content: []byte(`
function setup(ctx)
end

function teardown(ctx)
    local current = ctx.get_metadata("teardown_order") or ""
    ctx.set_metadata("teardown_order", current .. "first,")
end
`),
	}

	plugin2 := &LuaPlugin{
		Name:            "second-plugin",
		Version:         "1.0.0",
		SupportedAgents: []string{"test-agent"},
		Order:           20,
		content: []byte(`
function setup(ctx)
end

function teardown(ctx)
    local current = ctx.get_metadata("teardown_order") or ""
    ctx.set_metadata("teardown_order", current .. "second,")
end
`),
	}

	plugins := []*LuaPlugin{plugin1, plugin2}

	// Run teardown
	err := executor.Teardown(context.Background(), plugins, "test-agent", sb)
	if err != nil {
		t.Fatalf("Teardown() error = %v", err)
	}

	// Verify teardown ran in reverse order (second, then first)
	order := sb.GetMetadata()["teardown_order"]
	if order != "second,first," {
		t.Errorf("Teardown order incorrect: got %q, want %q", order, "second,first,")
	}
}

func TestPluginExecutor_Teardown_ContinuesOnError(t *testing.T) {
	executor := NewPluginExecutor()

	tmpDir := t.TempDir()
	sb := newMockSandbox("test-pod", tmpDir)

	// First plugin (higher order, runs first in teardown) fails
	failingPlugin := &LuaPlugin{
		Name:            "failing-teardown",
		Version:         "1.0.0",
		SupportedAgents: []string{"test-agent"},
		Order:           20,
		content: []byte(`
function teardown(ctx)
    error("Teardown failure")
end
`),
	}

	// Second plugin (lower order, runs second in teardown) should still run
	successPlugin := &LuaPlugin{
		Name:            "success-teardown",
		Version:         "1.0.0",
		SupportedAgents: []string{"test-agent"},
		Order:           10,
		content: []byte(`
function teardown(ctx)
    ctx.add_env("TEARDOWN_SUCCESS", "true")
end
`),
	}

	plugins := []*LuaPlugin{failingPlugin, successPlugin}

	// Teardown should not return error even when one plugin fails
	err := executor.Teardown(context.Background(), plugins, "test-agent", sb)
	if err != nil {
		t.Fatalf("Teardown() should not return error: %v", err)
	}

	// Success plugin should still have run
	if sb.GetEnvVars()["TEARDOWN_SUCCESS"] != "true" {
		t.Error("Success teardown should have run despite earlier failure")
	}
}

func TestPluginExecutor_TeardownPlugin_LuaSyntaxError(t *testing.T) {
	executor := NewPluginExecutor()

	tmpDir := t.TempDir()
	sb := newMockSandbox("test-pod", tmpDir)

	// Create a plugin with Lua syntax error
	plugin := &LuaPlugin{
		Name:            "syntax-error-plugin",
		Version:         "1.0.0",
		SupportedAgents: []string{"test-agent"},
		content:         []byte(`this is not valid lua {{{`),
	}

	plugins := []*LuaPlugin{plugin}

	// Teardown should log the error but not return it
	err := executor.Teardown(context.Background(), plugins, "test-agent", sb)
	if err != nil {
		t.Fatalf("Teardown() should not return error: %v", err)
	}
}

func TestPluginExecutor_TeardownPlugin_TeardownFunctionError(t *testing.T) {
	executor := NewPluginExecutor()

	tmpDir := t.TempDir()
	sb := newMockSandbox("test-pod", tmpDir)

	// Create a plugin where teardown function throws an error
	plugin := &LuaPlugin{
		Name:            "teardown-error-plugin",
		Version:         "1.0.0",
		SupportedAgents: []string{"test-agent"},
		content: []byte(`
function setup(ctx)
end

function teardown(ctx)
    error("Teardown intentional error")
end
`),
	}

	plugins := []*LuaPlugin{plugin}

	// Teardown should log the error but not return it
	err := executor.Teardown(context.Background(), plugins, "test-agent", sb)
	if err != nil {
		t.Fatalf("Teardown() should not return error: %v", err)
	}
}

func TestPluginManagerTeardown(t *testing.T) {
	// Create a real test to ensure manager.Teardown works
	userPluginsDir := t.TempDir()

	// Create a user plugin with teardown
	pluginContent := `
plugin = {
    name = "teardown-test",
    version = "1.0.0",
    supported_agents = {"test-agent"},
    order = 10,
}

function setup(ctx)
    ctx.add_env("SETUP_DONE", "yes")
end

function teardown(ctx)
    ctx.add_env("TEARDOWN_DONE", "yes")
end
`
	pluginPath := filepath.Join(userPluginsDir, "teardown_test.lua")
	os.WriteFile(pluginPath, []byte(pluginContent), 0644)

	manager := NewPluginManager(userPluginsDir)
	err := manager.LoadPlugins()
	if err != nil {
		t.Fatalf("LoadPlugins() error = %v", err)
	}

	tmpDir := t.TempDir()
	sb := newMockSandbox("test-pod", tmpDir)

	// First execute
	err = manager.Execute(context.Background(), "test-agent", sb, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if sb.GetEnvVars()["SETUP_DONE"] != "yes" {
		t.Error("Setup was not executed")
	}

	// Then teardown
	err = manager.Teardown(context.Background(), "test-agent", sb)
	if err != nil {
		t.Fatalf("Teardown() error = %v", err)
	}

	if sb.GetEnvVars()["TEARDOWN_DONE"] != "yes" {
		t.Error("Teardown was not executed")
	}
}
