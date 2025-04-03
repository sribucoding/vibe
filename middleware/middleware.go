package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/vibe-go/vibe/respond"
)

// HandlerFunc defines the signature for route handlers.
// It should return an error if processing fails.
// Update HandlerFunc definition
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

// WithTimeout returns a middleware that times out the request after the given duration.
func WithTimeout(timeout time.Duration) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) error {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			r = r.WithContext(ctx)

			done := make(chan error, 1)
			go func() {
				done <- next(w, r)
			}()

			select {
			case err := <-done:
				return err
			case <-ctx.Done():
				return respond.Error(w, http.StatusGatewayTimeout, "Request timeout")
			}
		}
	}
}

// Middleware defines a function to wrap a HandlerFunc.
type Middleware func(HandlerFunc) HandlerFunc

// Recovery returns a middleware that recovers from panics and logs the error.
// It takes a logger to record panic information.
func Recovery(logger *log.Logger) Middleware {
	// Use a default logger if none is provided
	if logger == nil {
		logger = log.New(log.Writer(), "[recovery] ", log.LstdFlags)
	}

	return func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) error {
			defer func() {
				if rec := recover(); rec != nil {
					err, ok := rec.(error)
					if !ok {
						err = fmt.Errorf("%v", rec)
					}
					logger.Printf("recovered from panic: %v", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			return next(w, r)
		}
	}
}

// corsConfig holds the configuration for CORS middleware
type corsConfig struct {
	allowOrigin      string
	allowMethods     string
	allowHeaders     string
	allowCredentials bool
	maxAge           int
}

// CORSOption defines a function that configures CORS options
type CORSOption func(*corsConfig)

// WithAllowOrigin sets the Access-Control-Allow-Origin header
func WithAllowOrigin(origin string) CORSOption {
	return func(c *corsConfig) {
		c.allowOrigin = origin
	}
}

// WithAllowMethods sets the Access-Control-Allow-Methods header
func WithAllowMethods(methods string) CORSOption {
	return func(c *corsConfig) {
		c.allowMethods = methods
	}
}

// WithAllowHeaders sets the Access-Control-Allow-Headers header
func WithAllowHeaders(headers string) CORSOption {
	return func(c *corsConfig) {
		c.allowHeaders = headers
	}
}

// WithAllowCredentials sets the Access-Control-Allow-Credentials header
func WithAllowCredentials(allow bool) CORSOption {
	return func(c *corsConfig) {
		c.allowCredentials = allow
	}
}

// WithMaxAge sets the Access-Control-Max-Age header
func WithMaxAge(seconds int) CORSOption {
	return func(c *corsConfig) {
		c.maxAge = seconds
	}
}

// CORS returns a middleware that adds CORS headers with customizable options.
// If no options are provided, sensible defaults are used.
func CORS(options ...CORSOption) Middleware {
	// Default configuration
	cfg := &corsConfig{
		allowOrigin:      "*",
		allowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		allowHeaders:     "Content-Type, Authorization",
		allowCredentials: false,
		maxAge:           86400, // 24 hours
	}

	for _, option := range options {
		option(cfg)
	}

	return func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) error {
			w.Header().Set("Access-Control-Allow-Origin", cfg.allowOrigin)
			w.Header().Set("Access-Control-Allow-Methods", cfg.allowMethods)
			w.Header().Set("Access-Control-Allow-Headers", cfg.allowHeaders)

			if cfg.allowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if cfg.maxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", cfg.maxAge))
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return nil
			}
			return next(w, r)
		}
	}
}

// Logger returns a middleware that logs each request with method, path, and duration.
func Logger(logger *log.Logger) Middleware {
	if logger == nil {
		logger = log.New(log.Writer(), "[http] ", log.LstdFlags)
	}

	return func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) error {
			logger.Printf("Request: %s %s", r.Method, r.URL.Path)
			err := next(w, r)
			if err != nil {
				logger.Printf("Error: %v", err)
			}
			return err
		}
	}
}
