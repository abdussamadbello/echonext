package middleware

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

// StructuredLoggerConfig extends Echo's logger with structured fields
type StructuredLoggerConfig struct {
	// Skipper defines a function to skip middleware
	Skipper middleware.Skipper

	// CustomFields allows adding custom fields to log entries
	CustomFields func(c *echo.Context) map[string]interface{}
}

// responseStats extracts status code and bytes written from the Echo response.
// Echo's Context.Response() returns http.ResponseWriter; the underlying type is
// *echo.Response, which tracks Status and Size.
func responseStats(w http.ResponseWriter) (status int, size int64) {
	if r, ok := w.(*echo.Response); ok {
		return r.Status, r.Size
	}
	return 0, 0
}

// StructuredLogger returns a middleware that adds structured logging
// This complements Echo's built-in logger middleware
func StructuredLogger(config StructuredLoggerConfig) echo.MiddlewareFunc {
	if config.Skipper == nil {
		config.Skipper = middleware.DefaultSkipper
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			start := time.Now()
			err := next(c)
			stop := time.Now()

			req := c.Request()
			status, size := responseStats(c.Response())

			// Build log fields
			fields := map[string]interface{}{
				"time":       stop.Format(time.RFC3339),
				"method":     req.Method,
				"uri":        req.RequestURI,
				"status":     status,
				"latency_ms": stop.Sub(start).Milliseconds(),
				"bytes_in":   req.ContentLength,
				"bytes_out":  size,
			}

			// Add request ID if available
			if rid := GetRequestID(c); rid != "" {
				fields["request_id"] = rid
			}

			// Add custom fields
			if config.CustomFields != nil {
				for k, v := range config.CustomFields(c) {
					fields[k] = v
				}
			}

			// Convert map to slog attrs
			attrs := make([]any, 0, len(fields)*2)
			for k, v := range fields {
				attrs = append(attrs, k, v)
			}

			logger := c.Logger()
			switch {
			case status >= 500:
				logger.Error("request", attrs...)
			case status >= 400:
				logger.Warn("request", attrs...)
			default:
				logger.Info("request", attrs...)
			}

			return err
		}
	}
}
