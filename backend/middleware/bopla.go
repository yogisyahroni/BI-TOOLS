package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// MassAssignmentProtectionConfig defines the allowed fields for different models
type MassAssignmentProtectionConfig struct {
	// AllowedFields maps model names to their allowed field names
	AllowedFields map[string][]string
}

// MassAssignmentProtection creates a middleware that prevents mass assignment attacks
// by validating that only allowed fields are present in request bodies
func MassAssignmentProtection(config MassAssignmentProtectionConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get the route path to determine which model to validate against
		path := c.Path()

		// Determine the model based on the path
		modelName := getModelNameFromPath(path)
		if modelName == "" {
			// If we can't determine the model, skip validation
			return c.Next()
		}

		// Get allowed fields for this model
		allowedFields, exists := config.AllowedFields[modelName]
		if !exists {
			// If no allowed fields are defined for this model, skip validation
			return c.Next()
		}

		// Parse the request body into a map
		var requestData map[string]interface{}
		if err := c.BodyParser(&requestData); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body format",
				"type":  "parse_error",
			})
		}

		// Validate that only allowed fields are present
		for field := range requestData {
			if !isFieldAllowed(field, allowedFields) {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error":   "Mass assignment protection: Field not allowed",
					"field":   field,
					"message": "The field '" + field + "' is not allowed in this request",
					"type":    "mass_assignment_protection",
				})
			}
		}

		// If validation passes, restore the original body for the next handler
		if err := c.BodyParser(&requestData); err != nil {
			return err
		}

		return c.Next()
	}
}

// getModelNameFromPath determines the model name based on the request path
func getModelNameFromPath(path string) string {
	// This is a simplified implementation - in a real app, you'd want more sophisticated routing
	switch {
	case containsPathSegment(path, "/users"):
		return "user"
	case containsPathSegment(path, "/connections"):
		return "connection"
	case containsPathSegment(path, "/queries"):
		return "query"
	case containsPathSegment(path, "/dashboards"):
		return "dashboard"
	case containsPathSegment(path, "/dashboards/cards"):
		return "dashboard_card"
	case containsPathSegment(path, "/ai-providers"):
		return "ai_provider"
	case containsPathSegment(path, "/collections"):
		return "collection"
	default:
		return ""
	}
}

// containsPathSegment checks if a path contains a specific segment
func containsPathSegment(path, segment string) bool {
	// This is a simplified check - you might want more sophisticated path matching
	return len(path) >= len(segment) && path[:len(segment)] == segment
}

// isFieldAllowed checks if a field is in the allowed list
func isFieldAllowed(field string, allowedFields []string) bool {
	for _, allowed := range allowedFields {
		if field == allowed {
			return true
		}
	}
	return false
}

// CreateDefaultBoplaConfig creates a default configuration with common models and their allowed fields
func CreateDefaultBoplaConfig() MassAssignmentProtectionConfig {
	return MassAssignmentProtectionConfig{
		AllowedFields: map[string][]string{
			"user": {
				"name", "email", "password", "role", "isActive", "preferences",
			},
			"connection": {
				"name", "type", "config", "encryptedConfig", "isActive", "userId",
			},
			"query": {
				"name", "sql", "connectionId", "userId", "description", "tags",
			},
			"dashboard": {
				"name", "description", "userId", "config", "isPublic",
			},
			"dashboard_card": {
				"dashboardId", "queryId", "position", "config", "title", "type",
			},
			"ai_provider": {
				"name", "type", "config", "encryptedConfig", "isActive", "userId",
			},
			"collection": {
				"name", "description", "userId", "parentId", "color", "icon",
			},
		},
	}
}
