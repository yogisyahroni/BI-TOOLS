package integration

import (
	"bytes"
	"encoding/json"
	"insight-engine-backend/handlers"
	"insight-engine-backend/services"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func setupAnomalyApp() *fiber.App {
	app := fiber.New()
	service := services.NewAnomalyDetectionService()
	handler := handlers.NewAnomalyHandler(service)
	app.Post("/api/analytics/anomalies", handler.DetectAnomalies)
	return app
}

func TestDetectAnomalies_ZScore(t *testing.T) {
	app := setupAnomalyApp()

	// Data with a clear spike - need enough points for Z-Score to work
	data := []services.AnomalyPoint{}
	for i := 0; i < 20; i++ {
		data = append(data, services.AnomalyPoint{Timestamp: "2023-01-XX", Value: 10})
	}
	data = append(data, services.AnomalyPoint{Timestamp: "2023-01-SPIKE", Value: 100}) // Anomaly

	reqBody := services.AnomalyRequest{
		Data:        data,
		Method:      services.MethodZScore,
		Sensitivity: 2.0, // Set explicit sensitivity lower than max possible for N=21
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/analytics/anomalies", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result services.AnomalyResult
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)

	assert.NotEmpty(t, result.Anomalies)
	found := false
	for _, a := range result.Anomalies {
		if a.Value == 100 {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected anomaly at value 100")
}

func TestDetectAnomalies_IQR(t *testing.T) {
	app := setupAnomalyApp()

	data := []services.AnomalyPoint{
		{Timestamp: "1", Value: 1},
		{Timestamp: "2", Value: 2},
		{Timestamp: "3", Value: 3},
		{Timestamp: "4", Value: 4},
		{Timestamp: "5", Value: 5},
		{Timestamp: "6", Value: 100}, // Outlier
	}

	reqBody := services.AnomalyRequest{
		Data:   data,
		Method: services.MethodIQR,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/analytics/anomalies", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result services.AnomalyResult
	json.NewDecoder(resp.Body).Decode(&result)

	assert.NotEmpty(t, result.Anomalies)
	assert.Equal(t, 100.0, result.Anomalies[0].Value)
}
