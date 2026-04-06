package claude

import (
	"bufio"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTransport_AskUserQuestion_EndToEnd tests the full AskUserQuestion flow:
// 1. CLI sends can_use_tool for AskUserQuestion → OnPermissionRequest fires
// 2. Host responds with allow + updatedInput (answers) → written to stdin
// 3. CLI receives control_response with answers
func TestTransport_AskUserQuestion_EndToEnd(t *testing.T) {
	f := newFixtureWithStdin()
	defer f.Close()

	close(f.Transport.initCh)
	f.Transport.sessionID = "test-session"

	// Collect stdin writes asynchronously (pipe requires concurrent read/write).
	type stdinMsg struct {
		data map[string]any
	}
	stdinCh := make(chan stdinMsg, 10)
	go func() {
		scanner := bufio.NewScanner(f.StdinPR)
		for scanner.Scan() {
			var msg map[string]any
			if json.Unmarshal(scanner.Bytes(), &msg) == nil {
				stdinCh <- stdinMsg{msg}
			}
		}
	}()

	// CLI sends can_use_tool for AskUserQuestion.
	writeLine(f.PW, map[string]any{
		"type":       "control_request",
		"request_id": "ask-q1",
		"request": map[string]any{
			"subtype":   "can_use_tool",
			"tool_name": "AskUserQuestion",
			"input": map[string]any{
				"questions": []map[string]any{{
					"question":    "Which database?",
					"header":      "DB",
					"options":     []map[string]any{{"label": "PostgreSQL"}, {"label": "MySQL"}},
					"multiSelect": false,
				}},
			},
		},
	})
	f.Drain()

	// Verify OnPermissionRequest fired.
	f.mu.Lock()
	require.Len(t, f.PermissionRequests, 1)
	perm := f.PermissionRequests[0]
	f.mu.Unlock()

	assert.Equal(t, "AskUserQuestion", perm.ToolName)
	assert.Equal(t, "ask-q1", perm.RequestID)

	// Respond with updatedInput containing user's answers.
	updatedInput := map[string]any{
		"answers": map[string]any{"Which database?": "PostgreSQL"},
	}
	err := f.Transport.RespondToPermission("ask-q1", true, updatedInput)
	require.NoError(t, err)

	// Read the control_response from the async stdin channel.
	select {
	case received := <-stdinCh:
		msg := received.data
		assert.Equal(t, "control_response", msg["type"])
		resp := msg["response"].(map[string]any)
		respData := resp["response"].(map[string]any)
		assert.Equal(t, "allow", respData["behavior"])
		ui := respData["updatedInput"].(map[string]any)
		answers := ui["answers"].(map[string]any)
		assert.Equal(t, "PostgreSQL", answers["Which database?"])
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for control_response on stdin")
	}
}

func TestTransport_PermissionApprove_WithoutUpdatedInput(t *testing.T) {
	f := newFixtureWithStdin()
	defer f.Close()

	close(f.Transport.initCh)
	f.Transport.sessionID = "s1"

	stdinCh := make(chan map[string]any, 10)
	go func() {
		scanner := bufio.NewScanner(f.StdinPR)
		for scanner.Scan() {
			var msg map[string]any
			if json.Unmarshal(scanner.Bytes(), &msg) == nil {
				stdinCh <- msg
			}
		}
	}()

	// CLI sends can_use_tool for Edit with original input.
	writeLine(f.PW, map[string]any{
		"type":       "control_request",
		"request_id": "edit-1",
		"request": map[string]any{
			"subtype":   "can_use_tool",
			"tool_name": "Edit",
			"input":     map[string]any{"file_path": "/tmp/test.go", "old_string": "foo"},
		},
	})
	f.Drain()

	// Approve with nil updatedInput → should use stored original input.
	err := f.Transport.RespondToPermission("edit-1", true, nil)
	require.NoError(t, err)

	select {
	case msg := <-stdinCh:
		resp := msg["response"].(map[string]any)
		respData := resp["response"].(map[string]any)
		assert.Equal(t, "allow", respData["behavior"])
		ui := respData["updatedInput"].(map[string]any)
		assert.Equal(t, "/tmp/test.go", ui["file_path"])
		assert.Equal(t, "foo", ui["old_string"])
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for control_response on stdin")
	}
}
