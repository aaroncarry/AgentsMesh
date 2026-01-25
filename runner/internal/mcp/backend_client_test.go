package mcp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anthropics/agentsmesh/runner/internal/mcp/tools"
)

func TestNewBackendClient(t *testing.T) {
	client := NewBackendClient("http://localhost:8080", "test-org", "test-pod")

	if client == nil {
		t.Fatal("NewBackendClient returned nil")
	}

	if client.baseURL != "http://localhost:8080" {
		t.Errorf("baseURL: got %v, want %v", client.baseURL, "http://localhost:8080")
	}

	if client.orgSlug != "test-org" {
		t.Errorf("orgSlug: got %v, want %v", client.orgSlug, "test-org")
	}

	if client.podKey != "test-pod" {
		t.Errorf("podKey: got %v, want %v", client.podKey, "test-pod")
	}

	if client.httpClient == nil {
		t.Error("httpClient should not be nil")
	}
}

func TestSetPodKey(t *testing.T) {
	client := NewBackendClient("http://localhost:8080", "test-org", "old-pod")
	client.SetPodKey("new-pod")

	if client.podKey != "new-pod" {
		t.Errorf("podKey: got %v, want %v", client.podKey, "new-pod")
	}
}

func TestGetPodKey(t *testing.T) {
	client := NewBackendClient("http://localhost:8080", "test-org", "test-pod")

	if client.GetPodKey() != "test-pod" {
		t.Errorf("GetPodKey: got %v, want %v", client.GetPodKey(), "test-pod")
	}
}

func TestSetOrgSlug(t *testing.T) {
	client := NewBackendClient("http://localhost:8080", "old-org", "test-pod")
	client.SetOrgSlug("new-org")

	if client.orgSlug != "new-org" {
		t.Errorf("orgSlug: got %v, want %v", client.orgSlug, "new-org")
	}
}

func TestGetOrgSlug(t *testing.T) {
	client := NewBackendClient("http://localhost:8080", "test-org", "test-pod")

	if client.GetOrgSlug() != "test-org" {
		t.Errorf("GetOrgSlug: got %v, want %v", client.GetOrgSlug(), "test-org")
	}
}

func TestRequestError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer server.Close()

	client := NewBackendClient(server.URL, "test-org", "test-pod")
	_, err := client.GetTicket(context.Background(), "AM-1")

	if err == nil {
		t.Error("expected error for 500 response")
	}
}

func TestBackendClientImplementsInterface(t *testing.T) {
	var _ tools.CollaborationClient = (*BackendClient)(nil)
}
