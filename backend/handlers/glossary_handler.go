package handlers

import (
	"insight-engine-backend/models"
	"insight-engine-backend/services"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type GlossaryHandler struct {
	service *services.GlossaryService
}

func NewGlossaryHandler(service *services.GlossaryService) *GlossaryHandler {
	return &GlossaryHandler{service: service}
}

// CreateTerm handles the creation of a new business term
func (h *GlossaryHandler) CreateTerm(c *fiber.Ctx) error {
	workspaceID := c.Locals("workspace_id").(string)
	var term models.BusinessTerm
	if err := c.BodyParser(&term); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	term.WorkspaceID = workspaceID
	// Default status if not set
	if term.Status == "" {
		term.Status = "draft"
	}

	if err := h.service.CreateTerm(&term); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create term"})
	}

	return c.Status(http.StatusCreated).JSON(term)
}

// GetTerm retrieves a term by ID
func (h *GlossaryHandler) GetTerm(c *fiber.Ctx) error {
	id := c.Params("id")
	term, err := h.service.GetTerm(id)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Term not found"})
	}
	return c.JSON(term)
}

// ListTerms retrieves all terms for the current workspace
func (h *GlossaryHandler) ListTerms(c *fiber.Ctx) error {
	workspaceID := c.Locals("workspace_id").(string)
	terms, err := h.service.ListTerms(workspaceID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list terms"})
	}
	return c.JSON(terms)
}

// UpdateTerm updates an existing term
func (h *GlossaryHandler) UpdateTerm(c *fiber.Ctx) error {
	id := c.Params("id")
	workspaceID := c.Locals("workspace_id").(string)

	// First check ownership/existence
	existing, err := h.service.GetTerm(id)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Term not found"})
	}

	if existing.WorkspaceID != workspaceID {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": "Access denied"})
	}

	var payload models.BusinessTerm
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	payload.ID = id
	payload.WorkspaceID = workspaceID // Ensure workspace doesn't change

	if err := h.service.UpdateTerm(&payload); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update term"})
	}

	return c.JSON(payload)
}

// DeleteTerm deletes a term
func (h *GlossaryHandler) DeleteTerm(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.service.DeleteTerm(id); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete term"})
	}
	return c.JSON(fiber.Map{"message": "Term deleted"})
}
