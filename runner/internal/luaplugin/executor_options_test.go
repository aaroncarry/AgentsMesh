package luaplugin

import (
	"context"
	"testing"
)

// mockLoader implements PluginLoaderInterface for testing
type mockLoader struct {
	builtinPlugins []*LuaPlugin
	userPlugins    []*LuaPlugin
}

func (m *mockLoader) LoadBuiltinPlugins() ([]*LuaPlugin, error) {
	return m.builtinPlugins, nil
}

func (m *mockLoader) LoadUserPlugins(dir string, loadedNames map[string]bool) ([]*LuaPlugin, error) {
	return m.userPlugins, nil
}

// mockExecutor implements PluginExecutorInterface for testing
type mockExecutor struct {
	executeCount  int
	teardownCount int
}

func (m *mockExecutor) Execute(ctx context.Context, plugins []*LuaPlugin, agentType string, sb SandboxAdapter, config map[string]interface{}) error {
	m.executeCount++
	return nil
}

func (m *mockExecutor) Teardown(ctx context.Context, plugins []*LuaPlugin, agentType string, sb SandboxAdapter) error {
	m.teardownCount++
	return nil
}

func TestWithLoader(t *testing.T) {
	testPlugin := &LuaPlugin{
		Name:    "mock-builtin",
		Version: "1.0.0",
	}

	loader := &mockLoader{
		builtinPlugins: []*LuaPlugin{testPlugin},
	}

	manager := NewPluginManager("", WithLoader(loader))

	err := manager.LoadPlugins()
	if err != nil {
		t.Fatalf("LoadPlugins() error = %v", err)
	}

	plugins := manager.GetPlugins()
	if len(plugins) != 1 {
		t.Errorf("Expected 1 plugin, got %d", len(plugins))
	}
	if plugins[0].Name != "mock-builtin" {
		t.Errorf("Expected plugin name 'mock-builtin', got %q", plugins[0].Name)
	}
}

func TestWithExecutor(t *testing.T) {
	tmpDir := t.TempDir()
	sb := newMockSandbox("test-pod", tmpDir)

	executor := &mockExecutor{}

	manager := NewPluginManager("", WithExecutor(executor))

	// Load plugins first
	err := manager.LoadPlugins()
	if err != nil {
		t.Fatalf("LoadPlugins() error = %v", err)
	}

	// Execute should use mock executor
	err = manager.Execute(context.Background(), "test-agent", sb, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if executor.executeCount != 1 {
		t.Errorf("Expected execute count 1, got %d", executor.executeCount)
	}

	// Teardown should use mock executor
	err = manager.Teardown(context.Background(), "test-agent", sb)
	if err != nil {
		t.Fatalf("Teardown() error = %v", err)
	}

	if executor.teardownCount != 1 {
		t.Errorf("Expected teardown count 1, got %d", executor.teardownCount)
	}
}

func TestWithLoaderAndExecutor(t *testing.T) {
	testPlugin := &LuaPlugin{
		Name:    "combined-test",
		Version: "1.0.0",
	}

	loader := &mockLoader{
		builtinPlugins: []*LuaPlugin{testPlugin},
	}
	executor := &mockExecutor{}

	manager := NewPluginManager("",
		WithLoader(loader),
		WithExecutor(executor),
	)

	err := manager.LoadPlugins()
	if err != nil {
		t.Fatalf("LoadPlugins() error = %v", err)
	}

	tmpDir := t.TempDir()
	sb := newMockSandbox("test-pod", tmpDir)

	err = manager.Execute(context.Background(), "test-agent", sb, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify both custom loader and executor were used
	plugins := manager.GetPlugins()
	if len(plugins) != 1 || plugins[0].Name != "combined-test" {
		t.Error("Custom loader was not used correctly")
	}

	if executor.executeCount != 1 {
		t.Error("Custom executor was not used correctly")
	}
}
