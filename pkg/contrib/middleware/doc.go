// Package middleware provides optional Echo middleware helpers for EchoNext applications.
//
// This package COMPLEMENTS Echo's built-in middleware rather than replacing it.
// Echo already has excellent middleware for logging, recovery, CORS, etc.
// These helpers add additional functionality that works alongside Echo's middleware.
//
// Features:
//   - RequestID: Add correlation IDs to requests
//   - Metrics: Simple request metrics collection
//   - StructuredLogger: Enhanced logging with structured fields
//
// Example usage:
//
//	import (
//	    "github.com/abdussamadbello/echonext"
//	    "github.com/abdussamadbello/echonext/pkg/contrib/middleware"
//	    echomw "github.com/labstack/echo/v4/middleware"
//	)
//
//	app := echonext.New()
//
//	// Use Echo's built-in middleware
//	app.Use(echomw.Recover())
//	app.Use(echomw.CORS())
//
//	// Add contrib middleware
//	app.Use(middleware.RequestID())
//
//	// Create metrics collector
//	metrics := middleware.NewMetrics()
//	app.Use(middleware.MetricsMiddleware(metrics))
//
//	// Expose metrics endpoint
//	app.GET("/metrics", middleware.MetricsHandler(metrics))
//
//	// Use structured logging with request IDs
//	app.Use(middleware.StructuredLogger(middleware.StructuredLoggerConfig{
//	    CustomFields: func(c echo.Context) map[string]interface{} {
//	        return map[string]interface{}{
//	            "user_agent": c.Request().UserAgent(),
//	        }
//	    },
//	}))
package middleware
