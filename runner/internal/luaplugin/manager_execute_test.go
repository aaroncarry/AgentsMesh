package luaplugin

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestPluginManagerExecuteClaudeCode(t *testing.T) {
	m := NewPluginManager("")
	if err := m.LoadPlugins(); err != nil {
		t.Fatalf("LoadPlugins() error = %v", err)
	}

	tmpDir := t.TempDir()
	sb := newMockSandbox("test-pod", tmpDir)
	sb.workDir = tmpDir

	config := map[string]interface{}{
		"mcp_port":         19000,
		"mcp_enabled":      true,
		"skills_enabled":   true,
		"model":            "opus",
		"permission_mode":  "plan",
		"skip_permissions": false,
	}

	err := m.Execute(context.Background(), "claude-code", sb, config)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify MCP config was created
	mcpConfigPath := filepath.Join(tmpDir, "mcp-config.json")
	if _, err := os.Stat(mcpConfigPath); os.IsNotExist(err) {
		t.Error("MCP config file was not created")
	}

	// Verify launch args contain expected flags
	args := sb.GetLaunchArgs()
	hasModel := false
	hasPermission := false
	hasMCPConfig := false

	for i, arg := range args {
		if arg == "--model" && i+1 < len(args) && args[i+1] == "opus" {
			hasModel = true
		}
		if arg == "--permission-mode" && i+1 < len(args) && args[i+1] == "plan" {
			hasPermission = true
		}
		if arg == "--mcp-config" {
			hasMCPConfig = true
		}
	}

	if !hasModel {
		t.Error("Expected --model opus in launch args")
	}
	if !hasPermission {
		t.Error("Expected --permission-mode plan in launch args")
	}
	if !hasMCPConfig {
		t.Error("Expected --mcp-config in launch args")
	}

	// Verify skills directory was created
	skillsDir := filepath.Join(tmpDir, ".claude", "skills")
	if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
		t.Error("Skills directory was not created")
	}
}

func TestPluginManagerExecuteEnvPlugin(t *testing.T) {
	m := NewPluginManager("")
	if err := m.LoadPlugins(); err != nil {
		t.Fatalf("LoadPlugins() error = %v", err)
	}

	tmpDir := t.TempDir()
	sb := newMockSandbox("test-pod", tmpDir)

	config := map[string]interface{}{
		"env_vars": map[string]interface{}{
			"API_KEY":     "test-key",
			"ANOTHER_VAR": "another-value",
		},
	}

	// Execute for any agent type since env plugin supports all
	err := m.Execute(context.Background(), "claude-code", sb, config)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify environment variables were set
	envVars := sb.GetEnvVars()
	if envVars["API_KEY"] != "test-key" {
		t.Errorf("Expected API_KEY=test-key, got %s", envVars["API_KEY"])
	}
	if envVars["ANOTHER_VAR"] != "another-value" {
		t.Errorf("Expected ANOTHER_VAR=another-value, got %s", envVars["ANOTHER_VAR"])
	}
}

func TestPluginManagerExecuteGeminiCLI(t *testing.T) {
	m := NewPluginManager("")
	if err := m.LoadPlugins(); err != nil {
		t.Fatalf("LoadPlugins() error = %v", err)
	}

	tmpDir := t.TempDir()
	sb := newMockSandbox("test-pod", tmpDir)
	sb.workDir = tmpDir

	config := map[string]interface{}{
		"mcp_port":    19000,
		"mcp_enabled": true,
		"model":       "gemini-2.0-flash",
	}

	err := m.Execute(context.Background(), "gemini-cli", sb, config)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify .gemini/settings.json was created
	geminiConfigPath := filepath.Join(tmpDir, ".gemini", "settings.json")
	if _, err := os.Stat(geminiConfigPath); os.IsNotExist(err) {
		t.Error("Gemini settings.json was not created")
	}
}

func TestPluginManagerExecuteCodexCLI(t *testing.T) {
	m := NewPluginManager("")
	if err := m.LoadPlugins(); err != nil {
		t.Fatalf("LoadPlugins() error = %v", err)
	}

	tmpDir := t.TempDir()
	sb := newMockSandbox("test-pod", tmpDir)
	sb.workDir = tmpDir

	config := map[string]interface{}{
		"mcp_port":    19000,
		"mcp_enabled": true,
		"model":       "o3",
	}

	err := m.Execute(context.Background(), "codex-cli", sb, config)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify launch args contain -c for MCP config
	args := sb.GetLaunchArgs()
	hasMinusC := false
	for _, arg := range args {
		if arg == "-c" {
			hasMinusC = true
			break
		}
	}

	if !hasMinusC {
		t.Error("Expected -c in launch args for codex-cli MCP config")
	}
}
