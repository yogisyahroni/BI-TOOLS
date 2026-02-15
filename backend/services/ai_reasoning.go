package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type ReasoningStep struct {
	StepNumber int    `json:"stepNumber"`
	Thought    string `json:"thought"`
	Action     string `json:"action"`
	Parameter  string `json:"parameter,omitempty"`
}

type ReasoningPlan struct {
	OriginalQuery string          `json:"originalQuery"`
	Steps         []ReasoningStep `json:"steps"`
}

type AIReasoningService struct {
	aiService *AIService
}

func NewAIReasoningService(aiService *AIService) *AIReasoningService {
	return &AIReasoningService{
		aiService: aiService,
	}
}

// BreakDownQuery breaks down a complex query into logical steps
func (s *AIReasoningService) BreakDownQuery(ctx context.Context, providerID, userID, query string) (*ReasoningPlan, error) {
	systemPrompt := `You are an expert data analyst and query planner.
Your task is to break down the user's natural language query into a sequence of logical steps required to answer it.
Output MUST be a valid JSON object with a "steps" array.
Each step object must have:
- "stepNumber": integer (1-based index)
- "thought": string (reasoning for this step)
- "action": string (one of: "QueryTable", "Join", "Filter", "Aggregate", "Calculate", "Visualize", "Explain")
- "parameter": string (details for the action, e.g., table names, filter conditions)

Example input: "Show me sales by region for last month"
Example output:
{
  "steps": [
    {"stepNumber": 1, "thought": "Identify the sales data table", "action": "QueryTable", "parameter": "sales"},
    {"stepNumber": 2, "thought": "Filter for dates in the last month", "action": "Filter", "parameter": "date >= last_month_start AND date <= last_month_end"},
    {"stepNumber": 3, "thought": "Group data by region and sum sales amount", "action": "Aggregate", "parameter": "GROUP BY region, SUM(amount)"},
    {"stepNumber": 4, "thought": "Visualize the result as a bar chart", "action": "Visualize", "parameter": "Bar Chart"}
  ]
}
`
	fullPrompt := fmt.Sprintf("%s\n\nUser Query: %s", systemPrompt, query)

	// Call AI Service
	// We assume JSON mode via prompt instructions, but ideally we'd use response_format if provider supports it.
	// For now, we rely on the prompt.
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
		Steps []ReasoningStep `json:"steps"`
	}

	if err := json.Unmarshal([]byte(responseText), &result); err != nil {
		// Fallback: try to just return the text wrapped in a simple plan if parsing fails
		// Or return error. Let's return error for now to enforce JSON.
		return nil, fmt.Errorf("failed to parse AI response as JSON: %w", err)
	}

	return &ReasoningPlan{
		OriginalQuery: query,
		Steps:         result.Steps,
	}, nil
}

func cleanJSONResponse(text string) string {
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	return strings.TrimSpace(text)
}
