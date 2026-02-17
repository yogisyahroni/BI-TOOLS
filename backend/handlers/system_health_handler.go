package handlers

import (
	"insight-engine-backend/services"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

// SystemHealthHandler handles system health API requests
type SystemHealthHandler struct {
	healthService *services.SystemHealthService
}

// NewSystemHealthHandler creates a new system health handler
func NewSystemHealthHandler(healthService *services.SystemHealthService) *SystemHealthHandler {
	return &SystemHealthHandler{
		healthService: healthService,
	}
}

// RegisterRoutes registers the health routes
func (h *SystemHealthHandler) RegisterRoutes(router fiber.Router) {
	// All routes under /api/admin/health
	router.Get("/health", h.GetHealth)
	router.Get("/health/database", h.GetDatabaseHealth)
	router.Get("/health/queries", h.GetQueryPerformance)
	router.Get("/health/cache", h.GetCacheStats)
	router.Get("/health/services", h.GetServiceStatus)
	router.Get("/health/metrics", h.GetSystemMetrics)
	router.Get("/health/errors", h.GetRecentErrors)
}

// GetHealth handles GET /api/admin/health
func (h *SystemHealthHandler) GetHealth(c *fiber.Ctx) error {
	ctx := c.UserContext()

	summary, err := h.healthService.GetHealthSummary(ctx)
	if err != nil {
		services.GlobalLogger.WithContext(ctx).Error("health_check_failed", "Failed to retrieve health summary", map[string]interface{}{
			"error": err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to retrieve health summary",
			"details": err.Error(),
		})
	}

	return c.JSON(summary)
}

// GetDatabaseHealth handles GET /api/admin/health/database
func (h *SystemHealthHandler) GetDatabaseHealth(c *fiber.Ctx) error {
	ctx := c.UserContext()

	health, err := h.healthService.GetDatabaseHealth(ctx)
	if err != nil {
		services.LogError("db_health_check_failed", "Failed to retrieve database health", map[string]interface{}{
			"error": err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to retrieve database health",
			"details": err.Error(),
		})
	}

	return c.JSON(health)
}

// GetQueryPerformance handles GET /api/admin/health/queries
func (h *SystemHealthHandler) GetQueryPerformance(c *fiber.Ctx) error {
	ctx := c.UserContext()

	perf, err := h.healthService.GetQueryPerformance(ctx)
	if err != nil {
		services.LogError("query_perf_check_failed", "Failed to retrieve query performance", map[string]interface{}{
			"error": err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to retrieve query performance",
			"details": err.Error(),
		})
	}

	return c.JSON(perf)
}

// GetCacheStats handles GET /api/admin/health/cache
func (h *SystemHealthHandler) GetCacheStats(c *fiber.Ctx) error {
	ctx := c.UserContext()

	stats, err := h.healthService.GetCacheStats(ctx)
	if err != nil {
		services.LogError("cache_stats_check_failed", "Failed to retrieve cache stats", map[string]interface{}{
			"error": err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to retrieve cache stats",
			"details": err.Error(),
		})
	}

	return c.JSON(stats)
}

// GetServiceStatus handles GET /api/admin/health/services
func (h *SystemHealthHandler) GetServiceStatus(c *fiber.Ctx) error {
	ctx := c.Context()

	services := h.healthService.GetServiceStatus(ctx)

	return c.JSON(fiber.Map{
		"services": services,
		"count":    len(services),
	})
}

// GetSystemMetrics handles GET /api/admin/health/metrics
func (h *SystemHealthHandler) GetSystemMetrics(c *fiber.Ctx) error {
	ctx := c.UserContext()

	metrics, err := h.healthService.GetSystemMetrics(ctx)
	if err != nil {
		services.LogError("system_metrics_check_failed", "Failed to retrieve system metrics", map[string]interface{}{
			"error": err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to retrieve system metrics",
			"details": err.Error(),
		})
	}

	return c.JSON(metrics)
}

// GetRecentErrors handles GET /api/admin/health/errors
func (h *SystemHealthHandler) GetRecentErrors(c *fiber.Ctx) error {
	ctx := c.UserContext()

	// Parse limit parameter
	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			if parsedLimit > 100 {
				parsedLimit = 100
			}
			limit = parsedLimit
		}
	}

	errors, err := h.healthService.GetRecentErrors(ctx, limit)
	if err != nil {
		services.LogError("recent_errors_check_failed", "Failed to retrieve recent errors", map[string]interface{}{
			"error": err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to retrieve recent errors",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"errors": errors,
		"count":  len(errors),
	})
}

// HealthCheckRequest represents health check query parameters
type HealthCheckRequest struct {
	Detailed bool `query:"detailed"`
}

// HealthCheck handles basic health check (can be used without auth for load balancers)
func (h *SystemHealthHandler) HealthCheck(c *fiber.Ctx) error {
	ctx := c.Context()

	result := h.healthService.HealthCheck(ctx)

	// Determine status code based on health status
	statusCode := fiber.StatusOK
	if result.Status == "critical" {
		statusCode = fiber.StatusServiceUnavailable
	} else if result.Status == "warning" {
		statusCode = fiber.StatusOK // Still OK but flagged
	}

	return c.Status(statusCode).JSON(result)
}

// HealthReportRequest represents health report export parameters
type HealthReportRequest struct {
	StartDate string `query:"start_date"`
	EndDate   string `query:"end_date"`
	Format    string `query:"format"` // "json" or "csv"
}

// ExportHealthReport handles GET /api/admin/health/export
func (h *SystemHealthHandler) ExportHealthReport(c *fiber.Ctx) error {
	ctx := c.UserContext()

	summary, err := h.healthService.GetHealthSummary(ctx)
	if err != nil {
		services.LogError("health_report_export_failed", "Failed to generate health report", map[string]interface{}{
			"error": err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to generate health report",
			"details": err.Error(),
		})
	}

	// Generate filename with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := "health_report_" + timestamp + ".json"

	// Set headers for file download
	c.Set("Content-Type", "application/json")
	c.Set("Content-Disposition", "attachment; filename="+filename)

	return c.JSON(summary)
}

// ReadinessProbe handles GET /health/ready for Kubernetes readiness checks.
// Returns 200 if the application can serve traffic (database is reachable).
// Returns 503 if any critical dependency is unavailable.
func (h *SystemHealthHandler) ReadinessProbe(c *fiber.Ctx) error {
	ctx := c.UserContext()

	checks := make(map[string]string)
	ready := true

	// Check database health
	dbHealth, err := h.healthService.GetDatabaseHealth(ctx)
	if err != nil || (dbHealth != nil && dbHealth.Status == "critical") {
		checks["database"] = "critical"
		ready = false
		services.LogError("readiness_probe_failed", "Database is critical", map[string]interface{}{
			"component": "database",
			"error":     err,
		})
	} else if dbHealth != nil {
		checks["database"] = dbHealth.Status
	}

	// Check cache health (non-critical: degraded cache doesn't block readiness)
	cacheHealth, err := h.healthService.GetCacheStats(ctx)
	if err != nil || (cacheHealth != nil && cacheHealth.Status == "critical") {
		checks["cache"] = "degraded"
		// Don't mark ready=false for cache, just log it
		services.LogWarn("readiness_probe_warning", "Cache is degraded", map[string]interface{}{
			"component": "cache",
			"error":     err,
		})
	} else if cacheHealth != nil {
		checks["cache"] = cacheHealth.Status
	}

	statusCode := fiber.StatusOK
	status := "ready"
	if !ready {
		statusCode = fiber.StatusServiceUnavailable
		status = "not_ready"
	}

	return c.Status(statusCode).JSON(fiber.Map{
		"status":    status,
		"checks":    checks,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// LivenessProbe handles GET /health/live for Kubernetes liveness checks.
// Returns 200 if the process is alive and the Go runtime is functioning.
// This is intentionally lightweight — no DB or external calls.
func (h *SystemHealthHandler) LivenessProbe(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":    "alive",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
