package bootstrap

import (
	"context"
	"os"
	"time"

	"insight-engine-backend/handlers"
	"insight-engine-backend/middleware"
	"insight-engine-backend/routes"
	"insight-engine-backend/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger"
	// "insight-engine-backend/docs" // Import docs for Swagger
)

// InitServer initializes the Fiber app
func InitServer(svc *ServiceContainer, h *HandlerContainer) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: "InsightEngine Backend (Go)",
	})

	// Middleware
	app.Use(logger.New())
	app.Use(middleware.RecoveryMiddleware)

	// Request ID (TASK-180: generates UUID v4 per request for log/trace correlation)
	app.Use(middleware.RequestIDMiddleware())

	// SSL/TLS Enforcement
	sslConfig := middleware.LoadSSLConfigFromEnv()
	app.Use(middleware.SSLRedirect(sslConfig))

	// SQL Injection Protection
	app.Use(middleware.SQLInjectionMiddleware())

	// Privacy Compliance
	privacyConfig := middleware.CreateDefaultPrivacyConfig()
	app.Use(middleware.PrivacyComplianceMiddleware(privacyConfig))

	// API Governance
	apiVersionConfig := middleware.CreateDefaultAPIVersionConfig()
	app.Use(middleware.APIVersionMiddleware(apiVersionConfig))

	// Observability - Distributed Tracing
	tracingConfig := middleware.CreateDefaultTracingConfig()
	if tracerProvider, err := middleware.InitTracer(tracingConfig); err != nil {
		services.LogWarn("tracing_init", "Failed to initialize tracer", map[string]interface{}{"error": err})
	} else {
		// Ensure tracer is shut down properly
		defer func() {
			if tracerProvider != nil {
				tracerProvider.Shutdown(context.Background())
			}
		}()
	}
	app.Use(middleware.DistributedTracingMiddleware(tracingConfig))

	// Observability - Enhanced Metrics
	app.Use(middleware.EnhancedMetricsMiddleware())

	// Observability - Performance Monitoring
	app.Use(middleware.PerformanceMonitoringMiddleware(5 * time.Second)) // 5 second threshold

	// Observability - Error Tracking
	app.Use(middleware.ErrorTrackingMiddleware())

	// Swagger
	app.Get("/swagger/*", swagger.HandlerDefault)

	// CORS
	corsConfig := middleware.LoadCORSConfigFromEnv()
	app.Use(middleware.HardenedCORS(corsConfig))

	// Mass Assignment Protection (BOPLA)
	boplaConfig := middleware.CreateDefaultBoplaConfig()
	app.Use(middleware.MassAssignmentProtection(boplaConfig))

	// Rate Limiting
	comprehensiveRateLimit := middleware.ComprehensiveRateLimit(middleware.ComprehensiveRateLimitConfig{
		RateLimiterService: svc.RateLimiterService,
		SkipPaths: map[string]bool{
			"/api/health":       true,
			"/api/health/ready": true,
			"/api/health/live":  true,
			"/api/metrics":      true,
		},
	})

	// Metrics
	services.InitMetrics()
	app.Use(middleware.MetricsMiddleware)

	// Degradation Middleware
	app.Use(middleware.DegradationMiddleware(middleware.DegradationConfig{
		QueryExecutor: svc.QueryExecutor,
	}))

	// Handlers & Middleware Containers
	// IMPORTANT: Reconstruct HandlerContainer to match routes.HandlerContainer
	// The bootstrap HandlerContainer fields must map 1:1 to routes package
	routeHandlers := &routes.HandlerContainer{
		AIHandler:                h.AIHandler,
		AuthHandler:              h.AuthHandler,
		OAuthHandler:             h.OAuthHandler,
		PermissionHandler:        h.PermissionHandler,
		QueryHandler:             h.QueryHandler,
		VisualQueryHandler:       h.VisualQueryHandler,
		ConnectionHandler:        h.ConnectionHandler,
		QueryAnalyzerHandler:     h.QueryAnalyzerHandler,
		MaterializedViewHandler:  h.MaterializedViewHandler,
		EngineHandler:            h.EngineHandler,
		GeoJSONHandler:           h.GeoJSONHandler,
		DataGovernanceHandler:    h.DataGovernanceHandler,
		SemanticLayerHandler:     h.SemanticLayerHandler,
		ModelingHandler:          h.ModelingHandler,
		DashboardHandler:         h.DashboardHandler,
		DashboardCardHandler:     h.DashboardCardHandler,
		NotificationHandler:      h.NotificationHandler,
		ActivityHandler:          h.ActivityHandler,
		SchedulerHandler:         h.SchedulerHandler,
		WebSocketHandler:         h.WebSocketHandler,
		CommentHandler:           h.CommentHandler,
		ShareHandler:             h.ShareHandler,
		EmbedHandler:             h.EmbedHandler,
		AuditHandler:             h.AuditHandler,
		FrontendLogHandler:       h.FrontendLogHandler,
		RateLimitHandler:         h.RateLimitHandler,
		AIUsageHandler:           h.AIUsageHandler,
		AlertHandler:             h.AlertHandler,
		AlertNotificationHandler: h.AlertNotificationHandler,
		AnalyticsHandler:         h.AnalyticsHandler,
		AdminOrgHandler:          h.AdminOrgHandler,
		AdminUserHandler:         h.AdminUserHandler,
		AdminSystemHandler:       h.AdminSystemHandler,
		ScheduledReportHandler:   h.ScheduledReportHandler,
		VersionHandler:           h.VersionHandler,
		QueryVersionHandler:      h.QueryVersionHandler,
		GlossaryHandler:          h.GlossaryHandler,
		NLHandler:                h.NLHandler,
		WebhookHandler:           h.WebhookHandler,
		CollectionHandler:        h.CollectionHandler,
		// Reporting, Forecasting, Anomaly handlers were in main.go but need check if they are in routes.HandlerContainer
		// They were routed manually in main.go
	}

	middlewareContainer := &routes.MiddlewareContainer{
		AuthMiddleware:            middleware.AuthMiddleware,
		AdminMiddleware:           middleware.RequireAdmin,
		RateLimitMiddleware:       comprehensiveRateLimit,
		AdaptiveTimeoutMiddleware: middleware.AdaptiveTimeoutMiddleware(),
	}

	// Setup Standard Routes
	routes.SetupRoutes(app, routeHandlers, middlewareContainer)

	// Setup Custom Manual Routes (from main.go legacy)
	api := app.Group("/api")
	api.Get("/metrics", handlers.MetricsHandler)
	api.Post("/forecast", h.ForecastingHandler.Forecast)
	api.Post("/reports/generate", h.ReportingHandler.GenerateReport)
	api.Post("/analytics/anomalies", h.AnomalyHandler.DetectAnomalies)

	// Lineage Route
	api.Get("/lineage", h.LineageController.GetLineage)

	return app
}

// RunServer runs the application
func (app *App) Run() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	services.LogInfo("server_start", "Server running", map[string]interface{}{"port": port})
	if err := app.FiberApp.Listen(":" + port); err != nil {
		services.LogFatal("server_start_failed", "Failed to start server", map[string]interface{}{"port": port, "error": err})
	}
}
