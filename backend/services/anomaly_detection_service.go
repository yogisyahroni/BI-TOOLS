package services

import (
	"errors"
	"math"
	"sort"
)

type AnomalyMethod string

const (
	MethodZScore AnomalyMethod = "z-score"
	MethodIQR    AnomalyMethod = "iqr"
)

type AnomalyPoint struct {
	Timestamp string  `json:"timestamp"`
	Value     float64 `json:"value"`
}

type AnomalyRequest struct {
	Data        []AnomalyPoint `json:"data"`
	Method      AnomalyMethod  `json:"method"`
	Sensitivity float64        `json:"sensitivity"` // Multiplier: 1.5 for IQR, 2.0 or 3.0 for Z-Score
}

type DetectedAnomaly struct {
	Index     int     `json:"index"`
	Timestamp string  `json:"timestamp"`
	Value     float64 `json:"value"`
	Score     float64 `json:"score"`    // Z-Score or deviations from median
	Severity  string  `json:"severity"` // "low", "medium", "high"
	Expected  float64 `json:"expected"` // meaningful for Z-Score (mean), for IQR (median)
}

type AnomalyResult struct {
	Anomalies []DetectedAnomaly `json:"anomalies"`
	Summary   string            `json:"summary"`
}

type AnomalyDetectionService struct{}

func NewAnomalyDetectionService() *AnomalyDetectionService {
	return &AnomalyDetectionService{}
}

func (s *AnomalyDetectionService) DetectAnomalies(req AnomalyRequest) (*AnomalyResult, error) {
	if len(req.Data) < 3 {
		return nil, errors.New("insufficient data points for anomaly detection (min 3)")
	}

	// default sensitivity
	if req.Sensitivity <= 0 {
		if req.Method == MethodIQR {
			req.Sensitivity = 1.5
		} else {
			req.Sensitivity = 3.0
		}
	}

	values := make([]float64, len(req.Data))
	for i, p := range req.Data {
		values[i] = p.Value
	}

	var anomalies []DetectedAnomaly
	var err error

	switch req.Method {
	case MethodIQR:
		anomalies, err = s.calculateIQR(req.Data, values, req.Sensitivity)
	case MethodZScore:
		fallthrough
	default:
		anomalies, err = s.calculateZScore(req.Data, values, req.Sensitivity)
	}

	if err != nil {
		return nil, err
	}

	return &AnomalyResult{
		Anomalies: anomalies,
		Summary:   dataSummary(values),
	}, nil
}

func (s *AnomalyDetectionService) calculateZScore(original []AnomalyPoint, data []float64, threshold float64) ([]DetectedAnomaly, error) {
	mean, stdDev := calculateStats(data)
	if stdDev == 0 {
		return []DetectedAnomaly{}, nil // No variation, no anomalies
	}

	var anomalies []DetectedAnomaly

	for i, val := range data {
		zScore := (val - mean) / stdDev
		if math.Abs(zScore) > threshold {
			severity := "low"
			if math.Abs(zScore) > threshold*1.5 {
				severity = "medium"
			}
			if math.Abs(zScore) > threshold*2.0 {
				severity = "high"
			}

			anomalies = append(anomalies, DetectedAnomaly{
				Index:     i,
				Timestamp: original[i].Timestamp,
				Value:     val,
				Score:     zScore,
				Severity:  severity,
				Expected:  mean,
			})
		}
	}

	return anomalies, nil
}

func (s *AnomalyDetectionService) calculateIQR(original []AnomalyPoint, data []float64, multiplier float64) ([]DetectedAnomaly, error) {
	// Create a copy to sort
	sorted := make([]float64, len(data))
	copy(sorted, data)
	sort.Float64s(sorted)

	q1 := percentile(sorted, 25)
	q3 := percentile(sorted, 75)
	iqr := q3 - q1
	median := percentile(sorted, 50)

	lowerBound := q1 - (multiplier * iqr)
	upperBound := q3 + (multiplier * iqr)

	var anomalies []DetectedAnomaly

	for i, val := range data {
		if val < lowerBound || val > upperBound {
			score := 0.0
			if iqr > 0 {
				score = (val - median) / iqr // Normalized deviation
			}

			severity := "low"
			diff := 0.0
			if val < lowerBound {
				diff = lowerBound - val
			} else {
				diff = val - upperBound
			}

			// Simple heuristic for severity based on how far past the bound
			if diff > iqr*0.5 {
				severity = "medium"
			}
			if diff > iqr*1.0 {
				severity = "high"
			}

			anomalies = append(anomalies, DetectedAnomaly{
				Index:     i,
				Timestamp: original[i].Timestamp,
				Value:     val,
				Score:     score,
				Severity:  severity,
				Expected:  median,
			})
		}
	}

	return anomalies, nil
}

func calculateStats(data []float64) (mean, stdDev float64) {
	if len(data) == 0 {
		return 0, 0
	}
	sum := 0.0
	for _, v := range data {
		sum += v
	}
	mean = sum / float64(len(data))

	varianceSum := 0.0
	for _, v := range data {
		varianceSum += math.Pow(v-mean, 2)
	}
	variance := varianceSum / float64(len(data))
	return mean, math.Sqrt(variance)
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	index := (p / 100) * float64(len(sorted)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))
	weight := index - float64(lower)

	if lower == upper {
		return sorted[lower]
	}
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

func dataSummary(data []float64) string {
	// Placeholder for simple summary statistics if needed
	return "Analysis complete"
}
