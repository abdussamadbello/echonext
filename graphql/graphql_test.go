package graphql

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConstants(t *testing.T) {
	// Test the constants and configuration behavior
	assert.Equal(t, "/graphql", DefaultPath)
	assert.Equal(t, "/playground", DefaultPlaygroundPath)
	assert.Equal(t, 1000, DefaultQueryCacheSize)
	assert.Equal(t, 0, DefaultComplexityLimit)
	assert.True(t, DefaultIntrospection)
	assert.Equal(t, "/graphql/subscriptions", DefaultSubscriptionsPath)
}

func TestConfig_Defaults(t *testing.T) {
	config := Config{}

	// Test that defaults can be applied
	if config.Path == "" {
		config.Path = DefaultPath
	}
	if config.QueryCacheSize == 0 {
		config.QueryCacheSize = DefaultQueryCacheSize
	}

	assert.Equal(t, "/graphql", config.Path)
	assert.Equal(t, 1000, config.QueryCacheSize)
}

func TestRoute(t *testing.T) {
	route := Route{
		Summary:           "GraphQL API",
		Description:       "Main GraphQL endpoint",
		Tags:              []string{"graphql", "api"},
		DocumentInOpenAPI: true,
	}

	assert.Equal(t, "GraphQL API", route.Summary)
	assert.Equal(t, "Main GraphQL endpoint", route.Description)
	assert.Len(t, route.Tags, 2)
	assert.True(t, route.DocumentInOpenAPI)
}

func TestSubscriptionConfig(t *testing.T) {
	config := SubscriptionConfig{
		Path:              "/ws/graphql",
		KeepAliveInterval: 30,
	}

	assert.Equal(t, "/ws/graphql", config.Path)
	assert.Equal(t, 30, config.KeepAliveInterval)
}

func TestGetEchoContext_NotFound(t *testing.T) {
	ctx := context.Background()
	c := GetEchoContext(ctx)
	assert.Nil(t, c)
}

func TestMustGetEchoContext_Panics(t *testing.T) {
	ctx := context.Background()
	assert.Panics(t, func() {
		MustGetEchoContext(ctx)
	})
}

func TestGetUserFromContext_NoContext(t *testing.T) {
	ctx := context.Background()
	result := GetUserFromContext(ctx, "user")
	assert.Nil(t, result)
}

func TestEchoContextKey(t *testing.T) {
	// Test that the context key is unique
	assert.Equal(t, contextKey("echonext_echo_context"), EchoContextKey)
}

func TestConfig_WithWebSocketUpgrader(t *testing.T) {
	config := Config{
		WebSocketUpgrader: nil,
	}

	// Without upgrader, no WebSocket transport
	assert.Nil(t, config.WebSocketUpgrader)
}

func TestConfig_ComplexityLimit(t *testing.T) {
	config := Config{
		ComplexityLimit: 100,
	}

	assert.Equal(t, 100, config.ComplexityLimit)
}

func TestConfig_Tracing(t *testing.T) {
	config := Config{
		EnableTracing: true,
	}

	assert.True(t, config.EnableTracing)
}

func TestConfig_Introspection(t *testing.T) {
	config := Config{
		EnableIntrospection: true,
	}

	assert.True(t, config.EnableIntrospection)
}

func TestConfig_ErrorPresenter(t *testing.T) {
	config := Config{
		ErrorPresenter: nil,
	}

	assert.Nil(t, config.ErrorPresenter)
}

func TestConfig_RecoverFunc(t *testing.T) {
	config := Config{
		RecoverFunc: nil,
	}

	assert.Nil(t, config.RecoverFunc)
}

func TestHandler_PanicsWithoutSchema(t *testing.T) {
	config := Config{
		Schema: nil,
	}

	assert.Panics(t, func() {
		Handler(config)
	})
}
