package services

import (
	"fmt"
	"insight-engine-backend/models"
	"math"
	"time"
)

type ForecastingService struct{}

func NewForecastingService() *ForecastingService {
	return &ForecastingService{}
}

// Forecast generates predictions based on the request
func (s *ForecastingService) Forecast(req models.ForecastRequest) (*models.ForecastResult, error) {
	if len(req.Series) < 2 {
		return nil, fmt.Errorf("insufficient data points")
	}

	switch req.ModelType {
	case "linear":
		return s.forecastLinear(req)
	case "moving_average":
		return s.forecastMovingAverage(req)
	default:
		return nil, fmt.Errorf("unsupported model type: %s", req.ModelType)
	}
}

// Linear Regression using OLS
func (s *ForecastingService) forecastLinear(req models.ForecastRequest) (*models.ForecastResult, error) {
	// Determine time interval (avg difference between last few points)
	lastIdx := len(req.Series) - 1
	lastPoint := req.Series[lastIdx]
	prevPoint := req.Series[lastIdx-1]
	interval := lastPoint.Timestamp.Sub(prevPoint.Timestamp)

	// Manual OLS
	var sumX, sumY, sumXY, sumXX float64
	n := float64(len(req.Series))

	for i, p := range req.Series {
		x := float64(i)
		y := p.Value
		sumX += x
		sumY += y
		sumXY += x * y
		sumXX += x * x
	}

	m := (n*sumXY - sumX*sumY) / (n*sumXX - sumX*sumX)
	c := (sumY - m*sumX) / n

	forecast := make([]models.DataPoint, req.Horizon)
	for i := 0; i < req.Horizon; i++ {
		nextX := float64(len(req.Series) + i)
		nextY := m*nextX + c

		forecast[i] = models.DataPoint{
			Timestamp: lastPoint.Timestamp.Add(interval * time.Duration(i+1)),
			Value:     nextY,
		}
	}

	// Calculate simple error metric (MSE) on training data
	var mse float64
	for i, p := range req.Series {
		x := float64(i)
		predicted := m*x + c
		mse += math.Pow(p.Value-predicted, 2)
	}
	mse /= n

	return &models.ForecastResult{
		Forecast:  forecast,
		ModelUsed: "Linear Regression (OLS)",
		Evaluation: models.Metrics{
			MSE: mse,
		},
	}, nil
}

// Simple Moving Average
func (s *ForecastingService) forecastMovingAverage(req models.ForecastRequest) (*models.ForecastResult, error) {
	window := 3
	if len(req.Series) < window {
		window = len(req.Series)
	}

	// Determine interval
	lastIdx := len(req.Series) - 1
	lastPoint := req.Series[lastIdx]
	interval := lastPoint.Timestamp.Sub(req.Series[lastIdx-1].Timestamp)

	forecast := make([]models.DataPoint, req.Horizon)

	// For projection, we can use the last N points average as the next point,
	// then append it to history to predict the one after (simulating trend dampening).
	// Or just project the last calculated average flatly (Naive).
	// Iterative approach allows for some dampening/trend following if combined with weights,
	// but for simple SMA, usually the next point IS the average.

	// Let's use Iterative SMA:
	// P_{t+1} = Avg(last 3 actuals).
	// P_{t+2} = Avg(last 2 actuals + P_{t+1}).

	currentSeries := make([]float64, len(req.Series))
	for i, p := range req.Series {
		currentSeries[i] = p.Value
	}

	for i := 0; i < req.Horizon; i++ {
		sum := 0.0
		start := len(currentSeries) - window
		for j := start; j < len(currentSeries); j++ {
			sum += currentSeries[j]
		}
		avg := sum / float64(window)

		currentSeries = append(currentSeries, avg)

		forecast[i] = models.DataPoint{
			Timestamp: lastPoint.Timestamp.Add(interval * time.Duration(i+1)),
			Value:     avg,
		}
	}

	return &models.ForecastResult{
		Forecast:  forecast,
		ModelUsed: "Simple Moving Average (Window=3)",
		Evaluation: models.Metrics{
			MSE: 0, // Not calculating for SMA simplicity
		},
	}, nil
}
