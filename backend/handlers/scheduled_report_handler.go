package handlers

import (
	"strconv"

	"insight-engine-backend/models"
	"insight-engine-backend/services"

	"github.com/gofiber/fiber/v2"
)

// ScheduledReportHandler handles scheduled report API endpoints
type ScheduledReportHandler struct {
	service *services.ScheduledReportService
}

// NewScheduledReportHandler creates a new scheduled report handler
func NewScheduledReportHandler(service *services.ScheduledReportService) *ScheduledReportHandler {
	return &ScheduledReportHandler{service: service}
}

// CreateScheduledReport creates a new scheduled report
func (h *ScheduledReportHandler) CreateScheduledReport(c *fiber.Ctx) error {
	userID, ok := c.Locals("userId").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	var req models.CreateScheduledReportRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body: " + err.Error(),
		})
	}

	// Validate request
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name is required",
		})
	}
	if req.ResourceType == "" || req.ResourceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Resource type and ID are required",
		})
	}
	if req.ScheduleType == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Schedule type is required",
		})
	}
	if req.Format == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Format is required",
		})
	}
	if len(req.Recipients) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "At least one recipient is required",
		})
	}

	report, err := h.service.CreateScheduledReport(userID, &req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(report)
}

// GetScheduledReports lists all scheduled reports for the user
func (h *ScheduledReportHandler) GetScheduledReports(c *fiber.Ctx) error {
	userID, ok := c.Locals("userId").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Parse query parameters
	filter := &models.ScheduledReportFilter{}

	if resourceType := c.Query("resourceType"); resourceType != "" {
		t := models.ReportResourceType(resourceType)
		filter.ResourceType = &t
	}

	if resourceID := c.Query("resourceId"); resourceID != "" {
		filter.ResourceID = &resourceID
	}

	if isActive := c.Query("isActive"); isActive != "" {
		b := isActive == "true"
		filter.IsActive = &b
	}

	if scheduleType := c.Query("scheduleType"); scheduleType != "" {
		t := models.ReportScheduleType(scheduleType)
		filter.ScheduleType = &t
	}

	if search := c.Query("search"); search != "" {
		filter.Search = &search
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	filter.Page = page

	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	filter.Limit = limit

	filter.OrderBy = c.Query("orderBy", "created_at DESC")

	response, err := h.service.GetScheduledReports(userID, filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(response)
}

// GetScheduledReport retrieves a single scheduled report
func (h *ScheduledReportHandler) GetScheduledReport(c *fiber.Ctx) error {
	userID, ok := c.Locals("userId").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	reportID := c.Params("id")
	if reportID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Report ID is required",
		})
	}

	report, err := h.service.GetScheduledReport(reportID, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(report)
}

// UpdateScheduledReport updates a scheduled report
func (h *ScheduledReportHandler) UpdateScheduledReport(c *fiber.Ctx) error {
	userID, ok := c.Locals("userId").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	reportID := c.Params("id")
	if reportID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Report ID is required",
		})
	}

	var req models.UpdateScheduledReportRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body: " + err.Error(),
		})
	}

	report, err := h.service.UpdateScheduledReport(reportID, userID, &req)
	if err != nil {
		if err.Error() == "scheduled report not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(report)
}

// DeleteScheduledReport deletes a scheduled report
func (h *ScheduledReportHandler) DeleteScheduledReport(c *fiber.Ctx) error {
	userID, ok := c.Locals("userId").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	reportID := c.Params("id")
	if reportID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Report ID is required",
		})
	}

	if err := h.service.DeleteScheduledReport(reportID, userID); err != nil {
		if err.Error() == "scheduled report not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// TriggerScheduledReport manually triggers a scheduled report
func (h *ScheduledReportHandler) TriggerScheduledReport(c *fiber.Ctx) error {
	userID, ok := c.Locals("userId").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	reportID := c.Params("id")
	if reportID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Report ID is required",
		})
	}

	response, err := h.service.ExecuteScheduledReport(c.Context(), reportID, "manual")
	if err != nil {
		if err.Error() == "scheduled report not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(response)
}

// GetScheduledReportRuns retrieves run history for a scheduled report
func (h *ScheduledReportHandler) GetScheduledReportRuns(c *fiber.Ctx) error {
	userID, ok := c.Locals("userId").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	reportID := c.Params("id")
	if reportID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Report ID is required",
		})
	}

	// Parse query parameters
	filter := &models.ScheduledReportRunFilter{}

	if status := c.Query("status"); status != "" {
		s := models.ReportRunStatus(status)
		filter.Status = &s
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	filter.Page = page

	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	filter.Limit = limit

	filter.OrderBy = c.Query("orderBy", "started_at DESC")

	response, err := h.service.GetScheduledReportRuns(reportID, userID, filter)
	if err != nil {
		if err.Error() == "scheduled report not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(response)
}

// PreviewScheduledReport generates a preview of a report
func (h *ScheduledReportHandler) PreviewScheduledReport(c *fiber.Ctx) error {
	userID, ok := c.Locals("userId").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	reportID := c.Params("id")
	if reportID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Report ID is required",
		})
	}

	// Get the report
	report, err := h.service.GetScheduledReport(reportID, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Create preview request from report
	req := &models.ReportPreviewRequest{
		ResourceType:   report.ResourceType,
		ResourceID:     report.ResourceID,
		Format:         report.Format,
		IncludeFilters: report.IncludeFilters,
	}

	preview, err := h.service.PreviewReport(userID, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(preview)
}

// ToggleScheduledReportActive toggles the active status of a report
func (h *ScheduledReportHandler) ToggleScheduledReportActive(c *fiber.Ctx) error {
	userID, ok := c.Locals("userId").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	reportID := c.Params("id")
	if reportID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Report ID is required",
		})
	}

	report, err := h.service.ToggleReportActive(reportID, userID)
	if err != nil {
		if err.Error() == "scheduled report not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"id":       report.ID,
		"isActive": report.IsActive,
		"message":  "Report status updated successfully",
	})
}

// GetReportRunDownload returns a download URL for a report run
func (h *ScheduledReportHandler) GetReportRunDownload(c *fiber.Ctx) error {
	userID, ok := c.Locals("userId").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	runID := c.Params("runId")
	if runID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Run ID is required",
		})
	}

	url, err := h.service.GetRunDownloadURL(runID, userID)
	if err != nil {
		if err.Error() == "access denied" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"downloadUrl": url,
	})
}

// PreviewReport generates a preview without saving
func (h *ScheduledReportHandler) PreviewReport(c *fiber.Ctx) error {
	userID, ok := c.Locals("userId").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	var req models.ReportPreviewRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body: " + err.Error(),
		})
	}

	if req.ResourceType == "" || req.ResourceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Resource type and ID are required",
		})
	}
	if req.Format == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Format is required",
		})
	}

	preview, err := h.service.PreviewReport(userID, &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(preview)
}

// GetTimezones returns a list of supported timezones
func (h *ScheduledReportHandler) GetTimezones(c *fiber.Ctx) error {
	// Return common timezones
	timezones := []map[string]string{
		{"value": "UTC", "label": "UTC (Coordinated Universal Time)"},
		{"value": "America/New_York", "label": "Eastern Time (US & Canada)"},
		{"value": "America/Chicago", "label": "Central Time (US & Canada)"},
		{"value": "America/Denver", "label": "Mountain Time (US & Canada)"},
		{"value": "America/Los_Angeles", "label": "Pacific Time (US & Canada)"},
		{"value": "America/Anchorage", "label": "Alaska Time"},
		{"value": "America/Honolulu", "label": "Hawaii Time"},
		{"value": "Europe/London", "label": "London (GMT)"},
		{"value": "Europe/Paris", "label": "Paris (CET)"},
		{"value": "Europe/Berlin", "label": "Berlin (CET)"},
		{"value": "Asia/Tokyo", "label": "Tokyo (JST)"},
		{"value": "Asia/Shanghai", "label": "Shanghai (CST)"},
		{"value": "Asia/Singapore", "label": "Singapore (SGT)"},
		{"value": "Asia/Dubai", "label": "Dubai (GST)"},
		{"value": "Asia/Mumbai", "label": "Mumbai (IST)"},
		{"value": "Australia/Sydney", "label": "Sydney (AEDT)"},
		{"value": "Pacific/Auckland", "label": "Auckland (NZDT)"},
	}

	return c.JSON(fiber.Map{
		"timezones": timezones,
	})
}

// RegisterRoutes registers the scheduled report routes
func (h *ScheduledReportHandler) RegisterRoutes(app fiber.Router) {
	// Main CRUD routes
	app.Post("/scheduled-reports", h.CreateScheduledReport)
	app.Get("/scheduled-reports", h.GetScheduledReports)
	app.Get("/scheduled-reports/timezones", h.GetTimezones)
	app.Post("/scheduled-reports/preview", h.PreviewReport)
	app.Get("/scheduled-reports/:id", h.GetScheduledReport)
	app.Put("/scheduled-reports/:id", h.UpdateScheduledReport)
	app.Delete("/scheduled-reports/:id", h.DeleteScheduledReport)

	// Action routes
	app.Post("/scheduled-reports/:id/trigger", h.TriggerScheduledReport)
	app.Post("/scheduled-reports/:id/toggle", h.ToggleScheduledReportActive)
	app.Get("/scheduled-reports/:id/history", h.GetScheduledReportRuns)
	app.Get("/scheduled-reports/:id/preview", h.PreviewScheduledReport)

	// Download route
	app.Get("/scheduled-reports/runs/:runId/download", h.GetReportRunDownload)
}
