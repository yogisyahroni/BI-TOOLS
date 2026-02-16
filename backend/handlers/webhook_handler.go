package handlers

import (
	"context"
	"fmt"
	"insight-engine-backend/models"
	"insight-engine-backend/services"
	"net/url"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WebhookHandler handles CRUD for external webhook configurations (Slack/Teams)
type WebhookHandler struct {
	db *gorm.DB
}

// NewWebhookHandler creates a new WebhookHandler
func NewWebhookHandler(db *gorm.DB) *WebhookHandler {
	return &WebhookHandler{db: db}
}

// CreateWebhook registers a new Slack/Teams webhook
func (h *WebhookHandler) CreateWebhook(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	var input struct {
		Name        string `json:"name" validate:"required"`
		ChannelType string `json:"channelType" validate:"required"`
		WebhookURL  string `json:"webhookUrl" validate:"required"`
		Description string `json:"description"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate channel type
	if input.ChannelType != "slack" && input.ChannelType != "teams" {
		return c.Status(400).JSON(fiber.Map{"error": "channelType must be 'slack' or 'teams'"})
	}

	// Validate webhook URL format
	parsedURL, err := url.ParseRequestURI(input.WebhookURL)
	if err != nil || (parsedURL.Scheme != "https") {
		return c.Status(400).JSON(fiber.Map{"error": "webhookUrl must be a valid HTTPS URL"})
	}

	// Validate Slack webhook URL pattern
	if input.ChannelType == "slack" {
		if parsedURL.Host != "hooks.slack.com" && parsedURL.Host != "hooks.slack-gov.com" {
			return c.Status(400).JSON(fiber.Map{
				"error":   "Invalid Slack webhook URL",
				"details": "Slack webhook URLs must start with https://hooks.slack.com/",
			})
		}
	}

	webhook := models.WebhookConfig{
		UserID:      parsedUserID,
		Name:        input.Name,
		ChannelType: models.WebhookChannelType(input.ChannelType),
		WebhookURL:  input.WebhookURL,
		Description: input.Description,
		IsActive:    true,
	}

	if err := h.db.Create(&webhook).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create webhook configuration"})
	}

	// Mask the webhook URL in response for security
	webhook.WebhookURL = maskWebhookURL(webhook.WebhookURL)

	return c.Status(201).JSON(webhook)
}

// ListWebhooks returns all webhook configurations for the current user
func (h *WebhookHandler) ListWebhooks(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var webhooks []models.WebhookConfig
	if err := h.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&webhooks).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to list webhook configurations"})
	}

	// Mask URLs in response
	for i := range webhooks {
		webhooks[i].WebhookURL = maskWebhookURL(webhooks[i].WebhookURL)
	}

	return c.JSON(fiber.Map{
		"data":  webhooks,
		"total": len(webhooks),
	})
}

// DeleteWebhook removes a webhook configuration
func (h *WebhookHandler) DeleteWebhook(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	webhookID := c.Params("id")

	result := h.db.Where("id = ? AND user_id = ?", webhookID, userID).
		Delete(&models.WebhookConfig{})

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete webhook"})
	}

	if result.RowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Webhook not found or access denied"})
	}

	return c.JSON(fiber.Map{"message": "Webhook deleted successfully"})
}

// ToggleWebhook enables or disables a webhook
func (h *WebhookHandler) ToggleWebhook(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	webhookID := c.Params("id")

	var input struct {
		IsActive bool `json:"isActive"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	result := h.db.Model(&models.WebhookConfig{}).
		Where("id = ? AND user_id = ?", webhookID, userID).
		Update("is_active", input.IsActive)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update webhook"})
	}

	if result.RowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Webhook not found or access denied"})
	}

	status := "disabled"
	if input.IsActive {
		status = "enabled"
	}

	return c.JSON(fiber.Map{"message": fmt.Sprintf("Webhook %s", status)})
}

// TestWebhook sends a test message to the configured webhook
func (h *WebhookHandler) TestWebhook(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	webhookID := c.Params("id")

	var webhook models.WebhookConfig
	if err := h.db.Where("id = ? AND user_id = ?", webhookID, userID).
		First(&webhook).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Webhook not found or access denied"})
	}

	if !webhook.IsActive {
		return c.Status(400).JSON(fiber.Map{"error": "Webhook is disabled. Enable it before testing."})
	}

	notifier, err := services.CreateNotifierForType(string(webhook.ChannelType), webhook.WebhookURL)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	testFields := map[string]string{
		"Webhook Name": webhook.Name,
		"Channel Type": string(webhook.ChannelType),
		"Status":       "Active",
	}

	if err := notifier.SendMessage(
		context.Background(),
		"🔔 InsightEngine Test Notification",
		"This is a test message from InsightEngine AI. If you see this, your webhook integration is working correctly!",
		testFields,
	); err != nil {
		return c.Status(502).JSON(fiber.Map{
			"error":   "Test message delivery failed",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{"message": "Test message sent successfully"})
}

// GetWebhooks is an alias for ListWebhooks — satisfies the TASK-134 route registration
func (h *WebhookHandler) GetWebhooks(c *fiber.Ctx) error {
	return h.ListWebhooks(c)
}

// GetWebhook retrieves a single webhook configuration by ID
func (h *WebhookHandler) GetWebhook(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	webhookID := c.Params("id")

	var webhook models.WebhookConfig
	if err := h.db.Where("id = ? AND user_id = ?", webhookID, userID).
		First(&webhook).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Webhook not found or access denied"})
	}

	webhook.WebhookURL = maskWebhookURL(webhook.WebhookURL)
	return c.JSON(webhook)
}

// UpdateWebhook updates webhook configuration fields
func (h *WebhookHandler) UpdateWebhook(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	webhookID := c.Params("id")

	var webhook models.WebhookConfig
	if err := h.db.Where("id = ? AND user_id = ?", webhookID, userID).
		First(&webhook).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Webhook not found or access denied"})
	}

	var input struct {
		Name        *string `json:"name"`
		ChannelType *string `json:"channelType"`
		WebhookURL  *string `json:"webhookUrl"`
		Description *string `json:"description"`
		IsActive    *bool   `json:"isActive"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	updates := map[string]interface{}{}
	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.ChannelType != nil {
		if *input.ChannelType != "slack" && *input.ChannelType != "teams" {
			return c.Status(400).JSON(fiber.Map{"error": "channelType must be 'slack' or 'teams'"})
		}
		updates["channel_type"] = *input.ChannelType
	}
	if input.WebhookURL != nil {
		parsedURL, err := url.ParseRequestURI(*input.WebhookURL)
		if err != nil || parsedURL.Scheme != "https" {
			return c.Status(400).JSON(fiber.Map{"error": "webhookUrl must be a valid HTTPS URL"})
		}
		updates["webhook_url"] = *input.WebhookURL
	}
	if input.Description != nil {
		updates["description"] = *input.Description
	}
	if input.IsActive != nil {
		updates["is_active"] = *input.IsActive
	}

	if len(updates) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "No fields to update"})
	}

	if err := h.db.Model(&webhook).Updates(updates).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update webhook"})
	}

	webhook.WebhookURL = maskWebhookURL(webhook.WebhookURL)
	return c.JSON(webhook)
}

// GetWebhookLogs retrieves delivery logs for a webhook (placeholder — expand with actual log table)
func (h *WebhookHandler) GetWebhookLogs(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"data":  []interface{}{},
		"total": 0,
	})
}

// GetActiveWebhooks retrieves all active webhooks for a user (used internally by notification service)
func GetActiveWebhooks(db *gorm.DB, userID uuid.UUID) ([]models.WebhookConfig, error) {
	var webhooks []models.WebhookConfig
	if err := db.Where("user_id = ? AND is_active = ?", userID, true).
		Find(&webhooks).Error; err != nil {
		return nil, fmt.Errorf("failed to get active webhooks: %w", err)
	}
	return webhooks, nil
}

// maskWebhookURL masks the webhook URL for security — shows only last 8 chars
func maskWebhookURL(rawURL string) string {
	if len(rawURL) <= 20 {
		return "****"
	}
	return rawURL[:20] + "****" + rawURL[len(rawURL)-8:]
}
