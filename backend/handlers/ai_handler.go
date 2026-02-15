package handlers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"insight-engine-backend/database"
	"insight-engine-backend/models"
	"insight-engine-backend/pkg/validator"
	"insight-engine-backend/services"

	"github.com/gofiber/fiber/v2"
)

// AIHandler handles AI generation operations
type AIHandler struct {
	aiService             *services.AIService
	aiReasoningService    *services.AIReasoningService
	aiOptimizerService    *services.AIOptimizerService
	storyGeneratorService *services.StoryGeneratorService
}

// NewAIHandler creates a new AI handler
func NewAIHandler(aiService *services.AIService, aiReasoningService *services.AIReasoningService, aiOptimizerService *services.AIOptimizerService, storyGeneratorService *services.StoryGeneratorService) *AIHandler {
	return &AIHandler{
		aiService:             aiService,
		aiReasoningService:    aiReasoningService,
		aiOptimizerService:    aiOptimizerService,
		storyGeneratorService: storyGeneratorService,
	}
}

// Generate generates content using AI
func (h *AIHandler) Generate(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var input struct {
		ProviderID *string                `json:"providerId"` // Optional, uses default if not provided
		Prompt     string                 `json:"prompt" validate:"required"`
		Context    map[string]interface{} `json:"context"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := validator.GetValidator().ValidateStruct(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Get provider ID
	providerID := ""
	if input.ProviderID != nil {
		providerID = *input.ProviderID
	} else {
		// Use default provider
		provider, err := h.aiService.GetDefaultProvider(userID)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "No default provider found. Please set up an AI provider first."})
		}
		providerID = provider.ID
	}

	// Generate content
	aiRequest, err := h.aiService.Generate(c.Context(), providerID, userID, input.Prompt, input.Context)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(aiRequest)
}

// GetRequests gets AI request history
func (h *AIHandler) GetRequests(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	// Get query params
	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)
	providerID := c.Query("providerId")

	query := database.DB.Where("user_id = ?", userID)

	if providerID != "" {
		query = query.Where("provider_id = ?", providerID)
	}

	var requests []models.AIRequest
	var total int64

	query.Model(&models.AIRequest{}).Count(&total)
	query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&requests)

	return c.JSON(fiber.Map{
		"data":   requests,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetRequest gets a single AI request
func (h *AIHandler) GetRequest(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	id := c.Params("id")

	var request models.AIRequest
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&request).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Request not found"})
	}

	return c.JSON(request)
}

// GetUsageStats gets usage statistics
func (h *AIHandler) GetUsageStats(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var stats struct {
		TotalRequests   int64   `json:"totalRequests"`
		TotalTokens     int64   `json:"totalTokens"`
		TotalCost       float64 `json:"totalCost"`
		SuccessfulReqs  int64   `json:"successfulRequests"`
		FailedReqs      int64   `json:"failedRequests"`
		AvgTokensPerReq float64 `json:"avgTokensPerRequest"`
	}

	// Get total requests
	database.DB.Model(&models.AIRequest{}).Where("user_id = ?", userID).Count(&stats.TotalRequests)

	// Get successful/failed counts
	database.DB.Model(&models.AIRequest{}).Where("user_id = ? AND status = ?", userID, models.RequestStatusSuccess).Count(&stats.SuccessfulReqs)
	database.DB.Model(&models.AIRequest{}).Where("user_id = ? AND status = ?", userID, models.RequestStatusError).Count(&stats.FailedReqs)

	// Get sum of tokens and cost
	var result struct {
		TotalTokens int64
		TotalCost   float64
	}
	database.DB.Model(&models.AIRequest{}).
		Select("COALESCE(SUM(tokens_used), 0) as total_tokens, COALESCE(SUM(cost), 0) as total_cost").
		Where("user_id = ?", userID).
		Scan(&result)

	stats.TotalTokens = result.TotalTokens
	stats.TotalCost = result.TotalCost

	// Calculate average
	if stats.TotalRequests > 0 {
		stats.AvgTokensPerReq = float64(stats.TotalTokens) / float64(stats.TotalRequests)
	}

	return c.JSON(stats)
}

// Reason breaks down a complex query
func (h *AIHandler) Reason(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var input struct {
		ProviderID *string `json:"providerId"`
		Query      string  `json:"query" validate:"required"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := validator.GetValidator().ValidateStruct(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	providerID := ""
	if input.ProviderID != nil {
		providerID = *input.ProviderID
	} else {
		provider, err := h.aiService.GetDefaultProvider(userID)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "No default provider found. Please set up an AI provider first."})
		}
		providerID = provider.ID
	}

	plan, err := h.aiReasoningService.BreakDownQuery(c.Context(), providerID, userID, input.Query)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(plan)
}

// Optimize optimizes a SQL query
func (h *AIHandler) Optimize(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var input struct {
		ProviderID *string `json:"providerId"`
		Query      string  `json:"query" validate:"required"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := validator.GetValidator().ValidateStruct(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	providerID := ""
	if input.ProviderID != nil {
		providerID = *input.ProviderID
	} else {
		provider, err := h.aiService.GetDefaultProvider(userID)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "No default provider found. Please set up an AI provider first."})
		}
		providerID = provider.ID
	}

	suggestions, err := h.aiOptimizerService.OptimizeQuery(c.Context(), providerID, userID, input.Query)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(suggestions)
}

// StreamGenerate generates content using AI and streams the response
func (h *AIHandler) StreamGenerate(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var input struct {
		ProviderID *string                `json:"providerId"`
		Prompt     string                 `json:"prompt" validate:"required"`
		Context    map[string]interface{} `json:"context"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := validator.GetValidator().ValidateStruct(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Get provider ID
	providerID := ""
	if input.ProviderID != nil {
		providerID = *input.ProviderID
	} else {
		// Use default provider
		provider, err := h.aiService.GetDefaultProvider(userID)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "No default provider found. Please set up an AI provider first."})
		}
		providerID = provider.ID
	}

	// Get stream channel
	streamChan, err := h.aiService.StreamGenerate(c.Context(), providerID, userID, input.Prompt, input.Context)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Set headers for SSE
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		for resp := range streamChan {
			data, err := json.Marshal(resp)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
			if err := w.Flush(); err != nil {
				// Connection likely closed by client
				return
			}
		}
	})

	return nil
}

// GeneratePresentation generates a slide deck based on dashboard data
func (h *AIHandler) GeneratePresentation(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var input struct {
		DashboardID string  `json:"dashboardId" validate:"required"`
		Prompt      string  `json:"prompt"`
		ProviderID  *string `json:"providerId"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := validator.GetValidator().ValidateStruct(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Retrieve Dashboard Metadata
	var dashboard models.Dashboard
	if err := database.DB.Where("id = ?", input.DashboardID).First(&dashboard).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Dashboard not found"})
	}

	providerID := ""
	if input.ProviderID != nil {
		providerID = *input.ProviderID
	}

	// Call Service
	slideDeck, err := h.storyGeneratorService.GenerateSlides(c.Context(), &dashboard, userID, input.Prompt, providerID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(slideDeck)
}
