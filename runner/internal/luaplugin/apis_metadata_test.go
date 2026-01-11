package luaplugin

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func TestSetMetadataAPI(t *testing.T) {
	tmpDir := t.TempDir()
	sb := newMockSandbox("test-pod", tmpDir)

	L := lua.NewState()
	defer L.Close()

	fn := setMetadataAPI(sb)
	L.SetGlobal("set_metadata", L.NewFunction(fn))

	// Test setting string metadata
	err := L.DoString(`set_metadata("key1", "string_value")`)
	if err != nil {
		t.Fatalf("DoString failed: %v", err)
	}

	// Test setting number metadata
	err = L.DoString(`set_metadata("key2", 42)`)
	if err != nil {
		t.Fatalf("DoString failed: %v", err)
	}

	// Test setting boolean metadata
	err = L.DoString(`set_metadata("key3", true)`)
	if err != nil {
		t.Fatalf("DoString failed: %v", err)
	}

	metadata := sb.GetMetadata()
	if metadata["key1"] != "string_value" {
		t.Errorf("Metadata key1 mismatch: got %v, want %q", metadata["key1"], "string_value")
	}
	if metadata["key2"] != float64(42) {
		t.Errorf("Metadata key2 mismatch: got %v, want %v", metadata["key2"], 42)
	}
	if metadata["key3"] != true {
		t.Errorf("Metadata key3 mismatch: got %v, want %v", metadata["key3"], true)
	}
}

func TestGetMetadataAPI(t *testing.T) {
	tmpDir := t.TempDir()
	sb := newMockSandbox("test-pod", tmpDir)

	// Pre-populate metadata
	sb.SetMetadata("existing_key", "existing_value")
	sb.SetMetadata("number_key", float64(123))

	L := lua.NewState()
	defer L.Close()

	fn := getMetadataAPI(sb)
	L.SetGlobal("get_metadata", L.NewFunction(fn))

	// Test getting existing string metadata
	t.Run("get existing string", func(t *testing.T) {
		err := L.DoString(`
			local value = get_metadata("existing_key")
			if value ~= "existing_value" then
				error("Expected 'existing_value', got: " .. tostring(value))
			end
		`)
		if err != nil {
			t.Fatalf("DoString failed: %v", err)
		}
	})

	// Test getting existing number metadata
	t.Run("get existing number", func(t *testing.T) {
		err := L.DoString(`
			local value = get_metadata("number_key")
			if value ~= 123 then
				error("Expected 123, got: " .. tostring(value))
			end
		`)
		if err != nil {
			t.Fatalf("DoString failed: %v", err)
		}
	})

	// Test getting non-existing metadata
	t.Run("get non-existing", func(t *testing.T) {
		err := L.DoString(`
			local value = get_metadata("nonexistent_key")
			if value ~= nil then
				error("Expected nil, got: " .. tostring(value))
			end
		`)
		if err != nil {
			t.Fatalf("DoString failed: %v", err)
		}
	})
}

func TestGetMetadataAPI_NilMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	sb := &mockSandbox{
		podKey:   "test-pod",
		rootPath: tmpDir,
		workDir:  tmpDir,
		metadata: nil, // Explicitly nil
	}

	L := lua.NewState()
	defer L.Close()

	fn := getMetadataAPI(sb)
	L.SetGlobal("get_metadata", L.NewFunction(fn))

	err := L.DoString(`
		local value = get_metadata("any_key")
		if value ~= nil then
			error("Expected nil for nil metadata")
		end
	`)
	if err != nil {
		t.Fatalf("DoString failed: %v", err)
	}
}

func TestAppendPromptAPI(t *testing.T) {
	tmpDir := t.TempDir()
	sb := newMockSandbox("test-pod", tmpDir)

	L := lua.NewState()
	defer L.Close()

	fn := appendPromptAPI(sb)
	L.SetGlobal("append_prompt", L.NewFunction(fn))

	// Test appending to empty prompt_suffix
	err := L.DoString(`append_prompt("\n\nultrathink")`)
	if err != nil {
		t.Fatalf("DoString failed: %v", err)
	}

	metadata := sb.GetMetadata()
	if metadata["prompt_suffix"] != "\n\nultrathink" {
		t.Errorf("prompt_suffix mismatch: got %q, want %q", metadata["prompt_suffix"], "\n\nultrathink")
	}

	// Test appending to existing prompt_suffix
	err = L.DoString(`append_prompt(" additional")`)
	if err != nil {
		t.Fatalf("DoString failed: %v", err)
	}

	metadata = sb.GetMetadata()
	if metadata["prompt_suffix"] != "\n\nultrathink additional" {
		t.Errorf("prompt_suffix mismatch: got %q, want %q", metadata["prompt_suffix"], "\n\nultrathink additional")
	}
}
