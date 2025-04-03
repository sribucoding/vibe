# Vibe

Vibe is a lightweight, flexible Go web framework designed for building modern web applications and APIs with minimal boilerplate. Built on top of Go's standard `net/http` library, Vibe provides a thin layer of convenience without sacrificing performance or compatibility.

## Features

- **Built on standard library**: Leverages Go's `net/http` package for maximum compatibility and performance
- **Simple and intuitive API**: Build web applications with a clean, expressive syntax
- **Middleware support**: Add global or route-specific middleware for cross-cutting concerns
- **JSON handling**: Built-in utilities for working with JSON requests and responses
- **Error handling**: Comprehensive error handling with helpful response utilities
- **Flexible routing**: HTTP method-based routing with path parameters
- **CORS support**: Built-in middleware for handling Cross-Origin Resource Sharing
- **Route groups**: Organize routes with common prefixes and middleware

## Why Vibe?

Vibe takes a pragmatic approach by building on Go's battle-tested `net/http` package rather than reinventing the wheel. This means:

- **No magic**: Clear, predictable behavior that follows Go idioms
- **Easy to learn**: If you know `net/http`, you'll feel right at home
- **Excellent performance**: Minimal overhead compared to raw `net/http`
- **Compatibility**: Works with existing `net/http` middleware and handlers
- **Future-proof**: Benefits from improvements to the standard library

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

## Documentation

For detailed documentation and examples, please visit the GoDoc page .

## Examples

Check out the examples directory for complete working examples:

- Todo API : A simple todo list API demonstrating basic CRUD operations

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch ( git checkout -b feature/amazing-feature )
3. Commit your changes ( git commit -m 'Add some amazing feature' )
4. Push to the branch ( git push origin feature/amazing-feature )
5. Open a Pull Request
   Please make sure your code follows the project's coding standards and includes appropriate tests.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Code of Conduct

Please read our Code of Conduct to keep our community approachable and respectable.

## Security

If you discover a security vulnerability within Vibe, please send an email to sribucoding@gmail.com instead of using the issue tracker.

## Acknowledgments

- The Go community for inspiration and best practices
- All contributors who have helped shape this project

## Future Improvements

- [ ] **Request validation**: Add built-in request validation with custom error messages
- [ ] **WebSocket support**: Real-time communication capabilities
- [ ] **Rate limiting middleware**: Protect APIs from abuse
- [ ] **Authentication middleware**: Common auth patterns (JWT, OAuth)
- [ ] **Swagger/OpenAPI integration**: Automatic API documentation
- [ ] **Hot reload**: Development mode with automatic reloading
