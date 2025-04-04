package vibe_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/vibe-go/vibe"
	"github.com/vibe-go/vibe/httpx"
)

func TestBasicRouting(t *testing.T) {
	router := vibe.New()

	router.Get("/hello", func(w http.ResponseWriter, _ *http.Request) error {
		return httpx.JSON(w, map[string]string{"message": "Hello, World!"}, http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Hello, World!") {
		t.Errorf("Expected response to contain 'Hello, World!', got %s", string(body))
	}
}

func TestMethodRouting(t *testing.T) {
	router := vibe.New()

	// Register handlers for different HTTP methods
	router.Get("/resource", func(w http.ResponseWriter, _ *http.Request) error {
		return httpx.JSON(w, map[string]string{"method": "GET"}, http.StatusOK)
	})

	router.Post("/resource", func(w http.ResponseWriter, _ *http.Request) error {
		return httpx.JSON(w, map[string]string{"method": "POST"}, http.StatusOK)
	})

	router.Put("/resource", func(w http.ResponseWriter, _ *http.Request) error {
		return httpx.JSON(w, map[string]string{"method": "PUT"}, http.StatusOK)
	})

	router.Delete("/resource", func(w http.ResponseWriter, _ *http.Request) error {
		return httpx.JSON(w, map[string]string{"method": "DELETE"}, http.StatusOK)
	})

	router.Patch("/resource", func(w http.ResponseWriter, _ *http.Request) error {
		return httpx.JSON(w, map[string]string{"method": "PATCH"}, http.StatusOK)
	})

	router.Options("/resource", func(w http.ResponseWriter, _ *http.Request) error {
		return httpx.JSON(w, map[string]string{"method": "OPTIONS"}, http.StatusOK)
	})

	router.Head("/resource", func(w http.ResponseWriter, _ *http.Request) error {
		w.Header().Set("X-Test", "HEAD")
		return nil
	})

	// Test each method
	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodPatch,
		http.MethodOptions,
		http.MethodHead,
	}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/resource", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Method %s: Expected status code %d, got %d", method, http.StatusOK, resp.StatusCode)
			}

			if method == http.MethodHead {
				if resp.Header.Get("X-Test") != "HEAD" {
					t.Errorf("HEAD method: Expected X-Test header to be set")
				}
			} else {
				body, _ := io.ReadAll(resp.Body)
				var result map[string]string
				json.Unmarshal(body, &result)

				if result["method"] != method {
					t.Errorf("Expected method %s, got %s", method, result["method"])
				}
			}
		})
	}
}

func TestPathParameters(t *testing.T) {
	router := vibe.New()

	router.Get("/users/{id}", func(w http.ResponseWriter, r *http.Request) error {
		id := r.PathValue("id")
		return httpx.JSON(w, map[string]string{"id": id}, http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]string
	json.Unmarshal(body, &result)

	if result["id"] != "123" {
		t.Errorf("Expected id '123', got '%s'", result["id"])
	}
}

func TestRouteGroups(t *testing.T) {
	router := vibe.New()

	api := router.Group("/api")
	v1 := api.Group("/v1")

	v1.Get("/users", func(w http.ResponseWriter, _ *http.Request) error {
		return httpx.JSON(w, map[string]string{"path": "api/v1/users"}, http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]string
	json.Unmarshal(body, &result)

	if result["path"] != "api/v1/users" {
		t.Errorf("Expected path 'api/v1/users', got '%s'", result["path"])
	}
}

func TestGroupMiddleware(t *testing.T) {
	router := vibe.New()

	// Create a middleware that adds a header
	headerMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test", "middleware-applied")
			next.ServeHTTP(w, r)
		})
	}

	// Apply middleware to a group
	api := router.Group("/api", headerMiddleware)

	api.Get("/test", func(w http.ResponseWriter, _ *http.Request) error {
		return httpx.JSON(w, map[string]string{"status": "ok"}, http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	if resp.Header.Get("X-Test") != "middleware-applied" {
		t.Errorf("Expected X-Test header to be set by middleware")
	}
}

func TestWithoutRecovery(_ *testing.T) {
	router := vibe.New(vibe.WithoutRecovery())

	// This would normally panic without recovery middleware
	router.Get("/panic", func(_ http.ResponseWriter, _ *http.Request) error {
		panic("test panic")
	})

	// We can't really test that it panics in a unit test without crashing the test
	// This is more of a configuration test
	req := httptest.NewRequest(http.MethodGet, "/other-route", nil)
	w := httptest.NewRecorder()

	// This should not panic
	router.ServeHTTP(w, req)
}

func TestWithTimeout(t *testing.T) {
	// Create router with a very short timeout
	router := vibe.New(vibe.WithTimeout(50 * time.Millisecond))

	router.Get("/slow", func(w http.ResponseWriter, _ *http.Request) error {
		time.Sleep(100 * time.Millisecond) // This should trigger timeout
		return httpx.JSON(w, map[string]string{"status": "completed"}, http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusRequestTimeout {
		t.Errorf("Expected timeout status code %d, got %d", http.StatusRequestTimeout, resp.StatusCode)
	}
}

func TestWithoutTimeout(t *testing.T) {
	router := vibe.New(vibe.WithoutTimeout())

	router.Get("/slow", func(w http.ResponseWriter, _ *http.Request) error {
		time.Sleep(100 * time.Millisecond) // This should not trigger timeout
		return httpx.JSON(w, map[string]string{"status": "completed"}, http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestErrorHandling(t *testing.T) {
	router := vibe.New()

	router.Get("/error", func(_ http.ResponseWriter, _ *http.Request) error {
		return errors.New("test error")
	})

	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "test error") {
		t.Errorf("Expected response to contain error message, got %s", string(body))
	}
}

func TestNotFound(t *testing.T) {
	router := vibe.New()

	// Custom not found handler
	router.NotFound(func(w http.ResponseWriter, r *http.Request) error {
		return httpx.JSON(w, map[string]string{
			"error": "custom not found",
			"path":  r.URL.Path,
		}, http.StatusNotFound)
	})

	req := httptest.NewRequest(http.MethodGet, "/non-existent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "custom not found") {
		t.Errorf("Expected response to contain custom not found message, got %s", string(body))
	}
}

func TestMiddlewareChaining(t *testing.T) {
	router := vibe.New()

	// Create middlewares that add headers
	middleware1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Middleware-1", "applied")
			next.ServeHTTP(w, r)
		})
	}

	middleware2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Middleware-2", "applied")
			next.ServeHTTP(w, r)
		})
	}

	// Apply global middleware
	router.Use(middleware1)

	// Apply route-specific middleware
	router.Get("/test", func(w http.ResponseWriter, _ *http.Request) error {
		return httpx.JSON(w, map[string]string{"status": "ok"}, http.StatusOK)
	}, middleware2)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	if resp.Header.Get("X-Middleware-1") != "applied" {
		t.Errorf("Expected X-Middleware-1 header to be set")
	}

	if resp.Header.Get("X-Middleware-2") != "applied" {
		t.Errorf("Expected X-Middleware-2 header to be set")
	}
}
