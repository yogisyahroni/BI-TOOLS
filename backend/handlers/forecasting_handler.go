package handlers

import (
	"insight-engine-backend/models"
	"insight-engine-backend/services"

	"github.com/gofiber/fiber/v2"
)

type ForecastingHandler struct {
	service *services.ForecastingService
}

func NewForecastingHandler(service *services.ForecastingService) *ForecastingHandler {
	return &ForecastingHandler{service: service}
}

func (h *ForecastingHandler) Forecast(c *fiber.Ctx) error {
	var req models.ForecastRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate request defaults
	if req.Horizon <= 0 {
		req.Horizon = 12
	}
	if req.ModelType == "" {
		req.ModelType = "linear"
	}

	result, err := h.service.Forecast(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(result)
}
