package gitops

// CommandError is a structured error returned by Git executor operations.
type CommandError struct {
	Code    string
	Message string
}

func (e *CommandError) Error() string {
	return e.Message
}

func newCommandError(code, message string) *CommandError {
	return &CommandError{
		Code:    code,
		Message: message,
	}
}
