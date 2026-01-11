package luaplugin

// mockSandbox implements SandboxAdapter for testing
type mockSandbox struct {
	podKey     string
	rootPath   string
	workDir    string
	launchArgs []string
	envVars    map[string]string
	metadata   map[string]interface{}
}

func newMockSandbox(podKey, rootPath string) *mockSandbox {
	return &mockSandbox{
		podKey:     podKey,
		rootPath:   rootPath,
		workDir:    rootPath,
		launchArgs: make([]string, 0),
		envVars:    make(map[string]string),
		metadata:   make(map[string]interface{}),
	}
}

func (m *mockSandbox) GetPodKey() string                         { return m.podKey }
func (m *mockSandbox) GetRootPath() string                       { return m.rootPath }
func (m *mockSandbox) GetWorkDir() string                        { return m.workDir }
func (m *mockSandbox) GetLaunchArgs() []string                   { return m.launchArgs }
func (m *mockSandbox) SetLaunchArgs(args []string)               { m.launchArgs = args }
func (m *mockSandbox) AppendLaunchArgs(args ...string)           { m.launchArgs = append(m.launchArgs, args...) }
func (m *mockSandbox) GetEnvVars() map[string]string             { return m.envVars }
func (m *mockSandbox) SetEnvVar(key, value string)               { m.envVars[key] = value }
func (m *mockSandbox) GetMetadata() map[string]interface{}       { return m.metadata }
func (m *mockSandbox) SetMetadata(key string, value interface{}) { m.metadata[key] = value }

// errorMockLoader returns an error from LoadBuiltinPlugins
type errorMockLoader struct{}

func (m *errorMockLoader) LoadBuiltinPlugins() ([]*LuaPlugin, error) {
	return nil, nil // No error, just empty
}

func (m *errorMockLoader) LoadUserPlugins(dir string, loadedNames map[string]bool) ([]*LuaPlugin, error) {
	return nil, nil // No error, just empty
}
