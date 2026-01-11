package runner

import (
	"testing"
)

// --- Test Capabilities (JSONB type) ---

func TestCapabilitiesScanNil(t *testing.T) {
	var caps Capabilities
	err := caps.Scan(nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if caps != nil {
		t.Error("expected nil Capabilities")
	}
}

func TestCapabilitiesScanValid(t *testing.T) {
	var caps Capabilities
	jsonData := []byte(`[{"name":"claude-code","version":"1.0.0","supported_agents":["claude-code"]}]`)
	err := caps.Scan(jsonData)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(caps) != 1 {
		t.Errorf("expected 1 capability, got %d", len(caps))
	}
	if caps[0].Name != "claude-code" {
		t.Errorf("expected name 'claude-code', got %s", caps[0].Name)
	}
}

func TestCapabilitiesScanInvalidType(t *testing.T) {
	var caps Capabilities
	err := caps.Scan("not bytes")
	if err == nil {
		t.Error("expected error for invalid type")
	}
}

func TestCapabilitiesScanInvalidJSON(t *testing.T) {
	var caps Capabilities
	err := caps.Scan([]byte(`invalid json`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestCapabilitiesValueNil(t *testing.T) {
	var caps Capabilities
	val, err := caps.Value()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if val != nil {
		t.Error("expected nil value")
	}
}

func TestCapabilitiesValueValid(t *testing.T) {
	caps := Capabilities{
		{Name: "test-plugin", Version: "1.0.0"},
	}
	val, err := caps.Value()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if val == nil {
		t.Error("expected non-nil value")
	}
	// Verify it's valid JSON
	bytes, ok := val.([]byte)
	if !ok {
		t.Error("expected []byte value")
	}
	if len(bytes) == 0 {
		t.Error("expected non-empty JSON")
	}
}

// --- Test PluginCapability ---

func TestPluginCapabilityValidate(t *testing.T) {
	tests := []struct {
		name        string
		capability  PluginCapability
		expectError bool
	}{
		{
			name: "valid capability",
			capability: PluginCapability{
				Name:            "test-plugin",
				Version:         "1.0.0",
				SupportedAgents: []string{"claude-code"},
			},
			expectError: false,
		},
		{
			name: "empty name",
			capability: PluginCapability{
				Name:    "",
				Version: "1.0.0",
			},
			expectError: true,
		},
		{
			name: "with UI config",
			capability: PluginCapability{
				Name:    "ui-plugin",
				Version: "1.0.0",
				UI: &UIConfig{
					Configurable: true,
					Fields:       []UIField{{Name: "enabled", Type: "boolean"}},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.capability.Validate()
			if tt.expectError && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateCapabilities(t *testing.T) {
	tests := []struct {
		name        string
		caps        []PluginCapability
		expectError bool
	}{
		{
			name:        "empty list",
			caps:        []PluginCapability{},
			expectError: false,
		},
		{
			name: "valid capabilities",
			caps: []PluginCapability{
				{Name: "plugin-1", Version: "1.0.0"},
				{Name: "plugin-2", Version: "2.0.0"},
			},
			expectError: false,
		},
		{
			name: "one invalid capability",
			caps: []PluginCapability{
				{Name: "valid-plugin", Version: "1.0.0"},
				{Name: "", Version: "1.0.0"}, // invalid - empty name
			},
			expectError: true,
		},
		{
			name: "first capability invalid",
			caps: []PluginCapability{
				{Name: "", Version: "1.0.0"},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCapabilities(tt.caps)
			if tt.expectError && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// --- Test Runner with Capabilities ---

func TestRunnerWithCapabilities(t *testing.T) {
	runner := Runner{
		ID:             1,
		OrganizationID: 100,
		NodeID:         "node-001",
		Status:         RunnerStatusOnline,
		IsEnabled:      true,
		Capabilities: Capabilities{
			{
				Name:            "claude-code",
				Version:         "1.0.0",
				Description:     "Claude Code Plugin",
				SupportedAgents: []string{"claude-code"},
				UI: &UIConfig{
					Configurable: true,
					Fields: []UIField{
						{Name: "mcp_enabled", Type: "boolean", Label: "Enable MCP", Default: true},
					},
				},
			},
			{
				Name:            "env",
				Version:         "1.0.0",
				Description:     "Environment Variables Plugin",
				SupportedAgents: []string{}, // supports all agents
			},
		},
	}

	if len(runner.Capabilities) != 2 {
		t.Errorf("expected 2 capabilities, got %d", len(runner.Capabilities))
	}

	claudePlugin := runner.Capabilities[0]
	if claudePlugin.Name != "claude-code" {
		t.Errorf("expected name 'claude-code', got %s", claudePlugin.Name)
	}
	if claudePlugin.UI == nil {
		t.Error("expected UI config to be present")
	}
	if len(claudePlugin.UI.Fields) != 1 {
		t.Errorf("expected 1 UI field, got %d", len(claudePlugin.UI.Fields))
	}

	envPlugin := runner.Capabilities[1]
	if len(envPlugin.SupportedAgents) != 0 {
		t.Errorf("expected empty SupportedAgents for env plugin, got %v", envPlugin.SupportedAgents)
	}
}

// --- Test Capabilities JSON Roundtrip ---

func TestCapabilitiesJSONRoundtrip(t *testing.T) {
	original := Capabilities{
		{
			Name:            "test-plugin",
			Version:         "1.0.0",
			Description:     "Test Plugin",
			SupportedAgents: []string{"agent-1", "agent-2"},
			UI: &UIConfig{
				Configurable: true,
				Fields: []UIField{
					{
						Name:    "enabled",
						Type:    "boolean",
						Label:   "Enable",
						Default: true,
					},
					{
						Name:  "mode",
						Type:  "select",
						Label: "Mode",
						Options: []UIOption{
							{Value: "auto", Label: "Auto"},
							{Value: "manual", Label: "Manual"},
						},
					},
				},
			},
		},
	}

	// Marshal to JSON (Value)
	val, err := original.Value()
	if err != nil {
		t.Fatalf("Value() error: %v", err)
	}

	// Unmarshal from JSON (Scan)
	var restored Capabilities
	err = restored.Scan(val)
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	// Verify roundtrip
	if len(restored) != len(original) {
		t.Fatalf("expected %d capabilities, got %d", len(original), len(restored))
	}

	restoredPlugin := restored[0]
	originalPlugin := original[0]

	if restoredPlugin.Name != originalPlugin.Name {
		t.Errorf("expected name %s, got %s", originalPlugin.Name, restoredPlugin.Name)
	}
	if restoredPlugin.Version != originalPlugin.Version {
		t.Errorf("expected version %s, got %s", originalPlugin.Version, restoredPlugin.Version)
	}
	if len(restoredPlugin.SupportedAgents) != len(originalPlugin.SupportedAgents) {
		t.Errorf("expected %d supported agents, got %d",
			len(originalPlugin.SupportedAgents), len(restoredPlugin.SupportedAgents))
	}
	if restoredPlugin.UI == nil {
		t.Fatal("expected UI config to be restored")
	}
	if len(restoredPlugin.UI.Fields) != len(originalPlugin.UI.Fields) {
		t.Errorf("expected %d UI fields, got %d",
			len(originalPlugin.UI.Fields), len(restoredPlugin.UI.Fields))
	}
}
