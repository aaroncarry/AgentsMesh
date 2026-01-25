package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anthropics/agentsmesh/runner/internal/mcp/tools"
)

func TestRequestBinding(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		if body["target_pod"] != "target-pod" {
			t.Errorf("target_pod: got %v, want target-pod", body["target_pod"])
		}

		resp := map[string]interface{}{
			"binding": tools.Binding{
				ID:           1,
				InitiatorPod: "test-pod",
				TargetPod:    "target-pod",
				Status:       tools.BindingStatusPending,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewBackendClient(server.URL, "test-org", "test-pod")
	binding, err := client.RequestBinding(context.Background(), "target-pod", []tools.BindingScope{tools.ScopeTerminalRead})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if binding.Status != tools.BindingStatusPending {
		t.Errorf("status: got %v, want pending", binding.Status)
	}
}

func TestAcceptBinding(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		if body["binding_id"].(float64) != 1 {
			t.Errorf("binding_id: got %v, want 1", body["binding_id"])
		}

		resp := map[string]interface{}{
			"binding": tools.Binding{
				ID:     1,
				Status: tools.BindingStatusActive,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewBackendClient(server.URL, "test-org", "test-pod")
	binding, err := client.AcceptBinding(context.Background(), 1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if binding.Status != tools.BindingStatusActive {
		t.Errorf("status: got %v, want active", binding.Status)
	}
}

func TestRejectBinding(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		if body["reason"] != "not allowed" {
			t.Errorf("reason: got %v, want not allowed", body["reason"])
		}

		resp := map[string]interface{}{
			"binding": tools.Binding{
				ID:     1,
				Status: tools.BindingStatusRejected,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewBackendClient(server.URL, "test-org", "test-pod")
	binding, err := client.RejectBinding(context.Background(), 1, "not allowed")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if binding.Status != tools.BindingStatusRejected {
		t.Errorf("status: got %v, want rejected", binding.Status)
	}
}

func TestUnbindPod(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewBackendClient(server.URL, "test-org", "test-pod")
	err := client.UnbindPod(context.Background(), "target-pod")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetBindings(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"bindings": []tools.Binding{
				{ID: 1, Status: tools.BindingStatusActive},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewBackendClient(server.URL, "test-org", "test-pod")
	bindings, err := client.GetBindings(context.Background(), nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(bindings) != 1 {
		t.Errorf("bindings count: got %v, want 1", len(bindings))
	}
}

func TestGetBindingsWithStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("status") != "active" {
			t.Errorf("status param: got %v, want active", r.URL.Query().Get("status"))
		}

		resp := map[string]interface{}{
			"bindings": []tools.Binding{},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewBackendClient(server.URL, "test-org", "test-pod")
	status := tools.BindingStatusActive
	_, err := client.GetBindings(context.Background(), &status)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetBoundPods(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Backend returns pod keys as strings, not AvailablePod objects
		resp := map[string]interface{}{
			"pods":  []string{"bound-pod-1", "bound-pod-2"},
			"count": 2,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewBackendClient(server.URL, "test-org", "test-pod")
	pods, err := client.GetBoundPods(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(pods) != 2 {
		t.Errorf("pods count: got %v, want 2", len(pods))
	}

	if pods[0] != "bound-pod-1" {
		t.Errorf("first pod: got %v, want bound-pod-1", pods[0])
	}
}
