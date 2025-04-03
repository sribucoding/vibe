package respond

import (
	"fmt"
	"net/http"
)

// Error responds with a JSON error message and the given status code.
func Error(w http.ResponseWriter, status int, message string) error {
	return JSON(w, status, map[string]string{"error": message})
}

// NotFound is a convenience function for 404 responses.
func NotFound(w http.ResponseWriter, message string) error {
	if message == "" {
		message = "Resource not found"
	}
	return Error(w, http.StatusNotFound, message)
}

// BadRequest is a convenience function for 400 responses.
func BadRequest(w http.ResponseWriter, message string) error {
	return Error(w, http.StatusBadRequest, message)
}

// InternalError is a convenience function for 500 responses.
func InternalError(w http.ResponseWriter, err error) error {
	message := "Internal server error"
	if err != nil {
		message = fmt.Sprintf("Internal server error: %v", err)
	}
	return Error(w, http.StatusInternalServerError, message)
}
