package bootstrap

import (
	"insight-engine-backend/database"
	"insight-engine-backend/pkg/resilience"
	"insight-engine-backend/services"
	"os"
	"time"
)

// InitServices initializes all services
func InitServices() *ServiceContainer {
	// 1. Encryption Service
	encryptionService, err := services.NewEncryptionService()
	if err != nil {
		services.LogFatal("encryption_init", "Failed to initialize encryption service. Set ENCRYPTION_KEY environment variable (32 bytes). Generate with: openssl rand -base64 32", map[string]interface{}{"error": err})
	}
	services.LogInfo("encryption_init", "Encryption service initialized successfully", nil)

	// 2. AI Services
	aiService := services.NewAIService(encryptionService)
	aiReasoningService := services.NewAIReasoningService(aiService)
	aiOptimizerService := services.NewAIOptimizerService(aiService)
	storyGeneratorService := services.NewStoryGeneratorService(aiService)

	// 3. Database-dependent Services
	rateLimiterService := services.NewRateLimiter(database.DB)
	usageTrackerService := services.NewUsageTracker(database.DB)
	wsHub := services.NewWebSocketHub()
	go wsHub.Run() // Start WS Hub immediately

	notificationService := services.NewNotificationService(database.DB, wsHub)
	activityService := services.NewActivityService(database.DB, wsHub)

	// Cron & Scheduler
	cronService := services.NewCronService(database.DB)
	cronService.Start()

	schedulerService := services.NewSchedulerService(database.DB)
	schedulerService.Start()

	auditService := services.NewAuditService(database.DB)

	// Job Queue
	services.InitJobQueue(5)

	// Core Query Architecture
	cbConfig := resilience.CircuitBreakerConfig{
		Name:        "query-executor",
		MaxRequests: 5,
		Interval:    60 * time.Second,
		Timeout:     30 * time.Second,
	}
	circuitBreaker := resilience.NewCircuitBreaker(cbConfig)
	queryExecutor := services.NewQueryExecutor(circuitBreaker)
	queryQueueService := services.NewQueryQueueService(queryExecutor, 10)
	schemaDiscovery := services.NewSchemaDiscovery(queryExecutor)
	queryValidator := services.NewQueryValidator([]string{})

	// Redis
	redisConfig := services.RedisCacheConfig{
		Host:       os.Getenv("REDIS_HOST"),
		Password:   os.Getenv("REDIS_PASSWORD"),
		DB:         0,
		MaxRetries: 3,
		PoolSize:   10,
	}
	if redisConfig.Host == "" {
		redisConfig.Host = "localhost:6379"
	}

	var queryCache *services.QueryCache
	if redisCache, err := services.NewRedisCache(redisConfig); err == nil {
		queryCache = services.NewQueryCache(redisCache, 5*time.Minute)
	}

	// Business Services
	rlsService := services.NewRLSService(database.DB)
	dataGovernanceService := services.NewDataGovernanceService(database.DB)
	engineService := services.NewEngineService(queryExecutor)
	paginationService := services.NewPaginationService()

	queryBuilder := services.NewQueryBuilder(queryValidator, schemaDiscovery, queryCache, rlsService, paginationService, queryQueueService)
	geoJSONService := services.NewGeoJSONService(database.DB)

	// Auth
	emailService := services.NewEmailService()
	authService := services.NewAuthService(database.DB, emailService)
	oauthService := services.NewOAuthService(database.DB)

	// Additional Features
	materializedViewService := services.NewMaterializedViewService(database.DB, queryExecutor)
	reportingService := services.NewReportingService()
	forecastingService := services.NewForecastingService()
	anomalyDetectionService := services.NewAnomalyDetectionService()
	insightsService := services.NewInsightsService()
	correlationService := services.NewCorrelationService()
	glossaryService := services.NewGlossaryService(database.DB)
	nlService := services.NewNLService(database.DB, aiService)
	webhookService := services.NewWebhookService(database.DB)

	embedService := services.NewEmbedService(database.DB)
	commentService := services.NewCommentService(database.DB, notificationService)

	semanticLayerService := services.NewSemanticLayerService(database.DB)
	modelingService := services.NewModelingService(database.DB)

	// Alerts
	baseURL := os.Getenv("APP_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:3000"
	}
	alertNotificationService := services.NewAlertNotificationService(database.DB, emailService, notificationService, baseURL)
	alertService := services.NewAlertService(database.DB, queryExecutor, alertNotificationService)
	organizationService := services.NewOrganizationService(database.DB, auditService)

	// Scheduled Reports
	scheduledReportService, err := services.NewScheduledReportService(database.DB, emailService, "./exports", baseURL)
	if err != nil {
		services.LogWarn("scheduled_report_init", "Failed to initialize scheduled report service", map[string]interface{}{"error": err})
	}

	return &ServiceContainer{
		EncryptionService:        encryptionService,
		AIService:                aiService,
		AIReasoningService:       aiReasoningService,
		AIOptimizerService:       aiOptimizerService,
		StoryGeneratorService:    storyGeneratorService,
		SemanticLayerService:     semanticLayerService,
		ModelingService:          modelingService,
		RateLimiterService:       rateLimiterService,
		UsageTrackerService:      usageTrackerService,
		CronService:              cronService,
		WebSocketHub:             wsHub,
		NotificationService:      notificationService,
		ActivityService:          activityService,
		SchedulerService:         schedulerService,
		AuditService:             auditService,
		QueryExecutor:            queryExecutor,
		QueryQueueService:        queryQueueService,
		SchemaDiscovery:          schemaDiscovery,
		QueryValidator:           queryValidator,
		ReportingService:         reportingService,
		ForecastingService:       forecastingService,
		AnomalyDetectionService:  anomalyDetectionService,
		InsightsService:          insightsService,
		CorrelationService:       correlationService,
		GlossaryService:          glossaryService,
		NLService:                nlService,
		WebhookService:           webhookService,
		QueryCache:               queryCache,
		RLSService:               rlsService,
		DataGovernanceService:    dataGovernanceService,
		EngineService:            engineService,
		PaginationService:        paginationService,
		QueryBuilder:             queryBuilder,
		GeoJSONService:           geoJSONService,
		EmailService:             emailService,
		AuthService:              authService,
		OAuthService:             oauthService,
		MaterializedViewService:  materializedViewService,
		AlertNotificationService: alertNotificationService,
		AlertService:             alertService,
		OrganizationService:      organizationService,
		EmbedService:             embedService,
		CommentService:           commentService,
		ScheduledReportService:   scheduledReportService,
	}
}
