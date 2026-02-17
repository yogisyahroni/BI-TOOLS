package main

import (
	"context"
	"time"
	"insight-engine-backend/bootstrap"
	"insight-engine-backend/services"
)

func main() {
	// 0. Initialize Tracing
	shutdown := bootstrap.InitTracing()
	defer shutdown()

	// 1. Initialize Security Logger First
	// Security logging must be available from application start
	services.InitSecurityLogger("InsightEngine Backend (Go)")

	// 2. Initialize APM Service
	apmService := services.NewAPMService("InsightEngine Backend (Go)")
	
	// Start performance monitoring in background
	ctx := context.Background()
	go apmService.MonitorPerformance(ctx, 30*time.Second, services.PerformanceThresholds{
		CPUUsagePercent:    80.0,
		MemoryUsagePercent: 85.0,
		DiskUsagePercent:   90.0,
		GoroutinesCount:    1000,
	})

	// 3. Initialize Bootstrap
	// This is the single entry point for the entire application lifecycle
	app := bootstrap.NewApp()

	// 4. Run Application
	// This handles the server listen and graceful shutdown logic
	app.Run()

	services.LogInfo("main", "Application exited cleanly", nil)
}
