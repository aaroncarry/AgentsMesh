package runner

import "errors"

// Service errors
var (
	ErrRunnerNotFound      = errors.New("runner not found")
	ErrRunnerOffline       = errors.New("runner is offline")
	ErrInvalidToken        = errors.New("invalid registration token")
	ErrInvalidAuth         = errors.New("invalid runner authentication")
	ErrTokenExpired        = errors.New("registration token expired")
	ErrTokenExhausted      = errors.New("registration token usage exhausted")
	ErrRunnerAlreadyExists = errors.New("runner already exists")
	ErrRunnerDisabled      = errors.New("runner is disabled")
	ErrRunnerQuotaExceeded = errors.New("runner quota exceeded")
)
