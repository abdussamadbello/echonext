package middleware

import (
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

// Metrics holds request metrics
type Metrics struct {
	mu             sync.RWMutex
	TotalRequests  int64
	TotalErrors    int64
	RequestsByCode map[int]int64
	TotalDuration  time.Duration
}

// NewMetrics creates a new metrics instance
func NewMetrics() *Metrics {
	return &Metrics{
		RequestsByCode: make(map[int]int64),
	}
}

// Record records a request metric
func (m *Metrics) Record(statusCode int, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalRequests++
	m.TotalDuration += duration
	m.RequestsByCode[statusCode]++

	if statusCode >= 400 {
		m.TotalErrors++
	}
}

// GetMetrics returns current metrics (thread-safe)
func (m *Metrics) GetMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	avgDuration := int64(0)
	if m.TotalRequests > 0 {
		avgDuration = m.TotalDuration.Milliseconds() / m.TotalRequests
	}

	codeMap := make(map[int]int64)
	for code, count := range m.RequestsByCode {
		codeMap[code] = count
	}

	return map[string]interface{}{
		"total_requests":   m.TotalRequests,
		"total_errors":     m.TotalErrors,
		"avg_duration_ms":  avgDuration,
		"requests_by_code": codeMap,
	}
}

// MetricsMiddleware returns middleware that collects request metrics
func MetricsMiddleware(metrics *Metrics) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			// Process request
			err := next(c)

			// Record metrics
			duration := time.Since(start)
			statusCode := c.Response().Status
			metrics.Record(statusCode, duration)

			return err
		}
	}
}

// MetricsHandler returns a handler that exposes metrics
func MetricsHandler(metrics *Metrics) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(200, metrics.GetMetrics())
	}
}
