package runner

import (
	"testing"
)

// --- Test UIConfig ---

func TestUIConfigStruct(t *testing.T) {
	config := UIConfig{
		Configurable: true,
		Fields: []UIField{
			{Name: "enabled", Type: "boolean", Label: "Enable", Default: true},
			{Name: "model", Type: "select", Label: "Model"},
		},
	}

	if !config.Configurable {
		t.Error("expected Configurable to be true")
	}
	if len(config.Fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(config.Fields))
	}
}

// --- Test UIField ---

func TestUIFieldStruct(t *testing.T) {
	minVal := float64(1)
	maxVal := float64(100)

	field := UIField{
		Name:        "count",
		Type:        "number",
		Label:       "Count",
		Default:     10,
		Description: "Number of items",
		Placeholder: "Enter count",
		Min:         &minVal,
		Max:         &maxVal,
		Required:    true,
	}

	if field.Name != "count" {
		t.Errorf("expected name 'count', got %s", field.Name)
	}
	if field.Type != "number" {
		t.Errorf("expected type 'number', got %s", field.Type)
	}
	if *field.Min != 1 {
		t.Errorf("expected min 1, got %f", *field.Min)
	}
	if *field.Max != 100 {
		t.Errorf("expected max 100, got %f", *field.Max)
	}
	if !field.Required {
		t.Error("expected Required to be true")
	}
}

func TestUIFieldWithOptions(t *testing.T) {
	field := UIField{
		Name:  "mode",
		Type:  "select",
		Label: "Mode",
		Options: []UIOption{
			{Value: "auto", Label: "Auto"},
			{Value: "manual", Label: "Manual"},
		},
	}

	if len(field.Options) != 2 {
		t.Errorf("expected 2 options, got %d", len(field.Options))
	}
	if field.Options[0].Value != "auto" {
		t.Errorf("expected first option value 'auto', got %s", field.Options[0].Value)
	}
	if field.Options[1].Label != "Manual" {
		t.Errorf("expected second option label 'Manual', got %s", field.Options[1].Label)
	}
}

// --- Test UIFieldType Constants ---

func TestUIFieldTypeConstants(t *testing.T) {
	if UIFieldTypeBoolean != "boolean" {
		t.Errorf("expected 'boolean', got %s", UIFieldTypeBoolean)
	}
	if UIFieldTypeString != "string" {
		t.Errorf("expected 'string', got %s", UIFieldTypeString)
	}
	if UIFieldTypeSelect != "select" {
		t.Errorf("expected 'select', got %s", UIFieldTypeSelect)
	}
	if UIFieldTypeNumber != "number" {
		t.Errorf("expected 'number', got %s", UIFieldTypeNumber)
	}
	if UIFieldTypeSecret != "secret" {
		t.Errorf("expected 'secret', got %s", UIFieldTypeSecret)
	}
}

// --- Test UIOption ---

func TestUIOptionStruct(t *testing.T) {
	option := UIOption{
		Value: "test-value",
		Label: "Test Label",
	}

	if option.Value != "test-value" {
		t.Errorf("expected value 'test-value', got %s", option.Value)
	}
	if option.Label != "Test Label" {
		t.Errorf("expected label 'Test Label', got %s", option.Label)
	}
}
