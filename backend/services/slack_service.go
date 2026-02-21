package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// HTTPClient interface for mocking
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// SlackService handles sending messages to Slack
type SlackService struct {
	WebhookURL string
	BotToken   string
	Client     HTTPClient
}

// SlackAttachment represents a rich message attachment
type SlackAttachment struct {
	Color     string `json:"color,omitempty"`
	Pretext   string `json:"pretext,omitempty"`
	Title     string `json:"title,omitempty"`
	TitleLink string `json:"title_link,omitempty"`
	Text      string `json:"text,omitempty"`
	ImageURL  string `json:"image_url,omitempty"`
	Footer    string `json:"footer,omitempty"`
	Ts        int64  `json:"ts,omitempty"`
}

// SlackPayload represents the message payload sent to Slack
type SlackPayload struct {
	Channel     string            `json:"channel,omitempty"`
	Text        string            `json:"text,omitempty"`
	Attachments []SlackAttachment `json:"attachments,omitempty"`
}

// NewSlackService creates a new instance of SlackService
func NewSlackService() *SlackService {
	return &SlackService{
		WebhookURL: os.Getenv("SLACK_WEBHOOK_URL"),
		BotToken:   os.Getenv("SLACK_BOT_TOKEN"),
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SendNotification sends a simple text message to a Slack channel
// If channel is empty and WebhookURL is set, it uses the default webhook channel
func (s *SlackService) SendNotification(channel string, message string, attachments []SlackAttachment) error {
	if s.WebhookURL == "" && s.BotToken == "" {
		log.Println("⚠️ SlackService: No SLACK_WEBHOOK_URL or SLACK_BOT_TOKEN configured. Skipping notification.")
		return nil
	}

	payload := SlackPayload{
		Text:        message,
		Attachments: attachments,
	}

	if channel != "" {
		payload.Channel = channel
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal slack payload: %w", err)
	}

	var req *http.Request
	var url string

	if s.WebhookURL != "" {
		url = s.WebhookURL
		req, err = http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	} else {
		// Use Chat PostMessage API if using Bot Token
		url = "https://slack.com/api/chat.postMessage"
		req, err = http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
		req.Header.Set("Authorization", "Bearer "+s.BotToken)
	}

	if err != nil {
		return fmt.Errorf("failed to create slack request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send slack request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("slack API error: status code %d", resp.StatusCode)
	}

	return nil
}

// SendAlert sends a formatted alert message to Slack
func (s *SlackService) SendAlert(channel string, title string, message string, severity string) error {
	color := "#36a64f" // Green (Info)
	if severity == "error" || severity == "critical" {
		color = "#ff0000" // Red
	} else if severity == "warning" {
		color = "#ffcc00" // Yellow
	}

	attachment := SlackAttachment{
		Color:  color,
		Title:  title,
		Text:   message,
		Footer: "InsightEngine AI",
		Ts:     time.Now().Unix(),
	}

	return s.SendNotification(channel, "", []SlackAttachment{attachment})
}
