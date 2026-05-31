# GraphQL — Reference

Package `github.com/abdussamadbello/echonext/graphql`. Integrates
[gqlgen](https://gqlgen.com): you provide a generated `ExecutableSchema`, and
EchoNext mounts the endpoint + playground and injects the Echo context into
resolvers.

## Config

```go
type Config struct {
    Path                string                    // default "/graphql"
    PlaygroundPath      string                    // default "/playground"
    Schema              graphql.ExecutableSchema  // gqlgen-generated schema (required)
    ComplexityLimit     int                       // 0 = unlimited
    QueryCacheSize      int                       // default 1000
    EnableIntrospection bool
    EnableTracing       bool
    WebSocketUpgrader   *websocket.Upgrader       // for subscriptions
    RecoverFunc         graphql.RecoverFunc
    ErrorPresenter      graphql.ErrorPresenterFunc
}

type Route struct {
    Summary           string
    Description       string
    Tags              []string
    DocumentInOpenAPI bool
}

func DefaultConfig(schema graphql.ExecutableSchema) Config
```

## Mounting

```go
import (
    echographql "github.com/abdussamadbello/echonext/graphql"
    "yourapp/graph/generated"
    "yourapp/graph"
)

schema := generated.NewExecutableSchema(generated.Config{
    Resolvers: &graph.Resolver{ /* deps */ },
})

app.GraphQL(echographql.Config{
    Path:                "/graphql",
    PlaygroundPath:      "/playground",
    Schema:              schema,
    EnableIntrospection: true,
    QueryCacheSize:      1000,
}, echographql.Route{
    Summary:     "GraphQL API",
    Description: "Primary GraphQL endpoint",
    Tags:        []string{"GraphQL"},
})
```

Or accept defaults: `app.GraphQL(echographql.DefaultConfig(schema))`.

## Accessing the request inside resolvers

The Echo context is propagated through `context.Context`, so resolvers can read
headers, auth, etc.:

```go
func (r *queryResolver) Me(ctx context.Context) (*model.User, error) {
    ec := echographql.GetEchoContext(ctx)        // *echo.Context
    token := ec.Request().Header.Get("Authorization")

    // Or pull a value your auth middleware stashed on the context:
    user := echographql.GetUserFromContext(ctx, "user")
    _ = user
    _ = token
    // ...
}
```

## Scaffolding

`echonext generate graphql` emits gqlgen boilerplate (schema, resolvers,
generated package wiring). Define your `.graphql` schema, run gqlgen to
regenerate, then mount with `app.GraphQL` as above.
