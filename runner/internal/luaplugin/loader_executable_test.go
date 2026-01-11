package luaplugin

import (
	"testing"
)

func TestCheckExecutable(t *testing.T) {
	loader := NewPluginLoader()

	tests := []struct {
		name       string
		plugin     *LuaPlugin
		wantAvail  bool
	}{
		{
			name: "no executable requirement",
			plugin: &LuaPlugin{
				Name:       "test-plugin",
				Executable: "",
			},
			wantAvail: true,
		},
		{
			name: "common executable exists (ls)",
			plugin: &LuaPlugin{
				Name:       "test-plugin",
				Executable: "ls",
			},
			wantAvail: true,
		},
		{
			name: "non-existent executable",
			plugin: &LuaPlugin{
				Name:       "test-plugin",
				Executable: "non-existent-command-xyz-12345",
			},
			wantAvail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset availability
			tt.plugin.isAvailable = false

			loader.checkExecutable(tt.plugin)

			if tt.plugin.IsAvailable() != tt.wantAvail {
				t.Errorf("IsAvailable() = %v, want %v", tt.plugin.IsAvailable(), tt.wantAvail)
			}
		})
	}
}

func TestGetCapabilitiesIncludesAvailabilityStatus(t *testing.T) {
	// Create a manager with mock plugins
	m := &PluginManager{
		plugins: []*LuaPlugin{
			{
				Name:        "available-plugin",
				Version:     "1.0.0",
				Description: "An available plugin",
				Executable:  "ls",
				isAvailable: true,
			},
			{
				Name:        "unavailable-plugin",
				Version:     "1.0.0",
				Description: "An unavailable plugin",
				Executable:  "non-existent-command",
				isAvailable: false,
			},
			{
				Name:        "another-available-plugin",
				Version:     "2.0.0",
				Description: "Another available plugin",
				isAvailable: true,
			},
		},
	}

	caps := m.GetCapabilities()

	// Should include all 3 plugins
	if len(caps) != 3 {
		t.Errorf("GetCapabilities() returned %d plugins, want 3", len(caps))
	}

	// Check availability status for each plugin
	capMap := make(map[string]PluginCapability)
	for _, cap := range caps {
		capMap[cap.Name] = cap
	}

	// Check available-plugin
	if cap, ok := capMap["available-plugin"]; !ok {
		t.Error("GetCapabilities() should include available-plugin")
	} else {
		if !cap.Available {
			t.Error("available-plugin should have Available=true")
		}
		if cap.Executable != "ls" {
			t.Errorf("available-plugin.Executable = %q, want %q", cap.Executable, "ls")
		}
	}

	// Check unavailable-plugin
	if cap, ok := capMap["unavailable-plugin"]; !ok {
		t.Error("GetCapabilities() should include unavailable-plugin")
	} else {
		if cap.Available {
			t.Error("unavailable-plugin should have Available=false")
		}
		if cap.Executable != "non-existent-command" {
			t.Errorf("unavailable-plugin.Executable = %q, want %q", cap.Executable, "non-existent-command")
		}
	}

	// Check another-available-plugin
	if cap, ok := capMap["another-available-plugin"]; !ok {
		t.Error("GetCapabilities() should include another-available-plugin")
	} else {
		if !cap.Available {
			t.Error("another-available-plugin should have Available=true")
		}
	}
}

func TestPluginParserParsesExecutable(t *testing.T) {
	parser := NewPluginParser()

	content := []byte(`
plugin = {
    name = "test-plugin",
    version = "1.0.0",
    description = "Test plugin",
    executable = "test-command",
    supported_agents = {"test-agent"},
}

function setup(ctx)
end
`)

	plugin, err := parser.Parse("test.lua", content, true)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if plugin.Executable != "test-command" {
		t.Errorf("Executable = %q, want %q", plugin.Executable, "test-command")
	}
}

func TestLoadFromContentSetsAvailability(t *testing.T) {
	loader := NewPluginLoader()

	tests := []struct {
		name        string
		content     string
		wantAvail   bool
	}{
		{
			name: "plugin with existing executable",
			content: `
plugin = {
    name = "test-ls",
    version = "1.0.0",
    description = "Test with ls",
    executable = "ls",
}
function setup(ctx) end
`,
			wantAvail: true,
		},
		{
			name: "plugin with non-existing executable",
			content: `
plugin = {
    name = "test-missing",
    version = "1.0.0",
    description = "Test with missing",
    executable = "non-existent-xyz-99999",
}
function setup(ctx) end
`,
			wantAvail: false,
		},
		{
			name: "plugin without executable requirement",
			content: `
plugin = {
    name = "test-no-exec",
    version = "1.0.0",
    description = "No executable needed",
}
function setup(ctx) end
`,
			wantAvail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin, err := loader.loadFromContent("test.lua", []byte(tt.content), true)
			if err != nil {
				t.Fatalf("loadFromContent() error = %v", err)
			}

			if plugin.IsAvailable() != tt.wantAvail {
				t.Errorf("IsAvailable() = %v, want %v", plugin.IsAvailable(), tt.wantAvail)
			}
		})
	}
}

func TestIsAvailableMethod(t *testing.T) {
	// Test the IsAvailable() accessor method
	plugin := &LuaPlugin{
		Name:        "test",
		isAvailable: true,
	}

	if !plugin.IsAvailable() {
		t.Error("IsAvailable() should return true when isAvailable is true")
	}

	plugin.isAvailable = false
	if plugin.IsAvailable() {
		t.Error("IsAvailable() should return false when isAvailable is false")
	}
}
