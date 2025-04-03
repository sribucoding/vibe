package httpjson_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vibe-go/vibe/httpjson"
)

type testStruct struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func TestDecode(t *testing.T) {
	t.Run("ValidJSON", func(t *testing.T) {
		jsonBody := `{"name":"test","value":123}`
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		var result testStruct
		err := httpjson.Decode(req, &result)

		if err != nil {
			t.Errorf("Decode() returned error for valid JSON: %v", err)
		}

		if result.Name != "test" || result.Value != 123 {
			t.Errorf("Decode() didn't parse correctly, got %+v", result)
		}
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		jsonBody := `{"name":"test",value:123}` // Missing quotes around value
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		var result testStruct
		err := httpjson.Decode(req, &result)

		if err == nil {
			t.Error("Decode() didn't return error for invalid JSON")
		}
	})

	t.Run("EmptyBody", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
		req.Header.Set("Content-Type", "application/json")

		var result testStruct
		err := httpjson.Decode(req, &result)

		if err == nil {
			t.Error("Decode() didn't return error for empty body")
		}
	})

	t.Run("DecodeNilBody", func(t *testing.T) {
		// Test with nil body
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("Content-Type", "application/json")
		var result testStruct
		err := httpjson.Decode(req, &result)
		if err == nil {
			t.Error("Decode() didn't return error for nil body")
		}
	})
}
