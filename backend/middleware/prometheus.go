package middleware

import (
	"strconv"
	"time"

	"insight-engine-backend/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// PrometheusMiddleware records metrics for each request
func PrometheusMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		path := c.Path()
		method := c.Method()

		// Skip metrics endpoint to avoid skewing data
		if path == "/metrics" {
			return c.Next()
		}

		services.ActiveConnections.Inc()
		defer services.ActiveConnections.Dec()

		err := c.Next()

		status := strconv.Itoa(c.Response().StatusCode())
		duration := time.Since(start).Seconds()

		services.HttpRequestsTotal.WithLabelValues(method, path, status).Inc()
		services.HttpRequestDuration.WithLabelValues(method, path, status).Observe(duration)

		return err
	}
}

// MetricsHandler exposes the Prometheus metrics endpoint
func MetricsHandler() fiber.Handler {
	return adaptor.HTTPHandler(promhttp.Handler())
}
