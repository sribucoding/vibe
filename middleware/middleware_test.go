package middleware_test

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/vibe-go/vibe/httpx"
	"github.com/vibe-go/vibe/middleware"
)

func TestWithTimeout(t *testing.T) {
	// Test case: handler completes before timeout
	t.Run("CompletesBeforeTimeout", func(t *testing.T) {
		handler := httpx.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) error {
			time.Sleep(10 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message":"OK"}`))
			return nil
		})

		wrapped := middleware.WithTimeout(100 * time.Millisecond)(handler)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}
	})

	// Test case: handler times out
	t.Run("TimesOut", func(t *testing.T) {
		handler := httpx.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) error {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
			return nil
		})

		wrapped := middleware.WithTimeout(50 * time.Millisecond)(handler)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusRequestTimeout {
			t.Errorf("Expected status code %d, got %d", http.StatusRequestTimeout, resp.StatusCode)
		}
	})

	// Test case: handler returns an error
	t.Run("HandlerReturnsError", func(t *testing.T) {
		expectedErr := errors.New("handler error")
		handler := httpx.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) error {
			return expectedErr
		})

		wrapped := middleware.WithTimeout(100 * time.Millisecond)(handler)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, resp.StatusCode)
		}
	})

	// Test case: concurrent requests
	t.Run("ConcurrentRequests", func(t *testing.T) {
		handler := httpx.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) error {
			time.Sleep(10 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
			return nil
		})

		wrapped := middleware.WithTimeout(100 * time.Millisecond)(handler)

		var wg sync.WaitGroup
		for range 10 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				w := httptest.NewRecorder()
				wrapped.ServeHTTP(w, req)
				resp := w.Result()
				if resp.StatusCode != http.StatusOK {
					t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
				}
			}()
		}
		wg.Wait()
	})
}

func TestRecovery(t *testing.T) {
	// Test case: no panic
	t.Run("NoPanic", func(t *testing.T) {
		handler := httpx.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) error {
			w.WriteHeader(http.StatusOK)
			return nil
		})

		wrapped := middleware.Recovery(nil)(handler)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}
	})

	// Test case: panic with error
	t.Run("PanicWithError", func(t *testing.T) {
		expectedErr := errors.New("test panic error")
		handler := httpx.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) error {
			panic(expectedErr)
		})

		var buf bytes.Buffer
		logger := log.New(&buf, "[test] ", 0)
		wrapped := middleware.Recovery(logger)(handler)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, resp.StatusCode)
		}

		logOutput := buf.String()
		if !strings.Contains(logOutput, "test panic error") {
			t.Errorf("Expected log to contain panic error, got: %s", logOutput)
		}
	})

	// Test case: panic with non-error value
	t.Run("PanicWithNonError", func(t *testing.T) {
		handler := httpx.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) error {
			panic("string panic")
		})

		var buf bytes.Buffer
		logger := log.New(&buf, "[test] ", 0)
		wrapped := middleware.Recovery(logger)(handler)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, resp.StatusCode)
		}

		logOutput := buf.String()
		if !strings.Contains(logOutput, "string panic") {
			t.Errorf("Expected log to contain panic message, got: %s", logOutput)
		}
	})
}

func TestLogger(t *testing.T) {
	// Test case: with default logger
	t.Run("DefaultLogger", func(t *testing.T) {
		handler := httpx.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) error {
			w.WriteHeader(http.StatusOK)
			return nil
		})

		wrapped := middleware.Logger(nil)(handler)

		req := httptest.NewRequest(http.MethodGet, "/test-path", nil)
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}
	})

	// Test case: with custom logger
	t.Run("CustomLogger", func(t *testing.T) {
		handler := httpx.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) error {
			w.WriteHeader(http.StatusOK)
			return nil
		})

		var buf bytes.Buffer
		logger := log.New(&buf, "[custom] ", 0)
		wrapped := middleware.Logger(logger)(handler)

		req := httptest.NewRequest(http.MethodGet, "/custom-path", nil)
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}

		logOutput := buf.String()
		if !strings.Contains(logOutput, "Request: GET /custom-path") {
			t.Errorf("Expected log to contain request info, got: %s", logOutput)
		}
		if !strings.Contains(logOutput, "Completed: GET /custom-path") {
			t.Errorf("Expected log to contain completion info, got: %s", logOutput)
		}
	})
}

func TestResponseCapturer(t *testing.T) {
	// Test Write method
	t.Run("Write", func(t *testing.T) {
		w := httptest.NewRecorder()
		capturer := middleware.NewResponseCapturer(w)

		n, err := capturer.Write([]byte("test data"))
		if err != nil {
			t.Errorf("Write returned unexpected error: %v", err)
		}
		if n != 9 {
			t.Errorf("Expected to write 9 bytes, got %d", n)
		}

		if w.Body.String() != "test data" {
			t.Errorf("Expected body to be 'test data', got '%s'", w.Body.String())
		}
	})

	// Test WriteHeader with success status
	t.Run("WriteHeaderSuccess", func(t *testing.T) {
		w := httptest.NewRecorder()
		capturer := middleware.NewResponseCapturer(w)

		capturer.WriteHeader(http.StatusOK)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}

		if capturer.Error() != nil {
			t.Errorf("Expected no error for success status, got: %v", capturer.Error())
		}
	})

	// Test WriteHeader with error status
	t.Run("WriteHeaderError", func(t *testing.T) {
		w := httptest.NewRecorder()
		capturer := middleware.NewResponseCapturer(w)

		capturer.WriteHeader(http.StatusInternalServerError)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, w.Code)
		}

		if capturer.Error() == nil {
			t.Error("Expected error for error status, got nil")
		}
	})

	// Test error propagation
	t.Run("ErrorPropagation", func(t *testing.T) {
		// Create a custom ResponseWriter that returns an error on Write
		errorWriter := &errorResponseWriter{
			err: errors.New("write error"),
		}

		capturer := middleware.NewResponseCapturer(errorWriter)

		_, err := capturer.Write([]byte("test"))
		if err == nil || err.Error() != "write error" {
			t.Errorf("Expected 'write error', got: %v", err)
		}

		if capturer.Error() == nil || capturer.Error().Error() != "write error" {
			t.Errorf("Expected capturer to store 'write error', got: %v", capturer.Error())
		}
	})
}

// and returns an error on Write.
type errorResponseWriter struct {
	err error
}

func (e *errorResponseWriter) Header() http.Header {
	return http.Header{}
}

func (e *errorResponseWriter) Write([]byte) (int, error) {
	return 0, e.err
}

func (e *errorResponseWriter) WriteHeader(_ int) {
	// Do nothing
}
