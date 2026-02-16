package services

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"insight-engine-backend/models"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type WebhookService struct {
	db *gorm.DB
}

func NewWebhookService(db *gorm.DB) *WebhookService {
	return &WebhookService{db: db}
}

// CreateWebhook creates a new webhook
func (s *WebhookService) CreateWebhook(req models.WebhookRequest, userID uuid.UUID) (*models.Webhook, error) {
	// Generate secret
	secret := uuid.New().String()

	eventsJSON, err := json.Marshal(req.Events)
	if err != nil {
		return nil, err
	}

	headersJSON, err := json.Marshal(req.Headers)
	if err != nil {
		return nil, err
	}

	webhook := &models.Webhook{
		Name:        req.Name,
		URL:         req.URL,
		Events:      datatypes.JSON(eventsJSON),
		Secret:      secret,
		IsActive:    req.IsActive,
		Headers:     datatypes.JSON(headersJSON),
		Description: req.Description,
		UserID:      userID,
	}

	if err := s.db.Create(webhook).Error; err != nil {
		return nil, err
	}

	return webhook, nil
}

// GetWebhooks returns all webhooks for a user
func (s *WebhookService) GetWebhooks(userID uuid.UUID) ([]models.Webhook, error) {
	var webhooks []models.Webhook
	if err := s.db.Where("user_id = ?", userID).Find(&webhooks).Error; err != nil {
		return nil, err
	}
	return webhooks, nil
}

// GetWebhook returns a single webhook
func (s *WebhookService) GetWebhook(id uuid.UUID, userID uuid.UUID) (*models.Webhook, error) {
	var webhook models.Webhook
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&webhook).Error; err != nil {
		return nil, err
	}
	return &webhook, nil
}

// UpdateWebhook updates an existing webhook
func (s *WebhookService) UpdateWebhook(id uuid.UUID, req models.WebhookRequest, userID uuid.UUID) (*models.Webhook, error) {
	var webhook models.Webhook
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&webhook).Error; err != nil {
		return nil, err
	}

	eventsJSON, err := json.Marshal(req.Events)
	if err != nil {
		return nil, err
	}

	headersJSON, err := json.Marshal(req.Headers)
	if err != nil {
		return nil, err
	}

	webhook.Name = req.Name
	webhook.URL = req.URL
	webhook.Events = datatypes.JSON(eventsJSON)
	webhook.IsActive = req.IsActive
	webhook.Headers = datatypes.JSON(headersJSON)
	webhook.Description = req.Description

	if err := s.db.Save(&webhook).Error; err != nil {
		return nil, err
	}

	return &webhook, nil
}

// DeleteWebhook deletes a webhook
func (s *WebhookService) DeleteWebhook(id uuid.UUID, userID uuid.UUID) error {
	result := s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Webhook{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("webhook not found")
	}
	return nil
}

// GetWebhookLogs returns logs for a webhook
func (s *WebhookService) GetWebhookLogs(webhookID uuid.UUID, userID uuid.UUID, limit int) ([]models.WebhookLog, error) {
	// Verify ownership
	var webhook models.Webhook
	if err := s.db.Where("id = ? AND user_id = ?", webhookID, userID).First(&webhook).Error; err != nil {
		return nil, err
	}

	var logs []models.WebhookLog
	if err := s.db.Where("webhook_id = ?", webhookID).Order("created_at desc").Limit(limit).Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

// DispatchEvent dispatches an event to all matching webhooks
// This should be called asynchronously (go s.DispatchEvent(...))
func (s *WebhookService) DispatchEvent(eventType string, payload interface{}, userID *uuid.UUID) {
	var webhooks []models.Webhook

	query := s.db.Where("is_active = ?", true)
	if userID != nil {
		query = query.Where("user_id = ?", userID)
	}

	// Ideally we filter by event type in DB using JSONB queries, but for simplicity fetch all active and filter in code
	// Postgres JSONB query: query.Where("events @> ?", fmt.Sprintf(`["%s"]`, eventType))
	if err := query.Find(&webhooks).Error; err != nil {
		fmt.Printf("Error fetching webhooks: %v\n", err)
		return
	}

	for _, webhook := range webhooks {
		// Check if webhook subscribes to this event
		var events []string
		if err := json.Unmarshal(webhook.Events, &events); err != nil {
			continue
		}

		subscribed := false
		for _, e := range events {
			if e == "*" || e == eventType {
				subscribed = true
				break
			}
		}

		if subscribed {
			go s.sendWebhook(webhook, eventType, payload)
		}
	}
}

func (s *WebhookService) sendWebhook(webhook models.Webhook, eventType string, payload interface{}) {
	startTime := time.Now()

	requestBody := map[string]interface{}{
		"event":     eventType,
		"timestamp": startTime.UTC().Format(time.RFC3339),
		"payload":   payload,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		s.logAttempt(webhook.ID, eventType, nil, 0, "", 0, "failure", fmt.Sprintf("Marshal error: %v", err))
		return
	}

	// Calculate HMAC signature
	signature := computeHMAC(bodyBytes, webhook.Secret)

	req, err := http.NewRequest("POST", webhook.URL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		s.logAttempt(webhook.ID, eventType, bodyBytes, 0, "", 0, "failure", fmt.Sprintf("Request creation error: %v", err))
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Insight-Signature", signature)
	req.Header.Set("X-Insight-Event", eventType)
	req.Header.Set("User-Agent", "InsightEngine-Webhook/1.0")

	// Add custom headers
	var customHeaders map[string]string
	if len(webhook.Headers) > 0 {
		_ = json.Unmarshal(webhook.Headers, &customHeaders)
		for k, v := range customHeaders {
			req.Header.Set(k, v)
		}
	}

	// Retry logic (Simplified exponential backoff)
	client := &http.Client{Timeout: 10 * time.Second}
	if os.Getenv("SKIP_TLS_VERIFY") == "true" {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client.Transport = tr
	}
	var resp *http.Response
	var reqErr error

	for i := 0; i < 3; i++ {
		resp, reqErr = client.Do(req)
		if reqErr == nil && resp.StatusCode < 500 {
			break
		}
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	duration := time.Since(startTime).Milliseconds()

	if reqErr != nil {
		s.incrementFailureCount(webhook)
		s.logAttempt(webhook.ID, eventType, bodyBytes, 0, "", duration, "failure", reqErr.Error())
		return
	}
	defer resp.Body.Close()

	// Read response body (limited)
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	respBody := buf.String()
	if len(respBody) > 1000 {
		respBody = respBody[:1000] + "..."
	}

	status := "success"
	if resp.StatusCode >= 400 {
		status = "failure"
		s.incrementFailureCount(webhook)
	} else {
		// Reset failure count on success
		s.resetFailureCount(webhook)
	}

	s.logAttempt(webhook.ID, eventType, bodyBytes, resp.StatusCode, respBody, duration, status, "")

	// Update LastTriggeredAt
	now := time.Now()
	s.db.Model(&webhook).Update("last_triggered_at", now)
}

func (s *WebhookService) logAttempt(webhookID uuid.UUID, eventType string, reqBody []byte, status int, respBody string, duration int64, logStatus string, errMsg string) {
	log := models.WebhookLog{
		WebhookID:      webhookID,
		EventType:      eventType,
		RequestPayload: datatypes.JSON(reqBody),
		ResponseStatus: status,
		ResponseBody:   respBody,
		DurationMs:     duration,
		Status:         logStatus,
		ErrorMessage:   errMsg,
	}
	s.db.Create(&log)
}

func (s *WebhookService) incrementFailureCount(webhook models.Webhook) {
	s.db.Model(&webhook).Update("failure_count", gorm.Expr("failure_count + ?", 1))
}

func (s *WebhookService) resetFailureCount(webhook models.Webhook) {
	if webhook.FailureCount > 0 {
		s.db.Model(&webhook).Update("failure_count", 0)
	}
}

// SendTestWebhook sends a test event to a specific webhook
func (s *WebhookService) SendTestWebhook(webhookID uuid.UUID, userID uuid.UUID, eventType string, payload interface{}) error {
	webhook, err := s.GetWebhook(webhookID, userID)
	if err != nil {
		return err
	}

	go s.sendWebhook(*webhook, eventType, payload)
	return nil
}

func computeHMAC(message []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(message)
	return hex.EncodeToString(mac.Sum(nil))
}
