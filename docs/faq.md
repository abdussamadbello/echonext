# Frequently Asked Questions (FAQ)

## General Questions

### What is EchoNext?

EchoNext is a type-safe wrapper around the Echo web framework that provides automatic OpenAPI specification generation, built-in validation, and reduced boilerplate code. It allows you to build robust APIs with compile-time type safety while maintaining full compatibility with Echo.

### How is EchoNext different from Echo?

EchoNext builds on top of Echo and adds:
- **Type-safe handlers** - Automatic request parsing and response serialization
- **Automatic OpenAPI generation** - Generate API docs from your code
- **Built-in validation** - Validate requests using struct tags
- **Less boilerplate** - Focus on business logic, not HTTP details

You still have access to all Echo features and middleware.

### Is EchoNext production-ready?

Yes! EchoNext is built on top of the mature and battle-tested Echo framework. It adds a thin layer for type safety and OpenAPI generation without compromising performance or stability.

### Can I use Echo middleware with EchoNext?

Yes, absolutely! EchoNext is fully compatible with all Echo middleware:

```go
import "github.com/labstack/echo/v5/middleware"

app := echonext.New()
app.Use(middleware.RequestLogger())
app.Use(middleware.CORS("*"))
app.Use(middleware.RateLimiter(...))
```

### Can I mix EchoNext handlers with regular Echo handlers?

Yes, you can use both in the same application:

```go
app := echonext.New()

// EchoNext typed handler
app.GET("/typed", typedHandler, echonext.Route{...})

// Regular Echo handler
app.GET("/standard", func(c *echo.Context) error {
    return c.String(200, "OK")
})
```

## Installation and Setup

### How do I install EchoNext?

```bash
go get github.com/abdussamadbello/echonext
```

For the CLI tool:

```bash
go install github.com/abdussamadbello/echonext/cmd/echonext-cli@latest
```

### What Go version is required?

EchoNext requires Go 1.24.0 or later.

### Do I need to install the CLI tool?

No, the CLI is optional. It helps with project scaffolding and code generation, but you can use EchoNext without it.

## Type Safety and Handlers

### What handler signatures are supported?

EchoNext supports these signatures:

```go
// No request body
func(c *echo.Context) (Response, error)
func(c *echo.Context) error

// With request body
func(c *echo.Context, req Request) (Response, error)
func(c *echo.Context, req Request) error
```

### How do I handle path parameters?

Use `c.Param()` from the Echo context:

```go
func getUser(c *echo.Context) (UserResponse, error) {
    id := c.Param("id")
    // Use id...
}
```

### How do I handle query parameters?

Use `query` tags in your request struct:

```go
type ListRequest struct {
    Page  int `query:"page" validate:"min=1"`
    Limit int `query:"limit" validate:"min=1,max=100"`
}

func list(c *echo.Context, req ListRequest) (Response, error) {
    // req.Page and req.Limit are automatically populated
}
```

### Can I access request headers?

Yes, through the Echo context:

```go
func handler(c *echo.Context, req Request) (Response, error) {
    token := c.Request().Header.Get("Authorization")
    userAgent := c.Request().UserAgent()
}
```

### How do I handle file uploads?

Use Echo's file handling methods:

```go
func upload(c *echo.Context) error {
    file, err := c.FormFile("file")
    if err != nil {
        return err
    }
    // Handle file...
}
```

## Validation

### What validation tags are available?

EchoNext uses `go-playground/validator`. Common tags:

- `required` - Field is required
- `min=n` - Minimum length/value
- `max=n` - Maximum length/value
- `email` - Email format
- `url` - URL format
- `oneof=a b c` - Value must be one of the options
- `dive` - Validate array/slice elements

See [Validation Reference](./api-reference/validation.md) for complete list.

### How do I validate nested structs?

```go
type Address struct {
    Street string `validate:"required"`
    City   string `validate:"required"`
}

type User struct {
    Name    string  `validate:"required"`
    Address Address `validate:"required"`
}
```

### How do I make a field optional but validated?

Use `omitempty`:

```go
type Request struct {
    Email string `json:"email" validate:"omitempty,email"`
}
```

If provided, it must be a valid email. If not provided, validation passes.

### Can I add custom validation?

Yes, register custom validators:

```go
import "github.com/go-playground/validator/v10"

v := validator.New()
v.RegisterValidation("customtag", customValidationFunc)
```

## OpenAPI and Documentation

### How do I generate OpenAPI specs?

```go
app := echonext.New()
app.SetInfo("My API", "1.0.0", "Description")
app.ServeOpenAPISpec("/api/openapi.json")
app.ServeSwaggerUI("/api/docs", "/api/openapi.json")
```

### How do I add examples to my API docs?

Use the `example` tag:

```go
type User struct {
    Name string `json:"name" example:"John Doe"`
    Age  int    `json:"age" example:"30"`
}
```

### How do I document security requirements?

```go
app.AddSecurityScheme("bearerAuth", echonext.Security{
    Type:   "bearer",
    Scheme: "JWT",
})

app.POST("/protected", handler, echonext.Route{
    Security: []echonext.Security{{Type: "bearer"}},
})
```

### Can I customize the OpenAPI spec?

Yes, you can modify the generated spec:

```go
spec := app.GetOpenAPISpec()
// Modify spec...
```

## Database and Persistence

### Does EchoNext include database support?

EchoNext provides optional database helpers in `pkg/contrib/database`. These are GORM-based utilities that you can use if needed.

### What databases are supported?

Any database supported by GORM:
- PostgreSQL
- MySQL
- SQLite
- SQL Server
- And more...

### How do I use the database helpers?

```go
import "github.com/abdussamadbello/echonext/pkg/contrib/database"

cfg := database.DefaultConfig()
db, err := database.Connect(postgres.Open(dsn), cfg)

repo := database.NewRepository[User](db)
user, err := repo.Find(1)
```

### Can I use a different ORM?

Yes! The database contrib package is optional. Use any database library you prefer.

## Testing

### How do I test EchoNext handlers?

Use the testing contrib package:

```go
import echonexttest "github.com/abdussamadbello/echonext/pkg/contrib/testing"

func TestAPI(t *testing.T) {
    app := echonext.New()
    // Register routes...
    
    client := echonexttest.NewAPIClient(app)
    resp := client.GET("/users/1")
    resp.AssertStatus(t, 200)
}
```

### Can I use regular Go testing?

Yes, EchoNext handlers are just Go functions. Test them like any other function:

```go
func TestHandler(t *testing.T) {
    e := echo.New()
    req := httptest.NewRequest(http.MethodGet, "/", nil)
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)
    
    result, err := handler(c)
    // Assert result...
}
```

## Performance

### Does type safety impact performance?

The performance impact is minimal. EchoNext uses reflection only during route registration (at startup), not during request handling.

### How does EchoNext compare to raw Echo?

EchoNext adds a small overhead for request parsing and response serialization, but it's negligible in most applications. The benefits of type safety and reduced boilerplate usually outweigh the tiny performance cost.

### Can I optimize OpenAPI generation?

OpenAPI specs are generated once at startup, so there's no runtime overhead. Serving the spec is just serving a JSON file.

## CLI Tool

### Where is the CLI tool code?

The CLI tool is in `cmd/echonext-cli/`.

### Can I customize code generation templates?

Not yet, but this is planned for a future release.

### What project structure does `echonext init` create?

See [CLI Overview](./cli/overview.md#generated-structure) for the complete structure.

## Deployment

### How do I deploy an EchoNext app?

Deploy it like any Go application:

1. Build: `go build -o app ./cmd/api`
2. Run: `./app`

See [Deployment Guide](./guides/deployment.md) for detailed instructions.

### Does EchoNext support Docker?

Yes, create a standard Go Dockerfile:

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o api ./cmd/api

FROM alpine:latest
COPY --from=builder /app/api /api
CMD ["/api"]
```

### Can I use EchoNext in Kubernetes?

Yes, EchoNext apps are regular Go binaries that work great in Kubernetes.

## Troubleshooting

### Why am I getting validation errors?

Check that:
1. Required fields are provided
2. Values meet validation constraints (min, max, format, etc.)
3. JSON field names match struct tags

### Why isn't my handler being called?

Ensure:
1. Route is registered before `app.Start()`
2. HTTP method matches (GET, POST, etc.)
3. Path matches exactly
4. Handler signature is correct

### Why is the OpenAPI spec empty?

Make sure you:
1. Call `app.SetInfo()` before registering routes
2. Use `echonext.Route{}` when registering handlers
3. Call `app.ServeOpenAPISpec()` to serve the spec

### How do I enable CORS?

Use Echo's CORS middleware:

```go
import "github.com/labstack/echo/v5/middleware"

app.Use(middleware.CORS("*"))
```

Or with custom config:

```go
app.Use(middleware.CORSWithConfig(middleware.CORSConfig{
    AllowOrigins: []string{"https://example.com"},
    AllowMethods: []string{http.MethodGet, http.MethodPost},
}))
```

## Contributing

### How can I contribute?

See [Contributing Guide](./contributing/guide.md) for details.

### Where do I report bugs?

Open an issue on [GitHub](https://github.com/abdussamadbello/echonext/issues).

### Can I request features?

Yes! Open a feature request on GitHub Issues.

## Migration

### How do I migrate from Echo to EchoNext?

See [Migration Guide](./examples/migration.md) for step-by-step instructions.

### Can I migrate gradually?

Yes! You can use EchoNext and Echo handlers in the same application, allowing gradual migration.

## Still Have Questions?

- Check the [Troubleshooting Guide](./troubleshooting.md)
- Browse [Example Projects](./examples/README.md)
- Open an [issue on GitHub](https://github.com/abdussamadbello/echonext/issues)
