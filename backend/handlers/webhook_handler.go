package handlers

import (
	"fmt"
	"insight-engine-backend/pkg/validator"
	"strconv"

	"insight-engine-backend/models"
	"insight-engine-backend/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type WebhookHandler struct {
	webhookService *services.WebhookService
}

func NewWebhookHandler(webhookService *services.WebhookService) *WebhookHandler {
	return &WebhookHandler{webhookService: webhookService}
}

// CreateWebhook creates a new webhook
func (h *WebhookHandler) CreateWebhook(c *fiber.Ctx) error {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var req struct {
		Name        string   `json:"name" validate:"required"`
		URL         string   `json:"url" validate:"required,url"`
		Events      []string `json:"events" validate:"required,min=1"`
		Description string   `json:"description"`
		IsActive    bool     `json:"isActive"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := validator.GetValidator().ValidateStruct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Map to model
	webhookReq := models.WebhookRequest{
		Name:        req.Name,
		URL:         req.URL,
		Events:      req.Events,
		Description: req.Description,
		IsActive:    req.IsActive,
	}

	webhook, err := h.webhookService.CreateWebhook(webhookReq, userID)
	if err != nil {
		fmt.Printf("DEBUG: CreateWebhook Service Error: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create webhook"})
	}

	return c.Status(fiber.StatusCreated).JSON(webhook)
}

// GetWebhooks returns all webhooks for the current user
func (h *WebhookHandler) GetWebhooks(c *fiber.Ctx) error {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	webhooks, err := h.webhookService.GetWebhooks(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch webhooks"})
	}

	return c.JSON(webhooks)
}

// GetWebhook returns a single webhook
func (h *WebhookHandler) GetWebhook(c *fiber.Ctx) error {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid webhook ID"})
	}

	webhook, err := h.webhookService.GetWebhook(id, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Webhook not found"})
	}

	return c.JSON(webhook)
}

// UpdateWebhook updates a webhook
func (h *WebhookHandler) UpdateWebhook(c *fiber.Ctx) error {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid webhook ID"})
	}

	var req struct {
		Name        string   `json:"name"`
		URL         string   `json:"url" validate:"omitempty,url"`
		Events      []string `json:"events" validate:"omitempty,min=1"`
		Description string   `json:"description"`
		IsActive    bool     `json:"isActive"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := validator.GetValidator().ValidateStruct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Map to model
	webhookReq := models.WebhookRequest{
		Name:        req.Name,
		URL:         req.URL,
		Events:      req.Events,
		Description: req.Description,
		IsActive:    req.IsActive,
	}

	webhook, err := h.webhookService.UpdateWebhook(id, webhookReq, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update webhook"})
	}

	return c.JSON(webhook)
}

// DeleteWebhook deletes a webhook
func (h *WebhookHandler) DeleteWebhook(c *fiber.Ctx) error {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid webhook ID"})
	}

	if err := h.webhookService.DeleteWebhook(id, userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete webhook"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// GetWebhookLogs returns logs for a webhook
func (h *WebhookHandler) GetWebhookLogs(c *fiber.Ctx) error {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid webhook ID"})
	}

	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	logs, err := h.webhookService.GetWebhookLogs(id, userID, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch logs"})
	}

	return c.JSON(logs)
}

// TestWebhook triggers a test event for a webhook
func (h *WebhookHandler) TestWebhook(c *fiber.Ctx) error {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid webhook ID"})
	}

	var req struct {
		Event   string                 `json:"event"`
		Payload map[string]interface{} `json:"payload"`
	}
	if err := c.BodyParser(&req); err != nil {
		// Use default test payload if none provided
		req.Event = "test.event"
		req.Payload = map[string]interface{}{"message": "This is a test event"}
	}

	// Use the explicit test method
	if err := h.webhookService.SendTestWebhook(id, userID, req.Event, req.Payload); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to send test webhook: " + err.Error()})
	}

	// Since we haven't updated service yet, let's do that next.
	// For now return success that test initiated

	// We will implement SendTestEvent in service in next step
	return c.JSON(fiber.Map{"message": "Test event dispatched"})
}

// Helper to get user ID from context (duplicated from other handlers, common util needed)
func getUserIDFromContext(c *fiber.Ctx) (uuid.UUID, error) {
	// Assuming AuthMiddleware sets "userID" or "userId" in Locals
	userIDInterface := c.Locals("userID")
	if userIDInterface == nil {
		userIDInterface = c.Locals("userId")
	}
	if userIDInterface == nil {
		userIDInterface = c.Locals("user_id") // Fail-safe
	}

	if userIDInterface == nil {
		// Fallback to "user" object if "user_id" not directly set
		fmt.Println("DEBUG: getUserIDFromContext: userID is nil in Locals. Checking 'user' object.")
		userInterface := c.Locals("user")
		if userInterface != nil {
			if user, ok := userInterface.(*models.User); ok {
				return uuid.Parse(user.ID) // Parse string ID to UUID
			}
			// Map?
			if userMap, ok := userInterface.(map[string]interface{}); ok {
				if idStr, ok := userMap["id"].(string); ok {
					return uuid.Parse(idStr)
				}
			}
		}
		return uuid.Nil, fmt.Errorf("user not found in context")
	}

	fmt.Printf("DEBUG: getUserIDFromContext: Found userIDInterface: %v, Type: %T\n", userIDInterface, userIDInterface)

	if id, ok := userIDInterface.(string); ok {
		return uuid.Parse(id)
	}
	if id, ok := userIDInterface.(uuid.UUID); ok {
		return id, nil
	}
	return uuid.Nil, fmt.Errorf("invalid user ID type: %T", userIDInterface)
}
