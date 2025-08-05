# EchoNext

EchoNext is a type-safe wrapper around the Echo web framework that automatically generates OpenAPI specifications and provides request validation. Build robust, well-documented APIs with compile-time type safety.

## Features

- 🔒 **Type-Safe Handlers** - Define handlers with strongly-typed request and response structs
- 📚 **Automatic OpenAPI Generation** - Generate OpenAPI 3.0 specs from your code
- ✅ **Built-in Validation** - Validate requests using struct tags
- 📖 **Swagger UI** - Interactive API documentation out of the box
- 🚀 **Zero Boilerplate** - Focus on business logic, not HTTP details
- 🔌 **Echo Compatible** - Use all Echo middleware and features

## Installation

```bash
go get github.com/abdussamadbello/echonext
```

## Quick Start

```go
package main

import (
    "github.com/abdussamadbello/echonext"
    "github.com/labstack/echo/v4"
)

// Define your request/response types
type CreateUserRequest struct {
    Name  string `json:"name" validate:"required,min=2"`
    Email string `json:"email" validate:"required,email"`
}

type UserResponse struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

func main() {
    // Create new EchoNext app
    app := echonext.New()
    
    // Set API info
    app.SetInfo("User API", "1.0.0", "User management service")
    
    // Register typed routes
    app.POST("/users", createUser, echonext.Route{
        Summary:     "Create a new user",
        Description: "Creates a new user with the provided information",
        Tags:        []string{"Users"},
    })
    
    app.GET("/users/:id", getUser, echonext.Route{
        Summary: "Get user by ID",
        Tags:    []string{"Users"},
    })
    
    // Serve OpenAPI spec and Swagger UI
    app.ServeOpenAPISpec("/api/openapi.json")
    app.ServeSwaggerUI("/api/docs", "/api/openapi.json")
    
    // Start server
    app.Start(":8080")
}

// Handlers with typed parameters
func createUser(c echo.Context, req CreateUserRequest) (UserResponse, error) {
    // Your business logic here
    user := UserResponse{
        ID:    "123",
        Name:  req.Name,
        Email: req.Email,
    }
    return user, nil
}

func getUser(c echo.Context) (UserResponse, error) {
    id := c.Param("id")
    // Fetch user logic
    return UserResponse{
        ID:    id,
        Name:  "John Doe",
        Email: "john@example.com",
    }, nil
}
```

## Handler Signatures

EchoNext supports various handler signatures:

```go
// No request body (GET, DELETE)
func handler(c echo.Context) (ResponseType, error)

// With request body (POST, PUT, PATCH)
func handler(c echo.Context, req RequestType) (ResponseType, error)

// No response body
func handler(c echo.Context) error
```

## Validation

Use struct tags for validation:

```go
type CreatePostRequest struct {
    Title   string   `json:"title" validate:"required,min=3,max=200"`
    Content string   `json:"content" validate:"required,min=10"`
    Tags    []string `json:"tags" validate:"max=5,dive,min=2,max=20"`
    Status  string   `json:"status" validate:"required,oneof=draft published"`
}
```

## Query Parameters

For GET requests, use `query` tags:

```go
type ListUsersRequest struct {
    Page  int    `query:"page" validate:"min=1"`
    Limit int    `query:"limit" validate:"min=1,max=100"`
    Sort  string `query:"sort" validate:"omitempty,oneof=name email created_at"`
}

func listUsers(c echo.Context, req ListUsersRequest) (ListResponse, error) {
    // Access validated query params from req
}
```

## Error Handling

Return errors from handlers for automatic error responses:

```go
func getUser(c echo.Context) (UserResponse, error) {
    id := c.Param("id")
    user, err := db.GetUser(id)
    if err != nil {
        return UserResponse{}, echo.NewHTTPError(404, "user not found")
    }
    return user, nil
}
```

## Middleware & Echo Compatibility

EchoNext is fully compatible with all Echo middleware and features. Since it wraps `*echo.Echo`, you have access to everything Echo provides:

```go
import "github.com/labstack/echo/v4/middleware"

app := echonext.New()

// All standard Echo middleware works
app.Use(middleware.Logger())
app.Use(middleware.Recover())
app.Use(middleware.CORS())
app.Use(middleware.Gzip())
app.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(20)))
app.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
    return username == "admin" && password == "secret", nil
}))
```

### Echo Features Available

- **Context methods**: `c.Param()`, `c.QueryParam()`, `c.FormValue()`, etc.
- **File uploads**: `c.FormFile()`, `c.MultipartForm()`
- **Static files**: `app.Static("/static", "assets")`
- **Route groups**: `api := app.Group("/api")`
- **Custom binders**: Custom request binding logic
- **Error handling**: Echo's centralized error handler
- **Server options**: TLS, graceful shutdown, etc.

### Example with Echo Features

```go
app := echonext.New()

// Use Echo middleware
app.Use(middleware.Logger())
app.Use(middleware.CORS())

// Create route groups (standard Echo)
api := app.Group("/api/v1")

// Static files (standard Echo)
app.Static("/assets", "public")

// EchoNext typed routes work within groups
api.POST("/users", createUser, echonext.Route{
    Summary: "Create user",
    Tags:    []string{"Users"},
})

// Mix typed and standard Echo handlers
app.POST("/upload", func(c echo.Context) error {
    file, err := c.FormFile("upload")
    if err != nil {
        return err
    }
    // Standard Echo file handling
    return c.String(200, "Uploaded: "+file.Filename)
})
```

EchoNext adds type safety and OpenAPI generation **on top of** Echo without removing any functionality!

## Advanced OpenAPI Features

### Security Schemes

Define security requirements for your API:

```go
app := echonext.New()

// Add security schemes
app.AddSecurityScheme("bearerAuth", echonext.Security{
    Type:   "bearer",
    Scheme: "JWT",
})
app.AddSecurityScheme("apiKey", echonext.Security{
    Type: "apiKey",
    Name: "X-API-Key",
    In:   "header",
})

// Apply security to routes
app.POST("/protected", handler, echonext.Route{
    Security: []echonext.Security{
        {Type: "bearer"},
        {Type: "apiKey", Name: "X-API-Key"},
    },
})
```

### Custom Response Status Codes

Use appropriate HTTP status codes:

```go
app.POST("/users", createUser, echonext.Route{
    SuccessStatus: 201, // Returns 201 Created instead of 200 OK
})

app.DELETE("/users/:id", deleteUser, echonext.Route{
    SuccessStatus: 204, // Returns 204 No Content
})
```

### Request/Response Headers

Document required and optional headers:

```go
app.POST("/upload", uploadHandler, echonext.Route{
    RequestHeaders: map[string]echonext.HeaderInfo{
        "X-Request-ID": {
            Description: "Unique request identifier",
            Required:    true,
            Schema:      "string",
        },
    },
    ResponseHeaders: map[string]echonext.HeaderInfo{
        "X-Upload-ID": {
            Description: "ID of uploaded file",
            Schema:      "string",
        },
    },
})
```

### Content Types and Examples

Support multiple content types and provide examples:

```go
type CreateUserRequest struct {
    Name string `json:"name" example:"John Doe"`
    Age  int    `json:"age" example:"30"`
}

app.POST("/users", createUser, echonext.Route{
    ContentTypes: []string{"application/json", "application/xml"},
    Examples: map[string]interface{}{
        "basic": map[string]interface{}{
            "name": "John Doe",
            "age":  30,
        },
    },
})
```

### Complete API Configuration

```go
app := echonext.New()

// Set comprehensive API information
app.SetInfo("My API", "1.0.0", "A comprehensive API example")
app.SetContact("API Team", "https://example.com/support", "api@example.com")
app.SetLicense("MIT", "https://opensource.org/licenses/MIT")
app.SetServers([]echonext.Server{
    {URL: "https://api.example.com/v1", Description: "Production"},
    {URL: "https://staging.example.com/v1", Description: "Staging"},
})
```

## Example Application

Run the example Todo API:

```bash
go run example/main.go
```

Then visit:
- API Server: http://localhost:8080
- API Documentation: http://localhost:8080/api/docs
- OpenAPI Spec: http://localhost:8080/api/openapi.json

## Development

### Running Tests

```bash
go test ./...                    # Run all tests
go test -v ./...                # Run with verbose output
go test -bench=.                # Run benchmarks
go test -cover                  # Run with coverage
```

### Project Structure

```
echonext/
├── echonext.go        # Main package implementation
├── echonext_test.go   # Test suite
├── example/
│   └── main.go        # Complete example application
├── go.mod
└── README.md
```

## API Response Format

All responses are wrapped in a consistent format:

```json
{
  "success": true,
  "data": { ... },
  "error": ""
}
```

Error responses:

```json
{
  "success": false,
  "data": null,
  "error": "Validation failed: Name is required"
}
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT License - see the [LICENSE](LICENSE) file for details.

## Roadmap

- [ ] Support for file uploads
- [ ] WebSocket support
- [ ] GraphQL integration
- [ ] Database integration helpers
- [ ] Code generation from OpenAPI spec
- [ ] Authentication/Authorization helpers