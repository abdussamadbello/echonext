# Security Schemes — Reference

The `echonext.Security` struct passed to `app.AddSecurityScheme(name, sec)`
(defined in `echonext.go`).

```go
type Security struct {
    Type             string       // "bearer" | "apiKey" | "oauth2" | "basic" | "openIdConnect"
    Name             string       // apiKey: header/query/cookie name (e.g. "X-API-Key")
    Scheme           string       // bearer: "bearer"/"JWT"; basic: "basic"
    In               string       // apiKey: "header" | "query" | "cookie"
    Description      string       // human-readable description
    OAuth2Flows      *OAuth2Flows // required when Type == "oauth2"
    OpenIDConnectURL string       // required when Type == "openIdConnect"
}
```

## Bearer / JWT

```go
app.AddSecurityScheme("bearerAuth", echonext.Security{
    Type:        "bearer",
    Scheme:      "JWT",
    Description: "JWT access token",
})
```

## API key

```go
app.AddSecurityScheme("apiKey", echonext.Security{
    Type: "apiKey",
    In:   "header", // or "query" / "cookie"
    Name: "X-API-Key",
})
```

## Basic auth

```go
app.AddSecurityScheme("basicAuth", echonext.Security{
    Type:   "basic",
    Scheme: "basic",
})
```

## OAuth2

```go
type OAuth2Flows struct {
    Implicit          *OAuth2Flow
    Password          *OAuth2Flow
    ClientCredentials *OAuth2Flow
    AuthorizationCode *OAuth2Flow
}

type OAuth2Flow struct {
    AuthorizationURL string            // implicit, authorizationCode
    TokenURL         string            // password, clientCredentials, authorizationCode
    RefreshURL       string            // optional
    Scopes           map[string]string // scope name -> description
}
```

```go
app.AddSecurityScheme("oauth2", echonext.Security{
    Type: "oauth2",
    OAuth2Flows: &echonext.OAuth2Flows{
        AuthorizationCode: &echonext.OAuth2Flow{
            AuthorizationURL: "https://auth.example.com/authorize",
            TokenURL:         "https://auth.example.com/token",
            Scopes: map[string]string{
                "read":  "Read access",
                "write": "Write access",
                "admin": "Admin access",
            },
        },
    },
})

// Require specific scopes on a route:
app.POST("/admin/users", createUser, echonext.Route{
    Security: []echonext.SecurityRequirement{
        {SchemeName: "oauth2", Scopes: []string{"admin", "write"}},
    },
})
```

## OpenID Connect

```go
app.AddSecurityScheme("oidc", echonext.Security{
    Type:             "openIdConnect",
    OpenIDConnectURL: "https://auth.example.com/.well-known/openid-configuration",
})
```

## Requirements: global vs per-route

```go
// Apply a default to all routes:
app.SetGlobalSecurity(
    echonext.SecurityRequirement{SchemeName: "bearerAuth"},
)

// Clear it later if needed:
app.ClearGlobalSecurity()

// Opt one route out of the global default:
app.GET("/healthz", health, echonext.Route{DisableGlobalSecurity: true})

// Override per route:
app.GET("/admin", admin, echonext.Route{
    Security: []echonext.SecurityRequirement{
        {SchemeName: "oauth2", Scopes: []string{"admin"}},
    },
})
```

`SecurityRequirement`:

```go
type SecurityRequirement struct {
    SchemeName string   // must match a name passed to AddSecurityScheme
    Scopes     []string // OAuth2 scopes (empty for non-OAuth2 schemes)
}
```

`AddSecurityScheme` and `SetGlobalSecurity` return an `error` (e.g. unknown
scheme name, missing OAuth2 flows) — check it during setup.
