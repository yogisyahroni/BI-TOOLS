package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ChannelNotifier is the interface for sending messages to external collaboration platforms
type ChannelNotifier interface {
	// SendMessage sends a formatted notification to the configured channel
	SendMessage(ctx context.Context, title, body string, fields map[string]string) error
	// Type returns the channel type identifier (e.g., "slack", "teams")
	Type() string
}

// SlackNotifier sends notifications to Slack via incoming webhook
type SlackNotifier struct {
	webhookURL string
	client     *http.Client
}

// NewSlackNotifier creates a new Slack notifier for the given webhook URL
func NewSlackNotifier(webhookURL string) *SlackNotifier {
	return &SlackNotifier{
		webhookURL: webhookURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Type returns the channel type
func (s *SlackNotifier) Type() string { return "slack" }

// SendMessage sends a Block Kit formatted message to Slack
func (s *SlackNotifier) SendMessage(ctx context.Context, title, body string, fields map[string]string) error {
	// Build Slack Block Kit payload
	blocks := []map[string]interface{}{
		{
			"type": "header",
			"text": map[string]interface{}{
				"type": "plain_text",
				"text": title,
			},
		},
		{
			"type": "section",
			"text": map[string]interface{}{
				"type": "mrkdwn",
				"text": body,
			},
		},
	}

	// Add fields as a section block if present
	if len(fields) > 0 {
		fieldBlocks := make([]map[string]interface{}, 0, len(fields))
		for k, v := range fields {
			fieldBlocks = append(fieldBlocks, map[string]interface{}{
				"type": "mrkdwn",
				"text": fmt.Sprintf("*%s:*\n%s", k, v),
			})
		}
		blocks = append(blocks, map[string]interface{}{
			"type":   "section",
			"fields": fieldBlocks,
		})
	}

	// Divider and branding
	blocks = append(blocks, map[string]interface{}{
		"type": "divider",
	})
	blocks = append(blocks, map[string]interface{}{
		"type": "context",
		"elements": []map[string]interface{}{
			{
				"type": "mrkdwn",
				"text": ":bar_chart: Sent from *InsightEngine AI*",
			},
		},
	})

	payload := map[string]interface{}{
		"blocks": blocks,
	}

	return s.postJSON(ctx, payload)
}

// postJSON sends a JSON payload to the webhook URL with retry
func (s *SlackNotifier) postJSON(ctx context.Context, payload interface{}) error {
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack payload: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		if attempt > 0 {
			time.Sleep(500 * time.Millisecond)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.webhookURL, bytes.NewReader(jsonBytes))
		if err != nil {
			return fmt.Errorf("failed to create Slack request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := s.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("Slack webhook request failed: %w", err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}

		respBody, _ := io.ReadAll(resp.Body)
		lastErr = fmt.Errorf("Slack webhook returned HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return lastErr
}

// TeamsNotifier sends notifications to Microsoft Teams via incoming webhook
type TeamsNotifier struct {
	webhookURL string
	client     *http.Client
}

// NewTeamsNotifier creates a new Teams notifier for the given webhook URL
func NewTeamsNotifier(webhookURL string) *TeamsNotifier {
	return &TeamsNotifier{
		webhookURL: webhookURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Type returns the channel type
func (t *TeamsNotifier) Type() string { return "teams" }

// SendMessage sends an Adaptive Card formatted message to Microsoft Teams
func (t *TeamsNotifier) SendMessage(ctx context.Context, title, body string, fields map[string]string) error {
	// Build Adaptive Card body elements
	cardBody := []map[string]interface{}{
		{
			"type":   "TextBlock",
			"text":   title,
			"size":   "Large",
			"weight": "Bolder",
			"color":  "Accent",
		},
		{
			"type":      "TextBlock",
			"text":      body,
			"wrap":      true,
			"separator": true,
		},
	}

	// Add fields as a FactSet
	if len(fields) > 0 {
		facts := make([]map[string]string, 0, len(fields))
		for k, v := range fields {
			facts = append(facts, map[string]string{
				"title": k,
				"value": v,
			})
		}
		cardBody = append(cardBody, map[string]interface{}{
			"type":  "FactSet",
			"facts": facts,
		})
	}

	// Branding
	cardBody = append(cardBody, map[string]interface{}{
		"type":      "TextBlock",
		"text":      "📊 Sent from InsightEngine AI",
		"size":      "Small",
		"color":     "Light",
		"separator": true,
	})

	// Wrap in Adaptive Card + Teams message envelope
	payload := map[string]interface{}{
		"type": "message",
		"attachments": []map[string]interface{}{
			{
				"contentType": "application/vnd.microsoft.card.adaptive",
				"content": map[string]interface{}{
					"$schema": "http://adaptivecards.io/schemas/adaptive-card.json",
					"type":    "AdaptiveCard",
					"version": "1.4",
					"body":    cardBody,
				},
			},
		},
	}

	return t.postJSON(ctx, payload)
}

// postJSON sends a JSON payload to the Teams webhook URL with retry
func (t *TeamsNotifier) postJSON(ctx context.Context, payload interface{}) error {
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Teams payload: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		if attempt > 0 {
			time.Sleep(500 * time.Millisecond)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.webhookURL, bytes.NewReader(jsonBytes))
		if err != nil {
			return fmt.Errorf("failed to create Teams request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := t.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("Teams webhook request failed: %w", err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}

		respBody, _ := io.ReadAll(resp.Body)
		lastErr = fmt.Errorf("Teams webhook returned HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return lastErr
}

// CreateNotifierForType creates the appropriate ChannelNotifier based on channel type
func CreateNotifierForType(channelType, webhookURL string) (ChannelNotifier, error) {
	switch channelType {
	case "slack":
		return NewSlackNotifier(webhookURL), nil
	case "teams":
		return NewTeamsNotifier(webhookURL), nil
	default:
		return nil, fmt.Errorf("unsupported channel type: %s", channelType)
	}
}
