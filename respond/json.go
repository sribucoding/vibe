package respond

import (
	"encoding/json"
	"net/http"
)

// JSON sets the Content-Type to "application/json", sets the provided status code,
// and encodes the data as JSON.
func JSON(w http.ResponseWriter, statusCode int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(data)
}
