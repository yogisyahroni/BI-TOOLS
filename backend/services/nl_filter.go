package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type NLFilterService struct {
	aiService *AIService
}

func NewNLFilterService(aiService *AIService) *NLFilterService {
	return &NLFilterService{aiService: aiService}
}

type FilterIntent struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
	Logic    string      `json:"logic,omitempty"` // AND, OR
}

// ParseFilter converts natural language text into a structured filter
func (s *NLFilterService) ParseFilter(ctx context.Context, text string, userID string, contextData []string) ([]FilterIntent, error) {
	provider, err := s.aiService.GetDefaultProvider(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI provider: %v", err)
	}

	prompt := fmt.Sprintf(`
You are a data analyst helper. Convert the following natural language filter request into a JSON array of filter objects.
Available fields: %s
Request: "%s"

Output format:
[
  {"field": "field_name", "operator": "=", "value": "value", "logic": "AND"}
]
Supported operators: =, !=, >, <, >=, <=, LIKE, IN, BETWEEN
Return ONLY the JSON array.
`, strings.Join(contextData, ", "), text)

	// Context for AI
	aiCtx := map[string]interface{}{
		"type": "filter_generation",
	}

	aiReq, err := s.aiService.Generate(ctx, provider.ID, userID, prompt, aiCtx)
	if err != nil {
		return nil, err
	}

	if aiReq == nil || aiReq.Response == nil {
		return nil, fmt.Errorf("empty response from AI")
	}

	response := *aiReq.Response

	// Clean up response (remove markdown code blocks if any)
	cleanResponse := strings.TrimPrefix(response, "```json")
	cleanResponse = strings.TrimPrefix(cleanResponse, "```")
	cleanResponse = strings.TrimSuffix(cleanResponse, "```")
	cleanResponse = strings.TrimSpace(cleanResponse)

	var filters []FilterIntent
	err = json.Unmarshal([]byte(cleanResponse), &filters)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %v", err)
	}

	return filters, nil
}
