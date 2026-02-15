package middleware

import (
	"insight-engine-backend/services"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

func MetricsMiddleware(c *fiber.Ctx) error {
	start := time.Now()
	path := c.Path()
	method := c.Method()

	services.ActiveConnections.Inc()
	defer services.ActiveConnections.Dec()

	err := c.Next()

	duration := time.Since(start).Seconds()
	status := strconv.Itoa(c.Response().StatusCode())

	services.HttpRequestsTotal.WithLabelValues(status, method, path).Inc()
	services.HttpRequestDuration.WithLabelValues(method, path).Observe(duration)

	return err
}
