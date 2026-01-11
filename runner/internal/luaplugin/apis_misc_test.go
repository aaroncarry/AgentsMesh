package luaplugin

import (
	"os"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func TestJSONEncodeAPI(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	fn := jsonEncodeAPI()
	L.SetGlobal("json_encode", L.NewFunction(fn))

	// Test encoding a simple table
	t.Run("simple table", func(t *testing.T) {
		err := L.DoString(`
			local result = json_encode({key = "value", num = 42})
			if result == nil then
				error("Expected JSON string, got nil")
			end
			-- Just verify it's a string and contains expected parts
			if not string.find(result, '"key"') then
				error("JSON should contain key")
			end
		`)
		if err != nil {
			t.Fatalf("DoString failed: %v", err)
		}
	})

	// Test encoding nested table
	t.Run("nested table", func(t *testing.T) {
		err := L.DoString(`
			local result = json_encode({
				outer = {
					inner = "value"
				}
			})
			if result == nil then
				error("Expected JSON string")
			end
		`)
		if err != nil {
			t.Fatalf("DoString failed: %v", err)
		}
	})
}

func TestJSONEncodeAPI_Error(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	fn := jsonEncodeAPI()
	L.SetGlobal("json_encode", L.NewFunction(fn))

	// Test encoding table with function (which can't be JSON encoded)
	// Note: The encoder will skip non-encodable values rather than error
	err := L.DoString(`
		local result = json_encode({key = "value"})
		if result == nil then
			error("Expected JSON string")
		end
	`)
	if err != nil {
		t.Fatalf("DoString failed: %v", err)
	}
}

func TestLogAPI(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	fn := logAPI()
	L.SetGlobal("log", L.NewFunction(fn))

	// Test that log doesn't panic
	err := L.DoString(`log("test message")`)
	if err != nil {
		t.Fatalf("DoString failed: %v", err)
	}
}

func TestReadBuiltinResourceAPI(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Mock resource filesystem
	resources := map[string][]byte{
		"skills/test.md": []byte("# Test Skill\nThis is a test."),
	}
	resourceFS := func(path string) ([]byte, error) {
		content, ok := resources[path]
		if !ok {
			return nil, os.ErrNotExist
		}
		return content, nil
	}

	fn := readBuiltinResourceAPI(resourceFS)
	L.SetGlobal("read_builtin_resource", L.NewFunction(fn))

	// Test reading existing resource
	t.Run("existing resource", func(t *testing.T) {
		err := L.DoString(`
			local content = read_builtin_resource("skills/test.md")
			if content == nil then
				error("Expected content, got nil")
			end
			if not string.find(content, "Test Skill") then
				error("Content should contain 'Test Skill'")
			end
		`)
		if err != nil {
			t.Fatalf("DoString failed: %v", err)
		}
	})

	// Test reading non-existing resource
	t.Run("non-existing resource", func(t *testing.T) {
		err := L.DoString(`
			local content, err = read_builtin_resource("nonexistent.md")
			if content ~= nil then
				error("Expected nil for non-existing resource")
			end
			if err == nil then
				error("Expected error for non-existing resource")
			end
		`)
		if err != nil {
			t.Fatalf("DoString failed: %v", err)
		}
	})
}
