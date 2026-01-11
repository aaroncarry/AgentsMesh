package luaplugin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

// validatePath checks if the given path is within the sandbox boundaries.
// It allows paths within the sandbox root path or work directory.
// Returns the cleaned absolute path if valid, or an error if path traversal is detected.
func validatePath(sb SandboxAdapter, path string) (string, error) {
	// Clean and resolve to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}
	absPath = filepath.Clean(absPath)

	// Get sandbox boundaries
	rootPath := filepath.Clean(sb.GetRootPath())
	workDir := filepath.Clean(sb.GetWorkDir())

	// Check if path is within root_path or work_dir
	// Use HasPrefix with path separator to avoid matching partial directory names
	// e.g., /sandbox-test should not match /sandbox
	inRootPath := strings.HasPrefix(absPath, rootPath+string(filepath.Separator)) || absPath == rootPath
	inWorkDir := strings.HasPrefix(absPath, workDir+string(filepath.Separator)) || absPath == workDir

	if !inRootPath && !inWorkDir {
		return "", fmt.Errorf("path outside sandbox boundaries: %s (root=%s, workdir=%s)", path, rootPath, workDir)
	}

	return absPath, nil
}

// isSensitiveFile checks if the file path indicates a sensitive configuration file
// that should have restricted permissions.
func isSensitiveFile(path string) bool {
	base := filepath.Base(path)
	// MCP config files and other sensitive configurations
	sensitivePatterns := []string{
		"mcp-config.json",
		"settings.json",  // Gemini settings
		"opencode.json",  // OpenCode config
		"credentials",
		".env",
	}
	for _, pattern := range sensitivePatterns {
		if base == pattern {
			return true
		}
	}
	return false
}

// writeFileAPI creates a Lua function that writes content to a file.
// Usage: ctx.write_file(path, content)
// Security: Path must be within sandbox boundaries (root_path or work_dir).
// Sensitive files (MCP configs, credentials) are written with 0600 permissions.
func writeFileAPI(sb SandboxAdapter) lua.LGFunction {
	return func(L *lua.LState) int {
		path := L.CheckString(1)
		content := L.CheckString(2)

		// Validate path is within sandbox boundaries
		validPath, err := validatePath(sb, path)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		// Ensure directory exists
		dir := filepath.Dir(validPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		// Use restricted permissions for sensitive files
		perm := os.FileMode(0644)
		if isSensitiveFile(validPath) {
			perm = 0600
		}

		if err := os.WriteFile(validPath, []byte(content), perm); err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LBool(true))
		return 1
	}
}

// readFileAPI creates a Lua function that reads file content.
// Usage: content = ctx.read_file(path)
// Security: Path must be within sandbox boundaries (root_path or work_dir).
func readFileAPI(sb SandboxAdapter) lua.LGFunction {
	return func(L *lua.LState) int {
		path := L.CheckString(1)

		// Validate path is within sandbox boundaries
		validPath, err := validatePath(sb, path)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		data, err := os.ReadFile(validPath)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LString(string(data)))
		return 1
	}
}

// readJSONAPI creates a Lua function that reads and parses a JSON file.
// Usage: table = ctx.read_json(path)
// Security: Path must be within sandbox boundaries (root_path or work_dir).
func readJSONAPI(sb SandboxAdapter) lua.LGFunction {
	return func(L *lua.LState) int {
		path := L.CheckString(1)

		// Validate path is within sandbox boundaries
		validPath, err := validatePath(sb, path)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		data, err := os.ReadFile(validPath)
		if err != nil {
			// File doesn't exist - return nil (not an error for read_json)
			L.Push(lua.LNil)
			return 1
		}

		var result map[string]interface{}
		if err := json.Unmarshal(data, &result); err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(goValueToLua(L, result))
		return 1
	}
}

// mkdirAPI creates a Lua function that creates a directory.
// Usage: ctx.mkdir(path)
// Security: Path must be within sandbox boundaries (root_path or work_dir).
func mkdirAPI(sb SandboxAdapter) lua.LGFunction {
	return func(L *lua.LState) int {
		path := L.CheckString(1)

		// Validate path is within sandbox boundaries
		validPath, err := validatePath(sb, path)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		if err := os.MkdirAll(validPath, 0755); err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LBool(true))
		return 1
	}
}

// fileExistsAPI creates a Lua function that checks if a file exists.
// Usage: exists = ctx.file_exists(path)
// Security: Path must be within sandbox boundaries (root_path or work_dir).
func fileExistsAPI(sb SandboxAdapter) lua.LGFunction {
	return func(L *lua.LState) int {
		path := L.CheckString(1)

		// Validate path is within sandbox boundaries
		validPath, err := validatePath(sb, path)
		if err != nil {
			// Path outside boundaries - return false (file doesn't exist from sandbox's perspective)
			L.Push(lua.LBool(false))
			return 1
		}

		_, err = os.Stat(validPath)
		L.Push(lua.LBool(err == nil))
		return 1
	}
}
