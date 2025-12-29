// Package main demonstrates GraphQL integration with EchoNext
//
// This is a fully working GraphQL example using gqlgen with EchoNext.
// It provides CRUD operations for Users and Posts.
package main

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/abdussamadbello/echonext"
	echographql "github.com/abdussamadbello/echonext/graphql"
	"github.com/abdussamadbello/echonext/examples/graphql-demo/graph"
	"github.com/abdussamadbello/echonext/examples/graphql-demo/graph/generated"
)

func main() {
	app := echonext.New()

	// Set API information
	app.SetInfo("GraphQL Demo API", "1.0.0", "Demonstrates GraphQL integration with EchoNext")

	// Create resolver with sample data
	resolver := graph.NewResolver()

	// Create gqlgen schema
	schema := generated.NewExecutableSchema(generated.Config{
		Resolvers: resolver,
	})

	// Register GraphQL endpoint using EchoNext's GraphQL integration
	app.GraphQL(echographql.Config{
		Path:                "/graphql",
		PlaygroundPath:      "/playground",
		Schema:              schema,
		EnableIntrospection: true,
		QueryCacheSize:      1000,
	}, echographql.Route{
		Summary:     "GraphQL API",
		Description: "GraphQL API endpoint for queries and mutations",
		Tags:        []string{"GraphQL"},
	})

	// Add REST endpoints alongside GraphQL
	app.GET("/health", func(c echo.Context) (map[string]string, error) {
		return map[string]string{"status": "healthy"}, nil
	}, echonext.Route{
		Summary: "Health check",
		Tags:    []string{"System"},
	})

	// Serve OpenAPI spec
	app.ServeOpenAPISpec("/api/spec")
	app.ServeSwaggerUI("/api/docs", "/api/spec")

	log.Println("Server starting on :8080")
	log.Println("GraphQL Playground: http://localhost:8080/playground")
	log.Println("GraphQL Endpoint:   http://localhost:8080/graphql")
	log.Println("Swagger UI:         http://localhost:8080/api/docs")
	log.Println("")
	log.Println("Try these queries in the playground:")
	log.Println("  query { users { id name email } }")
	log.Println("  query { posts { id title author { name } } }")
	log.Println("  mutation { createUser(input: {name: \"Dan\", email: \"dan@example.com\"}) { id name } }")

	if err := app.Start(":8080"); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
