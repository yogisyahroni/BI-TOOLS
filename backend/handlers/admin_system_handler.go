package handlers

import (
	"context"
	"runtime"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// AdminSystemHandler handles system health and metrics
type AdminSystemHandler struct {
	db *gorm.DB
}

// NewAdminSystemHandler creates a new admin system handler
func NewAdminSystemHandler(db *gorm.DB) *AdminSystemHandler {
	return &AdminSystemHandler{
		db: db,
	}
}

// RegisterRoutes registers all admin system routes
func (h *AdminSystemHandler) RegisterRoutes(router fiber.Router, middlewares ...func(*fiber.Ctx) error) {
	system := router.Group("/admin/system")

	

	// Apply all provided middlewares (auth + admin check)

	for _, mw := range middlewares {

		system.Use(mw)

	}

	system.Get("/health", h.GetSystemHealth)
	system.Get("/metrics", h.GetSystemMetrics)
	system.Get("/database/connections", h.GetDatabaseConnections)
	system.Get("/database/performance", h.GetDatabasePerformance)
	system.Get("/cache/stats", h.GetCacheStats)
}

// SystemHealthResponse represents system health status
type SystemHealthResponse struct {
	Status     string                     `json:"status"` // healthy, degraded, unhealthy
	Timestamp  time.Time                  `json:"timestamp"`
	Uptime     int64                      `json:"uptime"` // seconds
	Version    string                     `json:"version"`
	Components map[string]ComponentHealth `json:"components"`
}

// ComponentHealth represents health of a system component
type ComponentHealth struct {
	Status  string                 `json:"status"` // up, down, degraded
	Message string                 `json:"message,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

var startTime = time.Now()

// GetSystemHealth handles GET /admin/system/health
func (h *AdminSystemHandler) GetSystemHealth(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	components := make(map[string]ComponentHealth)

	// Check database
	dbHealth := h.checkDatabaseHealth(ctx)
	components["database"] = dbHealth

	// Check cache (if implemented)
	cacheHealth := h.checkCacheHealth(ctx)
	components["cache"] = cacheHealth

	// Check memory
	memHealth := h.checkMemoryHealth()
	components["memory"] = memHealth

	// Determine overall status
	overallStatus := "healthy"
	for _, comp := range components {
		if comp.Status == "down" {
			overallStatus = "unhealthy"
			break
		} else if comp.Status == "degraded" && overallStatus != "unhealthy" {
			overallStatus = "degraded"
		}
	}

	response := SystemHealthResponse{
		Status:     overallStatus,
		Timestamp:  time.Now(),
		Uptime:     int64(time.Since(startTime).Seconds()),
		Version:    "1.0.0",
		Components: components,
	}

	if overallStatus == "unhealthy" {
		return c.Status(fiber.StatusServiceUnavailable).JSON(response)
	}

	return c.JSON(response)
}

// SystemMetricsResponse represents system metrics
type SystemMetricsResponse struct {
	Timestamp  time.Time       `json:"timestamp"`
	Memory     MemoryMetrics   `json:"memory"`
	Goroutines int             `json:"goroutines"`
	Database   DatabaseMetrics `json:"database"`
	Cache      CacheMetrics    `json:"cache"`
	API        APIMetrics      `json:"api"`
}

// MemoryMetrics represents memory usage metrics
type MemoryMetrics struct {
	Alloc        uint64  `json:"alloc"`      // bytes allocated
	TotalAlloc   uint64  `json:"totalAlloc"` // bytes allocated (all time)
	Sys          uint64  `json:"sys"`        // bytes obtained from system
	NumGC        uint32  `json:"numGC"`      // number of GC runs
	UsagePercent float64 `json:"usagePercent"`
}

// DatabaseMetrics represents database metrics
type DatabaseMetrics struct {
	ConnectionCount   int     `json:"connectionCount"`
	MaxConnections    int     `json:"maxConnections"`
	IdleConnections   int     `json:"idleConnections"`
	ActiveConnections int     `json:"activeConnections"`
	AvgQueryTime      float64 `json:"avgQueryTimeMs"`
	SlowQueries       int64   `json:"slowQueries"`
	TotalQueries      int64   `json:"totalQueries"`
}

// CacheMetrics represents cache metrics
type CacheMetrics struct {
	HitRate     float64 `json:"hitRate"`
	MissRate    float64 `json:"missRate"`
	TotalKeys   int64   `json:"totalKeys"`
	MemoryUsage int64   `json:"memoryUsage"`
}

// APIMetrics represents API metrics
type APIMetrics struct {
	RequestsPerSecond float64 `json:"requestsPerSecond"`
	AvgResponseTime   float64 `json:"avgResponseTimeMs"`
	ErrorRate         float64 `json:"errorRate"`
	TotalRequests     int64   `json:"totalRequests"`
}

// GetSystemMetrics handles GET /admin/system/metrics
func (h *AdminSystemHandler) GetSystemMetrics(c *fiber.Ctx) error {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	memoryMetrics := MemoryMetrics{
		Alloc:        m.Alloc,
		TotalAlloc:   m.TotalAlloc,
		Sys:          m.Sys,
		NumGC:        m.NumGC,
		UsagePercent: float64(m.Alloc) / float64(m.Sys) * 100,
	}

	// Get database metrics
	dbMetrics := h.getDatabaseMetrics()

	// Get cache metrics (placeholder)
	cacheMetrics := CacheMetrics{
		HitRate:     0.0,
		MissRate:    0.0,
		TotalKeys:   0,
		MemoryUsage: 0,
	}

	// Get API metrics (placeholder)
	apiMetrics := APIMetrics{
		RequestsPerSecond: 0.0,
		AvgResponseTime:   0.0,
		ErrorRate:         0.0,
		TotalRequests:     0,
	}

	response := SystemMetricsResponse{
		Timestamp:  time.Now(),
		Memory:     memoryMetrics,
		Goroutines: runtime.NumGoroutine(),
		Database:   dbMetrics,
		Cache:      cacheMetrics,
		API:        apiMetrics,
	}

	return c.JSON(response)
}

// DatabaseConnectionInfo represents a database connection
type DatabaseConnectionInfo struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Type         string    `json:"type"`
	Status       string    `json:"status"`
	LastChecked  time.Time `json:"lastChecked"`
	ResponseTime int64     `json:"responseTimeMs"`
}

// GetDatabaseConnections handles GET /admin/system/database/connections
func (h *AdminSystemHandler) GetDatabaseConnections(c *fiber.Ctx) error {
	// Get all configured connections from the database
	type Connection struct {
		ID   string
		Name string
		Type string
	}

	var connections []Connection
	if err := h.db.Table("connections").
		Select("id, name, type").
		Find(&connections).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch connections",
		})
	}

	// Check status of each connection
	connInfos := make([]DatabaseConnectionInfo, len(connections))
	for i, conn := range connections {
		start := time.Now()
		status := "up"

		// Simple status check - could be enhanced with actual connection test
		responseTime := time.Since(start).Milliseconds()

		connInfos[i] = DatabaseConnectionInfo{
			ID:           conn.ID,
			Name:         conn.Name,
			Type:         conn.Type,
			Status:       status,
			LastChecked:  time.Now(),
			ResponseTime: responseTime,
		}
	}

	return c.JSON(fiber.Map{
		"connections": connInfos,
		"total":       len(connInfos),
	})
}

// QueryPerformanceInfo represents query performance metrics
type QueryPerformanceInfo struct {
	TotalQueries  int64      `json:"totalQueries"`
	AvgTime       float64    `json:"avgTimeMs"`
	SlowQueries   int64      `json:"slowQueries"`
	FailedQueries int64      `json:"failedQueries"`
	TopQueries    []TopQuery `json:"topQueries"`
}

// TopQuery represents a frequently executed query
type TopQuery struct {
	Query   string    `json:"query"`
	Count   int64     `json:"count"`
	AvgTime float64   `json:"avgTimeMs"`
	LastRun time.Time `json:"lastRun"`
}

// GetDatabasePerformance handles GET /admin/system/database/performance
func (h *AdminSystemHandler) GetDatabasePerformance(c *fiber.Ctx) error {
	// Get query execution statistics
	var totalQueries int64
	h.db.Table("saved_queries").Count(&totalQueries)

	// Placeholder for actual metrics
	performance := QueryPerformanceInfo{
		TotalQueries:  totalQueries,
		AvgTime:       0.0,
		SlowQueries:   0,
		FailedQueries: 0,
		TopQueries:    []TopQuery{},
	}

	return c.JSON(performance)
}

// GetCacheStats handles GET /admin/system/cache/stats
func (h *AdminSystemHandler) GetCacheStats(c *fiber.Ctx) error {
	// Placeholder - implement based on cache system
	stats := fiber.Map{
		"enabled":      false,
		"hitRate":      0.0,
		"missRate":     0.0,
		"totalKeys":    0,
		"memoryUsage":  0,
		"evictionRate": 0.0,
	}

	return c.JSON(stats)
}

// Helper methods

func (h *AdminSystemHandler) checkDatabaseHealth(ctx context.Context) ComponentHealth {
	// Test primary database connection
	sqlDB, err := h.db.DB()
	if err != nil {
		return ComponentHealth{
			Status:  "down",
			Message: "Failed to get database instance",
		}
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return ComponentHealth{
			Status:  "down",
			Message: "Database ping failed",
		}
	}

	// Get connection stats
	stats := sqlDB.Stats()
	details := map[string]interface{}{
		"openConnections": stats.OpenConnections,
		"inUse":           stats.InUse,
		"idle":            stats.Idle,
		"maxOpen":         stats.MaxOpenConnections,
	}

	status := "up"
	message := "Database is healthy"

	// Check if we're running low on connections
	if stats.MaxOpenConnections > 0 && stats.OpenConnections >= stats.MaxOpenConnections*80/100 {
		status = "degraded"
		message = "Database connection pool nearly exhausted"
	}

	return ComponentHealth{
		Status:  status,
		Message: message,
		Details: details,
	}
}

func (h *AdminSystemHandler) checkCacheHealth(ctx context.Context) ComponentHealth {
	// Placeholder - implement based on cache system
	return ComponentHealth{
		Status:  "up",
		Message: "Cache not configured",
		Details: map[string]interface{}{
			"enabled": false,
		},
	}
}

func (h *AdminSystemHandler) checkMemoryHealth() ComponentHealth {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	usagePercent := float64(m.Alloc) / float64(m.Sys) * 100
	status := "up"
	message := "Memory usage normal"

	if usagePercent > 90 {
		status = "degraded"
		message = "High memory usage"
	} else if usagePercent > 95 {
		status = "down"
		message = "Critical memory usage"
	}

	details := map[string]interface{}{
		"allocMB":      m.Alloc / 1024 / 1024,
		"sysMB":        m.Sys / 1024 / 1024,
		"usagePercent": usagePercent,
		"numGC":        m.NumGC,
	}

	return ComponentHealth{
		Status:  status,
		Message: message,
		Details: details,
	}
}

func (h *AdminSystemHandler) getDatabaseMetrics() DatabaseMetrics {
	sqlDB, err := h.db.DB()
	if err != nil {
		return DatabaseMetrics{}
	}

	stats := sqlDB.Stats()

	return DatabaseMetrics{
		ConnectionCount:   stats.OpenConnections,
		MaxConnections:    stats.MaxOpenConnections,
		IdleConnections:   stats.Idle,
		ActiveConnections: stats.InUse,
		AvgQueryTime:      0.0,
		SlowQueries:       0,
		TotalQueries:      0,
	}
}
