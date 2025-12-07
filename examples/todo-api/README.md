# Todo List API - EchoNext Example

A simple but complete Todo List API built with EchoNext, demonstrating CRUD operations, validation, and OpenAPI documentation.

## 🚀 Quick Start

This example shows what you get when you run:

```bash
echonext init todo-api
echonext generate domain todo
echonext db init
```

## 📁 Project Structure

```
todo-api/
├── cmd/api/main.go           # API server entry point
├── domain/todo/              # Todo domain (generated)
│   ├── model.go             # GORM model
│   ├── service.go           # Business logic
│   ├── handler.go           # HTTP handlers
│   └── dto.go               # Request/Response DTOs
├── internal/
│   ├── config/config.go     # Configuration
│   ├── database/database.go # Database setup
│   └── server/server.go     # Server setup
├── configs/
│   ├── development.yaml     # Dev config
│   └── production.yaml      # Prod config
└── go.mod
```

## 🏃 Run the Example

```bash
# 1. Install dependencies
go mod tidy

# 2. Run with Docker (includes PostgreSQL)
docker-compose up

# Or run locally (requires PostgreSQL)
# Update configs/development.yaml with your DB connection
go run ./cmd/api

# 3. Visit API docs
open http://localhost:8080/api/docs
```

## 📝 API Endpoints

| Method | Endpoint       | Description      |
|--------|----------------|------------------|
| POST   | /todos         | Create todo      |
| GET    | /todos         | List all todos   |
| GET    | /todos/:id     | Get todo by ID   |
| PUT    | /todos/:id     | Update todo      |
| DELETE | /todos/:id     | Delete todo      |
| GET    | /health        | Health check     |
| GET    | /api/docs      | Swagger UI       |
| GET    | /api/openapi.json | OpenAPI spec  |

## 🧪 Try It Out

### Create a Todo

```bash
curl -X POST http://localhost:8080/todos \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Buy groceries",
    "description": "Milk, eggs, bread",
    "completed": false
  }'
```

### Get All Todos

```bash
curl http://localhost:8080/todos
```

### Update a Todo

```bash
curl -X PUT http://localhost:8080/todos/1 \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Buy groceries",
    "completed": true
  }'
```

### Delete a Todo

```bash
curl -X DELETE http://localhost:8080/todos/1
```

## 🎓 What This Example Demonstrates

### 1. **Type-Safe Handlers**
```go
func (h *Handler) Create(c echo.Context, req CreateTodoRequest) (TodoResponse, error) {
    // Request is automatically validated and bound
    // Response is automatically JSON-encoded
}
```

### 2. **Automatic Validation**
```go
type CreateTodoRequest struct {
    Title       string `json:"title" validate:"required,min=3,max=200"`
    Description string `json:"description" validate:"max=1000"`
    Completed   bool   `json:"completed"`
}
```

### 3. **OpenAPI Generation**
All endpoints are automatically documented with:
- Request/response schemas
- Validation rules
- Error responses
- Interactive Swagger UI

### 4. **Clean Architecture**
- **Model Layer**: GORM models with timestamps
- **Service Layer**: Business logic and database operations
- **Handler Layer**: HTTP request handling
- **DTO Layer**: Request/Response transformation

### 5. **Database Integration**
- GORM for ORM
- Auto-migration support
- Connection pooling
- Transaction handling

## 🔧 Customization Points

### Add New Fields to Todo

Edit `domain/todo/model.go`:
```go
type Todo struct {
    // ... existing fields ...
    Priority   string     `gorm:"type:varchar(20)"`
    DueDate    *time.Time `gorm:"index"`
    Tags       []string   `gorm:"type:text[]"`
}
```

### Add Custom Business Logic

Edit `domain/todo/service.go`:
```go
func (s *Service) CompleteAll() error {
    return s.db.Model(&Todo{}).Where("completed = ?", false).
        Update("completed", true).Error
}
```

### Add Custom Endpoints

Edit `domain/todo/handler.go`:
```go
app.POST("/todos/complete-all", h.CompleteAll, echonext.Route{
    Summary: "Mark all todos as completed",
    Tags:    []string{"Todos"},
})
```

## 📦 Using Contrib Packages

This example uses EchoNext contrib packages:

### Database Package
```go
import "github.com/abdussamadbello/echonext/pkg/contrib/database"

cfg := database.DefaultConfig()
db, err := database.Connect(postgres.Open(dsn), cfg)
```

### Config Package
```go
import "github.com/abdussamadbello/echonext/pkg/contrib/config"

var cfg MyConfig
config.LoadSimple(&cfg)
```

### Testing Package
```go
import echonexttest "github.com/abdussamadbello/echonext/pkg/contrib/testing"

client := echonexttest.NewAPIClient(app)
resp := client.POST("/todos", newTodo)
resp.AssertStatus(t, 201)
```

## 🧪 Running Tests

```bash
# Run all tests
go test ./...

# With coverage
go test -cover ./...

# Run specific tests
go test ./domain/todo/...
```

## 🐳 Docker Deployment

```bash
# Build
docker build -t todo-api -f infrastructure/Dockerfile.api .

# Run
docker-compose up
```

## 📖 Next Steps

1. **Add Authentication**: Generate an auth middleware
   ```bash
   echonext generate middleware auth
   ```

2. **Add More Domains**: Generate user management
   ```bash
   echonext generate domain user
   ```

3. **Add Background Jobs**: Use the worker executable
   ```bash
   go run ./cmd/worker
   ```

4. **Deploy to Production**: Use the provided Docker files

## 🤝 Learn More

- [EchoNext Documentation](../../README.md)
- [CLI Tool Guide](../../cmd/echonext-cli/)
- [Contrib Packages](../../pkg/contrib/)
- [More Examples](../)
