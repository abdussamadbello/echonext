# Observability with OpenTelemetry

Learn how to implement comprehensive observability in your EchoNext applications using OpenTelemetry for distributed tracing, metrics, and logging.

## Table of Contents

- [Overview](#overview)
- [OpenTelemetry Setup](#opentelemetry-setup)
- [Distributed Tracing](#distributed-tracing)
- [Metrics Collection](#metrics-collection)
- [Logging Integration](#logging-integration)
- [Service-to-Service Communication](#service-to-service-communication)
- [Production Best Practices](#production-best-practices)
- [Troubleshooting](#troubleshooting)

## Overview

EchoNext provides built-in OpenTelemetry (OTEL) support for comprehensive observability:

- **Distributed Tracing** - Track requests across services
- **Metrics Collection** - Monitor performance and health
- **Trace Context Propagation** - Correlate logs and traces
- **Request ID Correlation** - Track individual requests
- **Automatic Instrumentation** - Zero-code tracing for HTTP

### Why OpenTelemetry?

- **Vendor-Agnostic** - Works with Jaeger, Zipkin, Datadog, etc.
- **Cloud-Native** - CNCF standard for observability
- **Comprehensive** - Traces, metrics, and logs in one framework
- **Future-Proof** - Industry standard with wide adoption

## OpenTelemetry Setup

### 1. Install Dependencies

OpenTelemetry dependencies are already included in EchoNext:

```go
import (
    "github.com/abdussamadbello/echonext"
    "github.com/abdussamadbello/echonext/pkg/contrib/middleware"
)
```

### 2. Initialize OTEL

Initialize OpenTelemetry in your application startup:

```go
package main

import (
    "context"
    "log"
    "github.com/abdussamadbello/echonext"
    "github.com/abdussamadbello/echonext/pkg/contrib/middleware"
)

func main() {
    ctx := context.Background()
    
    // Initialize OTEL with default config
    shutdown, err := middleware.InitOTEL(ctx, middleware.DefaultOTELConfig())
    if err != nil {
        log.Printf("OTEL init failed (running without tracing): %v", err)
    } else {
        defer shutdown.Shutdown(ctx)
    }
    
    // Create app
    app := echonext.New()
    
    // Add OTEL middleware
    app.Use(middleware.RequestID())
    app.Use(middleware.OTELMiddleware("my-service"))
    
    // Register routes...
    app.Start(":8080")
}
```

### 3. Configuration Options

#### Default Configuration

```go
cfg := middleware.DefaultOTELConfig()
// Uses environment variables:
// - OTEL_SERVICE_NAME (default: "echonext-service")
// - OTEL_SERVICE_VERSION (default: "1.0.0")
// - OTEL_ENVIRONMENT (default: "development")
// - OTEL_EXPORTER_OTLP_ENDPOINT (default: "localhost:4317")
// - OTEL_INSECURE (default: "true")
// - OTEL_SAMPLE_RATE (default: "1.0")
// - OTEL_ENABLE_TRACING (default: "true")
// - OTEL_ENABLE_METRICS (default: "true")

shutdown, err := middleware.InitOTEL(ctx, cfg)
```

#### Custom Configuration

```go
cfg := middleware.OTELConfig{
    ServiceName:    "order-service",
    ServiceVersion: "1.2.3",
    Environment:    "production",
    Endpoint:       "otel-collector:4317",
    Insecure:       false, // Use TLS
    SampleRate:     0.1,   // Sample 10% in production
    EnableTracing:  true,
    EnableMetrics:  true,
}

shutdown, err := middleware.InitOTEL(ctx, cfg)
if err != nil {
    log.Fatal(err)
}
defer shutdown.Shutdown(ctx)
```

#### Environment Variables

Set these in your deployment:

```bash
export OTEL_SERVICE_NAME="my-api"
export OTEL_SERVICE_VERSION="1.0.0"
export OTEL_ENVIRONMENT="production"
export OTEL_EXPORTER_OTLP_ENDPOINT="collector.example.com:4317"
export OTEL_INSECURE="false"
export OTEL_SAMPLE_RATE="0.1"
```

## Distributed Tracing

### Automatic Request Tracing

All incoming HTTP requests are automatically traced:

```go
app.Use(middleware.OTELMiddleware("my-service"))

// Every request automatically gets:
// - Trace ID
// - Span ID
// - Request ID
// - Timing information
// - HTTP method, path, status code
```

### Adding Span Events

Add timeline markers within a trace:

```go
func createOrder(c echo.Context, req CreateOrderRequest) (OrderResponse, error) {
    // Add event to current span
    middleware.AddSpanEvent(c, "validating order")
    
    if err := validateOrder(req); err != nil {
        return OrderResponse{}, err
    }
    
    middleware.AddSpanEvent(c, "saving to database")
    order, err := service.Save(req)
    if err != nil {
        middleware.RecordError(c, err)
        return OrderResponse{}, echo.NewHTTPError(500, err.Error())
    }
    
    middleware.AddSpanEvent(c, "order created",
        attribute.String("order_id", order.ID),
        attribute.Int("quantity", order.Quantity),
    )
    
    return ToOrderResponse(order), nil
}
```

### Recording Errors

Capture errors in traces for debugging:

```go
func getUser(c echo.Context) (UserResponse, error) {
    id := c.Param("id")
    
    user, err := service.GetByID(id)
    if err != nil {
        // Record error in trace
        middleware.RecordError(c, err)
        return UserResponse{}, echo.NewHTTPError(404, "user not found")
    }
    
    return ToUserResponse(user), nil
}
```

### Getting Trace Context

Access trace information in handlers:

```go
func handler(c echo.Context) error {
    traceID := middleware.GetTraceID(c)
    spanID := middleware.GetSpanID(c)
    requestID := middleware.GetRequestID(c)
    
    log.Printf("Processing request %s (trace: %s)", requestID, traceID)
    
    return c.JSON(200, map[string]string{
        "trace_id":   traceID,
        "span_id":    spanID,
        "request_id": requestID,
    })
}
```

### Custom Span Attributes

Add custom attributes to spans:

```go
app.Use(middleware.OTELMiddleware("my-service",
    middleware.WithCustomAttributes(
        attribute.String("deployment.region", "us-east-1"),
        attribute.String("deployment.zone", "1a"),
    ),
))

// In handlers, add dynamic attributes
func handler(c echo.Context) error {
    middleware.AddSpanEvent(c, "processing",
        attribute.String("user_id", getUserID(c)),
        attribute.String("tenant", getTenant(c)),
    )
    // ...
}
```

## Service-to-Service Communication

### Traced HTTP Client

Use the traced HTTP client for outgoing requests:

```go
// Create traced client
tracedClient := middleware.NewTracedHTTPClient(
    middleware.WithClientTimeout(30 * time.Second),
)

func handler(c echo.Context) error {
    // Get context from incoming request
    ctx := c.Request().Context()
    
    // Make traced outgoing request
    // Trace context is automatically propagated
    resp, err := tracedClient.Get(ctx, "https://api.example.com/data")
    if err != nil {
        middleware.RecordError(c, err)
        return echo.NewHTTPError(500, "external API error")
    }
    defer resp.Body.Close()
    
    // Process response...
    return c.JSON(200, data)
}
```

### Making POST Requests

```go
func callExternalAPI(c echo.Context, data interface{}) error {
    ctx := c.Request().Context()
    
    // Marshal data
    body, err := json.Marshal(data)
    if err != nil {
        return err
    }
    
    // Create traced request
    req, err := middleware.NewRequestWithTrace(
        c,
        "POST",
        "https://payment-service.example.com/process",
        bytes.NewReader(body),
    )
    if err != nil {
        return err
    }
    
    req.Header.Set("Content-Type", "application/json")
    
    // Execute with traced client
    resp, err := tracedClient.Do(req)
    if err != nil {
        middleware.RecordError(c, err)
        return err
    }
    defer resp.Body.Close()
    
    return nil
}
```

### Wrapping Existing Clients

```go
// Wrap an existing http.Client
existingClient := &http.Client{
    Timeout: 30 * time.Second,
    // ... other config
}

tracedClient := middleware.WrapHTTPClient(existingClient)
```

### Distributed Trace Example

When Service A calls Service B:

```
Service A (order-service)
    └── POST /orders
        ├── Span: validate request
        ├── Span: save to database
        └── GET https://payment-service/validate
            └── Service B receives trace context
                └── Span: validate payment method
```

Jaeger shows the complete distributed trace across both services.

## Metrics Collection

### Request Metrics

The OTEL middleware automatically collects:

- Request count
- Request duration
- Response status codes
- Error rates

### Custom Metrics

Add custom metrics with Prometheus:

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    ordersCreated = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "orders_created_total",
        Help: "Total number of orders created",
    })
    
    orderValue = prometheus.NewHistogram(prometheus.HistogramOpts{
        Name:    "order_value_dollars",
        Help:    "Order values in dollars",
        Buckets: prometheus.LinearBuckets(0, 50, 20),
    })
)

func init() {
    prometheus.MustRegister(ordersCreated)
    prometheus.MustRegister(orderValue)
}

func createOrder(c echo.Context, req CreateOrderRequest) (OrderResponse, error) {
    order, err := service.Create(req)
    if err != nil {
        return OrderResponse{}, err
    }
    
    // Update metrics
    ordersCreated.Inc()
    orderValue.Observe(order.TotalAmount)
    
    return ToOrderResponse(order), nil
}

// Expose metrics endpoint
app.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
```

### EchoNext Metrics Middleware

Use the built-in metrics middleware:

```go
import "github.com/abdussamadbello/echonext/pkg/contrib/middleware"

metrics := middleware.NewMetrics()
app.Use(middleware.MetricsMiddleware(metrics))

// Expose metrics
app.GET("/metrics", middleware.MetricsHandler(metrics))
```

## Logging Integration

### Structured Logging with Trace Context

Include trace IDs in logs:

```go
import "github.com/labstack/gommon/log"

func handler(c echo.Context) error {
    traceID := middleware.GetTraceID(c)
    requestID := middleware.GetRequestID(c)
    
    c.Logger().Infof(
        "Processing request trace_id=%s request_id=%s",
        traceID,
        requestID,
    )
    
    // Process request...
    return c.JSON(200, data)
}
```

### Structured Logger Middleware

Use the structured logger middleware:

```go
app.Use(middleware.StructuredLogger(middleware.StructuredLoggerConfig{
    CustomFields: func(c echo.Context) map[string]interface{} {
        return map[string]interface{}{
            "trace_id":   middleware.GetTraceID(c),
            "request_id": middleware.GetRequestID(c),
            "user_agent": c.Request().UserAgent(),
        }
    },
}))
```

### JSON Logging

```go
import "encoding/json"

type LogEntry struct {
    Timestamp string `json:"timestamp"`
    Level     string `json:"level"`
    Message   string `json:"message"`
    TraceID   string `json:"trace_id"`
    RequestID string `json:"request_id"`
    Extra     map[string]interface{} `json:"extra,omitempty"`
}

func logJSON(c echo.Context, level, message string, extra map[string]interface{}) {
    entry := LogEntry{
        Timestamp: time.Now().UTC().Format(time.RFC3339),
        Level:     level,
        Message:   message,
        TraceID:   middleware.GetTraceID(c),
        RequestID: middleware.GetRequestID(c),
        Extra:     extra,
    }
    
    data, _ := json.Marshal(entry)
    fmt.Println(string(data))
}
```

## Production Best Practices

### 1. Sampling Strategy

Don't trace every request in production:

```go
cfg := middleware.OTELConfig{
    ServiceName: "my-service",
    SampleRate:  0.1, // Sample 10% of requests
}
```

### 2. Secure Connections

Use TLS for OTLP:

```go
cfg := middleware.OTELConfig{
    Endpoint: "collector.example.com:4317",
    Insecure: false, // Use TLS
}
```

### 3. Resource Attributes

Add deployment metadata:

```go
cfg := middleware.OTELConfig{
    ServiceName:    "my-service",
    ServiceVersion: os.Getenv("APP_VERSION"),
    Environment:    os.Getenv("ENVIRONMENT"),
}
```

### 4. Skip Health Checks

Don't trace health check endpoints:

```go
app.Use(middleware.OTELMiddleware("my-service",
    middleware.WithSkipper(func(c echo.Context) bool {
        return c.Path() == "/health" || c.Path() == "/ready"
    }),
))
```

### 5. Graceful Shutdown

Ensure traces are flushed:

```go
shutdown, err := middleware.InitOTEL(ctx, cfg)
if err != nil {
    log.Fatal(err)
}

// Defer with timeout
defer func() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    shutdown.Shutdown(ctx)
}()
```

### 6. Error Handling

Don't fail startup if OTEL is unavailable:

```go
shutdown, err := middleware.InitOTEL(ctx, cfg)
if err != nil {
    log.Printf("OTEL init failed (running without tracing): %v", err)
} else {
    defer shutdown.Shutdown(ctx)
}
```

## Backend Integration

### Jaeger

Development setup:

```bash
docker run -d \
  --name jaeger \
  -p 16686:16686 \
  -p 4317:4317 \
  jaegertracing/all-in-one:latest
```

Access UI: http://localhost:16686

### Zipkin

```bash
docker run -d \
  --name zipkin \
  -p 9411:9411 \
  openzipkin/zipkin
```

Configure:

```go
cfg := middleware.OTELConfig{
    Endpoint: "localhost:9411",
}
```

### Datadog

Set environment:

```bash
export OTEL_EXPORTER_OTLP_ENDPOINT="http://localhost:4317"
export DD_TRACE_ENABLED="true"
```

### Honeycomb

```bash
export OTEL_EXPORTER_OTLP_ENDPOINT="api.honeycomb.io:443"
export OTEL_EXPORTER_OTLP_HEADERS="x-honeycomb-team=YOUR_API_KEY"
```

### AWS X-Ray

```bash
export OTEL_EXPORTER_OTLP_ENDPOINT="localhost:4317"
# Run AWS OTEL Collector as sidecar
```

## Troubleshooting

### No Traces Appearing

1. **Check OTEL collector is running:**

```bash
# For Jaeger
docker ps | grep jaeger

# Check port
nc -zv localhost 4317
```

2. **Check application logs:**

```
OTEL initialization failed: ...
```

3. **Verify configuration:**

```go
log.Printf("OTEL Config: %+v", cfg)
```

### High Memory Usage

Reduce sample rate:

```go
cfg.SampleRate = 0.01 // Sample 1%
```

### Missing Spans

Ensure traced client is used:

```go
// ❌ Wrong - uses default http.Client
resp, err := http.Get(url)

// ✅ Correct - uses traced client
resp, err := tracedClient.Get(ctx, url)
```

### Trace Context Not Propagating

Ensure context is passed:

```go
// ❌ Wrong - creates new context
ctx := context.Background()

// ✅ Correct - uses request context
ctx := c.Request().Context()
```

### Performance Impact

OTEL has minimal overhead (~1-2% CPU). To verify:

```bash
go test -bench=. -benchmem
```

## Complete Example

See the [OTEL Demo](../../examples/otel-demo/) for a complete working example with:

- OTEL initialization
- Incoming request tracing
- Outgoing request tracing
- Span events and attributes
- Error recording
- Request ID correlation
- Integration with Jaeger

```bash
# Run the example
cd examples/otel-demo
docker run -d -p 16686:16686 -p 4317:4317 jaegertracing/all-in-one
go run main.go

# Make requests
curl http://localhost:8080/demo/external-call

# View traces
open http://localhost:16686
```

## Next Steps

- [OpenTelemetry Demo Example](../../examples/otel-demo/README.md)
- [Middleware Documentation](../../pkg/contrib/middleware/doc.go)
- [Deployment Guide](../guides/deployment.md)
- [Performance Optimization](./performance.md)

## External Resources

- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
- [OpenTelemetry Go SDK](https://opentelemetry.io/docs/instrumentation/go/)
- [Jaeger Documentation](https://www.jaegertracing.io/docs/)
- [W3C Trace Context](https://www.w3.org/TR/trace-context/)
