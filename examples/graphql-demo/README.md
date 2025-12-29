# GraphQL Demo with EchoNext

This example demonstrates how to integrate GraphQL with EchoNext using the `graphql` subpackage.

## Quick Start

### 1. Install gqlgen

```bash
go install github.com/99designs/gqlgen@latest
```

### 2. Initialize gqlgen

```bash
cd examples/graphql-demo
gqlgen init
```

### 3. Define your schema

Edit `graph/schema.graphqls`:

```graphql
type User {
  id: ID!
  name: String!
  email: String!
  posts: [Post!]!
}

type Post {
  id: ID!
  title: String!
  content: String!
  author: User!
}

type Query {
  users: [User!]!
  user(id: ID!): User
  posts: [Post!]!
  post(id: ID!): Post
}

type Mutation {
  createUser(name: String!, email: String!): User!
  createPost(title: String!, content: String!, authorId: ID!): Post!
}

type Subscription {
  postCreated: Post!
}
```

### 4. Generate code

```bash
gqlgen generate
```

### 5. Implement resolvers

Edit `graph/resolver.go` and implement your resolvers.

### 6. Wire up with EchoNext

```go
package main

import (
    "log"

    "github.com/abdussamadbello/echonext"
    "github.com/abdussamadbello/echonext/graphql"
    "yourmodule/graph"
    "yourmodule/graph/generated"
)

func main() {
    app := echonext.New()

    // Create gqlgen schema
    schema := generated.NewExecutableSchema(generated.Config{
        Resolvers: &graph.Resolver{},
    })

    // Register GraphQL endpoint
    app.GraphQL(graphql.Config{
        Path:                "/graphql",
        PlaygroundPath:      "/playground",
        Schema:              schema,
        EnableIntrospection: true,
        ComplexityLimit:     100,
    }, graphql.Route{
        Summary:     "GraphQL API",
        Description: "Main GraphQL endpoint",
        Tags:        []string{"GraphQL"},
    })

    log.Fatal(app.Start(":8080"))
}
```

## Using the graphql Package

### Config Options

```go
graphql.Config{
    Path:                "/graphql",       // GraphQL endpoint path
    PlaygroundPath:      "/playground",    // Playground path (empty to disable)
    Schema:              schema,           // gqlgen ExecutableSchema
    ComplexityLimit:     100,              // Query complexity limit (0 = no limit)
    QueryCacheSize:      1000,             // LRU cache size for parsed queries
    EnableIntrospection: true,             // Enable schema introspection
    EnableTracing:       false,            // Enable Apollo tracing
    WebSocketUpgrader:   upgrader,         // Custom WebSocket upgrader
    RecoverFunc:         recoverFn,        // Custom panic recovery
    ErrorPresenter:      errorFn,          // Custom error formatting
}
```

### Accessing Echo Context in Resolvers

```go
import "github.com/abdussamadbello/echonext/graphql"

func (r *queryResolver) Me(ctx context.Context) (*User, error) {
    // Get Echo context
    c := graphql.GetEchoContext(ctx)
    if c == nil {
        return nil, errors.New("no context")
    }

    // Access headers
    token := c.Request().Header.Get("Authorization")

    // Access cookies
    cookie, _ := c.Cookie("session")

    // Get values set by middleware
    userID := graphql.GetUserFromContext(ctx, "user_id")

    return &User{ID: userID.(string)}, nil
}
```

### With Authentication Middleware

```go
import (
    "github.com/labstack/echo/v4/middleware"
)

// Add JWT middleware before GraphQL
app.Echo.Use(middleware.JWTWithConfig(middleware.JWTConfig{
    SigningKey: []byte("secret"),
    Skipper: func(c echo.Context) bool {
        // Skip auth for playground
        return c.Path() == "/playground"
    },
}))
```

### Group Support

```go
// GraphQL under /api/v1 prefix
api := app.Group("/api/v1")
api.GraphQL(graphql.Config{
    Path:           "/graphql",    // Will be /api/v1/graphql
    PlaygroundPath: "/playground", // Will be /api/v1/playground
    Schema:         schema,
})
```

## Directory Structure

After setup, your project should look like:

```
graphql-demo/
├── main.go              # Application entry point
├── graph/
│   ├── schema.graphqls  # GraphQL schema
│   ├── resolver.go      # Resolver implementations
│   └── generated/       # gqlgen generated code
└── gqlgen.yml           # gqlgen configuration
```

## Running the Example

```bash
go run main.go
```

Then open:
- GraphQL Playground: http://localhost:8080/playground
- GraphQL Endpoint: http://localhost:8080/graphql
- Swagger UI: http://localhost:8080/api/docs
