package bootstrap

import (
	"insight-engine-backend/routes"
	"insight-engine-backend/services"
	"insight-engine-backend/services/formula_engine"

	"github.com/gofiber/fiber/v2"
)

// App is the main application struct
type App struct {
	FiberApp   *fiber.App
	Services   *ServiceContainer
	Handlers   *routes.HandlerContainer // Use routes.HandlerContainer
	Middleware *routes.MiddlewareContainer
}

// ServiceContainer holds all initialized services
type ServiceContainer struct {
	EncryptionService        *services.EncryptionService
	AIService                *services.AIService
	AIReasoningService       *services.AIReasoningService
	AIOptimizerService       *services.AIOptimizerService
	StoryGeneratorService    *services.StoryGeneratorService
	PPTXGenerator            *services.PPTXGenerator // TASK-161
	SemanticLayerService     *services.SemanticLayerService
	ModelingService          *services.ModelingService
	RateLimiterService       *services.RateLimiter
	UsageTrackerService      *services.UsageTracker
	CronService              *services.CronService
	WebSocketHub             *services.WebSocketHub
	NotificationService      *services.NotificationService
	ActivityService          *services.ActivityService
	SchedulerService         *services.SchedulerService
	AuditService             *services.AuditService
	QueryExecutor            *services.QueryExecutor
	QueryQueueService        *services.QueryQueueService
	SchemaDiscovery          *services.SchemaDiscovery
	QueryValidator           *services.QueryValidator
	ReportingService         *services.ReportingService
	ForecastingService       *services.ForecastingService
	AnomalyDetectionService  *services.AnomalyDetectionService
	InsightsService          *services.InsightsService
	CorrelationService       *services.CorrelationService
	GlossaryService          *services.GlossaryService
	NLService                *services.NLService
	WebhookService           *services.WebhookService
	QueryCache               *services.QueryCache
	RLSService               *services.RLSService
	DataGovernanceService    *services.DataGovernanceService
	EngineService            *services.EngineService
	PaginationService        *services.PaginationService
	QueryBuilder             *services.QueryBuilder
	GeoJSONService           *services.GeoJSONService
	EmailService             *services.EmailService
	AuthService              *services.AuthService
	OAuthService             *services.OAuthService
	MaterializedViewService  *services.MaterializedViewService
	AlertNotificationService *services.AlertNotificationService
	AlertService             *services.AlertService
	OrganizationService      *services.OrganizationService
	EmbedService             *services.EmbedService
	CommentService           *services.CommentService
	ScheduledReportService   *services.ScheduledReportService
	SecurityLogService       *services.SecurityLogService
	SystemHealthService      *services.SystemHealthService
	FormulaEngine            *formula_engine.FormulaEngine
	RedisCache               *services.RedisCache
}

// NewApp initializes the entire application
func NewApp() *App {
	// 1. Initialize Logger
	InitLogger()

	// 2. Load Config
	LoadConfig()

	// 3. Connect DB
	ConnectDatabase()

	// 4. Initialize Services
	svcContainer := InitServices()

	// 5. Initialize Handlers
	hdlContainer := InitHandlers(svcContainer)

	// 6. Initialize Fiber App
	app := InitServer(svcContainer, hdlContainer)

	return &App{
		FiberApp: app,
		Services: svcContainer,
		Handlers: hdlContainer,
	}
}
