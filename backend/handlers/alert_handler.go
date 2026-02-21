package handlers

import (
	"insight-engine-backend/models"
	"insight-engine-backend/services"

	"github.com/gofiber/fiber/v2"
)

// AlertHandler handles alert-related requests
type AlertHandler struct {
	alertService *services.AlertService
}

// NewAlertHandler creates a new alert handler
func NewAlertHandler(alertService *services.AlertService) *AlertHandler {
	return &AlertHandler{
		alertService: alertService,
	}
}

// RegisterRoutes registers all alert routes
func (h *AlertHandler) RegisterRoutes(app *fiber.App, authMiddleware func(*fiber.Ctx) error) {
	alertRoutes := app.Group("/api/alerts", authMiddleware)

	alertRoutes.Get("/", h.ListAlerts)
	alertRoutes.Post("/", h.CreateAlert)
	alertRoutes.Get("/triggered", h.GetTriggeredAlerts)
	alertRoutes.Get("/stats", h.GetAlertStats)
	alertRoutes.Post("/test", h.TestAlert)
	alertRoutes.Get("/:id", h.GetAlert)
	alertRoutes.Put("/:id", h.UpdateAlert)
	alertRoutes.Delete("/:id", h.DeleteAlert)
	alertRoutes.Post("/:id/execute", h.ExecuteAlertCheck)
	alertRoutes.Post("/:id/acknowledge", h.AcknowledgeAlert)
	alertRoutes.Post("/:id/mute", h.MuteAlert)
	alertRoutes.Post("/:id/unmute", h.UnmuteAlert)
	alertRoutes.Get("/:id/history", h.GetAlertHistory)
}

// ListAlerts handles GET /api/alerts
func (h *AlertHandler) ListAlerts(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	filter := &models.AlertFilter{
		UserID: &userID,
	}

	result, err := h.alertService.ListAlerts(filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch alerts",
		})
	}

	return c.JSON(result)
}

// CreateAlert handles POST /api/alerts
func (h *AlertHandler) CreateAlert(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var req models.CreateAlertRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	alert, err := h.alertService.CreateAlert(userID, &req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(alert)
}

// GetAlert handles GET /api/alerts/:id
func (h *AlertHandler) GetAlert(c *fiber.Ctx) error {
	alertID := c.Params("id")
	userID := c.Locals("userID").(string)

	alert, err := h.alertService.GetAlert(alertID, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Alert not found",
		})
	}

	return c.JSON(alert)
}

// UpdateAlert handles PUT /api/alerts/:id
func (h *AlertHandler) UpdateAlert(c *fiber.Ctx) error {
	alertID := c.Params("id")
	userID := c.Locals("userID").(string)

	var req models.UpdateAlertRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	alert, err := h.alertService.UpdateAlert(alertID, userID, &req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(alert)
}

// DeleteAlert handles DELETE /api/alerts/:id
func (h *AlertHandler) DeleteAlert(c *fiber.Ctx) error {
	alertID := c.Params("id")
	userID := c.Locals("userID").(string)

	if err := h.alertService.DeleteAlert(alertID, userID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Alert deleted successfully",
	})
}

// ExecuteAlertCheck handles POST /api/alerts/:id/execute
func (h *AlertHandler) ExecuteAlertCheck(c *fiber.Ctx) error {
	alertID := c.Params("id")

	history, err := h.alertService.ExecuteAlertCheck(c.Context(), alertID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(history)
}

// AcknowledgeAlert handles POST /api/alerts/:id/acknowledge
func (h *AlertHandler) AcknowledgeAlert(c *fiber.Ctx) error {
	alertID := c.Params("id")
	userID := c.Locals("userID").(string)

	var req models.AcknowledgeAlertRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	alert, err := h.alertService.AcknowledgeAlert(alertID, userID, &req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(alert)
}

// MuteAlert handles POST /api/alerts/:id/mute
func (h *AlertHandler) MuteAlert(c *fiber.Ctx) error {
	alertID := c.Params("id")
	userID := c.Locals("userID").(string)

	var req models.MuteAlertRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	alert, err := h.alertService.MuteAlert(alertID, userID, &req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(alert)
}

// UnmuteAlert handles POST /api/alerts/:id/unmute
func (h *AlertHandler) UnmuteAlert(c *fiber.Ctx) error {
	alertID := c.Params("id")
	userID := c.Locals("userID").(string)

	alert, err := h.alertService.UnmuteAlert(alertID, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(alert)
}

// GetAlertHistory handles GET /api/alerts/:id/history
func (h *AlertHandler) GetAlertHistory(c *fiber.Ctx) error {
	alertID := c.Params("id")
	userID := c.Locals("userID").(string)

	filter := &models.AlertHistoryFilter{}
	// Parse query parameters if needed

	result, err := h.alertService.GetAlertHistory(alertID, userID, filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(result)
}

// GetTriggeredAlerts handles GET /api/alerts/triggered
func (h *AlertHandler) GetTriggeredAlerts(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	alerts, err := h.alertService.GetTriggeredAlerts(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"alerts": alerts,
		"total":  len(alerts),
	})
}

// GetAlertStats handles GET /api/alerts/stats
func (h *AlertHandler) GetAlertStats(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	stats, err := h.alertService.GetAlertStats(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(stats)
}

// TestAlert handles POST /api/alerts/test
func (h *AlertHandler) TestAlert(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var req models.TestAlertRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	result, err := h.alertService.TestAlert(c.Context(), userID, &req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(result)
}
