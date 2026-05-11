package factory

import (
	"log/slog"

	"github.com/anthropics/agentsmesh/runner/internal/acp"
	"github.com/anthropics/agentsmesh/runner/internal/agentkit"
)

const TransportType = "factory-droid-acp"

func init() {
	acp.RegisterAgent("droid", TransportType, func(cb acp.EventCallbacks, l *slog.Logger) acp.Transport {
		return acp.NewACPTransportWithOptions(cb, l, acp.ACPTransportOptions{
			SendInitialized: true,
		})
	})

	agentkit.RegisterProcessNames("droid")
}
