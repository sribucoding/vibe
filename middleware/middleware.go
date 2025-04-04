package middleware

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/vibe-go/vibe/httpx"
)

func WithTimeout(timeout time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return httpx.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			r = r.WithContext(ctx)

			done := make(chan struct{}, 1)
			var err error

			go func() {
				defer func() {
					done <- struct{}{}
				}()

				respCapturer := NewResponseCapturer(w)
				next.ServeHTTP(respCapturer, r)

				if respCapturer.Error() != nil {
					err = respCapturer.Error()
				}
			}()

			select {
			case <-done:
				if err != nil {
					return err
				}
				return nil
			case <-ctx.Done():
				return httpx.Error(w, errors.New("request timed out"), http.StatusRequestTimeout)
			}
		})
	}
}

// Recovery returns a middleware that recovers from panics and logs the error.
// It takes a logger to record panic information.
func Recovery(logger *log.Logger) func(next http.Handler) http.Handler {
	// Use a default logger if none is provided
	if logger == nil {
		logger = log.New(log.Writer(), "[recovery] ", log.LstdFlags)
	}

	return func(next http.Handler) http.Handler {
		return httpx.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			defer func() {
				if rec := recover(); rec != nil {
					err, ok := rec.(error)
					if !ok {
						err = fmt.Errorf("%v", rec)
					}
					logger.Printf("recovered from panic: %v", err)
					err = httpx.InternalError(w, err)
					if err != nil {
						logger.Printf("failed to write error response: %v", err)
					}
				}
			}()
			next.ServeHTTP(w, r)
			return nil
		})
	}
}

// Logger returns a middleware that logs each request with method, path, and duration.
func Logger(logger *log.Logger) func(next http.Handler) http.Handler {
	if logger == nil {
		logger = log.New(log.Writer(), "[http] ", log.LstdFlags)
	}

	return func(next http.Handler) http.Handler {
		return httpx.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			start := time.Now()
			logger.Printf("Request: %s %s", r.Method, r.URL.Path)

			next.ServeHTTP(w, r)

			logger.Printf("Completed: %s %s in %v", r.Method, r.URL.Path, time.Since(start))
			return nil
		})
	}
}

// ResponseCapturer is a wrapper for http.ResponseWriter that captures errors.
type ResponseCapturer struct {
	http.ResponseWriter
	Err error
}

// NewResponseCapturer creates a new response capturer that wraps a ResponseWriter.
func NewResponseCapturer(w http.ResponseWriter) *ResponseCapturer {
	return &ResponseCapturer{ResponseWriter: w}
}

func (r *ResponseCapturer) setError(err error) {
	r.Err = err
}

// Write overrides the underlying ResponseWriter's Write method to capture errors.
func (r *ResponseCapturer) Write(b []byte) (int, error) {
	n, err := r.ResponseWriter.Write(b)
	if err != nil {
		r.setError(err)
	}
	return n, err
}

// WriteHeader overrides the underlying ResponseWriter's WriteHeader method.
func (r *ResponseCapturer) WriteHeader(statusCode int) {
	// Optionally capture non-2xx status codes as errors
	if statusCode >= http.StatusBadRequest {
		r.setError(fmt.Errorf("response status code: %d", statusCode))
	}
	r.ResponseWriter.WriteHeader(statusCode)
}

// Error returns the captured error.
func (r *ResponseCapturer) Error() error {
	return r.Err
}
