package runner

import (
	"testing"

	"github.com/anthropics/agentmesh/backend/internal/domain/runner"
)

// --- GetPluginOptions Tests ---

func TestGetPluginOptionsNilCapabilities(t *testing.T) {
	service := NewService(nil)

	r := &runner.Runner{
		ID:           1,
		Capabilities: nil,
	}

	result := service.GetPluginOptions(r, "claude-code")
	if result != nil {
		t.Errorf("expected nil result for nil capabilities, got %v", result)
	}
}

func TestGetPluginOptionsEmptyCapabilities(t *testing.T) {
	service := NewService(nil)

	r := &runner.Runner{
		ID:           1,
		Capabilities: runner.Capabilities{},
	}

	result := service.GetPluginOptions(r, "claude-code")
	if len(result) != 0 {
		t.Errorf("expected empty result for empty capabilities, got %d", len(result))
	}
}

func TestGetPluginOptionsFilterByAgentType(t *testing.T) {
	service := NewService(nil)

	r := &runner.Runner{
		ID: 1,
		Capabilities: runner.Capabilities{
			{
				Name:            "claude-code",
				SupportedAgents: []string{"claude-code"},
				UI:              &runner.UIConfig{Configurable: true},
			},
			{
				Name:            "gemini-cli",
				SupportedAgents: []string{"gemini-cli"},
				UI:              &runner.UIConfig{Configurable: true},
			},
			{
				Name:            "env",
				SupportedAgents: []string{}, // supports all agents
				UI:              &runner.UIConfig{Configurable: true},
			},
		},
	}

	// Filter for claude-code
	result := service.GetPluginOptions(r, "claude-code")
	if len(result) != 2 {
		t.Errorf("expected 2 plugins for claude-code (claude-code + env), got %d", len(result))
	}

	// Verify the correct plugins are returned
	names := make(map[string]bool)
	for _, cap := range result {
		names[cap.Name] = true
	}
	if !names["claude-code"] {
		t.Error("expected claude-code plugin to be included")
	}
	if !names["env"] {
		t.Error("expected env plugin to be included (supports all agents)")
	}
	if names["gemini-cli"] {
		t.Error("expected gemini-cli plugin to NOT be included")
	}
}

func TestGetPluginOptionsNoAgentFilter(t *testing.T) {
	service := NewService(nil)

	r := &runner.Runner{
		ID: 1,
		Capabilities: runner.Capabilities{
			{
				Name:            "claude-code",
				SupportedAgents: []string{"claude-code"},
				UI:              &runner.UIConfig{Configurable: true},
			},
			{
				Name:            "gemini-cli",
				SupportedAgents: []string{"gemini-cli"},
				UI:              &runner.UIConfig{Configurable: true},
			},
		},
	}

	// No agent filter - should return all configurable plugins
	result := service.GetPluginOptions(r, "")
	if len(result) != 2 {
		t.Errorf("expected 2 plugins with no filter, got %d", len(result))
	}
}

func TestGetPluginOptionsOnlyConfigurable(t *testing.T) {
	service := NewService(nil)

	r := &runner.Runner{
		ID: 1,
		Capabilities: runner.Capabilities{
			{
				Name:            "configurable-plugin",
				SupportedAgents: []string{"test-agent"},
				UI:              &runner.UIConfig{Configurable: true, Fields: []runner.UIField{{Name: "enabled"}}},
			},
			{
				Name:            "non-configurable-plugin",
				SupportedAgents: []string{"test-agent"},
				UI:              &runner.UIConfig{Configurable: false},
			},
			{
				Name:            "no-ui-plugin",
				SupportedAgents: []string{"test-agent"},
				UI:              nil,
			},
		},
	}

	result := service.GetPluginOptions(r, "test-agent")
	if len(result) != 1 {
		t.Errorf("expected 1 configurable plugin, got %d", len(result))
	}
	if result[0].Name != "configurable-plugin" {
		t.Errorf("expected 'configurable-plugin', got %s", result[0].Name)
	}
}

func TestGetPluginOptionsWithUIFields(t *testing.T) {
	service := NewService(nil)

	minVal := float64(1)
	maxVal := float64(100)

	r := &runner.Runner{
		ID: 1,
		Capabilities: runner.Capabilities{
			{
				Name:            "test-plugin",
				Version:         "1.0.0",
				Description:     "Test Plugin",
				SupportedAgents: []string{"test-agent"},
				UI: &runner.UIConfig{
					Configurable: true,
					Fields: []runner.UIField{
						{Name: "enabled", Type: "boolean", Label: "Enable", Default: true},
						{Name: "mode", Type: "select", Label: "Mode", Options: []runner.UIOption{
							{Value: "auto", Label: "Auto"},
							{Value: "manual", Label: "Manual"},
						}},
						{Name: "count", Type: "number", Label: "Count", Min: &minVal, Max: &maxVal},
					},
				},
			},
		},
	}

	result := service.GetPluginOptions(r, "test-agent")
	if len(result) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(result))
	}

	plugin := result[0]
	if plugin.Name != "test-plugin" {
		t.Errorf("expected name 'test-plugin', got %s", plugin.Name)
	}
	if plugin.UI == nil {
		t.Fatal("expected UI config")
	}
	if len(plugin.UI.Fields) != 3 {
		t.Errorf("expected 3 UI fields, got %d", len(plugin.UI.Fields))
	}

	// Verify field types
	fieldTypes := make(map[string]string)
	for _, f := range plugin.UI.Fields {
		fieldTypes[f.Name] = f.Type
	}
	if fieldTypes["enabled"] != "boolean" {
		t.Errorf("expected 'enabled' type 'boolean', got %s", fieldTypes["enabled"])
	}
	if fieldTypes["mode"] != "select" {
		t.Errorf("expected 'mode' type 'select', got %s", fieldTypes["mode"])
	}
	if fieldTypes["count"] != "number" {
		t.Errorf("expected 'count' type 'number', got %s", fieldTypes["count"])
	}
}

func TestGetPluginOptionsMultipleAgentsSupported(t *testing.T) {
	service := NewService(nil)

	r := &runner.Runner{
		ID: 1,
		Capabilities: runner.Capabilities{
			{
				Name:            "multi-agent-plugin",
				SupportedAgents: []string{"agent-a", "agent-b", "agent-c"},
				UI:              &runner.UIConfig{Configurable: true},
			},
		},
	}

	// Should match agent-b
	result := service.GetPluginOptions(r, "agent-b")
	if len(result) != 1 {
		t.Errorf("expected 1 plugin for agent-b, got %d", len(result))
	}

	// Should not match agent-d
	result = service.GetPluginOptions(r, "agent-d")
	if len(result) != 0 {
		t.Errorf("expected 0 plugins for agent-d, got %d", len(result))
	}
}
