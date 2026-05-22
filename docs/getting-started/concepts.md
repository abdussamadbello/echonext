# Core Concepts

Understanding the core concepts of EchoNext will help you build better APIs faster.

## What is EchoNext?

EchoNext is a **type-safe wrapper** around the Echo web framework that provides:

1. **Type-safe handlers** with automatic request/response handling
2. **Automatic OpenAPI 3.0 specification** generation
3. **Built-in validation** using struct tags
4. **Zero boilerplate** - focus on business logic
5. **Full Echo compatibility** - use all Echo features

## Core Components

### 1. The App

The `App` is the main EchoNext application:

```go
app := echonext.New()
```

It wraps `*echo.Echo`, so you have access to all Echo methods:

```go
app.Use(middleware.Logger())    // Echo middleware
app.Static("/assets", "public") // Echo static files
app.Group("/api/v1")           // Echo route groups
```

### 2. Type-Safe Handlers

Traditional Echo handlers:

```go
func handler(c *echo.Context) error {
    // Manual parsing, binding, validation
    var req Request
    if err := c.Bind(&req); err != nil {
        return err
    }
    if err := c.Validate(&req); err != nil {
        return err
    }
    
    // Business logic
    result := process(req)
    
    // Manual response
    return c.JSON(200, result)
}
```

EchoNext handlers:

```go
func handler(c *echo.Context, req Request) (Response, error) {
    // Request is already parsed and validated!
    result := process(req)
    return result, nil  // Response is automatically serialized
}
```

### 3. Handler Signatures

EchoNext supports multiple handler signatures:

```go
// No request body (GET, DELETE)
func handler(c *echo.Context) (Response, error)

// With request body (POST, PUT, PATCH)
func handler(c *echo.Context, req Request) (Response, error)

// No response body
func handler(c *echo.Context) error
func handler(c *echo.Context, req Request) error

// Access to Echo context is always available
func handler(c *echo.Context, req Request) (Response, error) {
    userId := c.Param("id")        // Path parameters
    token := c.Request().Header.Get("Authorization")  // Headers
    // ...
}
```

### 4. Request Types

Define request structures with validation:

```go
type CreateUserRequest struct {
    Name     string   `json:"name" validate:"required,min=2,max=100"`
    Email    string   `json:"email" validate:"required,email"`
    Age      int      `json:"age" validate:"min=18,max=120"`
    Tags     []string `json:"tags" validate:"max=5,dive,min=2"`
    Password string   `json:"password" validate:"required,min=8"`
}
```

For query parameters, use `query` tag:

```go
type ListUsersRequest struct {
    Page   int    `query:"page" validate:"min=1"`
    Limit  int    `query:"limit" validate:"min=1,max=100"`
    Sort   string `query:"sort" validate:"omitempty,oneof=name email created"`
    Search string `query:"search" validate:"omitempty,min=2"`
}
```

### 5. Response Types

Define response structures:

```go
type UserResponse struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
}

type ListUsersResponse struct {
    Users []UserResponse `json:"users"`
    Total int           `json:"total"`
    Page  int           `json:"page"`
}
```

Responses are automatically wrapped:

```json
{
  "success": true,
  "data": {
    "id": "123",
    "name": "John Doe",
    "email": "john@example.com"
  },
  "error": ""
}
```

### 6. Route Configuration

Configure routes with metadata for OpenAPI:

```go
app.POST("/users", createUser, echonext.Route{
    Summary:       "Create a new user",
    Description:   "Creates a new user account with the provided information",
    Tags:          []string{"Users"},
    SuccessStatus: 201,
})

app.GET("/users/:id", getUser, echonext.Route{
    Summary:     "Get user by ID",
    Description: "Retrieves a user by their unique identifier",
    Tags:        []string{"Users"},
})
```

### 7. Validation

EchoNext uses the `validator` package for validation:

**Common Validation Tags:**

```go
type Request struct {
    // Required fields
    Name string `validate:"required"`
    
    // String length
    Username string `validate:"min=3,max=20"`
    
    // Email format
    Email string `validate:"email"`
    
    // Numeric ranges
    Age int `validate:"min=18,max=120"`
    
    // Allowed values
    Status string `validate:"oneof=active inactive pending"`
    
    // URL format
    Website string `validate:"url"`
    
    // Array validation
    Tags []string `validate:"max=5,dive,min=2,max=20"`
    
    // Optional with validation
    Phone string `validate:"omitempty,e164"` // E.164 phone format
    
    // Complex validation
    Password string `validate:"required,min=8,containsany=!@#$%"`
}
```

See [Validation Reference](../api-reference/validation.md) for all available tags.

### 8. OpenAPI Generation

EchoNext automatically generates OpenAPI 3.0 specifications:

```go
app := echonext.New()

// Set API information
app.SetInfo("My API", "1.0.0", "API description")
app.SetContact("API Team", "https://example.com", "api@example.com")
app.SetLicense("MIT", "https://opensource.org/licenses/MIT")

// Add servers
app.SetServers([]echonext.Server{
    {URL: "https://api.example.com", Description: "Production"},
    {URL: "https://staging.example.com", Description: "Staging"},
})

// Serve OpenAPI spec
app.ServeOpenAPISpec("/api/openapi.json")
app.ServeSwaggerUI("/api/docs", "/api/openapi.json")
```

### 9. Error Handling

Return errors from handlers:

```go
func getUser(c *echo.Context) (UserResponse, error) {
    id := c.Param("id")
    
    user, err := db.FindUser(id)
    if err != nil {
        // Return HTTP error
        return UserResponse{}, echo.NewHTTPError(404, "user not found")
    }
    
    return user, nil
}
```

Common HTTP errors:

```go
echo.NewHTTPError(400, "bad request")
echo.NewHTTPError(401, "unauthorized")
echo.NewHTTPError(403, "forbidden")
echo.NewHTTPError(404, "not found")
echo.NewHTTPError(409, "conflict")
echo.NewHTTPError(422, "validation failed")
echo.NewHTTPError(500, "internal server error")
```

### 10. Middleware

EchoNext is fully compatible with Echo middleware:

```go
import "github.com/labstack/echo/v5/middleware"

app := echonext.New()

// Global middleware
app.Use(middleware.Logger())
app.Use(middleware.Recover())
app.Use(middleware.CORS())
app.Use(middleware.Gzip())

// Route-specific middleware
app.GET("/admin", adminHandler, echonext.Route{
    Summary: "Admin endpoint",
}, middleware.BasicAuth(func(username, password string, c *echo.Context) (bool, error) {
    return username == "admin" && password == "secret", nil
}))
```

Use EchoNext contrib middleware:

```go
import "github.com/abdussamadbello/echonext/pkg/contrib/middleware"

app.Use(middleware.RequestID())
app.Use(middleware.MetricsMiddleware(metrics))
app.Use(middleware.OTELMiddleware("my-service"))
```

### 11. Route Groups

Organize routes into groups:

```go
app := echonext.New()

// Create API v1 group
v1 := app.Group("/api/v1")

// Add routes to group
v1.GET("/users", listUsers, echonext.Route{...})
v1.POST("/users", createUser, echonext.Route{...})

// Create API v2 group
v2 := app.Group("/api/v2")

// Add routes to v2
v2.GET("/users", listUsersV2, echonext.Route{...})
```

## Type Safety Benefits

### Compile-Time Checks

```go
type Request struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}

// ✅ Correct
func handler(c *echo.Context, req Request) (Response, error) {
    name := req.Name  // Type-safe access
    age := req.Age    // No type assertions needed
    return Response{}, nil
}

// ❌ Won't compile - wrong type
func handler(c *echo.Context, req string) (Response, error) {
    // Compiler catches this mistake!
}
```

### Auto-Complete Support

Your IDE can provide auto-complete for:
- Request fields
- Response fields
- Validation tags
- Route configuration

### Refactoring Safety

Changing field names or types is safe:

```go
// Change field name
type Request struct {
    UserName string `json:"username"`  // Was: Name
}

// Compiler shows all places that need updating
func handler(c *echo.Context, req Request) (Response, error) {
    name := req.UserName  // IDE suggests the new name
}
```

## Echo Compatibility

EchoNext is **fully compatible** with Echo:

```go
app := echonext.New()

// ✅ All Echo methods work
app.Use(echoMiddleware.Logger())
app.Static("/static", "assets")
app.File("/favicon.ico", "public/favicon.ico")

// ✅ Mix EchoNext and standard Echo handlers
app.GET("/typed", typedHandler, echonext.Route{...})
app.GET("/standard", func(c *echo.Context) error {
    return c.String(200, "Standard Echo handler")
})

// ✅ Access underlying Echo instance
echoInstance := app.Echo  // *echo.Echo

// ✅ Use Echo context in handlers
func handler(c *echo.Context, req Request) (Response, error) {
    // All Echo context methods available
    c.Param("id")
    c.QueryParam("filter")
    c.Request().Header.Get("Authorization")
    c.Set("user", user)
    c.Get("user")
}
```

## Best Practices

### 1. Use Descriptive Types

```go
// ❌ Generic types
type Request struct {
    Data map[string]interface{}
}

// ✅ Specific types
type CreateUserRequest struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
}
```

### 2. Add Validation Rules

```go
// ❌ No validation
type Request struct {
    Email string `json:"email"`
}

// ✅ With validation
type Request struct {
    Email string `json:"email" validate:"required,email"`
}
```

### 3. Document Your Routes

```go
// ❌ No documentation
app.POST("/users", createUser)

// ✅ With documentation
app.POST("/users", createUser, echonext.Route{
    Summary:     "Create user",
    Description: "Creates a new user account",
    Tags:        []string{"Users"},
})
```

### 4. Use Appropriate Status Codes

```go
app.POST("/users", createUser, echonext.Route{
    SuccessStatus: 201,  // 201 Created for POST
})

app.DELETE("/users/:id", deleteUser, echonext.Route{
    SuccessStatus: 204,  // 204 No Content for DELETE
})
```

### 5. Handle Errors Gracefully

```go
func handler(c *echo.Context, req Request) (Response, error) {
    result, err := service.Process(req)
    if err != nil {
        // Return appropriate HTTP error
        return Response{}, echo.NewHTTPError(500, err.Error())
    }
    return result, nil
}
```

## Next Steps

- [API Development Guide](../guides/api-development.md) - Build complete APIs
- [Validation Guide](../guides/validation.md) - Master request validation
- [OpenAPI Guide](../guides/openapi.md) - Generate great documentation
- [Error Handling Guide](../guides/error-handling.md) - Handle errors properly
