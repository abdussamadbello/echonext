// Package graphql provides GraphQL integration with automatic OpenAPI documentation
package graphql

import (
	"context"
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v5"
	"github.com/vektah/gqlparser/v2/ast"
)

// GraphQL configuration defaults
const (
	DefaultPath              = "/graphql"
	DefaultPlaygroundPath    = "/playground"
	DefaultQueryCacheSize    = 1000
	DefaultComplexityLimit   = 0 // No limit
	DefaultIntrospection     = true
	DefaultSubscriptionsPath = "/graphql/subscriptions"
)

// Config configures GraphQL endpoint behavior
type Config struct {
	// Path is the endpoint path for GraphQL queries/mutations (default: /graphql)
	Path string
	// PlaygroundPath is the endpoint for GraphQL Playground (default: /playground, empty to disable)
	PlaygroundPath string
	// Schema is the gqlgen executable schema
	Schema graphql.ExecutableSchema
	// ComplexityLimit sets the maximum query complexity (0 = no limit)
	ComplexityLimit int
	// QueryCacheSize sets the LRU cache size for parsed queries
	QueryCacheSize int
	// EnableIntrospection enables schema introspection (default: true)
	EnableIntrospection bool
	// EnableTracing enables Apollo tracing
	EnableTracing bool
	// WebSocketUpgrader for subscriptions (nil = use default)
	WebSocketUpgrader *websocket.Upgrader
	// RecoverFunc handles panics in resolvers
	RecoverFunc graphql.RecoverFunc
	// ErrorPresenter customizes error responses
	ErrorPresenter graphql.ErrorPresenterFunc
}

// Route configures GraphQL endpoint metadata for OpenAPI documentation
type Route struct {
	Summary           string
	Description       string
	Tags              []string
	DocumentInOpenAPI bool // Whether to include in OpenAPI spec
}

// SubscriptionConfig configures WebSocket subscriptions
type SubscriptionConfig struct {
	// Path is the WebSocket endpoint path (default: /graphql/subscriptions)
	Path string
	// KeepAliveInterval is the interval for keep-alive pings
	KeepAliveInterval int // in seconds
}

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// EchoContextKey is the key used to store echo.Context in GraphQL context
	EchoContextKey contextKey = "echonext_echo_context"
)

// DefaultConfig returns sensible default configuration
func DefaultConfig(schema graphql.ExecutableSchema) Config {
	return Config{
		Path:                DefaultPath,
		PlaygroundPath:      DefaultPlaygroundPath,
		Schema:              schema,
		ComplexityLimit:     DefaultComplexityLimit,
		QueryCacheSize:      DefaultQueryCacheSize,
		EnableIntrospection: DefaultIntrospection,
		EnableTracing:       false,
		WebSocketUpgrader: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins by default
			},
		},
	}
}

// GetEchoContext retrieves the Echo context from a GraphQL resolver context
// Returns nil if not found
func GetEchoContext(ctx context.Context) *echo.Context {
	if c, ok := ctx.Value(EchoContextKey).(*echo.Context); ok {
		return c
	}
	return nil
}

// MustGetEchoContext retrieves the Echo context from a GraphQL resolver context
// Panics if not found
func MustGetEchoContext(ctx context.Context) *echo.Context {
	c := GetEchoContext(ctx)
	if c == nil {
		panic("echo.Context not found in GraphQL context")
	}
	return c
}

// GetUserFromContext is a helper to get user information from the Echo context
// Useful for authentication in resolvers
func GetUserFromContext(ctx context.Context, key string) interface{} {
	c := GetEchoContext(ctx)
	if c == nil {
		return nil
	}
	return c.Get(key)
}

// Middleware creates Echo middleware that can be used with GraphQL handlers
// This is useful for authentication, logging, etc.
type Middleware func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler

// WithMiddleware wraps a GraphQL handler with Echo middleware context
func WithMiddleware(middlewares ...echo.MiddlewareFunc) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		// Chain all middlewares
		handler := next
		for i := len(middlewares) - 1; i >= 0; i-- {
			handler = middlewares[i](handler)
		}
		return handler
	}
}

// Handler creates a gqlgen handler from a Config
// This is used internally by echonext.App.GraphQL but can be used directly
func Handler(config Config) *handler.Server {
	// Apply defaults
	if config.QueryCacheSize == 0 {
		config.QueryCacheSize = DefaultQueryCacheSize
	}
	if config.Schema == nil {
		panic("GraphQL schema is required")
	}

	// Create gqlgen handler
	srv := handler.New(config.Schema)

	// Configure transports
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{})

	// Add WebSocket transport for subscriptions
	if config.WebSocketUpgrader != nil {
		srv.AddTransport(&transport.Websocket{
			Implementation: gorillaWebsocketImplementation{
				upgrader: *config.WebSocketUpgrader,
			},
		})
	}

	// Configure query cache
	srv.SetQueryCache(lru.New[*ast.QueryDocument](config.QueryCacheSize))

	// Configure extensions
	if config.EnableIntrospection {
		srv.Use(extension.Introspection{})
	}
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	// Add complexity limit
	if config.ComplexityLimit > 0 {
		srv.Use(extension.FixedComplexityLimit(config.ComplexityLimit))
	}

	// Set error presenter
	if config.ErrorPresenter != nil {
		srv.SetErrorPresenter(config.ErrorPresenter)
	}

	// Set recover function
	if config.RecoverFunc != nil {
		srv.SetRecoverFunc(config.RecoverFunc)
	}

	return srv
}

// SubscriptionHandler creates a gqlgen handler for subscriptions
func SubscriptionHandler(schema graphql.ExecutableSchema) *handler.Server {
	srv := handler.New(schema)
	srv.AddTransport(&transport.Websocket{
		Implementation: gorillaWebsocketImplementation{
			upgrader: websocket.Upgrader{
				ReadBufferSize:  1024,
				WriteBufferSize: 1024,
				CheckOrigin: func(r *http.Request) bool {
					return true
				},
			},
		},
	})
	return srv
}

// PlaygroundHandler creates a playground handler for the given GraphQL path
func PlaygroundHandler(title, graphqlPath string) http.Handler {
	return playground.Handler(title, graphqlPath)
}

// WrapWithEchoContext creates an Echo handler that injects context into GraphQL requests
func WrapWithEchoContext(srv *handler.Server) echo.HandlerFunc {
	return func(c *echo.Context) error {
		// Add Echo context to GraphQL context
		ctx := context.WithValue(c.Request().Context(), EchoContextKey, c)
		c.SetRequest(c.Request().WithContext(ctx))
		srv.ServeHTTP(c.Response(), c.Request())
		return nil
	}
}
