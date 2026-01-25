package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anthropics/agentsmesh/runner/internal/mcp/tools"
)

func TestSearchTickets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"tickets": []tools.Ticket{
				{ID: 1, Identifier: "AM-1", Title: "Test Ticket"},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewBackendClient(server.URL, "test-org", "test-pod")
	tickets, err := client.SearchTickets(context.Background(), nil, nil, nil, nil, nil, nil, "", 20, 1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tickets) != 1 {
		t.Errorf("tickets count: got %v, want 1", len(tickets))
	}
}

func TestGetTicket(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"ticket": tools.Ticket{
				ID:         1,
				Identifier: "AM-123",
				Title:      "Test Ticket",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewBackendClient(server.URL, "test-org", "test-pod")
	ticket, err := client.GetTicket(context.Background(), "AM-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ticket.Identifier != "AM-123" {
		t.Errorf("identifier: got %v, want AM-123", ticket.Identifier)
	}
}

func TestCreateTicket(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		if body["title"] != "New Ticket" {
			t.Errorf("title: got %v, want New Ticket", body["title"])
		}

		resp := map[string]interface{}{
			"ticket": tools.Ticket{
				ID:         1,
				Identifier: "AM-1",
				Title:      "New Ticket",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewBackendClient(server.URL, "test-org", "test-pod")
	// Test without repository_id (nil)
	ticket, err := client.CreateTicket(context.Background(), nil, "New Ticket", "Description", tools.TicketTypeTask, tools.TicketPriorityMedium, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ticket.Title != "New Ticket" {
		t.Errorf("title: got %v, want New Ticket", ticket.Title)
	}
}

func TestCreateTicketWithRepositoryID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		if body["title"] != "New Ticket" {
			t.Errorf("title: got %v, want New Ticket", body["title"])
		}
		// Check repository_id is passed
		if body["repository_id"] == nil {
			t.Error("repository_id should be set")
		} else if int64(body["repository_id"].(float64)) != 123 {
			t.Errorf("repository_id: got %v, want 123", body["repository_id"])
		}

		resp := map[string]interface{}{
			"ticket": tools.Ticket{
				ID:         1,
				Identifier: "AM-1",
				Title:      "New Ticket",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewBackendClient(server.URL, "test-org", "test-pod")
	repoID := int64(123)
	ticket, err := client.CreateTicket(context.Background(), &repoID, "New Ticket", "Description", tools.TicketTypeTask, tools.TicketPriorityMedium, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ticket.Title != "New Ticket" {
		t.Errorf("title: got %v, want New Ticket", ticket.Title)
	}
}

func TestUpdateTicket(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		if body["title"] != "Updated Title" {
			t.Errorf("title: got %v, want Updated Title", body["title"])
		}

		resp := map[string]interface{}{
			"ticket": tools.Ticket{
				ID:    1,
				Title: "Updated Title",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewBackendClient(server.URL, "test-org", "test-pod")
	title := "Updated Title"
	ticket, err := client.UpdateTicket(context.Background(), "AM-1", &title, nil, nil, nil, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ticket.Title != "Updated Title" {
		t.Errorf("title: got %v, want Updated Title", ticket.Title)
	}
}

func TestUpdateTicketWithAllParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		if body["title"] != "Updated Title" {
			t.Errorf("title: got %v, want Updated Title", body["title"])
		}
		if body["description"] != "Updated Description" {
			t.Errorf("description: got %v, want Updated Description", body["description"])
		}
		if body["status"] != "done" {
			t.Errorf("status: got %v, want done", body["status"])
		}
		if body["priority"] != "high" {
			t.Errorf("priority: got %v, want high", body["priority"])
		}
		if body["type"] != "bug" {
			t.Errorf("type: got %v, want bug", body["type"])
		}

		resp := map[string]interface{}{
			"ticket": tools.Ticket{
				ID:    1,
				Title: "Updated Title",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewBackendClient(server.URL, "test-org", "test-pod")
	title := "Updated Title"
	description := "Updated Description"
	status := tools.TicketStatusDone
	priority := tools.TicketPriorityHigh
	ticketType := tools.TicketTypeBug
	_, err := client.UpdateTicket(context.Background(), "AM-1", &title, &description, &status, &priority, &ticketType)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSearchTicketsWithAllParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("repository_id") != "1" {
			t.Errorf("repository_id param: got %v, want 1", r.URL.Query().Get("repository_id"))
		}
		if r.URL.Query().Get("status") != "todo" {
			t.Errorf("status param: got %v, want todo", r.URL.Query().Get("status"))
		}
		if r.URL.Query().Get("type") != "task" {
			t.Errorf("type param: got %v, want task", r.URL.Query().Get("type"))
		}
		if r.URL.Query().Get("priority") != "high" {
			t.Errorf("priority param: got %v, want high", r.URL.Query().Get("priority"))
		}
		if r.URL.Query().Get("assignee_id") != "2" {
			t.Errorf("assignee_id param: got %v, want 2", r.URL.Query().Get("assignee_id"))
		}
		if r.URL.Query().Get("parent_id") != "3" {
			t.Errorf("parent_id param: got %v, want 3", r.URL.Query().Get("parent_id"))
		}
		if r.URL.Query().Get("query") != "test query" {
			t.Errorf("query param: got %v, want 'test query'", r.URL.Query().Get("query"))
		}

		resp := map[string]interface{}{
			"tickets": []tools.Ticket{},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewBackendClient(server.URL, "test-org", "test-pod")
	repositoryID := 1
	assigneeID := 2
	parentID := 3
	status := tools.TicketStatusTodo
	ticketType := tools.TicketTypeTask
	priority := tools.TicketPriorityHigh
	_, err := client.SearchTickets(context.Background(), &repositoryID, &status, &ticketType, &priority, &assigneeID, &parentID, "test query", 20, 1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateTicketWithParentID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		if body["parent_ticket_id"] == nil {
			t.Error("parent_ticket_id should be set")
		} else if int64(body["parent_ticket_id"].(float64)) != 100 {
			t.Errorf("parent_ticket_id: got %v, want 100", body["parent_ticket_id"])
		}

		resp := map[string]interface{}{
			"ticket": tools.Ticket{
				ID:         2,
				Identifier: "AM-2",
				Title:      "Subtask",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewBackendClient(server.URL, "test-org", "test-pod")
	parentID := int64(100)
	_, err := client.CreateTicket(context.Background(), nil, "Subtask", "Description", tools.TicketTypeTask, tools.TicketPriorityMedium, &parentID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
