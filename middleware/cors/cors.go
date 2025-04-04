// Package cors provides Cross-Origin Resource Sharing (CORS) middleware for the Vibe framework.
package cors

import (
	"net/http"
	"strconv"

	"github.com/vibe-go/vibe/httpx"
)

// DefaultMaxAge is the default max age for CORS preflight requests (24 hours).
const DefaultMaxAge = 86400

// Config holds the configuration for CORS middleware.
type Config struct {
	allowOrigin      string
	allowMethods     string
	allowHeaders     string
	allowCredentials bool
	maxAge           int
}

// Option defines a function that configures CORS options.
type Option func(*Config)

// WithAllowOrigin sets the Access-Control-Allow-Origin header.
func WithAllowOrigin(origin string) Option {
	return func(c *Config) {
		c.allowOrigin = origin
	}
}

// WithAllowMethods sets the Access-Control-Allow-Methods header.
func WithAllowMethods(methods string) Option {
	return func(c *Config) {
		c.allowMethods = methods
	}
}

// WithAllowHeaders sets the Access-Control-Allow-Headers header.
func WithAllowHeaders(headers string) Option {
	return func(c *Config) {
		c.allowHeaders = headers
	}
}

// WithAllowCredentials sets the Access-Control-Allow-Credentials header.
func WithAllowCredentials(allow bool) Option {
	return func(c *Config) {
		c.allowCredentials = allow
	}
}

// WithMaxAge sets the Access-Control-Max-Age header.
func WithMaxAge(seconds int) Option {
	return func(c *Config) {
		c.maxAge = seconds
	}
}

// New returns a middleware that adds CORS headers with customizable options.
// If no options are provided, sensible defaults are used.
func New(options ...Option) func(next http.Handler) http.Handler {
	// Default configuration
	cfg := &Config{
		allowOrigin:      "*",
		allowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		allowHeaders:     "Content-Type, Authorization",
		allowCredentials: false,
		maxAge:           DefaultMaxAge,
	}

	for _, option := range options {
		option(cfg)
	}

	return func(next http.Handler) http.Handler {
		return httpx.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			w.Header().Set("Access-Control-Allow-Origin", cfg.allowOrigin)
			w.Header().Set("Access-Control-Allow-Methods", cfg.allowMethods)
			w.Header().Set("Access-Control-Allow-Headers", cfg.allowHeaders)

			if cfg.allowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if cfg.maxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", strconv.Itoa(cfg.maxAge))
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return nil
			}
			next.ServeHTTP(w, r)
			return nil
		})
	}
}
