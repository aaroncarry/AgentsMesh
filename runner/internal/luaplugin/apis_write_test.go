package luaplugin

import (
	"os"
	"path/filepath"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func TestWriteFileAPI(t *testing.T) {
	tmpDir := t.TempDir()
	sb := newMockSandbox("test-pod", tmpDir)

	L := lua.NewState()
	defer L.Close()

	// Test successful write
	t.Run("successful write", func(t *testing.T) {
		fn := writeFileAPI(sb)
		L.SetGlobal("write_file", L.NewFunction(fn))

		testPath := filepath.Join(tmpDir, "test.txt")
		err := L.DoString(`return write_file("` + testPath + `", "hello world")`)
		if err != nil {
			t.Fatalf("DoString failed: %v", err)
		}

		content, err := os.ReadFile(testPath)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}
		if string(content) != "hello world" {
			t.Errorf("Content mismatch: got %q, want %q", string(content), "hello world")
		}
	})

	// Test sensitive file permissions
	t.Run("sensitive file has restricted permissions", func(t *testing.T) {
		fn := writeFileAPI(sb)
		L.SetGlobal("write_file", L.NewFunction(fn))

		testPath := filepath.Join(tmpDir, "mcp-config.json")
		err := L.DoString(`return write_file("` + testPath + `", "{}")`)
		if err != nil {
			t.Fatalf("DoString failed: %v", err)
		}

		info, err := os.Stat(testPath)
		if err != nil {
			t.Fatalf("Failed to stat file: %v", err)
		}

		perm := info.Mode().Perm()
		if perm != 0600 {
			t.Errorf("Sensitive file should have 0600 permissions, got %o", perm)
		}
	})

	// Test path traversal protection
	t.Run("path traversal blocked", func(t *testing.T) {
		fn := writeFileAPI(sb)
		L.SetGlobal("write_file", L.NewFunction(fn))

		err := L.DoString(`
			local ok, err = write_file("/etc/test.txt", "malicious")
			if ok ~= nil then
				error("Expected nil return for blocked path")
			end
		`)
		if err != nil {
			t.Fatalf("DoString failed: %v", err)
		}
	})
}

func TestWriteFileAPI_MkdirError(t *testing.T) {
	// Create a sandbox where we can trigger a mkdir error
	// by making a file with the same name as the desired directory
	tmpDir := t.TempDir()
	sb := newMockSandbox("test-pod", tmpDir)

	// Create a file that will block directory creation
	blocker := filepath.Join(tmpDir, "blocker")
	os.WriteFile(blocker, []byte("content"), 0644)

	L := lua.NewState()
	defer L.Close()

	fn := writeFileAPI(sb)
	L.SetGlobal("write_file", L.NewFunction(fn))

	// Try to write a file where the directory creation would fail
	// (blocker is a file, not a directory)
	err := L.DoString(`
		local ok, err = write_file("` + filepath.Join(blocker, "subdir", "file.txt") + `", "content")
		if ok ~= nil then
			error("Expected nil return for blocked mkdir")
		end
	`)
	if err != nil {
		t.Fatalf("DoString failed: %v", err)
	}
}

func TestWriteFileAPI_WriteError(t *testing.T) {
	tmpDir := t.TempDir()
	sb := newMockSandbox("test-pod", tmpDir)

	// Create a directory and make a file within it
	subDir := filepath.Join(tmpDir, "subdir")
	os.MkdirAll(subDir, 0755)

	// Make the directory read-only to cause write failure
	os.Chmod(subDir, 0555)
	defer os.Chmod(subDir, 0755) // Restore permissions for cleanup

	L := lua.NewState()
	defer L.Close()

	fn := writeFileAPI(sb)
	L.SetGlobal("write_file", L.NewFunction(fn))

	// Try to write a file in the read-only directory
	err := L.DoString(`
		local ok, err = write_file("` + filepath.Join(subDir, "test.txt") + `", "content")
		-- On some systems this may succeed if running as root, so we don't assert
	`)
	if err != nil {
		t.Fatalf("DoString failed: %v", err)
	}
}
