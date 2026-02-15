package middleware

import (
	"insight-engine-backend/services"

	"github.com/gofiber/fiber/v2"
)

// DegradationConfig holds configuration for the degradation middleware
type DegradationConfig struct {
	QueryExecutor *services.QueryExecutor
}

// DegradationMiddleware creates a middleware that checks system health
func DegradationMiddleware(config DegradationConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check database health via circuit breaker
		if config.QueryExecutor != nil && !config.QueryExecutor.IsHealthy() {
			c.Set("X-System-Status", "degraded")

			// Optional: Block state-changing operations if system is degraded
			// method := c.Method()
			// if method == "POST" || method == "PUT" || method == "DELETE" || method == "PATCH" {
			// 	return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			// 		"error": "System is in degraded mode. Write operations are temporarily disabled.",
			// 	})
			// }
		} else {
			c.Set("X-System-Status", "healthy")
		}

		return c.Next()
	}
}
