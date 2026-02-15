package models

import "time"

// InsightType defines the type of insight
type InsightType string

const (
	InsightTypeTrend       InsightType = "trend"
	InsightTypeAnomaly     InsightType = "anomaly"
	InsightTypeCorrelation InsightType = "correlation"
	InsightTypeDescriptive InsightType = "descriptive"
)

// Insight represents an automatically generated data insight
type Insight struct {
	ID          string                 `json:"id"`
	Type        InsightType            `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Metric      string                 `json:"metric"`
	Value       interface{}            `json:"value"`
	Confidence  float64                `json:"confidence"` // 0.0 to 1.0
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"createdAt"`
}

// CorrelationResult represents a correlation analysis result
type CorrelationResult struct {
	VariableA    string  `json:"variableA"`
	VariableB    string  `json:"variableB"`
	Coefficient  float64 `json:"coefficient"`  // -1.0 to 1.0
	Strength     string  `json:"strength"`     // Strong, Moderate, Weak, None
	Significance float64 `json:"significance"` // p-value (optional/simulated)
}

// GenerateInsightsRequest represents the request payload for insights generation
type GenerateInsightsRequest struct {
	Data      []map[string]interface{} `json:"data"`
	MetricCol string                   `json:"metricCol"`
	TimeCol   string                   `json:"timeCol,omitempty"`
}

// CalculateCorrelationRequest represents the request payload for correlation
type CalculateCorrelationRequest struct {
	Data []map[string]interface{} `json:"data"`
	Cols []string                 `json:"cols"` // Columns to correlate
}
