package agentpod

import (
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
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

func TestBuildAgentCommand(t *testing.T) {
	tests := []struct {
		name            string
		model           string
		permissionMode  string
		skipPermissions bool
		contains        []string
		notContains     []string
	}{
		{
			name:            "basic command",
			model:           "opus",
			permissionMode:  "default",
			skipPermissions: false,
			contains:        []string{"claude", "--model opus", "--permission-mode default"},
			notContains:     []string{"--dangerously-skip-permissions"},
		},
		{
			name:            "skip permissions",
			model:           "sonnet",
			permissionMode:  "plan",
			skipPermissions: true,
			contains:        []string{"claude", "--dangerously-skip-permissions", "--model sonnet"},
		},
		{
			name:            "empty values",
			model:           "",
			permissionMode:  "",
			skipPermissions: false,
			contains:        []string{"claude"},
			notContains:     []string{"--model", "--permission-mode"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := BuildAgentCommand(tt.model, tt.permissionMode, tt.skipPermissions)
			for _, s := range tt.contains {
				if !containsStr(cmd, s) {
					t.Errorf("Command does not contain %q: %s", s, cmd)
				}
			}
			for _, s := range tt.notContains {
				if containsStr(cmd, s) {
					t.Errorf("Command should not contain %q: %s", s, cmd)
				}
			}
		})
	}
}

func TestBuildInitialPrompt(t *testing.T) {
	tests := []struct {
		name       string
		prompt     string
		thinkLevel string
		expected   string
	}{
		{
			name:       "with ultrathink",
			prompt:     "Do something",
			thinkLevel: "ultrathink",
			expected:   "Do something\n\nultrathink",
		},
		{
			name:       "with megathink",
			prompt:     "Do something",
			thinkLevel: "megathink",
			expected:   "Do something\n\nmegathink",
		},
		{
			name:       "with none",
			prompt:     "Do something",
			thinkLevel: agentpod.ThinkLevelNone,
			expected:   "Do something",
		},
		{
			name:       "empty think level",
			prompt:     "Do something",
			thinkLevel: "",
			expected:   "Do something",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildInitialPrompt(tt.prompt, tt.thinkLevel)
			if result != tt.expected {
				t.Errorf("Result = %q, want %q", result, tt.expected)
			}
		})
	}
}
