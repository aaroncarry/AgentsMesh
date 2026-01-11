package luaplugin

import (
	"os"
	"path/filepath"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func TestReadFileAPI(t *testing.T) {
	tmpDir := t.TempDir()
	sb := newMockSandbox("test-pod", tmpDir)

	L := lua.NewState()
	defer L.Close()

	// Create test file
	testPath := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testPath, []byte("test content"), 0644)

	// Test successful read
	t.Run("successful read", func(t *testing.T) {
		fn := readFileAPI(sb)
		L.SetGlobal("read_file", L.NewFunction(fn))

		err := L.DoString(`
			local content = read_file("` + testPath + `")
			if content ~= "test content" then
				error("Content mismatch: " .. tostring(content))
			end
		`)
		if err != nil {
			t.Fatalf("DoString failed: %v", err)
		}
	})

	// Test reading non-existent file
	t.Run("non-existent file", func(t *testing.T) {
		fn := readFileAPI(sb)
		L.SetGlobal("read_file", L.NewFunction(fn))

		err := L.DoString(`
			local content, err = read_file("` + filepath.Join(tmpDir, "nonexistent.txt") + `")
			if content ~= nil then
				error("Expected nil for non-existent file")
			end
			if err == nil then
				error("Expected error message")
			end
		`)
		if err != nil {
			t.Fatalf("DoString failed: %v", err)
		}
	})

	// Test path traversal protection
	t.Run("path traversal blocked", func(t *testing.T) {
		fn := readFileAPI(sb)
		L.SetGlobal("read_file", L.NewFunction(fn))

		err := L.DoString(`
			local content, err = read_file("/etc/passwd")
			if content ~= nil then
				error("Expected nil for blocked path")
			end
		`)
		if err != nil {
			t.Fatalf("DoString failed: %v", err)
		}
	})
}

func TestReadJSONAPI(t *testing.T) {
	tmpDir := t.TempDir()
	sb := newMockSandbox("test-pod", tmpDir)

	L := lua.NewState()
	defer L.Close()

	// Create test JSON file
	testPath := filepath.Join(tmpDir, "test.json")
	os.WriteFile(testPath, []byte(`{"key": "value", "number": 42}`), 0644)

	// Test successful JSON read
	t.Run("successful read", func(t *testing.T) {
		fn := readJSONAPI(sb)
		L.SetGlobal("read_json", L.NewFunction(fn))

		err := L.DoString(`
			local data = read_json("` + testPath + `")
			if data == nil then
				error("Expected table, got nil")
			end
			if data.key ~= "value" then
				error("Expected key='value', got: " .. tostring(data.key))
			end
			if data.number ~= 42 then
				error("Expected number=42, got: " .. tostring(data.number))
			end
		`)
		if err != nil {
			t.Fatalf("DoString failed: %v", err)
		}
	})

	// Test reading non-existent JSON file (returns nil, not error)
	t.Run("non-existent file returns nil", func(t *testing.T) {
		fn := readJSONAPI(sb)
		L.SetGlobal("read_json", L.NewFunction(fn))

		err := L.DoString(`
			local data = read_json("` + filepath.Join(tmpDir, "nonexistent.json") + `")
			if data ~= nil then
				error("Expected nil for non-existent file")
			end
		`)
		if err != nil {
			t.Fatalf("DoString failed: %v", err)
		}
	})

	// Test invalid JSON
	t.Run("invalid JSON returns error", func(t *testing.T) {
		invalidPath := filepath.Join(tmpDir, "invalid.json")
		os.WriteFile(invalidPath, []byte(`not valid json`), 0644)

		fn := readJSONAPI(sb)
		L.SetGlobal("read_json", L.NewFunction(fn))

		err := L.DoString(`
			local data, err = read_json("` + invalidPath + `")
			if data ~= nil then
				error("Expected nil for invalid JSON")
			end
			if err == nil then
				error("Expected error for invalid JSON")
			end
		`)
		if err != nil {
			t.Fatalf("DoString failed: %v", err)
		}
	})
}

func TestReadJSONAPI_PathValidationError(t *testing.T) {
	tmpDir := t.TempDir()
	sb := newMockSandbox("test-pod", tmpDir)

	L := lua.NewState()
	defer L.Close()

	fn := readJSONAPI(sb)
	L.SetGlobal("read_json", L.NewFunction(fn))

	// Try to read JSON from outside sandbox
	err := L.DoString(`
		local data, err = read_json("/etc/passwd")
		if data ~= nil then
			error("Expected nil for path outside sandbox")
		end
	`)
	if err != nil {
		t.Fatalf("DoString failed: %v", err)
	}
}
