package services

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
)

// APMService provides Application Performance Monitoring capabilities
type APMService struct {
	serviceName string
	startTime   time.Time
}

// NewAPMService creates a new APM service
func NewAPMService(serviceName string) *APMService {
	return &APMService{
		serviceName: serviceName,
		startTime:   time.Now(),
	}
}

// PerformanceMetrics represents application performance metrics
type PerformanceMetrics struct {
	ServiceName string             `json:"service_name"`
	Uptime      time.Duration      `json:"uptime"`
	CPU         CPUMetrics         `json:"cpu"`
	Memory      MemoryMetrics      `json:"memory"`
	Disk        DiskMetrics        `json:"disk"`
	Application ApplicationMetrics `json:"application"`
	Timestamp   time.Time          `json:"timestamp"`
}

// CPUMetrics represents CPU usage metrics
type CPUMetrics struct {
	PercentUsed float64 `json:"percent_used"`
	Count       int     `json:"count"`
	LoadAvg     LoadAvg `json:"load_avg"`
}

// LoadAvg represents system load average
type LoadAvg struct {
	OneMin     float64 `json:"one_min"`
	FiveMin    float64 `json:"five_min"`
	FifteenMin float64 `json:"fifteen_min"`
}

// MemoryMetrics represents memory usage metrics
type MemoryMetrics struct {
	Total       uint64  `json:"total"`
	Available   uint64  `json:"available"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"used_percent"`
}

// DiskMetrics represents disk usage metrics
type DiskMetrics struct {
	Total       uint64  `json:"total"`
	Free        uint64  `json:"free"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"used_percent"`
}

// ApplicationMetrics represents application-specific metrics
type ApplicationMetrics struct {
	Goroutines int     `json:"goroutines"`
	GCStats    GCStats `json:"gc_stats"`
	Threads    int     `json:"threads"`
}

// GCStats represents garbage collector statistics
type GCStats struct {
	NumGC        uint32   `json:"num_gc"`
	PauseTotalNs uint64   `json:"pause_total_ns"`
	LastPauseNs  uint64   `json:"last_pause_ns"`
	PauseNs      []uint64 `json:"pause_ns"`
}

// GetPerformanceMetrics collects and returns performance metrics
func (a *APMService) GetPerformanceMetrics(ctx context.Context) (*PerformanceMetrics, error) {
	metrics := &PerformanceMetrics{
		ServiceName: a.serviceName,
		Uptime:      time.Since(a.startTime),
		Timestamp:   time.Now(),
	}

	// Collect CPU metrics
	cpuPercent, err := cpu.PercentWithContext(ctx, 0, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU percent: %w", err)
	}

	cpuCount, err := cpu.Counts(false)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU count: %w", err)
	}

	metrics.CPU = CPUMetrics{
		PercentUsed: cpuPercent[0],
		Count:       cpuCount,
	}

	// Collect memory metrics
	vmStat, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get memory stats: %w", err)
	}

	metrics.Memory = MemoryMetrics{
		Total:       vmStat.Total,
		Available:   vmStat.Available,
		Used:        vmStat.Used,
		UsedPercent: vmStat.UsedPercent,
	}

	// Collect disk metrics
	diskStat, err := disk.UsageWithContext(ctx, "/")
	if err != nil {
		// Try to get disk usage for current working directory
		diskStat, err = disk.UsageWithContext(ctx, ".")
		if err != nil {
			return nil, fmt.Errorf("failed to get disk usage: %w", err)
		}
	}

	metrics.Disk = DiskMetrics{
		Total:       diskStat.Total,
		Free:        diskStat.Free,
		Used:        diskStat.Used,
		UsedPercent: diskStat.UsedPercent,
	}

	// Collect application metrics
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	metrics.Application = ApplicationMetrics{
		Goroutines: runtime.NumGoroutine(),
		GCStats: GCStats{
			NumGC:        m.NumGC,
			PauseTotalNs: m.PauseTotalNs,
			LastPauseNs:  m.LastGC,
			PauseNs:      []uint64{m.PauseNs[(m.NumGC+255)%256]}, // Last GC pause
		},
		Threads: runtime.GOMAXPROCS(0),
	}

	return metrics, nil
}

// LogPerformanceMetrics logs performance metrics
func (a *APMService) LogPerformanceMetrics(ctx context.Context) error {
	metrics, err := a.GetPerformanceMetrics(ctx)
	if err != nil {
		return fmt.Errorf("failed to get performance metrics: %w", err)
	}

	LogInfo("performance_metrics", "Application performance metrics collected", map[string]interface{}{
		"service":             a.serviceName,
		"uptime":              metrics.Uptime.Seconds(),
		"cpu_percent":         metrics.CPU.PercentUsed,
		"memory_used_percent": metrics.Memory.UsedPercent,
		"disk_used_percent":   metrics.Disk.UsedPercent,
		"goroutines":          metrics.Application.Goroutines,
		"timestamp":           metrics.Timestamp,
	})

	return nil
}

// HealthCheck performs a health check on the application
func (a *APMService) HealthCheck(ctx context.Context) (bool, map[string]interface{}) {
	health := make(map[string]interface{})

	// Check if we can collect metrics
	_, err := a.GetPerformanceMetrics(ctx)
	if err != nil {
		health["status"] = "unhealthy"
		health["error"] = err.Error()
		return false, health
	}

	// Check if CPU usage is too high
	if _, err := cpu.PercentWithContext(ctx, 0, false); err != nil {
		health["status"] = "degraded"
		health["cpu_error"] = err.Error()
		return true, health
	}

	// Check if memory usage is too high
	vmStat, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		health["status"] = "degraded"
		health["memory_error"] = err.Error()
		return true, health
	}

	if vmStat.UsedPercent > 90 {
		health["status"] = "degraded"
		health["memory_warning"] = fmt.Sprintf("High memory usage: %.2f%%", vmStat.UsedPercent)
	}

	health["status"] = "healthy"
	health["uptime"] = time.Since(a.startTime).String()
	health["timestamp"] = time.Now()

	return true, health
}

// GetApplicationMetrics returns application-specific metrics
func (a *APMService) GetApplicationMetrics() ApplicationMetrics {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return ApplicationMetrics{
		Goroutines: runtime.NumGoroutine(),
		GCStats: GCStats{
			NumGC:        m.NumGC,
			PauseTotalNs: m.PauseTotalNs,
			LastPauseNs:  m.LastGC,
			PauseNs:      []uint64{m.PauseNs[(m.NumGC+255)%256]},
		},
		Threads: runtime.GOMAXPROCS(0),
	}
}

// MonitorPerformance continuously monitors performance and logs warnings
func (a *APMService) MonitorPerformance(ctx context.Context, interval time.Duration, warningThresholds PerformanceThresholds) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			metrics, err := a.GetPerformanceMetrics(ctx)
			if err != nil {
				LogError("performance_monitor", "Failed to collect performance metrics", map[string]interface{}{
					"error": err.Error(),
				})
				continue
			}

			// Check thresholds and log warnings
			a.checkThresholds(metrics, warningThresholds)
		}
	}
}

// PerformanceThresholds defines warning thresholds for performance metrics
type PerformanceThresholds struct {
	CPUUsagePercent    float64
	MemoryUsagePercent float64
	DiskUsagePercent   float64
	GoroutinesCount    int
}

// checkThresholds checks if metrics exceed warning thresholds
func (a *APMService) checkThresholds(metrics *PerformanceMetrics, thresholds PerformanceThresholds) {
	if thresholds.CPUUsagePercent > 0 && metrics.CPU.PercentUsed > thresholds.CPUUsagePercent {
		LogWarn("performance_threshold", "CPU usage above threshold", map[string]interface{}{
			"current":   metrics.CPU.PercentUsed,
			"threshold": thresholds.CPUUsagePercent,
		})
	}

	if thresholds.MemoryUsagePercent > 0 && metrics.Memory.UsedPercent > thresholds.MemoryUsagePercent {
		LogWarn("performance_threshold", "Memory usage above threshold", map[string]interface{}{
			"current":   metrics.Memory.UsedPercent,
			"threshold": thresholds.MemoryUsagePercent,
		})
	}

	if thresholds.DiskUsagePercent > 0 && metrics.Disk.UsedPercent > thresholds.DiskUsagePercent {
		LogWarn("performance_threshold", "Disk usage above threshold", map[string]interface{}{
			"current":   metrics.Disk.UsedPercent,
			"threshold": thresholds.DiskUsagePercent,
		})
	}

	if thresholds.GoroutinesCount > 0 && metrics.Application.Goroutines > thresholds.GoroutinesCount {
		LogWarn("performance_threshold", "Goroutines count above threshold", map[string]interface{}{
			"current":   metrics.Application.Goroutines,
			"threshold": thresholds.GoroutinesCount,
		})
	}
}
