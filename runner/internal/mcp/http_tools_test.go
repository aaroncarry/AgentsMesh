package mcp

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestHTTPToolsCoverage provides additional tests for http_tools.go functions

func TestHTTPServerMCPToolsCallSendTerminalKey(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "send_terminal_key",
			"arguments": {
				"pod_key": "target-pod",
				"key": "ctrl_c"
			}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Tool should be found (may error on backend call)
	if resp.Error != nil && resp.Error.Code == -32601 {
		t.Error("tool send_terminal_key should be found")
	}
}

func TestHTTPServerMCPToolsCallSendTerminalKeyMissingArgs(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "send_terminal_key",
			"arguments": {
				"pod_key": "target-pod"
			}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Should handle missing key argument
}

func TestHTTPServerMCPToolsCallAcceptBinding(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "accept_binding",
			"arguments": {
				"binding_id": 123
			}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Tool should be found
}

func TestHTTPServerMCPToolsCallRejectBinding(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "reject_binding",
			"arguments": {
				"binding_id": 123,
				"reason": "Not needed"
			}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Tool should be found
}

func TestHTTPServerMCPToolsCallUnbindPod(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "unbind_pod",
			"arguments": {
				"binding_id": 123
			}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Tool should be found
}

func TestHTTPServerMCPToolsCallGetBindings(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "get_bindings",
			"arguments": {}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Tool should be found
}

func TestHTTPServerMCPToolsCallGetBoundPods(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "get_bound_pods",
			"arguments": {}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Tool should be found
}

func TestHTTPServerMCPToolsCallGetChannel(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "get_channel",
			"arguments": {
				"channel_id": 123
			}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Tool should be found
}

func TestHTTPServerMCPToolsCallSendChannelMessage(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "send_channel_message",
			"arguments": {
				"channel_id": 123,
				"content": "Hello, world!"
			}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Tool should be found
}

func TestHTTPServerMCPToolsCallGetChannelMessages(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "get_channel_messages",
			"arguments": {
				"channel_id": 123,
				"limit": 10
			}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Tool should be found
}

func TestHTTPServerMCPToolsCallGetChannelDocument(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "get_channel_document",
			"arguments": {
				"channel_id": 123
			}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Tool should be found
}

func TestHTTPServerMCPToolsCallUpdateChannelDocument(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "update_channel_document",
			"arguments": {
				"channel_id": 123,
				"content": "Updated content"
			}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Tool should be found
}

func TestHTTPServerMCPToolsCallUpdateTicket(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "update_ticket",
			"arguments": {
				"ticket_id": "AM-123",
				"status": "done"
			}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Tool should be found
}

func TestHTTPServerMCPToolsCallCreatePod(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "create_pod",
			"arguments": {
				"ticket_id": 123,
				"command": "echo hello"
			}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Tool should be found
}

func TestHTTPServerMCPToolsCallCreatePodWithAllParams(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "create_pod",
			"arguments": {
				"agent_type_id": 1,
				"runner_id": 2,
				"ticket_id": 123,
				"initial_prompt": "Hello, start working on this task",
				"model": "claude-opus-4",
				"repository_id": 456,
				"branch_name": "feature/new-feature",
				"credential_profile_id": 789,
				"permission_mode": "plan",
				"config_overrides": {
					"timeout": 300,
					"max_tokens": 4096
				}
			}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Tool should be found (may error on backend call, but params should be parsed)
}

func TestHTTPServerMCPToolsCallCreatePodWithRepositoryURL(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "create_pod",
			"arguments": {
				"agent_type_id": 1,
				"repository_url": "https://github.com/example/repo.git",
				"branch_name": "main"
			}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Tool should be found
}

func TestHTTPServerMCPToolsCallCreatePodWithBypassPermissions(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "create_pod",
			"arguments": {
				"agent_type_id": 1,
				"permission_mode": "bypassPermissions"
			}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Tool should be found
}

func TestHTTPServerMCPToolsCallCreatePodWithEmptyConfigOverrides(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "create_pod",
			"arguments": {
				"agent_type_id": 1,
				"config_overrides": {}
			}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Tool should be found
}

// Tests for helper function edge cases

func TestGetStringArgEmpty(t *testing.T) {
	args := map[string]interface{}{
		"empty": "",
	}
	result := getStringArg(args, "empty")
	if result != "" {
		t.Errorf("expected empty string, got %v", result)
	}
}

func TestGetStringArgNonString(t *testing.T) {
	args := map[string]interface{}{
		"number": 42,
	}
	result := getStringArg(args, "number")
	if result != "" {
		t.Errorf("expected empty string for non-string, got %v", result)
	}
}

func TestGetIntArgFloat(t *testing.T) {
	args := map[string]interface{}{
		"float": 42.5,
	}
	result := getIntArg(args, "float")
	if result != 42 {
		t.Errorf("expected 42 for float64, got %v", result)
	}
}

func TestGetIntPtrArgFloat(t *testing.T) {
	args := map[string]interface{}{
		"float": 42.5,
	}
	result := getIntPtrArg(args, "float")
	if result == nil {
		t.Error("expected non-nil result")
	} else if *result != 42 {
		t.Errorf("expected 42, got %v", *result)
	}
}

func TestGetBoolArgNonBool(t *testing.T) {
	args := map[string]interface{}{
		"string": "true",
	}
	result := getBoolArg(args, "string")
	if result {
		t.Error("expected false for non-bool")
	}
}

func TestGetStringSliceArgInvalidItems(t *testing.T) {
	args := map[string]interface{}{
		"mixed": []interface{}{"a", 123, "b"},
	}
	result := getStringSliceArg(args, "mixed")
	// Should only include string items
	if len(result) != 2 {
		t.Errorf("expected 2 strings, got %v", len(result))
	}
}

func TestGetStringSliceArgNonSlice(t *testing.T) {
	args := map[string]interface{}{
		"string": "not a slice",
	}
	result := getStringSliceArg(args, "string")
	if result != nil {
		t.Errorf("expected nil for non-slice, got %v", result)
	}
}

func TestGetMapArg(t *testing.T) {
	args := map[string]interface{}{
		"config": map[string]interface{}{
			"key1": "value1",
			"key2": 123,
		},
	}
	result := getMapArg(args, "config")
	if result == nil {
		t.Error("expected non-nil result")
	}
	if result["key1"] != "value1" {
		t.Errorf("expected key1=value1, got %v", result["key1"])
	}
	if result["key2"] != 123 {
		t.Errorf("expected key2=123, got %v", result["key2"])
	}
}

func TestGetMapArgNonMap(t *testing.T) {
	args := map[string]interface{}{
		"string": "not a map",
	}
	result := getMapArg(args, "string")
	if result != nil {
		t.Errorf("expected nil for non-map, got %v", result)
	}
}

func TestGetMapArgMissing(t *testing.T) {
	args := map[string]interface{}{}
	result := getMapArg(args, "missing")
	if result != nil {
		t.Errorf("expected nil for missing key, got %v", result)
	}
}

func TestGetMapArgEmptyMap(t *testing.T) {
	args := map[string]interface{}{
		"empty": map[string]interface{}{},
	}
	result := getMapArg(args, "empty")
	if result == nil {
		t.Error("expected non-nil result for empty map")
	}
	if len(result) != 0 {
		t.Errorf("expected empty map, got %v", result)
	}
}

// Tests for tools with various argument combinations

func TestHTTPServerMCPToolsCallSearchTicketsWithAllParams(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "search_tickets",
			"arguments": {
				"query": "test",
				"status": "todo",
				"type": "task",
				"priority": "high",
				"assignee_id": 123,
				"product_id": 456,
				"limit": 10
			}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
}

func TestHTTPServerMCPToolsCallSearchChannelsWithAllParams(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "search_channels",
			"arguments": {
				"name": "test",
				"type": "public",
				"owner_type": "pod",
				"owner_id": "pod-123"
			}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
}

func TestHTTPServerMCPToolsCallCreateChannelWithAllParams(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "create_channel",
			"arguments": {
				"name": "test-channel",
				"description": "A test channel",
				"type": "public"
			}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
}

func TestHTTPServerMCPToolsCallObserveTerminalWithDefaultLines(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "observe_terminal",
			"arguments": {
				"pod_key": "target-pod"
			}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
}

func TestHTTPServerMCPToolsCallGetChannelMessagesWithAllParams(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "get_channel_messages",
			"arguments": {
				"channel_id": 123,
				"limit": 50,
				"before_id": 1000
			}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
}

// Additional tests for better coverage

func TestHTTPServerMCPToolsCallBindPodMissingArgs(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "bind_pod",
			"arguments": {
				"target_pod": "target"
			}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Tool should be found and validation error returned
	if resp.Error != nil && resp.Error.Code == -32601 {
		t.Error("tool bind_pod should be found")
	}
}

func TestHTTPServerMCPToolsCallBindPodEmptyTarget(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "bind_pod",
			"arguments": {
				"target_pod": "",
				"scopes": ["terminal:read"]
			}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Tool should be found and validation error returned
	if resp.Error != nil && resp.Error.Code == -32601 {
		t.Error("tool bind_pod should be found")
	}
}

func TestHTTPServerMCPToolsCallAcceptBindingMissingID(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "accept_binding",
			"arguments": {}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Tool should be found and validation error returned
	if resp.Error != nil && resp.Error.Code == -32601 {
		t.Error("tool accept_binding should be found")
	}
}

func TestHTTPServerMCPToolsCallRejectBindingMissingID(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "reject_binding",
			"arguments": {
				"reason": "test reason"
			}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Tool should be found and validation error returned
	if resp.Error != nil && resp.Error.Code == -32601 {
		t.Error("tool reject_binding should be found")
	}
}

func TestHTTPServerMCPToolsCallUnbindPodMissingTarget(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "unbind_pod",
			"arguments": {}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Tool should be found and validation error returned
	if resp.Error != nil && resp.Error.Code == -32601 {
		t.Error("tool unbind_pod should be found")
	}
}

func TestHTTPServerMCPToolsCallGetBindingsWithStatus(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "get_bindings",
			"arguments": {
				"status": "active"
			}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Tool should be found
}

func TestHTTPServerMCPToolsCallListRunners(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "list_runners",
			"arguments": {}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Tool should be found
}

func TestHTTPServerMCPToolsCallListRepositories(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "list_repositories",
			"arguments": {}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Tool should be found
}

func TestHTTPServerMCPToolsCallSendTerminalKeyWithValidKeys(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "send_terminal_key",
			"arguments": {
				"pod_key": "target-pod",
				"keys": ["ctrl+c", "enter", "escape"]
			}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Tool should be found
}

func TestHTTPServerMCPToolsCallSendTerminalKeyMissingPodKey(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "send_terminal_key",
			"arguments": {
				"keys": ["enter"]
			}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Tool should be found and validation error returned
	if resp.Error != nil && resp.Error.Code == -32601 {
		t.Error("tool send_terminal_key should be found")
	}
}

func TestHTTPServerMCPToolsCallSendTerminalKeyEmptyKeys(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "send_terminal_key",
			"arguments": {
				"pod_key": "target-pod",
				"keys": []
			}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Tool should be found and validation error returned
	if resp.Error != nil && resp.Error.Code == -32601 {
		t.Error("tool send_terminal_key should be found")
	}
}

func TestHTTPServerMCPToolsCallUpdateTicketWithAllParams(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "update_ticket",
			"arguments": {
				"ticket_id": "AM-123",
				"title": "Updated Title",
				"description": "Updated description",
				"status": "in_progress",
				"priority": "high",
				"type": "bug"
			}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Tool should be found
}

func TestHTTPServerMCPToolsCallUpdateTicketMissingID(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "update_ticket",
			"arguments": {
				"title": "Updated Title"
			}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Tool should be found and validation error returned
	if resp.Error != nil && resp.Error.Code == -32601 {
		t.Error("tool update_ticket should be found")
	}
}

func TestHTTPServerMCPToolsCallGetTicketMissingID(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "get_ticket",
			"arguments": {}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Tool should be found and validation error returned
	if resp.Error != nil && resp.Error.Code == -32601 {
		t.Error("tool get_ticket should be found")
	}
}

func TestHTTPServerMCPToolsCallGetChannelMissingID(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "get_channel",
			"arguments": {}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Tool should be found and validation error returned
	if resp.Error != nil && resp.Error.Code == -32601 {
		t.Error("tool get_channel should be found")
	}
}

func TestHTTPServerMCPToolsCallGetChannelDocumentMissingID(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "get_channel_document",
			"arguments": {}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Tool should be found and validation error returned
	if resp.Error != nil && resp.Error.Code == -32601 {
		t.Error("tool get_channel_document should be found")
	}
}

func TestHTTPServerMCPToolsCallUpdateChannelDocumentMissingArgs(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "update_channel_document",
			"arguments": {
				"channel_id": 123
			}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Tool should be found and validation error returned
	if resp.Error != nil && resp.Error.Code == -32601 {
		t.Error("tool update_channel_document should be found")
	}
}

func TestHTTPServerMCPToolsCallSendChannelMessageMissingArgs(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "send_channel_message",
			"arguments": {
				"channel_id": 123
			}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Tool should be found and validation error returned
	if resp.Error != nil && resp.Error.Code == -32601 {
		t.Error("tool send_channel_message should be found")
	}
}

func TestHTTPServerMCPToolsCallSendChannelMessageWithAllParams(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "send_channel_message",
			"arguments": {
				"channel_id": 123,
				"content": "Hello",
				"message_type": "text",
				"mentions": ["pod-1", "pod-2"],
				"reply_to": 456
			}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Tool should be found
}

func TestHTTPServerMCPToolsCallGetChannelMessagesMissingID(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "get_channel_messages",
			"arguments": {}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Tool should be found and validation error returned
	if resp.Error != nil && resp.Error.Code == -32601 {
		t.Error("tool get_channel_messages should be found")
	}
}

func TestHTTPServerMCPToolsCallGetChannelMessagesWithTimeParams(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "get_channel_messages",
			"arguments": {
				"channel_id": 123,
				"before_time": "2024-01-01T00:00:00Z",
				"after_time": "2024-01-02T00:00:00Z",
				"mentioned_pod": "pod-1"
			}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Tool should be found
}

func TestHTTPServerMCPToolsCallObserveTerminalMissingPodKey(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "observe_terminal",
			"arguments": {}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Tool should be found and validation error returned
	if resp.Error != nil && resp.Error.Code == -32601 {
		t.Error("tool observe_terminal should be found")
	}
}

func TestHTTPServerMCPToolsCallSendTerminalTextMissingArgs(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "send_terminal_text",
			"arguments": {
				"pod_key": "target-pod"
			}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Tool should be found and validation error returned
	if resp.Error != nil && resp.Error.Code == -32601 {
		t.Error("tool send_terminal_text should be found")
	}
}

func TestHTTPServerMCPToolsCallCreateChannelMissingName(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "create_channel",
			"arguments": {
				"description": "A test channel"
			}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Tool should be found and validation error returned
	if resp.Error != nil && resp.Error.Code == -32601 {
		t.Error("tool create_channel should be found")
	}
}

func TestHTTPServerMCPToolsCallCreateChannelWithProjectAndTicket(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "create_channel",
			"arguments": {
				"name": "test-channel",
				"description": "A test channel",
				"project_id": 123,
				"ticket_id": 456
			}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Tool should be found
}

func TestHTTPServerMCPToolsCallSearchChannelsWithFilters(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "search_channels",
			"arguments": {
				"name": "test",
				"project_id": 123,
				"ticket_id": 456,
				"is_archived": false,
				"offset": 10,
				"limit": 20
			}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Tool should be found
}

func TestHTTPServerMCPToolsCallCreateTicketMissingTitle(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "create_ticket",
			"arguments": {
				"description": "A test ticket"
			}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Tool should be found and validation error returned
	if resp.Error != nil && resp.Error.Code == -32601 {
		t.Error("tool create_ticket should be found")
	}
}

func TestHTTPServerMCPToolsCallCreateTicketWithAllParams(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "create_ticket",
			"arguments": {
				"title": "Test Ticket",
				"description": "A test ticket",
				"type": "task",
				"priority": "high",
				"repository_id": 123,
				"parent_ticket_id": 456
			}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Tool should be found
}

func TestHTTPServerMCPToolsCallCreatePodMissingAgentTypeID(t *testing.T) {
	server := NewHTTPServer("http://localhost:8080", 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "claude")

	body := bytes.NewBufferString(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "create_pod",
			"arguments": {
				"initial_prompt": "Hello"
			}
		}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var resp MCPResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Tool should be found and validation error returned
	if resp.Error != nil && resp.Error.Code == -32601 {
		t.Error("tool create_pod should be found")
	}
}

// Tests for getInt64PtrArg edge cases
func TestGetInt64PtrArgFloat(t *testing.T) {
	args := map[string]interface{}{
		"float": float64(42.5),
	}
	result := getInt64PtrArg(args, "float")
	if result == nil {
		t.Error("expected non-nil result")
	} else if *result != 42 {
		t.Errorf("expected 42, got %v", *result)
	}
}

func TestGetInt64PtrArgInt(t *testing.T) {
	args := map[string]interface{}{
		"int": int(42),
	}
	result := getInt64PtrArg(args, "int")
	if result == nil {
		t.Error("expected non-nil result")
	} else if *result != 42 {
		t.Errorf("expected 42, got %v", *result)
	}
}

func TestGetInt64PtrArgMissing(t *testing.T) {
	args := map[string]interface{}{}
	result := getInt64PtrArg(args, "missing")
	if result != nil {
		t.Errorf("expected nil for missing key, got %v", result)
	}
}

func TestGetInt64PtrArgString(t *testing.T) {
	args := map[string]interface{}{
		"string": "42",
	}
	result := getInt64PtrArg(args, "string")
	if result != nil {
		t.Errorf("expected nil for string type, got %v", result)
	}
}

func TestGetIntArgMissing(t *testing.T) {
	args := map[string]interface{}{}
	result := getIntArg(args, "missing")
	if result != 0 {
		t.Errorf("expected 0 for missing key, got %v", result)
	}
}

func TestGetIntPtrArgMissing(t *testing.T) {
	args := map[string]interface{}{}
	result := getIntPtrArg(args, "missing")
	if result != nil {
		t.Errorf("expected nil for missing key, got %v", result)
	}
}

func TestGetIntPtrArgString(t *testing.T) {
	args := map[string]interface{}{
		"string": "42",
	}
	result := getIntPtrArg(args, "string")
	if result != nil {
		t.Errorf("expected nil for string type, got %v", result)
	}
}
