package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

type AIOptimizationSuggestion struct {
	Suggestion string `json:"suggestion"`
	Rationale  string `json:"rationale"`
	Impact     string `json:"impact"` // High, Medium, Low
}

type AIOptimizerService struct {
	aiService *AIService
}

func NewAIOptimizerService(aiService *AIService) *AIOptimizerService {
	return &AIOptimizerService{aiService: aiService}
}

// OptimizeQuery analyzes a SQL query and suggests optimizations
func (s *AIOptimizerService) OptimizeQuery(ctx context.Context, providerID, userID, sqlQuery string) ([]AIOptimizationSuggestion, error) {
	systemPrompt := `You are an expert database administrator and SQL optimizer.
Analyze the following SQL query for performance improvements, potential bottlenecks, and adherence to best practices.
Return a JSON object with a "suggestions" array.
Each suggestion object must have:
- "suggestion": string (the specific improvement or rewritten query part)
- "rationale": string (why this improves performance)
- "impact": string (one of: "High", "Medium", "Low")

If the query is already optimal, return an empty array.
Output MUST be valid JSON.`

	fullPrompt := fmt.Sprintf("%s\n\nSQL Query:\n%s", systemPrompt, sqlQuery)

	// Call AI Service
	aiReq, err := s.aiService.Generate(ctx, providerID, userID, fullPrompt, nil)
	if err != nil {
		return nil, err
	}

	if aiReq.Response == nil {
		return nil, errors.New("empty response from AI")
	}

	responseText := *aiReq.Response

	// Attempt to clean markdown code blocks if present
	responseText = cleanJSONResponse(responseText)

	var result struct {
		Suggestions []AIOptimizationSuggestion `json:"suggestions"`
	}

	if err := json.Unmarshal([]byte(responseText), &result); err != nil {
		return nil, fmt.Errorf("failed to parse AI response as JSON: %w", err)
	}

	return result.Suggestions, nil
}
