package runner

import (
	"path/filepath"
	"strings"

	"github.com/anthropics/agentsmesh/runner/internal/terminal/aggregator"
)

type ptyTerminalProfile struct {
	name string
}

func selectPTYTerminalProfile(launchCommand string) ptyTerminalProfile {
	if isDroidExecutable(launchCommand) {
		return ptyTerminalProfile{name: "droid-raw-passthrough"}
	}
	return ptyTerminalProfile{name: "default"}
}

func (p ptyTerminalProfile) NewAggregator() *aggregator.SmartAggregator {
	switch p.name {
	case "droid-raw-passthrough":
		return aggregator.NewSmartAggregator(nil,
			aggregator.WithRawPassthrough(),
			aggregator.WithOutputTransform(aggregator.StripSynchronizedOutputSequences),
		)
	default:
		return aggregator.NewSmartAggregator(nil, aggregator.WithFullRedrawThrottling())
	}
}

func isDroidExecutable(launchCommand string) bool {
	name := strings.ToLower(filepath.Base(launchCommand))
	return name == "droid"
}
