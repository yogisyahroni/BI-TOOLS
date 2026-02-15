package handlers

import (
	"insight-engine-backend/services"

	"github.com/gofiber/fiber/v2"
)

type AnomalyHandler struct {
	service *services.AnomalyDetectionService
}

func NewAnomalyHandler(service *services.AnomalyDetectionService) *AnomalyHandler {
	return &AnomalyHandler{service: service}
}

func (h *AnomalyHandler) DetectAnomalies(c *fiber.Ctx) error {
	var req services.AnomalyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	result, err := h.service.DetectAnomalies(req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(result)
}
