package middleware

import (
	"fmt"
	"time"

	"insight-engine-backend/services"

	"github.com/gofiber/fiber/v2"
)

// CacheConfig defines configuration for the cache middleware
type CacheConfig struct {
	RedisCache *services.RedisCache
	TTL        time.Duration
	KeyPrefix  string
}

// CacheMiddleware creates a fiber handler for response caching
func CacheMiddleware(config CacheConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Only cache GET requests
		if c.Method() != fiber.MethodGet {
			return c.Next()
		}

		// Skip if cache is not configured
		if config.RedisCache == nil {
			return c.Next()
		}

		// Generate cache key
		// Use URL and query string
		key := fmt.Sprintf("%scache:api:%s?%s", config.KeyPrefix, c.Path(), string(c.Request().URI().QueryString()))

		// 1. Try to get from cache
		cached, err := config.RedisCache.Get(c.Context(), key)
		if err == nil && cached != nil {
			c.Set("X-Cache", "HIT")
			c.Set("Content-Type", "application/json") // Assuming JSON for now, could store content type too
			return c.Send(cached)
		}

		// 2. Execute handler
		c.Set("X-Cache", "MISS")
		if err := c.Next(); err != nil {
			return err
		}

		// 3. Cache successful responses
		if c.Response().StatusCode() == fiber.StatusOK {
			body := c.Response().Body()
			if len(body) > 0 {
				_ = config.RedisCache.Set(c.Context(), key, body, config.TTL)
			}
		}

		return nil
	}
}
