package middleware

import (
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
	"insight-engine-backend/services"
)

// PrivacyComplianceConfig holds configuration for privacy compliance
type PrivacyComplianceConfig struct {
	Enabled         bool     // Whether privacy compliance is enabled
	AllowedRegions  []string // List of allowed regions for data processing
	PIIFields       []string // List of fields considered PII
	RequireConsent  bool     // Whether consent is required for data processing
}

// PrivacyComplianceMiddleware provides privacy and data residency compliance
func PrivacyComplianceMiddleware(config PrivacyComplianceConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip compliance checks if disabled
		if !config.Enabled {
			return c.Next()
		}

		// Check data residency requirements
		clientIP := c.IP()
		clientRegion := determineRegionFromIP(clientIP)

		// If specific regions are required, validate against them
		if len(config.AllowedRegions) > 0 {
			allowed := false
			for _, allowedRegion := range config.AllowedRegions {
				if strings.EqualFold(clientRegion, allowedRegion) {
					allowed = true
					break
				}
			}

			if !allowed {
				services.LogSecurityEventWithContext(c, "data_residency_violation", 
					"Request from disallowed region", 
					map[string]interface{}{
						"client_region": clientRegion,
						"allowed_regions": config.AllowedRegions,
						"client_ip": clientIP,
					})

				return c.Status(403).JSON(fiber.Map{
					"error": "Access restricted",
					"type":  "compliance_violation",
					"message": "Access not allowed from your geographic region due to data residency requirements",
				})
			}
		}

		// Check for PII in request body
		var requestBody map[string]interface{}
		if err := c.BodyParser(&requestBody); err == nil {
			// Check if request contains PII
			piiDetected := false
			var piiFields []string

			for field, value := range requestBody {
				for _, piiField := range config.PIIFields {
					if strings.EqualFold(field, piiField) {
						piiDetected = true
						piiFields = append(piiFields, field)
						break
					}
				}

				// Also check if the value looks like PII (basic pattern matching)
				if strValue, ok := value.(string); ok {
					if isLikelyPII(strValue) {
						piiDetected = true
						piiFields = append(piiFields, field+" (detected)")
					}
				}
			}

			if piiDetected {
				// Log PII detection for audit purposes
				services.LogSecurityEventWithContext(c, "pii_detected", 
					"Personal Identifiable Information detected in request", 
					map[string]interface{}{
						"pii_fields": piiFields,
						"client_ip": clientIP,
						"path": c.Path(),
					})

				// If consent is required, check if it's provided
				if config.RequireConsent {
					consentProvided := checkConsent(c)
					if !consentProvided {
						return c.Status(400).JSON(fiber.Map{
							"error": "Consent required",
							"type":  "privacy_violation",
							"message": "Consent required for processing personal data",
						})
					}
				}
			}
		}

		// Continue to next middleware
		return c.Next()
	}
}

// determineRegionFromIP attempts to determine the geographic region from an IP address
// In a real implementation, this would use a geolocation service
func determineRegionFromIP(ip string) string {
	// This is a simplified implementation
	// In a real system, you'd use a geolocation service like MaxMind, IPinfo, etc.
	
	// For demo purposes, return a default region
	// A real implementation would look up the IP in a geolocation database
	return "US" // Default to US for demonstration
}

// isLikelyPII checks if a string value looks like PII based on patterns
func isLikelyPII(value string) bool {
	// Check for email pattern
	emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	if match, _ := regexp.MatchString(emailPattern, value); match {
		return true
	}

	// Check for phone number pattern (simplified)
	phonePattern := `^\+?[1-9]\d{1,14}$`
	if match, _ := regexp.MatchString(phonePattern, value); match {
		return true
	}

	// Check for credit card pattern (simplified)
	ccPattern := `^(?:\d{4}[-\s]?){3}\d{4}$|^(\d{4}){4}$`
	if match, _ := regexp.MatchString(ccPattern, value); match {
		return true
	}

	// Check for SSN pattern (US, simplified)
	ssnPattern := `^\d{3}-?\d{2}-?\d{4}$`
	if match, _ := regexp.MatchString(ssnPattern, value); match {
		return true
	}

	return false
}

// checkConsent checks if consent has been provided in the request
func checkConsent(c *fiber.Ctx) bool {
	// Check for consent in headers
	consentHeader := c.Get("X-Consent-Given")
	if consentHeader != "" && strings.ToLower(consentHeader) == "true" {
		return true
	}

	// Check for consent in cookies
	consentCookie := c.Cookies("consent-given")
	if consentCookie != "" && strings.ToLower(consentCookie) == "true" {
		return true
	}

	// Check for consent in request body (if applicable)
	var requestBody map[string]interface{}
	if err := c.BodyParser(&requestBody); err == nil {
		if consent, ok := requestBody["consent"]; ok {
			if consentBool, ok := consent.(bool); ok && consentBool {
				return true
			}
			if consentStr, ok := consent.(string); ok && strings.ToLower(consentStr) == "true" {
				return true
			}
		}
	}

	return false
}

// CreateDefaultPrivacyConfig creates a default privacy compliance configuration
func CreateDefaultPrivacyConfig() PrivacyComplianceConfig {
	return PrivacyComplianceConfig{
		Enabled:        true,
		AllowedRegions: []string{"US", "EU", "CA"}, // Common regions
		PIIFields: []string{
			"email", "phone", "address", "ssn", "social_security_number",
			"credit_card", "bank_account", "national_id", "passport",
			"driver_license", "medical_record", "health_info",
		},
		RequireConsent: false, // Can be enabled based on jurisdiction
	}
}