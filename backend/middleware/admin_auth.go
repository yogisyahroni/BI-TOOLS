package middleware

import (
	"insight-engine-backend/services"

	"github.com/gofiber/fiber/v2"
)

// RequireAdmin ensures the current user has admin role before allowing access
func RequireAdmin(c *fiber.Ctx) error {
	// Get user ID from context (set by AuthMiddleware)
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized - invalid session",
		})
	}

	// Get user role from context (set by AuthMiddleware)
	userRole, ok := c.Locals("userRole").(string)
	if !ok {
		// If role not in context, log and deny
		services.LogError("admin_auth_failed", "Role not found in user context", map[string]interface{}{
			"user_id": userID,
			"path":    c.Path(),
			"method":  c.Method(),
		})
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Forbidden - insufficient permissions",
		})
	}

	// Check if user is admin
	if userRole != "admin" {
		services.LogWarn("unauthorized_admin_access_attempt", "Non-admin user attempted admin endpoint access", map[string]interface{}{
			"user_id":   userID,
			"user_role": userRole,
			"path":      c.Path(),
			"method":    c.Method(),
			"ip":        c.IP(),
		})
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Forbidden - admin access required",
		})
	}

	// Log successful admin access for audit trail
	services.LogInfo("admin_access_granted", "Admin user accessed admin endpoint", map[string]interface{}{
		"user_id": userID,
		"path":    c.Path(),
		"method":  c.Method(),
	})

	return c.Next()
}
