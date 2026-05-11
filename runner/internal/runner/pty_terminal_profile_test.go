package runner

import (
	"bytes"
	"sync"
	"testing"
	"time"

	"github.com/anthropics/agentsmesh/runner/internal/terminal/vt"
)

type captureRelayWriter struct {
	mu   sync.Mutex
	data bytes.Buffer
}

func (w *captureRelayWriter) SendOutput(data []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	_, _ = w.data.Write(data)
	return nil
}

func (w *captureRelayWriter) IsConnected() bool {
	return true
}

func (w *captureRelayWriter) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.data.String()
}

func TestPTYTerminalProfile_DroidUsesRawPassthroughForIncompleteSyncFrame(t *testing.T) {
	virtualTerm := vt.NewVirtualTerminal(80, 24, 100)
	profile := selectPTYTerminalProfile("droid")
	agg := profile.NewAggregator()
	defer agg.Stop()

	relay := &captureRelayWriter{}
	agg.SetRelayClient(relay)

	data := []byte("\x1b[?2026hHello from droid")
	virtualTerm.Feed(data)
	agg.Write(data)

	deadline := time.After(500 * time.Millisecond)
	for {
		output := relay.String()
		outputBytes := []byte(output)
		if bytes.Contains(outputBytes, []byte("Hello")) &&
			bytes.Contains(outputBytes, []byte("from")) &&
			bytes.Contains(outputBytes, []byte("droid")) {
			if bytes.Contains(outputBytes, []byte("\x1b[?2026h")) {
				t.Fatalf("raw passthrough should not forward synchronized-output start: %q", output)
			}
			return
		}

		select {
		case <-deadline:
			t.Fatalf("timed out waiting for droid raw passthrough output, got %q", output)
		case <-time.After(10 * time.Millisecond):
		}
	}
}

func TestSelectPTYTerminalProfileMatchesDroidExecutable(t *testing.T) {
	droid := selectPTYTerminalProfile("/usr/local/bin/droid")
	if droid.name != "droid-raw-passthrough" {
		t.Fatalf("expected droid raw passthrough profile, got %q", droid.name)
	}

	bash := selectPTYTerminalProfile("/usr/local/bin/bash")
	if bash.name != "default" {
		t.Fatalf("expected default profile for bash, got %q", bash.name)
	}
}
