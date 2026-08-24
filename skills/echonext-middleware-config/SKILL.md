---
name: echonext-middleware-config
description: >-
  Register, order and write EchoNext middleware, and load application
  configuration. Use when adding middleware (auth, CORS, logging, request IDs,
  metrics, tracing) to an echonext app or group, writing a custom
  echo.MiddlewareFunc, or loading config from YAML/env with the contrib config
  package.
license: MIT
metadata:
  version: 0.1.0
---

# EchoNext Middleware & Config

`*App` embeds `*echo.Echo`, so middleware is plain Echo middleware —
`echo.MiddlewareFunc`. EchoNext adds no middleware type of its own. Config is
handled by the optional `pkg/contrib/config` package (a thin Viper wrapper).

## Three packages are called `middleware`

This trips up every import block. Alias them:

```go
import (
    echomw "github.com/labstack/echo/v5/middleware"                     // Echo built-ins
    "github.com/abdussamadbello/echonext/pkg/contrib/middleware"        // contrib helpers
    appmw "myapp/internal/middleware"                                   // your own
)
```

Echo's package and the contrib package **both export `RequestID`** with
different configs, so an unaliased import silently picks the wrong one.

## Registering middleware

```go
app := echonext.New()

// Global — runs for every route, in registration order.
app.Use(echomw.Recover())
app.Use(middleware.RequestID())

// Group-scoped — passed at construction, or added later with Use.
api := app.Group("/api/v1", appmw.APIKeyAuth(cfg.JWT.Secret))
api.Use(echomw.Gzip())

// Sub-groups inherit the parent's middleware.
admin := api.Group("/admin", appmw.RequireRole("admin"))
```

`echonext.Route{}` has **no** middleware field — it carries OpenAPI metadata
only. To scope middleware to a single route, put that route in its own group.

Order matters: `Recover` first so it catches panics from everything after it,
`RequestID` before any logger that reports the ID, auth before handlers.

## Echo v5 gotchas

- **`echomw.CORS()` panics with no arguments** — "at least one AllowOrigins is
  required". Echo v4 defaulted to `*`; v5 makes you say so:

  ```go
  app.Use(echomw.CORS("*"))                       // explicit permissive
  app.Use(echomw.CORS("https://app.example.com")) // or a real origin list
  ```

- **`echomw.Logger()` no longer exists** — it is `echomw.RequestLogger()`.
- Middleware signatures take `*echo.Context` (a pointer), not `echo.Context`.

## Contrib middleware

`pkg/contrib/middleware` complements Echo's set rather than replacing it.

### RequestID

```go
app.Use(middleware.RequestID())              // UUID v4, X-Request-ID header
id := middleware.GetRequestID(c)             // read it in a handler

app.Use(middleware.RequestIDWithConfig(middleware.RequestIDConfig{
    RequestIDHeader: "X-Correlation-ID",
    ContextKey:      "correlation_id",
    Generator:       func() string { return myULID() },
    Skipper:         func(c *echo.Context) bool { return c.Path() == "/health" },
}))
```

### Metrics

```go
metrics := middleware.NewMetrics()
app.Use(middleware.MetricsMiddleware(metrics))
```

`metrics.GetMetrics()` returns `total_requests`, `total_errors`,
`avg_duration_ms`, `requests_by_code`. Two ways to expose it:

```go
// Type-safe route — appears in the OpenAPI spec. Preferred.
app.GET("/metrics", func(c *echo.Context) (map[string]interface{}, error) {
    return metrics.GetMetrics(), nil
}, echonext.Route{Summary: "Request metrics", Tags: []string{"System"}})

// Raw Echo route — not in the spec.
app.Echo.GET("/metrics", middleware.MetricsHandler(metrics))
```

`MetricsHandler` returns an `echo.HandlerFunc`, which writes the response
itself. Register it on `app.Echo`, **not** on `app.GET` — the type-safe router
wraps handler return values and will try to write a second response, logging
"echo: response already written to client".

### StructuredLogger

Logs one slog line per request through `c.Logger()`, at Info/Warn/Error by
status class, and folds in the request ID automatically:

```go
app.Use(middleware.StructuredLogger(middleware.StructuredLoggerConfig{
    Skipper: echomw.DefaultSkipper,
    CustomFields: func(c *echo.Context) map[string]interface{} {
        return map[string]interface{}{"user_id": c.Get("user_id")}
    },
}))
```

Pick either this or `echomw.RequestLogger()` — enabling both logs every
request twice.

### OpenTelemetry

`middleware.InitOTEL` + `middleware.OTELMiddleware` give tracing and metrics,
configured from `OTEL_*` environment variables, plus a traced HTTP client for
outgoing calls. See [`references/otel.md`](references/otel.md).

## Writing custom middleware

The shape is a function returning `echo.MiddlewareFunc`:

```go
package middleware

func APIKeyAuth(validKey string) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c *echo.Context) error {
            if c.Request().Header.Get("X-API-Key") != validKey {
                return echo.NewHTTPError(http.StatusUnauthorized, "invalid api key")
            }
            c.Set("user_id", lookupUser(c))   // handlers read this with c.Get
            return next(c)                     // omit to short-circuit
        }
    }
}
```

Add a `Config` struct with a `Skipper` when the middleware needs options:

```go
type AuthConfig struct {
    Skipper echomw.Skipper
    Key     string
}

func AuthWithConfig(cfg AuthConfig) echo.MiddlewareFunc {
    if cfg.Skipper == nil {
        cfg.Skipper = echomw.DefaultSkipper
    }
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c *echo.Context) error {
            if cfg.Skipper(c) {
                return next(c)
            }
            // ...
            return next(c)
        }
    }
}
```

**Error shape differs from handlers.** An error returned from middleware is
rendered by Echo's error handler as `{"message":"invalid api key"}`. It does
*not* go through EchoNext's `Response[T]` envelope, so it lacks the
`"success": false` field that a handler error produces. If clients depend on a
uniform envelope, install a custom `app.HTTPErrorHandler`.

Scaffold one with the CLI — it writes `internal/middleware/<name>.go` with both
the plain and `WithConfig` variants:

```bash
echonext generate middleware auth
```

## Configuration

### In a scaffolded project

`echonext init` generates `internal/config/config.go` (plain Viper) plus
`configs/{development,production,test}.yaml`. `config.Load()` picks the file by
`APP_ENV` (default `development`) and overlays env vars prefixed with the
uppercased project name:

```bash
APP_ENV=production MYAPP_APP_PORT=9090 MYAPP_DATABASE_DSN=postgres://... ./api
```

Precedence: explicit env var → config file → `viper.SetDefault` value.

### With the contrib package

`pkg/contrib/config` is optional and generic over your config struct:

```go
import "github.com/abdussamadbello/echonext/pkg/contrib/config"

type Config struct {
    App      config.AppConfig      `mapstructure:"app"`
    Database config.DatabaseConfig `mapstructure:"database"`
    Server   config.ServerConfig   `mapstructure:"server"`
    CORS     config.CORSConfig     `mapstructure:"cors"`
}

var cfg Config
if err := config.LoadSimple(&cfg); err != nil {   // ./configs + ., yaml, env
    log.Fatal(err)
}
```

| Function | Use for |
|---|---|
| `config.LoadSimple(&cfg)` | defaults: name `config`, yaml, `./configs` and `.` |
| `config.LoadWithEnv(&cfg, "MYAPP")` | same, with an env-var prefix |
| `config.Load(&cfg, opts)` | full control via `LoadOptions` |
| `config.LoadFromFile(&cfg, path)` | one exact file, no env binding |
| `config.Watch(&cfg, opts, cb)` | reload on file change |

A missing config file is **not** an error for `Load`/`LoadSimple`/`LoadWithEnv`
— they fall through to env vars and zero values. `LoadFromFile` *does* error.

Prebuilt structs in `pkg/contrib/config`: `AppConfig`, `DatabaseConfig`,
`CacheConfig`, `LoggerConfig`, `ServerConfig`, `JWTConfig`, `CORSConfig`.
Duration fields are strings parsed on demand —
`cfg.Server.ParseReadTimeout()`, `cfg.Database.ParseMaxLifetime()`.

```go
opts := config.DefaultLoadOptions()
opts.ConfigName = os.Getenv("APP_ENV")   // development / production / test
opts.EnvPrefix  = "MYAPP"
opts.WatchConfig = true
err := config.Load(&cfg, opts)
```

### Feeding config into middleware

```go
app.Use(echomw.CORSWithConfig(echomw.CORSConfig{
    AllowOrigins:     cfg.CORS.AllowOrigins,
    AllowMethods:     cfg.CORS.AllowMethods,
    AllowCredentials: cfg.CORS.AllowCredentials,
    MaxAge:           cfg.CORS.MaxAge,
}))
```

`AllowOrigins` must be non-empty here too, so validate it after loading —
an empty `cors.allow_origins` in YAML panics at startup.

## Quick checklist

- Alias the Echo middleware import as `echomw`; the contrib one keeps the name.
- `Recover` → `RequestID` → logger → auth, in that order.
- Never call `echomw.CORS()` bare; pass origins.
- Route-scoped middleware means a group, not a `Route{}` field.
- Custom middleware returns `echo.MiddlewareFunc` and calls `next(c)`.
- Register raw `echo.HandlerFunc` values on `app.Echo`, not `app.GET`.
- Middleware errors bypass the `Response[T]` envelope.
- `LoadFromFile` errors on a missing file; the other loaders do not.
