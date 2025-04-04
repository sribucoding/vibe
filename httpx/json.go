package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// DecodeJSON decodes the JSON request body into the provided value.
func DecodeJSON(r *http.Request, v interface{}) error {
	if r.Body == nil {
		return errors.New("request body is empty")
	}
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return fmt.Errorf("failed to decode JSON: %w", err)
	}

	return nil
}

// JSON sets the Content-Type to "application/json", sets the provided status code,
// and encodes the data as JSON.
func JSON(w http.ResponseWriter, data interface{}, statusCode int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(data)
}
