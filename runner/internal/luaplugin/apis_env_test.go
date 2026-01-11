package luaplugin

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func TestAddEnvAPI(t *testing.T) {
	tmpDir := t.TempDir()
	sb := newMockSandbox("test-pod", tmpDir)

	L := lua.NewState()
	defer L.Close()

	fn := addEnvAPI(sb)
	L.SetGlobal("add_env", L.NewFunction(fn))

	err := L.DoString(`add_env("MY_VAR", "my_value")`)
	if err != nil {
		t.Fatalf("DoString failed: %v", err)
	}

	envVars := sb.GetEnvVars()
	if envVars["MY_VAR"] != "my_value" {
		t.Errorf("Env var mismatch: got %q, want %q", envVars["MY_VAR"], "my_value")
	}
}
