# GraphQL Integration

EchoNext provides seamless integration with [gqlgen](https://gqlgen.com/) for GraphQL support.

## Quick Start

### 1. Generate GraphQL Boilerplate

```bash
echonext generate graphql
```

This creates:

```
graph/
├── schema.graphqls    # GraphQL schema
├── resolver.go        # Resolver struct
gqlgen.yml             # gqlgen configuration
tools/tools.go         # Tool dependencies
```

### 2. Define Your Schema

Edit `graph/schema.graphqls`:

```graphql
type Query {
    users: [User!]!
    user(id: ID!): User
}

type Mutation {
    createUser(input: CreateUserInput!): User!
    updateUser(id: ID!, input: UpdateUserInput!): User!
    deleteUser(id: ID!): Boolean!
}

type Subscription {
    userCreated: User!
}

type User {
    id: ID!
    name: String!
    email: String!
    createdAt: Time!
}

input CreateUserInput {
    name: String!
    email: String!
}

input UpdateUserInput {
    name: String
    email: String
}

scalar Time
```

### 3. Generate Code

```bash
go generate ./...
```

### 4. Implement Resolvers

Edit `graph/schema.resolvers.go`:

```go
func (r *queryResolver) Users(ctx context.Context) ([]*model.User, error) {
    return r.userService.GetAll(ctx)
}

func (r *queryResolver) User(ctx context.Context, id string) (*model.User, error) {
    return r.userService.GetByID(ctx, id)
}

func (r *mutationResolver) CreateUser(ctx context.Context, input model.CreateUserInput) (*model.User, error) {
    return r.userService.Create(ctx, input)
}
```

### 5. Integrate with EchoNext

```go
import (
    "github.com/abdussamadbello/echonext"
    "github.com/abdussamadbello/echonext/graphql"
    "myapp/graph"
)

func main() {
    app := echonext.New()

    app.GraphQL(graphql.Config{
        Path:           "/graphql",
        PlaygroundPath: "/playground",
        Schema: graph.NewExecutableSchema(graph.Config{
            Resolvers: graph.NewResolver(),
        }),
    })

    app.Start(":8080")
}
```

## Configuration Options

```go
graphql.Config{
    // GraphQL endpoint path
    Path: "/graphql",

    // GraphQL Playground path (empty to disable)
    PlaygroundPath: "/playground",

    // Generated executable schema
    Schema: schema,

    // Query complexity limit (0 = unlimited)
    ComplexityLimit: 100,

    // Query cache size
    QueryCacheSize: 1000,

    // Enable introspection (disable in production)
    EnableIntrospection: true,

    // Enable Apollo tracing
    EnableTracing: false,
}
```

## Accessing Echo Context

Access the Echo context in your resolvers:

```go
import "github.com/abdussamadbello/echonext/graphql"

func (r *queryResolver) CurrentUser(ctx context.Context) (*model.User, error) {
    // Get Echo context
    echoCtx := graphql.GetEchoContext(ctx)
    if echoCtx == nil {
        return nil, errors.New("no context available")
    }

    // Access request data
    userID := echoCtx.Get("user_id").(string)
    requestID := echoCtx.Request().Header.Get("X-Request-ID")

    return r.userService.GetByID(ctx, userID)
}
```

## Authentication

Use Echo middleware for authentication:

```go
import (
    "github.com/labstack/echo/v5/middleware"
)

func main() {
    app := echonext.New()

    // Add JWT middleware
    app.Use(middleware.JWTWithConfig(middleware.JWTConfig{
        SigningKey: []byte("secret"),
        Skipper: func(c *echo.Context) bool {
            // Skip auth for playground
            return c.Path() == "/playground"
        },
    }))

    app.GraphQL(graphql.Config{
        Path:           "/graphql",
        PlaygroundPath: "/playground",
        Schema:         schema,
    })
}
```

Access auth info in resolvers:

```go
func (r *queryResolver) Me(ctx context.Context) (*model.User, error) {
    echoCtx := graphql.GetEchoContext(ctx)

    // Get user from JWT token (set by middleware)
    user := echoCtx.Get("user").(*jwt.Token)
    claims := user.Claims.(jwt.MapClaims)
    userID := claims["user_id"].(string)

    return r.userService.GetByID(ctx, userID)
}
```

## Subscriptions

Implement real-time subscriptions:

### Schema

```graphql
type Subscription {
    userCreated: User!
    messageReceived(roomId: ID!): Message!
}
```

### Resolver

```go
type Resolver struct {
    userCreated chan *model.User
}

func NewResolver() *Resolver {
    return &Resolver{
        userCreated: make(chan *model.User, 100),
    }
}

func (r *subscriptionResolver) UserCreated(ctx context.Context) (<-chan *model.User, error) {
    return r.userCreated, nil
}

// Call this when a user is created
func (r *Resolver) NotifyUserCreated(user *model.User) {
    select {
    case r.userCreated <- user:
    default:
        // Channel full, skip notification
    }
}
```

### Mutation with Notification

```go
func (r *mutationResolver) CreateUser(ctx context.Context, input model.CreateUserInput) (*model.User, error) {
    user, err := r.userService.Create(ctx, input)
    if err != nil {
        return nil, err
    }

    // Notify subscribers
    r.NotifyUserCreated(user)

    return user, nil
}
```

## Error Handling

Return GraphQL errors:

```go
import "github.com/vektah/gqlparser/v2/gqlerror"

func (r *queryResolver) User(ctx context.Context, id string) (*model.User, error) {
    user, err := r.userService.GetByID(ctx, id)
    if err != nil {
        if errors.Is(err, ErrNotFound) {
            return nil, &gqlerror.Error{
                Message: "User not found",
                Extensions: map[string]interface{}{
                    "code": "NOT_FOUND",
                    "id":   id,
                },
            }
        }
        return nil, err
    }
    return user, nil
}
```

## DataLoader Pattern

Prevent N+1 queries with dataloaders:

```go
// loaders/user.go
type UserLoader struct {
    userService *user.Service
}

func (l *UserLoader) BatchGetUsers(ctx context.Context, ids []string) []*dataloader.Result[*model.User] {
    users, err := l.userService.GetByIDs(ctx, ids)
    if err != nil {
        // Return error for all
        results := make([]*dataloader.Result[*model.User], len(ids))
        for i := range ids {
            results[i] = &dataloader.Result[*model.User]{Error: err}
        }
        return results
    }

    // Map users by ID
    userMap := make(map[string]*model.User)
    for _, u := range users {
        userMap[u.ID] = u
    }

    // Return in order
    results := make([]*dataloader.Result[*model.User], len(ids))
    for i, id := range ids {
        if u, ok := userMap[id]; ok {
            results[i] = &dataloader.Result[*model.User]{Data: u}
        } else {
            results[i] = &dataloader.Result[*model.User]{Error: ErrNotFound}
        }
    }
    return results
}
```

## Query Complexity

Limit query complexity to prevent abuse:

```go
app.GraphQL(graphql.Config{
    Path:            "/graphql",
    Schema:          schema,
    ComplexityLimit: 100, // Max complexity score
})
```

Add complexity to fields in schema:

```graphql
type Query {
    users(first: Int = 10): [User!]! @complexity(value: 10, multipliers: ["first"])
}
```

## Caching

Query caching is enabled by default:

```go
app.GraphQL(graphql.Config{
    Path:           "/graphql",
    Schema:         schema,
    QueryCacheSize: 1000, // Cache up to 1000 parsed queries
})
```

## Testing GraphQL

```go
func TestUserQuery(t *testing.T) {
    app := echonext.New()
    app.GraphQL(graphql.Config{
        Path:   "/graphql",
        Schema: testSchema,
    })

    query := `
        query {
            user(id: "1") {
                id
                name
            }
        }
    `

    req := httptest.NewRequest("POST", "/graphql", strings.NewReader(
        fmt.Sprintf(`{"query": %q}`, query),
    ))
    req.Header.Set("Content-Type", "application/json")
    rec := httptest.NewRecorder()

    app.ServeHTTP(rec, req)

    assert.Equal(t, 200, rec.Code)

    var response map[string]interface{}
    json.Unmarshal(rec.Body.Bytes(), &response)

    data := response["data"].(map[string]interface{})
    user := data["user"].(map[string]interface{})
    assert.Equal(t, "1", user["id"])
}
```

## Best Practices

1. **Use DataLoaders** - Prevent N+1 query problems
2. **Limit complexity** - Set complexity limits for production
3. **Disable introspection** - Disable in production for security
4. **Use persisted queries** - Cache queries for performance
5. **Implement proper auth** - Use Echo middleware for authentication
6. **Handle errors properly** - Return meaningful GraphQL errors
7. **Monitor performance** - Enable tracing in development

## Example Project

See [examples/graphql-demo/](../../examples/graphql-demo/) for a complete working example with:

- Full CRUD operations
- Query and Mutation examples
- Subscription support
- Health check endpoint
- GraphQL Playground
