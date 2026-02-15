package handlers

import (
	"insight-engine-backend/models"
	"insight-engine-backend/services"

	"github.com/gofiber/fiber/v2"
)

// AnalyticsHandler handles analytics related requests
type AnalyticsHandler struct {
	insightsService    *services.InsightsService
	correlationService *services.CorrelationService
}

// NewAnalyticsHandler creates a new analytics handler
func NewAnalyticsHandler(insightsService *services.InsightsService, correlationService *services.CorrelationService) *AnalyticsHandler {
	return &AnalyticsHandler{
		insightsService:    insightsService,
		correlationService: correlationService,
	}
}

// GenerateInsights generates insights from data
// @Summary Generate insights from data
// @Description Analyzes the provided data and returns a list of insights including trends, anomalies, and descriptive statistics.
// @Tags analytics
// @Accept json
// @Produce json
// @Param request body models.GenerateInsightsRequest true "Data and configuration"
// @Success 200 {array} models.Insight
// @Router /api/analytics/insights [post]
func (h *AnalyticsHandler) GenerateInsights(c *fiber.Ctx) error {
	var req models.GenerateInsightsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if len(req.Data) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No data provided"})
	}

	if req.MetricCol == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "metricCol is required"})
	}

	insights, err := h.insightsService.GenerateInsights(c.UserContext(), req.Data, req.MetricCol, req.TimeCol)
	if err != nil {
		services.LogError("generate_insights_error", "Failed to generate insights", map[string]interface{}{"error": err.Error()})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate insights"})
	}

	return c.JSON(insights)
}

// CalculateCorrelation calculates correlation between columns
// @Summary Calculate correlation between columns
// @Description Calculates Pearson correlation coefficient between specified columns in the dataset.
// @Tags analytics
// @Accept json
// @Produce json
// @Param request body models.CalculateCorrelationRequest true "Data and columns"
// @Success 200 {array} models.CorrelationResult
// @Router /api/analytics/correlations [post]
func (h *AnalyticsHandler) CalculateCorrelation(c *fiber.Ctx) error {
	var req models.CalculateCorrelationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if len(req.Data) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No data provided"})
	}

	if len(req.Cols) < 2 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "At least 2 columns required for correlation"})
	}

	results, err := h.correlationService.CalculateCorrelation(c.UserContext(), req.Data, req.Cols)
	if err != nil {
		services.LogError("calculate_correlation_error", "Failed to calculate correlation", map[string]interface{}{"error": err.Error()})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to calculate correlation"})
	}

	return c.JSON(results)
}
