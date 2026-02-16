package middleware

import (
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
	"insight-engine-backend/services"
)

// SQLInjectionMiddleware provides protection against SQL injection attacks
func SQLInjectionMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check request body for potential SQL injection
		body := string(c.Body())
		if isSQLInjection(body) {
			ip := c.IP()
			path := c.Path()
			
			services.LogSQLInjectionAttempt(
				ip,
				path,
				body,
				"Potential SQL injection pattern detected in request body",
			)
			
			return c.Status(400).JSON(fiber.Map{
				"error": "Malformed request",
				"type":  "security_violation",
				"message": "Request blocked due to potential security threat",
			})
		}

		// Check query parameters
		query := c.Request().URI().String()
		if isSQLInjection(query) {
			ip := c.IP()
			path := c.Path()
			
			services.LogSQLInjectionAttempt(
				ip,
				path,
				query,
				"Potential SQL injection pattern detected in query parameters",
			)
			
			return c.Status(400).JSON(fiber.Map{
				"error": "Malformed request",
				"type":  "security_violation",
				"message": "Request blocked due to potential security threat",
			})
		}

		// Continue to next middleware
		return c.Next()
	}
}

// isSQLInjection checks if a string contains potential SQL injection patterns
func isSQLInjection(input string) bool {
	if input == "" {
		return false
	}

	// Convert to lowercase for pattern matching
	lowerInput := strings.ToLower(input)

	// Dangerous SQL keywords/patterns
	dangerousPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(\b(drop|truncate|alter|create|delete|update|insert|exec|execute|sp_|xp_|sysobjects|syscolumns)\b)`),
		regexp.MustCompile(`(;\s*(drop|truncate|alter|create|delete|update|insert|exec|execute))`),
		regexp.MustCompile(`('|--)`),
		regexp.MustCompile(`(union\s+select)`),
		regexp.MustCompile(`(waitfor\s+delay|pg_sleep|sleep|benchmark\s*\(|time_to_sec)`),
		regexp.MustCompile(`(exec\s*\(|execute\s*\(|sp_|xp_)`),
		regexp.MustCompile(`(char\s*\(|ascii\s*\(|substring\s*\(|mid\s*\(|like\s+char|like\s+ascii)`),
		regexp.MustCompile(`(or\s+1\s*=\s*1|and\s+1\s*=\s*1)`),
		regexp.MustCompile(`(order\s+by\s+\d+)`), // Often used in SQL injection enumeration
	}

	for _, pattern := range dangerousPatterns {
		if pattern.MatchString(lowerInput) {
			return true
		}
	}

	return false
}

// ValidateInputForSQL validates input strings to ensure they don't contain SQL injection patterns
func ValidateInputForSQL(input string) error {
	if isSQLInjection(input) {
		return &fiber.Error{
			Code:    400,
			Message: "Input contains potential SQL injection patterns",
		}
	}
	return nil
}