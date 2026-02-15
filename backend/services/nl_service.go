package services

import (
	"context"
	"encoding/json"
	"fmt"
	"insight-engine-backend/models"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type NLService struct {
	db        *gorm.DB
	aiService *AIService
}

func NewNLService(db *gorm.DB, aiService *AIService) *NLService {
	return &NLService{
		db:        db,
		aiService: aiService,
	}
}

// ParseNaturalLanguageFilter converts "sales last month" to structured filter
func (s *NLService) ParseNaturalLanguageFilter(ctx context.Context, text string, userID string) (map[string]interface{}, error) {
	// 1. Get default AI provider
	provider, err := s.aiService.GetDefaultProvider(userID)
	if err != nil {
		return nil, fmt.Errorf("no active AI provider found: %w", err)
	}

	// 2. Construct Prompt
	prompt := fmt.Sprintf(`
		You are a data filtering assistant. Convert the following natural language query into a JSON object representing filters.
		Query: "%s"
		
		Output Format:
		{
			"date_range": { "start": "YYYY-MM-DD", "end": "YYYY-MM-DD" }, // If applicable
			"filters": { "column": "value", "column": { "operator": "gt", "value": 100 } }
		}
		Return ONLY VALID JSON.
	`, text)

	// 3. Call AI
	aiReq, err := s.aiService.Generate(ctx, provider.ID, userID, prompt, nil)
	if err != nil {
		// For now, if AI fails, we error out. In future, fallback to random/template.
		return nil, err
	}

	// 4. Parse Response
	var result map[string]interface{}

	content := *aiReq.Response
	// Simple cleanup to extract JSON if wrapped in markdown code blocks
	if len(content) > 0 {
		start := 0
		end := len(content)

		// Find start of JSON
		if idx := 0; idx < len(content) {
			if content[idx] == '{' {
				start = idx
			}
		}
		// Better approach: look for first '{'
		for i, c := range content {
			if c == '{' {
				start = i
				break
			}
		}

		// Find end of JSON
		for i := len(content) - 1; i >= 0; i-- {
			if content[i] == '}' {
				end = i + 1
				break
			}
		}

		if start < end {
			jsonPart := content[start:end]
			err = json.Unmarshal([]byte(jsonPart), &result)
			if err != nil {
				return nil, fmt.Errorf("failed to parse AI JSON response: %w. Content: %s", err, content)
			}
		} else {
			return nil, fmt.Errorf("no JSON object found in AI response: %s", content)
		}
	}

	return result, nil
}

// GenerateDashboardFromText creates a dashboard configuration from a description
func (s *NLService) GenerateDashboardFromText(ctx context.Context, text string, userID string, workspaceID string) (*models.Dashboard, error) {
	provider, err := s.aiService.GetDefaultProvider(userID)
	if err != nil {
		return nil, fmt.Errorf("no active AI provider found: %w", err)
	}

	// Prompt to generate dashboard structure
	prompt := fmt.Sprintf(`
		Create a dashboard configuration for: "%s".
		Return a JSON object with:
		- title: string
		- description: string
		- layout: array of objects { x, y, w, h, i }
		- cards: array of objects { title, type (line, bar, metric), query_config }
		
		Return ONLY JSON.
	`, text)

	aiReq, err := s.aiService.Generate(ctx, provider.ID, userID, prompt, nil)
	if err != nil {
		return nil, err
	}

	layoutJSON := []byte("{}")
	if aiReq.Response != nil {
		// Attempt to extract JSON from reference
		content := *aiReq.Response
		// Simple cleanup to extract JSON if wrapped
		start := 0
		end := len(content)
		for i, c := range content {
			if c == '{' {
				start = i
				break
			}
		}
		for i := len(content) - 1; i >= 0; i-- {
			if content[i] == '}' {
				end = i + 1
				break
			}
		}
		if start < end {
			layoutJSON = []byte(content[start:end])
		}
	}

	// Mocking successful parsing for this step since we handled the architecture
	layout := datatypes.JSON(layoutJSON)
	desc := "Generated from prompt: " + text

	dashboard := &models.Dashboard{
		ID:           uuid.New().String(),
		CollectionID: workspaceID, // Assuming workspace mapping for now
		UserID:       userID,
		Name:         "AI Generated Dashboard: " + text,
		Description:  &desc,
		Layout:       &layout,
	}

	// Parse aiReq.Response into dashboard struct fields ideally
	// s.db.Create(dashboard) // Optional: persist immediately or return draft

	return dashboard, nil
}

// GenerateDataStory generates a narrative from data
func (s *NLService) GenerateDataStory(ctx context.Context, data interface{}, userID string) (string, error) {
	provider, err := s.aiService.GetDefaultProvider(userID)
	if err != nil {
		return "", fmt.Errorf("no active AI provider found: %w", err)
	}

	dataJSON, _ := json.Marshal(data)
	prompt := fmt.Sprintf(`
		Analyze the following data and provide a concise business narrative, highlighting key trends and anomalies.
		Data: %s
	`, string(dataJSON))

	aiReq, err := s.aiService.Generate(ctx, provider.ID, userID, prompt, nil)
	if err != nil {
		return "", err
	}

	return *aiReq.Response, nil
}
