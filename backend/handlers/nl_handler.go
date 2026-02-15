package handlers

import (
	"insight-engine-backend/pkg/validator"
	"insight-engine-backend/services"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type NLHandler struct {
	service *services.NLService
}

func NewNLHandler(service *services.NLService) *NLHandler {
	return &NLHandler{service: service}
}

// ParseFilter handles natural language filter parsing
func (h *NLHandler) ParseFilter(c *fiber.Ctx) error {
	type Request struct {
		Text string `json:"text" validate:"required"`
	}
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	if err := validator.GetValidator().ValidateStruct(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	userID := c.Locals("user_id").(string) // Assuming Auth middleware sets this

	filters, err := h.service.ParseNaturalLanguageFilter(c.Context(), req.Text, userID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(filters)
}

// GenerateDashboard handles AI dashboard creation
func (h *NLHandler) GenerateDashboard(c *fiber.Ctx) error {
	type Request struct {
		Text string `json:"text" validate:"required"`
	}
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	if err := validator.GetValidator().ValidateStruct(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	userID := c.Locals("user_id").(string)
	workspaceID := c.Get("X-Workspace-ID") // Or from context

	dashboard, err := h.service.GenerateDashboardFromText(c.Context(), req.Text, userID, workspaceID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(dashboard)
}

// GenerateStory handles data story generation
func (h *NLHandler) GenerateStory(c *fiber.Ctx) error {
	type Request struct {
		Data interface{} `json:"data" validate:"required"`
	}
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	if err := validator.GetValidator().ValidateStruct(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	userID := c.Locals("user_id").(string)

	story, err := h.service.GenerateDataStory(c.Context(), req.Data, userID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"story": story})
}
