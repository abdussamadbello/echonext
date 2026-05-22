# Architecture Overview

This document explains how EchoNext works under the hood and provides a comprehensive view of the codebase structure.

## Project Structure

```
echonext/
├── echonext.go              # Core package - type-safe wrapper around Echo
├── echonext_test.go         # Core package tests
├── go.mod                   # Go module definition
│
├── pkg/                     # Contrib packages (optional utilities)
│   └── contrib/
│       ├── config/          # Viper-based configuration management
│       ├── database/        # GORM integration, repository pattern, Atlas migrations
│       ├── middleware/      # Echo middleware extensions (RequestID, Metrics, OTEL)
│       └── testing/         # Testing utilities (APIClient, Fixtures, Suites)
│
├── cmd/                     # CLI tool
│   └── echonext-cli/
│       ├── main.go          # CLI entry point
│       ├── commands.go      # Cobra command definitions
│       ├── init.go          # Project initialization
│       ├── generators.go    # Code generation logic
│       └── generator/
│           ├── engine.go    # Template engine with Sprig
│           ├── project.go   # Project generator methods
│           └── templates/   # Embedded Go templates
│
├── example/                 # Basic example application
│   └── main.go
│
├── examples/                # Multiple example projects
│   ├── quickstart/          # Beginner example
│   ├── todo-api/            # Basic CRUD example
│   ├── blog-api/            # Relationships example
│   ├── ecommerce-api/       # Advanced transactions
│   ├── microservice/        # Microservices template
│   └── otel-demo/           # Observability example
│
└── docs/                    # Documentation
    ├── architecture/
    ├── guides/
    ├── contrib/
    └── cli/
```

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                         EchoNext App                                 │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌──────────────────┐  ┌─────────────────┐  ┌───────────────────┐  │
│  │   Type-Safe      │  │    OpenAPI      │  │    Validation     │  │
│  │   Handlers       │  │   Generation    │  │     Engine        │  │
│  └──────────────────┘  └─────────────────┘  └───────────────────┘  │
│                                                                      │
│  ┌──────────────────┐  ┌─────────────────┐  ┌───────────────────┐  │
│  │    Security      │  │     Groups      │  │    Response       │  │
│  │    Schemes       │  │   & Routing     │  │    Wrapping       │  │
│  └──────────────────┘  └─────────────────┘  └───────────────────┘  │
│                                                                      │
├─────────────────────────────────────────────────────────────────────┤
│                        Echo Framework                                │
├─────────────────────────────────────────────────────────────────────┤
│                   Standard Library (net/http)                        │
└─────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────┐
│                     Contrib Packages (Optional)                      │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌──────────────────┐  ┌─────────────────┐  ┌───────────────────┐  │
│  │    Database      │  │     Config      │  │    Middleware     │  │
│  │  (GORM + Atlas)  │  │    (Viper)      │  │   (OTEL, etc.)    │  │
│  └──────────────────┘  └─────────────────┘  └───────────────────┘  │
│                                                                      │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │                       Testing Utilities                       │  │
│  │              (APIClient, Fixtures, Suites)                    │  │
│  └──────────────────────────────────────────────────────────────┘  │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

## Core Components

### 1. App Structure

The `App` struct wraps `*echo.Echo` and adds EchoNext functionality:

```go
type App struct {
    *echo.Echo                      // Embedded Echo instance
    spec           *openapi3.T     // OpenAPI specification
    validator      *validator.Validate // Validation engine
    routes         []RouteInfo     // Route metadata
    securitySchemes map[string]*openapi3.SecurityScheme
    globalSecurity []map[string][]string
}
```

**Key responsibilities:**
- Manage OpenAPI spec generation
- Configure validation
- Track route metadata
- Handle security schemes
- Provide type-safe route registration

### 2. Type-Safe Handler Signatures

EchoNext supports multiple handler signatures:

```go
// GET/DELETE without request body
func(c *echo.Context) (ResponseType, error)

// POST/PUT/PATCH with request body
func(c *echo.Context, req RequestType) (ResponseType, error)

// No response body (returns error only)
func(c *echo.Context) error
```

### 3. Handler Wrapper Creation

When you register a typed handler, EchoNext creates a wrapper:

```go
// Your handler
func createUser(c *echo.Context, req CreateUserRequest) (UserResponse, error)

// EchoNext creates this wrapper via createEchoHandler()
func wrapper(c *echo.Context) error {
    // 1. Parse and validate request
    var req CreateUserRequest
    if err := parseAndValidate(c, &req); err != nil {
        return err
    }

    // 2. Call your handler
    resp, err := createUser(c, req)
    if err != nil {
        return err
    }

    // 3. Wrap and serialize response
    return c.JSON(statusCode, Response[UserResponse]{
        Success: true,
        Data:    resp,
    })
}
```

### 4. Request Processing Pipeline

```
HTTP Request
    │
    ▼
Echo Middleware (Logger, Recover, CORS, etc.)
    │
    ▼
Contrib Middleware (RequestID, OTEL, Metrics)
    │
    ▼
EchoNext Handler Wrapper
    │
    ├── 1. Bind request (JSON body / Query params / Path params)
    │
    ├── 2. Validate struct (go-playground/validator)
    │
    ├── 3. Call user handler (typed)
    │
    ├── 4. Wrap response (Response[T])
    │
    └── 5. Serialize to JSON
    │
    ▼
HTTP Response
```

### 5. Response Wrapping

All responses are wrapped in a standard format:

```go
type Response[T any] struct {
    Data    T      `json:"data,omitempty"`
    Error   string `json:"error,omitempty"`
    Success bool   `json:"success"`
}
```

## OpenAPI Generation

### Generation Process

```
1. Route Registration
   ├── Capture handler type via reflection
   ├── Extract request/response types
   └── Store in routes metadata

2. GenerateOpenAPISpec() called
   ├── Iterate all registered routes
   └── For each route: addRouteToSpec()
       ├── Extract path parameters
       ├── Generate request schema
       ├── Generate response schema
       ├── Apply security requirements
       └── Add to OpenAPI paths

3. Schema Generation
   └── generateSchema() with circular reference detection
       ├── Primitive types (string, int, float, bool)
       ├── Collections (slice, map)
       ├── Structs (fields from JSON tags)
       └── Validation constraints (from validate tags)
```

### Schema Generation from Struct Types

```go
func generateSchema(t reflect.Type) *openapi3.Schema {
    schema := &openapi3.Schema{
        Type:       "object",
        Properties: make(map[string]*openapi3.SchemaRef),
    }

    for i := 0; i < t.NumField(); i++ {
        field := t.Field(i)

        jsonTag := field.Tag.Get("json")
        validateTag := field.Tag.Get("validate")

        fieldSchema := generateFieldSchema(field.Type, validateTag)
        schema.Properties[jsonTag] = openapi3.NewSchemaRef("", fieldSchema)

        if strings.Contains(validateTag, "required") {
            schema.Required = append(schema.Required, jsonTag)
        }
    }

    return schema
}
```

### Validation to OpenAPI Mapping

| Validation Tag | OpenAPI Schema |
|---------------|----------------|
| `required` | `required: [field]` |
| `min=N` | `minimum: N` |
| `max=N` | `maximum: N` |
| `email` | `format: email` |
| `url` | `format: uri` |
| `oneof=a b c` | `enum: [a, b, c]` |

## Security Scheme Support

### Supported Security Types

```go
// Bearer Token (JWT)
app.AddSecurityScheme("bearerAuth", echonext.SecurityScheme{
    Type:         echonext.SecurityHTTP,
    Scheme:       "bearer",
    BearerFormat: "JWT",
})

// API Key
app.AddSecurityScheme("apiKey", echonext.SecurityScheme{
    Type: echonext.SecurityAPIKey,
    In:   "header",
    Name: "X-API-Key",
})

// OAuth2
app.AddSecurityScheme("oauth2", echonext.SecurityScheme{
    Type: echonext.SecurityOAuth2,
    Flows: &echonext.OAuthFlows{
        AuthorizationCode: &echonext.OAuthFlow{
            AuthorizationURL: "https://auth.example.com/authorize",
            TokenURL:         "https://auth.example.com/token",
            Scopes: map[string]string{
                "read":  "Read access",
                "write": "Write access",
            },
        },
    },
})
```

### Security Resolution

```
Route Security Resolution Priority:

1. Route.Security (if specified)     → Use only these
2. Route.DisableGlobalSecurity       → Use none (public)
3. Global security (if configured)   → Use global
4. Default                           → No security required
```

## Contrib Packages Architecture

### Design Philosophy

Contrib packages are **completely optional** and provide convenience wrappers around popular libraries without vendor lock-in.

### pkg/contrib/database

```
database/
├── database.go      # Connection management, pool config
├── repository.go    # Generic Repository[T] pattern
├── transaction.go   # WithTx, WithTxResult utilities
├── migration.go     # GORM auto-migration helpers
├── atlas.go         # Atlas CLI wrapper for migrations
└── database_test.go # Tests
```

**Key Patterns:**
- **Repository Pattern**: Generic CRUD interface
- **Transaction Management**: Automatic rollback on panic
- **Atlas Integration**: Declarative schema migrations

### pkg/contrib/config

```
config/
├── config.go        # Load[T], Watch[T] functions
├── structs.go       # Standard config structures
└── config_test.go   # Tests
```

**Key Features:**
- Generic configuration loading with type safety
- Environment variable binding
- Hot reload support

### pkg/contrib/middleware

```
middleware/
├── requestid.go     # UUID-based request ID tracking
├── metrics.go       # Request metrics collection
├── logger.go        # Structured logging
├── otel.go          # OpenTelemetry integration
└── middleware_test.go # Comprehensive tests
```

**Middleware Chain Order:**
```
Request → RequestID → OTEL → Metrics → Logger → Handler
```

### pkg/contrib/testing

```
testing/
├── client.go        # APIClient for HTTP testing
├── fixtures.go      # Test data management
├── suite.go         # Suite, IntegrationSuite
└── client_test.go   # Tests
```

**Key Components:**
- **APIClient**: Fluent HTTP client (`client.WithAuth("token").GET("/users")`)
- **Response**: Built-in assertions (`resp.AssertStatus(t, 200)`)
- **FixtureManager**: Database test data
- **Factory[T]**: Generic entity factory

## CLI Tool Architecture

### Command Structure

```
echonext
├── init              # Initialize new project
├── generate          # Generate code
│   ├── domain        # Full domain (model, service, handler, dto)
│   ├── handler       # HTTP handler only
│   ├── service       # Service layer only
│   ├── model         # GORM model only
│   ├── dto           # Request/Response DTOs
│   ├── middleware    # Custom middleware
│   └── otel          # OpenTelemetry setup
├── db                # Database commands (Atlas)
│   ├── init          # Initialize Atlas setup
│   ├── migrate       # Apply migrations
│   ├── migrate:status # Check status
│   ├── migrate:new   # Create empty migration
│   ├── migrate:diff  # Generate from schema changes
│   ├── migrate:down  # Rollback
│   ├── migrate:lint  # Lint migrations
│   └── schema:inspect # Inspect schema
└── version           # Version info
```

### Template Engine

```go
// generator/engine.go
type TemplateEngine struct {
    embedded embed.FS           // Built-in templates
    userDir  string             // ~/.echonext/templates/
    funcMap  template.FuncMap   // Sprig + custom functions
}
```

**Template Resolution:**
1. Check `~/.echonext/templates/` (user overrides)
2. Fall back to embedded templates

**Custom Template Functions:**
- `pluralize` - "user" → "users"
- `singularize` - "users" → "user"
- `pascalCase` - "user_account" → "UserAccount"
- `lowerCamelCase` - "user_account" → "userAccount"

### Project Data Structure

```go
type ProjectData struct {
    Name         string // Project name (e.g., "my-api")
    Module       string // Go module (e.g., "github.com/user/my-api")
    Template     string // Template type (standard, minimal)
    EchoNextPath string // Local echonext path for development
}
```

## Design Patterns Used

| Pattern | Usage |
|---------|-------|
| **Repository** | Generic CRUD in `database.Repository[T]` |
| **Builder** | Fluent API in `APIClient.WithAuth().GET()` |
| **Decorator** | Middleware wraps handlers |
| **Factory** | `NewRepository[T]`, `NewAPIClient`, `NewFactory[T]` |
| **Template Method** | Transaction management (`WithTx`, `WithTxResult`) |
| **Functional Options** | Configuration (`LoadOptions`, `OTELConfig`) |
| **Generics** | Type-safe handlers, repositories, factories |

## Component Interactions

### Request Flow with All Components

```
┌─────────────────────────────────────────────────────────────────────┐
│                           HTTP Request                               │
└─────────────────────────────────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────┐
│ Echo Router                                                          │
└─────────────────────────────────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────┐
│ Middleware Chain                                                     │
│ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐    │
│ │  Recover    │→│  RequestID  │→│    OTEL     │→│   Metrics   │    │
│ └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘    │
└─────────────────────────────────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────┐
│ EchoNext Handler Wrapper (createEchoHandler)                        │
│ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐    │
│ │    Bind     │→│  Validate   │→│   Handler   │→│    Wrap     │    │
│ └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘    │
└─────────────────────────────────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────┐
│ Business Logic (User Handler)                                        │
│ ┌─────────────────────────────────────────────────────────────────┐ │
│ │  Service Layer (uses Repository, may call external APIs)        │ │
│ └─────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        JSON Response[T]                              │
└─────────────────────────────────────────────────────────────────────┘
```

## Performance Considerations

### Minimal Runtime Overhead

- **Type inspection once** - During route registration, not per request
- **Reflection minimal** - Only for binding and validation
- **No runtime code generation** - All wrappers created at startup
- **Direct Echo integration** - Minimal abstraction overhead

### Benchmarks

Typical overhead compared to raw Echo:

```
BenchmarkRawEcho       100000    10234 ns/op
BenchmarkEchoNext       95000    10789 ns/op  (~5% overhead)
```

## Key Dependencies

### Core Framework

| Package | Purpose |
|---------|---------|
| `github.com/labstack/echo/v5` | Web framework |
| `github.com/getkin/kin-openapi` | OpenAPI spec generation |
| `github.com/go-playground/validator/v10` | Validation |

### Contrib Packages

| Package | Purpose |
|---------|---------|
| `github.com/spf13/viper` | Configuration management |
| `gorm.io/gorm` | ORM |
| `go.opentelemetry.io/*` | Observability/tracing |
| `github.com/google/uuid` | UUID generation |

### CLI Tool

| Package | Purpose |
|---------|---------|
| `github.com/spf13/cobra` | CLI framework |
| `github.com/Masterminds/sprig/v3` | Template functions |

## Extensibility Points

1. **Custom Middleware** - Implement `echo.MiddlewareFunc`
2. **Custom Validators** - Register with `validator.Validate`
3. **Template Customization** - Override in `~/.echonext/templates/`
4. **Repository Extension** - Extend `BaseRepository[T]`
5. **Security Schemes** - Add custom via `AddSecurityScheme()`
6. **Configuration Types** - Extend with any type via generics

## See Also

- [Design Philosophy](./philosophy.md)
- [Type System](./types.md)
- [OpenAPI Generation](./openapi-generation.md)
- [Database Documentation](../contrib/database.md)
- [Testing Documentation](../contrib/testing.md)
- [Middleware Documentation](../contrib/middleware.md)
