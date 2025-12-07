# OpenTelemetry Demo

This example demonstrates how to use EchoNext's OpenTelemetry (OTEL) middleware for automatic distributed tracing of incoming and outgoing HTTP requests.

## Features Demonstrated

- **OTEL Initialization** - Configure tracing and metrics exporters
- **Incoming Request Tracing** - Automatic spans for all HTTP requests
- **Outgoing Request Tracing** - Traced HTTP client for downstream services
- **Span Events** - Adding timeline markers within a trace
- **Error Recording** - Capturing errors in traces for debugging
- **Request ID Correlation** - X-Request-ID header propagation
- **Trace Context Propagation** - W3C traceparent/tracestate headers

## Prerequisites

### 1. Start Jaeger (Trace Viewer)

```bash
docker run -d \
  --name jaeger \
  -p 16686:16686 \
  -p 4317:4317 \
  jaegertracing/all-in-one:latest
```

- **Jaeger UI**: http://localhost:16686
- **OTLP gRPC**: localhost:4317

### 2. Dependencies

```bash
go mod tidy
```

## Running the Example

```bash
# From the echonext repository root
go run examples/otel-demo/main.go

# Or from this directory
go run main.go
```

**Endpoints:**
- API Server: http://localhost:8080
- API Docs: http://localhost:8080/api/docs
- Jaeger UI: http://localhost:16686

## API Endpoints

### Business Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check with trace ID |
| POST | `/orders` | Create order (with span events) |
| GET | `/orders` | List all orders |
| GET | `/orders/:id` | Get order by ID |

### OTEL Demo Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/demo/external-call` | Makes traced outgoing HTTP request |
| GET | `/demo/trace-info` | Shows current trace/span/request IDs |
| GET | `/demo/nested-spans` | Demonstrates span events for sub-operations |
| GET | `/demo/error-recording` | Shows how errors appear in traces |

## Try It Out

### 1. View Basic Traces

```bash
# Make a request
curl http://localhost:8080/health

# Open Jaeger UI
open http://localhost:16686

# Select "order-service" and click "Find Traces"
```

### 2. See Distributed Tracing

```bash
# This makes an outgoing HTTP request to httpbin.org
curl http://localhost:8080/demo/external-call
```

In Jaeger, you'll see:
- Parent span for the incoming request
- Child span for the outgoing HTTP request to httpbin.org

### 3. See Span Events

```bash
curl http://localhost:8080/demo/nested-spans
```

In Jaeger, expand the span to see event markers in the timeline.

### 4. See Error Recording

```bash
curl http://localhost:8080/demo/error-recording
```

In Jaeger, the span will be marked with an error icon.

### 5. Create Orders with Tracing

```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"product": "Test Product", "quantity": 5}'
```

## Code Walkthrough

### 1. Initialize OTEL

```go
shutdown, err := middleware.InitOTEL(ctx, middleware.OTELConfig{
    ServiceName:    "order-service",
    ServiceVersion: "1.0.0",
    Environment:    "development",
    Endpoint:       "localhost:4317",
    Insecure:       true,
    SampleRate:     1.0,
    EnableTracing:  true,
    EnableMetrics:  true,
})
if err != nil {
    log.Printf("OTEL init failed: %v", err)
} else {
    defer shutdown.Shutdown(ctx)
}
```

### 2. Add Middleware

```go
// RequestID adds X-Request-ID header
app.Use(middleware.RequestID())

// OTELMiddleware instruments all incoming requests
app.Use(middleware.OTELMiddleware("order-service"))
```

### 3. Create Traced HTTP Client

```go
tracedClient := middleware.NewTracedHTTPClient(
    middleware.WithClientTimeout(30 * time.Second),
)

// Use in handlers
resp, err := tracedClient.Get(ctx, "https://api.example.com/data")
```

### 4. Add Span Events

```go
func myHandler(c echo.Context) error {
    // Add event with attributes
    middleware.AddSpanEvent(c, "processing order",
        attribute.String("order_id", "123"),
        attribute.Int("quantity", 5),
    )
    // ...
}
```

### 5. Record Errors

```go
if err != nil {
    middleware.RecordError(c, err)
    return echo.NewHTTPError(500, "something went wrong")
}
```

### 6. Get Trace Context

```go
traceID := middleware.GetTraceID(c)
spanID := middleware.GetSpanID(c)
requestID := middleware.GetRequestID(c)
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `OTEL_SERVICE_NAME` | Service name in traces | (from config) |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | OTLP collector endpoint | localhost:4317 |
| `OTEL_TRACES_SAMPLER` | Sampling strategy | parentbased_always_on |

## Integration with Other Services

When your service calls another traced service:

1. The `TracedHTTPClient` automatically injects the `traceparent` header
2. The downstream service extracts this context
3. Jaeger shows the complete distributed trace across services

```
Service A (order-service)
    └── POST /orders
        └── GET https://payment-service/validate
            └── Span in payment-service
```

## Production Considerations

1. **Sampling**: Use lower sample rates in production
   ```go
   SampleRate: 0.1, // Sample 10% of requests
   ```

2. **Secure Endpoint**: Use TLS for OTLP
   ```go
   Insecure: false,
   ```

3. **Environment-Based Config**:
   ```go
   cfg := middleware.OTELConfig{
       ServiceName: os.Getenv("OTEL_SERVICE_NAME"),
       Endpoint:    os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
   }
   ```

## Troubleshooting

### Traces Not Appearing in Jaeger

1. Verify Jaeger is running:
   ```bash
   docker ps | grep jaeger
   ```

2. Check OTLP port is accessible:
   ```bash
   nc -zv localhost 4317
   ```

3. Look for connection errors in app logs

### App Works But No Tracing

This is expected if Jaeger isn't running. The app gracefully degrades:
```
OTEL initialization failed (running without tracing): ...
```

## Related Documentation

- [EchoNext OTEL Middleware](../../pkg/contrib/middleware/doc.go)
- [OpenTelemetry Go](https://opentelemetry.io/docs/instrumentation/go/)
- [Jaeger Documentation](https://www.jaegertracing.io/docs/)
