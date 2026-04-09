package claude

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/anthropics/agentsmesh/runner/internal/acp"
)

const controlRequestTimeout = 30 * time.Second

// controlResult carries the outcome of an outgoing control_request.
type controlResult struct {
	response map[string]any
	err      error
}

// controlRequestTracker manages outgoing control_request → control_response matching.
type controlRequestTracker struct {
	counter  atomic.Int64
	mu       sync.Mutex
	pending  map[string]chan controlResult
}

func newControlRequestTracker() *controlRequestTracker {
	return &controlRequestTracker{
		pending: make(map[string]chan controlResult),
	}
}

// nextID generates a unique request ID: "req_<counter>_<hex>".
func (t *controlRequestTracker) nextID() string {
	n := t.counter.Add(1)
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("req_%d_%s", n, hex.EncodeToString(b))
}

// track registers a pending request and returns a channel for the result.
func (t *controlRequestTracker) track(reqID string) chan controlResult {
	ch := make(chan controlResult, 1)
	t.mu.Lock()
	t.pending[reqID] = ch
	t.mu.Unlock()
	return ch
}

// resolve delivers a control_response to the matching pending request.
// Returns true if a pending request was found and resolved.
func (t *controlRequestTracker) resolve(reqID string, response map[string]any, err error) bool {
	t.mu.Lock()
	ch, ok := t.pending[reqID]
	if ok {
		delete(t.pending, reqID)
	}
	t.mu.Unlock()
	if ok {
		ch <- controlResult{response: response, err: err}
	}
	return ok
}

// sendAndWait sends a control_request via writeFunc and blocks until
// a matching control_response arrives or the timeout expires.
func (t *controlRequestTracker) sendAndWait(
	subtype string,
	payload map[string]any,
	writeFunc func(any) error,
) (map[string]any, error) {
	reqID := t.nextID()
	ch := t.track(reqID)

	// Build the request envelope
	request := map[string]any{"subtype": subtype}
	for k, v := range payload {
		request[k] = v
	}
	msg := map[string]any{
		"type":       "control_request",
		"request_id": reqID,
		"request":    request,
	}

	if err := writeFunc(msg); err != nil {
		t.mu.Lock()
		delete(t.pending, reqID)
		t.mu.Unlock()
		return nil, fmt.Errorf("write control_request %s: %w", subtype, err)
	}

	select {
	case result := <-ch:
		return result.response, result.err
	case <-time.After(controlRequestTimeout):
		t.mu.Lock()
		delete(t.pending, reqID)
		t.mu.Unlock()
		return nil, acp.ErrControlTimeout
	}
}
