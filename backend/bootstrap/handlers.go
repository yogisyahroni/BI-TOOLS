package bootstrap

import (
	"insight-engine-backend/controllers"
	"insight-engine-backend/database"
	"insight-engine-backend/handlers"
)

// InitHandlers initializes all handlers
func InitHandlers(svc *ServiceContainer) *HandlerContainer {

	// Core Handlers
	aiHandler := handlers.NewAIHandler(svc.AIService, svc.AIReasoningService, svc.AIOptimizerService, svc.StoryGeneratorService)
	authHandler := handlers.NewAuthHandler(svc.AuthService)
	oauthHandler := handlers.NewOAuthHandler(svc.OAuthService)

	visualQueryHandler := handlers.NewVisualQueryHandler(database.DB, svc.QueryBuilder, svc.QueryExecutor, svc.SchemaDiscovery, svc.QueryCache)
	connectionHandler := handlers.NewConnectionHandler(svc.QueryExecutor, svc.SchemaDiscovery)
	queryHandler := handlers.NewQueryHandler(svc.QueryExecutor)
	queryAnalyzerHandler := handlers.NewQueryAnalyzerHandler(database.DB, svc.QueryExecutor)

	materializedViewHandler := handlers.NewMaterializedViewHandler(database.DB, svc.MaterializedViewService)
	engineHandler := handlers.NewEngineHandler(svc.EngineService)
	geoJSONHandler := handlers.NewGeoJSONHandler(svc.GeoJSONService)
	dataGovernanceHandler := handlers.NewDataGovernanceHandler(svc.DataGovernanceService)

	semanticLayerHandler := handlers.NewSemanticLayerHandler(svc.SemanticLayerService)
	modelingHandler := handlers.NewModelingHandler(svc.ModelingService)

	dashboardHandler := handlers.NewDashboardHandler()
	dashboardCardHandler := handlers.NewDashboardCardHandler()

	// Monitoring Handlers
	notificationHandler := handlers.NewNotificationHandler(svc.NotificationService)
	activityHandler := handlers.NewActivityHandler(svc.ActivityService)
	schedulerHandler := handlers.NewSchedulerHandler(svc.SchedulerService)
	wsHandler := handlers.NewWebSocketHandler(svc.WebSocketHub)

	// Additional Handlers
	commentHandler := handlers.NewCommentHandler(svc.CommentService)
	shareHandler := handlers.NewShareHandler(database.DB, svc.AuditService)
	embedHandler := handlers.NewEmbedHandler(svc.EmbedService)
	auditHandler := handlers.NewAuditHandler(svc.AuditService)

	frontendLogHandler := handlers.NewFrontendLogHandler(database.DB)
	rateLimitHandler := handlers.NewRateLimitHandler(svc.RateLimiterService)
	aiUsageHandler := handlers.NewAIUsageHandler(svc.UsageTrackerService)

	alertHandler := handlers.NewAlertHandler(svc.AlertService)
	alertNotificationHandler := handlers.NewAlertNotificationHandler(svc.AlertNotificationService)

	analyticsHandler := handlers.NewAnalyticsHandler(svc.InsightsService, svc.CorrelationService)

	// Admin Handlers
	adminOrgHandler := handlers.NewAdminOrganizationHandler(svc.OrganizationService)
	adminUserHandler := handlers.NewAdminUserHandler(database.DB, svc.AuditService)
	adminSystemHandler := handlers.NewAdminSystemHandler(database.DB)

	// Report & Analysis
	var reportHandler *handlers.ScheduledReportHandler
	if svc.ScheduledReportService != nil {
		reportHandler = handlers.NewScheduledReportHandler(svc.ScheduledReportService)
	}

	versionHandler := handlers.NewVersionHandler(database.DB, svc.NotificationService)
	queryVersionHandler := handlers.NewQueryVersionHandler(database.DB, svc.NotificationService)
	glossaryHandler := handlers.NewGlossaryHandler(svc.GlossaryService)
	nlHandler := handlers.NewNLHandler(svc.NLService)
	webhookHandler := handlers.NewWebhookHandler(database.DB)

	reportingHandler := handlers.NewReportingHandler(svc.ReportingService)
	forecastingHandler := handlers.NewForecastingHandler(svc.ForecastingService)
	anomalyHandler := handlers.NewAnomalyHandler(svc.AnomalyDetectionService)

	lineageController := controllers.NewLineageController()
	permissionHandler := handlers.NewPermissionHandler(database.DB)

	return &HandlerContainer{
		AIHandler:                aiHandler,
		AuthHandler:              authHandler,
		OAuthHandler:             oauthHandler,
		PermissionHandler:        permissionHandler,
		QueryHandler:             queryHandler,
		VisualQueryHandler:       visualQueryHandler,
		ConnectionHandler:        connectionHandler,
		QueryAnalyzerHandler:     queryAnalyzerHandler,
		MaterializedViewHandler:  materializedViewHandler,
		EngineHandler:            engineHandler,
		GeoJSONHandler:           geoJSONHandler,
		DataGovernanceHandler:    dataGovernanceHandler,
		SemanticLayerHandler:     semanticLayerHandler,
		ModelingHandler:          modelingHandler,
		DashboardHandler:         dashboardHandler,
		DashboardCardHandler:     dashboardCardHandler,
		NotificationHandler:      notificationHandler,
		ActivityHandler:          activityHandler,
		SchedulerHandler:         schedulerHandler,
		WebSocketHandler:         wsHandler,
		CommentHandler:           commentHandler,
		ShareHandler:             shareHandler,
		EmbedHandler:             embedHandler,
		AuditHandler:             auditHandler,
		FrontendLogHandler:       frontendLogHandler,
		RateLimitHandler:         rateLimitHandler,
		AIUsageHandler:           aiUsageHandler,
		AlertHandler:             alertHandler,
		AlertNotificationHandler: alertNotificationHandler,
		AnalyticsHandler:         analyticsHandler,
		AdminOrgHandler:          adminOrgHandler,
		AdminUserHandler:         adminUserHandler,
		AdminSystemHandler:       adminSystemHandler,
		ScheduledReportHandler:   reportHandler,
		VersionHandler:           versionHandler,
		QueryVersionHandler:      queryVersionHandler,
		GlossaryHandler:          glossaryHandler,
		NLHandler:                nlHandler,
		WebhookHandler:           webhookHandler,
		ReportingHandler:         reportingHandler,
		ForecastingHandler:       forecastingHandler,
		AnomalyHandler:           anomalyHandler,
		LineageController:        lineageController,
		CollectionHandler:        handlers.NewCollectionHandler(),
	}
}
