package httpx_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vibe-go/vibe/httpx"
)

type testStruct struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func TestJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"message": "test"}

	err := httpx.JSON(w, data, http.StatusCreated)
	if err != nil {
		t.Errorf("JSON() returned error: %v", err)
	}

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	// Check status code
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	// Check content type
	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}

	// Check body
	var result map[string]string
	err = json.Unmarshal(body, &result)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if result["message"] != "test" {
		t.Errorf("Expected message 'test', got '%s'", result["message"])
	}
}

func TestError(t *testing.T) {
	w := httptest.NewRecorder()
	testErr := errors.New("Invalid request")

	err := httpx.Error(w, testErr, http.StatusBadRequest)
	if err != nil {
		t.Errorf("Error() returned error: %v", err)
	}

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	// Check status code
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}

	// Check body
	expected := `{"error":"Invalid request"}`
	if strings.TrimSpace(string(body)) != expected {
		t.Errorf("Expected body %s, got %s", expected, string(body))
	}
}

func TestNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	testErr := errors.New("Resource not found")

	err := httpx.NotFound(w, testErr)
	if err != nil {
		t.Errorf("NotFound() returned error: %v", err)
	}

	resp := w.Result()

	// Check status code
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

func TestBadRequest(t *testing.T) {
	w := httptest.NewRecorder()
	testErr := errors.New("Invalid input")

	err := httpx.BadRequest(w, testErr)
	if err != nil {
		t.Errorf("BadRequest() returned error: %v", err)
	}

	resp := w.Result()

	// Check status code
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestInternalError(t *testing.T) {
	w := httptest.NewRecorder()
	testErr := io.EOF

	err := httpx.InternalError(w, testErr)
	if err != nil {
		t.Errorf("InternalError() returned error: %v", err)
	}

	resp := w.Result()

	// Check status code
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}
}

func TestNotFoundEmptyMessage(t *testing.T) {
	w := httptest.NewRecorder()

	err := httpx.NotFound(w, nil)
	if err != nil {
		t.Errorf("NotFound() returned error: %v", err)
	}

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	// Check status code
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, resp.StatusCode)
	}

	// Check that default message is used
	expected := `{"error":"resource not found"}`
	if strings.TrimSpace(string(body)) != expected {
		t.Errorf("Expected body %s, got %s", expected, string(body))
	}
}

func TestWithStatusCode(t *testing.T) {
	w := httptest.NewRecorder()

	httpx.WithStatusCode(w, http.StatusTeapot)

	resp := w.Result()
	if resp.StatusCode != http.StatusTeapot {
		t.Errorf("Expected status code %d, got %d", http.StatusTeapot, resp.StatusCode)
	}
}

func TestDecode(t *testing.T) {
	t.Run("ValidJSON", func(t *testing.T) {
		jsonBody := `{"name":"test","value":123}`
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		var result testStruct
		err := httpx.DecodeJSON(req, &result)

		if err != nil {
			t.Errorf("JSONDecode() returned error for valid JSON: %v", err)
		}

		if result.Name != "test" || result.Value != 123 {
			t.Errorf("JSONDecode() didn't parse correctly, got %+v", result)
		}
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		jsonBody := `{"name":"test",value:123}` // Missing quotes around value
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		var result testStruct
		err := httpx.DecodeJSON(req, &result)

		if err == nil {
			t.Error("JSONDecode() didn't return error for invalid JSON")
		}
	})

	t.Run("EmptyBody", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
		req.Header.Set("Content-Type", "application/json")

		var result testStruct
		err := httpx.DecodeJSON(req, &result)

		if err == nil {
			t.Error("JSONDecode() didn't return error for empty body")
		}
	})

	t.Run("DecodeNilBody", func(t *testing.T) {
		// Test with nil body
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("Content-Type", "application/json")
		var result testStruct
		err := httpx.DecodeJSON(req, &result)
		if err == nil {
			t.Error("JSONDecode() didn't return error for nil body")
		}
	})
}
