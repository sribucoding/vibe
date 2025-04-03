package httpjson

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Decode decodes the JSON request body into the provided value.
func Decode(r *http.Request, v interface{}) error {
	if r.Body == nil {
		return fmt.Errorf("request body is empty")
	}
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return fmt.Errorf("failed to decode JSON: %v", err)
	}

	return nil
}

// Encode sets the Content-Type to "application/json" and encodes the data as JSON.
func Encode(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(data)
}
