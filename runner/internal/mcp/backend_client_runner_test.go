package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anthropics/agentsmesh/runner/internal/mcp/tools"
)

func TestListRepositories(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method: got %v, want GET", r.Method)
		}

		expectedPath := "/api/v1/orgs/test-org/pod/repositories"
		if r.URL.Path != expectedPath {
			t.Errorf("path: got %v, want %v", r.URL.Path, expectedPath)
		}

		resp := map[string]interface{}{
			"repositories": []tools.Repository{
				{
					ID:            1,
					ProviderType:  "gitlab",
					Name:          "test-repo",
					FullPath:      "org/test-repo",
					DefaultBranch: "main",
					Visibility:    "private",
					IsActive:      true,
				},
				{
					ID:            2,
					ProviderType:  "github",
					Name:          "another-repo",
					FullPath:      "org/another-repo",
					DefaultBranch: "master",
					Visibility:    "public",
					IsActive:      true,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewBackendClient(server.URL, "test-org", "test-pod")
	repos, err := client.ListRepositories(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(repos) != 2 {
		t.Fatalf("repos count: got %v, want 2", len(repos))
	}

	repo := repos[0]
	if repo.ID != 1 {
		t.Errorf("repo ID: got %v, want 1", repo.ID)
	}
	if repo.Name != "test-repo" {
		t.Errorf("repo Name: got %v, want test-repo", repo.Name)
	}
	if repo.ProviderType != "gitlab" {
		t.Errorf("repo ProviderType: got %v, want gitlab", repo.ProviderType)
	}
	if repo.DefaultBranch != "main" {
		t.Errorf("repo DefaultBranch: got %v, want main", repo.DefaultBranch)
	}
}

func TestListRepositoriesEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"repositories": []tools.Repository{},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewBackendClient(server.URL, "test-org", "test-pod")
	repos, err := client.ListRepositories(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(repos) != 0 {
		t.Errorf("repos count: got %v, want 0", len(repos))
	}
}

func TestListRunners(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method: got %v, want GET", r.Method)
		}

		expectedPath := "/api/v1/orgs/test-org/pod/runners"
		if r.URL.Path != expectedPath {
			t.Errorf("path: got %v, want %v", r.URL.Path, expectedPath)
		}

		resp := map[string]interface{}{
			"runners": []tools.RunnerSummary{
				{
					ID:                1,
					NodeID:            "dev-machine",
					Description:       "Development runner",
					Status:            "online",
					CurrentPods:       2,
					MaxConcurrentPods: 5,
					AvailableAgents: []tools.AgentTypeSummary{
						{
							ID:          1,
							Slug:        "claude-code",
							Name:        "Claude Code",
							Description: "AI coding assistant",
							Config: []tools.ConfigFieldSummary{
								{
									Name:     "model",
									Type:     "select",
									Default:  "claude-sonnet-4-20250514",
									Options:  []string{"claude-sonnet-4-20250514", "claude-opus-4-20250514"},
									Required: true,
								},
							},
							UserConfig: map[string]interface{}{
								"model": "claude-opus-4-20250514",
							},
						},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewBackendClient(server.URL, "test-org", "test-pod")
	runners, err := client.ListRunners(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(runners) != 1 {
		t.Fatalf("runners count: got %v, want 1", len(runners))
	}

	runner := runners[0]
	if runner.ID != 1 {
		t.Errorf("runner ID: got %v, want 1", runner.ID)
	}
	if runner.NodeID != "dev-machine" {
		t.Errorf("runner NodeID: got %v, want dev-machine", runner.NodeID)
	}
	if runner.Status != "online" {
		t.Errorf("runner Status: got %v, want online", runner.Status)
	}
	if runner.CurrentPods != 2 {
		t.Errorf("runner CurrentPods: got %v, want 2", runner.CurrentPods)
	}
	if runner.MaxConcurrentPods != 5 {
		t.Errorf("runner MaxConcurrentPods: got %v, want 5", runner.MaxConcurrentPods)
	}

	// Verify available agents
	if len(runner.AvailableAgents) != 1 {
		t.Fatalf("available agents count: got %v, want 1", len(runner.AvailableAgents))
	}

	agent := runner.AvailableAgents[0]
	if agent.ID != 1 {
		t.Errorf("agent ID: got %v, want 1", agent.ID)
	}
	if agent.Slug != "claude-code" {
		t.Errorf("agent Slug: got %v, want claude-code", agent.Slug)
	}
	if agent.Name != "Claude Code" {
		t.Errorf("agent Name: got %v, want Claude Code", agent.Name)
	}

	// Verify config fields
	if len(agent.Config) != 1 {
		t.Fatalf("config fields count: got %v, want 1", len(agent.Config))
	}

	config := agent.Config[0]
	if config.Name != "model" {
		t.Errorf("config Name: got %v, want model", config.Name)
	}
	if config.Type != "select" {
		t.Errorf("config Type: got %v, want select", config.Type)
	}
	if !config.Required {
		t.Error("config Required: expected true")
	}
	if len(config.Options) != 2 {
		t.Errorf("config Options count: got %v, want 2", len(config.Options))
	}

	// Verify user config
	if agent.UserConfig == nil {
		t.Fatal("user config should not be nil")
	}
	if agent.UserConfig["model"] != "claude-opus-4-20250514" {
		t.Errorf("user config model: got %v, want claude-opus-4-20250514", agent.UserConfig["model"])
	}
}

func TestListRunnersEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"runners": []tools.RunnerSummary{},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewBackendClient(server.URL, "test-org", "test-pod")
	runners, err := client.ListRunners(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(runners) != 0 {
		t.Errorf("runners count: got %v, want 0", len(runners))
	}
}

func TestListRunnersMultipleAgents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"runners": []tools.RunnerSummary{
				{
					ID:                1,
					NodeID:            "multi-agent-runner",
					Status:            "online",
					CurrentPods:       0,
					MaxConcurrentPods: 10,
					AvailableAgents: []tools.AgentTypeSummary{
						{ID: 1, Slug: "claude-code", Name: "Claude Code"},
						{ID: 2, Slug: "aider", Name: "Aider"},
						{ID: 3, Slug: "codex", Name: "Codex CLI"},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewBackendClient(server.URL, "test-org", "test-pod")
	runners, err := client.ListRunners(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(runners) != 1 {
		t.Fatalf("runners count: got %v, want 1", len(runners))
	}

	if len(runners[0].AvailableAgents) != 3 {
		t.Errorf("available agents count: got %v, want 3", len(runners[0].AvailableAgents))
	}

	// Verify agent slugs
	expectedSlugs := []string{"claude-code", "aider", "codex"}
	for i, agent := range runners[0].AvailableAgents {
		if agent.Slug != expectedSlugs[i] {
			t.Errorf("agent[%d] Slug: got %v, want %v", i, agent.Slug, expectedSlugs[i])
		}
	}
}
