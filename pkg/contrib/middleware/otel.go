package middleware

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/labstack/echo/v5"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// OTELConfig configures OpenTelemetry instrumentation
type OTELConfig struct {
	// ServiceName is the name of the service (required)
	ServiceName string

	// ServiceVersion is the version of the service
	ServiceVersion string

	// Environment is the deployment environment (e.g., "production", "staging")
	Environment string

	// Endpoint is the OTLP collector endpoint (e.g., "localhost:4317")
	// If empty, uses OTEL_EXPORTER_OTLP_ENDPOINT env var
	Endpoint string

	// Insecure disables TLS for the OTLP connection
	Insecure bool

	// SampleRate is the trace sampling rate (0.0 to 1.0)
	// Default is 1.0 (sample everything)
	SampleRate float64

	// EnableTracing enables distributed tracing
	// Default is true
	EnableTracing bool

	// EnableMetrics enables metrics collection
	// Default is true
	EnableMetrics bool

	// Skipper defines a function to skip OTEL for certain requests
	Skipper func(echo.Context) bool

	// CustomAttributes adds custom attributes to all spans
	CustomAttributes []attribute.KeyValue

	// PropagateHeaders specifies additional headers to propagate
	PropagateHeaders []string
}

// OTELShutdown holds shutdown functions for OTEL providers
type OTELShutdown struct {
	tracerShutdown func(context.Context) error
	meterShutdown  func(context.Context) error
}

// Shutdown gracefully shuts down OTEL providers
func (s *OTELShutdown) Shutdown(ctx context.Context) error {
	var firstErr error
	if s.tracerShutdown != nil {
		if err := s.tracerShutdown(ctx); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if s.meterShutdown != nil {
		if err := s.meterShutdown(ctx); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// DefaultOTELConfig returns a default OTEL configuration
// It reads from environment variables when available
func DefaultOTELConfig() OTELConfig {
	cfg := OTELConfig{
		ServiceName:    getEnvOrDefault("OTEL_SERVICE_NAME", "echonext-service"),
		ServiceVersion: getEnvOrDefault("OTEL_SERVICE_VERSION", "1.0.0"),
		Environment:    getEnvOrDefault("OTEL_ENVIRONMENT", "development"),
		Endpoint:       getEnvOrDefault("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
		Insecure:       getEnvBool("OTEL_INSECURE", true),
		SampleRate:     getEnvFloat("OTEL_SAMPLE_RATE", 1.0),
		EnableTracing:  getEnvBool("OTEL_ENABLE_TRACING", true),
		EnableMetrics:  getEnvBool("OTEL_ENABLE_METRICS", true),
	}
	return cfg
}

// InitOTEL initializes OpenTelemetry with the given configuration
// Returns a shutdown function that should be called on application exit
func InitOTEL(ctx context.Context, cfg OTELConfig) (*OTELShutdown, error) {
	shutdown := &OTELShutdown{}

	// Apply defaults
	if cfg.SampleRate <= 0 {
		cfg.SampleRate = 1.0
	}
	if cfg.Endpoint == "" {
		cfg.Endpoint = getEnvOrDefault("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317")
	}

	// Create resource with service information
	res, err := sdkresource.New(ctx,
		sdkresource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			semconv.DeploymentEnvironment(cfg.Environment),
		),
		sdkresource.WithHost(),
		sdkresource.WithProcess(),
		sdkresource.WithTelemetrySDK(),
	)
	if err != nil {
		return nil, err
	}

	// Setup gRPC connection options
	var dialOpts []grpc.DialOption
	if cfg.Insecure {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Initialize tracing
	if cfg.EnableTracing {
		traceExporter, err := otlptracegrpc.New(ctx,
			otlptracegrpc.WithEndpoint(cfg.Endpoint),
			otlptracegrpc.WithDialOption(dialOpts...),
		)
		if err != nil {
			return nil, err
		}

		// Create sampler based on sample rate
		var sampler sdktrace.Sampler
		if cfg.SampleRate >= 1.0 {
			sampler = sdktrace.AlwaysSample()
		} else if cfg.SampleRate <= 0 {
			sampler = sdktrace.NeverSample()
		} else {
			sampler = sdktrace.TraceIDRatioBased(cfg.SampleRate)
		}

		tp := sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(traceExporter),
			sdktrace.WithResource(res),
			sdktrace.WithSampler(sampler),
		)

		otel.SetTracerProvider(tp)
		shutdown.tracerShutdown = tp.Shutdown
	}

	// Initialize metrics
	if cfg.EnableMetrics {
		metricExporter, err := otlpmetricgrpc.New(ctx,
			otlpmetricgrpc.WithEndpoint(cfg.Endpoint),
			otlpmetricgrpc.WithDialOption(dialOpts...),
		)
		if err != nil {
			return nil, err
		}

		mp := sdkmetric.NewMeterProvider(
			sdkmetric.WithResource(res),
			sdkmetric.WithReader(
				sdkmetric.NewPeriodicReader(metricExporter,
					sdkmetric.WithInterval(15*time.Second),
				),
			),
		)

		otel.SetMeterProvider(mp)
		shutdown.meterShutdown = mp.Shutdown
	}

	// Set up propagators for distributed tracing
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return shutdown, nil
}

// OTELMiddleware returns Echo middleware that instruments HTTP requests with OpenTelemetry
// This is the primary middleware for automatic OTEL instrumentation
func OTELMiddleware(serviceName string, opts ...OTELMiddlewareOption) echo.MiddlewareFunc {
	cfg := otelMiddlewareConfig{
		serviceName: serviceName,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	// Use otelecho middleware as the base
	baseMiddleware := otelecho.Middleware(serviceName)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		// Wrap with otelecho first
		handler := baseMiddleware(next)

		return func(c echo.Context) error {
			// Skip if configured
			if cfg.skipper != nil && cfg.skipper(c) {
				return next(c)
			}

			// Add custom attributes to span if configured
			if len(cfg.customAttributes) > 0 {
				span := trace.SpanFromContext(c.Request().Context())
				span.SetAttributes(cfg.customAttributes...)
			}

			// Add request ID to span if available
			if rid := GetRequestID(c); rid != "" {
				span := trace.SpanFromContext(c.Request().Context())
				span.SetAttributes(attribute.String("request.id", rid))
			}

			return handler(c)
		}
	}
}

// otelMiddlewareConfig holds middleware configuration
type otelMiddlewareConfig struct {
	serviceName      string
	skipper          func(echo.Context) bool
	customAttributes []attribute.KeyValue
}

// OTELMiddlewareOption configures the OTEL middleware
type OTELMiddlewareOption func(*otelMiddlewareConfig)

// WithSkipper sets a skipper function for the middleware
func WithSkipper(skipper func(echo.Context) bool) OTELMiddlewareOption {
	return func(cfg *otelMiddlewareConfig) {
		cfg.skipper = skipper
	}
}

// WithCustomAttributes adds custom attributes to all spans
func WithCustomAttributes(attrs ...attribute.KeyValue) OTELMiddlewareOption {
	return func(cfg *otelMiddlewareConfig) {
		cfg.customAttributes = append(cfg.customAttributes, attrs...)
	}
}

// OTELTracingMiddleware returns a simpler tracing-only middleware
// Use this if you only need distributed tracing without metrics
func OTELTracingMiddleware(serviceName string) echo.MiddlewareFunc {
	return otelecho.Middleware(serviceName)
}

// GetTraceID extracts the trace ID from the current request context
func GetTraceID(c echo.Context) string {
	span := trace.SpanFromContext(c.Request().Context())
	if span.SpanContext().HasTraceID() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

// GetSpanID extracts the span ID from the current request context
func GetSpanID(c echo.Context) string {
	span := trace.SpanFromContext(c.Request().Context())
	if span.SpanContext().HasSpanID() {
		return span.SpanContext().SpanID().String()
	}
	return ""
}

// AddSpanEvent adds an event to the current span
func AddSpanEvent(c echo.Context, name string, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(c.Request().Context())
	span.AddEvent(name, trace.WithAttributes(attrs...))
}

// SetSpanAttributes adds attributes to the current span
func SetSpanAttributes(c echo.Context, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(c.Request().Context())
	span.SetAttributes(attrs...)
}

// SetSpanStatus sets the status of the current span
func SetSpanStatus(c echo.Context, code trace.SpanKind, description string) {
	// Note: This is intentionally a no-op as SpanKind is set at span creation
	// Use RecordError for error status
}

// RecordError records an error in the current span
func RecordError(c echo.Context, err error, opts ...trace.EventOption) {
	span := trace.SpanFromContext(c.Request().Context())
	span.RecordError(err, opts...)
}

// StartSpan starts a new child span from the request context
// Remember to call span.End() when done
func StartSpan(c echo.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	tracer := otel.Tracer("")
	return tracer.Start(c.Request().Context(), name, opts...)
}

// Helper functions for environment variable parsing

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		}
	}
	return defaultValue
}

// =============================================================================
// Instrumented HTTP Client for Outgoing Requests
// =============================================================================

// TracedHTTPClient is an HTTP client with OpenTelemetry instrumentation
// for tracing outgoing requests
type TracedHTTPClient struct {
	*http.Client
}

// NewTracedHTTPClient creates a new HTTP client with OTEL instrumentation
// All outgoing requests will automatically create spans and propagate trace context
func NewTracedHTTPClient(opts ...TracedClientOption) *TracedHTTPClient {
	cfg := &tracedClientConfig{
		timeout:   30 * time.Second,
		transport: http.DefaultTransport,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	transport := otelhttp.NewTransport(cfg.transport,
		otelhttp.WithSpanNameFormatter(cfg.spanNameFormatter),
	)

	return &TracedHTTPClient{
		Client: &http.Client{
			Transport: transport,
			Timeout:   cfg.timeout,
		},
	}
}

// tracedClientConfig holds configuration for traced HTTP client
type tracedClientConfig struct {
	timeout           time.Duration
	transport         http.RoundTripper
	spanNameFormatter func(operation string, r *http.Request) string
}

// TracedClientOption configures the traced HTTP client
type TracedClientOption func(*tracedClientConfig)

// WithClientTimeout sets the client timeout
func WithClientTimeout(timeout time.Duration) TracedClientOption {
	return func(cfg *tracedClientConfig) {
		cfg.timeout = timeout
	}
}

// WithBaseTransport sets the base transport to wrap
func WithBaseTransport(transport http.RoundTripper) TracedClientOption {
	return func(cfg *tracedClientConfig) {
		cfg.transport = transport
	}
}

// WithSpanNameFormatter sets a custom span name formatter
func WithSpanNameFormatter(formatter func(operation string, r *http.Request) string) TracedClientOption {
	return func(cfg *tracedClientConfig) {
		cfg.spanNameFormatter = formatter
	}
}

// Do executes the request with tracing
// Use this when you have access to echo.Context to propagate the trace
func (c *TracedHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return c.Client.Do(req)
}

// DoWithContext executes the request with the given context
// The context should contain the trace information (from echo.Context.Request().Context())
func (c *TracedHTTPClient) DoWithContext(ctx context.Context, req *http.Request) (*http.Response, error) {
	return c.Client.Do(req.WithContext(ctx))
}

// Get performs a traced GET request
func (c *TracedHTTPClient) Get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

// Post performs a traced POST request
func (c *TracedHTTPClient) Post(ctx context.Context, url string, contentType string, body interface{}) (*http.Response, error) {
	var bodyReader *http.Request
	var err error

	switch b := body.(type) {
	case nil:
		bodyReader, err = http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	case *http.Request:
		bodyReader = b.WithContext(ctx)
	default:
		// For io.Reader types
		if reader, ok := body.(interface{ Read([]byte) (int, error) }); ok {
			bodyReader, err = http.NewRequestWithContext(ctx, http.MethodPost, url, readerWrapper{reader})
		} else {
			bodyReader, err = http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
		}
	}

	if err != nil {
		return nil, err
	}
	if contentType != "" {
		bodyReader.Header.Set("Content-Type", contentType)
	}
	return c.Client.Do(bodyReader)
}

// readerWrapper wraps a reader interface
type readerWrapper struct {
	r interface{ Read([]byte) (int, error) }
}

func (rw readerWrapper) Read(p []byte) (int, error) {
	return rw.r.Read(p)
}

// WrapTransport wraps an existing http.RoundTripper with OTEL instrumentation
// Use this to add tracing to an existing HTTP client
func WrapTransport(transport http.RoundTripper) http.RoundTripper {
	if transport == nil {
		transport = http.DefaultTransport
	}
	return otelhttp.NewTransport(transport)
}

// WrapHTTPClient wraps an existing HTTP client with OTEL instrumentation
func WrapHTTPClient(client *http.Client) *TracedHTTPClient {
	if client == nil {
		client = http.DefaultClient
	}

	transport := client.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	return &TracedHTTPClient{
		Client: &http.Client{
			Transport:     otelhttp.NewTransport(transport),
			Timeout:       client.Timeout,
			CheckRedirect: client.CheckRedirect,
			Jar:           client.Jar,
		},
	}
}

// HTTPClientFromContext creates a request with trace context from echo.Context
// This is a convenience function for making outgoing requests that continue the trace
func HTTPClientFromContext(c echo.Context) context.Context {
	return c.Request().Context()
}

// NewRequestWithTrace creates an http.Request with trace context from echo.Context
func NewRequestWithTrace(c echo.Context, method, url string, body interface{}) (*http.Request, error) {
	ctx := c.Request().Context()

	var req *http.Request
	var err error

	switch b := body.(type) {
	case nil:
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
	default:
		if reader, ok := b.(interface{ Read([]byte) (int, error) }); ok {
			req, err = http.NewRequestWithContext(ctx, method, url, readerWrapper{reader})
		} else {
			req, err = http.NewRequestWithContext(ctx, method, url, nil)
		}
	}

	return req, err
}

// DefaultTracedClient is a pre-configured traced HTTP client
// Initialize after calling InitOTEL
var DefaultTracedClient *TracedHTTPClient

// InitDefaultTracedClient initializes the default traced HTTP client
// Call this after InitOTEL
func InitDefaultTracedClient(opts ...TracedClientOption) {
	DefaultTracedClient = NewTracedHTTPClient(opts...)
}
