package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anthropics/agentsmesh/runner/internal/mcp/tools"
)

func TestObserveTerminal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method: got %v, want GET", r.Method)
		}

		if r.Header.Get("X-Pod-Key") != "test-pod" {
			t.Errorf("X-Pod-Key: got %v, want test-pod", r.Header.Get("X-Pod-Key"))
		}

		resp := tools.TerminalOutput{
			PodKey:     "target-pod",
			Output:     "test output",
			CursorX:    10,
			CursorY:    5,
			TotalLines: 100,
			HasMore:    true,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewBackendClient(server.URL, "test-org", "test-pod")
	result, err := client.ObserveTerminal(context.Background(), "target-pod", 50, false, true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Output != "test output" {
		t.Errorf("Output: got %v, want test output", result.Output)
	}

	if result.CursorX != 10 {
		t.Errorf("CursorX: got %v, want 10", result.CursorX)
	}
}

func TestSendTerminalText(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: got %v, want POST", r.Method)
		}

		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		if body["input"] != "hello world" {
			t.Errorf("input: got %v, want hello world", body["input"])
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewBackendClient(server.URL, "test-org", "test-pod")
	err := client.SendTerminalText(context.Background(), "target-pod", "hello world")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSendTerminalKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		// The client converts keys to escape sequences and sends them as "input"
		input, ok := body["input"].(string)
		if !ok || input == "" {
			t.Errorf("input: got %v, want non-empty string", body["input"])
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewBackendClient(server.URL, "test-org", "test-pod")
	err := client.SendTerminalKey(context.Background(), "target-pod", []string{"ctrl+c", "enter"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSendTerminalKeyAllKeys(t *testing.T) {
	// Test all supported special keys
	allKeys := []string{
		"enter", "escape", "tab", "backspace", "delete",
		"ctrl+c", "ctrl+d", "ctrl+u", "ctrl+l", "ctrl+z",
		"ctrl+a", "ctrl+e", "ctrl+k", "ctrl+w",
		"up", "down", "left", "right",
		"home", "end", "pageup", "pagedown",
		"shift+tab",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		input, ok := body["input"].(string)
		if !ok || input == "" {
			t.Errorf("input: got %v, want non-empty string", body["input"])
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewBackendClient(server.URL, "test-org", "test-pod")
	err := client.SendTerminalKey(context.Background(), "target-pod", allKeys)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSendTerminalKeySingleChar(t *testing.T) {
	// Test single character key
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		input, ok := body["input"].(string)
		if !ok || input != "a" {
			t.Errorf("input: got %v, want 'a'", body["input"])
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewBackendClient(server.URL, "test-org", "test-pod")
	err := client.SendTerminalKey(context.Background(), "target-pod", []string{"a"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSendTerminalKeyUnknownKey(t *testing.T) {
	// Test unknown multi-character key (should be ignored)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		// Unknown keys should produce empty string
		input, ok := body["input"].(string)
		if !ok {
			t.Errorf("input should be string")
		}
		// "unknown_key" is multi-char and not in the switch, so should be empty
		if input != "" {
			t.Errorf("input: got %v, want empty for unknown key", input)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewBackendClient(server.URL, "test-org", "test-pod")
	err := client.SendTerminalKey(context.Background(), "target-pod", []string{"unknown_key"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
