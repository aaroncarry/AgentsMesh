package agentpod

import (
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/domain/ticket"
)

func TestBuildTicketPrompt(t *testing.T) {
	tests := []struct {
		name     string
		ticket   *ticket.Ticket
		contains []string
	}{
		{
			name: "basic ticket",
			ticket: &ticket.Ticket{
				Identifier: "PROJ-123",
				Title:      "Fix the bug",
			},
			contains: []string{"PROJ-123", "Fix the bug"},
		},
		{
			name: "ticket with description",
			ticket: &ticket.Ticket{
				Identifier:  "PROJ-456",
				Title:       "Add feature",
				Description: strPtr("Detailed description here"),
			},
			contains: []string{"PROJ-456", "Add feature", "Detailed description here"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := BuildTicketPrompt(tt.ticket)
			for _, s := range tt.contains {
				if !containsStr(prompt, s) {
					t.Errorf("Prompt does not contain %q: %s", s, prompt)
				}
			}
		})
	}
}
