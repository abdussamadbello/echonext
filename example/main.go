package main

import (
	"fmt"
	"log"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/abdussamadbello/echonext"
)

// Domain models
type Todo struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Request/Response DTOs
type CreateTodoRequest struct {
	Title       string `json:"title" validate:"required,min=3,max=200"`
	Description string `json:"description" validate:"max=1000"`
}

type UpdateTodoRequest struct {
	Title       string `json:"title,omitempty" validate:"omitempty,min=3,max=200"`
	Description string `json:"description,omitempty" validate:"omitempty,max=1000"`
	Completed   *bool  `json:"completed,omitempty"`
}

type ListTodosRequest struct {
	Page      int    `query:"page" validate:"min=1"`
	Limit     int    `query:"limit" validate:"min=1,max=100"`
	Completed *bool  `query:"completed"`
	Sort      string `query:"sort" validate:"omitempty,oneof=created_at updated_at title"`
}

type ListTodosResponse struct {
	Todos      []Todo `json:"todos"`
	TotalCount int    `json:"total_count"`
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
}

// In-memory storage
var todos = make(map[string]*Todo)

func main() {
	// Create EchoNext app
	app := echonext.New()

	// Configure API info
	app.SetInfo(
		"Todo API",
		"1.0.0",
		"A simple todo management API built with EchoNext",
	)

	// Add middleware
	app.Use(middleware.Logger())
	app.Use(middleware.Recover())
	app.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.PUT, echo.POST, echo.DELETE},
	}))

	// Health check
	app.GET("/health", healthCheck, echonext.Route{
		Summary: "Health check",
		Tags:    []string{"System"},
	})

	// Todo endpoints
	app.POST("/todos", createTodo, echonext.Route{
		Summary:     "Create a new todo",
		Description: "Creates a new todo item with the provided title and description",
		Tags:        []string{"Todos"},
	})

	app.GET("/todos", listTodos, echonext.Route{
		Summary:     "List todos",
		Description: "Returns a paginated list of todos with optional filtering",
		Tags:        []string{"Todos"},
	})

	app.GET("/todos/:id", getTodo, echonext.Route{
		Summary:     "Get todo by ID",
		Description: "Returns a single todo item by its ID",
		Tags:        []string{"Todos"},
	})

	app.PUT("/todos/:id", updateTodo, echonext.Route{
		Summary:     "Update todo",
		Description: "Updates an existing todo item",
		Tags:        []string{"Todos"},
	})

	app.DELETE("/todos/:id", deleteTodo, echonext.Route{
		Summary:     "Delete todo",
		Description: "Deletes a todo item by its ID",
		Tags:        []string{"Todos"},
	})

	// Serve API documentation
	app.ServeOpenAPISpec("/api/openapi.json")
	app.ServeSwaggerUI("/api/docs", "/api/openapi.json")

	// Add some sample data
	seedData()

	// Start server
	log.Println("Server starting on http://localhost:8080")
	log.Println("API Documentation: http://localhost:8080/api/docs")
	log.Fatal(app.Start(":8080"))
}

// Handler implementations
func healthCheck(c echo.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"status":  "healthy",
		"service": "todo-api",
		"time":    time.Now(),
	}, nil
}

func createTodo(c echo.Context, req CreateTodoRequest) (Todo, error) {
	todo := Todo{
		ID:          generateID(),
		Title:       req.Title,
		Description: req.Description,
		Completed:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	todos[todo.ID] = &todo
	return todo, nil
}

func listTodos(c echo.Context, req ListTodosRequest) (ListTodosResponse, error) {
	// Set defaults
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Limit == 0 {
		req.Limit = 10
	}

	// Filter todos
	var filteredTodos []Todo
	for _, todo := range todos {
		if req.Completed != nil && todo.Completed != *req.Completed {
			continue
		}
		filteredTodos = append(filteredTodos, *todo)
	}

	// Simple pagination
	start := (req.Page - 1) * req.Limit
	end := start + req.Limit
	if end > len(filteredTodos) {
		end = len(filteredTodos)
	}
	if start > len(filteredTodos) {
		start = len(filteredTodos)
	}

	return ListTodosResponse{
		Todos:      filteredTodos[start:end],
		TotalCount: len(filteredTodos),
		Page:       req.Page,
		Limit:      req.Limit,
	}, nil
}

func getTodo(c echo.Context) (Todo, error) {
	id := c.Param("id")
	todo, exists := todos[id]
	if !exists {
		return Todo{}, echo.NewHTTPError(404, "todo not found")
	}
	return *todo, nil
}

func updateTodo(c echo.Context, req UpdateTodoRequest) (Todo, error) {
	id := c.Param("id")
	todo, exists := todos[id]
	if !exists {
		return Todo{}, echo.NewHTTPError(404, "todo not found")
	}

	// Update fields if provided
	if req.Title != "" {
		todo.Title = req.Title
	}
	if req.Description != "" {
		todo.Description = req.Description
	}
	if req.Completed != nil {
		todo.Completed = *req.Completed
	}
	todo.UpdatedAt = time.Now()

	return *todo, nil
}

func deleteTodo(c echo.Context) error {
	id := c.Param("id")
	if _, exists := todos[id]; !exists {
		return echo.NewHTTPError(404, "todo not found")
	}

	delete(todos, id)
	return nil
}

// Helper functions
func generateID() string {
	return fmt.Sprintf("todo_%d", time.Now().UnixNano())
}

func seedData() {
	todos["todo_1"] = &Todo{
		ID:          "todo_1",
		Title:       "Build EchoNext framework",
		Description: "Create a type-safe wrapper around Echo with OpenAPI generation",
		Completed:   true,
		CreatedAt:   time.Now().Add(-48 * time.Hour),
		UpdatedAt:   time.Now().Add(-24 * time.Hour),
	}

	todos["todo_2"] = &Todo{
		ID:          "todo_2",
		Title:       "Write documentation",
		Description: "Create comprehensive docs and examples",
		Completed:   false,
		CreatedAt:   time.Now().Add(-24 * time.Hour),
		UpdatedAt:   time.Now().Add(-12 * time.Hour),
	}

	todos["todo_3"] = &Todo{
		ID:          "todo_3",
		Title:       "Add tests",
		Description: "Write unit and integration tests",
		Completed:   false,
		CreatedAt:   time.Now().Add(-12 * time.Hour),
		UpdatedAt:   time.Now().Add(-6 * time.Hour),
	}
}