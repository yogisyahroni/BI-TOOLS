package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"time"

	"insight-engine-backend/models"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

type PulseService struct {
	db                *gorm.DB
	screenshotService *ScreenshotService
	slackService      *SlackService
	adminToken        string // Token for screenshot service to access dashboards
}

func NewPulseService(db *gorm.DB, screenshotService *ScreenshotService, slackService *SlackService) *PulseService {
	// TODO: Admin token should be generated or retrieved securely
	// For now, we might rely on the screenshot service using a privileged session or a specific robot user token
	return &PulseService{
		db:                db,
		screenshotService: screenshotService,
		slackService:      slackService,
		adminToken:        "", // Needs to be set
	}
}

// SetAdminToken allows setting the token used for screenshots
func (s *PulseService) SetAdminToken(token string) {
	s.adminToken = token
}

// CreatePulse creates a new pulse
func (s *PulseService) CreatePulse(pulse *models.Pulse) error {
	// Validate cron schedule
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(pulse.Schedule)
	if err != nil {
		return fmt.Errorf("invalid cron schedule: %w", err)
	}

	now := time.Now()
	next := schedule.Next(now)
	pulse.NextRunAt = &next

	return s.db.Create(pulse).Error
}

func (s *PulseService) GetUserPulses(userID uuid.UUID) ([]models.Pulse, error) {
	var pulses []models.Pulse
	err := s.db.Where("user_id = ?", userID).Find(&pulses).Error
	return pulses, err
}

// GetPulse gets a single pulse by ID
func (s *PulseService) GetPulse(id uuid.UUID) (*models.Pulse, error) {
	var pulse models.Pulse
	err := s.db.First(&pulse, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &pulse, nil
}

func (s *PulseService) ProcessDuePulses(ctx context.Context) error {
	var duePulses []models.Pulse
	now := time.Now()

	err := s.db.Where("is_active = ? AND next_run_at <= ?", true, now).Find(&duePulses).Error
	if err != nil {
		return err
	}

	for _, pulse := range duePulses {
		go s.ExecutePulse(ctx, pulse)
	}

	return nil
}

func (s *PulseService) ExecutePulse(ctx context.Context, pulse models.Pulse) {
	logPrefix := fmt.Sprintf("Pulse %s:", pulse.ID)
	LogInfo("pulse_execution", logPrefix+" Starting execution", nil)

	// 1. Capture Screenshot
	// We need a valid token. If s.adminToken is not set, we might fail.
	// Future improvement: Impersonate user or use a service account.
	token := s.adminToken
	if token == "" {
		// Try to find a valid token for the user? Or just warn.
		// For MVP, we assume a system token is available or env var.
		// Or we can use a "public" mode for screenshots if dashboard allows.
	}

	// Default dimensions
	width := int64(1920)
	height := int64(1080)

	// Parse config for dimensions
	// TODO: Parse pulse.Config

	imgData, err := s.screenshotService.CaptureDashboard(pulse.DashboardID.String(), token, width, height)
	if err != nil {
		LogError("pulse_execution", logPrefix+" Screenshot failed", map[string]interface{}{"error": err})
		s.recordFailure(&pulse, err.Error())
		return
	}

	// 2. Send to Channel
	if pulse.ChannelType == models.PulseChannelSlack {
		err = s.sendToSlack(pulse, imgData)
	} else if pulse.ChannelType == models.PulseChannelTeams {
		// TODO: Implement Teams
		err = fmt.Errorf("teams not implemented yet")
	}

	if err != nil {
		LogError("pulse_execution", logPrefix+" Delivery failed", map[string]interface{}{"error": err})
		s.recordFailure(&pulse, err.Error())
		return
	}

	// 3. Update Next Run
	s.recordSuccess(&pulse)
}

func (s *PulseService) sendToSlack(pulse models.Pulse, imgData []byte) error {
	// Slack file upload requires "files.upload" API which needs a Token, not just a Webhook.
	// If WebhookURL is provided, we can't upload files directly to it easily (only JSON).
	// We might need to upload to S3/GCS first and send a link.
	// OR use the Bot Token if configured in SlackService.

	if s.slackService.BotToken == "" {
		return fmt.Errorf("slack bot token required for file uploads")
	}

	// Use Slack API to upload file
	url := "https://slack.com/api/files.upload"

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", "dashboard.png")
	if err != nil {
		return err
	}
	part.Write(imgData)

	writer.WriteField("channels", pulse.WebhookURL) // reusing WebhookURL field for Channel ID if using Bot Token
	writer.WriteField("initial_comment", fmt.Sprintf("Pulse: %s", pulse.Name))

	err = writer.Close()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+s.slackService.BotToken)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("slack api error: %d", resp.StatusCode)
	}

	// Check response body for "ok": true
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if ok, _ := result["ok"].(bool); !ok {
		return fmt.Errorf("slack error: %v", result["error"])
	}

	return nil
}

func (s *PulseService) recordSuccess(pulse *models.Pulse) {
	now := time.Now()
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, _ := parser.Parse(pulse.Schedule)
	next := schedule.Next(now)

	s.db.Model(pulse).Updates(map[string]interface{}{
		"last_run_at":   now,
		"next_run_at":   next,
		"failure_count": 0,
		"last_error":    "",
	})
}

func (s *PulseService) recordFailure(pulse *models.Pulse, errorMsg string) {
	// Still update next run time so we don't retry immediately in a loop?
	// Or should we retry? For now, standard schedule.
	now := time.Now()
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, _ := parser.Parse(pulse.Schedule)
	next := schedule.Next(now)

	s.db.Model(pulse).Updates(map[string]interface{}{
		"last_run_at":   now,
		"next_run_at":   next,
		"failure_count": gorm.Expr("failure_count + 1"),
		"last_error":    errorMsg,
	})
}
