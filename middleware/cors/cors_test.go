package cors_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vibe-go/vibe/httpx"
	"github.com/vibe-go/vibe/middleware/cors"
)

func TestCORSMiddleware(t *testing.T) {
	// Create a simple handler
	handler := httpx.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) error {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"OK"}`))
		return nil
	})

	t.Run("DefaultOptions", func(t *testing.T) {
		// Wrap with default CORS middleware
		wrapped := cors.New()(handler)

		// Create a test request and response recorder
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		// Execute the handler
		wrapped.ServeHTTP(w, req)

		// Check the response headers
		resp := w.Result()
		if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
			t.Errorf("Expected Access-Control-Allow-Origin to be '*', got '%s'",
				resp.Header.Get("Access-Control-Allow-Origin"))
		}
	})

	// Test with custom options
	t.Run("CustomOptions", func(t *testing.T) {
		// Wrap with custom CORS middleware
		wrapped := cors.New(
			cors.WithAllowOrigin("https://example.com"),
			cors.WithAllowCredentials(true),
		)(handler)

		// Create a test request and response recorder
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		// Execute the handler
		wrapped.ServeHTTP(w, req)

		// Check the response headers
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
		// Wrap with default CORS middleware
		wrapped := cors.New()(handler)

		// Create an OPTIONS request and response recorder
		req := httptest.NewRequest(http.MethodOptions, "/", nil)
		w := httptest.NewRecorder()

		// Execute the handler
		wrapped.ServeHTTP(w, req)

		// Check the response
		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d for OPTIONS request, got %d",
				http.StatusOK, resp.StatusCode)
		}
	})

	// Test all CORS options
	t.Run("AllOptions", func(t *testing.T) {
		// Test all CORS options
		wrapped := cors.New(
			cors.WithAllowOrigin("https://example.com"),
			cors.WithAllowMethods("GET, POST"),
			cors.WithAllowHeaders("X-Custom-Header"),
			cors.WithAllowCredentials(true),
			cors.WithMaxAge(3600),
		)(handler)

		// Create a test request and response recorder
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		// Execute the handler
		wrapped.ServeHTTP(w, req)

		// Check all headers
		resp := w.Result()

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
	})
}
