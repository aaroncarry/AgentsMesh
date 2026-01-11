package luaplugin

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func TestAddArgsAPI(t *testing.T) {
	tmpDir := t.TempDir()
	sb := newMockSandbox("test-pod", tmpDir)

	L := lua.NewState()
	defer L.Close()

	fn := addArgsAPI(sb)
	L.SetGlobal("add_args", L.NewFunction(fn))

	// Test single arg
	err := L.DoString(`add_args("--flag")`)
	if err != nil {
		t.Fatalf("DoString failed: %v", err)
	}

	// Test multiple args
	err = L.DoString(`add_args("--key", "value")`)
	if err != nil {
		t.Fatalf("DoString failed: %v", err)
	}

	args := sb.GetLaunchArgs()
	expected := []string{"--flag", "--key", "value"}
	if len(args) != len(expected) {
		t.Errorf("Args length mismatch: got %d, want %d", len(args), len(expected))
	}
	for i, arg := range expected {
		if args[i] != arg {
			t.Errorf("Arg %d mismatch: got %q, want %q", i, args[i], arg)
		}
	}
}
