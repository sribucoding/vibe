# Vibe

Vibe is a lightweight, flexible Go web framework designed for building modern web applications and APIs with minimal boilerplate.

## Features

- **Simple and intuitive API**: Build web applications with a clean, expressive syntax
- **Middleware support**: Add global or route-specific middleware for cross-cutting concerns
- **JSON handling**: Built-in utilities for working with JSON requests and responses
- **Error handling**: Comprehensive error handling with helpful response utilities
- **Flexible routing**: HTTP method-based routing with path parameters
- **CORS support**: Built-in middleware for handling Cross-Origin Resource Sharing
- **Route groups**: Organize routes with common prefixes and middleware

## Installation

```bash
go get github.com/sribucoding/vibe
```

## Quick Start

```go
package main

import (
    "net/http"
    "github.com/sribucoding/vibe"
    "github.com/sribucoding/vibe/respond"
)

func main() {
    router := vibe.New()

    // Create an API group with version prefix
    api := router.Group("/api/v1")

    // Add routes to the group
    api.Get("/users", func(w http.ResponseWriter, r *http.Request) error {
        return respond.JSON(w, http.StatusOK, map[string]string{"message": "List of users"})
    })

    // Nested groups for more organization
    admin := api.Group("/admin")
    admin.Get("/stats", func(w http.ResponseWriter, r *http.Request) error {
        return respond.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
    })

    http.ListenAndServe(":8080", router)
}
```
