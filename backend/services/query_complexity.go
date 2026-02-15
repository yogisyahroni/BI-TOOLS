package services

import (
	"time"

	"insight-engine-backend/models"
)

// QueryComplexityService handles invalidation and calculation of query complexity
type QueryComplexityService struct{}

// NewQueryComplexityService creates a new instance of QueryComplexityService
func NewQueryComplexityService() *QueryComplexityService {
	return &QueryComplexityService{}
}

// ComplexityLevel represents the categorization of query complexity
type ComplexityLevel string

const (
	ComplexitySimple  ComplexityLevel = "SIMPLE"
	ComplexityMedium  ComplexityLevel = "MEDIUM"
	ComplexityComplex ComplexityLevel = "COMPLEX"
	ComplexityHeavy   ComplexityLevel = "HEAVY"
)

// ComplexityResult holds the result of the complexity calculation
type ComplexityResult struct {
	Score       int
	Level       ComplexityLevel
	Timeout     time.Duration
	Description string
}

// CalculateComplexity analyzes the VisualQueryConfig and determines the appropriate timeout
func (s *QueryComplexityService) CalculateComplexity(config *models.VisualQueryConfig) ComplexityResult {
	score := 0

	// Base score for any query
	score += 1

	// Add points for tables involved
	score += len(config.Tables)

	// Add points for joins (joins are expensive)
	score += len(config.Joins) * 2

	// Add points for aggregations
	score += len(config.Aggregations)

	// Add points for grouping
	if len(config.GroupBy) > 0 {
		score += len(config.GroupBy)
	}

	// Add points for sorting
	if len(config.OrderBy) > 0 {
		score += 1
	}

	// Determine timeouts based on score buckets
	var level ComplexityLevel
	var timeout time.Duration
	var desc string

	switch {
	case score < 5:
		level = ComplexitySimple
		timeout = 30 * time.Second
		desc = "Simple query"
	case score < 10:
		level = ComplexityMedium
		timeout = 60 * time.Second
		desc = "Medium complexity query"
	case score < 20:
		level = ComplexityComplex
		timeout = 120 * time.Second
		desc = "Complex query with multiple joins/aggregations"
	default:
		level = ComplexityHeavy
		timeout = 300 * time.Second
		desc = "Heavy analytical query"
	}

	return ComplexityResult{
		Score:       score,
		Level:       level,
		Timeout:     timeout,
		Description: desc,
	}
}
