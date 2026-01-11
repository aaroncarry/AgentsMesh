package luaplugin

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func TestParseUIConfig(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Create UI config table
	uiTbl := L.NewTable()
	uiTbl.RawSetString("configurable", lua.LBool(true))

	// Create fields array
	fieldsTbl := L.NewTable()

	// Add field
	fieldTbl := L.NewTable()
	fieldTbl.RawSetString("name", lua.LString("mcp_enabled"))
	fieldTbl.RawSetString("type", lua.LString("boolean"))
	fieldTbl.RawSetString("label", lua.LString("Enable MCP"))
	fieldTbl.RawSetString("default", lua.LBool(true))
	fieldsTbl.Append(fieldTbl)

	uiTbl.RawSetString("fields", fieldsTbl)

	// Parse
	result := parseUIConfig(uiTbl)

	if !result.Configurable {
		t.Error("Expected Configurable to be true")
	}
	if len(result.Fields) != 1 {
		t.Errorf("Expected 1 field, got %d", len(result.Fields))
	}
	if result.Fields[0].Name != "mcp_enabled" {
		t.Errorf("Expected field name 'mcp_enabled', got %q", result.Fields[0].Name)
	}
}

func TestParseUIField(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Test full field with all properties
	t.Run("full field", func(t *testing.T) {
		fieldTbl := L.NewTable()
		fieldTbl.RawSetString("name", lua.LString("timeout"))
		fieldTbl.RawSetString("type", lua.LString("number"))
		fieldTbl.RawSetString("label", lua.LString("Timeout (sec)"))
		fieldTbl.RawSetString("description", lua.LString("Connection timeout"))
		fieldTbl.RawSetString("placeholder", lua.LString("Enter timeout"))
		fieldTbl.RawSetString("required", lua.LBool(true))
		fieldTbl.RawSetString("default", lua.LNumber(60))
		fieldTbl.RawSetString("min", lua.LNumber(1))
		fieldTbl.RawSetString("max", lua.LNumber(300))

		field := parseUIField(fieldTbl)

		if field.Name != "timeout" {
			t.Errorf("Expected name 'timeout', got %q", field.Name)
		}
		if field.Type != "number" {
			t.Errorf("Expected type 'number', got %q", field.Type)
		}
		if field.Label != "Timeout (sec)" {
			t.Errorf("Expected label 'Timeout (sec)', got %q", field.Label)
		}
		if field.Description != "Connection timeout" {
			t.Errorf("Expected description 'Connection timeout', got %q", field.Description)
		}
		if field.Placeholder != "Enter timeout" {
			t.Errorf("Expected placeholder 'Enter timeout', got %q", field.Placeholder)
		}
		if !field.Required {
			t.Error("Expected Required to be true")
		}
		if field.Default != float64(60) {
			t.Errorf("Expected default 60, got %v", field.Default)
		}
		if field.Min == nil || *field.Min != 1 {
			t.Errorf("Expected min 1, got %v", field.Min)
		}
		if field.Max == nil || *field.Max != 300 {
			t.Errorf("Expected max 300, got %v", field.Max)
		}
	})

	// Test select field with options
	t.Run("select field with options", func(t *testing.T) {
		fieldTbl := L.NewTable()
		fieldTbl.RawSetString("name", lua.LString("mode"))
		fieldTbl.RawSetString("type", lua.LString("select"))
		fieldTbl.RawSetString("label", lua.LString("Mode"))
		fieldTbl.RawSetString("default", lua.LString("plan"))

		optionsTbl := L.NewTable()

		opt1 := L.NewTable()
		opt1.RawSetString("value", lua.LString("plan"))
		opt1.RawSetString("label", lua.LString("Plan Mode"))
		optionsTbl.Append(opt1)

		opt2 := L.NewTable()
		opt2.RawSetString("value", lua.LString("auto"))
		opt2.RawSetString("label", lua.LString("Auto Mode"))
		optionsTbl.Append(opt2)

		fieldTbl.RawSetString("options", optionsTbl)

		field := parseUIField(fieldTbl)

		if field.Default != "plan" {
			t.Errorf("Expected default 'plan', got %v", field.Default)
		}
		if len(field.Options) != 2 {
			t.Errorf("Expected 2 options, got %d", len(field.Options))
		}
		if field.Options[0].Value != "plan" || field.Options[0].Label != "Plan Mode" {
			t.Errorf("First option mismatch: %v", field.Options[0])
		}
	})

	// Test boolean default
	t.Run("boolean default", func(t *testing.T) {
		fieldTbl := L.NewTable()
		fieldTbl.RawSetString("name", lua.LString("enabled"))
		fieldTbl.RawSetString("type", lua.LString("boolean"))
		fieldTbl.RawSetString("label", lua.LString("Enabled"))
		fieldTbl.RawSetString("default", lua.LBool(true))

		field := parseUIField(fieldTbl)

		if field.Default != true {
			t.Errorf("Expected default true, got %v", field.Default)
		}
	})
}
