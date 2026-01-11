package luaplugin

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func TestGoValueToLua(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Test nil
	t.Run("nil", func(t *testing.T) {
		result := goValueToLua(L, nil)
		if result != lua.LNil {
			t.Errorf("Expected LNil, got %v", result)
		}
	})

	// Test bool
	t.Run("bool true", func(t *testing.T) {
		result := goValueToLua(L, true)
		if result != lua.LBool(true) {
			t.Errorf("Expected LBool(true), got %v", result)
		}
	})

	t.Run("bool false", func(t *testing.T) {
		result := goValueToLua(L, false)
		if result != lua.LBool(false) {
			t.Errorf("Expected LBool(false), got %v", result)
		}
	})

	// Test int
	t.Run("int", func(t *testing.T) {
		result := goValueToLua(L, 42)
		if result != lua.LNumber(42) {
			t.Errorf("Expected LNumber(42), got %v", result)
		}
	})

	// Test int64
	t.Run("int64", func(t *testing.T) {
		result := goValueToLua(L, int64(9999))
		if result != lua.LNumber(9999) {
			t.Errorf("Expected LNumber(9999), got %v", result)
		}
	})

	// Test float64
	t.Run("float64", func(t *testing.T) {
		result := goValueToLua(L, 3.14)
		if result != lua.LNumber(3.14) {
			t.Errorf("Expected LNumber(3.14), got %v", result)
		}
	})

	// Test string
	t.Run("string", func(t *testing.T) {
		result := goValueToLua(L, "hello")
		if result != lua.LString("hello") {
			t.Errorf("Expected LString('hello'), got %v", result)
		}
	})

	// Test []string
	t.Run("[]string", func(t *testing.T) {
		result := goValueToLua(L, []string{"a", "b", "c"})
		tbl, ok := result.(*lua.LTable)
		if !ok {
			t.Fatalf("Expected table, got %T", result)
		}
		if tbl.RawGetInt(1) != lua.LString("a") {
			t.Error("First element should be 'a'")
		}
		if tbl.RawGetInt(2) != lua.LString("b") {
			t.Error("Second element should be 'b'")
		}
		if tbl.RawGetInt(3) != lua.LString("c") {
			t.Error("Third element should be 'c'")
		}
	})

	// Test map[string]string
	t.Run("map[string]string", func(t *testing.T) {
		result := goValueToLua(L, map[string]string{"key": "value"})
		tbl, ok := result.(*lua.LTable)
		if !ok {
			t.Fatalf("Expected table, got %T", result)
		}
		if tbl.RawGetString("key") != lua.LString("value") {
			t.Error("key should be 'value'")
		}
	})

	// Test map[string]interface{}
	t.Run("map[string]interface{}", func(t *testing.T) {
		result := goValueToLua(L, map[string]interface{}{
			"str":  "value",
			"num":  float64(42),
			"bool": true,
		})
		tbl, ok := result.(*lua.LTable)
		if !ok {
			t.Fatalf("Expected table, got %T", result)
		}
		if tbl.RawGetString("str") != lua.LString("value") {
			t.Error("str should be 'value'")
		}
		if tbl.RawGetString("num") != lua.LNumber(42) {
			t.Error("num should be 42")
		}
		if tbl.RawGetString("bool") != lua.LBool(true) {
			t.Error("bool should be true")
		}
	})

	// Test unsupported type
	t.Run("unsupported type", func(t *testing.T) {
		result := goValueToLua(L, struct{}{})
		if result != lua.LNil {
			t.Errorf("Expected LNil for unsupported type, got %v", result)
		}
	})
}

func TestLuaTableToGoMap(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	tbl := L.NewTable()
	tbl.RawSetString("str", lua.LString("value"))
	tbl.RawSetString("num", lua.LNumber(42))
	tbl.RawSetString("bool", lua.LBool(true))

	result := luaTableToGoMap(tbl)

	if result["str"] != "value" {
		t.Errorf("Expected str='value', got %v", result["str"])
	}
	if result["num"] != float64(42) {
		t.Errorf("Expected num=42, got %v", result["num"])
	}
	if result["bool"] != true {
		t.Errorf("Expected bool=true, got %v", result["bool"])
	}

	// Test with integer keys (should be ignored)
	tbl.RawSetInt(1, lua.LString("indexed"))
	result = luaTableToGoMap(tbl)
	if _, exists := result["1"]; exists {
		t.Error("Integer keys should be ignored")
	}
}

func TestLuaValueToGo(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Test bool
	t.Run("bool", func(t *testing.T) {
		result := luaValueToGo(lua.LBool(true))
		if result != true {
			t.Errorf("Expected true, got %v", result)
		}
	})

	// Test number
	t.Run("number", func(t *testing.T) {
		result := luaValueToGo(lua.LNumber(3.14))
		if result != float64(3.14) {
			t.Errorf("Expected 3.14, got %v", result)
		}
	})

	// Test string
	t.Run("string", func(t *testing.T) {
		result := luaValueToGo(lua.LString("hello"))
		if result != "hello" {
			t.Errorf("Expected 'hello', got %v", result)
		}
	})

	// Test table
	t.Run("table", func(t *testing.T) {
		tbl := L.NewTable()
		tbl.RawSetString("key", lua.LString("value"))
		result := luaValueToGo(tbl)
		m, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected map, got %T", result)
		}
		if m["key"] != "value" {
			t.Errorf("Expected key='value', got %v", m["key"])
		}
	})

	// Test nil
	t.Run("nil", func(t *testing.T) {
		result := luaValueToGo(lua.LNil)
		if result != nil {
			t.Errorf("Expected nil, got %v", result)
		}
	})

	// Test function (unsupported, returns nil)
	t.Run("function", func(t *testing.T) {
		fn := L.NewFunction(func(L *lua.LState) int { return 0 })
		result := luaValueToGo(fn)
		if result != nil {
			t.Errorf("Expected nil for function, got %v", result)
		}
	})
}
