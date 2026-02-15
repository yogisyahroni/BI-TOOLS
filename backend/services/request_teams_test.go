package services

import (
	"encoding/json"
	"insight-engine-backend/models"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func TestSendTeamsNotification(t *testing.T) {
	// Mock server to capture the request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var payload map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&payload)
		assert.NoError(t, err)

		// Verify key Teams MessageCard fields
		assert.Equal(t, "MessageCard", payload["@type"])
		assert.Equal(t, "http://schema.org/extensions", payload["@context"])

		// The service uses "dc2626" (not #dc2626) for Critical severity
		assert.Equal(t, "dc2626", payload["themeColor"])
		assert.Contains(t, payload["summary"], "Alert Triggered: High CPU")
		assert.Contains(t, payload["title"], "🚨 Alert Triggered: High CPU")

		// Verify sections
		sections, ok := payload["sections"].([]interface{})
		assert.True(t, ok)
		assert.NotEmpty(t, sections)

		section := sections[0].(map[string]interface{})
		assert.Equal(t, "Severity: CRITICAL", section["activityTitle"])

		// Verify facts
		facts, ok := section["facts"].([]interface{})
		assert.True(t, ok)
		assert.Len(t, facts, 2)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Setup mock data
	alertID := uuid.New()
	alert := &models.Alert{
		ID:        alertID.String(),
		Name:      "High CPU",
		Severity:  models.AlertSeverityCritical,
		Column:    "cpu_usage",
		Operator:  ">",
		Threshold: 90.0,
	}

	result := &models.AlertEvaluationResult{
		Triggered: true,
		Value:     95.5,
		Message:   "CPU > 90",
	}

	// Create config JSON
	configMap := map[string]interface{}{
		"url": server.URL,
	}
	configBytes, _ := json.Marshal(configMap)
	configJSON := datatypes.JSON(configBytes)

	channel := &models.AlertNotificationChannelConfig{
		ChannelType: models.AlertChannelTeams,
		IsEnabled:   true,
		Config:      configJSON,
	}

	// Initialize service with nil dependencies since they aren't used for this specific path
	svc := NewAlertNotificationService(&gorm.DB{}, nil, nil, "http://localhost:3000")

	// Call the method
	err := svc.SendAlertNotification(alert, nil, channel, result)
	assert.NoError(t, err)
}
