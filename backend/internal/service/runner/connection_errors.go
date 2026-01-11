package runner

import "errors"

// Connection-related errors
var (
	ErrRunnerNotConnected = errors.New("runner not connected")
	ErrConnectionClosed   = errors.New("connection closed")
)
