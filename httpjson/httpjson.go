package httpjson

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// Decode decodes the JSON request body into the provided value.
func Decode(r *http.Request, v interface{}) error {
	if r.Body == nil {
		return errors.New("request body is empty")
	}
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return fmt.Errorf("failed to decode JSON: %w", err)
	}

	return nil
}
