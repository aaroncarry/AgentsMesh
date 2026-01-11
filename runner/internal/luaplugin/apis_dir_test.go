package luaplugin

import (
	"os"
	"path/filepath"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func TestMkdirAPI(t *testing.T) {
	tmpDir := t.TempDir()
	sb := newMockSandbox("test-pod", tmpDir)

	L := lua.NewState()
	defer L.Close()

	// Test successful mkdir
	t.Run("successful mkdir", func(t *testing.T) {
		fn := mkdirAPI(sb)
		L.SetGlobal("mkdir", L.NewFunction(fn))

		testPath := filepath.Join(tmpDir, "new", "nested", "dir")
		err := L.DoString(`
			local ok = mkdir("` + testPath + `")
			if not ok then
				error("mkdir should return true")
			end
		`)
		if err != nil {
			t.Fatalf("DoString failed: %v", err)
		}

		if _, err := os.Stat(testPath); os.IsNotExist(err) {
			t.Error("Directory was not created")
		}
	})

	// Test path traversal protection
	t.Run("path traversal blocked", func(t *testing.T) {
		fn := mkdirAPI(sb)
		L.SetGlobal("mkdir", L.NewFunction(fn))

		err := L.DoString(`
			local ok, err = mkdir("/tmp/outside")
			if ok ~= nil then
				error("Expected nil for blocked path")
			end
		`)
		if err != nil {
			t.Fatalf("DoString failed: %v", err)
		}
	})
}

func TestMkdirAPI_ErrorPath(t *testing.T) {
	tmpDir := t.TempDir()
	sb := newMockSandbox("test-pod", tmpDir)

	// Create a file that will block directory creation
	blocker := filepath.Join(tmpDir, "blocker_file")
	os.WriteFile(blocker, []byte("content"), 0644)

	L := lua.NewState()
	defer L.Close()

	fn := mkdirAPI(sb)
	L.SetGlobal("mkdir", L.NewFunction(fn))

	// Try to create a directory where a file exists
	err := L.DoString(`
		local ok, err = mkdir("` + filepath.Join(blocker, "subdir") + `")
		if ok ~= nil then
			error("Expected nil return for mkdir error")
		end
		if err == nil then
			error("Expected error message")
		end
	`)
	if err != nil {
		t.Fatalf("DoString failed: %v", err)
	}
}

func TestFileExistsAPI(t *testing.T) {
	tmpDir := t.TempDir()
	sb := newMockSandbox("test-pod", tmpDir)

	L := lua.NewState()
	defer L.Close()

	// Create test file
	existingFile := filepath.Join(tmpDir, "exists.txt")
	os.WriteFile(existingFile, []byte("test"), 0644)

	// Test existing file
	t.Run("existing file", func(t *testing.T) {
		fn := fileExistsAPI(sb)
		L.SetGlobal("file_exists", L.NewFunction(fn))

		err := L.DoString(`
			local exists = file_exists("` + existingFile + `")
			if not exists then
				error("Expected file to exist")
			end
		`)
		if err != nil {
			t.Fatalf("DoString failed: %v", err)
		}
	})

	// Test non-existing file
	t.Run("non-existing file", func(t *testing.T) {
		fn := fileExistsAPI(sb)
		L.SetGlobal("file_exists", L.NewFunction(fn))

		err := L.DoString(`
			local exists = file_exists("` + filepath.Join(tmpDir, "nonexistent.txt") + `")
			if exists then
				error("Expected file to not exist")
			end
		`)
		if err != nil {
			t.Fatalf("DoString failed: %v", err)
		}
	})

	// Test path outside sandbox (returns false, not error)
	t.Run("path outside sandbox returns false", func(t *testing.T) {
		fn := fileExistsAPI(sb)
		L.SetGlobal("file_exists", L.NewFunction(fn))

		err := L.DoString(`
			local exists = file_exists("/etc/passwd")
			if exists then
				error("Should return false for paths outside sandbox")
			end
		`)
		if err != nil {
			t.Fatalf("DoString failed: %v", err)
		}
	})
}
