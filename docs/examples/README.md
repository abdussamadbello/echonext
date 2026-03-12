# Example Projects

Learn EchoNext through complete, working examples. Each example demonstrates different features and complexity levels.

## Available Examples

### 1. 🚀 Quickstart (Running Example)

**Location:** `example/main.go` (repository root)  
**Complexity:** Beginner  
**Type:** Working code you can run immediately

A complete Todo API demonstrating all core EchoNext features.

**Features:**
- Type-safe CRUD operations
- Request validation
- OpenAPI documentation
- Swagger UI
- Error handling

**Run it now:**
```bash
cd /path/to/echonext
go run example/main.go

# Visit http://localhost:8080/api/docs
```

**What you'll learn:**
- Basic EchoNext setup
- Handler signatures
- Route configuration
- Validation tags
- OpenAPI generation

[View Source](../../example/main.go) | [Read Guide](../../examples/quickstart/README.md)

---

### 2. 📝 Todo List API

**Location:** `examples/todo-api/`  
**Complexity:** Beginner  
**Type:** Generate and build yourself

Simple CRUD API with a single domain.

**Features:**
- Single domain (todos)
- In-memory storage
- Basic validation
- Complete CRUD operations

**Build it:**
```bash
echonext init todo-api --module=github.com/username/todo-api
cd todo-api
echonext generate domain todo
echonext db init
go mod tidy
go run ./cmd/api
```

**What you'll learn:**
- CLI project generation
- Domain structure
- Database integration
- Basic patterns

[Read Guide](../../examples/todo-api/README.md)

---

### 3. 📰 Blog API

**Location:** `examples/blog-api/`  
**Complexity:** Intermediate  
**Type:** Generate and build yourself

Multi-domain blog platform with relationships.

**Features:**
- Multiple domains (posts, comments, users, categories)
- Domain relationships (one-to-many, many-to-many)
- Authentication patterns
- Search and filtering
- File uploads

**Build it:**
```bash
echonext init blog-api
cd blog-api
echonext generate domain post
echonext generate domain comment
echonext generate domain user
echonext generate domain category
echonext generate middleware auth
echonext db init
go run ./cmd/api
```

**What you'll learn:**
- Multi-domain architecture
- Relationships between entities
- Authentication middleware
- Query parameters
- File handling

[Read Guide](../../examples/blog-api/README.md)

---

### 4. 🛒 E-commerce API

**Location:** `examples/ecommerce-api/`  
**Complexity:** Advanced  
**Type:** Generate and build yourself

Complete e-commerce backend with complex business logic.

**Features:**
- Multiple domains (products, orders, payments, inventory, customers, cart)
- Transaction handling
- Payment integration patterns
- Order workflows
- Inventory tracking
- Admin endpoints

**Build it:**
```bash
echonext init ecommerce-api
cd ecommerce-api
echonext generate domain product
echonext generate domain order
echonext generate domain payment
echonext generate domain inventory
echonext generate domain customer
echonext generate domain cart
echonext db init
go run ./cmd/api
```

**What you'll learn:**
- Complex domain modeling
- Transaction management
- Business workflows
- Role-based access
- Advanced patterns

[Read Guide](../../examples/ecommerce-api/README.md)

---

### 5. 🔧 Microservices Template

**Location:** `examples/microservice/`  
**Complexity:** Expert  
**Type:** Generate and build yourself

Distributed microservices architecture.

**Features:**
- Service-to-service communication
- Event-driven architecture
- Message queue integration
- Distributed tracing
- Service discovery patterns
- Health checks
- Monitoring

**Build it:**
```bash
# Create multiple services
echonext init user-service
echonext init order-service
echonext init notification-service

# Each service follows same pattern
cd user-service
echonext generate domain user
echonext db init
```

**What you'll learn:**
- Microservices architecture
- Inter-service communication
- Event-driven patterns
- Observability
- Service mesh concepts

[Read Guide](../../examples/microservice/README.md)

---

### 6. 📊 OpenTelemetry Demo

**Location:** `examples/otel-demo/`  
**Complexity:** Intermediate  
**Type:** Working code example

Distributed tracing and observability.

**Features:**
- OTEL initialization
- Automatic request tracing
- Traced HTTP clients
- Span events and attributes
- Request ID correlation
- Trace context propagation
- Integration with Jaeger

**Run it:**
```bash
# Start Jaeger (trace viewer)
docker run -d -p 16686:16686 -p 4317:4317 jaegertracing/all-in-one

# Run the example
go run examples/otel-demo/main.go

# Visit http://localhost:8080/api/docs for API
# Visit http://localhost:16686 for Jaeger UI
```

**What you'll learn:**
- OpenTelemetry setup
- Distributed tracing
- Observability best practices
- Monitoring integration

[View Source](../../examples/otel-demo/main.go) | [Read Guide](../../examples/otel-demo/README.md)

---

## Learning Path

### Path 1: Complete Beginner

1. **Start:** Run the Quickstart example
2. **Next:** Build the Todo API
3. **Then:** Try the Blog API
4. **Practice:** Build your own simple API

### Path 2: Experienced Go Developer

1. **Start:** Review the Quickstart
2. **Next:** Build the E-commerce API
3. **Explore:** OpenTelemetry Demo
4. **Advanced:** Microservices Template

### Path 3: Learning Specific Features

**Want to learn validation?**
→ Start with Todo API

**Want to learn relationships?**
→ Build the Blog API

**Want to learn transactions?**
→ Try the E-commerce API

**Want to learn observability?**
→ Run the OTEL Demo

## Common Patterns

All examples demonstrate these patterns:

### Project Structure

```
project/
├── cmd/
│   └── api/           # HTTP server entrypoint
├── domain/            # Business domains
│   └── user/
│       ├── model.go   # GORM model
│       ├── service.go # Business logic
│       ├── handler.go # HTTP handlers
│       └── dto.go     # Request/Response types
├── internal/
│   ├── config/        # Configuration
│   ├── database/      # DB setup
│   └── middleware/    # Custom middleware
├── configs/           # Config files
└── tests/             # Tests
```

### Handler Pattern

```go
// 1. Define types
type CreateRequest struct {
    Name string `json:"name" validate:"required,min=2"`
}

type Response struct {
    ID   uint   `json:"id"`
    Name string `json:"name"`
}

// 2. Implement handler
func create(c echo.Context, req CreateRequest) (Response, error) {
    // Business logic
    result := service.Create(req)
    return result, nil
}

// 3. Register route
app.POST("/resource", create, echonext.Route{
    Summary: "Create resource",
    Tags:    []string{"Resources"},
})
```

### Service Pattern

```go
type Service struct {
    db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
    return &Service{db: db}
}

func (s *Service) Create(req CreateRequest) (*Model, error) {
    model := &Model{Name: req.Name}
    if err := s.db.Create(model).Error; err != nil {
        return nil, err
    }
    return model, nil
}
```

### Repository Pattern

```go
import "github.com/abdussamadbello/echonext/pkg/contrib/database"

repo := database.NewRepository[User](db)

// CRUD operations
user, err := repo.Find(1)
users, err := repo.FindAll()
err = repo.Create(&user)
err = repo.Update(&user)
err = repo.Delete(1)

// Queries
users, err := repo.Where("active = ?", true).FindAll()
user, err := repo.Where("email = ?", email).First()
```

## Running Examples

### Prerequisites

```bash
# Install Go 1.24+
go version

# Install EchoNext
go get github.com/abdussamadbello/echonext

# Install CLI (optional but recommended)
go install github.com/abdussamadbello/echonext/cmd/echonext-cli@latest
```

### Running the Quickstart

```bash
git clone https://github.com/abdussamadbello/echonext
cd echonext
go run example/main.go
```

### Generating New Examples

```bash
# Create from scratch
echonext init myexample
cd myexample
echonext generate domain myentity
go mod tidy
go run ./cmd/api
```

## Example Code Snippets

### Complete Handler Example

```go
package user

import (
    "github.com/abdussamadbello/echonext"
    "github.com/labstack/echo/v5"
)

type Handler struct {
    service *Service
}

func NewHandler(service *Service) *Handler {
    return &Handler{service: service}
}

func (h *Handler) Register(app *echonext.App) {
    app.POST("/users", h.Create, echonext.Route{
        Summary:       "Create user",
        Tags:          []string{"Users"},
        SuccessStatus: 201,
    })
    
    app.GET("/users/:id", h.Get, echonext.Route{
        Summary: "Get user by ID",
        Tags:    []string{"Users"},
    })
    
    app.PUT("/users/:id", h.Update, echonext.Route{
        Summary: "Update user",
        Tags:    []string{"Users"},
    })
    
    app.DELETE("/users/:id", h.Delete, echonext.Route{
        Summary:       "Delete user",
        Tags:          []string{"Users"},
        SuccessStatus: 204,
    })
}

func (h *Handler) Create(c echo.Context, req CreateUserRequest) (UserResponse, error) {
    user, err := h.service.Create(req)
    if err != nil {
        return UserResponse{}, echo.NewHTTPError(500, err.Error())
    }
    return ToUserResponse(user), nil
}

func (h *Handler) Get(c echo.Context) (UserResponse, error) {
    id := parseID(c.Param("id"))
    user, err := h.service.GetByID(id)
    if err != nil {
        return UserResponse{}, echo.NewHTTPError(404, "user not found")
    }
    return ToUserResponse(user), nil
}
```

### Complete Service Example

```go
package user

import "gorm.io/gorm"

type Service struct {
    db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
    return &Service{db: db}
}

func (s *Service) Create(req CreateUserRequest) (*User, error) {
    user := &User{
        Name:  req.Name,
        Email: req.Email,
    }
    
    if err := s.db.Create(user).Error; err != nil {
        return nil, err
    }
    
    return user, nil
}

func (s *Service) GetByID(id uint) (*User, error) {
    var user User
    if err := s.db.First(&user, id).Error; err != nil {
        return nil, err
    }
    return &user, nil
}

func (s *Service) List(page, limit int) ([]User, int64, error) {
    var users []User
    var total int64
    
    offset := (page - 1) * limit
    
    if err := s.db.Model(&User{}).Count(&total).Error; err != nil {
        return nil, 0, err
    }
    
    if err := s.db.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
        return nil, 0, err
    }
    
    return users, total, nil
}
```

## Next Steps

1. **Run an example** - Start with the Quickstart
2. **Build one yourself** - Use the CLI to generate a project
3. **Customize** - Modify examples for your needs
4. **Share** - Contribute your own examples!

## Contributing Examples

Have a great example? We'd love to include it!

1. Create your example project
2. Add a detailed README
3. Submit a pull request
4. Follow the [Contributing Guide](../contributing/guide.md)

## Additional Resources

- [Quick Start Guide](../getting-started/quickstart.md)
- [API Development Guide](../guides/api-development.md)
- [CLI Overview](../cli/overview.md)
- [Contrib Packages](../contrib/database.md)
