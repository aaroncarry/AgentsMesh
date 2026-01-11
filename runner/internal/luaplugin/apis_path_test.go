package luaplugin

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidatePath(t *testing.T) {
	tmpDir := t.TempDir()
	sb := newMockSandbox("test-pod", tmpDir)
	sb.workDir = filepath.Join(tmpDir, "workdir")
	os.MkdirAll(sb.workDir, 0755)

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid path within root",
			path:    filepath.Join(tmpDir, "test.txt"),
			wantErr: false,
		},
		{
			name:    "valid path within workdir",
			path:    filepath.Join(sb.workDir, "test.txt"),
			wantErr: false,
		},
		{
			name:    "path traversal attempt",
			path:    filepath.Join(tmpDir, "..", "outside.txt"),
			wantErr: true,
		},
		{
			name:    "absolute path outside sandbox",
			path:    "/etc/passwd",
			wantErr: true,
		},
		{
			name:    "root path itself",
			path:    tmpDir,
			wantErr: false,
		},
		{
			name:    "work dir itself",
			path:    sb.workDir,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validatePath(sb, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsSensitiveFile(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"mcp-config.json", true},
		{"/path/to/mcp-config.json", true},
		{"settings.json", true},
		{"/home/user/.gemini/settings.json", true},
		{"opencode.json", true},
		{"credentials", true},
		{".env", true},
		{"normal.txt", false},
		{"config.yaml", false},
		{"my-settings.json", false}, // Not exact match
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := isSensitiveFile(tt.path)
			if result != tt.expected {
				t.Errorf("isSensitiveFile(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}
