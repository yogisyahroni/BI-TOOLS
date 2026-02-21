package bootstrap

import (
	"fmt"
	"insight-engine-backend/database"
	"insight-engine-backend/pkg/resilience"
	"insight-engine-backend/services"
	"insight-engine-backend/services/formula_engine"
	"os"
	"time"
)

// InitServices initializes all services
func InitServices() *ServiceContainer {
	// ... (imports) ...
	// 1. Encryption Service
	// 1. Encryption Service
	encryptionService, err := services.NewEncryptionService()
	if err != nil {
		services.LogFatal("encryption_init", "Failed to initialize encryption service. Set ENCRYPTION_KEY environment variable (32 bytes). Generate with: openssl rand -base64 32", map[string]interface{}{"error": err})
	}
	services.LogInfo("encryption_init", "Encryption service initialized successfully", nil)

	// Initialize generic services
	embeddingService := services.NewEmbeddingService(encryptionService)
	semanticLayerService := services.NewSemanticLayerService(database.DB) // Moved up for AI Service dependency
	aiService := services.NewAIService(encryptionService, embeddingService, semanticLayerService)
	aiReasoningService := services.NewAIReasoningService(aiService)
	aiOptimizerService := services.NewAIOptimizerService(aiService)
	storyGeneratorService := services.NewStoryGeneratorService(aiService)

	// 2. Security Log Service
	securityLogService := services.NewSecurityLogService("InsightEngine Backend (Go)")

	// 3. Database-dependent Services
	rateLimiterService := services.NewRateLimiter(database.DB)
	usageTrackerService := services.NewUsageTracker(database.DB)
	wsHub := services.NewWebSocketHub()
	go wsHub.Run() // Start WS Hub immediately

	// Slack Service (TASK-156)
	slackService := services.NewSlackService()

	notificationService := services.NewNotificationService(database.DB, wsHub, slackService)
	activityService := services.NewActivityService(database.DB, wsHub)

	// Pulse Service (TASK-156)
	fmt.Println("DEBUG: Initializing ScreenshotService")
	screenshotService := services.NewScreenshotService()
	fmt.Println("DEBUG: Initializing PulseService")
	pulseService := services.NewPulseService(database.DB, screenshotService, slackService)
	// Set admin token for screenshot service - ideally from ENV or internal auth
	pulseService.SetAdminToken(os.Getenv("INTERNAL_ADMIN_TOKEN"))
	fmt.Println("DEBUG: PulseService initialized")

	// Cron & Scheduler
	cronService := services.NewCronService(database.DB)
	cronService.SetPulseService(pulseService) // Inject PulseService
	cronService.Start()

	schedulerService := services.NewSchedulerService(database.DB)
	schedulerService.Start()

	auditService := services.NewAuditService(database.DB)

	// Job Queue
	services.InitJobQueue(5)

	// Redis (Moved up for QueryExecutor dependency)
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
	var redisCache *services.RedisCache
	if redisCache, err = services.NewRedisCache(redisConfig); err == nil {
		queryCache = services.NewQueryCache(redisCache, 5*time.Minute)
	} else {
		// Log error but continue (fallback to nil or handle gracefully if services allow nil)
		// QueryCache handles nil RedisCache? NewQueryCache might need valid RedisCache.
		// Let's check NewQueryCache. If it fails, queryCache is nil.
		services.LogWarn("redis_init", "Failed to initialize Redis cache", map[string]interface{}{"error": err})
	}

	// Core Query Architecture
	cbConfig := resilience.CircuitBreakerConfig{
		Name:        "query-executor",
		MaxRequests: 5,
		Interval:    60 * time.Second,
		Timeout:     30 * time.Second,
	}
	circuitBreaker := resilience.NewCircuitBreaker(cbConfig)
	queryOptimizer := services.NewQueryOptimizer()
	queryExecutor := services.NewQueryExecutor(circuitBreaker, queryOptimizer, queryCache)
	queryQueueService := services.NewQueryQueueService(queryExecutor, 10)
	schemaDiscovery := services.NewSchemaDiscovery(queryExecutor)
	queryValidator := services.NewQueryValidator([]string{})

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

	pptxGenerator := services.NewPPTXGenerator() // TASK-161

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

	// System Health (GAP-003)
	systemHealthService := services.NewSystemHealthService(database.DB, redisCache)

	// Formula Engine (GAP-004)
	formulaEngine := formula_engine.NewFormulaEngine()

	return &ServiceContainer{
		EncryptionService:  encryptionService,
		EmbeddingService:   embeddingService,
		AIService:          aiService,
		AIReasoningService: aiReasoningService,
		AIOptimizerService: aiOptimizerService,

		StoryGeneratorService:    storyGeneratorService,
		PPTXGenerator:            pptxGenerator, // TASK-161
		SemanticLayerService:     semanticLayerService,
		ModelingService:          modelingService,
		RateLimiterService:       rateLimiterService,
		UsageTrackerService:      usageTrackerService,
		CronService:              cronService,
		SchedulerService:         schedulerService,
		WebSocketHub:             wsHub,
		NotificationService:      notificationService,
		SlackService:             slackService,
		ActivityService:          activityService,
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

		ScheduledReportService: scheduledReportService,
		SecurityLogService:     securityLogService,
		SystemHealthService:    systemHealthService,
		PulseService:           pulseService,
		ScreenshotService:      screenshotService,
		// ...
		FormulaEngine: formulaEngine, // GAP-004
		RedisCache:    redisCache,
	}
}
