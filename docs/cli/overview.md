# CLI Overview

The EchoNext CLI is a powerful code generation tool that helps you scaffold projects, generate domains, and manage databases quickly.

## Installation

```bash
go install github.com/abdussamadbello/echonext/cmd/echonext-cli@latest
```

Verify installation:

```bash
echonext --version
```

## Available Commands

### Project Management

- `echonext init` - Initialize a new EchoNext project
- `echonext dev` - Start development server with hot reload (planned)

### Code Generation

- `echonext generate domain` - Generate complete domain (model, service, handler, DTOs)
- `echonext generate handler` - Generate HTTP handler
- `echonext generate service` - Generate service layer
- `echonext generate model` - Generate GORM model
- `echonext generate dto` - Generate request/response DTOs
- `echonext generate middleware` - Generate custom middleware
- `echonext generate otel` - Generate OpenTelemetry setup

### Database Management

- `echonext db init` - Initialize database and migrations
- `echonext db migrate` - Run database migrations
- `echonext db seed` - Seed database with test data

## Quick Start

Create a new project in seconds:

```bash
# Initialize project
echonext init myapp --module=github.com/username/myapp

# Navigate to project
cd myapp

# Generate a domain
echonext generate domain user

# Initialize database
echonext db init

# Run the app
go mod tidy
go run ./cmd/api
```

## Command Reference

### echonext init

Create a new EchoNext project with complete structure.

```bash
echonext init PROJECT_NAME [flags]
```

**Flags:**
- `--module` - Go module name (default: current directory name)
- `--dir` - Output directory (default: current directory)

**Example:**
```bash
echonext init blog-api --module=github.com/myuser/blog-api
```

**Generated Structure:**
```
blog-api/
├── cmd/
│   ├── api/          # HTTP server
│   ├── worker/       # Background worker
│   ├── cli/          # CLI tool
│   └── migration/    # DB migrations
├── domain/           # Business domains
├── internal/
│   ├── config/       # Configuration
│   ├── database/     # Database setup
│   ├── middleware/   # Custom middleware
│   └── server/       # Server setup
├── configs/          # Config files
├── tests/            # Tests
├── go.mod
└── README.md
```

### echonext generate domain

Generate a complete domain with model, service, handler, and DTOs.

```bash
echonext generate domain ENTITY_NAME [flags]
```

**Example:**
```bash
echonext generate domain product
```

**Generates:**
```
domain/product/
├── model.go       # GORM model
├── service.go     # Business logic
├── handler.go     # HTTP handlers
└── dto.go         # Request/Response types
```

**Files include:**
- ✅ CRUD operations
- ✅ Type-safe handlers
- ✅ Validation rules
- ✅ OpenAPI documentation
- ✅ Error handling

### echonext generate handler

Generate HTTP handler for an entity.

```bash
echonext generate handler ENTITY_NAME
```

**Example:**
```bash
echonext generate handler product
```

### echonext generate service

Generate service layer for business logic.

```bash
echonext generate service ENTITY_NAME
```

**Example:**
```bash
echonext generate service product
```

### echonext generate model

Generate GORM model.

```bash
echonext generate model ENTITY_NAME
```

**Example:**
```bash
echonext generate model product
```

### echonext generate dto

Generate request/response DTOs.

```bash
echonext generate dto ENTITY_NAME
```

**Example:**
```bash
echonext generate dto product
```

### echonext generate middleware

Generate custom middleware.

```bash
echonext generate middleware MIDDLEWARE_NAME
```

**Example:**
```bash
echonext generate middleware auth
echonext generate middleware ratelimit
```

### echonext generate otel

Generate OpenTelemetry instrumentation setup.

```bash
echonext generate otel
```

**Generates:**
- OTEL initialization code
- Configuration setup
- Middleware integration
- Traced HTTP client

### echonext db init

Initialize database migrations.

```bash
echonext db init
```

**Creates:**
```
migrations/
└── 000001_initial.sql
```

### echonext db migrate

Run database migrations.

```bash
echonext db migrate
```

### echonext db seed

Seed database with test data.

```bash
echonext db seed
```

## Common Workflows

### Starting a New Project

```bash
# 1. Create project
echonext init myapp --module=github.com/user/myapp
cd myapp

# 2. Generate domains
echonext generate domain user
echonext generate domain product
echonext generate domain order

# 3. Setup database
echonext db init

# 4. Run
go mod tidy
go run ./cmd/api
```

### Adding a New Feature

```bash
# Generate domain
echonext generate domain comment

# Customize the generated files
# - Edit domain/comment/model.go for database schema
# - Edit domain/comment/service.go for business logic
# - Edit domain/comment/handler.go for endpoints
# - Edit domain/comment/dto.go for request/response types

# Run migrations
echonext db migrate

# Test
go run ./cmd/api
```

### Adding Middleware

```bash
# Generate middleware
echonext generate middleware auth

# Edit internal/middleware/auth.go
# Add logic for authentication

# Use in main.go
# app.Use(middleware.Auth())
```

### Adding OpenTelemetry

```bash
# Generate OTEL setup
echonext generate otel

# Configure environment variables
export OTEL_SERVICE_NAME="myapp"
export OTEL_EXPORTER_OTLP_ENDPOINT="localhost:4317"

# Use in main.go
# shutdown := otel.MustInit(ctx, otel.DefaultConfig())
# defer shutdown()
# app.Use(middleware.OTELMiddleware("myapp"))
```

## Generated Code Structure

### Domain Structure

When you run `echonext generate domain user`, you get:

**model.go:**
```go
package user

import (
    "gorm.io/gorm"
    "time"
)

type User struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
    
    // TODO: Add your fields here
    Name  string `gorm:"not null" json:"name"`
    Email string `gorm:"unique;not null" json:"email"`
}
```

**service.go:**
```go
package user

import "gorm.io/gorm"

type Service struct {
    db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
    return &Service{db: db}
}

func (s *Service) Create(user *User) error {
    return s.db.Create(user).Error
}

func (s *Service) GetByID(id uint) (*User, error) {
    var user User
    err := s.db.First(&user, id).Error
    return &user, err
}

// ... more CRUD methods
```

**handler.go:**
```go
package user

import (
    "github.com/abdussamadbello/echonext"
    "github.com/labstack/echo/v4"
)

type Handler struct {
    service *Service
}

func NewHandler(service *Service) *Handler {
    return &Handler{service: service}
}

func (h *Handler) Register(app *echonext.App) {
    app.POST("/users", h.Create, echonext.Route{
        Summary: "Create user",
        Tags:    []string{"Users"},
    })
    
    app.GET("/users/:id", h.Get, echonext.Route{
        Summary: "Get user",
        Tags:    []string{"Users"},
    })
    
    // ... more routes
}

func (h *Handler) Create(c echo.Context, req CreateUserRequest) (UserResponse, error) {
    // Implementation
}
```

## Customization

### Modify Generated Files

All generated files have `// TODO` comments showing where to add your custom logic:

```go
type User struct {
    ID uint `gorm:"primaryKey"`
    
    // TODO: Add your fields here
    Name  string `gorm:"not null"`
    Email string `gorm:"unique;not null"`
}
```

### Add Custom Templates

(Feature planned for future release)

## Tips and Best Practices

1. **Use Consistent Naming** - Use singular names for entities (user, not users)
2. **Generate, Then Customize** - Use CLI for scaffolding, then add business logic
3. **One Domain Per Entity** - Keep domains focused on single business entities
4. **Use Database Migrations** - Don't modify existing migrations, create new ones
5. **Test After Generation** - Run and test generated code immediately

## Troubleshooting

### Command Not Found

Add Go bin to your PATH:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

### Generation Fails

Make sure you're in a Go module:

```bash
go mod init github.com/user/project
```

### Database Init Fails

Check that you have database configuration in `configs/config.yaml`.

## See Also

- [Project Initialization Guide](./init.md)
- [Code Generation Guide](./generate.md)
- [Database Commands Guide](./database.md)
- [Example Projects](../examples/README.md)
