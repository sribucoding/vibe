package respond_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vibe-go/vibe/respond"
)

func TestJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"message": "test"}

	err := respond.JSON(w, http.StatusCreated, data)
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

	err := respond.Error(w, http.StatusBadRequest, "Invalid request")
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

	err := respond.NotFound(w, "Resource not found")
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

	err := respond.BadRequest(w, "Invalid input")
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

	err := respond.InternalError(w, testErr)
	if err != nil {
		t.Errorf("InternalError() returned error: %v", err)
	}

	resp := w.Result()

	// Check status code
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}
}

// Add these test functions to your existing respond_test.go file

func TestNotFoundEmptyMessage(t *testing.T) {
	w := httptest.NewRecorder()

	err := respond.NotFound(w, "")
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
	expected := `{"error":"Resource not found"}`
	if strings.TrimSpace(string(body)) != expected {
		t.Errorf("Expected body %s, got %s", expected, string(body))
	}
}

func TestWithStatusCode(t *testing.T) {
	w := httptest.NewRecorder()

	respond.WithStatusCode(w, http.StatusTeapot)

	resp := w.Result()
	if resp.StatusCode != http.StatusTeapot {
		t.Errorf("Expected status code %d, got %d", http.StatusTeapot, resp.StatusCode)
	}
}
