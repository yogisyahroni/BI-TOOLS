package middleware

import (
	"context"
	"fmt"
	"insight-engine-backend/services"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware validates NextAuth JWT tokens AND Embed Tokens
func AuthMiddleware(c *fiber.Ctx) error {
	// 1. Extract token from Authorization header or cookie
	tokenString := extractToken(c)
	if tokenString == "" {
		services.LogWarn("auth_no_token", "No token found in request", nil)
		return c.Status(401).JSON(fiber.Map{
			"status":  "error",
			"message": "Unauthorized: No token provided",
		})
	}

	// services.LogDebug("auth_token_found", "Token found, validating", map[string]interface{}{"token_length": len(tokenString)})

	// 2. Parse and validate JWT
	// Try NEXTAUTH_SECRET first (User Session)
	secret := os.Getenv("NEXTAUTH_SECRET")
	isEmbed := false

	claims := jwt.MapClaims{}
	parsedToken, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		// Validate signing method for security
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})

	// If failed, try EMBED_SECRET (Embed Token)
	if err != nil || !parsedToken.Valid {
		embedSecret := os.Getenv("EMBED_SECRET")
		if embedSecret != "" && embedSecret != secret {
			claims = jwt.MapClaims{} // Reset claims
			parsedToken, err = jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
				// Validate signing method for security
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
				}
				return []byte(embedSecret), nil
			})
			if err == nil && parsedToken.Valid {
				// Check if it is actually an embed token
				if sub, ok := claims["sub"].(string); ok && sub == "embed-token" {
					isEmbed = true
				}
			}
		}
	}

	if err != nil || !parsedToken.Valid {
		// DEBUG: Log to file
		f, _ := os.OpenFile("auth_debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if f != nil {
			f.WriteString(fmt.Sprintf("Time: %s | Error: %v | Token: %s\n", time.Now().String(), err, tokenString))
			f.Close()
		}

		services.LogWarn("auth_validation_failed", "Token validation failed", map[string]interface{}{"error": err})
		return c.Status(401).JSON(fiber.Map{
			"status":  "error",
			"message": "Unauthorized: Invalid token",
			"error":   err.Error(),
		})
	}

	// Additional security checks
	if exp, ok := claims["exp"].(float64); ok {
		if int64(exp) < time.Now().Unix() {
			services.LogWarn("auth_expired_token", "Expired token attempted", nil)
			return c.Status(401).JSON(fiber.Map{
				"status":  "error",
				"message": "Unauthorized: Token expired",
			})
		}
	}

	// services.LogDebug("auth_success", "Token validated successfully", map[string]interface{}{"user_id": claims["sub"], "is_embed": isEmbed})

	if isEmbed {
		c.Locals("isEmbed", true)
		if dashboardID, ok := claims["dashboard_id"].(string); ok {
			c.Locals("embedDashboardID", dashboardID)
		}
		// Embed tokens don't have a user ID, so we might set a placeholder or handle nil in handlers
		// For now, let's set a distinct userID to avoid confusing handlers that check for empty string
		c.Locals("userID", "embed-user")
		c.Locals("userId", "embed-user")
		c.Locals("workspaceID", "embed-workspace") // Placeholder
		return c.Next()
	}

	// 3. Extract user ID from claims and store in context (Normal User Flow)
	// 3. Extract user ID from claims and store in context (Normal User Flow)
	if sub, ok := claims["sub"].(string); ok {
		c.Locals("userID", sub)
		c.Locals("userId", sub)

		// Inject into UserContext
		ctx := c.UserContext()
		if ctx == nil {
			ctx = c.Context()
		}
		ctx = context.WithValue(ctx, "userID", sub)
		c.SetUserContext(ctx)

	} else if id, ok := claims["id"].(string); ok {
		c.Locals("userID", id)
		c.Locals("userId", id)

		// Inject into UserContext
		ctx := c.UserContext()
		if ctx == nil {
			ctx = c.Context()
		}
		ctx = context.WithValue(ctx, "userID", id)
		c.SetUserContext(ctx)
	}

	if email, ok := claims["email"].(string); ok {
		c.Locals("userEmail", email)
	}

	// 4. Set Workspace Context
	userIDVal := c.Locals("userID")

	fmt.Printf("DEBUG: AuthMiddleware - Claims keys: %v\n", getKeys(claims))
	if userIDVal != nil {
		fmt.Printf("DEBUG: AuthMiddleware - Extracted userID: %v\n", userIDVal)
		setWorkspaceContext(c, userIDVal.(string))
	} else {
		fmt.Printf("DEBUG: AuthMiddleware - NO userID extracted. Fallthrough.\n")
		c.Locals("workspaceID", "")
	}

	return c.Next()
}

func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// extractToken retrieves JWT from Authorization header or cookie
func extractToken(c *fiber.Ctx) string {
	// Try Authorization header first
	authHeader := c.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}

	// Try query parameter (for WebSockets or direct downloads)
	queryToken := c.Query("token")
	if queryToken != "" {
		return queryToken
	}

	// Try next-auth.session-token cookie (NextAuth default)
	cookie := c.Cookies("next-auth.session-token")
	if cookie != "" {
		return cookie
	}

	// Try __Secure-next-auth.session-token (HTTPS)
	cookie = c.Cookies("__Secure-next-auth.session-token")
	if cookie != "" {
		return cookie
	}

	return ""
}

// setWorkspaceContext ensures workspaceID is available in context
func setWorkspaceContext(c *fiber.Ctx, userID string) {
	workspaceID := c.Get("X-Workspace-ID")
	if workspaceID != "" {
		c.Locals("workspaceID", workspaceID)
		return
	}
	// Fallback
	c.Locals("workspaceID", "")
}
