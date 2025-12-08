# Middleware

EchoNext provides optional middleware helpers that complement Echo's built-in middleware.

## Overview

The `pkg/contrib/middleware` package provides:

- **RequestID** - Add correlation IDs to requests
- **Metrics** - Simple request metrics collection
- **StructuredLogger** - Enhanced logging with structured fields
- **OpenTelemetry** - Automatic distributed tracing and metrics

These middleware **complement** Echo's built-in middleware rather than replacing them. Use Echo's excellent middleware for logging, recovery, CORS, etc., and add these helpers for additional functionality.

## RequestID Middleware

Add unique request IDs for correlation and debugging:

```go
import "github.com/abdussamadbello/echonext/pkg/contrib/middleware"

app := echonext.New()
app.Use(middleware.RequestID())

// Access in handlers
func handler(c echo.Context) error {
    requestID := middleware.GetRequestID(c)
    log.Printf("Processing request %s", requestID)
    return c.JSON(200, map[string]string{"request_id": requestID})
}
```

The Request ID is also added to the `X-Request-ID` header in responses.

## Metrics Middleware

Collect basic request metrics:

```go
import "github.com/abdussamadbello/echonext/pkg/contrib/middleware"

// Create metrics collector
metrics := middleware.NewMetrics()

// Add middleware
app.Use(middleware.MetricsMiddleware(metrics))

// Expose metrics endpoint
app.GET("/metrics", middleware.MetricsHandler(metrics))
```

Metrics collected:
- Request count
- Request duration
- Status codes
- Error rates

## Structured Logger Middleware

Enhanced logging with custom fields:

```go
app.Use(middleware.StructuredLogger(middleware.StructuredLoggerConfig{
    CustomFields: func(c echo.Context) map[string]interface{} {
        return map[string]interface{}{
            "request_id": middleware.GetRequestID(c),
            "user_agent": c.Request().UserAgent(),
            "user_id":    getUserID(c),
        }
    },
}))
```

## OpenTelemetry Middleware

For comprehensive distributed tracing and observability, see the [Observability Guide](../advanced/observability.md).

Quick example:

```go
import (
    "context"
    "github.com/abdussamadbello/echonext/pkg/contrib/middleware"
)

func main() {
    ctx := context.Background()
    
    // Initialize OTEL
    shutdown, err := middleware.InitOTEL(ctx, middleware.DefaultOTELConfig())
    if err != nil {
        log.Printf("OTEL init failed: %v", err)
    } else {
        defer shutdown.Shutdown(ctx)
    }
    
    app := echonext.New()
    
    // Add OTEL middleware
    app.Use(middleware.RequestID())
    app.Use(middleware.OTELMiddleware("my-service"))
    
    // All requests are now automatically traced
    app.Start(":8080")
}
```

### Features

- **Automatic Request Tracing** - Every HTTP request creates a trace span
- **Traced HTTP Client** - Track outgoing HTTP requests
- **Span Events** - Add timeline markers within traces
- **Error Recording** - Capture errors in traces
- **Request ID Correlation** - Link requests and traces
- **Trace Context Propagation** - W3C traceparent/tracestate headers

### Traced HTTP Client

For service-to-service communication:

```go
// Create traced client
tracedClient := middleware.NewTracedHTTPClient(
    middleware.WithClientTimeout(30 * time.Second),
)

func handler(c echo.Context) error {
    ctx := c.Request().Context()
    
    // Make traced outgoing request
    resp, err := tracedClient.Get(ctx, "https://api.example.com/data")
    if err != nil {
        middleware.RecordError(c, err)
        return echo.NewHTTPError(500, "API error")
    }
    defer resp.Body.Close()
    
    // Process response...
}
```

### Environment Variables

Configure OTEL via environment variables:

```bash
export OTEL_SERVICE_NAME="my-service"
export OTEL_SERVICE_VERSION="1.0.0"
export OTEL_ENVIRONMENT="production"
export OTEL_EXPORTER_OTLP_ENDPOINT="localhost:4317"
export OTEL_INSECURE="false"
export OTEL_SAMPLE_RATE="0.1"
export OTEL_ENABLE_TRACING="true"
export OTEL_ENABLE_METRICS="true"
```

### Helper Functions

Access trace information in handlers:

```go
func handler(c echo.Context) error {
    // Get trace context
    traceID := middleware.GetTraceID(c)
    spanID := middleware.GetSpanID(c)
    requestID := middleware.GetRequestID(c)
    
    // Add span event
    middleware.AddSpanEvent(c, "processing order",
        attribute.String("order_id", "123"),
        attribute.Int("quantity", 5),
    )
    
    // Record error
    if err != nil {
        middleware.RecordError(c, err)
        return err
    }
    
    return c.JSON(200, data)
}
```

## Combining Middleware

Use multiple middleware together:

```go
import (
    "github.com/abdussamadbello/echonext"
    "github.com/abdussamadbello/echonext/pkg/contrib/middleware"
    echomw "github.com/labstack/echo/v4/middleware"
)

app := echonext.New()

// Echo built-in middleware
app.Use(echomw.Recover())
app.Use(echomw.CORS())
app.Use(echomw.Gzip())

// EchoNext contrib middleware
app.Use(middleware.RequestID())
app.Use(middleware.OTELMiddleware("my-service"))

// Metrics
metrics := middleware.NewMetrics()
app.Use(middleware.MetricsMiddleware(metrics))
app.GET("/metrics", middleware.MetricsHandler(metrics))

// Structured logging
app.Use(middleware.StructuredLogger(middleware.StructuredLoggerConfig{
    CustomFields: func(c echo.Context) map[string]interface{} {
        return map[string]interface{}{
            "trace_id":   middleware.GetTraceID(c),
            "request_id": middleware.GetRequestID(c),
        }
    },
}))
```

## Custom Middleware

Create your own middleware following Echo's pattern:

```go
func AuthMiddleware(apiKey string) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            key := c.Request().Header.Get("X-API-Key")
            if key != apiKey {
                return echo.NewHTTPError(401, "invalid API key")
            }
            return next(c)
        }
    }
}

// Use it
app.Use(AuthMiddleware("secret-key"))
```

## See Also

- [Observability Guide](../advanced/observability.md) - Complete OpenTelemetry guide
- [OpenTelemetry Demo](../../examples/otel-demo/README.md) - Working example
- [Middleware Package Docs](../../pkg/contrib/middleware/doc.go) - Full API reference
- [Echo Middleware](https://echo.labstack.com/middleware/) - Echo's built-in middleware
