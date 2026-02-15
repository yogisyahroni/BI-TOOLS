package handlers

import (
	"insight-engine-backend/models"
	"insight-engine-backend/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// AlertNotificationHandler handles alert notification-related requests
type AlertNotificationHandler struct {
	notificationService *services.AlertNotificationService
}

// NewAlertNotificationHandler creates a new alert notification handler
func NewAlertNotificationHandler(notificationService *services.AlertNotificationService) *AlertNotificationHandler {
	return &AlertNotificationHandler{
		notificationService: notificationService,
	}
}

// RegisterRoutes registers all alert notification routes
func (h *AlertNotificationHandler) RegisterRoutes(app *fiber.App, authMiddleware fiber.Handler) {
	routes := app.Group("/api/alerts/notifications")
	routes.Use(authMiddleware)

	// Notification channels
	routes.Get("/channels", h.GetNotificationChannels)
	routes.Post("/channels/test", h.TestNotificationChannel)

	// Templates
	routes.Get("/templates", h.GetNotificationTemplates)
	routes.Get("/templates/:id", h.GetNotificationTemplate)
	routes.Put("/templates/:id", h.UpdateNotificationTemplate)
}

// GetNotificationChannels returns available notification channels
func (h *AlertNotificationHandler) GetNotificationChannels(c *fiber.Ctx) error {
	channels := []map[string]interface{}{
		{
			"type":        models.AlertChannelEmail,
			"name":        "Email",
			"description": "Send email notifications",
			"icon":        "mail",
		},
		{
			"type":        models.AlertChannelWebhook,
			"name":        "Webhook",
			"description": "Send HTTP POST requests",
			"icon":        "webhook",
		},
		{
			"type":        models.AlertChannelInApp,
			"name":        "In-App",
			"description": "Send in-app notifications",
			"icon":        "bell",
		},
		{
			"type":        models.AlertChannelSlack,
			"name":        "Slack",
			"description": "Send Slack messages",
			"icon":        "message-square",
		},
	}

	return c.JSON(fiber.Map{
		"channels": channels,
	})
}

// TestNotificationChannel tests a notification channel configuration
func (h *AlertNotificationHandler) TestNotificationChannel(c *fiber.Ctx) error {
	var req models.TestNotificationChannelRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.ChannelType == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Channel type is required",
		})
	}

	// For now, just validate the configuration
	// In production, you would actually send a test notification
	switch req.ChannelType {
	case models.AlertChannelEmail:
		if req.Config == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Email configuration is required",
			})
		}
		recipients, ok := req.Config["recipients"].([]interface{})
		if !ok || len(recipients) == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "At least one recipient is required",
			})
		}
	case models.AlertChannelWebhook:
		if req.Config == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Webhook configuration is required",
			})
		}
		url, ok := req.Config["url"].(string)
		if !ok || url == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Webhook URL is required",
			})
		}
	case models.AlertChannelSlack:
		if req.Config == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Slack configuration is required",
			})
		}
		url, ok := req.Config["url"].(string)
		if !ok || url == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Slack webhook URL is required",
			})
		}
	case models.AlertChannelInApp:
		// No additional configuration needed for in-app
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Unsupported channel type",
		})
	}

	return c.JSON(models.TestNotificationChannelResponse{
		Success: true,
		Message: "Channel configuration is valid",
	})
}

// GetNotificationTemplates returns available notification templates
func (h *AlertNotificationHandler) GetNotificationTemplates(c *fiber.Ctx) error {
	templates, err := h.notificationService.GetNotificationTemplates()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"templates": templates,
	})
}

// GetNotificationTemplate returns a single notification template
func (h *AlertNotificationHandler) GetNotificationTemplate(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Template ID is required",
		})
	}

	templateID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid template ID",
		})
	}

	// Get template from service (would need to implement this method)
	_ = templateID
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "Not implemented",
	})
}

// UpdateNotificationTemplate updates a notification template
func (h *AlertNotificationHandler) UpdateNotificationTemplate(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Template ID is required",
		})
	}

	templateID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid template ID",
		})
	}

	var req struct {
		Name        string `json:"name,omitempty"`
		Description string `json:"description,omitempty"`
		Content     string `json:"content,omitempty"`
		IsActive    *bool  `json:"isActive,omitempty"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Update template (would need to implement this method)
	_ = templateID
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "Not implemented",
	})
}
