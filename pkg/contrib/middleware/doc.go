// Package middleware provides optional Echo middleware helpers for EchoNext applications.
//
// This package COMPLEMENTS Echo's built-in middleware rather than replacing it.
// Echo already has excellent middleware for logging, recovery, CORS, etc.
// These helpers add additional functionality that works alongside Echo's middleware.
//
// Features:
//   - RequestID: Add correlation IDs to requests
//   - Metrics: Simple request metrics collection
//   - StructuredLogger: Enhanced logging with structured fields
//   - OpenTelemetry: Automatic distributed tracing and metrics
//
// # OpenTelemetry Integration
//
// The OTEL middleware provides automatic instrumentation for distributed tracing
// and metrics collection. It supports auto-configuration from environment variables.
//
// Environment Variables:
//   - OTEL_SERVICE_NAME: Service name (default: "echonext-service")
//   - OTEL_SERVICE_VERSION: Service version (default: "1.0.0")
//   - OTEL_ENVIRONMENT: Deployment environment (default: "development")
//   - OTEL_EXPORTER_OTLP_ENDPOINT: OTLP collector endpoint (default: "localhost:4317")
//   - OTEL_INSECURE: Use insecure connection (default: "true")
//   - OTEL_SAMPLE_RATE: Trace sampling rate 0.0-1.0 (default: "1.0")
//   - OTEL_ENABLE_TRACING: Enable tracing (default: "true")
//   - OTEL_ENABLE_METRICS: Enable metrics (default: "true")
//
// Example OTEL usage:
//
//	import (
//	    "context"
//	    "github.com/abdussamadbello/echonext"
//	    "github.com/abdussamadbello/echonext/pkg/contrib/middleware"
//	)
//
//	func main() {
//	    app := echonext.New()
//
//	    // Initialize OTEL with auto-configuration from environment
//	    shutdown, err := middleware.InitOTEL(context.Background(), middleware.DefaultOTELConfig())
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    defer shutdown.Shutdown(context.Background())
//
//	    // Add OTEL middleware for automatic request tracing
//	    app.Use(middleware.OTELMiddleware("my-service"))
//
//	    // Or with custom options
//	    app.Use(middleware.OTELMiddleware("my-service",
//	        middleware.WithSkipper(func(c *echo.Context) bool {
//	            return c.Path() == "/health"
//	        }),
//	        middleware.WithCustomAttributes(
//	            attribute.String("custom.key", "value"),
//	        ),
//	    ))
//
//	    // Access trace info in handlers
//	    app.GET("/users", func(c *echo.Context) error {
//	        traceID := middleware.GetTraceID(c)
//	        middleware.AddSpanEvent(c, "fetching users")
//	        // ...
//	    })
//	}
//
// # Traced HTTP Client (Outgoing Requests)
//
// The package provides an instrumented HTTP client for tracing outgoing requests.
// This enables full distributed tracing across service boundaries.
//
// Creating a traced client:
//
//	// Option 1: Create new traced client
//	client := middleware.NewTracedHTTPClient()
//
//	// Option 2: With custom options
//	client := middleware.NewTracedHTTPClient(
//	    middleware.WithClientTimeout(10 * time.Second),
//	    middleware.WithBaseTransport(customTransport),
//	)
//
//	// Option 3: Wrap an existing client
//	client := middleware.WrapHTTPClient(existingClient)
//
// Making traced outgoing requests in handlers:
//
//	app.GET("/users/:id", func(c *echo.Context) error {
//	    // Get context with trace info from incoming request
//	    ctx := c.Request().Context()
//
//	    // Make traced GET request - trace context is automatically propagated
//	    resp, err := client.Get(ctx, "http://user-service/api/users")
//	    if err != nil {
//	        return err
//	    }
//	    defer resp.Body.Close()
//
//	    // Or create a request manually with trace context
//	    req, _ := middleware.NewRequestWithTrace(c, "POST", "http://other-service/api", body)
//	    resp, err = client.Do(req)
//	})
//
// The traced client automatically:
//   - Creates spans for each outgoing HTTP request
//   - Propagates trace context (traceparent, tracestate headers)
//   - Records HTTP method, URL, status code, and duration
//   - Links outgoing spans to the parent incoming request span
//
// Helper functions:
//   - NewTracedHTTPClient: Create a new instrumented HTTP client
//   - WrapHTTPClient: Wrap an existing http.Client with tracing
//   - WrapTransport: Wrap an http.RoundTripper with tracing
//   - NewRequestWithTrace: Create http.Request with trace context from echo.Context
//   - HTTPClientFromContext: Get context with trace info from echo.Context
//
// # Basic Middleware Usage
//
// Example usage:
//
//	import (
//	    "github.com/abdussamadbello/echonext"
//	    "github.com/abdussamadbello/echonext/pkg/contrib/middleware"
//	    echomw "github.com/labstack/echo/v5/middleware"
//	)
//
//	app := echonext.New()
//
//	// Use Echo's built-in middleware
//	app.Use(echomw.Recover())
//	app.Use(echomw.CORS("*"))
//
//	// Add contrib middleware
//	app.Use(middleware.RequestID())
//
//	// Create metrics collector
//	metrics := middleware.NewMetrics()
//	app.Use(middleware.MetricsMiddleware(metrics))
//
//	// Expose metrics endpoint
//	app.Echo.GET("/metrics", middleware.MetricsHandler(metrics))
//
//	// Use structured logging with request IDs
//	app.Use(middleware.StructuredLogger(middleware.StructuredLoggerConfig{
//	    CustomFields: func(c *echo.Context) map[string]interface{} {
//	        return map[string]interface{}{
//	            "user_agent": c.Request().UserAgent(),
//	        }
//	    },
//	}))
package middleware
