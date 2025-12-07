// OTEL Demo - OpenTelemetry Distributed Tracing Example
//
// This example demonstrates how to use EchoNext's OpenTelemetry middleware
// for automatic distributed tracing of incoming and outgoing HTTP requests.
//
// Run with: go run main.go
// API docs: http://localhost:8080/api/docs
// Trace viewer: http://localhost:16686 (Jaeger)
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
	"github.com/abdussamadbello/echonext/pkg/contrib/middleware"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/otel/attribute"
)

// ============================================================================
// Domain Models
// ============================================================================

type Order struct {
	ID        string    `json:"id"`
	Product   string    `json:"product"`
	Quantity  int       `json:"quantity"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateOrderRequest struct {
	Product  string `json:"product" validate:"required,min=1,max=200"`
	Quantity int    `json:"quantity" validate:"required,min=1,max=1000"`
}

// In-memory storage
var orders = make(map[string]*Order)

// Traced HTTP client for outgoing requests
var tracedClient *middleware.TracedHTTPClient

// ============================================================================
// Main Application
// ============================================================================

func main() {
	app := echonext.New()
	ctx := context.Background()

	// ========================================================================
	// Step 1: Initialize OpenTelemetry
	// ========================================================================
	// This sets up trace and metric exporters. If the OTEL collector is not
	// running, the app will still work (just without tracing).
	//
	// Default OTLP gRPC endpoint: localhost:4317
	// Run Jaeger: docker run -d -p 16686:16686 -p 4317:4317 jaegertracing/all-in-one
	//
	shutdown, err := middleware.InitOTEL(ctx, middleware.OTELConfig{
		ServiceName:    "order-service",
		ServiceVersion: "1.0.0",
		Environment:    "development",
		Endpoint:       "localhost:4317",
		Insecure:       true,
		SampleRate:     1.0, // Sample all requests in development
		EnableTracing:  true,
		EnableMetrics:  true,
	})
	if err != nil {
		log.Printf("OTEL initialization failed (running without tracing): %v", err)
	} else {
		defer shutdown.Shutdown(ctx)
		log.Println("OpenTelemetry initialized - traces will be sent to localhost:4317")
	}

	// ========================================================================
	// Step 2: Create Traced HTTP Client for Outgoing Requests
	// ========================================================================
	// This client automatically propagates trace context (traceparent header)
	// to downstream services, enabling distributed tracing.
	//
	tracedClient = middleware.NewTracedHTTPClient(
		middleware.WithClientTimeout(30 * time.Second),
	)

	// ========================================================================
	// Step 3: Configure API Documentation
	// ========================================================================
	app.SetInfo(
		"Order Service API",
		"1.0.0",
		"Demonstrates OpenTelemetry distributed tracing with EchoNext",
	)
	app.SetServers([]echonext.Server{
		{URL: "http://localhost:8080", Description: "Development server"},
	})

	// ========================================================================
	// Step 4: Add Middleware
	// ========================================================================
	app.Use(echomw.Logger())
	app.Use(echomw.Recover())

	// RequestID adds X-Request-ID header for correlation
	app.Use(middleware.RequestID())

	// OTELMiddleware instruments all incoming HTTP requests
	// Each request gets a trace span with:
	// - HTTP method, path, status code
	// - Request/response timing
	// - Error recording (if any)
	app.Use(middleware.OTELMiddleware("order-service"))

	// ========================================================================
	// Step 5: Register Routes
	// ========================================================================

	// Health check endpoint
	app.GET("/health", healthCheck, echonext.Route{
		Summary: "Health check",
		Tags:    []string{"System"},
	})

	// -------------------------------------------------------------------------
	// Order CRUD endpoints
	// -------------------------------------------------------------------------

	app.POST("/orders", createOrder, echonext.Route{
		Summary:       "Create a new order",
		Description:   "Creates a new order and demonstrates span events",
		Tags:          []string{"Orders"},
		SuccessStatus: 201,
	})

	app.GET("/orders", listOrders, echonext.Route{
		Summary: "List all orders",
		Tags:    []string{"Orders"},
	})

	app.GET("/orders/:id", getOrder, echonext.Route{
		Summary: "Get order by ID",
		Tags:    []string{"Orders"},
	})

	// -------------------------------------------------------------------------
	// OTEL Demo endpoints - demonstrates tracing features
	// -------------------------------------------------------------------------

	app.GET("/demo/external-call", demoExternalCall, echonext.Route{
		Summary:     "Demo: Traced External API Call",
		Description: "Makes a traced HTTP request to an external API. The trace context is automatically propagated.",
		Tags:        []string{"OTEL Demo"},
	})

	app.GET("/demo/trace-info", demoTraceInfo, echonext.Route{
		Summary:     "Demo: Current Trace Information",
		Description: "Returns the current trace ID, span ID, and request ID for debugging.",
		Tags:        []string{"OTEL Demo"},
	})

	app.GET("/demo/nested-spans", demoNestedSpans, echonext.Route{
		Summary:     "Demo: Nested Span Operations",
		Description: "Demonstrates creating child spans for sub-operations within a request.",
		Tags:        []string{"OTEL Demo"},
	})

	app.GET("/demo/error-recording", demoErrorRecording, echonext.Route{
		Summary:     "Demo: Error Recording in Traces",
		Description: "Demonstrates how errors are recorded in spans for debugging.",
		Tags:        []string{"OTEL Demo"},
	})

	// Serve API documentation
	app.ServeOpenAPISpec("/api/openapi.json")
	app.ServeSwaggerUI("/api/docs", "/api/openapi.json")

	// Add sample data
	seedOrders()

	// Start server
	log.Println("=================================================")
	log.Println("Order Service (OTEL Demo) starting...")
	log.Println("=================================================")
	log.Println("API Server:     http://localhost:8080")
	log.Println("API Docs:       http://localhost:8080/api/docs")
	log.Println("Trace Viewer:   http://localhost:16686 (Jaeger)")
	log.Println("")
	log.Println("Quick start Jaeger:")
	log.Println("  docker run -d -p 16686:16686 -p 4317:4317 jaegertracing/all-in-one")
	log.Println("=================================================")

	log.Fatal(app.Start(":8080"))
}

// ============================================================================
// Handler Implementations
// ============================================================================

func healthCheck(c echo.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"status":    "healthy",
		"service":   "order-service",
		"trace_id":  middleware.GetTraceID(c),
		"timestamp": time.Now(),
	}, nil
}

func createOrder(c echo.Context, req CreateOrderRequest) (Order, error) {
	// Add a span event to mark the start of order creation
	middleware.AddSpanEvent(c, "creating order",
		attribute.String("product", req.Product),
		attribute.Int("quantity", req.Quantity),
	)

	order := Order{
		ID:        fmt.Sprintf("order_%d", time.Now().UnixNano()),
		Product:   req.Product,
		Quantity:  req.Quantity,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	orders[order.ID] = &order

	// Add another span event after order is stored
	middleware.AddSpanEvent(c, "order stored",
		attribute.String("order_id", order.ID),
	)

	return order, nil
}

func listOrders(c echo.Context) ([]Order, error) {
	middleware.AddSpanEvent(c, "listing orders",
		attribute.Int("count", len(orders)),
	)

	result := make([]Order, 0, len(orders))
	for _, order := range orders {
		result = append(result, *order)
	}
	return result, nil
}

func getOrder(c echo.Context) (Order, error) {
	id := c.Param("id")

	middleware.AddSpanEvent(c, "fetching order",
		attribute.String("order_id", id),
	)

	order, exists := orders[id]
	if !exists {
		middleware.RecordError(c, fmt.Errorf("order not found: %s", id))
		return Order{}, echo.NewHTTPError(404, "order not found")
	}

	return *order, nil
}

// ============================================================================
// OTEL Demo Handlers
// ============================================================================

// demoExternalCall demonstrates traced outgoing HTTP requests
func demoExternalCall(c echo.Context) (map[string]interface{}, error) {
	middleware.AddSpanEvent(c, "starting external API call")

	// Get the request context - it contains the trace context
	ctx := c.Request().Context()

	// Make a traced HTTP request
	// The traceparent header is automatically injected for distributed tracing
	resp, err := tracedClient.Get(ctx, "https://httpbin.org/json")
	if err != nil {
		middleware.RecordError(c, err)
		return nil, echo.NewHTTPError(http.StatusBadGateway, "failed to fetch external data")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		middleware.RecordError(c, err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to read response")
	}

	var externalData map[string]interface{}
	if err := json.Unmarshal(body, &externalData); err != nil {
		middleware.RecordError(c, err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to parse response")
	}

	middleware.AddSpanEvent(c, "external API call completed")

	return map[string]interface{}{
		"trace_id":      middleware.GetTraceID(c),
		"span_id":       middleware.GetSpanID(c),
		"external_data": externalData,
		"message":       "Check Jaeger to see the distributed trace including the outgoing HTTP call!",
	}, nil
}

// demoTraceInfo returns current trace context information
func demoTraceInfo(c echo.Context) (map[string]interface{}, error) {
	middleware.AddSpanEvent(c, "trace info requested")

	return map[string]interface{}{
		"trace_id":   middleware.GetTraceID(c),
		"span_id":    middleware.GetSpanID(c),
		"request_id": middleware.GetRequestID(c),
		"descriptions": map[string]string{
			"trace_id":   "Unique identifier for the entire distributed trace (same across all services)",
			"span_id":    "Unique identifier for this specific operation within the trace",
			"request_id": "Correlation ID added by RequestID middleware (X-Request-ID header)",
		},
		"how_to_view": "Run Jaeger: docker run -d -p 16686:16686 -p 4317:4317 jaegertracing/all-in-one",
	}, nil
}

// demoNestedSpans demonstrates adding span events for sub-operations
func demoNestedSpans(c echo.Context) (map[string]interface{}, error) {
	// Simulate a multi-step operation with span events
	middleware.AddSpanEvent(c, "step 1: validating input")
	time.Sleep(10 * time.Millisecond)

	middleware.AddSpanEvent(c, "step 2: fetching from database",
		attribute.String("table", "orders"),
	)
	time.Sleep(20 * time.Millisecond)

	middleware.AddSpanEvent(c, "step 3: transforming data")
	time.Sleep(5 * time.Millisecond)

	middleware.AddSpanEvent(c, "step 4: preparing response")

	return map[string]interface{}{
		"trace_id": middleware.GetTraceID(c),
		"message":  "Check Jaeger to see the span events as markers in the trace timeline!",
		"steps": []string{
			"step 1: validating input",
			"step 2: fetching from database",
			"step 3: transforming data",
			"step 4: preparing response",
		},
	}, nil
}

// demoErrorRecording demonstrates how errors appear in traces
func demoErrorRecording(c echo.Context) (map[string]interface{}, error) {
	// Simulate an error scenario
	simulatedErr := fmt.Errorf("simulated error for demonstration")

	// Record the error in the span (it will appear in Jaeger)
	middleware.RecordError(c, simulatedErr)

	middleware.AddSpanEvent(c, "error was recorded",
		attribute.String("error_type", "simulated"),
	)

	// Return success but note the error was recorded
	return map[string]interface{}{
		"trace_id": middleware.GetTraceID(c),
		"message":  "Error was recorded in the span - check Jaeger to see it marked as an error!",
		"note":     "In real scenarios, you'd return an error response after recording",
	}, nil
}

// ============================================================================
// Seed Data
// ============================================================================

func seedOrders() {
	orders["order_1"] = &Order{
		ID:        "order_1",
		Product:   "EchoNext Framework",
		Quantity:  1,
		Status:    "completed",
		CreatedAt: time.Now().Add(-24 * time.Hour),
	}

	orders["order_2"] = &Order{
		ID:        "order_2",
		Product:   "OTEL Collector",
		Quantity:  3,
		Status:    "pending",
		CreatedAt: time.Now().Add(-1 * time.Hour),
	}
}
