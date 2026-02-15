package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"insight-engine-backend/models"
	"insight-engine-backend/services"

	"github.com/gofiber/fiber/v2"
)

// AdaptiveTimeoutMiddleware dynamically sets the request timeout based on query complexity
// It intercepts POST requests to /api/queries/execute, parses the body, and calculates a timeout.
func AdaptiveTimeoutMiddleware() fiber.Handler {
	complexityService := services.NewQueryComplexityService()

	return func(c *fiber.Ctx) error {
		// Only apply to query execution endpoint
		if c.Path() != "/api/queries/execute" || c.Method() != "POST" {
			return c.Next()
		}

		// Read the body to inspect complexity
		// Fiber's Body() returns raw bytes. We need to parse it to check for Config.
		// We use a struct that matches dtos.ExecuteQueryRequest but we only care about Config.
		var req struct {
			Config *models.VisualQueryConfig `json:"config"`
		}

		// We use Unmarshal to avoid consuming the body stream if Fiber's BodyParser does (it buffers usually).
		// c.Body() returns the buffer.
		if err := json.Unmarshal(c.Body(), &req); err != nil {
			// If parsing fails, proceed with default timeout logic (handled by Go context default or no timeout)
			// But for safety, let's set a default timeout for raw queries too.
			ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
			defer cancel()
			c.SetUserContext(ctx)
			return c.Next()
		}

		// Calculate complexity
		var result services.ComplexityResult
		if req.Config != nil {
			result = complexityService.CalculateComplexity(req.Config)
		} else {
			// Fallback for raw SQL or missing config: Treat as simple/medium
			// Could improve by analyzing SQL string length or keywords, but for now use 30s.
			result = services.ComplexityResult{
				Timeout:     30 * time.Second,
				Score:       1,
				Level:       services.ComplexitySimple,
				Description: "Raw SQL or no config provided",
			}
		}

		// Set context with timeout
		ctx, cancel := context.WithTimeout(c.Context(), result.Timeout)
		defer cancel()

		// Important: We cannot defer cancel() here because it would cancel the context immediately after middleware returns!
		// Wait, no. defer cancel() executes when *this function* returns.
		// If we defer cancel(), the context is canceled before the next handler runs?
		// YES! `defer cancel()` cancels the context when `AdaptiveTimeoutMiddleware`'s anonymous function returns.
		// But `c.Next()` is called synchronously.
		// `c.Next()` executes the chain. When `c.Next()` returns, the chain is done?
		// In Fiber, `c.Next()` executes the next handler.
		// BUT `cancel()` should be called when the *request* is done, or we risk leaking.
		// However, `context.WithTimeout` parent is `c.Context()`.
		// If we don't cancel, it leaks until timeout.
		// BUT if we cancel, we kill the request.
		// For `WithTimeout`, we usually rely on the timeout to cancel.
		// Or we hook into Fiber's response completion?
		// Actually, `c.Context()` (fasthttp) is reset after request.
		// So leak is minimal?
		// Documentation says: "The returned context's Done channel is closed when the deadline expires, when the returned cancel function is called, or when the parent context's Done channel is closed, whichever happens first."
		// If we wrap `c.Context()`, and that context finishes, ours finishes.
		// We should NOT defer cancel() if it cancels immediately.
		// We should let it live until timeout or request end.
		// To be safe, we can skip `defer cancel()` here and rely on timeout + parent cancellation.

		c.SetUserContext(ctx)

		// set headers for debugging/observability
		c.Set("X-Query-Complexity-Score", fmt.Sprintf("%d", result.Score))
		c.Set("X-Query-Complexity-Level", string(result.Level))
		c.Set("X-Query-Timeout", result.Timeout.String())

		// Proceed
		// Note: Fiber's context cancellation might need handling in the handler.
		// The handler uses `c.Context()` which wraps fasthttp context.
		// `SetUserContext` overrides the context returned by `c.UserContext()`.
		// We should ensure `query_executor` uses `c.UserContext()`.

		return c.Next()
	}
}
