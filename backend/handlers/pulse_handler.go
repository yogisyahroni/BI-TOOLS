package handlers

import (
	"fmt"
	"insight-engine-backend/models"
	"insight-engine-backend/services"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type PulseHandler struct {
	pulseService *services.PulseService
}

func NewPulseHandler(pulseService *services.PulseService) *PulseHandler {
	return &PulseHandler{
		pulseService: pulseService,
	}
}

// CreatePulse creates a new pulse
// @Summary Create a new pulse
// @Description Create a new pulse for scheduled screenshots
// @Tags pulses
// @Accept json
// @Produce json
// @Param pulse body models.Pulse true "Pulse object"
// @Success 201 {object} models.Pulse
// @Failure 400 {object} fiber.Map
// @Failure 500 {object} fiber.Map
// @Router /api/v1/pulses [post]
func (h *PulseHandler) CreatePulse(c *fiber.Ctx) error {
	var pulse models.Pulse
	if err := c.BodyParser(&pulse); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	userID, ok := c.Locals("userId").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}
	fmt.Printf("DEBUG: CreatePulse - UserID from context: '%s'\n", userID)
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		fmt.Printf("ERROR: Invalid UserID format: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Invalid User ID in context"})
	}
	pulse.UserID = parsedUserID

	pulse.IsActive = true
	pulse.CreatedAt = time.Now()
	pulse.UpdatedAt = time.Now()

	if err := h.pulseService.CreatePulse(&pulse); err != nil {
		fmt.Printf("ERROR: CreatePulse failed: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(pulse)
}

// GetUserPulses gets all pulses for the current user
// @Summary Get user pulses
// @Description Get all pulses for the current user
// @Tags pulses
// @Produce json
// @Success 200 {array} models.Pulse
// @Failure 500 {object} fiber.Map
// @Router /api/v1/pulses [get]
func (h *PulseHandler) GetUserPulses(c *fiber.Ctx) error {
	userIDStr, ok := c.Locals("userId").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Invalid User ID in context"})
	}

	pulses, err := h.pulseService.GetUserPulses(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(pulses)
}

// TriggerPulse manually triggers a pulse
// @Summary Trigger a pulse
// @Description Manually trigger a pulse execution
// @Tags pulses
// @Param id path string true "Pulse ID"
// @Success 200 {object} fiber.Map
// @Failure 500 {object} fiber.Map
// @Router /api/v1/pulses/{id}/trigger [post]
func (h *PulseHandler) TriggerPulse(c *fiber.Ctx) error {
	pulseIDStr := c.Params("id")
	pulseID, err := uuid.Parse(pulseIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid pulse ID"})
	}

	pulse, err := h.pulseService.GetPulse(pulseID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Pulse not found"})
	}

	// Run in background to avoid blocking response
	go h.pulseService.ExecutePulse(c.Context(), *pulse)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Pulse triggered successfully"})
}
