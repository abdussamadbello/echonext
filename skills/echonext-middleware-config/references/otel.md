# OpenTelemetry middleware

`pkg/contrib/middleware` ships an OTLP-based tracing and metrics setup plus an
instrumented HTTP client, so traces span service boundaries.

## Initialise once, at startup

```go
import (
    "context"
    "github.com/abdussamadbello/echonext"
    "github.com/abdussamadbello/echonext/pkg/contrib/middleware"
)

func main() {
    ctx := context.Background()

    shutdown, err := middleware.InitOTEL(ctx, middleware.DefaultOTELConfig())
    if err != nil {
        log.Fatal(err)
    }
    defer shutdown.Shutdown(ctx)

    app := echonext.New()
    app.Use(middleware.RequestID())            // correlate logs with traces
    app.Use(middleware.OTELMiddleware("my-service"))

    app.Start(":8080")
}
```

`InitOTEL` must run before `OTELMiddleware`; the middleware only creates spans,
the init call wires up the exporters and providers. `Shutdown` flushes both the
tracer and meter providers and returns the first error.

## Configuration

`DefaultOTELConfig()` reads the environment, so most deployments need no code
change:

| Variable | Default |
|---|---|
| `OTEL_SERVICE_NAME` | `echonext-service` |
| `OTEL_SERVICE_VERSION` | `1.0.0` |
| `OTEL_ENVIRONMENT` | `development` |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `localhost:4317` |
| `OTEL_INSECURE` | `true` |
| `OTEL_SAMPLE_RATE` | `1.0` |
| `OTEL_ENABLE_TRACING` | `true` |
| `OTEL_ENABLE_METRICS` | `true` |

Override any field in code:

```go
cfg := middleware.DefaultOTELConfig()
cfg.ServiceName = "orders-api"
cfg.SampleRate  = 0.1
cfg.CustomAttributes = []attribute.KeyValue{
    attribute.String("deployment.region", "eu-west-1"),
}
shutdown, err := middleware.InitOTEL(ctx, cfg)
```

`OTELConfig` fields: `ServiceName`, `ServiceVersion`, `Environment`,
`Endpoint`, `Insecure`, `SampleRate`, `EnableTracing`, `EnableMetrics`,
`Skipper`, `CustomAttributes`, `PropagateHeaders`.

## Middleware options

```go
app.Use(middleware.OTELMiddleware("my-service",
    middleware.WithSkipper(func(c *echo.Context) bool {
        return c.Path() == "/health" || c.Path() == "/metrics"
    }),
    middleware.WithCustomAttributes(attribute.String("tier", "public")),
))
```

`middleware.OTELTracingMiddleware(name)` is the tracing-only variant.

## Span helpers inside handlers

```go
func getOrder(c *echo.Context, req GetOrderRequest) (Order, error) {
    middleware.AddSpanEvent(c, "loading order")
    middleware.SetSpanAttributes(c, attribute.String("order.id", req.ID))

    order, err := store.Get(req.ID)
    if err != nil {
        middleware.RecordError(c, err)
        return Order{}, err
    }

    // Nested span for an expensive step.
    _, span := middleware.StartSpan(c, "enrich-order")
    defer span.End()

    return order, nil
}
```

`middleware.GetTraceID(c)` and `middleware.GetSpanID(c)` return the current IDs
as strings — useful in log lines and error responses.

## Traced outgoing HTTP

Propagates W3C `traceparent`/`tracestate` so the downstream service's spans
attach to the caller's trace.

```go
client := middleware.NewTracedHTTPClient(
    middleware.WithClientTimeout(10 * time.Second),
)

app.GET("/users/:id", func(c *echo.Context, req GetUserRequest) (User, error) {
    resp, err := client.Get(c.Request().Context(), "http://user-service/api/users/"+req.ID)
    if err != nil {
        middleware.RecordError(c, err)
        return User{}, err
    }
    defer resp.Body.Close()
    // ...
})
```

Pass `c.Request().Context()`, not `context.Background()` — the request context
carries the parent span, and without it the outgoing span starts a new trace.

Other entry points:

- `middleware.WrapHTTPClient(existing)` — wrap a client you already configured
- `middleware.WrapTransport(rt)` — wrap a bare `http.RoundTripper`
- `middleware.NewRequestWithTrace(c, method, url, body)` — build an
  `*http.Request` carrying trace context, for use with `client.Do`
- `middleware.InitDefaultTracedClient(opts...)` — populate the package-level
  `middleware.DefaultTracedClient`
