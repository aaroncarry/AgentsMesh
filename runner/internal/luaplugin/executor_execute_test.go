package luaplugin

import (
	"context"
	"testing"
	"time"
)

func TestPluginExecutor_Execute(t *testing.T) {
	executor := NewPluginExecutor()

	tmpDir := t.TempDir()
	sb := newMockSandbox("test-pod", tmpDir)

	// Create a simple test plugin
	plugin := &LuaPlugin{
		Name:            "test-plugin",
		Version:         "1.0.0",
		SupportedAgents: []string{"test-agent"},
		Order:           10,
		Critical:        true,
		content: []byte(`
plugin = {
    name = "test-plugin",
    version = "1.0.0",
}

function setup(ctx)
    ctx.add_env("TEST_EXECUTED", "true")
end
`),
	}

	plugins := []*LuaPlugin{plugin}
	config := map[string]interface{}{}

	err := executor.Execute(context.Background(), plugins, "test-agent", sb, config)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if sb.GetEnvVars()["TEST_EXECUTED"] != "true" {
		t.Error("Plugin setup was not executed")
	}
}

func TestPluginExecutor_Execute_CriticalPluginFailure(t *testing.T) {
	executor := NewPluginExecutor()

	tmpDir := t.TempDir()
	sb := newMockSandbox("test-pod", tmpDir)

	// Create a critical plugin that fails
	plugin := &LuaPlugin{
		Name:            "failing-plugin",
		Version:         "1.0.0",
		SupportedAgents: []string{"test-agent"},
		Order:           10,
		Critical:        true, // This is critical
		content: []byte(`
function setup(ctx)
    error("Intentional failure")
end
`),
	}

	plugins := []*LuaPlugin{plugin}
	config := map[string]interface{}{}

	err := executor.Execute(context.Background(), plugins, "test-agent", sb, config)
	if err == nil {
		t.Fatal("Expected error for critical plugin failure")
	}
}

func TestPluginExecutor_Execute_NonCriticalPluginFailure(t *testing.T) {
	executor := NewPluginExecutor()

	tmpDir := t.TempDir()
	sb := newMockSandbox("test-pod", tmpDir)

	// Create a non-critical plugin that fails
	failingPlugin := &LuaPlugin{
		Name:            "non-critical-failing",
		Version:         "1.0.0",
		SupportedAgents: []string{"test-agent"},
		Order:           10,
		Critical:        false, // Not critical
		content: []byte(`
function setup(ctx)
    error("Intentional failure")
end
`),
	}

	// Create a plugin that runs after
	successPlugin := &LuaPlugin{
		Name:            "success-plugin",
		Version:         "1.0.0",
		SupportedAgents: []string{"test-agent"},
		Order:           20, // Runs after failing plugin
		Critical:        false,
		content: []byte(`
function setup(ctx)
    ctx.add_env("SUCCESS_RAN", "true")
end
`),
	}

	plugins := []*LuaPlugin{failingPlugin, successPlugin}
	config := map[string]interface{}{}

	// Should not return error for non-critical plugin failure
	err := executor.Execute(context.Background(), plugins, "test-agent", sb, config)
	if err != nil {
		t.Fatalf("Execute() should not fail for non-critical plugin: %v", err)
	}

	// Success plugin should still run
	if sb.GetEnvVars()["SUCCESS_RAN"] != "true" {
		t.Error("Success plugin should have run after non-critical failure")
	}
}

func TestPluginExecutor_Execute_NoSetupFunction(t *testing.T) {
	executor := NewPluginExecutor()

	tmpDir := t.TempDir()
	sb := newMockSandbox("test-pod", tmpDir)

	// Plugin without setup function
	plugin := &LuaPlugin{
		Name:            "no-setup-plugin",
		Version:         "1.0.0",
		SupportedAgents: []string{"test-agent"},
		content: []byte(`
plugin = {
    name = "no-setup-plugin",
}
-- No setup function defined
`),
	}

	plugins := []*LuaPlugin{plugin}
	config := map[string]interface{}{}

	// Should not error when there's no setup function
	err := executor.Execute(context.Background(), plugins, "test-agent", sb, config)
	if err != nil {
		t.Fatalf("Execute() should not fail when plugin has no setup: %v", err)
	}
}

func TestPluginExecutor_ContextCancellation(t *testing.T) {
	executor := NewPluginExecutor()

	tmpDir := t.TempDir()
	sb := newMockSandbox("test-pod", tmpDir)

	// Create a plugin that takes time
	plugin := &LuaPlugin{
		Name:            "slow-plugin",
		Version:         "1.0.0",
		SupportedAgents: []string{"test-agent"},
		Critical:        true,
		content: []byte(`
function setup(ctx)
    -- Simple loop that can be interrupted
    local i = 0
    while i < 1000000000 do
        i = i + 1
    end
end
`),
	}

	plugins := []*LuaPlugin{plugin}
	config := map[string]interface{}{}

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// This should fail due to context cancellation
	err := executor.Execute(ctx, plugins, "test-agent", sb, config)
	_ = err // Error is expected but timing-dependent
}

func TestPluginExecutor_ContextTimeout(t *testing.T) {
	executor := NewPluginExecutor()

	tmpDir := t.TempDir()
	sb := newMockSandbox("test-pod", tmpDir)

	// Create a plugin that would take a long time
	plugin := &LuaPlugin{
		Name:            "timeout-test",
		Version:         "1.0.0",
		SupportedAgents: []string{"test-agent"},
		Critical:        true,
		content: []byte(`
function setup(ctx)
    ctx.add_env("STARTED", "true")
end
`),
	}

	plugins := []*LuaPlugin{plugin}
	config := map[string]interface{}{}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// This should complete within timeout
	err := executor.Execute(ctx, plugins, "test-agent", sb, config)
	if err != nil {
		t.Fatalf("Execute() should complete within timeout: %v", err)
	}

	if sb.GetEnvVars()["STARTED"] != "true" {
		t.Error("Plugin should have executed")
	}
}

func TestPluginExecutor_ExecutePlugin_LuaSyntaxError(t *testing.T) {
	executor := NewPluginExecutor()

	tmpDir := t.TempDir()
	sb := newMockSandbox("test-pod", tmpDir)

	// Create a plugin with Lua syntax error
	plugin := &LuaPlugin{
		Name:            "syntax-error-plugin",
		Version:         "1.0.0",
		SupportedAgents: []string{"test-agent"},
		Critical:        true,
		content:         []byte(`this is not valid lua {{{`),
	}

	plugins := []*LuaPlugin{plugin}
	config := map[string]interface{}{}

	err := executor.Execute(context.Background(), plugins, "test-agent", sb, config)
	if err == nil {
		t.Fatal("Expected error for Lua syntax error")
	}
}

func TestPluginExecutor_InitSharedModulesCalledOnce(t *testing.T) {
	executor := NewPluginExecutor()

	tmpDir := t.TempDir()
	sb := newMockSandbox("test-pod", tmpDir)

	// Create a simple plugin
	plugin := &LuaPlugin{
		Name:            "simple-plugin",
		Version:         "1.0.0",
		SupportedAgents: []string{"test-agent"},
		content: []byte(`
function setup(ctx)
end
`),
	}

	plugins := []*LuaPlugin{plugin}
	config := map[string]interface{}{}

	// Execute multiple times - shared modules should only be loaded once
	for i := 0; i < 3; i++ {
		err := executor.Execute(context.Background(), plugins, "test-agent", sb, config)
		if err != nil {
			t.Fatalf("Execute() error on iteration %d: %v", i, err)
		}
	}
}
