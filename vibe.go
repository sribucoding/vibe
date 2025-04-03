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
//	    "github.com/sribucoding/vibe"
//	    "github.com/sribucoding/vibe/respond"
//	)
//
//	func main() {
//	    router := vibe.New()
//
//	    router.Get("/hello", func(w http.ResponseWriter, r *http.Request) error {
//	        return respond.JSON(w, http.StatusOK, map[string]string{
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

	"github.com/sribucoding/vibe/middleware"
)

// HandlerFunc defines the signature for route handlers.
// It returns an error if processing fails, which will be handled by the router.
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

// Middleware defines a function to wrap a HandlerFunc.
// Middleware can perform pre-processing before calling the next handler,
// and post-processing after the next handler returns.
type Middleware func(HandlerFunc) HandlerFunc

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

// Router wraps the standard library ServeMux and adds middleware and method-specific route registration.
// It provides a more expressive API for defining routes and applying middleware.
type Router struct {
	mux             *http.ServeMux
	middlewares     []middleware.Middleware // Global middlewares
	logger          *log.Logger
	disableRecovery bool
}

// New creates a new Router instance with default configuration.
// By default, it includes a recovery middleware to handle panics.
// Options can be provided to customize the router's behavior.
//
// Example:
//
//	// Default router with recovery middleware
//	router := vibe.New()
//
//	// Router without recovery middleware
//	router := vibe.New(vibe.WithoutRecovery())
func New(options ...RouterOption) *Router {
	router := &Router{
		mux:    http.NewServeMux(),
		logger: log.New(os.Stdout, "[vibe] ", log.LstdFlags),
	}

	for _, option := range options {
		option(router)
	}

	if !router.disableRecovery {
		router.Use(middleware.Recovery(router.logger))
	}

	return router
}

// Use adds a global middleware to the router.
// Global middlewares are applied to all routes.
func (r *Router) Use(mw middleware.Middleware) {
	r.middlewares = append(r.middlewares, mw)
}

// chainMiddleware chains a list of middlewares with the base handler.
// Middlewares are applied in reverse order so that the first middleware
// in the list is the outermost wrapper.
func chainMiddleware(h middleware.HandlerFunc, mws ...middleware.Middleware) middleware.HandlerFunc {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// registerRoute is a helper that registers a route with the given HTTP method and pattern.
func (r *Router) registerRoute(method, pattern string, handler middleware.HandlerFunc, mws ...middleware.Middleware) {
	// Chain the handler with middlewares
	chainedHandler := chainMiddleware(handler, append(r.middlewares, mws...)...)

	r.mux.HandleFunc(method+" "+pattern, func(w http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			if method != http.MethodGet && req.Method == http.MethodOptions {
				// Handle OPTIONS requests for non-GET endpoints
				w.WriteHeader(http.StatusOK)
				return
			}
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if err := chainedHandler(w, req); err != nil {
			r.logger.Printf("Error handling %s %s: %v", req.Method, req.URL.Path, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
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
func (r *Router) Get(pattern string, handler middleware.HandlerFunc, mws ...middleware.Middleware) {
	r.registerRoute(http.MethodGet, pattern, handler, mws...)
}

// Post registers a POST route.
// The pattern supports path parameters in the format "/{param}".
func (r *Router) Post(pattern string, handler middleware.HandlerFunc, mws ...middleware.Middleware) {
	r.registerRoute(http.MethodPost, pattern, handler, mws...)
}

// Put registers a PUT route.
// The pattern supports path parameters in the format "/{param}".
func (r *Router) Put(pattern string, handler middleware.HandlerFunc, mws ...middleware.Middleware) {
	r.registerRoute(http.MethodPut, pattern, handler, mws...)
}

// Group represents a group of routes with a common prefix and middleware.
// It allows for organizing routes into logical groups.
type Group struct {
	router     *Router
	prefix     string
	middleware []middleware.Middleware
}

// Group creates a new route group with the given prefix.
// All routes registered on the group will have the specified path prefix.
// Additional middleware can be applied to all routes in the group.
//
// Example:
//
//	api := router.Group("/api/v1")
//	api.Get("/users", listUsers)
func (r *Router) Group(prefix string, mws ...middleware.Middleware) *Group {
	return &Group{
		router:     r,
		prefix:     prefix,
		middleware: mws,
	}
}

// Use adds middleware to the group.
// The middleware will be applied to all routes in the group.
// Returns the group for method chaining.
func (g *Group) Use(mw middleware.Middleware) *Group {
	g.middleware = append(g.middleware, mw)
	return g
}

// Get registers a GET route in the group.
// The pattern is relative to the group's prefix.
func (g *Group) Get(pattern string, handler middleware.HandlerFunc, mws ...middleware.Middleware) {
	fullPath := g.prefix + pattern
	allMiddleware := append(g.middleware, mws...)
	g.router.Get(fullPath, handler, allMiddleware...)
}

// Post registers a POST route in the group.
// The pattern is relative to the group's prefix.
func (g *Group) Post(pattern string, handler middleware.HandlerFunc, mws ...middleware.Middleware) {
	fullPath := g.prefix + pattern
	allMiddleware := append(g.middleware, mws...)
	g.router.Post(fullPath, handler, allMiddleware...)
}

// Put registers a PUT route in the group.
// The pattern is relative to the group's prefix.
func (g *Group) Put(pattern string, handler middleware.HandlerFunc, mws ...middleware.Middleware) {
	fullPath := g.prefix + pattern
	allMiddleware := append(g.middleware, mws...)
	g.router.Put(fullPath, handler, allMiddleware...)
}

// Delete registers a DELETE route in the group.
// The pattern is relative to the group's prefix.
func (g *Group) Delete(pattern string, handler middleware.HandlerFunc, mws ...middleware.Middleware) {
	fullPath := g.prefix + pattern
	allMiddleware := append(g.middleware, mws...)
	g.router.Delete(fullPath, handler, allMiddleware...)
}

// Patch registers a PATCH route in the group.
// The pattern is relative to the group's prefix.
func (g *Group) Patch(pattern string, handler middleware.HandlerFunc, mws ...middleware.Middleware) {
	fullPath := g.prefix + pattern
	allMiddleware := append(g.middleware, mws...)
	g.router.Patch(fullPath, handler, allMiddleware...)
}

// Options registers an OPTIONS route in the group.
// The pattern is relative to the group's prefix.
func (g *Group) Options(pattern string, handler middleware.HandlerFunc, mws ...middleware.Middleware) {
	fullPath := g.prefix + pattern
	allMiddleware := append(g.middleware, mws...)
	g.router.Options(fullPath, handler, allMiddleware...)
}

// Head registers a HEAD route in the group.
// The pattern is relative to the group's prefix.
func (g *Group) Head(pattern string, handler middleware.HandlerFunc, mws ...middleware.Middleware) {
	fullPath := g.prefix + pattern
	allMiddleware := append(g.middleware, mws...)
	g.router.Head(fullPath, handler, allMiddleware...)
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
func (g *Group) Group(prefix string, mws ...middleware.Middleware) *Group {
	fullPrefix := g.prefix + prefix
	allMiddleware := append(g.middleware, mws...)
	return &Group{
		router:     g.router,
		prefix:     fullPrefix,
		middleware: allMiddleware,
	}
}

// Delete registers a DELETE route.
// The pattern supports path parameters in the format "/{param}".
func (r *Router) Delete(pattern string, handler middleware.HandlerFunc, mws ...middleware.Middleware) {
	r.registerRoute(http.MethodDelete, pattern, handler, mws...)
}

// Patch registers a PATCH route.
// The pattern supports path parameters in the format "/{param}".
func (r *Router) Patch(pattern string, handler middleware.HandlerFunc, mws ...middleware.Middleware) {
	r.registerRoute(http.MethodPatch, pattern, handler, mws...)
}

// Options registers an OPTIONS route.
// The pattern supports path parameters in the format "/{param}".
func (r *Router) Options(pattern string, handler middleware.HandlerFunc, mws ...middleware.Middleware) {
	r.registerRoute(http.MethodOptions, pattern, handler, mws...)
}

// Head registers a HEAD route.
// The pattern supports path parameters in the format "/{param}".
func (r *Router) Head(pattern string, handler middleware.HandlerFunc, mws ...middleware.Middleware) {
	r.registerRoute(http.MethodHead, pattern, handler, mws...)
}

// NotFound sets a custom handler for 404 Not Found responses.
// This allows customizing the behavior when no route matches the request.
//
// Example:
//
//	router.NotFound(func(w http.ResponseWriter, r *http.Request) error {
//	    return respond.JSON(w, http.StatusNotFound, map[string]string{
//	        "error": "Resource not found",
//	        "path": r.URL.Path,
//	    })
//	})
func (r *Router) NotFound(handler middleware.HandlerFunc) {
	// Chain the handler with global middlewares
	chainedHandler := chainMiddleware(handler, r.middlewares...)

	// Override the default NotFound handler
	r.mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		// Only handle actual 404s, not other routes
		if req.URL.Path != "/" {
			if err := chainedHandler(w, req); err != nil {
				r.logger.Printf("Error handling NotFound for %s: %v", req.URL.Path, err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
	})
}
