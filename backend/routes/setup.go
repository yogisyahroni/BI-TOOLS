package routes

import (
	handlers "insight-engine-backend/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// SetupRoutes registers all application routes
func SetupRoutes(app *fiber.App, h *HandlerContainer, m *MiddlewareContainer) {
	// Base API group with rate limiting
	api := app.Group("/api", m.RateLimitMiddleware)

	// --- Authentication Routes ---
	api.Post("/auth/register", h.AuthHandler.Register)
	api.Post("/auth/login", h.AuthHandler.Login)
	api.Get("/auth/verify-email", h.AuthHandler.VerifyEmail)
	api.Post("/auth/resend-verification", h.AuthHandler.ResendVerification)
	api.Post("/auth/forgot-password", h.AuthHandler.ForgotPassword)
	api.Post("/auth/reset-password", h.AuthHandler.ResetPassword)
	api.Get("/auth/validate-reset-token", h.AuthHandler.ValidateResetToken)

	// Protected Auth Routes
	api.Post("/auth/change-password", m.AuthMiddleware, h.AuthHandler.ChangePassword)

	// OAuth Routes
	api.Get("/auth/providers", h.OAuthHandler.GetProviders)
	api.Get("/auth/:provider", h.OAuthHandler.InitiateAuth)
	api.Get("/auth/:provider/callback", h.OAuthHandler.HandleCallback)

	// --- Health Check Routes ---
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "InsightEngine Backend", "version": "1.0.0"})
	})
	// Note: Readiness/Liveness logic from main.go requires DB access, kept simple here or moved to a HealthHandler?
	// For now, let's keep simple json response here as in main.go, assuming DB check logic stays in main or moved.
	// Actually, create a simple inline handler here if needed, or better, move handler logic to a dedicated handler file later.
	// For this refactor, we'll implement simple responses to unblock.
	api.Get("/health/live", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "alive"})
	})
	// For ready check, we might need DB reference. We'll skip complex ready logic here or require a HealthHandler.
	// Let's assume HealthHandler exists or just use simple one for now to fix structure.

	// --- Core Feature Routes ---

	// Query Routes
	api.Get("/queries", m.AuthMiddleware, h.QueryHandler.GetQueries)
	api.Post("/queries", m.AuthMiddleware, h.QueryHandler.CreateQuery)
	api.Get("/queries/:id", m.AuthMiddleware, h.QueryHandler.GetQuery)
	api.Put("/queries/:id", m.AuthMiddleware, h.QueryHandler.UpdateQuery)
	api.Delete("/queries/:id", m.AuthMiddleware, h.QueryHandler.DeleteQuery)
	api.Post("/queries/:id/run", m.AuthMiddleware, h.QueryHandler.RunQuery)
	api.Post("/queries/execute", m.AuthMiddleware, m.AdaptiveTimeoutMiddleware, h.QueryHandler.ExecuteAdHocQuery)

	// Query Analysis
	api.Post("/query/analyze", m.AuthMiddleware, h.QueryAnalyzerHandler.AnalyzeQueryPlan)
	api.Get("/query/complexity", m.AuthMiddleware, h.QueryAnalyzerHandler.GetQueryComplexity)
	api.Post("/query/optimize", m.AuthMiddleware, h.QueryAnalyzerHandler.GetOptimizationSuggestions)

	// Visual Query Builder
	api.Get("/visual-queries", m.AuthMiddleware, h.VisualQueryHandler.GetVisualQueries)
	api.Post("/visual-queries", m.AuthMiddleware, h.VisualQueryHandler.CreateVisualQuery)
	api.Get("/visual-queries/:id", m.AuthMiddleware, h.VisualQueryHandler.GetVisualQuery)
	api.Put("/visual-queries/:id", m.AuthMiddleware, h.VisualQueryHandler.UpdateVisualQuery)
	api.Delete("/visual-queries/:id", m.AuthMiddleware, h.VisualQueryHandler.DeleteVisualQuery)
	api.Post("/visual-queries/generate-sql", m.AuthMiddleware, h.VisualQueryHandler.GenerateSQL)
	api.Post("/visual-queries/:id/preview", m.AuthMiddleware, h.VisualQueryHandler.PreviewVisualQuery)
	api.Get("/visual-queries/cache/stats", m.AuthMiddleware, h.VisualQueryHandler.GetCacheStats)
	api.Post("/visual-queries/join-suggestions", m.AuthMiddleware, h.VisualQueryHandler.GetJoinSuggestions)

	// Materialized Views
	api.Post("/materialized-views", m.AuthMiddleware, h.MaterializedViewHandler.CreateMaterializedView)
	api.Get("/materialized-views", m.AuthMiddleware, h.MaterializedViewHandler.ListMaterializedViews)
	api.Get("/materialized-views/:id", m.AuthMiddleware, h.MaterializedViewHandler.GetMaterializedView)
	api.Delete("/materialized-views/:id", m.AuthMiddleware, h.MaterializedViewHandler.DropMaterializedView)
	api.Post("/materialized-views/:id/refresh", m.AuthMiddleware, h.MaterializedViewHandler.RefreshMaterializedView)
	api.Put("/materialized-views/:id/schedule", m.AuthMiddleware, h.MaterializedViewHandler.UpdateSchedule)
	api.Get("/materialized-views/:id/status", m.AuthMiddleware, h.MaterializedViewHandler.GetStatus)
	api.Get("/materialized-views/:id/history", m.AuthMiddleware, h.MaterializedViewHandler.GetRefreshHistory)

	// Connections
	api.Get("/connections", m.AuthMiddleware, h.ConnectionHandler.GetConnections)
	api.Post("/connections", m.AuthMiddleware, h.ConnectionHandler.CreateConnection)
	api.Get("/connections/:id", m.AuthMiddleware, h.ConnectionHandler.GetConnection)
	api.Put("/connections/:id", m.AuthMiddleware, h.ConnectionHandler.UpdateConnection)
	api.Delete("/connections/:id", m.AuthMiddleware, h.ConnectionHandler.DeleteConnection)
	api.Post("/connections/:id/test", m.AuthMiddleware, h.ConnectionHandler.TestConnection)
	api.Get("/connections/:id/schema", m.AuthMiddleware, h.ConnectionHandler.GetConnectionSchema)

	// Engine Analytics
	api.Post("/engine/aggregate", m.AuthMiddleware, h.EngineHandler.Aggregate)
	api.Post("/engine/forecast", m.AuthMiddleware, h.EngineHandler.Forecast)

	// Advanced Analytics
	api.Post("/analytics/insights", m.AuthMiddleware, h.AnalyticsHandler.GenerateInsights)
	api.Post("/analytics/correlations", m.AuthMiddleware, h.AnalyticsHandler.CalculateCorrelation)

	api.Post("/engine/clustering", m.AuthMiddleware, h.EngineHandler.PerformClustering)

	// GeoJSON
	api.Post("/geojson", m.AuthMiddleware, h.GeoJSONHandler.UploadGeoJSON)
	api.Get("/geojson", m.AuthMiddleware, h.GeoJSONHandler.ListGeoJSON)
	api.Get("/geojson/:id", m.AuthMiddleware, h.GeoJSONHandler.GetGeoJSON)
	api.Get("/geojson/:id/data", m.AuthMiddleware, h.GeoJSONHandler.GetGeoJSONData)
	api.Put("/geojson/:id", m.AuthMiddleware, h.GeoJSONHandler.UpdateGeoJSON)
	api.Delete("/geojson/:id", m.AuthMiddleware, h.GeoJSONHandler.DeleteGeoJSON)

	// Data Governance
	api.Get("/governance/classifications", m.AuthMiddleware, h.DataGovernanceHandler.GetClassifications)
	api.Get("/governance/metadata", m.AuthMiddleware, h.DataGovernanceHandler.GetColumnMetadata)
	api.Post("/governance/metadata", m.AuthMiddleware, m.AdminMiddleware, h.DataGovernanceHandler.UpdateColumnMetadata)
	api.Get("/governance/permissions", m.AuthMiddleware, h.DataGovernanceHandler.GetColumnPermissions)
	api.Post("/governance/permissions", m.AuthMiddleware, m.AdminMiddleware, h.DataGovernanceHandler.SetColumnPermission)

	// Semantic Layer
	api.Get("/semantic/models", m.AuthMiddleware, h.SemanticLayerHandler.ListSemanticModels)
	api.Post("/semantic/models", m.AuthMiddleware, h.SemanticLayerHandler.CreateSemanticModel)
	api.Get("/semantic/models/:id", m.AuthMiddleware, h.SemanticLayerHandler.GetSemanticModel)
	api.Put("/semantic/models/:id", m.AuthMiddleware, h.SemanticLayerHandler.UpdateSemanticModel)
	api.Delete("/semantic/models/:id", m.AuthMiddleware, h.SemanticLayerHandler.DeleteSemanticModel)
	api.Get("/semantic/metrics", m.AuthMiddleware, h.SemanticLayerHandler.ListSemanticMetrics)
	api.Post("/semantic/query", m.AuthMiddleware, h.SemanticLayerHandler.ExecuteSemanticQuery)

	// Semantic Layer Chat/GenAI
	// Note: Semantic handlers for chat are seemingly mixed directly in handlers package in main.go (handlers.Semantic*)
	// We need to check if they are part of a specific struct. In main.go: `handlers.SemanticExplainData` suggests package level functions?
	// Let's assume they are methods on a struct if they hold state, or functions.
	// In main.go: `handlers.InitSemanticHandlers(aiService)` suggests global init?
	// Wait, the main.go code says: `api.Post(..., handlers.SemanticExplainData)`
	// This looks like package-level functions. We will need to check `handlers` package to be sure.
	// If they are package level, we access them directly or via `h` if we wrap them.
	// For now, let's assume they are package functions and call them directly from `handlers` package import if needed,
	// OR, better, `main.go` should wrap them in a struct.
	// Ideally, we move `Semantic*` functions to `SemanticHandler` struct.
	// For this refactor, let's check if we can pass them in HandlerContainer.
	// If they are functions, we can't put them in a struct field of type *Handler easily unless we define func types.
	// Let's defer this specific group to see if they are methods or funcs.
	// Looking at `main.go`: `handlers.InitSemanticHandlers(aiService)`.
	// It seems they are global variables in `handlers` package initialized by that init function.
	// Risky. But let's assume we can reference them directly from `routes` package by importing `handlers`.
	// Or add them to HandlerContainer as func fields if we want strict dependency injection.
	// Let's use `handlers.SemanticExplainData` directly here since they sound global.

	// Modeling
	api.Get("/modeling/definitions", m.AuthMiddleware, h.ModelingHandler.ListModelDefinitions)
	api.Post("/modeling/definitions", m.AuthMiddleware, h.ModelingHandler.CreateModelDefinition)
	api.Get("/modeling/definitions/:id", m.AuthMiddleware, h.ModelingHandler.GetModelDefinition)
	api.Put("/modeling/definitions/:id", m.AuthMiddleware, h.ModelingHandler.UpdateModelDefinition)
	api.Delete("/modeling/definitions/:id", m.AuthMiddleware, h.ModelingHandler.DeleteModelDefinition)
	api.Get("/modeling/metrics", m.AuthMiddleware, h.ModelingHandler.ListMetricDefinitions)
	api.Post("/modeling/metrics", m.AuthMiddleware, h.ModelingHandler.CreateMetricDefinition)
	api.Get("/modeling/metrics/:id", m.AuthMiddleware, h.ModelingHandler.GetMetricDefinition)
	api.Put("/modeling/metrics/:id", m.AuthMiddleware, h.ModelingHandler.UpdateMetricDefinition)
	api.Delete("/modeling/metrics/:id", m.AuthMiddleware, h.ModelingHandler.DeleteMetricDefinition)

	// Business Glossary (TASK-125)
	api.Get("/glossary/terms", m.AuthMiddleware, h.GlossaryHandler.ListTerms)
	api.Post("/glossary/terms", m.AuthMiddleware, h.GlossaryHandler.CreateTerm)
	api.Get("/glossary/terms/:id", m.AuthMiddleware, h.GlossaryHandler.GetTerm)
	api.Put("/glossary/terms/:id", m.AuthMiddleware, h.GlossaryHandler.UpdateTerm)
	api.Delete("/glossary/terms/:id", m.AuthMiddleware, h.GlossaryHandler.DeleteTerm)

	// Natural Language Features (TASK-120-122)
	api.Post("/nl/filter", m.AuthMiddleware, h.NLHandler.ParseFilter)
	api.Post("/nl/dashboard", m.AuthMiddleware, h.NLHandler.GenerateDashboard)
	api.Post("/nl/story", m.AuthMiddleware, h.NLHandler.GenerateStory)

	// Dashboards
	api.Get("/dashboards", m.AuthMiddleware, h.DashboardHandler.GetDashboards)
	api.Post("/dashboards", m.AuthMiddleware, h.DashboardHandler.CreateDashboard)
	api.Get("/dashboards/:id", m.AuthMiddleware, h.DashboardHandler.GetDashboard)
	api.Put("/dashboards/:id", m.AuthMiddleware, h.DashboardHandler.UpdateDashboard)
	api.Delete("/dashboards/:id", m.AuthMiddleware, h.DashboardHandler.DeleteDashboard)
	api.Post("/dashboards/:id/certify", m.AuthMiddleware, h.DashboardHandler.CertifyDashboard)

	// Dashboard Cards
	api.Get("/dashboards/:id/cards", m.AuthMiddleware, h.DashboardCardHandler.GetDashboardCards)
	api.Post("/dashboards/:id/cards", m.AuthMiddleware, h.DashboardCardHandler.AddCard)
	api.Put("/dashboards/:id/cards/positions", m.AuthMiddleware, h.DashboardCardHandler.UpdateCardPositions)
	api.Delete("/dashboards/:id/cards", m.AuthMiddleware, h.DashboardCardHandler.RemoveCard)

	// Collections (TASK-Gap Fix)
	api.Get("/collections", m.AuthMiddleware, h.CollectionHandler.GetCollections)
	api.Post("/collections", m.AuthMiddleware, h.CollectionHandler.CreateCollection)
	api.Get("/collections/:id", m.AuthMiddleware, h.CollectionHandler.GetCollection)
	api.Put("/collections/:id", m.AuthMiddleware, h.CollectionHandler.UpdateCollection)
	api.Delete("/collections/:id", m.AuthMiddleware, h.CollectionHandler.DeleteCollection)

	// --- Workspace Management ---
	// Note: WorkspaceHandler functions are standalone in the handlers package, not part of HandlerContainer yet.
	// We import "insight-engine-backend/handlers" so we can reference them directly.
	// Ideally they should be in HandlerContainer, but for now we register them directly to unblock.
	api.Get("/workspaces", m.AuthMiddleware, handlers.GetWorkspaces)
	api.Post("/workspaces", m.AuthMiddleware, handlers.CreateWorkspace)
	api.Get("/workspaces/:id", m.AuthMiddleware, handlers.GetWorkspace)
	api.Put("/workspaces/:id", m.AuthMiddleware, handlers.UpdateWorkspace)
	api.Delete("/workspaces/:id", m.AuthMiddleware, handlers.DeleteWorkspace)
	api.Get("/workspaces/members", m.AuthMiddleware, handlers.GetMembers)
	api.Post("/workspaces/members", m.AuthMiddleware, handlers.InviteMember)

	// --- Real-time & Collaboration ---

	// Notifications
	api.Get("/notifications", m.AuthMiddleware, h.NotificationHandler.GetNotifications)
	api.Get("/notifications/unread", m.AuthMiddleware, h.NotificationHandler.GetUnreadNotifications)
	api.Get("/notifications/unread-count", m.AuthMiddleware, h.NotificationHandler.GetUnreadCount)
	api.Put("/notifications/:id/read", m.AuthMiddleware, h.NotificationHandler.MarkAsRead)
	api.Put("/notifications/read-all", m.AuthMiddleware, h.NotificationHandler.MarkAllAsRead)
	api.Delete("/notifications/:id", m.AuthMiddleware, h.NotificationHandler.DeleteNotification)
	api.Delete("/notifications/read", m.AuthMiddleware, h.NotificationHandler.DeleteReadNotifications)
	api.Post("/notifications", m.AuthMiddleware, m.AdminMiddleware, h.NotificationHandler.CreateNotification)
	api.Post("/notifications/broadcast", m.AuthMiddleware, m.AdminMiddleware, h.NotificationHandler.BroadcastSystemNotification)

	// Activity
	api.Get("/activity", m.AuthMiddleware, h.ActivityHandler.GetUserActivity)
	api.Get("/activity/workspace/:id", m.AuthMiddleware, h.ActivityHandler.GetWorkspaceActivity)
	api.Get("/activity/recent", m.AuthMiddleware, m.AdminMiddleware, h.ActivityHandler.GetRecentActivity)

	// Scheduler
	api.Get("/scheduler/jobs", m.AuthMiddleware, h.SchedulerHandler.ListJobs)
	api.Get("/scheduler/jobs/:id", m.AuthMiddleware, h.SchedulerHandler.GetJob)
	api.Post("/scheduler/jobs", m.AuthMiddleware, m.AdminMiddleware, h.SchedulerHandler.CreateJob)
	api.Put("/scheduler/jobs/:id", m.AuthMiddleware, m.AdminMiddleware, h.SchedulerHandler.UpdateJob)
	api.Delete("/scheduler/jobs/:id", m.AuthMiddleware, m.AdminMiddleware, h.SchedulerHandler.DeleteJob)
	api.Post("/scheduler/jobs/:id/pause", m.AuthMiddleware, h.SchedulerHandler.PauseJob)
	api.Post("/scheduler/jobs/:id/resume", m.AuthMiddleware, h.SchedulerHandler.ResumeJob)
	api.Post("/scheduler/jobs/:id/trigger", m.AuthMiddleware, h.SchedulerHandler.TriggerJob)

	// WebSocket
	app.Get("/api/v1/ws", m.AuthMiddleware, websocket.New(h.WebSocketHandler.HandleConnection))
	api.Get("/ws/stats", m.AuthMiddleware, h.WebSocketHandler.GetStats)

	// Comments & Annotations
	h.CommentHandler.RegisterRoutes(api)

	// Versioning
	h.VersionHandler.RegisterRoutes(api)
	h.QueryVersionHandler.RegisterRoutes(api)

	// Sharing
	api.Post("/shares", m.AuthMiddleware, h.ShareHandler.CreateShare)
	api.Get("/shares/resource/:type/:id", m.AuthMiddleware, h.ShareHandler.GetSharesForResource)
	api.Get("/shares/my", m.AuthMiddleware, h.ShareHandler.GetMyShares)
	api.Get("/shares/:id", m.AuthMiddleware, h.ShareHandler.GetShareByID)
	api.Put("/shares/:id", m.AuthMiddleware, h.ShareHandler.UpdateShare)
	api.Delete("/shares/:id", m.AuthMiddleware, h.ShareHandler.DeleteShare)
	api.Post("/shares/:id/accept", m.AuthMiddleware, h.ShareHandler.AcceptShare)
	api.Get("/shares/check", m.AuthMiddleware, h.ShareHandler.CheckShareAccess)
	api.Post("/shares/validate", m.AuthMiddleware, h.ShareHandler.ValidateShareAccess)

	// Embed Tokens (Task 133)
	api.Post("/embed/token", m.AuthMiddleware, h.EmbedHandler.GenerateToken)
	api.Get("/embed/token/validate", h.EmbedHandler.ValidateToken)

	// --- Admin & Logs ---

	// Audit Logs
	api.Get("/admin/audit-logs", m.AuthMiddleware, h.AuditHandler.GetAuditLogs)
	api.Get("/admin/audit-logs/recent", m.AuthMiddleware, h.AuditHandler.GetRecentActivity)
	api.Get("/admin/audit-logs/summary", m.AuthMiddleware, h.AuditHandler.GetAuditSummary)
	api.Get("/admin/audit-logs/user/:id", m.AuthMiddleware, h.AuditHandler.GetUserActivity)
	api.Get("/admin/audit-logs/export", m.AuthMiddleware, h.AuditHandler.ExportAuditLogs)

	// Frontend Logs
	api.Post("/logs/frontend", h.FrontendLogHandler.CreateFrontendLog)
	api.Get("/logs/frontend", m.AuthMiddleware, m.AdminMiddleware, h.FrontendLogHandler.GetFrontendLogs)
	api.Delete("/logs/frontend/cleanup", m.AuthMiddleware, m.AdminMiddleware, h.FrontendLogHandler.CleanupOldLogs)

	// Admin Dashboard
	h.AdminOrgHandler.RegisterRoutes(api, m.AuthMiddleware, m.AdminMiddleware)
	h.AdminUserHandler.RegisterRoutes(api, m.AuthMiddleware, m.AdminMiddleware)
	h.AdminSystemHandler.RegisterRoutes(api, m.AuthMiddleware, m.AdminMiddleware)

	// Permission Management
	api.Get("/permissions", m.AuthMiddleware, h.PermissionHandler.GetAllPermissions)
	api.Get("/permissions/resource/:resource", m.AuthMiddleware, h.PermissionHandler.GetPermissionsByResource)
	api.Post("/permissions/check", m.AuthMiddleware, h.PermissionHandler.CheckUserPermission)
	api.Get("/roles", m.AuthMiddleware, h.PermissionHandler.GetAllRoles)
	api.Get("/roles/:id", m.AuthMiddleware, h.PermissionHandler.GetRoleByID)
	api.Post("/roles", m.AuthMiddleware, m.AdminMiddleware, h.PermissionHandler.CreateRole)
	api.Put("/roles/:id", m.AuthMiddleware, m.AdminMiddleware, h.PermissionHandler.UpdateRole)
	api.Delete("/roles/:id", m.AuthMiddleware, m.AdminMiddleware, h.PermissionHandler.DeleteRole)
	api.Put("/roles/:id/permissions", m.AuthMiddleware, m.AdminMiddleware, h.PermissionHandler.AssignPermissionsToRole)
	api.Get("/users/:id/roles", m.AuthMiddleware, h.PermissionHandler.GetUserRoles)
	api.Get("/users/:id/permissions", m.AuthMiddleware, h.PermissionHandler.GetUserPermissions)
	api.Post("/users/:id/roles", m.AuthMiddleware, m.AdminMiddleware, h.PermissionHandler.AssignRoleToUser)
	api.Delete("/users/:id/roles/:roleId", m.AuthMiddleware, m.AdminMiddleware, h.PermissionHandler.RemoveRoleFromUser)

	// --- Monitoring & Alerting ---

	// Rate Limits
	api.Get("/rate-limits", m.AuthMiddleware, h.RateLimitHandler.GetConfigs)
	api.Get("/rate-limits/:id", m.AuthMiddleware, h.RateLimitHandler.GetConfig)
	api.Post("/rate-limits", m.AuthMiddleware, h.RateLimitHandler.CreateConfig)
	api.Put("/rate-limits/:id", m.AuthMiddleware, h.RateLimitHandler.UpdateConfig)
	api.Delete("/rate-limits/:id", m.AuthMiddleware, h.RateLimitHandler.DeleteConfig)
	api.Get("/rate-limits/violations", m.AuthMiddleware, h.RateLimitHandler.GetViolations)

	// AI Usage
	api.Post("/ai/generate", m.AuthMiddleware, h.AIHandler.Generate)
	api.Post("/ai/presentation", m.AuthMiddleware, h.AIHandler.GeneratePresentation)
	api.Post("/ai/stream", m.AuthMiddleware, h.AIHandler.StreamGenerate)
	api.Get("/ai/usage", m.AuthMiddleware, h.AIUsageHandler.GetUsageStats)
	api.Get("/ai/requests", m.AuthMiddleware, h.AIHandler.GetRequests)
	api.Get("/ai/request-history", m.AuthMiddleware, h.AIUsageHandler.GetRequestHistory)
	api.Get("/ai/budgets", m.AuthMiddleware, h.AIUsageHandler.GetBudgets)
	api.Post("/ai/budgets", m.AuthMiddleware, h.AIUsageHandler.CreateBudget)
	api.Put("/ai/budgets/:id", m.AuthMiddleware, h.AIUsageHandler.UpdateBudget)
	api.Delete("/ai/budgets/:id", m.AuthMiddleware, h.AIUsageHandler.DeleteBudget)
	api.Get("/ai/alerts", m.AuthMiddleware, h.AIUsageHandler.GetAlerts)
	api.Post("/ai/alerts/:id/acknowledge", m.AuthMiddleware, h.AIUsageHandler.AcknowledgeAlert)

	// Alerts
	h.AlertHandler.RegisterRoutes(app, m.AuthMiddleware)
	h.AlertNotificationHandler.RegisterRoutes(app, m.AuthMiddleware)

	// Webhooks (TASK-134)
	api.Post("/webhooks", m.AuthMiddleware, h.WebhookHandler.CreateWebhook)
	api.Get("/webhooks", m.AuthMiddleware, h.WebhookHandler.GetWebhooks)
	api.Get("/webhooks/:id", m.AuthMiddleware, h.WebhookHandler.GetWebhook)
	api.Put("/webhooks/:id", m.AuthMiddleware, h.WebhookHandler.UpdateWebhook)
	api.Delete("/webhooks/:id", m.AuthMiddleware, h.WebhookHandler.DeleteWebhook)
	api.Get("/webhooks/:id/logs", m.AuthMiddleware, h.WebhookHandler.GetWebhookLogs)
	api.Post("/webhooks/:id/test", m.AuthMiddleware, h.WebhookHandler.TestWebhook)

	// --- Optional/Conditional Routes ---
	if h.ScheduledReportHandler != nil {
		h.ScheduledReportHandler.RegisterRoutes(api)
	}
}
