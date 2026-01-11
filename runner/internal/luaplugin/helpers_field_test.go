package luaplugin

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func TestGetStringField(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	tbl := L.NewTable()
	tbl.RawSetString("name", lua.LString("test-value"))
	tbl.RawSetString("number", lua.LNumber(42))

	// Test getting string field
	result := getStringField(tbl, "name")
	if result != "test-value" {
		t.Errorf("Expected 'test-value', got %q", result)
	}

	// Test getting non-string field (should return empty string)
	result = getStringField(tbl, "number")
	if result != "" {
		t.Errorf("Expected empty string for non-string field, got %q", result)
	}

	// Test getting non-existent field
	result = getStringField(tbl, "nonexistent")
	if result != "" {
		t.Errorf("Expected empty string for non-existent field, got %q", result)
	}
}

func TestGetIntField(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	tbl := L.NewTable()
	tbl.RawSetString("order", lua.LNumber(42))
	tbl.RawSetString("name", lua.LString("test"))

	// Test getting int field
	result := getIntField(tbl, "order")
	if result != 42 {
		t.Errorf("Expected 42, got %d", result)
	}

	// Test getting non-number field (should return 0)
	result = getIntField(tbl, "name")
	if result != 0 {
		t.Errorf("Expected 0 for non-number field, got %d", result)
	}

	// Test getting non-existent field
	result = getIntField(tbl, "nonexistent")
	if result != 0 {
		t.Errorf("Expected 0 for non-existent field, got %d", result)
	}
}

func TestGetBoolField(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	tbl := L.NewTable()
	tbl.RawSetString("enabled", lua.LBool(true))
	tbl.RawSetString("disabled", lua.LBool(false))
	tbl.RawSetString("name", lua.LString("test"))

	// Test getting true bool field
	result := getBoolField(tbl, "enabled")
	if !result {
		t.Error("Expected true, got false")
	}

	// Test getting false bool field
	result = getBoolField(tbl, "disabled")
	if result {
		t.Error("Expected false, got true")
	}

	// Test getting non-bool field (should return false)
	result = getBoolField(tbl, "name")
	if result {
		t.Error("Expected false for non-bool field")
	}
}

func TestGetStringArrayField(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	tbl := L.NewTable()

	// Create array table
	arrTbl := L.NewTable()
	arrTbl.Append(lua.LString("value1"))
	arrTbl.Append(lua.LString("value2"))
	arrTbl.Append(lua.LString("value3"))
	tbl.RawSetString("agents", arrTbl)

	// Test getting string array
	result := getStringArrayField(tbl, "agents")
	if len(result) != 3 {
		t.Errorf("Expected 3 items, got %d", len(result))
	}
	if result[0] != "value1" || result[1] != "value2" || result[2] != "value3" {
		t.Errorf("Array content mismatch: %v", result)
	}

	// Test with mixed types (should only include strings)
	mixedTbl := L.NewTable()
	mixedTbl.Append(lua.LString("string"))
	mixedTbl.Append(lua.LNumber(42)) // should be skipped
	tbl.RawSetString("mixed", mixedTbl)

	result = getStringArrayField(tbl, "mixed")
	if len(result) != 1 || result[0] != "string" {
		t.Errorf("Mixed array should only include strings: %v", result)
	}

	// Test non-existent field
	result = getStringArrayField(tbl, "nonexistent")
	if result != nil {
		t.Errorf("Expected nil for non-existent field, got %v", result)
	}

	// Test non-table field
	tbl.RawSetString("notarray", lua.LString("string"))
	result = getStringArrayField(tbl, "notarray")
	if result != nil {
		t.Errorf("Expected nil for non-table field, got %v", result)
	}
}
