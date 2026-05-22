# Quick Start

Build your first type-safe API with EchoNext in just 5 minutes!

## Your First API

Let's create a simple Todo API with full OpenAPI documentation.

### Step 1: Create a New Project

```bash
# Create a new directory
mkdir todo-api && cd todo-api

# Initialize a Go module
go mod init github.com/yourusername/todo-api

# Install EchoNext
go get github.com/abdussamadbello/echonext
```

### Step 2: Create main.go

Create a `main.go` file:

```go
package main

import (
    "github.com/abdussamadbello/echonext"
    "github.com/labstack/echo/v5"
)

// Define your data types
type Todo struct {
    ID        int    `json:"id"`
    Title     string `json:"title"`
    Completed bool   `json:"completed"`
}

type CreateTodoRequest struct {
    Title string `json:"title" validate:"required,min=3,max=100"`
}

type UpdateTodoRequest struct {
    Title     *string `json:"title,omitempty" validate:"omitempty,min=3,max=100"`
    Completed *bool   `json:"completed,omitempty"`
}

// In-memory storage (for demo purposes)
var todos = []Todo{
    {ID: 1, Title: "Learn EchoNext", Completed: false},
    {ID: 2, Title: "Build an API", Completed: false},
}
var nextID = 3

func main() {
    // Create EchoNext app
    app := echonext.New()

    // Set API information
    app.SetInfo("Todo API", "1.0.0", "A simple Todo API built with EchoNext")

    // Register routes
    app.GET("/todos", listTodos, echonext.Route{
        Summary:     "List all todos",
        Description: "Returns a list of all todo items",
        Tags:        []string{"Todos"},
    })

    app.GET("/todos/:id", getTodo, echonext.Route{
        Summary: "Get a todo by ID",
        Tags:    []string{"Todos"},
    })

    app.POST("/todos", createTodo, echonext.Route{
        Summary:       "Create a new todo",
        Tags:          []string{"Todos"},
        SuccessStatus: 201,
    })

    app.PUT("/todos/:id", updateTodo, echonext.Route{
        Summary: "Update a todo",
        Tags:    []string{"Todos"},
    })

    app.DELETE("/todos/:id", deleteTodo, echonext.Route{
        Summary:       "Delete a todo",
        Tags:          []string{"Todos"},
        SuccessStatus: 204,
    })

    // Serve OpenAPI spec and Swagger UI
    app.ServeOpenAPISpec("/api/openapi.json")
    app.ServeSwaggerUI("/api/docs", "/api/openapi.json")

    // Start server
    app.Start(":8080")
}

// Handler functions
func listTodos(c *echo.Context) ([]Todo, error) {
    return todos, nil
}

func getTodo(c *echo.Context) (Todo, error) {
    id := c.Param("id")
    for _, todo := range todos {
        if todo.ID == parseID(id) {
            return todo, nil
        }
    }
    return Todo{}, echo.NewHTTPError(404, "todo not found")
}

func createTodo(c *echo.Context, req CreateTodoRequest) (Todo, error) {
    todo := Todo{
        ID:        nextID,
        Title:     req.Title,
        Completed: false,
    }
    nextID++
    todos = append(todos, todo)
    return todo, nil
}

func updateTodo(c *echo.Context, req UpdateTodoRequest) (Todo, error) {
    id := c.Param("id")
    for i, todo := range todos {
        if todo.ID == parseID(id) {
            if req.Title != nil {
                todos[i].Title = *req.Title
            }
            if req.Completed != nil {
                todos[i].Completed = *req.Completed
            }
            return todos[i], nil
        }
    }
    return Todo{}, echo.NewHTTPError(404, "todo not found")
}

func deleteTodo(c *echo.Context) error {
    id := c.Param("id")
    for i, todo := range todos {
        if todo.ID == parseID(id) {
            todos = append(todos[:i], todos[i+1:]...)
            return nil
        }
    }
    return echo.NewHTTPError(404, "todo not found")
}

// Helper function
func parseID(s string) int {
    var id int
    fmt.Sscanf(s, "%d", &id)
    return id
}
```

### Step 3: Run Your API

```bash
# Install dependencies
go mod tidy

# Run the server
go run main.go
```

### Step 4: Test Your API

Your API is now running! Visit:

- **Swagger UI**: http://localhost:8080/api/docs
- **OpenAPI Spec**: http://localhost:8080/api/openapi.json

Try these commands:

```bash
# List all todos
curl http://localhost:8080/todos

# Create a todo
curl -X POST http://localhost:8080/todos \
  -H "Content-Type: application/json" \
  -d '{"title": "My new todo"}'

# Get a specific todo
curl http://localhost:8080/todos/1

# Update a todo
curl -X PUT http://localhost:8080/todos/1 \
  -H "Content-Type: application/json" \
  -d '{"completed": true}'

# Delete a todo
curl -X DELETE http://localhost:8080/todos/1
```

## What You Just Built

In less than 100 lines of code, you created an API with:

✅ **Type-safe handlers** - No manual JSON parsing  
✅ **Automatic validation** - Using struct tags  
✅ **OpenAPI documentation** - Generated automatically  
✅ **Swagger UI** - Interactive API testing  
✅ **Clean code** - Focus on business logic  

## Key Features Demonstrated

### 1. Type-Safe Handlers

Instead of manually parsing JSON:

```go
// ❌ Traditional way
func handler(c *echo.Context) error {
    var req CreateTodoRequest
    if err := c.Bind(&req); err != nil {
        return err
    }
    if err := c.Validate(&req); err != nil {
        return err
    }
    // ... handle request
}

// ✅ EchoNext way
func handler(c *echo.Context, req CreateTodoRequest) (Todo, error) {
    // req is already parsed and validated!
    return todo, nil
}
```

### 2. Automatic Validation

Add validation rules with struct tags:

```go
type CreateTodoRequest struct {
    Title string `json:"title" validate:"required,min=3,max=100"`
}
```

EchoNext validates automatically and returns clear error messages.

### 3. OpenAPI Generation

Route metadata becomes OpenAPI documentation:

```go
app.POST("/todos", createTodo, echonext.Route{
    Summary:       "Create a new todo",
    Description:   "Creates a new todo item",
    Tags:          []string{"Todos"},
    SuccessStatus: 201,
})
```

## Using the CLI (Faster Way)

Want to create projects even faster? Use the EchoNext CLI:

```bash
# Install CLI
go install github.com/abdussamadbello/echonext/cmd/echonext-cli@latest

# Create a new project
echonext init todo-api --module=github.com/yourusername/todo-api

# Navigate to project
cd todo-api

# Generate a domain (model, service, handler, DTOs)
echonext generate domain todo

# Initialize database
echonext db init

# Run the API
go mod tidy
go run ./cmd/api
```

This generates a complete project structure with:
- Organized directory layout
- Database integration
- Configuration management
- Middleware setup
- Testing structure

## Next Steps

Now that you've built your first API:

1. **Learn Core Concepts** - [Core Concepts](./concepts.md)
2. **Add More Features** - [API Development Guide](../guides/api-development.md)
3. **Add Database** - [Database Guide](../contrib/database.md)
4. **Add Tests** - [Testing Guide](../guides/testing.md)
5. **Explore Examples** - [Example Projects](../examples/README.md)

## Common Patterns

### Adding Middleware

```go
import "github.com/labstack/echo/v5/middleware"

app := echonext.New()

// Use Echo middleware
app.Use(middleware.Logger())
app.Use(middleware.Recover())
app.Use(middleware.CORS())
```

### Query Parameters

```go
type ListTodosRequest struct {
    Page   int    `query:"page" validate:"min=1"`
    Limit  int    `query:"limit" validate:"min=1,max=100"`
    Status string `query:"status" validate:"omitempty,oneof=all active completed"`
}

func listTodos(c *echo.Context, req ListTodosRequest) ([]Todo, error) {
    // Use req.Page, req.Limit, req.Status
    return filteredTodos, nil
}
```

### Error Handling

```go
func getTodo(c *echo.Context) (Todo, error) {
    id := c.Param("id")
    todo, err := db.FindTodo(id)
    if err != nil {
        return Todo{}, echo.NewHTTPError(404, "todo not found")
    }
    return todo, nil
}
```

### Custom Status Codes

```go
app.POST("/todos", createTodo, echonext.Route{
    Summary:       "Create todo",
    SuccessStatus: 201, // Returns 201 Created
})

app.DELETE("/todos/:id", deleteTodo, echonext.Route{
    Summary:       "Delete todo",
    SuccessStatus: 204, // Returns 204 No Content
})
```

## Troubleshooting

### Validation Errors

If you get validation errors, check:
1. Struct tags are correct
2. Required fields are provided
3. Values meet constraints (min, max, etc.)

### Import Errors

Make sure to import:
```go
import (
    "github.com/abdussamadbello/echonext"
    "github.com/labstack/echo/v5"
)
```

### Port Already in Use

Change the port in `app.Start()`:
```go
app.Start(":3000")  // Use port 3000 instead
```

## Getting Help

- [FAQ](../faq.md)
- [Troubleshooting Guide](../troubleshooting.md)
- [GitHub Issues](https://github.com/abdussamadbello/echonext/issues)

Ready to build more complex APIs? Check out the [API Development Guide](../guides/api-development.md)!
