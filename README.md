# EchoNext

EchoNext is a type-safe wrapper around the Echo web framework that automatically generates OpenAPI specifications and provides request validation. Build robust, well-documented APIs with compile-time type safety.

## Features

- 🔒 **Type-Safe Handlers** - Define handlers with strongly-typed request and response structs
- 📚 **Automatic OpenAPI Generation** - Generate OpenAPI 3.0 specs from your code
- ✅ **Built-in Validation** - Validate requests using struct tags
- 📖 **Swagger UI** - Interactive API documentation out of the box
- 🚀 **Zero Boilerplate** - Focus on business logic, not HTTP details
- 🔌 **Echo Compatible** - Use all Echo middleware and features
- 📁 **File Uploads** - Type-safe file uploads with OpenAPI documentation
- 🔌 **WebSocket Support** - Real-time communication with Hub pattern
- 📊 **GraphQL Integration** - Seamless gqlgen integration with context sharing
- 🛠️ **CLI Tool** - Code generation, hot reload, and project scaffolding

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

## File Uploads

EchoNext provides type-safe file upload support with automatic OpenAPI documentation:

```go
import "github.com/abdussamadbello/echonext/upload"

type AvatarRequest struct {
    File *upload.File `form:"avatar" validate:"required"`
}

type AvatarResponse struct {
    URL string `json:"url"`
}

func uploadAvatar(c echo.Context, req AvatarRequest) (AvatarResponse, error) {
    // Access file metadata
    fmt.Printf("Filename: %s, Size: %d\n", req.File.Filename, req.File.Size)

    // Save the file easily
    if err := req.File.SaveTo("/uploads/" + req.File.Filename); err != nil {
        return AvatarResponse{}, err
    }

    return AvatarResponse{URL: "/uploads/" + req.File.Filename}, nil
}

// Register upload endpoint
app.Upload("/avatar", uploadAvatar, echonext.Route{
    Summary: "Upload avatar image",
})
```

### Multiple Files & Configuration

```go
type DocumentsRequest struct {
    Files []*upload.File `form:"documents" validate:"max=10"`
}

app.Upload("/documents", handler, echonext.Route{
    FileConfig: &echonext.FileUploadConfig{
        MaxFileSize:       10 << 20,  // 10MB per file
        MaxTotalSize:      50 << 20,  // 50MB total
        AllowedMIMETypes:  []string{"image/jpeg", "image/png", "application/pdf"},
        AllowedExtensions: []string{".jpg", ".png", ".pdf"},
        MaxFiles:          5,
    },
})
```

Generate upload handlers with the CLI:

```bash
echonext generate upload avatar
```

See [examples/upload-demo/](examples/upload-demo/) for a complete example.

## WebSocket Support

Type-safe WebSocket handlers with connection management:

```go
import "github.com/abdussamadbello/echonext/websocket"

// Simple handler
func chatHandler(conn *websocket.Connection) error {
    for {
        var msg ChatMessage
        if err := conn.ReadJSON(&msg); err != nil {
            return err
        }
        response := ChatResponse{Text: "Echo: " + msg.Text}
        if err := conn.WriteJSON(response); err != nil {
            return err
        }
    }
}

app.WS("/chat", chatHandler)
```

### Hub Pattern for Broadcasting

```go
type ChatHandler struct {
    hub *websocket.Hub
}

func (h *ChatHandler) OnConnect(conn *websocket.Connection) error {
    h.hub.Register(conn)
    return nil
}

func (h *ChatHandler) OnMessage(conn *websocket.Connection, msg []byte) error {
    return h.hub.Broadcast(msg)  // Broadcast to all connections
}

func (h *ChatHandler) OnDisconnect(conn *websocket.Connection, err error) {
    h.hub.Unregister(conn)
}

// Usage
hub := websocket.NewHub()
go hub.Run()

app.WS("/ws/chat", &ChatHandler{hub: hub})
```

Generate WebSocket handlers with the CLI:

```bash
echonext generate websocket chat
```

See [examples/websocket-demo/](examples/websocket-demo/) for a complete example.

## GraphQL Integration

Seamless integration with gqlgen:

```go
import "github.com/abdussamadbello/echonext/graphql"

app.GraphQL(graphql.Config{
    Path:           "/graphql",
    PlaygroundPath: "/playground",
    Schema:         graph.NewExecutableSchema(graph.Config{
        Resolvers: graph.NewResolver(),
    }),
})
```

### Access Echo Context in Resolvers

```go
func (r *queryResolver) CurrentUser(ctx context.Context) (*model.User, error) {
    echoCtx := graphql.GetEchoContext(ctx)
    userID := echoCtx.Get("user_id").(string)
    return r.userService.GetByID(userID)
}
```

### GraphQL Configuration Options

```go
graphql.Config{
    Path:                "/graphql",
    PlaygroundPath:      "/playground",  // Empty to disable
    Schema:              schema,
    ComplexityLimit:     100,
    QueryCacheSize:      1000,
    EnableIntrospection: true,
}
```

Generate GraphQL boilerplate with the CLI:

```bash
echonext generate graphql
```

See [examples/graphql-demo/](examples/graphql-demo/) for a complete example.

## Code Generation from OpenAPI

Generate EchoNext code from existing OpenAPI specifications:

```bash
# From local file
echonext generate openapi api.yaml

# From URL
echonext generate openapi https://api.example.com/openapi.json

# With options
echonext generate openapi api.yaml --output=./generated --package=api
```

Generated files:
- `models/models.go` - Data models from schema components
- `dto/dto.go` - Request/Response DTOs
- `handlers/handlers.go` - Handler function stubs
- `routes.go` - Route registration

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
├── echonext.go              # Main package implementation
├── echonext_test.go         # Test suite
├── upload/                  # File upload package
│   └── upload.go
├── websocket/               # WebSocket package
│   └── websocket.go
├── graphql/                 # GraphQL integration
│   └── graphql.go
├── cmd/echonext-cli/        # CLI tool
│   ├── commands.go
│   └── generator/           # Code generation templates
├── examples/
│   ├── graphql-demo/        # GraphQL example
│   ├── websocket-demo/      # WebSocket chat example
│   └── upload-demo/         # File upload example
├── pkg/contrib/             # Optional helper packages
└── example/main.go          # Quick start example
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

## Optional Contrib Packages

EchoNext provides optional helper packages in `pkg/contrib/` for common tasks. These are completely optional - you can use the underlying libraries directly if you prefer.

### 📦 Database (`pkg/contrib/database`)

GORM integration helpers with:
- Connection management with retry logic
- Generic Repository[T] pattern
- Transaction utilities
- Migration helpers
- **Atlas integration** for schema migrations

```go
import "github.com/abdussamadbello/echonext/pkg/contrib/database"

cfg := database.DefaultConfig()
db, err := database.Connect(postgres.Open(dsn), cfg)

// Use repository pattern with generics
userRepo := database.NewRepository[User](db)
user, err := userRepo.Find(1)
users, err := userRepo.Where("active = ?", true).FindAll()
```

### 🔄 Database Migrations (Atlas)

EchoNext uses [Atlas](https://atlasgo.io) for database schema management:

```bash
# Initialize Atlas in your project
echonext db init

# Apply migrations
echonext db migrate

# Generate migration from schema changes
echonext db migrate:diff add_users_table

# Check migration status
echonext db migrate:status

# Rollback migrations
echonext db migrate:down --count=1
```

**Declarative Schema** - Define your schema in `schema.hcl`:

```hcl
table "users" {
  schema = schema.public
  column "id" { type = bigserial }
  column "email" { type = varchar(255) }
  primary_key { columns = [column.id] }
}
```

See [CLAUDE.md](CLAUDE.md#database-migrations-with-atlas) for detailed Atlas documentation.

### ⚙️ Config (`pkg/contrib/config`)

Viper integration helpers with:
- Generic config loading
- Environment variable binding
- Hot reload support
- Standard config structures

```go
import "github.com/abdussamadbello/echonext/pkg/contrib/config"

type MyConfig struct {
    App      config.AppConfig      `mapstructure:"app"`
    Database config.DatabaseConfig `mapstructure:"database"`
}

var cfg MyConfig
config.LoadSimple(&cfg)
```

### 🧪 Testing (`pkg/contrib/testing`)

Testing utilities with:
- APIClient for testing endpoints
- FixtureManager for test data
- Test suite with setup/teardown
- Factory pattern for test entities

```go
import echonexttest "github.com/abdussamadbello/echonext/pkg/contrib/testing"

client := echonexttest.NewAPIClient(app)
resp := client.POST("/users", userRequest)
resp.AssertStatus(t, 201).AssertSuccess(t)
```

See [pkg/contrib/README.md](pkg/contrib/README.md) for detailed documentation.

## 🎓 Example Projects

Learn by example! Check out our complete example projects:

### [⚡ Quickstart](example/main.go) - Running Example
Complete working Todo API. Run it now!

```bash
go run example/main.go
# Visit http://localhost:8080/api/docs
```

### [📝 Todo List API](examples/todo-api/) - Beginner
Simple CRUD operations demonstrating the basics of EchoNext.

```bash
echonext init todo-api
echonext generate domain todo
go run ./cmd/api
```

### [📰 Blog API](examples/blog-api/) - Intermediate
Multi-domain blog platform with authentication, search, and relationships.

```bash
echonext init blog-api
echonext generate domain post
echonext generate domain comment
echonext generate domain user
```

### [🛒 E-commerce API](examples/ecommerce-api/) - Advanced
Complete e-commerce platform with orders, payments, and inventory.

```bash
echonext init ecommerce-api
echonext generate domain product
echonext generate domain order
echonext generate domain payment
```

### [🔧 Microservices](examples/microservice/) - Expert
Distributed system with service-to-service communication and events.

See [examples/README.md](examples/README.md) for detailed guides and more examples.

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT License - see the [LICENSE](LICENSE) file for details.

## Roadmap

**Completed:**
- [x] ✅ Database integration helpers (see `pkg/contrib/database`)
- [x] ✅ Configuration management helpers (see `pkg/contrib/config`)
- [x] ✅ Testing utilities (see `pkg/contrib/testing`)
- [x] ✅ Middleware helpers (see `pkg/contrib/middleware`)
- [x] ✅ CLI tool for project generation (`echonext init`)
- [x] ✅ Code generation commands (`echonext generate domain/handler/service/model/dto`)
- [x] ✅ Database management commands (`echonext db init/migrate/seed`)
- [x] ✅ Atlas migration integration for schema management
- [x] ✅ Complete example projects (Todo, Blog, E-commerce, Microservices)

**v1.4.0:**
- [x] ✅ Hot reload dev command (`echonext dev`)
- [x] ✅ Enhanced test runner (`echonext test`)
- [x] ✅ Build automation (`echonext build`)
- [x] ✅ Support for file uploads in OpenAPI spec
- [x] ✅ WebSocket support with type safety
- [x] ✅ GraphQL integration
- [x] ✅ Code generation from OpenAPI spec

**Planned:**
- [ ] Plugin system for custom generators
- [ ] OpenTelemetry integration
- [ ] gRPC support
- [ ] API versioning helpers