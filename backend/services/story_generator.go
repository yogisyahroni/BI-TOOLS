package services

import (
	"context"
	"encoding/json"
	"insight-engine-backend/models"
	"strings"
)

type StoryGeneratorService struct {
	aiService *AIService
}

func NewStoryGeneratorService(aiService *AIService) *StoryGeneratorService {
	return &StoryGeneratorService{aiService: aiService}
}

// DataStory struct for narrative text
type DataStory struct {
	Title     string    `json:"title"`
	Summary   string    `json:"summary"`
	KeyPoints []string  `json:"key_points"`
	Sections  []Section `json:"sections"`
}

type Section struct {
	Heading string `json:"heading"`
	Content string `json:"content"`
}

// GenerateStoryFromDashboard specific function for text narrative
func (s *StoryGeneratorService) GenerateStoryFromDashboard(dashboard *models.Dashboard) (*DataStory, error) {
	// Simple mock implementation for now, or specific prompt logic
	return &DataStory{
		Title:     "Executive Summary: " + dashboard.Name,
		Summary:   "This dashboard provides critical insights into key performance indicators.",
		KeyPoints: []string{"Trend is positive", "Anomalies detected in Q3"},
		Sections: []Section{
			{Heading: "Overview", Content: "Performance is up 15%."},
		},
	}, nil
}

// GenerateSlides generates slides based on dashboard data
func (s *StoryGeneratorService) GenerateSlides(ctx context.Context, dashboard *models.Dashboard, userID string, customPrompt string, providerID string) (*models.SlideDeck, error) {
	// 1. Prepare Context from Dashboard
	dashboardContext := map[string]interface{}{
		"title":       dashboard.Name, // Changed from dashboard.Title to dashboard.Name as per models.Dashboard
		"description": dashboard.Description,
		"widgets":     dashboard.Layout, // Simplified
	}

	// 2. Determine Provider
	if providerID == "" {
		defaultProvider, err := s.aiService.GetDefaultProvider(userID)
		if err != nil {
			return nil, err
		}
		providerID = defaultProvider.ID
	}

	// 3. Call AI Service
	// We ask for a JSON structure representing slides
	systemPrompt := `
You are an expert data storyteller. Create a presentation slide deck based on the provided dashboard context.
Output MUST be a valid JSON object with the following structure:
{
  "title": "Presentation Title",
  "slides": [
    {
      "title": "Slide Title",
      "layout": "title_and_body" | "two_columns" | "blank" | "title_only",
      "bullet_points": ["point 1", "point 2"],
      "speaker_notes": "notes for speaker"
    }
  ]
}
`
	fullPrompt := systemPrompt + "\nUser Prompt: " + customPrompt

	aiReq, err := s.aiService.Generate(ctx, providerID, userID, fullPrompt, dashboardContext)
	if err != nil {
		return nil, err
	}

	// 4. Parse Response
	content := *aiReq.Response
	// Clean up potential markdown code blocks
	content = cleanJSON(content)

	var slideDeck models.SlideDeck
	if err := json.Unmarshal([]byte(content), &slideDeck); err != nil {
		return nil, err
	}

	return &slideDeck, nil
}

func cleanJSON(content string) string {
	// Simple cleanup for markdown code blocks
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "```json") {
		content = content[7:]
	} else if strings.HasPrefix(content, "```") {
		content = content[3:]
	}

	if strings.HasSuffix(content, "```") {
		content = content[:len(content)-3]
	}
	return strings.TrimSpace(content)
}
