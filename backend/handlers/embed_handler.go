package handlers

import (
	"insight-engine-backend/dtos"
	"insight-engine-backend/services"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type EmbedHandler struct {
	embedService *services.EmbedService
}

func NewEmbedHandler(embedService *services.EmbedService) *EmbedHandler {
	return &EmbedHandler{
		embedService: embedService,
	}
}

// GenerateToken handles the request to create an embed token
// @Summary Generate Embed Token
// @Description Generates a signed JWT for embedding a dashboard
// @Tags Embed
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dtos.EmbedTokenRequest true "Embed Token Request"
// @Success 200 {object} dtos.EmbedTokenResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /embed/token [post]
func (h *EmbedHandler) GenerateToken(c *fiber.Ctx) error {
	var req dtos.EmbedTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validation (Basic)
	if req.DashboardID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "dashboard_id is required",
		})
	}

	resp, err := h.embedService.GenerateEmbedToken(req)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(resp)
}

// ValidateToken validates the token and returns the dashboard configuration
// @Summary Validate Embed Token
// @Description Validates the signed JWT and returns the dashboard configuration
// @Tags Embed
// @Accept json
// @Produce json
// @Param token query string true "Embed Token"
// @Success 200 {object} models.Dashboard
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /embed/token/validate [get]
func (h *EmbedHandler) ValidateToken(c *fiber.Ctx) error {
	token := c.Query("token")
	if token == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "token query parameter is required",
		})
	}

	dashboard, err := h.embedService.ValidateEmbedToken(token)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Invalid or expired token",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    dashboard,
	})
}
