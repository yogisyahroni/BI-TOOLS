package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"insight-engine-backend/database"
	"insight-engine-backend/models"
	"insight-engine-backend/services"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
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

// DataBinding represents the connection to a Dashboard or Card
type DataBinding struct {
	DashboardID *string `json:"dashboard_id,omitempty"`
	CardID      *string `json:"card_id,omitempty"`
}

// ManualSlide is the request shape for a single slide in manual creation
type ManualSlide struct {
	Title       string       `json:"title"`
	Content     string       `json:"content"`
	Layout      string       `json:"layout"` // title | bullet_points | image_text | chart
	Notes       string       `json:"notes"`
	DataBinding *DataBinding `json:"data_binding,omitempty"`
}

// CreateManualStoryRequest represents the request body for a manual (no-AI) story
type CreateManualStoryRequest struct {
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Slides      []ManualSlide `json:"slides"`
}

// CreateStory generates a new story from a dashboard
func (h *StoryHandler) CreateStory(c *fiber.Ctx) error {
	var req GenerateStoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	userIDVal := c.Locals("userID")
	if userIDVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized: User ID missing"})
	}
	userID, ok := userIDVal.(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized: Invalid User ID"})
	}

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

// CreateManualStory creates a blank Story without AI generation.
// POST /stories/manual
func (h *StoryHandler) CreateManualStory(c *fiber.Ctx) error {
	var req CreateManualStoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "title is required"})
	}

	userIDVal := c.Locals("userID")
	if userIDVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized: User ID missing"})
	}
	userID, ok := userIDVal.(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized: Invalid User ID"})
	}

	// Ensure the user exists in the users table to satisfy the FK constraint.
	// Users authenticated via SSO/OAuth might have a valid JWT but not yet a row in users.
	userUUID, uuidErr := uuid.Parse(userID)
	if uuidErr == nil {
		email, _ := c.Locals("userEmail").(string)
		if email == "" {
			email = userID + "@sso.local" // fallback if email not in claims
		}
		// Use raw SQL INSERT ... ON CONFLICT DO NOTHING so existing users are not overwritten
		database.DB.Exec(
			`INSERT INTO users (id, email, username, name, role, status, email_verified, created_at, updated_at)
			 VALUES (?, ?, ?, ?, 'user', 'active', true, NOW(), NOW())
			 ON CONFLICT (id) DO NOTHING`,
			userUUID, email, userID, email,
		)
	}

	// Build the SlideDeck content from the request
	type slideContent struct {
		Title       string       `json:"title"`
		Content     string       `json:"content"`
		Layout      string       `json:"layout"`
		Notes       string       `json:"notes"`
		DataBinding *DataBinding `json:"data_binding,omitempty"`
	}
	type slideDeckContent struct {
		Slides []slideContent `json:"slides"`
	}

	slides := make([]slideContent, 0, len(req.Slides))
	for _, s := range req.Slides {
		layout := s.Layout
		if layout == "" {
			layout = "title"
		}
		slides = append(slides, slideContent{
			Title:       s.Title,
			Content:     s.Content,
			Layout:      layout,
			Notes:       s.Notes,
			DataBinding: s.DataBinding,
		})
	}

	// If no slides supplied, start with one blank title slide
	if len(slides) == 0 {
		slides = append(slides, slideContent{
			Title:   req.Title,
			Content: "",
			Layout:  "title",
			Notes:   "",
		})
	}

	deckContent := slideDeckContent{Slides: slides}
	contentJSON, err := json.Marshal(deckContent)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to serialize content"})
	}

	story := models.Story{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Content:     datatypes.JSON(contentJSON),
	}

	// Use a session that skips association saves so GORM does not try to create/update
	// the User or Dashboard records when inserting a new Story.
	if err := database.DB.Session(&gorm.Session{FullSaveAssociations: false}).Create(&story).Error; err != nil {
		// TEMP DEBUG: print the actual DB error to terminal
		fmt.Printf("[CreateManualStory] DB error: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to save story",
			"detail": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(story)
}

// GetStory retrieves a story by ID
func (h *StoryHandler) GetStory(c *fiber.Ctx) error {
	id := c.Params("id")
	userIDVal := c.Locals("userID")
	if userIDVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized: User ID missing"})
	}
	userID, ok := userIDVal.(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized: Invalid User ID"})
	}

	var story models.Story
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&story).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Story not found"})
	}

	// Enrich story with live data if charts are present
	h.enrichStoryWithLiveData(c.Context(), &story)

	return c.JSON(story)
}

// GetPublicStory retrieves a story by its share token without requiring authentication.
func (h *StoryHandler) GetPublicStory(c *fiber.Ctx) error {
	token := c.Params("token")
	if token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid token"})
	}

	var story models.Story
	if err := database.DB.Where("share_token = ? AND is_public = ?", token, true).First(&story).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Story not found or not public"})
	}

	// Enrich story with live data if charts are present
	h.enrichStoryWithLiveData(c.Context(), &story)

	return c.JSON(story)
}

// TogglePublicShare enables or disables public sharing for a story.
func (h *StoryHandler) TogglePublicShare(c *fiber.Ctx) error {
	id := c.Params("id")
	userIDVal := c.Locals("userID")
	if userIDVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized: User ID missing"})
	}
	userID, ok := userIDVal.(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized: Invalid User ID"})
	}

	var body struct {
		IsPublic bool `json:"is_public"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	var story models.Story
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&story).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Story not found"})
	}

	story.IsPublic = body.IsPublic
	if story.IsPublic && story.ShareToken == "" {
		// Generate a new secure token for sharing using uuid
		story.ShareToken = uuid.New().String()
	}

	if err := database.DB.Save(&story).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update sharing settings"})
	}

	return c.JSON(fiber.Map{
		"is_public":   story.IsPublic,
		"share_token": story.ShareToken,
	})
}

// GetStories retrieves all stories for the user
func (h *StoryHandler) GetStories(c *fiber.Ctx) error {
	userIDVal := c.Locals("userID")
	if userIDVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized: User ID missing"})
	}
	userID, ok := userIDVal.(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized: Invalid User ID"})
	}

	var stories []models.Story
	if err := database.DB.Where("user_id = ?", userID).Order("created_at desc").Find(&stories).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch stories"})
	}

	return c.JSON(stories)
}

// UpdateStory updates a story's content
func (h *StoryHandler) UpdateStory(c *fiber.Ctx) error {
	id := c.Params("id")
	userIDVal := c.Locals("userID")
	if userIDVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized: User ID missing"})
	}
	userID, ok := userIDVal.(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized: Invalid User ID"})
	}

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
	userIDVal := c.Locals("userID")
	if userIDVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized: User ID missing"})
	}
	userID, ok := userIDVal.(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized: Invalid User ID"})
	}

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
	userIDVal := c.Locals("userID")
	if userIDVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized: User ID missing"})
	}
	userID, ok := userIDVal.(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized: Invalid User ID"})
	}

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

// enrichStoryWithLiveData iterates through a story's slides and injects live query results for charts
func (h *StoryHandler) enrichStoryWithLiveData(ctx context.Context, story *models.Story) {
	var deck struct {
		Slides []map[string]interface{} `json:"slides"`
	}

	if err := json.Unmarshal(story.Content, &deck); err != nil {
		fmt.Printf("[enrichStoryWithLiveData] Error unmarshaling story content: %v\n", err)
		return
	}

	// For each slide, check if it's a chart and has data binding
	for i, slide := range deck.Slides {
		if layout, ok := slide["layout"].(string); ok && layout == "chart" {
			if dataBinding, ok := slide["data_binding"].(map[string]interface{}); ok {
				cardIDRaw, ok1 := dataBinding["card_id"]

				if ok1 && cardIDRaw != nil {
					cardID := cardIDRaw.(string)

					// Fetch the Dashboard Card to get the query and visualization config
					var card models.DashboardCard
					if err := database.DB.Where("id = ?", cardID).Preload("Query").First(&card).Error; err == nil {
						// Inject visualization config
						if card.VisualizationConfig != nil {
							slide["visualization_config"] = card.VisualizationConfig
						} else if card.Query != nil && card.Query.VisualizationConfig != nil {
							slide["visualization_config"] = card.Query.VisualizationConfig
						}

						// Execute Query if it exists
						if card.Query != nil && card.Query.ConnectionID != "" {
							var conn models.Connection
							if err := database.DB.Where("id = ?", card.Query.ConnectionID).First(&conn).Error; err == nil {
								// Instantiate QueryExecutor
								qe := services.NewQueryExecutor(nil, nil, nil) // we don't strictly need query cache here if not initialized, or we can use the app's cache if available. We'll skip cache for now or fetch it from context if we had it.

								limit := 1000 // default presentation limit
								// Note: For public charts, ideally we would use a read-only specialized executor.
								result, err := qe.Execute(ctx, &conn, card.Query.SQL, nil, &limit, nil)
								if err == nil {
									slide["query_result"] = result
								} else {
									fmt.Printf("[enrichStoryWithLiveData] Query execution failed for card %s: %v\n", cardID, err)
									slide["query_error"] = err.Error()
								}
							}
						}
					}
				}
			}
		}
		deck.Slides[i] = slide
	}

	// Re-marshal the enriched content back into the story object
	if enrichedContent, err := json.Marshal(deck); err == nil {
		story.Content = datatypes.JSON(enrichedContent)
	}
}
