package services

import (
	"context"
	"fmt"
	"insight-engine-backend/models"
	"math"
)

// CorrelationService provides correlation analysis
type CorrelationService struct{}

// NewCorrelationService creates a new correlation service
func NewCorrelationService() *CorrelationService {
	return &CorrelationService{}
}

// CalculateCorrelation calculates Pearson correlation between specified columns
func (s *CorrelationService) CalculateCorrelation(ctx context.Context, data []map[string]interface{}, cols []string) ([]models.CorrelationResult, error) {
	if len(data) < 2 || len(cols) < 2 {
		return []models.CorrelationResult{}, nil
	}

	results := []models.CorrelationResult{}

	// Iterate over all unique pairs of columns
	for i := 0; i < len(cols); i++ {
		for j := i + 1; j < len(cols); j++ {
			colA := cols[i]
			colB := cols[j]

			// Extract values for the pair
			valsA, valsB := s.extractPairedValues(data, colA, colB)

			if len(valsA) < 2 {
				continue
			}

			coef := s.pearsonCorrelation(valsA, valsB)

			// Determine strength
			strength := "None"
			absCoef := math.Abs(coef)
			if absCoef > 0.7 {
				strength = "Strong"
			} else if absCoef > 0.4 {
				strength = "Moderate"
			} else if absCoef > 0.2 {
				strength = "Weak"
			}

			results = append(results, models.CorrelationResult{
				VariableA:   colA,
				VariableB:   colB,
				Coefficient: coef,
				Strength:    strength,
			})
		}
	}

	return results, nil
}

func (s *CorrelationService) extractPairedValues(data []map[string]interface{}, colA, colB string) ([]float64, []float64) {
	valsA := []float64{}
	valsB := []float64{}

	for _, row := range data {
		valA, okA := row[colA]
		valB, okB := row[colB]

		if okA && okB {
			fA, errA := s.toFloat(valA)
			fB, errB := s.toFloat(valB)

			if errA == nil && errB == nil {
				valsA = append(valsA, fA)
				valsB = append(valsB, fB)
			}
		}
	}
	return valsA, valsB
}

func (s *CorrelationService) toFloat(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	default:
		return 0, fmt.Errorf("not a number")
	}
}

func (s *CorrelationService) pearsonCorrelation(x, y []float64) float64 {
	n := float64(len(x))
	if n != float64(len(y)) || n == 0 {
		return 0
	}

	sumX, sumY, sumXY, sumXX, sumYY := 0.0, 0.0, 0.0, 0.0, 0.0

	for i := 0; i < len(x); i++ {
		sumX += x[i]
		sumY += y[i]
		sumXY += x[i] * y[i]
		sumXX += x[i] * x[i]
		sumYY += y[i] * y[i]
	}

	numerator := n*sumXY - sumX*sumY
	denominator := math.Sqrt((n*sumXX - sumX*sumX) * (n*sumYY - sumY*sumY))

	if denominator == 0 {
		return 0
	}

	return numerator / denominator
}
