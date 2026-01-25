package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anthropics/agentsmesh/runner/internal/mcp/tools"
)

func TestListAvailablePods(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"pods": []tools.AvailablePod{
				{PodKey: "pod-1", Status: tools.PodStatusRunning},
				{PodKey: "pod-2", Status: tools.PodStatusRunning},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewBackendClient(server.URL, "test-org", "test-pod")
	pods, err := client.ListAvailablePods(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(pods) != 2 {
		t.Errorf("pods count: got %v, want 2", len(pods))
	}
}

func TestCreatePod(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body tools.PodCreateRequest
		json.NewDecoder(r.Body).Decode(&body)

		if body.InitialPrompt != "Hello" {
			t.Errorf("initial_prompt: got %v, want Hello", body.InitialPrompt)
		}

		resp := map[string]interface{}{
			"pod": map[string]interface{}{
				"pod_key": "new-pod",
				"status":  "created",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewBackendClient(server.URL, "test-org", "test-pod")
	resp, err := client.CreatePod(context.Background(), &tools.PodCreateRequest{
		InitialPrompt: "Hello",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.PodKey != "new-pod" {
		t.Errorf("pod_key: got %v, want new-pod", resp.PodKey)
	}
}
