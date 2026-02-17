package middleware

import (
    "context"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const (
	// HeaderXRequestID is the canonical header for request tracing.
	HeaderXRequestID = "X-Request-ID"

	// LocalsKeyRequestID is the Fiber locals key for storing the request ID.
	LocalsKeyRequestID = "requestID"
)

// RequestIDMiddleware generates a unique request ID for each incoming request.
// If the client sends an X-Request-ID header, it is preserved. Otherwise, a new
// UUID v4 is generated. The ID is stored in Fiber locals for downstream use
// by loggers, tracing middleware, and error handlers.
func RequestIDMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := c.Get(HeaderXRequestID)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Store in locals for logger/tracing correlation
		c.Locals(LocalsKeyRequestID, requestID)

        // Inject into UserContext for logger.WithContext()
        // We use the string key "requestID" to match what services/logger.go expects
        ctx := c.UserContext()
        if ctx == nil {
            ctx = c.Context()
        }
        ctx = context.WithValue(ctx, "requestID", requestID)
        c.SetUserContext(ctx)

		// Set response header so clients can correlate requests
		c.Set(HeaderXRequestID, requestID)

		return c.Next()
	}
}

// GetRequestID retrieves the request ID from Fiber context locals.
// Returns empty string if no request ID is set.
func GetRequestID(c *fiber.Ctx) string {
	if id, ok := c.Locals(LocalsKeyRequestID).(string); ok {
		return id
	}
	return ""
}
