// Package vibe provides a lightweight, flexible web framework for building
// modern web applications and APIs in Go with minimal boilerplate.
//
// Built on top of Go's standard net/http library, Vibe provides a thin layer
// of convenience without sacrificing performance or compatibility.
//
// Vibe focuses on simplicity and expressiveness, offering features like
// middleware support, flexible routing, and built-in utilities for common web
// development tasks.
//
// Key features:
//   - Built on standard library for maximum compatibility and performance
//   - Simple and intuitive API with minimal boilerplate
//   - Middleware support for cross-cutting concerns
//   - Flexible routing with HTTP method-based handlers
//   - JSON utilities for request/response handling
//   - Route groups for organizing endpoints
//   - CORS support via middleware
//
// Basic usage example:
//
//	package main
//
//	import (
//	    "net/http"
//	    "github.com/vibe-go/vibe"
//	    "github.com/vibe-go/vibe/respond"
//	)
//
//	func main() {
//	    router := vibe.New()
//
//	    router.Get("/hello", func(w http.ResponseWriter, r *http.Request) error {
//	        return httpx.JSON(w, http.StatusOK, map[string]string{
//	            "message": "Hello, World!",
//	        })
//	    })
//
//	    http.ListenAndServe(":8080", router)
//	}
package vibe

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/vibe-go/vibe/httpx"
	"github.com/vibe-go/vibe/middleware"
)

// MiddlewareFunc follows the standard http middleware pattern in Go.
type MiddlewareFunc func(http.Handler) http.Handler

// RouterOption defines a function to configure Router options.
// It follows the functional options pattern for flexible configuration.
type RouterOption func(*Router)

// WithoutRecovery disables the default recovery middleware.
// By default, the router includes a recovery middleware to handle panics.
func WithoutRecovery() RouterOption {
	return func(r *Router) {
		r.disableRecovery = true
	}
}

// WithoutTimeout disables the default timeout middleware.
// By default, the router includes a timeout middleware with 1-minute timeout.
func WithoutTimeout() RouterOption {
	return func(r *Router) {
		r.disableTimeout = true
	}
}

// WithTimeout sets a custom timeout duration for the default timeout middleware.
// By default, the router uses 1-minute timeout if timeout middleware is enabled.
func WithTimeout(duration time.Duration) RouterOption {
	return func(r *Router) {
		r.timeout = duration
	}
}

// Router wraps the standard library ServeMux and adds middleware and method-specific route registration.
// It provides a more expressive API for defining routes and applying middleware.
type Router struct {
	mux             *http.ServeMux
	middlewares     []MiddlewareFunc
	logger          *log.Logger
	disableRecovery bool
	disableTimeout  bool
	timeout         time.Duration
}

// New creates a new Router instance with default configuration.
// By default, it includes a recovery middleware to handle panics and
// a timeout middleware with a 30-second timeout.
// Options can be provided to customize the router's behavior.
//
// Example:
//
//	// Default router with recovery and timeout middleware
//	router := vibe.New()
//
//	// Router without recovery middleware
//	router := vibe.New(vibe.WithoutRecovery())
//
//	// Router with custom timeout
//	router := vibe.New(vibe.WithTimeout(10 * time.Second))
//
//	// Router without timeout middleware
//	router := vibe.New(vibe.WithoutTimeout())
func New(options ...RouterOption) *Router {
	const timeout = 60 * time.Second

	router := &Router{
		mux:     http.NewServeMux(),
		logger:  log.New(os.Stdout, "[vibe] ", log.LstdFlags),
		timeout: timeout,
	}

	for _, option := range options {
		option(router)
	}

	if !router.disableRecovery {
		router.Use(middleware.Recovery(router.logger))
	}

	if !router.disableTimeout {
		router.Use(middleware.WithTimeout(router.timeout))
	}

	return router
}

// Use adds a global middleware to the router.
// Global middlewares are applied to all routes.
func (r *Router) Use(mw MiddlewareFunc) {
	r.middlewares = append(r.middlewares, mw)
}

// chainMiddleware chains a list of middlewares with the base handler.
// Middlewares are applied in reverse order so that the first middleware
// in the list is the outermost wrapper.
func chainMiddleware(h http.Handler, mws ...MiddlewareFunc) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// registerRoute is a helper that registers a route with the given HTTP method and pattern.
func (r *Router) registerRoute(method, pattern string, handler httpx.HandlerFunc, mws ...MiddlewareFunc) {
	// Chain the handler with middlewares
	chainedHandler := chainMiddleware(handler, append(r.middlewares, mws...)...)

	r.mux.Handle(method+" "+pattern, chainedHandler)
}

// ServeHTTP implements the http.Handler interface.
// This allows the Router to be used with the standard library's http.ListenAndServe.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

// JSON sets the Content-Type to "application/json" and encodes the data as JSON.
// It's a convenience method for returning JSON responses.
func (r *Router) JSON(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(data)
}

// Get registers a GET route.
// The pattern supports path parameters in the format "/{param}".
func (r *Router) Get(pattern string, handler httpx.HandlerFunc, mws ...MiddlewareFunc) {
	r.registerRoute(http.MethodGet, pattern, handler, mws...)
}

// Post registers a POST route.
// The pattern supports path parameters in the format "/{param}".
func (r *Router) Post(pattern string, handler httpx.HandlerFunc, mws ...MiddlewareFunc) {
	r.registerRoute(http.MethodPost, pattern, handler, mws...)
}

// Put registers a PUT route.
// The pattern supports path parameters in the format "/{param}".
func (r *Router) Put(pattern string, handler httpx.HandlerFunc, mws ...MiddlewareFunc) {
	r.registerRoute(http.MethodPut, pattern, handler, mws...)
}

// Group represents a group of routes with a common prefix and middleware.
// It allows for organizing routes into logical groups.
type Group struct {
	router     *Router
	prefix     string
	middleware []MiddlewareFunc
}

// Group creates a new route group with the given prefix.
// All routes registered on the group will have the specified path prefix.
// Additional middleware can be applied to all routes in the group.
//
// Example:
//
//	api := router.Group("/api/v1")
//	api.Get("/users", listUsers)
func (r *Router) Group(prefix string, mws ...MiddlewareFunc) *Group {
	return &Group{
		router:     r,
		prefix:     prefix,
		middleware: mws,
	}
}

// Use adds middleware to the group.
// The middleware will be applied to all routes in the group.
// Returns the group for method chaining.
func (g *Group) Use(mw MiddlewareFunc) *Group {
	g.middleware = append(g.middleware, mw)
	return g
}

// Get registers a GET route in the group.
// The pattern is relative to the group's prefix.
func (g *Group) Get(pattern string, handler httpx.HandlerFunc, mws ...MiddlewareFunc) {
	fullPath := g.prefix + pattern
	g.router.Get(fullPath, handler, append(g.middleware, mws...)...)
}

// Post registers a POST route in the group.
// The pattern is relative to the group's prefix.
func (g *Group) Post(pattern string, handler httpx.HandlerFunc, mws ...MiddlewareFunc) {
	fullPath := g.prefix + pattern
	g.router.Post(fullPath, handler, append(g.middleware, mws...)...)
}

// Put registers a PUT route in the group.
// The pattern is relative to the group's prefix.
func (g *Group) Put(pattern string, handler httpx.HandlerFunc, mws ...MiddlewareFunc) {
	fullPath := g.prefix + pattern
	g.router.Put(fullPath, handler, append(g.middleware, mws...)...)
}

// Delete registers a DELETE route in the group.
// The pattern is relative to the group's prefix.
func (g *Group) Delete(pattern string, handler httpx.HandlerFunc, mws ...MiddlewareFunc) {
	fullPath := g.prefix + pattern
	g.router.Delete(fullPath, handler, append(g.middleware, mws...)...)
}

// Patch registers a PATCH route in the group.
// The pattern is relative to the group's prefix.
func (g *Group) Patch(pattern string, handler httpx.HandlerFunc, mws ...MiddlewareFunc) {
	fullPath := g.prefix + pattern
	g.router.Patch(fullPath, handler, append(g.middleware, mws...)...)
}

// Options registers an OPTIONS route in the group.
// The pattern is relative to the group's prefix.
func (g *Group) Options(pattern string, handler httpx.HandlerFunc, mws ...MiddlewareFunc) {
	fullPath := g.prefix + pattern
	g.router.Options(fullPath, handler, append(g.middleware, mws...)...)
}

// Head registers a HEAD route in the group.
// The pattern is relative to the group's prefix.
func (g *Group) Head(pattern string, handler httpx.HandlerFunc, mws ...MiddlewareFunc) {
	fullPath := g.prefix + pattern
	g.router.Head(fullPath, handler, append(g.middleware, mws...)...)
}

// Group creates a sub-group with the given prefix.
// The prefix is relative to the parent group's prefix.
// This allows for nested route organization.
//
// Example:
//
//	api := router.Group("/api/v1")
//	admin := api.Group("/admin")
//	admin.Get("/stats", getStats)  // Route: /api/v1/admin/stats
func (g *Group) Group(prefix string, mws ...MiddlewareFunc) *Group {
	fullPrefix := g.prefix + prefix
	return &Group{
		router:     g.router,
		prefix:     fullPrefix,
		middleware: append(g.middleware, mws...),
	}
}

// Delete registers a DELETE route.
// The pattern supports path parameters in the format "/{param}".
func (r *Router) Delete(pattern string, handler httpx.HandlerFunc, mws ...MiddlewareFunc) {
	r.registerRoute(http.MethodDelete, pattern, handler, mws...)
}

// Patch registers a PATCH route.
// The pattern supports path parameters in the format "/{param}".
func (r *Router) Patch(pattern string, handler httpx.HandlerFunc, mws ...MiddlewareFunc) {
	r.registerRoute(http.MethodPatch, pattern, handler, mws...)
}

// Options registers an OPTIONS route.
// The pattern supports path parameters in the format "/{param}".
func (r *Router) Options(pattern string, handler httpx.HandlerFunc, mws ...MiddlewareFunc) {
	r.registerRoute(http.MethodOptions, pattern, handler, mws...)
}

// Head registers a HEAD route.
// The pattern supports path parameters in the format "/{param}".
func (r *Router) Head(pattern string, handler httpx.HandlerFunc, mws ...MiddlewareFunc) {
	r.registerRoute(http.MethodHead, pattern, handler, mws...)
}

// NotFound sets a custom handler for 404 Not Found responses.
// This allows customizing the behavior when no route matches the request.
//
// Example:
//
//	router.NotFound(func(w http.ResponseWriter, r *http.Request) error {
//	    return httpx.JSON(w, map[string]string{
//	        "error": "Resource not found",
//	        "path": r.URL.Path,
//	    }, http.StatusNotFound)
//	})
func (r *Router) NotFound(handler httpx.HandlerFunc) {
	// Chain the handler with global middlewares
	chainedHandler := chainMiddleware(handler, r.middlewares...)

	// Override the default NotFound handler
	r.mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		// Only handle paths that aren't the root path itself
		// This prevents the NotFound handler from handling the root path
		if req.URL.Path != "/" {
			chainedHandler.ServeHTTP(w, req)
		}
	})
}
