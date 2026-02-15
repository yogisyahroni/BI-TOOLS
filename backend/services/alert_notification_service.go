package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"insight-engine-backend/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AlertNotificationService handles alert notification operations
type AlertNotificationService struct {
	db                  *gorm.DB
	emailService        *EmailService
	notificationService *NotificationService
	baseURL             string
}

// NewAlertNotificationService creates a new alert notification service
func NewAlertNotificationService(db *gorm.DB, emailService *EmailService, notificationService *NotificationService, baseURL string) *AlertNotificationService {
	return &AlertNotificationService{
		db:                  db,
		emailService:        emailService,
		notificationService: notificationService,
		baseURL:             baseURL,
	}
}

// SendAlertNotification sends a notification through the specified channel
func (s *AlertNotificationService) SendAlertNotification(alert *models.Alert, history *models.AlertHistory, channel *models.AlertNotificationChannelConfig, result *models.AlertEvaluationResult) error {
	if !channel.IsEnabled {
		return nil
	}

	switch channel.ChannelType {
	case models.AlertChannelEmail:
		return s.sendEmailNotification(alert, history, channel, result)
	case models.AlertChannelWebhook:
		return s.sendWebhookNotification(alert, history, channel, result)
	case models.AlertChannelInApp:
		return s.sendInAppNotification(alert, history, channel, result)
	case models.AlertChannelSlack:
		return s.sendSlackNotification(alert, history, channel, result)
	case models.AlertChannelTeams:
		return s.sendTeamsNotification(alert, history, channel, result)
	default:
		return fmt.Errorf("unsupported notification channel: %s", channel.ChannelType)
	}
}

// sendEmailNotification sends an email notification for an alert
func (s *AlertNotificationService) sendEmailNotification(alert *models.Alert, history *models.AlertHistory, channel *models.AlertNotificationChannelConfig, result *models.AlertEvaluationResult) error {
	// Get channel config
	config, err := channel.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get channel config: %w", err)
	}

	// Get recipients
	recipients := []string{alert.User.Email}
	if customRecipients, ok := config["recipients"].([]interface{}); ok {
		recipients = []string{}
		for _, r := range customRecipients {
			if email, ok := r.(string); ok {
				recipients = append(recipients, email)
			}
		}
	}

	if len(recipients) == 0 {
		return fmt.Errorf("no recipients configured")
	}

	// Build email subject
	severityEmoji := map[models.AlertSeverity]string{
		models.AlertSeverityCritical: "🚨",
		models.AlertSeverityWarning:  "⚠️",
		models.AlertSeverityInfo:     "ℹ️",
	}

	emoji := severityEmoji[alert.Severity]
	if emoji == "" {
		emoji = "⚠️"
	}

	subject := fmt.Sprintf("%s [%s] Alert Triggered: %s", emoji, string(alert.Severity), alert.Name)

	// Build email body
	queryName := "Unknown Query"
	if alert.Query != nil {
		queryName = alert.Query.Name
	}

	bodyHTML := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Alert Notification</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f4f4f4;
        }
        .container {
            background-color: #ffffff;
            border-radius: 8px;
            padding: 40px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .header {
            text-align: center;
            margin-bottom: 30px;
        }
        .logo {
            font-size: 32px;
            font-weight: bold;
            color: #4F46E5;
            margin-bottom: 10px;
        }
        .alert-box {
            background-color: %s;
            border-left: 4px solid %s;
            padding: 20px;
            margin: 20px 0;
            border-radius: 4px;
        }
        .alert-box h2 {
            margin: 0 0 10px 0;
            color: %s;
        }
        .metric {
            background-color: #f9fafb;
            padding: 15px;
            border-radius: 4px;
            margin: 15px 0;
        }
        .metric-row {
            display: flex;
            justify-content: space-between;
            margin: 5px 0;
        }
        .button {
            display: inline-block;
            background-color: #4F46E5;
            color: #ffffff;
            text-decoration: none;
            padding: 12px 30px;
            border-radius: 6px;
            font-weight: 600;
            margin: 20px 0;
        }
        .button:hover {
            background-color: #4338ca;
        }
        .footer {
            margin-top: 30px;
            padding-top: 20px;
            border-top: 1px solid #e5e5e5;
            text-align: center;
            color: #666;
            font-size: 14px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="logo">InsightEngine</div>
        </div>
        
        <div class="alert-box">
            <h2>🚨 Alert Triggered: %s</h2>
            <p><strong>Severity:</strong> %s</p>
            <p><strong>Time:</strong> %s</p>
        </div>
        
        <div class="metric">
            <div class="metric-row">
                <strong>Query:</strong>
                <span>%s</span>
            </div>
            <div class="metric-row">
                <strong>Condition:</strong>
                <span>%s %s %s</span>
            </div>
            <div class="metric-row">
                <strong>Current Value:</strong>
                <span style="color: #dc2626; font-weight: bold;">%.2f</span>
            </div>
            <div class="metric-row">
                <strong>Threshold:</strong>
                <span>%.2f</span>
            </div>
        </div>
        
        <p>%s</p>
        
        <div style="text-align: center;">
            <a href="%s/alerts" class="button">View Alert</a>
        </div>
        
        <div class="footer">
            <p>This is an automated alert from InsightEngine.</p>
            <p style="margin-top: 10px;">&copy; 2026 InsightEngine. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`,
		getAlertBackgroundColor(alert.Severity),
		getAlertBorderColor(alert.Severity),
		getAlertTextColor(alert.Severity),
		alert.Name,
		string(alert.Severity),
		time.Now().Format("2006-01-02 15:04:05 MST"),
		queryName,
		alert.Column, alert.Operator, formatThreshold(alert.Threshold),
		result.Value,
		alert.Threshold,
		getAlertDescription(alert),
		s.baseURL,
	)

	// Send email using email service
	req := &SendReportEmailRequest{
		To:       recipients,
		Subject:  subject,
		BodyHTML: bodyHTML,
		BodyText: fmt.Sprintf("Alert Triggered: %s\n\nQuery: %s\nCondition: %s %s %s\nCurrent Value: %.2f\nThreshold: %.2f\n\nView at: %s/alerts",
			alert.Name, queryName, alert.Column, alert.Operator, formatThreshold(alert.Threshold),
			result.Value, alert.Threshold, s.baseURL),
	}

	return s.emailService.SendReportEmail(req)
}

// sendWebhookNotification sends a webhook notification
func (s *AlertNotificationService) sendWebhookNotification(alert *models.Alert, history *models.AlertHistory, channel *models.AlertNotificationChannelConfig, result *models.AlertEvaluationResult) error {
	// Get channel config
	config, err := channel.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get channel config: %w", err)
	}

	// Get webhook URL
	webhookURL, ok := config["url"].(string)
	if !ok || webhookURL == "" {
		// Fallback to alert's webhook URL
		if alert.WebhookURL != nil && *alert.WebhookURL != "" {
			webhookURL = *alert.WebhookURL
		} else {
			return fmt.Errorf("webhook URL not configured")
		}
	}

	// Build payload
	payload := map[string]interface{}{
		"alert": map[string]interface{}{
			"id":          alert.ID,
			"name":        alert.Name,
			"description": alert.Description,
			"severity":    alert.Severity,
			"state":       alert.State,
			"column":      alert.Column,
			"operator":    alert.Operator,
			"threshold":   alert.Threshold,
		},
		"result": map[string]interface{}{
			"triggered": result.Triggered,
			"value":     result.Value,
			"message":   result.Message,
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"link":      fmt.Sprintf("%s/alerts/%s", s.baseURL, alert.ID),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Build request
	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "InsightEngine-Alert/1.0")

	// Add custom headers if configured
	if headers, ok := config["headers"].(map[string]interface{}); ok {
		for key, value := range headers {
			if strValue, ok := value.(string); ok {
				req.Header.Set(key, strValue)
			}
		}
	}

	// Also use alert's webhook headers if available
	if alert.WebhookHeaders != nil {
		alertHeaders, err := alert.GetWebhookHeaders()
		if err == nil {
			for key, value := range alertHeaders {
				req.Header.Set(key, value)
			}
		}
	}

	// Execute request with retry
	client := &http.Client{Timeout: 30 * time.Second}

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}

		lastErr = fmt.Errorf("webhook returned status %d", resp.StatusCode)
		if resp.StatusCode >= 500 {
			// Retry on server errors
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}
		// Don't retry on client errors
		return lastErr
	}

	return fmt.Errorf("webhook failed after 3 attempts: %w", lastErr)
}

// sendInAppNotification sends an in-app notification
func (s *AlertNotificationService) sendInAppNotification(alert *models.Alert, history *models.AlertHistory, channel *models.AlertNotificationChannelConfig, result *models.AlertEvaluationResult) error {
	if s.notificationService == nil {
		return fmt.Errorf("notification service not configured")
	}

	userID, err := uuid.Parse(alert.UserID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	title := fmt.Sprintf("Alert: %s", alert.Name)
	message := fmt.Sprintf("%s %s %s (value: %.2f)", alert.Column, alert.Operator, formatThreshold(alert.Threshold), result.Value)

	link := fmt.Sprintf("/alerts/%s", alert.ID)

	metadata := map[string]interface{}{
		"alertId":   alert.ID,
		"severity":  alert.Severity,
		"value":     result.Value,
		"threshold": alert.Threshold,
	}

	return s.notificationService.SendNotification(userID, title, message, string(alert.Severity), link, metadata)
}

// sendSlackNotification sends a Slack notification via webhook
func (s *AlertNotificationService) sendSlackNotification(alert *models.Alert, history *models.AlertHistory, channel *models.AlertNotificationChannelConfig, result *models.AlertEvaluationResult) error {
	// Get channel config
	config, err := channel.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get channel config: %w", err)
	}

	// Get webhook URL
	webhookURL, ok := config["url"].(string)
	if !ok || webhookURL == "" {
		return fmt.Errorf("Slack webhook URL not configured")
	}

	// Build color based on severity
	color := map[models.AlertSeverity]string{
		models.AlertSeverityCritical: "#dc2626",
		models.AlertSeverityWarning:  "#f59e0b",
		models.AlertSeverityInfo:     "#3b82f6",
	}[alert.Severity]
	if color == "" {
		color = "#6b7280"
	}

	// Build Slack message blocks
	payload := map[string]interface{}{
		"attachments": []map[string]interface{}{
			{
				"color": color,
				"blocks": []map[string]interface{}{
					{
						"type": "header",
						"text": map[string]string{
							"type":  "plain_text",
							"text":  fmt.Sprintf("🚨 Alert Triggered: %s", alert.Name),
							"emoji": "true",
						},
					},
					{
						"type": "section",
						"fields": []map[string]string{
							{
								"type": "mrkdwn",
								"text": fmt.Sprintf("*Severity:*\n%s", alert.Severity),
							},
							{
								"type": "mrkdwn",
								"text": fmt.Sprintf("*Time:*\n%s", time.Now().Format("2006-01-02 15:04:05")),
							},
							{
								"type": "mrkdwn",
								"text": fmt.Sprintf("*Condition:*\n%s %s %s", alert.Column, alert.Operator, formatThreshold(alert.Threshold)),
							},
							{
								"type": "mrkdwn",
								"text": fmt.Sprintf("*Current Value:*\n`%.2f`", result.Value),
							},
						},
					},
					{
						"type": "actions",
						"elements": []map[string]interface{}{
							{
								"type": "button",
								"text": map[string]string{
									"type": "plain_text",
									"text": "View Alert",
								},
								"url":   fmt.Sprintf("%s/alerts/%s", s.baseURL, alert.ID),
								"style": "primary",
							},
						},
					},
				},
			},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Send request
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to send Slack notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("Slack webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// sendTeamsNotification sends a Microsoft Teams notification via webhook
func (s *AlertNotificationService) sendTeamsNotification(alert *models.Alert, history *models.AlertHistory, channel *models.AlertNotificationChannelConfig, result *models.AlertEvaluationResult) error {
	// Get channel config
	config, err := channel.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get channel config: %w", err)
	}

	// Get webhook URL
	webhookURL, ok := config["url"].(string)
	if !ok || webhookURL == "" {
		return fmt.Errorf("Teams webhook URL not configured")
	}

	// Build color based on severity (Teams uses hex without # sometimes, but # is safe for modern cards)
	themeColor := map[models.AlertSeverity]string{
		models.AlertSeverityCritical: "dc2626", // Red
		models.AlertSeverityWarning:  "f59e0b", // Amber
		models.AlertSeverityInfo:     "3b82f6", // Blue
	}[alert.Severity]
	if themeColor == "" {
		themeColor = "6b7280" // Gray
	}

	// Build Teams MessageCard payload (Legacy but widely supported)
	// Docs: https://learn.microsoft.com/en-us/outlook/actionable-messages/message-card-reference
	payload := map[string]interface{}{
		"@type":      "MessageCard",
		"@context":   "http://schema.org/extensions",
		"themeColor": themeColor,
		"summary":    fmt.Sprintf("Alert Triggered: %s", alert.Name),
		"title":      fmt.Sprintf("🚨 Alert Triggered: %s", alert.Name),
		"sections": []map[string]interface{}{
			{
				"activityTitle":    fmt.Sprintf("Severity: %s", alert.Severity),
				"activitySubtitle": fmt.Sprintf("Time: %s", time.Now().Format("2006-01-02 15:04:05")),
				"facts": []map[string]string{
					{
						"name":  "Condition",
						"value": fmt.Sprintf("%s %s %s", alert.Column, alert.Operator, formatThreshold(alert.Threshold)),
					},
					{
						"name":  "Current Value",
						"value": fmt.Sprintf("%.2f", result.Value),
					},
				},
				"text": getAlertDescription(alert),
			},
		},
		"potentialAction": []map[string]interface{}{
			{
				"@type": "OpenUri",
				"name":  "View Alert",
				"targets": []map[string]string{
					{
						"os":  "default",
						"uri": fmt.Sprintf("%s/alerts/%s", s.baseURL, alert.ID),
					},
				},
			},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Send request
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to send Teams notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("Teams webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// GetNotificationTemplates returns available notification templates
func (s *AlertNotificationService) GetNotificationTemplates() ([]models.AlertNotificationTemplate, error) {
	var templates []models.AlertNotificationTemplate
	if err := s.db.Where("is_active = ?", true).Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

// RenderNotificationTemplate renders a template with alert data
func (s *AlertNotificationService) RenderNotificationTemplate(template *models.AlertNotificationTemplate, alert *models.Alert, result *models.AlertEvaluationResult) (string, error) {
	data := map[string]interface{}{
		"AlertName":     alert.Name,
		"AlertSeverity": alert.Severity,
		"AlertState":    alert.State,
		"Column":        alert.Column,
		"Operator":      alert.Operator,
		"Threshold":     alert.Threshold,
		"Value":         result.Value,
		"Triggered":     result.Triggered,
		"Message":       result.Message,
		"Timestamp":     time.Now().Format(time.RFC3339),
	}

	// Simple template replacement
	content := []byte(template.Content)
	for key, value := range data {
		placeholder := fmt.Sprintf("{{%s}}", key)
		content = bytes.ReplaceAll(content, []byte(placeholder), []byte(fmt.Sprintf("%v", value)))
	}

	return string(content), nil
}

// TrackNotificationDelivery tracks the delivery status of a notification
func (s *AlertNotificationService) TrackNotificationDelivery(historyID uuid.UUID, channelType models.AlertNotificationChannel, status string, error string) error {
	log := &models.AlertNotificationLog{
		ID:          uuid.New(),
		HistoryID:   historyID,
		ChannelType: channelType,
		Status:      status,
	}

	if error != "" {
		log.Error = &error
	}

	if status == "sent" {
		now := time.Now()
		log.SentAt = &now
	}

	return s.db.Create(log).Error
}

// Helper functions

func getAlertBackgroundColor(severity models.AlertSeverity) string {
	switch severity {
	case models.AlertSeverityCritical:
		return "#fef2f2"
	case models.AlertSeverityWarning:
		return "#fffbeb"
	case models.AlertSeverityInfo:
		return "#eff6ff"
	default:
		return "#f9fafb"
	}
}

func getAlertBorderColor(severity models.AlertSeverity) string {
	switch severity {
	case models.AlertSeverityCritical:
		return "#dc2626"
	case models.AlertSeverityWarning:
		return "#f59e0b"
	case models.AlertSeverityInfo:
		return "#3b82f6"
	default:
		return "#6b7280"
	}
}

func getAlertTextColor(severity models.AlertSeverity) string {
	switch severity {
	case models.AlertSeverityCritical:
		return "#991b1b"
	case models.AlertSeverityWarning:
		return "#92400e"
	case models.AlertSeverityInfo:
		return "#1e40af"
	default:
		return "#374151"
	}
}

func getAlertDescription(alert *models.Alert) string {
	// if alert.Description != nil && *alert.Description != "" {
	// 	return *alert.Description
	// }
	return fmt.Sprintf("This alert monitors when %s %s the threshold value.", alert.Column, getOperatorDescription(alert.Operator))
}

func getOperatorDescription(operator string) string {
	switch operator {
	case ">":
		return "exceeds"
	case "<":
		return "falls below"
	case "=", "==":
		return "equals"
	case ">=":
		return "reaches or exceeds"
	case "<=":
		return "reaches or falls below"
	case "!=":
		return "differs from"
	default:
		return "compares to"
	}
}

func formatThreshold(threshold float64) string {
	if threshold == float64(int64(threshold)) {
		return fmt.Sprintf("%.0f", threshold)
	}
	return fmt.Sprintf("%.2f", threshold)
}
