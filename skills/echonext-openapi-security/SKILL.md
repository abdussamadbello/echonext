---
name: echonext-openapi-security
description: >-
  Configure EchoNext's auto-generated OpenAPI spec and authentication. Use when
  setting API metadata (title/servers/contact), serving the OpenAPI JSON or
  Swagger UI, or adding auth (bearer/JWT, API key, OAuth2, OpenID Connect) via
  security schemes and per-route or global security requirements.
license: MIT
metadata:
  version: 0.1.0
---

# EchoNext OpenAPI & Security

EchoNext generates an OpenAPI 3.0 spec from your type-safe routes and structs
automatically. You only add metadata (info, servers, security) and choose how
to serve it. Security schemes are declared once, then referenced by name on
routes or globally.

## API metadata

```go
app := echonext.New()

app.SetInfo("Todo API", "1.0.0", "Manage todos")
app.SetContact("API Team", "https://example.com", "api@example.com")
app.SetLicense("MIT", "https://opensource.org/licenses/MIT")
app.SetServers([]echonext.Server{
    {URL: "https://api.example.com", Description: "production"},
    {URL: "http://localhost:8080",   Description: "local"},
})
```

## Serving the spec and docs

```go
app.ServeOpenAPISpec("/api/openapi.json")          // raw OpenAPI JSON
app.ServeSwaggerUI("/api/docs", "/api/openapi.json") // Swagger UI at /api/docs
```

If you need the spec object directly (e.g. to write it to a file in a build
step): `spec := app.GenerateOpenAPISpec()`.

## Per-route OpenAPI metadata

The `Route{}` options on any route feed the spec:

```go
app.POST("/todos", createTodo, echonext.Route{
    Summary:       "Create a todo",
    Description:   "Creates a todo and returns it.",
    Tags:          []string{"Todos"},
    SuccessStatus: http.StatusCreated,
})
```

Request/response schemas and parameters are derived from your handler's typed
`req`/return structs and their `json:`/`query:`/`param:`/`validate:` tags — you
do not hand-write schemas.

## Authentication

### 1. Declare a security scheme (once)

```go
// Bearer / JWT
app.AddSecurityScheme("bearerAuth", echonext.Security{
    Type:   "bearer",
    Scheme: "JWT",
})

// API key in a header
app.AddSecurityScheme("apiKey", echonext.Security{
    Type: "apiKey",
    In:   "header",
    Name: "X-API-Key",
})
```

### 2. Require it — globally or per route

```go
// Global default for every route:
app.SetGlobalSecurity(echonext.SecurityRequirement{SchemeName: "bearerAuth"})

// Per route (and opt a route OUT of the global default):
app.GET("/public/health", health, echonext.Route{
    DisableGlobalSecurity: true,
})

// Per route requirement with OAuth2 scopes:
app.GET("/admin", adminHandler, echonext.Route{
    Security: []echonext.SecurityRequirement{
        {SchemeName: "oauth2", Scopes: []string{"admin"}},
    },
})
```

> Security schemes document the API and drive Swagger UI's "Authorize" button.
> They do **not** enforce auth by themselves — enforce it with middleware
> (e.g. an Echo JWT middleware) on the routes/groups that need it.

See [`references/security-schemes.md`](references/security-schemes.md) for the
full `Security` field reference and OAuth2 / OpenID Connect setups.

## Checklist

- `SetInfo` + `SetServers` early in setup.
- `AddSecurityScheme(name, …)` before referencing it.
- Use `SetGlobalSecurity` for the common case; `DisableGlobalSecurity` /
  `Route.Security` for exceptions.
- Add real enforcement middleware — schemes are documentation, not a guard.
