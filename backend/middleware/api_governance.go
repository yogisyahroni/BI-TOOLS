package middleware

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"insight-engine-backend/services"
)

// APIVersionConfig holds configuration for API versioning
type APIVersionConfig struct {
	DefaultVersion string            // Default API version if none specified
	SupportedVersions []string      // List of supported API versions
	DeprecatedVersions []string     // List of deprecated API versions
	MinSupportedVersion string      // Minimum supported version
}

// APIVersionMiddleware handles API versioning and deprecation
func APIVersionMiddleware(config APIVersionConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract version from header, query parameter, or path
		version := extractAPIVersion(c, config.DefaultVersion)

		// Validate version
		if !isVersionSupported(version, config.SupportedVersions) {
			services.LogSecurityEventWithContext(c, "api_version_unsupported", 
				"Unsupported API version requested", 
				map[string]interface{}{
					"requested_version": version,
					"supported_versions": config.SupportedVersions,
					"client_ip": c.IP(),
					"path": c.Path(),
				})

			return c.Status(400).JSON(fiber.Map{
				"error": "Unsupported API version",
				"type":  "api_version_error",
				"supported_versions": config.SupportedVersions,
			})
		}

		// Check if version is deprecated
		if isVersionDeprecated(version, config.DeprecatedVersions) {
			services.LogWarn("api_version_deprecated", 
				"Deprecated API version used", 
				map[string]interface{}{
					"version": version,
					"path": c.Path(),
					"client_ip": c.IP(),
				})
				
			// Add warning header
			c.Set("Warning", fmt.Sprintf("299 - \"API version %s is deprecated\"", version))
		}

		// Check if version is below minimum supported
		if isVersionBelowMinimum(version, config.MinSupportedVersion) {
			services.LogSecurityEventWithContext(c, "api_version_obsolete", 
				"Obsolete API version used", 
				map[string]interface{}{
					"version": version,
					"minimum_supported": config.MinSupportedVersion,
					"path": c.Path(),
					"client_ip": c.IP(),
				})

			return c.Status(426).JSON(fiber.Map{
				"error": "API version too old",
				"type":  "upgrade_required",
				"minimum_supported_version": config.MinSupportedVersion,
				"upgrade_url": "/docs/api-upgrade-guide",
			})
		}

		// Store version in context for handlers to use
		c.Locals("apiVersion", version)

		return c.Next()
	}
}

// extractAPIVersion extracts the API version from headers, query params, or path
func extractAPIVersion(c *fiber.Ctx, defaultVersion string) string {
	// 1. Check Accept header for version (e.g., application/vnd.api.v1+json)
	acceptHeader := c.Get("Accept")
	if strings.Contains(acceptHeader, "application/vnd.api.v") {
		// Extract version from accept header
		for _, version := range []string{"v1", "v2", "v3"} {
			if strings.Contains(acceptHeader, version) {
				return strings.TrimPrefix(version, "v")
			}
		}
	}

	// 2. Check custom header (X-API-Version)
	apiVersion := c.Get("X-API-Version")
	if apiVersion != "" {
		return strings.TrimPrefix(apiVersion, "v")
	}

	// 3. Check query parameter
	queryVersion := c.Query("api-version")
	if queryVersion != "" {
		return strings.TrimPrefix(queryVersion, "v")
	}

	// 4. Check path (e.g., /api/v1/users)
	path := c.Path()
	if strings.HasPrefix(path, "/api/v") {
		parts := strings.Split(path, "/")
		if len(parts) > 2 && strings.HasPrefix(parts[2], "v") {
			return strings.TrimPrefix(parts[2], "v")
		}
	}

	// 5. Return default if no version specified
	return defaultVersion
}

// isVersionSupported checks if the version is in the supported list
func isVersionSupported(version string, supported []string) bool {
	for _, supportedVersion := range supported {
		if version == supportedVersion {
			return true
		}
	}
	return false
}

// isVersionDeprecated checks if the version is in the deprecated list
func isVersionDeprecated(version string, deprecated []string) bool {
	for _, depVersion := range deprecated {
		if version == depVersion {
			return true
		}
	}
	return false
}

// isVersionBelowMinimum checks if the version is below the minimum supported
func isVersionBelowMinimum(version, minVersion string) bool {
	// Simple string comparison for now - in a real system you'd want semantic version comparison
	return version < minVersion
}

// CreateDefaultAPIVersionConfig creates a default API version configuration
func CreateDefaultAPIVersionConfig() APIVersionConfig {
	return APIVersionConfig{
		DefaultVersion: "1",
		SupportedVersions: []string{"1", "2"},
		DeprecatedVersions: []string{"0"},
		MinSupportedVersion: "1",
	}
}

// APIRateLimitConfig holds configuration for API rate limiting
type APIRateLimitConfig struct {
	Enabled bool
	// Add more rate limit configuration options as needed
}

// APIRateLimitMiddleware provides API-specific rate limiting
func APIRateLimitMiddleware(config APIRateLimitConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !config.Enabled {
			return c.Next()
		}

		// This would integrate with the existing rate limiting system
		// For now, we'll just continue to the next middleware
		return c.Next()
	}
}