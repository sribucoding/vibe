package vibe

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sribucoding/vibe/middleware"
	"github.com/sribucoding/vibe/respond"
)

func TestRouterBasicRouting(t *testing.T) {
	router := New()

	// Register a simple route
	router.Get("/hello", func(w http.ResponseWriter, r *http.Request) error {
		return respond.JSON(w, http.StatusOK, map[string]string{"message": "Hello, World!"})
	})

	// Create a test request
	req := httptest.NewRequest("GET", "/hello", nil)
	w := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(w, req)

	// Check the response
	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	expected := `{"message":"Hello, World!"}`
	if strings.TrimSpace(string(body)) != expected {
		t.Errorf("Expected body %s, got %s", expected, string(body))
	}
}

func TestRouterMethodNotAllowed(t *testing.T) {
	router := New()

	// Register a POST route
	router.Post("/api", func(w http.ResponseWriter, r *http.Request) error {
		return respond.JSON(w, http.StatusOK, map[string]string{"message": "OK"})
	})

	// Try to access with GET
	req := httptest.NewRequest("GET", "/api", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Check the response
	resp := w.Result()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, resp.StatusCode)
	}
}

func TestRouterWithMiddleware(t *testing.T) {
	router := New()

	// Add a test middleware that adds a header
	testMiddleware := func(next middleware.HandlerFunc) middleware.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) error {
			w.Header().Set("X-Test", "middleware-executed")
			return next(w, r)
		}
	}

	router.Use(testMiddleware)

	// Register a route
	router.Get("/test", func(w http.ResponseWriter, r *http.Request) error {
		return respond.JSON(w, http.StatusOK, map[string]string{"message": "OK"})
	})

	// Create a test request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(w, req)

	// Check the response
	resp := w.Result()
	if resp.Header.Get("X-Test") != "middleware-executed" {
		t.Error("Middleware was not executed")
	}
}

func TestRouterGroups(t *testing.T) {
	router := New()

	// Create a group
	api := router.Group("/api")

	// Add a route to the group
	api.Get("/users", func(w http.ResponseWriter, r *http.Request) error {
		return respond.JSON(w, http.StatusOK, map[string]string{"message": "users"})
	})

	// Create a nested group
	v1 := api.Group("/v1")
	v1.Get("/products", func(w http.ResponseWriter, r *http.Request) error {
		return respond.JSON(w, http.StatusOK, map[string]string{"message": "products"})
	})

	// Test the first group route
	req1 := httptest.NewRequest("GET", "/api/users", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	resp1 := w1.Result()
	if resp1.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp1.StatusCode)
	}

	// Test the nested group route
	req2 := httptest.NewRequest("GET", "/api/v1/products", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	resp2 := w2.Result()
	if resp2.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp2.StatusCode)
	}
}

// Add these test functions to your existing vibe_test.go file

func TestRouterOptions(t *testing.T) {
	// Test WithPrefix option
	t.Run("WithPrefix", func(t *testing.T) {
		router := New()

		api := router.Group("/api")

		api.Get("/users", func(w http.ResponseWriter, r *http.Request) error {
			return respond.JSON(w, http.StatusOK, map[string]string{"message": "OK"})
		})

		// Should match /api/users
		req := httptest.NewRequest("GET", "/api/users", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}

		// Should not match /users
		req = httptest.NewRequest("GET", "/users", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		resp = w.Result()
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status code %d, got %d", http.StatusNotFound, resp.StatusCode)
		}
	})

	// Test WithoutRecovery option
	t.Run("WithoutRecovery", func(t *testing.T) {
		router := New(WithoutRecovery())

		// Register a route that will panic
		router.Get("/panic", func(w http.ResponseWriter, r *http.Request) error {
			panic("test panic")
		})

		// This should panic, so we need to recover ourselves for the test
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic, but none occurred")
			}
		}()

		req := httptest.NewRequest("GET", "/panic", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	})
}

func TestAllHTTPMethods(t *testing.T) {
	methods := []struct {
		method     string
		routerFunc func(string, middleware.HandlerFunc, ...middleware.Middleware) *Router
	}{
		{"GET", func(p string, h middleware.HandlerFunc, m ...middleware.Middleware) *Router {
			r := New()
			r.Get(p, h, m...)
			return r
		}},
		{"POST", func(p string, h middleware.HandlerFunc, m ...middleware.Middleware) *Router {
			r := New()
			r.Post(p, h, m...)
			return r
		}},
		{"PUT", func(p string, h middleware.HandlerFunc, m ...middleware.Middleware) *Router {
			r := New()
			r.Put(p, h, m...)
			return r
		}},
		{"DELETE", func(p string, h middleware.HandlerFunc, m ...middleware.Middleware) *Router {
			r := New()
			r.Delete(p, h, m...)
			return r
		}},
		{"PATCH", func(p string, h middleware.HandlerFunc, m ...middleware.Middleware) *Router {
			r := New()
			r.Patch(p, h, m...)
			return r
		}},
		{"OPTIONS", func(p string, h middleware.HandlerFunc, m ...middleware.Middleware) *Router {
			r := New()
			r.Options(p, h, m...)
			return r
		}},
		{"HEAD", func(p string, h middleware.HandlerFunc, m ...middleware.Middleware) *Router {
			r := New()
			r.Head(p, h, m...)
			return r
		}},
	}

	for _, m := range methods {
		t.Run(m.method, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) error {
				return respond.JSON(w, http.StatusOK, map[string]string{"method": r.Method})
			}

			router := m.routerFunc("/test", handler)

			req := httptest.NewRequest(m.method, "/test", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status code %d for %s, got %d", http.StatusOK, m.method, resp.StatusCode)
			}
		})
	}
}

func TestGroupMiddleware(t *testing.T) {
	router := New()

	// Create middleware that adds a header
	headerMiddleware := func(next middleware.HandlerFunc) middleware.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) error {
			w.Header().Set("X-Group-Test", "true")
			return next(w, r)
		}
	}

	// Create a group with middleware
	api := router.Group("/api", headerMiddleware)

	// Add a route to the group
	api.Get("/test", func(w http.ResponseWriter, r *http.Request) error {
		return respond.JSON(w, http.StatusOK, map[string]string{"message": "OK"})
	})

	// Test the route
	req := httptest.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	resp := w.Result()
	if resp.Header.Get("X-Group-Test") != "true" {
		t.Error("Group middleware was not applied")
	}
}

func TestNotFoundHandler(t *testing.T) {
	router := New()

	router.NotFound(func(w http.ResponseWriter, r *http.Request) error {
		return respond.JSON(w, http.StatusNotFound, map[string]string{"error": "Custom not found"})
	})

	req := httptest.NewRequest("GET", "/non-existent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, resp.StatusCode)
	}

	expected := `{"error":"Custom not found"}`
	if strings.TrimSpace(string(body)) != expected {
		t.Errorf("Expected body %s, got %s", expected, string(body))
	}
}
