package factory

import (
	"testing"

	"github.com/anthropics/agentsmesh/runner/internal/acp"
)

func TestRegisterDroidTransport(t *testing.T) {
	if got := acp.TransportTypeForCommand("droid"); got != TransportType {
		t.Fatalf("TransportTypeForCommand(droid) = %q, want %q", got, TransportType)
	}
}
