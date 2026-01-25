package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anthropics/agentsmesh/runner/internal/mcp/tools"
)

func TestSearchChannels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("name") != "test" {
			t.Errorf("name param: got %v, want test", r.URL.Query().Get("name"))
		}

		resp := map[string]interface{}{
			"channels": []tools.Channel{
				{ID: 1, Name: "test-channel"},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewBackendClient(server.URL, "test-org", "test-pod")
	channels, err := client.SearchChannels(context.Background(), "test", nil, nil, nil, 0, 20)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(channels) != 1 {
		t.Errorf("channels count: got %v, want 1", len(channels))
	}
}

func TestCreateChannel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		if body["name"] != "new-channel" {
			t.Errorf("name: got %v, want new-channel", body["name"])
		}

		resp := map[string]interface{}{
			"channel": tools.Channel{
				ID:   1,
				Name: "new-channel",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewBackendClient(server.URL, "test-org", "test-pod")
	channel, err := client.CreateChannel(context.Background(), "new-channel", "description", nil, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if channel.Name != "new-channel" {
		t.Errorf("name: got %v, want new-channel", channel.Name)
	}
}

func TestGetChannel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"channel": tools.Channel{
				ID:   1,
				Name: "test-channel",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewBackendClient(server.URL, "test-org", "test-pod")
	channel, err := client.GetChannel(context.Background(), 1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if channel.ID != 1 {
		t.Errorf("ID: got %v, want 1", channel.ID)
	}
}

func TestSendMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		if body["content"] != "Hello" {
			t.Errorf("content: got %v, want Hello", body["content"])
		}

		resp := map[string]interface{}{
			"message": tools.ChannelMessage{
				ID:      1,
				Content: "Hello",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewBackendClient(server.URL, "test-org", "test-pod")
	msg, err := client.SendMessage(context.Background(), 1, "Hello", tools.ChannelMessageTypeText, nil, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if msg.Content != "Hello" {
		t.Errorf("content: got %v, want Hello", msg.Content)
	}
}

func TestGetMessages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"messages": []tools.ChannelMessage{
				{ID: 1, Content: "Message 1"},
				{ID: 2, Content: "Message 2"},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewBackendClient(server.URL, "test-org", "test-pod")
	messages, err := client.GetMessages(context.Background(), 1, nil, nil, nil, 50)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(messages) != 2 {
		t.Errorf("messages count: got %v, want 2", len(messages))
	}
}

func TestGetDocument(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"document": "test document content",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewBackendClient(server.URL, "test-org", "test-pod")
	doc, err := client.GetDocument(context.Background(), 1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if doc != "test document content" {
		t.Errorf("document: got %v, want test document content", doc)
	}
}

func TestUpdateDocument(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		if body["document"] != "updated content" {
			t.Errorf("document: got %v, want updated content", body["document"])
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewBackendClient(server.URL, "test-org", "test-pod")
	err := client.UpdateDocument(context.Background(), 1, "updated content")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSearchChannelsWithAllParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify all query params
		if r.URL.Query().Get("name") != "test" {
			t.Errorf("name param: got %v, want test", r.URL.Query().Get("name"))
		}
		if r.URL.Query().Get("repository_id") != "1" {
			t.Errorf("repository_id param: got %v, want 1", r.URL.Query().Get("repository_id"))
		}
		if r.URL.Query().Get("ticket_id") != "2" {
			t.Errorf("ticket_id param: got %v, want 2", r.URL.Query().Get("ticket_id"))
		}
		if r.URL.Query().Get("is_archived") != "true" {
			t.Errorf("is_archived param: got %v, want true", r.URL.Query().Get("is_archived"))
		}

		resp := map[string]interface{}{
			"channels": []tools.Channel{},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewBackendClient(server.URL, "test-org", "test-pod")
	repositoryID := 1
	ticketID := 2
	isArchived := true
	_, err := client.SearchChannels(context.Background(), "test", &repositoryID, &ticketID, &isArchived, 10, 20)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateChannelWithAllParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		if body["repository_id"] == nil {
			t.Error("repository_id should be set")
		}
		if body["ticket_id"] == nil {
			t.Error("ticket_id should be set")
		}

		resp := map[string]interface{}{
			"channel": tools.Channel{
				ID:   1,
				Name: "new-channel",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewBackendClient(server.URL, "test-org", "test-pod")
	repositoryID := 1
	ticketID := 2
	_, err := client.CreateChannel(context.Background(), "new-channel", "description", &repositoryID, &ticketID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSendMessageWithAllParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		if body["mentions"] == nil {
			t.Error("mentions should be set")
		}
		if body["reply_to"] == nil {
			t.Error("reply_to should be set")
		}

		resp := map[string]interface{}{
			"message": tools.ChannelMessage{
				ID:      1,
				Content: "Hello",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewBackendClient(server.URL, "test-org", "test-pod")
	replyTo := 5
	_, err := client.SendMessage(context.Background(), 1, "Hello", tools.ChannelMessageTypeText, []string{"pod-1", "pod-2"}, &replyTo)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetMessagesWithAllParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("before_time") != "2024-01-01T00:00:00Z" {
			t.Errorf("before_time param missing or wrong")
		}
		if r.URL.Query().Get("after_time") != "2024-01-02T00:00:00Z" {
			t.Errorf("after_time param missing or wrong")
		}
		if r.URL.Query().Get("mentioned_pod") != "pod-1" {
			t.Errorf("mentioned_pod param missing or wrong")
		}

		resp := map[string]interface{}{
			"messages": []tools.ChannelMessage{},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewBackendClient(server.URL, "test-org", "test-pod")
	beforeTime := "2024-01-01T00:00:00Z"
	afterTime := "2024-01-02T00:00:00Z"
	mentionedPod := "pod-1"
	_, err := client.GetMessages(context.Background(), 1, &beforeTime, &afterTime, &mentionedPod, 50)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
