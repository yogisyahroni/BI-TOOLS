package handlers

import (
	"encoding/json"
	"fmt"
	"insight-engine-backend/database"
	"insight-engine-backend/models"
	"insight-engine-backend/services"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/datatypes"
)

// StoryHandler handles story generation and management
type StoryHandler struct {
	storyService *services.StoryGeneratorService
	pptxGen      *services.PPTXGenerator
}

// NewStoryHandler creates a new StoryHandler
func NewStoryHandler(storyService *services.StoryGeneratorService, pptxGen *services.PPTXGenerator) *StoryHandler {
	return &StoryHandler{
		storyService: storyService,
		pptxGen:      pptxGen,
	}
}

// GenerateStoryRequest represents the request body for generating a story
type GenerateStoryRequest struct {
	DashboardID string `json:"dashboard_id"`
	Prompt      string `json:"prompt"`
	ProviderID  string `json:"provider_id"`
}

// CreateStory generates a new story from a dashboard
func (h *StoryHandler) CreateStory(c *fiber.Ctx) error {
	var req GenerateStoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	userID := c.Locals("userID").(string)
	ctx := c.UserContext()

	// 1. Fetch Dashboard
	var dashboard models.Dashboard
	if err := database.DB.Where("id = ?", req.DashboardID).First(&dashboard).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Dashboard not found"})
	}

	// 2. Generate Slides
	slideDeck, err := h.storyService.GenerateSlides(ctx, &dashboard, userID, req.Prompt, req.ProviderID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate story: " + err.Error()})
	}

	// 3. Persist Story
	slideDeckJSON, err := json.Marshal(slideDeck)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to serialize slide deck"})
	}

	story := models.Story{
		UserID:      userID,
		DashboardID: &req.DashboardID,
		Title:       slideDeck.Title,
		Description: slideDeck.Description,
		Content:     datatypes.JSON(slideDeckJSON),
		ProviderID:  req.ProviderID,
		Prompt:      req.Prompt,
	}

	if err := database.DB.Create(&story).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save story"})
	}

	return c.Status(fiber.StatusCreated).JSON(story)
}

// GetStory retrieves a story by ID
func (h *StoryHandler) GetStory(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("userID").(string)

	var story models.Story
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&story).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Story not found"})
	}

	return c.JSON(story)
}

// GetStories retrieves all stories for the user
func (h *StoryHandler) GetStories(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var stories []models.Story
	if err := database.DB.Where("user_id = ?", userID).Order("created_at desc").Find(&stories).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch stories"})
	}

	return c.JSON(stories)
}

// UpdateStory updates a story's content
func (h *StoryHandler) UpdateStory(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("userID").(string)

	var updates struct {
		Title   string         `json:"title"`
		Content datatypes.JSON `json:"content"`
	}

	if err := c.BodyParser(&updates); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	var story models.Story
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&story).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Story not found"})
	}

	story.Title = updates.Title
	story.Content = updates.Content
	story.UpdatedAt = time.Now()

	if err := database.DB.Save(&story).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update story"})
	}

	return c.JSON(story)
}

// ExportPPTX exports a story as a .pptx file
func (h *StoryHandler) ExportPPTX(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("userID").(string)

	var story models.Story
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&story).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Story not found"})
	}

	var deck models.SlideDeck
	if err := json.Unmarshal(story.Content, &deck); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse story content"})
	}

	pptxBytes, err := h.pptxGen.GeneratePPTX(&deck)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate PPTX: " + err.Error()})
	}

	filename := fmt.Sprintf("%s.pptx", slugify(story.Title))
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.presentationml.presentation")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	return c.Send(pptxBytes)
}

// DeleteStory deletes a story
func (h *StoryHandler) DeleteStory(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("userID").(string)

	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Story{}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete story"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// Helper to sanitize filename
func slugify(s string) string {
	// Simple replacement for now, could be more robust
	return s // Keeping it simple as backend logic might not have slug library
}
