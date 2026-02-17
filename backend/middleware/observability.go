package middleware

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"insight-engine-backend/services"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

// TracingConfig holds configuration for distributed tracing
type TracingConfig struct {
	Enabled        bool
	Endpoint       string
	ServiceName    string
	SamplingRatio  float64
	HeadersToTrace []string
}

// InitTracer initializes the OpenTelemetry tracer
func InitTracer(config TracingConfig) (*sdktrace.TracerProvider, error) {
	if !config.Enabled {
		return nil, nil
	}

	// Create the OTLP exporter
	exporter, err := otlptracegrpc.New(
		context.Background(),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(config.Endpoint),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create resource with service information
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(config.ServiceName),
		semconv.ServiceVersion("1.0.0"),
	)

	// Create trace provider with sampler
	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(config.SamplingRatio))),
	)

	// Set global trace provider
	otel.SetTracerProvider(traceProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return traceProvider, nil
}

// DistributedTracingMiddleware provides distributed tracing for requests
func DistributedTracingMiddleware(config TracingConfig) fiber.Handler {
	tracer := otel.Tracer(config.ServiceName)

	return func(c *fiber.Ctx) error {
		// Extract trace context from headers
		propagator := otel.GetTextMapPropagator()
		ctx := propagator.Extract(context.Background(), propagation.HeaderCarrier(c.GetReqHeaders()))

		// Create a new span for this request
		spanName := fmt.Sprintf("%s %s", c.Method(), c.Path())
		ctx, span := tracer.Start(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				semconv.HTTPMethodKey.String(c.Method()),
				semconv.HTTPTargetKey.String(c.OriginalURL()),
				attribute.String("net.peer.ip", c.IP()),
				attribute.String("http.user_agent", c.Get("User-Agent")),
				attribute.String("http.route", c.Route().Path),
			),
		)
		defer span.End()

		// Add custom attributes from headers if configured
		for _, header := range config.HeadersToTrace {
			if value := c.Get(header); value != "" {
				span.SetAttributes(attribute.String("http.header."+strings.ToLower(header), value))
			}
		}

		// Store span context in Fiber locals for handlers to use
		c.Locals("trace_ctx", ctx)
		c.Locals("trace_id", span.SpanContext().TraceID().String())
		c.Locals("span_id", span.SpanContext().SpanID().String())

		// Continue processing the request
		err := c.Next()

		// Add response attributes to span
		span.SetAttributes(
			semconv.HTTPStatusCodeKey.Int(c.Response().StatusCode()),
		)

		// Record error if occurred
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}

		return err
	}
}

// EnhancedMetricsMiddleware provides enhanced metrics collection
func EnhancedMetricsMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		path := c.Path()
		method := c.Method()



		// Continue processing
		err := c.Next()

		// Calculate duration
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Response().StatusCode())

		// Record metrics

		// Standard metrics (HttpRequestsTotal, HttpRequestDuration, ActiveConnections) 

		// are now handled by PrometheusMiddleware in prometheus.go to avoid duplication and conflicts.

		// Note: error responses are already counted in the general counter above.
		// No duplicate increment needed.

		// Route-specific business metrics
		switch {
		case strings.HasPrefix(path, "/api/queries"):
			recordQueryMetrics(method, path, duration, status)
		case strings.HasPrefix(path, "/api/connections"):
			recordConnectionMetrics(method, path, duration, status)
		case strings.HasPrefix(path, "/api/ai/"):
			recordAIMetrics(method, path, duration, status)
		case strings.HasPrefix(path, "/api/dashboards"):
			recordDashboardMetrics(method, path)
		}

		return err
	}
}

// recordQueryMetrics records query-specific Prometheus metrics.
func recordQueryMetrics(method string, path string, duration float64, status string) {
	resultStatus := "success"
	if status >= "400" {
		resultStatus = "error"
	}

	queryType := "read"
	if method == "POST" {
		if strings.Contains(path, "/run") || strings.Contains(path, "/execute") {
			queryType = "execute"
		} else {
			queryType = "write"
		}
	}

	services.RecordQueryExecution("sql", resultStatus, queryType, duration)
}

// recordConnectionMetrics records connection-specific Prometheus metrics.
func recordConnectionMetrics(method string, path string, duration float64, status string) {
	resultStatus := "success"
	if status >= "400" {
		resultStatus = "error"
	}

	if strings.Contains(path, "/test") {
		services.RecordQueryExecution("connection_test", resultStatus, "test", duration)
	} else if strings.Contains(path, "/schema") {
		services.RecordQueryExecution("schema_discovery", resultStatus, "read", duration)
	}
}

// recordAIMetrics records AI-specific Prometheus metrics.
func recordAIMetrics(method string, path string, duration float64, status string) {
	resultStatus := "success"
	if status >= "400" {
		resultStatus = "error"
	}

	requestType := "generate"
	if strings.Contains(path, "/presentation") {
		requestType = "presentation"
	} else if strings.Contains(path, "/stream") {
		requestType = "stream"
	}

	services.RecordAIRequest(requestType, resultStatus, duration)
}

// recordDashboardMetrics records dashboard view metrics.
func recordDashboardMetrics(method string, path string) {
	if method == "GET" && !strings.HasSuffix(path, "/cards") && !strings.HasSuffix(path, "/positions") {
		// Extract dashboard ID if present (e.g., /api/dashboards/{id})
		parts := strings.Split(strings.TrimPrefix(path, "/api/dashboards"), "/")
		if len(parts) >= 2 && parts[1] != "" {
			services.RecordDashboardView(parts[1])
		}
	}
}

// PerformanceMonitoringMiddleware monitors performance and detects anomalies
func PerformanceMonitoringMiddleware(threshold time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		duration := time.Since(start)

		// Check if request took longer than threshold
		if duration > threshold {
			// Log slow request
			services.LogWarn("slow_request", "Request took longer than threshold", map[string]interface{}{
				"path":        c.Path(),
				"method":      c.Method(),
				"duration":    duration.String(),
				"threshold":   threshold.String(),
				"status_code": c.Response().StatusCode(),
				"client_ip":   c.IP(),
			})

			// Add slow request attribute to span if tracing is enabled
			if span := trace.SpanFromContext(c.Locals("trace_ctx").(context.Context)); span != nil {
				span.SetAttributes(attribute.Bool("performance.slow_request", true))
				span.SetAttributes(attribute.String("performance.duration", duration.String()))
			}
		}

		return err
	}
}

// ErrorTrackingMiddleware captures and tracks errors
func ErrorTrackingMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()

		// Capture errors for monitoring
		if err != nil || c.Response().StatusCode() >= 500 {
			errorType := "unknown"
			if err != nil {
				errorType = fmt.Sprintf("%T", err)
			}

			// Record error metrics
			// Standard metrics (HttpRequestsTotal) are already handled by PrometheusMiddleware.
			// We only need to log the error here.

			// Log error
			services.LogError("request_error", "Request resulted in error", map[string]interface{}{
				"path":        c.Path(),
				"method":      c.Method(),
				"error_type":  errorType,
				"status_code": c.Response().StatusCode(),
				"client_ip":   c.IP(),
				"user_agent":  c.Get("User-Agent"),
			})
		}

		return err
	}
}

// RequestSizeTrackingMiddleware tracks request and response sizes
func RequestSizeTrackingMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get request size
		reqSize := len(c.Request().Body())

		err := c.Next()

		// Get response size
		respSize := len(c.Response().Body())

		// Add to span if tracing is enabled
		if ctx, ok := c.Locals("trace_ctx").(context.Context); ok {
			if span := trace.SpanFromContext(ctx); span != nil {
				span.SetAttributes(
					attribute.Int("http.request.size", reqSize),
					attribute.Int("http.response.size", respSize),
				)
			}
		}

		return err
	}
}

// CreateDefaultTracingConfig creates a default tracing configuration
func CreateDefaultTracingConfig() TracingConfig {
	return TracingConfig{
		Enabled:        true,
		Endpoint:       "localhost:4317", // Default OTLP gRPC endpoint
		ServiceName:    "insight-engine-backend",
		SamplingRatio:  0.1, // 10% sampling rate
		HeadersToTrace: []string{"X-Request-ID", "X-Correlation-ID", "Authorization"},
	}
}
