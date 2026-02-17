package services

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	"insight-engine-backend/models"

	"gorm.io/gorm"
)

// SystemHealthService handles system health monitoring
type SystemHealthService struct {
	db         *gorm.DB
	redisCache *RedisCache
}

// NewSystemHealthService creates a new system health service
func NewSystemHealthService(db *gorm.DB, redisCache *RedisCache) *SystemHealthService {
	return &SystemHealthService{
		db:         db,
		redisCache: redisCache,
	}
}

// DatabaseHealth represents database health metrics
type DatabaseHealth struct {
	Status            string  `json:"status"`
	Latency           int64   `json:"latency_ms"`
	ActiveConnections int     `json:"active_connections"`
	MaxConnections    int     `json:"max_connections"`
	ErrorRate         float64 `json:"error_rate"`
}

// CacheHealth represents cache health metrics
type CacheHealth struct {
	Status      string  `json:"status"`
	HitRate     float64 `json:"hit_rate"`
	MemoryUsage int64   `json:"memory_usage_bytes"`
	KeysCount   int     `json:"keys_count"`
}

// ServiceStatus represents individual service status
type ServiceStatus struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"` // "up", "down", "degraded"
	Uptime    string    `json:"uptime"`
	LastCheck time.Time `json:"last_check"`
}

// SystemMetrics represents system resource metrics
type SystemMetrics struct {
	CPUUsage    float64 `json:"cpu_usage_percent"`
	MemoryUsage float64 `json:"memory_usage_percent"`
	DiskUsage   float64 `json:"disk_usage_percent"`
}

// ErrorLog represents a recent error entry
type ErrorLog struct {
	ID        uint      `json:"id"`
	Message   string    `json:"message"`
	Source    string    `json:"source"`
	CreatedAt time.Time `json:"created_at"`
}

// SystemHealth represents complete system health status
type SystemHealth struct {
	OverallStatus string                   `json:"overall_status"`
	HealthScore   int                      `json:"health_score"`
	Database      DatabaseHealth           `json:"database"`
	Cache         CacheHealth              `json:"cache"`
	Services      map[string]ServiceStatus `json:"services"`
	Metrics       SystemMetrics            `json:"metrics"`
	RecentErrors  []ErrorLog               `json:"recent_errors"`
	LastUpdated   time.Time                `json:"last_updated"`
}

// GetDatabaseHealth retrieves database health metrics
func (s *SystemHealthService) GetDatabaseHealth(ctx context.Context) (*DatabaseHealth, error) {
	health := &DatabaseHealth{}

	// Check connection with latency measurement
	start := time.Now()
	sqlDB, err := s.db.DB()
	if err != nil {
		health.Status = "critical"
		health.ErrorRate = 1.0
		return health, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Ping database
	if err := sqlDB.PingContext(ctx); err != nil {
		health.Status = "critical"
		health.ErrorRate = 1.0
		return health, fmt.Errorf("database ping failed: %w", err)
	}

	health.Latency = time.Since(start).Milliseconds()

	// Get connection stats
	stats := sqlDB.Stats()
	health.ActiveConnections = stats.InUse + stats.Idle
	health.MaxConnections = stats.MaxOpenConnections

	// Calculate error rate from wait count
	if stats.WaitCount > 0 {
		health.ErrorRate = float64(stats.WaitCount) / float64(stats.WaitCount+int64(stats.OpenConnections))
	}

	// Determine status
	if health.Latency > 5000 || health.ErrorRate > 0.5 {
		health.Status = "critical"
	} else if health.Latency > 1000 || health.ErrorRate > 0.1 {
		health.Status = "warning"
	} else {
		health.Status = "healthy"
	}

	return health, nil
}

// QueryPerformance represents query performance metrics
type QueryPerformance struct {
	TotalQueries     int64   `json:"total_queries"`
	SlowQueries      int64   `json:"slow_queries"`
	AvgExecutionTime float64 `json:"avg_execution_time_ms"`
	MaxExecutionTime float64 `json:"max_execution_time_ms"`
	MinExecutionTime float64 `json:"min_execution_time_ms"`
	QueriesPerSecond float64 `json:"queries_per_second"`
}

// GetQueryPerformance retrieves query performance metrics
func (s *SystemHealthService) GetQueryPerformance(ctx context.Context) (*QueryPerformance, error) {
	perf := &QueryPerformance{}

	// Get database stats
	sqlDB, err := s.db.DB()
	if err != nil {
		return perf, fmt.Errorf("failed to get database instance: %w", err)
	}

	stats := sqlDB.Stats()
	perf.TotalQueries = int64(stats.OpenConnections)

	// Calculate average wait duration (in ms)
	if stats.WaitCount > 0 {
		perf.AvgExecutionTime = float64(stats.WaitDuration.Milliseconds()) / float64(stats.WaitCount)
	}

	return perf, nil
}

// GetCacheStats retrieves cache statistics
func (s *SystemHealthService) GetCacheStats(ctx context.Context) (*CacheHealth, error) {
	health := &CacheHealth{}

	if s.redisCache == nil {
		health.Status = "unconfigured"
		return health, nil
	}

	// use RedisCache's Ping
	if err := s.redisCache.Ping(ctx); err != nil {
		health.Status = "critical"
		return health, fmt.Errorf("redis ping failed: %w", err)
	}

	// use RedisCache's GetStats
	stats, err := s.redisCache.GetStats(ctx)
	if err != nil {
		// If GetStats fails but Ping succeeded, it might be partial failure or just stats failure
		// But GetStats also handles connection failure.
		// If we are here, we might just log and return degraded or continue.
		LogWarn("cache_stats_error", "Failed to get detailed cache stats", map[string]interface{}{"error": err})
		health.Status = "degraded"
	}

	if stats != nil {
		health.KeysCount = int(stats.TotalKeys)
		health.MemoryUsage = stats.MemoryUsage
		health.HitRate = stats.HitRate / 100.0 // RedisCache returns percentage (0-100), we might want 0-1
	}

	// Determine status logic (simplified, real logic depends on thresholds)
	if health.Status == "" {
		health.Status = "healthy"
	}

	return health, nil
}

// GetSystemMetrics retrieves system resource metrics
func (s *SystemHealthService) GetSystemMetrics(ctx context.Context) (*SystemMetrics, error) {
	metrics := &SystemMetrics{}

	// Get memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Calculate memory usage percentage
	// Note: This is process memory, not system memory
	totalMemory := memStats.Sys
	usedMemory := memStats.Alloc
	if totalMemory > 0 {
		metrics.MemoryUsage = float64(usedMemory) / float64(totalMemory) * 100
	}

	// Get disk usage (simplified - checks current directory)
	// In production, use proper disk monitoring
	wd, _ := os.Getwd()
	_ = wd                // Placeholder for disk usage calculation
	metrics.DiskUsage = 0 // Would calculate from syscall.Statfs in production

	// CPU usage would typically come from system monitoring tools
	// This is a simplified placeholder
	metrics.CPUUsage = 0

	return metrics, nil
}

// GetServiceStatus checks all service statuses
func (s *SystemHealthService) GetServiceStatus(ctx context.Context) map[string]ServiceStatus {
	services := make(map[string]ServiceStatus)

	// Check Database
	dbHealth, _ := s.GetDatabaseHealth(ctx)
	services["database"] = ServiceStatus{
		Name:      "Database",
		Status:    dbHealth.Status,
		Uptime:    "running",
		LastCheck: time.Now(),
	}

	// Check Cache
	cacheHealth, _ := s.GetCacheStats(ctx)
	services["cache"] = ServiceStatus{
		Name:      "Redis Cache",
		Status:    cacheHealth.Status,
		Uptime:    "running",
		LastCheck: time.Now(),
	}

	// Email service (check via recent activity)
	services["email"] = ServiceStatus{
		Name:      "Email Service",
		Status:    "healthy", // Would check actual email queue
		Uptime:    "running",
		LastCheck: time.Now(),
	}

	// Scheduler service
	services["scheduler"] = ServiceStatus{
		Name:      "Scheduler",
		Status:    "healthy", // Would check scheduler health
		Uptime:    "running",
		LastCheck: time.Now(),
	}

	// WebSocket service
	services["websocket"] = ServiceStatus{
		Name:      "WebSocket",
		Status:    "healthy", // Would check WebSocket hub
		Uptime:    "running",
		LastCheck: time.Now(),
	}

	return services
}

// GetRecentErrors retrieves recent error logs
func (s *SystemHealthService) GetRecentErrors(ctx context.Context, limit int) ([]ErrorLog, error) {
	if limit == 0 {
		limit = 10
	}

	// Query recent audit logs with errors or from error sources
	var logs []models.AuditLog
	err := s.db.Where("action IN ?", []string{"ERROR", "FAILED"}).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve error logs: %w", err)
	}

	var errorLogs []ErrorLog
	for _, log := range logs {
		errorLogs = append(errorLogs, ErrorLog{
			ID:        log.ID,
			Message:   log.ResourceName,
			Source:    log.ResourceType,
			CreatedAt: log.CreatedAt,
		})
	}

	return errorLogs, nil
}

// GetHealthSummary calculates overall system health
func (s *SystemHealthService) GetHealthSummary(ctx context.Context) (*SystemHealth, error) {
	summary := &SystemHealth{
		Services:     make(map[string]ServiceStatus),
		LastUpdated:  time.Now(),
		RecentErrors: []ErrorLog{},
	}

	// Get database health
	dbHealth, err := s.GetDatabaseHealth(ctx)
	if err != nil {
		dbHealth = &DatabaseHealth{Status: "critical", ErrorRate: 1.0}
	}
	summary.Database = *dbHealth

	// Get cache health
	cacheHealth, err := s.GetCacheStats(ctx)
	if err != nil {
		cacheHealth = &CacheHealth{Status: "critical"}
	}
	summary.Cache = *cacheHealth

	// Get service statuses
	summary.Services = s.GetServiceStatus(ctx)

	// Get system metrics
	metrics, err := s.GetSystemMetrics(ctx)
	if err != nil {
		metrics = &SystemMetrics{}
	}
	summary.Metrics = *metrics

	// Get recent errors
	errors, err := s.GetRecentErrors(ctx, 10)
	if err == nil {
		summary.RecentErrors = errors
	}

	// Calculate overall health score (0-100)
	healthScore := 100

	// Database contributes 30 points
	if dbHealth.Status == "critical" {
		healthScore -= 30
	} else if dbHealth.Status == "warning" {
		healthScore -= 15
	}

	// Cache contributes 20 points
	if cacheHealth.Status == "critical" {
		healthScore -= 20
	} else if cacheHealth.Status == "warning" {
		healthScore -= 10
	} else if cacheHealth.Status == "unconfigured" {
		healthScore -= 5
	}

	// Services contribute 40 points (8 points each)
	for _, svc := range summary.Services {
		if svc.Status == "critical" {
			healthScore -= 8
		} else if svc.Status == "warning" || svc.Status == "degraded" {
			healthScore -= 4
		}
	}

	// Recent errors contribute 10 points
	if len(errors) > 5 {
		healthScore -= 10
	} else if len(errors) > 0 {
		healthScore -= len(errors) * 2
	}

	if healthScore < 0 {
		healthScore = 0
	}

	summary.HealthScore = healthScore

	// Determine overall status
	if healthScore >= 80 {
		summary.OverallStatus = "healthy"
	} else if healthScore >= 50 {
		summary.OverallStatus = "warning"
	} else {
		summary.OverallStatus = "critical"
	}

	return summary, nil
}

// HealthCheckResult represents a simple health check result
type HealthCheckResult struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

// HealthCheck performs a basic health check
func (s *SystemHealthService) HealthCheck(ctx context.Context) *HealthCheckResult {
	dbHealth, _ := s.GetDatabaseHealth(ctx)

	status := "healthy"
	if dbHealth.Status == "critical" {
		status = "critical"
	} else if dbHealth.Status == "warning" {
		status = "warning"
	}

	return &HealthCheckResult{
		Status:    status,
		Timestamp: time.Now().Format(time.RFC3339),
	}
}