package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Request ID Middleware Tests
// =============================================================================

func TestRequestID(t *testing.T) {
	e := echo.New()
	e.Use(RequestID())

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	// Should have X-Request-ID header in response
	rid := rec.Header().Get(HeaderXRequestID)
	assert.NotEmpty(t, rid)
	assert.Len(t, rid, 36) // UUID v4 format
}

func TestRequestIDPreserveExisting(t *testing.T) {
	e := echo.New()
	e.Use(RequestID())

	existingID := "existing-request-id-12345"

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(HeaderXRequestID, existingID)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	// Should preserve the existing request ID
	rid := rec.Header().Get(HeaderXRequestID)
	assert.Equal(t, existingID, rid)
}

func TestRequestIDWithSkipper(t *testing.T) {
	e := echo.New()
	e.Use(RequestIDWithConfig(RequestIDConfig{
		Skipper: func(c echo.Context) bool {
			return c.Request().URL.Path == "/skip"
		},
	}))

	e.GET("/skip", func(c echo.Context) error {
		return c.String(http.StatusOK, "skipped")
	})

	e.GET("/noskip", func(c echo.Context) error {
		return c.String(http.StatusOK, "not skipped")
	})

	// Test skipped path
	req1 := httptest.NewRequest(http.MethodGet, "/skip", nil)
	rec1 := httptest.NewRecorder()
	e.ServeHTTP(rec1, req1)
	assert.Empty(t, rec1.Header().Get(HeaderXRequestID))

	// Test non-skipped path
	req2 := httptest.NewRequest(http.MethodGet, "/noskip", nil)
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)
	assert.NotEmpty(t, rec2.Header().Get(HeaderXRequestID))
}

func TestRequestIDCustomGenerator(t *testing.T) {
	e := echo.New()
	customID := "custom-generated-id"

	e.Use(RequestIDWithConfig(RequestIDConfig{
		Generator: func() string { return customID },
	}))

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	rid := rec.Header().Get(HeaderXRequestID)
	assert.Equal(t, customID, rid)
}

func TestRequestIDCustomHeader(t *testing.T) {
	e := echo.New()
	customHeader := "X-Custom-Request-ID"

	e.Use(RequestIDWithConfig(RequestIDConfig{
		RequestIDHeader: customHeader,
	}))

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	// Should use custom header
	rid := rec.Header().Get(customHeader)
	assert.NotEmpty(t, rid)
	// Default header should be empty
	assert.Empty(t, rec.Header().Get(HeaderXRequestID))
}

func TestGetRequestID(t *testing.T) {
	e := echo.New()

	var retrievedID string
	e.Use(RequestID())
	e.GET("/test", func(c echo.Context) error {
		retrievedID = GetRequestID(c)
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	// Retrieved ID should match response header
	assert.Equal(t, rec.Header().Get(HeaderXRequestID), retrievedID)
}

func TestGetRequestIDEmpty(t *testing.T) {
	e := echo.New()

	var retrievedID string
	// Don't use RequestID middleware
	e.GET("/test", func(c echo.Context) error {
		retrievedID = GetRequestID(c)
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	// Should return empty string when no request ID is set
	assert.Empty(t, retrievedID)
}

func TestDefaultRequestIDConfig(t *testing.T) {
	config := DefaultRequestIDConfig()

	assert.Nil(t, config.Skipper)
	assert.NotNil(t, config.Generator)
	assert.Equal(t, HeaderXRequestID, config.RequestIDHeader)
	assert.Equal(t, ContextKeyRequestID, config.ContextKey)

	// Generator should produce valid UUIDs
	id := config.Generator()
	assert.Len(t, id, 36)
}

// =============================================================================
// Metrics Middleware Tests
// =============================================================================

func TestNewMetrics(t *testing.T) {
	m := NewMetrics()

	assert.NotNil(t, m)
	assert.NotNil(t, m.RequestsByCode)
	assert.Equal(t, int64(0), m.TotalRequests)
	assert.Equal(t, int64(0), m.TotalErrors)
}

func TestMetricsRecord(t *testing.T) {
	m := NewMetrics()

	m.Record(200, 100*time.Millisecond)
	m.Record(201, 50*time.Millisecond)
	m.Record(400, 20*time.Millisecond)
	m.Record(500, 200*time.Millisecond)

	assert.Equal(t, int64(4), m.TotalRequests)
	assert.Equal(t, int64(2), m.TotalErrors) // 400 and 500
	assert.Equal(t, int64(1), m.RequestsByCode[200])
	assert.Equal(t, int64(1), m.RequestsByCode[201])
	assert.Equal(t, int64(1), m.RequestsByCode[400])
	assert.Equal(t, int64(1), m.RequestsByCode[500])
}

func TestMetricsGetMetrics(t *testing.T) {
	m := NewMetrics()

	m.Record(200, 100*time.Millisecond)
	m.Record(200, 200*time.Millisecond)

	metrics := m.GetMetrics()

	assert.Equal(t, int64(2), metrics["total_requests"])
	assert.Equal(t, int64(0), metrics["total_errors"])
	assert.Equal(t, int64(150), metrics["avg_duration_ms"]) // (100+200)/2

	codeMap := metrics["requests_by_code"].(map[int]int64)
	assert.Equal(t, int64(2), codeMap[200])
}

func TestMetricsGetMetricsEmpty(t *testing.T) {
	m := NewMetrics()

	metrics := m.GetMetrics()

	assert.Equal(t, int64(0), metrics["total_requests"])
	assert.Equal(t, int64(0), metrics["total_errors"])
	assert.Equal(t, int64(0), metrics["avg_duration_ms"])
}

func TestMetricsThreadSafety(t *testing.T) {
	m := NewMetrics()

	var wg sync.WaitGroup
	numGoroutines := 100
	requestsPerGoroutine := 100

	// Record metrics from multiple goroutines
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < requestsPerGoroutine; j++ {
				m.Record(200, 10*time.Millisecond)
			}
		}()
	}

	wg.Wait()

	expected := int64(numGoroutines * requestsPerGoroutine)
	assert.Equal(t, expected, m.TotalRequests)
}

func TestMetricsMiddleware(t *testing.T) {
	e := echo.New()
	m := NewMetrics()

	e.Use(MetricsMiddleware(m))

	e.GET("/ok", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	e.GET("/error", func(c echo.Context) error {
		return c.String(http.StatusInternalServerError, "Error")
	})

	// Make successful request
	req1 := httptest.NewRequest(http.MethodGet, "/ok", nil)
	rec1 := httptest.NewRecorder()
	e.ServeHTTP(rec1, req1)

	// Make error request
	req2 := httptest.NewRequest(http.MethodGet, "/error", nil)
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)

	assert.Equal(t, int64(2), m.TotalRequests)
	assert.Equal(t, int64(1), m.TotalErrors)
	assert.Equal(t, int64(1), m.RequestsByCode[200])
	assert.Equal(t, int64(1), m.RequestsByCode[500])
}

func TestMetricsHandler(t *testing.T) {
	e := echo.New()
	m := NewMetrics()

	m.Record(200, 100*time.Millisecond)

	e.GET("/metrics", MetricsHandler(m))

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "total_requests")
}

// =============================================================================
// Structured Logger Middleware Tests
// =============================================================================

func TestStructuredLogger(t *testing.T) {
	e := echo.New()

	e.Use(StructuredLogger(StructuredLoggerConfig{}))

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	// Should not panic
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestStructuredLoggerWithSkipper(t *testing.T) {
	e := echo.New()

	skipped := false
	e.Use(StructuredLogger(StructuredLoggerConfig{
		Skipper: func(c echo.Context) bool {
			skipped = c.Request().URL.Path == "/skip"
			return skipped
		},
	}))

	e.GET("/skip", func(c echo.Context) error {
		return c.String(http.StatusOK, "skipped")
	})

	req := httptest.NewRequest(http.MethodGet, "/skip", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.True(t, skipped)
}

func TestStructuredLoggerWithCustomFields(t *testing.T) {
	e := echo.New()

	customFieldsCalled := false
	e.Use(StructuredLogger(StructuredLoggerConfig{
		CustomFields: func(c echo.Context) map[string]interface{} {
			customFieldsCalled = true
			return map[string]interface{}{
				"custom_field": "custom_value",
			}
		},
	}))

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.True(t, customFieldsCalled)
}

func TestStructuredLoggerWithRequestID(t *testing.T) {
	e := echo.New()

	// Use RequestID middleware first, then structured logger
	e.Use(RequestID())
	e.Use(StructuredLogger(StructuredLoggerConfig{}))

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	// Should not panic and should include request ID in logs
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}
