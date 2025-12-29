// Package main demonstrates GraphQL integration with EchoNext
//
// This example shows how to set up a GraphQL API using the graphql subpackage.
// For a full GraphQL implementation, you would use gqlgen to generate the schema
// and resolvers.
//
// Setup steps:
// 1. Install gqlgen: go install github.com/99designs/gqlgen@latest
// 2. Initialize schema: gqlgen init
// 3. Define your schema in schema.graphqls
// 4. Generate code: gqlgen generate
// 5. Implement resolvers
// 6. Wire up with EchoNext as shown below
package main

import (
	"context"
	"log"
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/labstack/echo/v4"
	"github.com/vektah/gqlparser/v2/ast"

	"github.com/abdussamadbello/echonext"
	echographql "github.com/abdussamadbello/echonext/graphql"
)

// User represents a user in our system
type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Post represents a blog post
type Post struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	AuthorID string `json:"authorId"`
}

// In-memory data store for demo
var users = []User{
	{ID: "1", Name: "Alice", Email: "alice@example.com"},
	{ID: "2", Name: "Bob", Email: "bob@example.com"},
}

var posts = []Post{
	{ID: "1", Title: "Hello GraphQL", Content: "Getting started with GraphQL", AuthorID: "1"},
	{ID: "2", Title: "EchoNext Tips", Content: "Building APIs with EchoNext", AuthorID: "2"},
}

// SimpleSchema implements a minimal ExecutableSchema for demonstration
// In a real application, you would use gqlgen to generate this
type SimpleSchema struct{}

func (s *SimpleSchema) Schema() *ast.Schema {
	return nil // In real use, this returns the parsed schema
}

func (s *SimpleSchema) Complexity(typeName, fieldName string, childComplexity int, args map[string]interface{}) (int, bool) {
	return childComplexity + 1, true
}

func (s *SimpleSchema) Exec(ctx context.Context) graphql.ResponseHandler {
	return func(ctx context.Context) *graphql.Response {
		// This is a simplified handler - real implementation uses gqlgen resolvers
		return &graphql.Response{
			Data: []byte(`{"message": "Use gqlgen for full GraphQL implementation"}`),
		}
	}
}

func main() {
	app := echonext.New()

	// Set API information
	app.SetInfo("GraphQL Demo API", "1.0.0", "Demonstrates GraphQL integration with EchoNext")

	// Option 1: Using echonext.GraphQL() with graphql.Config
	// This is the recommended approach for most use cases
	//
	// schema := generated.NewExecutableSchema(generated.Config{
	//     Resolvers: &resolver.Resolver{},
	// })
	//
	// app.GraphQL(echographql.Config{
	//     Path:                "/graphql",
	//     PlaygroundPath:      "/playground",
	//     Schema:              schema,
	//     EnableIntrospection: true,
	//     ComplexityLimit:     100,
	// }, echographql.Route{
	//     Summary:     "GraphQL API",
	//     Description: "Main GraphQL endpoint for queries and mutations",
	//     Tags:        []string{"GraphQL"},
	// })

	// Option 2: Manual setup with more control
	// Use this when you need custom middleware or special configuration
	setupManualGraphQL(app)

	// Add some REST endpoints alongside GraphQL
	app.GET("/health", func(c echo.Context) (map[string]string, error) {
		return map[string]string{"status": "healthy"}, nil
	}, echonext.Route{
		Summary: "Health check",
		Tags:    []string{"System"},
	})

	// REST API for users (can coexist with GraphQL)
	app.GET("/api/users", func(c echo.Context) ([]User, error) {
		return users, nil
	}, echonext.Route{
		Summary: "List all users",
		Tags:    []string{"Users"},
	})

	// Serve OpenAPI spec
	app.ServeOpenAPISpec("/api/spec")
	app.ServeSwaggerUI("/api/docs", "/api/spec")

	log.Println("Server starting on :8080")
	log.Println("GraphQL Playground: http://localhost:8080/playground")
	log.Println("GraphQL Endpoint:   http://localhost:8080/graphql")
	log.Println("Swagger UI:         http://localhost:8080/api/docs")

	if err := app.Start(":8080"); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

// setupManualGraphQL demonstrates manual GraphQL setup with full control
func setupManualGraphQL(app *echonext.App) {
	// In a real application, you would use your gqlgen-generated schema:
	// schema := generated.NewExecutableSchema(generated.Config{
	//     Resolvers: &resolver.Resolver{},
	// })

	// For this demo, we'll create a simple handler
	srv := handler.NewDefaultServer(nil) // Replace nil with your schema

	// Configure transports
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{})

	// Add query cache
	srv.SetQueryCache(lru.New(1000))

	// Enable introspection
	srv.Use(extension.Introspection{})

	// Add automatic persisted queries
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New(100),
	})

	// Custom Echo handler that injects context
	graphqlHandler := func(c echo.Context) error {
		// Add Echo context to GraphQL context for use in resolvers
		ctx := context.WithValue(c.Request().Context(), echographql.EchoContextKey, c)
		c.SetRequest(c.Request().WithContext(ctx))
		srv.ServeHTTP(c.Response(), c.Request())
		return nil
	}

	// Register endpoints
	app.Echo.POST("/graphql", graphqlHandler)
	app.Echo.GET("/graphql", graphqlHandler)

	// Playground
	playgroundHandler := playground.Handler("GraphQL Playground", "/graphql")
	app.Echo.GET("/playground", echo.WrapHandler(playgroundHandler))
}

// Example resolver showing how to access Echo context in GraphQL resolvers
//
// func (r *queryResolver) CurrentUser(ctx context.Context) (*User, error) {
//     // Get Echo context from GraphQL context
//     c := echographql.GetEchoContext(ctx)
//     if c == nil {
//         return nil, errors.New("no echo context")
//     }
//
//     // Access request headers, cookies, etc.
//     authHeader := c.Request().Header.Get("Authorization")
//
//     // Get user from context (set by auth middleware)
//     userID := echographql.GetUserFromContext(ctx, "user_id")
//
//     return &User{ID: userID.(string)}, nil
// }
