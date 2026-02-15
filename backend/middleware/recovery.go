package middleware

import (
	"fmt"
	"insight-engine-backend/services"

	"github.com/gofiber/fiber/v2"
)

func RecoveryMiddleware(c *fiber.Ctx) error {
	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}

			services.DefaultErrorTracker.TrackPanic("middleware_recovery", err, map[string]interface{}{
				"path":   c.Path(),
				"method": c.Method(),
				"ip":     c.IP(),
			})

			// Return 500
			c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "An unexpected error occurred. Administrators have been notified.",
			})
		}
	}()

	return c.Next()
}
