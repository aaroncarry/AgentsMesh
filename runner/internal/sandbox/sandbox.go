package sandbox

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Sandbox represents a pod's isolated runtime environment.
type Sandbox struct {
	// Pod identification
	PodKey   string `json:"pod_key"`
	RootPath string `json:"root_path"` // Sandbox root directory

	// Outputs filled by plugin chain
	WorkDir    string            `json:"work_dir"`    // Final working directory
	EnvVars    map[string]string `json:"env_vars"`    // Environment variables
	LaunchArgs []string          `json:"launch_args"` // Additional launch arguments (e.g., --mcp-config)

	// Metadata from plugins
	Metadata map[string]interface{} `json:"metadata"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Internal state (not serialized)
	plugins []Plugin   `json:"-"` // Applied plugins (for Teardown)
	mu      sync.Mutex `json:"-"` // Protects concurrent access to maps
}

// NewSandbox creates a new Sandbox instance.
func NewSandbox(podKey, rootPath string) *Sandbox {
	now := time.Now()
	return &Sandbox{
		PodKey: podKey,
		RootPath:   rootPath,
		EnvVars:    make(map[string]string),
		LaunchArgs: make([]string, 0),
		Metadata:   make(map[string]interface{}),
		CreatedAt:  now,
		UpdatedAt:  now,
		plugins:    make([]Plugin, 0),
	}
}

// AddPlugin records a plugin that was applied (for Teardown).
func (s *Sandbox) AddPlugin(p Plugin) {
	s.plugins = append(s.plugins, p)
}

// GetPlugins returns the list of applied plugins.
func (s *Sandbox) GetPlugins() []Plugin {
	return s.plugins
}

// Save persists the sandbox metadata to sandbox.json.
func (s *Sandbox) Save() error {
	s.UpdatedAt = time.Now()

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	metaPath := filepath.Join(s.RootPath, "sandbox.json")
	return os.WriteFile(metaPath, data, 0644)
}

// Load reads sandbox metadata from sandbox.json.
func (s *Sandbox) Load() error {
	metaPath := filepath.Join(s.RootPath, "sandbox.json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, s)
}

// GetLogsDir returns the path to the logs directory.
func (s *Sandbox) GetLogsDir() string {
	return filepath.Join(s.RootPath, "logs")
}

// EnsureLogsDir creates the logs directory if it doesn't exist.
func (s *Sandbox) EnsureLogsDir() error {
	return os.MkdirAll(s.GetLogsDir(), 0755)
}

// The following methods implement the luaplugin.SandboxAdapter interface
// to allow Lua plugins to interact with the sandbox without import cycles.

// GetPodKey returns the pod key.
func (s *Sandbox) GetPodKey() string {
	return s.PodKey
}

// GetRootPath returns the sandbox root path.
func (s *Sandbox) GetRootPath() string {
	return s.RootPath
}

// GetWorkDir returns the working directory.
func (s *Sandbox) GetWorkDir() string {
	return s.WorkDir
}

// GetLaunchArgs returns a copy of the launch arguments.
// Thread-safe: protected by mutex for concurrent access.
func (s *Sandbox) GetLaunchArgs() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.LaunchArgs == nil {
		return nil
	}
	// Return a copy to prevent concurrent modification
	result := make([]string, len(s.LaunchArgs))
	copy(result, s.LaunchArgs)
	return result
}

// SetLaunchArgs sets the launch arguments.
// Thread-safe: protected by mutex for concurrent access.
func (s *Sandbox) SetLaunchArgs(args []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LaunchArgs = args
}

// AppendLaunchArgs appends arguments to the launch args atomically.
// Thread-safe: this is the preferred method for adding arguments.
func (s *Sandbox) AppendLaunchArgs(args ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LaunchArgs = append(s.LaunchArgs, args...)
}

// GetEnvVars returns a copy of the environment variables.
// Thread-safe: protected by mutex for concurrent access.
func (s *Sandbox) GetEnvVars() map[string]string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.EnvVars == nil {
		return nil
	}
	// Return a copy to prevent concurrent modification
	result := make(map[string]string, len(s.EnvVars))
	for k, v := range s.EnvVars {
		result[k] = v
	}
	return result
}

// SetEnvVar sets an environment variable.
// Thread-safe: protected by mutex for concurrent Lua plugin execution.
func (s *Sandbox) SetEnvVar(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.EnvVars == nil {
		s.EnvVars = make(map[string]string)
	}
	s.EnvVars[key] = value
}

// GetMetadata returns a copy of the metadata map.
// Thread-safe: protected by mutex for concurrent access.
// Note: nested objects are not deep-copied.
func (s *Sandbox) GetMetadata() map[string]interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Metadata == nil {
		return nil
	}
	// Return a shallow copy to prevent concurrent modification
	result := make(map[string]interface{}, len(s.Metadata))
	for k, v := range s.Metadata {
		result[k] = v
	}
	return result
}

// SetMetadata sets a metadata value.
// Thread-safe: protected by mutex for concurrent Lua plugin execution.
func (s *Sandbox) SetMetadata(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Metadata == nil {
		s.Metadata = make(map[string]interface{})
	}
	s.Metadata[key] = value
}
