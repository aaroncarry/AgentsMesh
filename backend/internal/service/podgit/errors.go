package podgit

import "net/http"

// Error represents a structured Pod Git API error.
type Error struct {
	Code       string
	Message    string
	HTTPStatus int
}

func (e *Error) Error() string {
	return e.Message
}

func newError(code, message string, httpStatus int) *Error {
	return &Error{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

func notFound(message string) *Error {
	return newError("RESOURCE_NOT_FOUND", message, http.StatusNotFound)
}

func forbidden(message string) *Error {
	return newError("ACCESS_DENIED", message, http.StatusForbidden)
}

func badRequest(code, message string) *Error {
	return newError(code, message, http.StatusBadRequest)
}

func serviceUnavailable(code, message string) *Error {
	return newError(code, message, http.StatusServiceUnavailable)
}

func internalError(message string) *Error {
	return newError("INTERNAL_ERROR", message, http.StatusInternalServerError)
}

func gatewayTimeout(code, message string) *Error {
	return newError(code, message, http.StatusGatewayTimeout)
}
