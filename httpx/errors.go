package httpx

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorResponder is an interface for responding with errors in different formats.
type ErrorResponder interface {
	// Error writes an error response in the appropriate format.
	Error(w http.ResponseWriter, err error, status int) error
}

// JSONErrorResponder implements ErrorResponder for JSON responses.
type JSONErrorResponder struct{}

// Error writes a JSON error response.
func (r JSONErrorResponder) Error(w http.ResponseWriter, err error, status int) error {
	message := "unknown error"
	if err != nil {
		message = err.Error()
	}
	return JSON(w, map[string]string{"error": message}, status)
}

// defaultResponder is the default error responder (JSON).
var defaultResponder ErrorResponder = JSONErrorResponder{}

// DefaultResponder returns the current default error responder.
func DefaultResponder() ErrorResponder {
	return defaultResponder
}

// SetDefaultResponder sets the default error responder.
func SetDefaultResponder(responder ErrorResponder) {
	defaultResponder = responder
}

// Error responds with an error message in the default format and the given status code.
func Error(w http.ResponseWriter, err error, status int) error {
	return DefaultResponder().Error(w, err, status)
}

// NotFound is a convenience function for 404 responses.
func NotFound(w http.ResponseWriter, err error) error {
	if err == nil {
		err = errors.New("resource not found")
	}
	return Error(w, err, http.StatusNotFound)
}

// BadRequest is a convenience function for 400 responses.
func BadRequest(w http.ResponseWriter, err error) error {
	return Error(w, err, http.StatusBadRequest)
}

// InternalError is a convenience function for 500 responses.
func InternalError(w http.ResponseWriter, err error) error {
	if err == nil {
		err = errors.New("internal server error")
	} else {
		err = fmt.Errorf("internal server error: %w", err)
	}
	return Error(w, err, http.StatusInternalServerError)
}
