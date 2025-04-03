package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/vibe-go/vibe/respond"
)

func TestRecoveryMiddleware(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) error {
		panic("test panic")
	}

	wrapped := Recovery(nil)(handler)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	_ = wrapped(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}
}

func TestTimeoutMiddleware(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) error {
		time.Sleep(100 * time.Millisecond)
		return respond.JSON(w, http.StatusOK, map[string]string{"message": "OK"})
	}

	wrapped := WithTimeout(50 * time.Millisecond)(handler)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// Execute the handler
	wrapped(w, req)

	// Check the response status code instead of the error
	resp := w.Result()
	if resp.StatusCode != http.StatusGatewayTimeout {
		t.Errorf("Expected status code %d for timeout, got %d",
			http.StatusGatewayTimeout, resp.StatusCode)
	}
}

func TestCORSMiddleware(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) error {
		return respond.JSON(w, http.StatusOK, map[string]string{"message": "OK"})
	}

	t.Run("DefaultOptions", func(t *testing.T) {
		wrapped := CORS()(handler)

		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		_ = wrapped(w, req)

		resp := w.Result()
		if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
			t.Errorf("Expected Access-Control-Allow-Origin to be '*', got '%s'",
				resp.Header.Get("Access-Control-Allow-Origin"))
		}
	})

	// Test with custom options
	t.Run("CustomOptions", func(t *testing.T) {
		wrapped := CORS(
			WithAllowOrigin("https://example.com"),
			WithAllowCredentials(true),
		)(handler)

		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		_ = wrapped(w, req)

		resp := w.Result()
		if resp.Header.Get("Access-Control-Allow-Origin") != "https://example.com" {
			t.Errorf("Expected Access-Control-Allow-Origin to be 'https://example.com', got '%s'",
				resp.Header.Get("Access-Control-Allow-Origin"))
		}

		if resp.Header.Get("Access-Control-Allow-Credentials") != "true" {
			t.Errorf("Expected Access-Control-Allow-Credentials to be 'true', got '%s'",
				resp.Header.Get("Access-Control-Allow-Credentials"))
		}
	})

	// Test OPTIONS request
	t.Run("OptionsRequest", func(t *testing.T) {
		wrapped := CORS()(handler)

		req := httptest.NewRequest("OPTIONS", "/", nil)
		w := httptest.NewRecorder()

		_ = wrapped(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d for OPTIONS request, got %d",
				http.StatusOK, resp.StatusCode)
		}
	})
}

// Add these test functions to your existing middleware_test.go file

func TestLoggerMiddleware(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) error {
		return respond.JSON(w, http.StatusOK, map[string]string{"message": "OK"})
	}

	// Test with nil logger (should use default)
	wrapped := Logger(nil)(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	err := wrapped(w, req)
	if err != nil {
		t.Errorf("Logger middleware returned error: %v", err)
	}

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Test with error from handler
	errorHandler := func(w http.ResponseWriter, r *http.Request) error {
		return fmt.Errorf("test error")
	}

	wrapped = Logger(nil)(errorHandler)

	req = httptest.NewRequest("GET", "/test", nil)
	w = httptest.NewRecorder()

	err = wrapped(w, req)
	if err == nil || err.Error() != "test error" {
		t.Errorf("Expected 'test error', got %v", err)
	}
}

func TestCORSAllOptions(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) error {
		return respond.JSON(w, http.StatusOK, map[string]string{"message": "OK"})
	}

	// Test all CORS options
	wrapped := CORS(
		WithAllowOrigin("https://example.com"),
		WithAllowMethods("GET, POST"),
		WithAllowHeaders("X-Custom-Header"),
		WithAllowCredentials(true),
		WithMaxAge(3600),
	)(handler)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	_ = wrapped(w, req)

	resp := w.Result()

	// Check all headers
	if resp.Header.Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Errorf("Expected Access-Control-Allow-Origin to be 'https://example.com', got '%s'",
			resp.Header.Get("Access-Control-Allow-Origin"))
	}

	if resp.Header.Get("Access-Control-Allow-Methods") != "GET, POST" {
		t.Errorf("Expected Access-Control-Allow-Methods to be 'GET, POST', got '%s'",
			resp.Header.Get("Access-Control-Allow-Methods"))
	}

	if resp.Header.Get("Access-Control-Allow-Headers") != "X-Custom-Header" {
		t.Errorf("Expected Access-Control-Allow-Headers to be 'X-Custom-Header', got '%s'",
			resp.Header.Get("Access-Control-Allow-Headers"))
	}

	if resp.Header.Get("Access-Control-Allow-Credentials") != "true" {
		t.Errorf("Expected Access-Control-Allow-Credentials to be 'true', got '%s'",
			resp.Header.Get("Access-Control-Allow-Credentials"))
	}

	if resp.Header.Get("Access-Control-Max-Age") != "3600" {
		t.Errorf("Expected Access-Control-Max-Age to be '3600', got '%s'",
			resp.Header.Get("Access-Control-Max-Age"))
	}
}
