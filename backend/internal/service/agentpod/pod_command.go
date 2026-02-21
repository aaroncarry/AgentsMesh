package agentpod

import (
	"fmt"
	"strings"

	"github.com/anthropics/agentsmesh/backend/internal/domain/ticket"
)

// BuildTicketPrompt builds an initial prompt from ticket context
func BuildTicketPrompt(t *ticket.Ticket) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Working on ticket: %s", t.Identifier))
	parts = append(parts, fmt.Sprintf("Title: %s", t.Title))
	if t.Description != nil && *t.Description != "" {
		parts = append(parts, fmt.Sprintf("Description: %s", *t.Description))
	}
	return strings.Join(parts, "\n")
}
