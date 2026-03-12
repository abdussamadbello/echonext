package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/abdussamadbello/echonext"
	contribmw "github.com/abdussamadbello/echonext/pkg/contrib/middleware"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
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
	Title       string `json:"title" validate:"required,min=3,max=200" example:"Learn Go"`
	Description string `json:"description" validate:"max=1000" example:"Study Go programming language"`
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

// Traced HTTP client for outgoing requests (demonstrates OTEL propagation)
var tracedClient *contribmw.TracedHTTPClient

func main() {
	// Create EchoNext app
	app := echonext.New()

	// ==========================================================================
	// OpenTelemetry Setup (Optional - runs without collector if unavailable)
	// ==========================================================================
	// Initialize OTEL - traces and metrics will be sent to the OTLP collector
	// If no collector is running, the app will still work (just no tracing)
	ctx := context.Background()
	shutdown, err := contribmw.InitOTEL(ctx, contribmw.OTELConfig{
		ServiceName:    "todo-api",
		ServiceVersion: "1.0.0",
		Environment:    "development",
		Endpoint:       "localhost:4317", // Default OTLP gRPC endpoint
		Insecure:       true,
		SampleRate:     1.0, // Sample all requests in development
		EnableTracing:  true,
		EnableMetrics:  true,
	})
	if err != nil {
		log.Printf("⚠️  OTEL initialization failed (running without tracing): %v", err)
	} else {
		defer shutdown.Shutdown(ctx)
		log.Println("✅ OpenTelemetry initialized - traces will be sent to localhost:4317")
	}

	// Initialize traced HTTP client for outgoing requests
	tracedClient = contribmw.NewTracedHTTPClient()

	// Configure API info
	app.SetInfo(
		"Todo API",
		"1.0.0",
		"A simple todo management API built with EchoNext",
	)
	app.SetContact("Todo Team", "https://example.com/support", "support@example.com")
	app.SetLicense("MIT", "https://opensource.org/licenses/MIT")
	app.SetServers([]echonext.Server{
		{URL: "http://localhost:8080", Description: "Development server"},
		{URL: "https://api.todo.com", Description: "Production server"},
	})

	// Add security schemes
	_ = app.AddSecurityScheme("bearerAuth", echonext.Security{
		Type:        "bearer",
		Scheme:      "JWT",
		Description: "JWT Bearer token authentication",
	})
	_ = app.AddSecurityScheme("apiKeyAuth", echonext.Security{
		Type:        "apiKey",
		Name:        "X-API-Key",
		In:          "header",
		Description: "API key for service identification",
	})

	// Add middleware
	app.Use(middleware.Logger())
	app.Use(middleware.Recover())
	app.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.PUT, echo.POST, echo.DELETE},
	}))

	// Add OTEL middleware for automatic request tracing
	// This creates spans for all incoming HTTP requests
	app.Use(contribmw.RequestID()) // Adds X-Request-ID header
	app.Use(contribmw.OTELMiddleware("todo-api"))

	// Health check
	app.GET("/health", healthCheck, echonext.Route{
		Summary: "Health check",
		Tags:    []string{"System"},
	})

	// Todo endpoints
	app.POST("/todos", createTodo, echonext.Route{
		Summary:       "Create a new todo",
		Description:   "Creates a new todo item with the provided title and description",
		Tags:          []string{"Todos"},
		SuccessStatus: 201,
		Security: []echonext.SecurityRequirement{
			{SchemeName: "bearerAuth"},
		},
		RequestHeaders: map[string]echonext.HeaderInfo{
			"X-Request-ID": {
				Description: "Unique request identifier",
				Required:    false,
				Schema:      "string",
			},
		},
		ResponseHeaders: map[string]echonext.HeaderInfo{
			"X-Todo-ID": {
				Description: "ID of the created todo",
				Schema:      "string",
			},
		},
		Examples: map[string]interface{}{
			"basic": map[string]interface{}{
				"title":       "Learn EchoNext",
				"description": "Study the EchoNext framework",
			},
		},
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
		Security: []echonext.SecurityRequirement{
			{SchemeName: "bearerAuth"},
			{SchemeName: "apiKeyAuth"},
		},
	})

	app.DELETE("/todos/:id", deleteTodo, echonext.Route{
		Summary:       "Delete todo",
		Description:   "Deletes a todo item by its ID",
		Tags:          []string{"Todos"},
		SuccessStatus: 204,
		Security: []echonext.SecurityRequirement{
			{SchemeName: "bearerAuth"},
		},
	})

	// ==========================================================================
	// OTEL Demo Endpoints - Demonstrates distributed tracing features
	// ==========================================================================

	// External API call with traced HTTP client
	app.GET("/external", fetchExternalData, echonext.Route{
		Summary:     "Fetch external data (OTEL demo)",
		Description: "Demonstrates traced outgoing HTTP requests. The trace context is automatically propagated to external services.",
		Tags:        []string{"OTEL Demo"},
	})

	// Trace info endpoint - shows current trace context
	app.GET("/trace-info", getTraceInfo, echonext.Route{
		Summary:     "Get trace information (OTEL demo)",
		Description: "Returns the current trace ID and span ID for debugging distributed traces.",
		Tags:        []string{"OTEL Demo"},
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

// =============================================================================
// OTEL Demo Handlers - Demonstrates distributed tracing features
// =============================================================================

// fetchExternalData demonstrates making traced outgoing HTTP requests
// The trace context is automatically propagated to external services
func fetchExternalData(c echo.Context) (map[string]interface{}, error) {
	// Add a custom span event (visible in trace viewers like Jaeger)
	contribmw.AddSpanEvent(c, "starting external API call")

	// Get context with trace info - this will be propagated to external service
	ctx := c.Request().Context()

	// Make traced HTTP request using our instrumented client
	// The traceparent header is automatically injected
	resp, err := tracedClient.Get(ctx, "https://httpbin.org/json")
	if err != nil {
		contribmw.RecordError(c, err)
		return nil, echo.NewHTTPError(http.StatusBadGateway, "failed to fetch external data")
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		contribmw.RecordError(c, err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to read response")
	}

	// Parse JSON response
	var externalData map[string]interface{}
	if err := json.Unmarshal(body, &externalData); err != nil {
		contribmw.RecordError(c, err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to parse response")
	}

	// Add another span event
	contribmw.AddSpanEvent(c, "external API call completed")

	// Return response with trace info for debugging
	return map[string]interface{}{
		"trace_id":      contribmw.GetTraceID(c),
		"span_id":       contribmw.GetSpanID(c),
		"external_data": externalData,
		"message":       "This request was traced! Check your OTEL collector (Jaeger/Zipkin) to see the distributed trace.",
	}, nil
}

// getTraceInfo returns the current trace context information
// Useful for debugging distributed traces
func getTraceInfo(c echo.Context) (map[string]interface{}, error) {
	traceID := contribmw.GetTraceID(c)
	spanID := contribmw.GetSpanID(c)
	requestID := contribmw.GetRequestID(c)

	// Add a span event
	contribmw.AddSpanEvent(c, "trace info requested")

	return map[string]interface{}{
		"trace_id":   traceID,
		"span_id":    spanID,
		"request_id": requestID,
		"info": map[string]string{
			"trace_id_desc":   "Unique identifier for the entire distributed trace",
			"span_id_desc":    "Unique identifier for this specific operation within the trace",
			"request_id_desc": "Correlation ID added by RequestID middleware",
		},
		"how_to_view": "Set up Jaeger (docker run -d -p 16686:16686 -p 4317:4317 jaegertracing/all-in-one) and visit http://localhost:16686",
	}, nil
}
