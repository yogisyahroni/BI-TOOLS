package models

import "time"

// ForecastRequest represents the input data for forecasting
type ForecastRequest struct {
	Series    []DataPoint `json:"series" validate:"required,min=2"`
	Horizon   int         `json:"horizon" validate:"required,min=1,max=100"`
	ModelType string      `json:"model_type" validate:"required,oneof=linear exponential moving_average"`
	Interval  string      `json:"interval"` // e.g., "daily", "monthly" (optional, for labeling)
}

// DataPoint represents a single point in the time series
type DataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// ForecastResult represents the output of the forecasting engine
type ForecastResult struct {
	Forecast   []DataPoint `json:"forecast"`
	ModelUsed  string      `json:"model_used"`
	Evaluation Metrics     `json:"evaluation,omitempty"` // Simple error metrics like MSE
}

// Metrics represents accuracy metrics
type Metrics struct {
	MSE float64 `json:"mse"` // Mean Squared Error
	MAE float64 `json:"mae"` // Mean Absolute Error
}
