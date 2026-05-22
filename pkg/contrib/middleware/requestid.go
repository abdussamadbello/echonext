// Package middleware provides optional Echo middleware helpers for EchoNext applications.
// These complement Echo's built-in middleware rather than replacing them.
package middleware

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

const (
	// HeaderXRequestID is the header name for request ID
	HeaderXRequestID = "X-Request-ID"
	// ContextKeyRequestID is the context key for storing request ID
	ContextKeyRequestID = "request_id"
)

// RequestIDConfig defines configuration for RequestID middleware
type RequestIDConfig struct {
	// Skipper defines a function to skip middleware
	Skipper func(c *echo.Context) bool

	// Generator defines a function to generate request ID
	// Default: generates UUID v4
	Generator func() string

	// RequestIDHeader is the header name for request ID
	// Default: X-Request-ID
	RequestIDHeader string

	// ContextKey is the key used to store request ID in context
	// Default: request_id
	ContextKey string
}

// DefaultRequestIDConfig returns default configuration
func DefaultRequestIDConfig() RequestIDConfig {
	return RequestIDConfig{
		Skipper:         nil,
		Generator:       func() string { return uuid.New().String() },
		RequestIDHeader: HeaderXRequestID,
		ContextKey:      ContextKeyRequestID,
	}
}

// RequestID returns a middleware that adds a request ID to each request
// The request ID is added to the response header and stored in context
func RequestID() echo.MiddlewareFunc {
	return RequestIDWithConfig(DefaultRequestIDConfig())
}

// RequestIDWithConfig returns RequestID middleware with custom configuration
func RequestIDWithConfig(config RequestIDConfig) echo.MiddlewareFunc {
	// Set defaults
	if config.Generator == nil {
		config.Generator = DefaultRequestIDConfig().Generator
	}
	if config.RequestIDHeader == "" {
		config.RequestIDHeader = HeaderXRequestID
	}
	if config.ContextKey == "" {
		config.ContextKey = ContextKeyRequestID
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			// Skip if configured
			if config.Skipper != nil && config.Skipper(c) {
				return next(c)
			}

			req := c.Request()
			res := c.Response()

			// Check if request ID already exists in header
			rid := req.Header.Get(config.RequestIDHeader)
			if rid == "" {
				rid = config.Generator()
			}

			// Store in context
			c.Set(config.ContextKey, rid)

			// Add to response header
			res.Header().Set(config.RequestIDHeader, rid)

			return next(c)
		}
	}
}

// GetRequestID retrieves the request ID from context
func GetRequestID(c *echo.Context) string {
	if rid, ok := c.Get(ContextKeyRequestID).(string); ok {
		return rid
	}
	return ""
}
