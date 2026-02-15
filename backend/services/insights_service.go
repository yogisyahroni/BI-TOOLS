package services

import (
	"context"
	"fmt"
	"insight-engine-backend/models"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
)

// InsightsService provides automated data insights
type InsightsService struct{}

// NewInsightsService creates a new insights service
func NewInsightsService() *InsightsService {
	return &InsightsService{}
}

// GenerateInsights analyzes data and returns a list of insights
func (s *InsightsService) GenerateInsights(ctx context.Context, data []map[string]interface{}, metricCol string, timeCol string) ([]models.Insight, error) {
	if len(data) < 2 {
		return []models.Insight{}, nil
	}

	var values []float64
	var times []time.Time

	// Extract data
	for _, row := range data {
		if val, ok := row[metricCol]; ok {
			// Try to convert to float64
			var floatVal float64
			switch v := val.(type) {
			case float64:
				floatVal = v
			case int:
				floatVal = float64(v)
			case int64:
				floatVal = float64(v)
			default:
				continue // Skip non-numeric
			}
			values = append(values, floatVal)

			if timeCol != "" {
				if timeVal, ok := row[timeCol]; ok {
					// Parsing logic here simplified to basic string or time.Time assumption
					// In production, would be more robust
					if t, ok := timeVal.(time.Time); ok {
						times = append(times, t)
					} else if tStr, ok := timeVal.(string); ok {
						if parsed, err := time.Parse(time.RFC3339, tStr); err == nil {
							times = append(times, parsed)
						}
					}
				}
			}
		}
	}

	if len(values) < 2 {
		return []models.Insight{}, nil
	}

	insights := []models.Insight{}

	// 1. Trend Analysis (Simple Linear Regression)
	slope, intercept := s.calculateTrend(values)
	trendInsight := s.generateTrendInsight(slope, intercept, values, metricCol)
	if trendInsight != nil {
		insights = append(insights, *trendInsight)
	}

	// 2. Anomaly Detection (Simple Z-Score)
	anomalies := s.detectAnomalies(values)
	if len(anomalies) > 0 {
		// Group anomalies if too many
		if len(anomalies) > 3 {
			insights = append(insights, models.Insight{
				ID:          uuid.New().String(),
				Type:        models.InsightTypeAnomaly,
				Title:       fmt.Sprintf("Multiple Anomalies Detected"),
				Description: fmt.Sprintf("Found %d data points that deviate significantly from the norm.", len(anomalies)),
				Metric:      metricCol,
				Confidence:  0.9,
				CreatedAt:   time.Now(),
			})
		} else {
			for _, idx := range anomalies {
				insights = append(insights, models.Insight{
					ID:          uuid.New().String(),
					Type:        models.InsightTypeAnomaly,
					Title:       "Anomaly Detected",
					Description: fmt.Sprintf("Value %.2f is an outlier (Z-Score > 3).", values[idx]),
					Metric:      metricCol,
					Value:       values[idx],
					Confidence:  0.95,
					CreatedAt:   time.Now(),
				})
			}
		}
	}

	// 3. Descriptive Stats
	statsInsight := s.generateDescriptiveStats(values, metricCol)
	if statsInsight != nil {
		insights = append(insights, *statsInsight)
	}

	return insights, nil
}

// calculateTrend performs simple linear regression (OLS)
// Returns slope and intercept
func (s *InsightsService) calculateTrend(y []float64) (float64, float64) {
	n := float64(len(y))
	if n < 2 {
		return 0, 0
	}

	sumX, sumY, sumXY, sumXX := 0.0, 0.0, 0.0, 0.0

	for i, val := range y {
		x := float64(i)
		sumX += x
		sumY += val
		sumXY += x * val
		sumXX += x * x
	}

	slope := (n*sumXY - sumX*sumY) / (n*sumXX - sumX*sumX)
	intercept := (sumY - slope*sumX) / n

	return slope, intercept
}

func (s *InsightsService) generateTrendInsight(slope, intercept float64, values []float64, metric string) *models.Insight {
	// Determine trend direction and magnitude
	firstVal := values[0]
	lastVal := values[len(values)-1]

	// Avoid division by zero
	if firstVal == 0 {
		firstVal = 0.0001
	}

	percentChange := ((lastVal - firstVal) / math.Abs(firstVal)) * 100

	var title, description string
	var confidence float64 = 0.8

	if math.Abs(percentChange) < 5 {
		title = "Stable Trend"
		description = fmt.Sprintf("%s has remained relatively stable (%.1f%% change).", metric, percentChange)
	} else if percentChange > 0 {
		title = "Upward Trend"
		description = fmt.Sprintf("%s is trending upwards by %.1f%% over the period.", metric, percentChange)
	} else {
		title = "Downward Trend"
		description = fmt.Sprintf("%s is trending downwards by %.1f%% over the period.", metric, math.Abs(percentChange))
	}

	return &models.Insight{
		ID:          uuid.New().String(),
		Type:        models.InsightTypeTrend,
		Title:       title,
		Description: description,
		Metric:      metric,
		Value:       slope,
		Confidence:  confidence,
		CreatedAt:   time.Now(),
	}
}

func (s *InsightsService) detectAnomalies(values []float64) []int {
	n := float64(len(values))
	if n < 4 {
		return []int{}
	}

	// Calculate Mean
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / n

	// Calculate StdDev
	varianceSum := 0.0
	for _, v := range values {
		varianceSum += math.Pow(v-mean, 2)
	}
	stdDev := math.Sqrt(varianceSum / n)

	if stdDev == 0 {
		return []int{}
	}

	anomalies := []int{}
	for i, v := range values {
		zScore := (v - mean) / stdDev
		if math.Abs(zScore) > 3 { // 3 sigma
			anomalies = append(anomalies, i)
		}
	}

	return anomalies
}

func (s *InsightsService) generateDescriptiveStats(values []float64, metric string) *models.Insight {
	sort.Float64s(values)
	min := values[0]
	max := values[len(values)-1]

	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))

	return &models.Insight{
		ID:          uuid.New().String(),
		Type:        models.InsightTypeDescriptive,
		Title:       "Summary Statistics",
		Description: fmt.Sprintf("Average: %.2f, Min: %.2f, Max: %.2f", mean, min, max),
		Metric:      metric,
		Value:       mean,
		Confidence:  1.0,
		CreatedAt:   time.Now(),
	}
}
