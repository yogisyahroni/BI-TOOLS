package handlers

import (
	"insight-engine-backend/models"
	"insight-engine-backend/services"

	"github.com/gofiber/fiber/v2"
)

// CommentHandler handles comment-related requests
type CommentHandler struct {
	commentService *services.CommentService
}

// NewCommentHandler creates a new comment handler
func NewCommentHandler(commentService *services.CommentService) *CommentHandler {
	return &CommentHandler{
		commentService: commentService,
	}
}

// RegisterRoutes registers all comment routes
func (h *CommentHandler) RegisterRoutes(api fiber.Router) {
	comments := api.Group("/comments")

	comments.Get("/", h.GetComments)
	comments.Post("/", h.CreateComment)
	comments.Put("/:id", h.UpdateComment)
	comments.Delete("/:id", h.DeleteComment)
}

// GetComments returns comments for a specific entity
func (h *CommentHandler) GetComments(c *fiber.Ctx) error {
	entityType := c.Query("entityType")
	entityID := c.Query("entityId")

	if entityType == "" || entityID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "entityType and entityId are required"})
	}

	// Validate entity type
	validEntityType, valid := models.ValidateEntityType(entityType)
	if !valid {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid entity type"})
	}

	filter := &models.CommentFilter{
		EntityType: &validEntityType,
		EntityID:   &entityID,
	}

	comments, total, err := h.commentService.GetCommentsByEntity(c.Locals("userID").(string), filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch comments"})
	}

	return c.JSON(fiber.Map{
		"comments": comments,
		"total":    total,
	})
}

// CreateComment creates a new comment
func (h *CommentHandler) CreateComment(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var input struct {
		EntityType string `json:"entityType"` // 'pipeline', 'dataflow', 'collection'
		EntityID   string `json:"entityId"`
		Content    string `json:"content"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.Content == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Content is required"})
	}

	// Validate and convert entity type
	entityType, valid := models.ValidateEntityType(input.EntityType)
	if !valid {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid entity type"})
	}

	req := &models.CommentCreateRequest{
		EntityType: string(entityType),
		EntityID:   input.EntityID,
		Content:    input.Content,
	}

	comment, err := h.commentService.CreateComment(userID, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(comment)
}

// UpdateComment updates a comment (only by owner)
func (h *CommentHandler) UpdateComment(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("userID").(string)

	var input struct {
		Content *string `json:"content"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.Content == nil || *input.Content == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Content is required"})
	}

	comment, err := h.commentService.UpdateComment(id, userID, *input.Content)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(comment)
}

// DeleteComment deletes a comment (only by owner)
func (h *CommentHandler) DeleteComment(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("userID").(string)

	if err := h.commentService.DeleteComment(id, userID); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Comment not found or access denied"})
	}

	return c.JSON(fiber.Map{"message": "Comment deleted successfully"})
}
