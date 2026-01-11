package luaplugin

import (
	"context"
	"testing"
)

func TestPluginExecutor_FilterPlugins(t *testing.T) {
	executor := NewPluginExecutor()

	plugins := []*LuaPlugin{
		{Name: "all-agents", SupportedAgents: []string{}},
		{Name: "claude-only", SupportedAgents: []string{"claude-code"}},
		{Name: "gemini-only", SupportedAgents: []string{"gemini-cli"}},
		{Name: "multi-agent", SupportedAgents: []string{"claude-code", "gemini-cli"}},
	}

	// Test filtering for claude-code
	claudePlugins := executor.filterPlugins(plugins, "claude-code")
	names := make(map[string]bool)
	for _, p := range claudePlugins {
		names[p.Name] = true
	}

	if !names["all-agents"] {
		t.Error("all-agents should be included for claude-code")
	}
	if !names["claude-only"] {
		t.Error("claude-only should be included for claude-code")
	}
	if names["gemini-only"] {
		t.Error("gemini-only should NOT be included for claude-code")
	}
	if !names["multi-agent"] {
		t.Error("multi-agent should be included for claude-code")
	}

	// Test filtering for unknown agent
	unknownPlugins := executor.filterPlugins(plugins, "unknown-agent")
	if len(unknownPlugins) != 1 || unknownPlugins[0].Name != "all-agents" {
		t.Error("Only all-agents should be included for unknown agent")
	}
}

func TestPluginExecutor_LoadSharedModules(t *testing.T) {
	executor := NewPluginExecutor()

	tmpDir := t.TempDir()
	sb := newMockSandbox("test-pod", tmpDir)
	sb.workDir = tmpDir

	// Create a plugin that uses mcp_utils (shared module)
	plugin := &LuaPlugin{
		Name:            "uses-mcp-utils",
		Version:         "1.0.0",
		SupportedAgents: []string{"test-agent"},
		content: []byte(`
function setup(ctx)
    -- Access mcp_utils which should be loaded as shared module
    if mcp_utils == nil then
        error("mcp_utils not loaded")
    end
    if mcp_utils.DEFAULT_PORT ~= 19000 then
        error("mcp_utils.DEFAULT_PORT incorrect")
    end
    ctx.add_env("MCP_UTILS_LOADED", "true")
end
`),
	}

	plugins := []*LuaPlugin{plugin}
	config := map[string]interface{}{}

	err := executor.Execute(context.Background(), plugins, "test-agent", sb, config)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if sb.GetEnvVars()["MCP_UTILS_LOADED"] != "true" {
		t.Error("Shared module mcp_utils was not loaded")
	}
}

func TestPluginContent(t *testing.T) {
	plugin := &LuaPlugin{
		Name:    "test",
		content: []byte("test content"),
	}

	content := plugin.Content()
	if string(content) != "test content" {
		t.Errorf("Content() = %q, want %q", string(content), "test content")
	}
}

func TestPluginIsBuiltin(t *testing.T) {
	builtinPlugin := &LuaPlugin{
		Name:      "builtin-test",
		isBuiltin: true,
	}

	userPlugin := &LuaPlugin{
		Name:      "user-test",
		isBuiltin: false,
	}

	if !builtinPlugin.IsBuiltin() {
		t.Error("Builtin plugin should return true for IsBuiltin()")
	}

	if userPlugin.IsBuiltin() {
		t.Error("User plugin should return false for IsBuiltin()")
	}
}

func TestPluginExecutor_Execute_PluginWithAllConfig(t *testing.T) {
	executor := NewPluginExecutor()

	tmpDir := t.TempDir()
	sb := newMockSandbox("test-pod", tmpDir)
	sb.workDir = tmpDir

	// Create a plugin that uses various config and context features
	plugin := &LuaPlugin{
		Name:            "config-test-plugin",
		Version:         "1.0.0",
		SupportedAgents: []string{"test-agent"},
		content: []byte(`
function setup(ctx)
    -- Test config access
    local port = ctx.config.port or 8080
    ctx.add_env("PORT", tostring(port))

    -- Test sandbox info access
    ctx.add_env("POD_KEY", ctx.sandbox.pod_key)
    ctx.add_env("ROOT_PATH", ctx.sandbox.root_path)
    ctx.add_env("WORK_DIR", ctx.sandbox.work_dir)

    -- Test file operations
    local test_file = ctx.sandbox.root_path .. "/test.txt"
    ctx.write_file(test_file, "hello")

    local exists = ctx.file_exists(test_file)
    if exists then
        ctx.add_env("FILE_EXISTS", "true")
    end

    -- Test metadata
    ctx.set_metadata("test_key", "test_value")
    local value = ctx.get_metadata("test_key")
    ctx.add_env("METADATA_VALUE", value)
end
`),
	}

	plugins := []*LuaPlugin{plugin}
	config := map[string]interface{}{
		"port": float64(9000),
	}

	err := executor.Execute(context.Background(), plugins, "test-agent", sb, config)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	envVars := sb.GetEnvVars()

	if envVars["PORT"] != "9000" {
		t.Errorf("Expected PORT=9000, got %s", envVars["PORT"])
	}
	if envVars["POD_KEY"] != "test-pod" {
		t.Errorf("Expected POD_KEY=test-pod, got %s", envVars["POD_KEY"])
	}
	if envVars["FILE_EXISTS"] != "true" {
		t.Errorf("Expected FILE_EXISTS=true, got %s", envVars["FILE_EXISTS"])
	}
	if envVars["METADATA_VALUE"] != "test_value" {
		t.Errorf("Expected METADATA_VALUE=test_value, got %s", envVars["METADATA_VALUE"])
	}
}
