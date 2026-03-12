package middleware

import (
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

// StructuredLoggerConfig extends Echo's logger with structured fields
type StructuredLoggerConfig struct {
	// Skipper defines a function to skip middleware
	Skipper middleware.Skipper

	// CustomFields allows adding custom fields to log entries
	CustomFields func(c echo.Context) map[string]interface{}
}

// StructuredLogger returns a middleware that adds structured logging
// This complements Echo's built-in logger middleware
func StructuredLogger(config StructuredLoggerConfig) echo.MiddlewareFunc {
	if config.Skipper == nil {
		config.Skipper = middleware.DefaultSkipper
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			start := time.Now()
			err := next(c)
			stop := time.Now()

			req := c.Request()
			res := c.Response()

			// Build log fields
			fields := map[string]interface{}{
				"time":       stop.Format(time.RFC3339),
				"method":     req.Method,
				"uri":        req.RequestURI,
				"status":     res.Status,
				"latency_ms": stop.Sub(start).Milliseconds(),
				"bytes_in":   req.ContentLength,
				"bytes_out":  res.Size,
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

			// Log based on status code
			if res.Status >= 500 {
				c.Logger().Errorj(fields)
			} else if res.Status >= 400 {
				c.Logger().Warnj(fields)
			} else {
				c.Logger().Infoj(fields)
			}

			return err
		}
	}
}
