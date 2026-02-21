package routes

import (
	"insight-engine-backend/controllers"
	"insight-engine-backend/handlers"

	"github.com/gofiber/fiber/v2"
)

// HandlerContainer holds all application handlers to be passed to the route setup function.
// This avoids having functions with excessive arguments.
type HandlerContainer struct {
	// Auth Handlers
	AuthHandler       *handlers.AuthHandler
	OAuthHandler      *handlers.OAuthHandler
	PermissionHandler *handlers.PermissionHandler

	// Core Feature Handlers
	QueryHandler            *handlers.QueryHandler
	VisualQueryHandler      *handlers.VisualQueryHandler
	ConnectionHandler       *handlers.ConnectionHandler
	QueryAnalyzerHandler    *handlers.QueryAnalyzerHandler
	MaterializedViewHandler *handlers.MaterializedViewHandler
	EngineHandler           *handlers.EngineHandler
	GeoJSONHandler          *handlers.GeoJSONHandler
	DataGovernanceHandler   *handlers.DataGovernanceHandler
	SemanticLayerHandler    *handlers.SemanticLayerHandler
	ModelingHandler         *handlers.ModelingHandler
	FormulaHandler          *handlers.FormulaHandler // GAP-004

	// Real-time & Collaboration Handlers
	NotificationHandler  *handlers.NotificationHandler
	DashboardHandler     *handlers.DashboardHandler
	DashboardCardHandler *handlers.DashboardCardHandler
	ActivityHandler      *handlers.ActivityHandler
	SchedulerHandler     *handlers.SchedulerHandler
	WebSocketHandler     *handlers.WebSocketHandler
	CommentHandler       *handlers.CommentHandler
	ShareHandler         *handlers.ShareHandler
	EmbedHandler         *handlers.EmbedHandler

	// Monitoring & Logging Handlers
	AuditHandler             *handlers.AuditHandler
	FrontendLogHandler       *handlers.FrontendLogHandler
	RateLimitHandler         *handlers.RateLimitHandler
	AIHandler                *handlers.AIHandler
	StoryHandler             *handlers.StoryHandler // TASK-161
	AIUsageHandler           *handlers.AIUsageHandler
	PulseHandler             *handlers.PulseHandler // TASK-156
	AlertHandler             *handlers.AlertHandler
	AlertNotificationHandler *handlers.AlertNotificationHandler
	AnalyticsHandler         *handlers.AnalyticsHandler
	GlossaryHandler          *handlers.GlossaryHandler // TASK-125
	NLHandler                *handlers.NLHandler       // TASK-120-122
	WebhookHandler           *handlers.WebhookHandler

	// Analytics & ML Handlers
	ReportingHandler   *handlers.ReportingHandler
	ForecastingHandler *handlers.ForecastingHandler
	AnomalyHandler     *handlers.AnomalyHandler

	// Admin Handlers
	AdminOrgHandler    *handlers.AdminOrganizationHandler
	AdminUserHandler   *handlers.AdminUserHandler
	AdminSystemHandler *handlers.AdminSystemHandler

	// Optional Handlers (may be nil if init failed)
	ScheduledReportHandler *handlers.ScheduledReportHandler

	// Versioning Handlers
	VersionHandler      *handlers.VersionHandler
	QueryVersionHandler *handlers.QueryVersionHandler

	// Collection Handler
	CollectionHandler *handlers.CollectionHandler

	// Lineage Controller
	LineageController *controllers.LineageController

	// System Health Handler
	SystemHealthHandler *handlers.SystemHealthHandler
}

// MiddlewareContainer holds middleware functions
type MiddlewareContainer struct {
	AuthMiddleware            fiber.Handler
	AdminMiddleware           fiber.Handler
	RateLimitMiddleware       fiber.Handler
	AdaptiveTimeoutMiddleware fiber.Handler
	CacheMiddleware           fiber.Handler
}
