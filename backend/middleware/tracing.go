package middleware

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// TracerName is the name of the tracer for this package
const TracerName = "insight-engine-backend/middleware"

// TracingMiddleware adds OpenTelemetry tracing to Fiber requests
func TracingMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		tracer := otel.Tracer(TracerName)
		globalPropagator := otel.GetTextMapPropagator()

		// Extract context from headers
		ctx := globalPropagator.Extract(c.Context(), propagation.HeaderCarrier(c.GetReqHeaders()))

		// Create span
		spanName := fmt.Sprintf("%s %s", c.Method(), c.Path())
		ctx, span := tracer.Start(ctx, spanName, trace.WithAttributes(
			attribute.String("http.method", c.Method()),
			attribute.String("http.path", c.Path()),
			attribute.String("http.url", c.OriginalURL()),
			attribute.String("http.user_agent", c.Get("User-Agent")),
			attribute.String("http.client_ip", c.IP()),
		))
		defer span.End()

		// Inject context into Fiber locals for use in handlers
		c.Locals("ctx", ctx)
		c.Locals("tracer", tracer)
        
        // Critical: Update Fiber's UserContext so c.UserContext() returns the traced context
        c.SetUserContext(ctx)

		// Also set the trace ID in the response header for easier debugging
		if sc := span.SpanContext(); sc.IsValid() {
			c.Set("X-Trace-ID", sc.TraceID().String())
		}

		// Continue processing
		err := c.Next()

		// Add response attributes
		span.SetAttributes(
			attribute.Int("http.status_code", c.Response().StatusCode()),
		)

		if err != nil {
			span.RecordError(err)
			span.SetAttributes(attribute.String("error", err.Error()))
		}

		return err
	}
}

// GetTraceContext retrieves the OpenTelemetry context from Fiber locals
func GetTraceContext(c *fiber.Ctx) context.Context {
	if ctx, ok := c.Locals("ctx").(context.Context); ok {
		return ctx
	}
	return c.Context()
}

// GetTracer retrieves the OpenTelemetry tracer from Fiber locals
func GetTracer(c *fiber.Ctx) trace.Tracer {
	if tracer, ok := c.Locals("tracer").(trace.Tracer); ok {
		return tracer
	}
	return otel.Tracer(TracerName)
}
