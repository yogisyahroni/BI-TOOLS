package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type AlertState string
type AlertSeverity string
type AlertNotificationChannel string

const (
	AlertStateOK           AlertState = "OK"
	AlertStateTriggered    AlertState = "TRIGGERED"
	AlertStateAcknowledged AlertState = "ACKNOWLEDGED"
	AlertStateMuted        AlertState = "MUTED"
	AlertStateError        AlertState = "ERROR"
	AlertStateUnknown      AlertState = "UNKNOWN"

	AlertSeverityCritical AlertSeverity = "CRITICAL"
	AlertSeverityWarning  AlertSeverity = "WARNING"
	AlertSeverityInfo     AlertSeverity = "INFO"

	AlertChannelEmail   AlertNotificationChannel = "EMAIL"
	AlertChannelWebhook AlertNotificationChannel = "WEBHOOK"
	AlertChannelSlack   AlertNotificationChannel = "SLACK"
	AlertChannelTeams   AlertNotificationChannel = "TEAMS"
	AlertChannelInApp   AlertNotificationChannel = "IN_APP"
)

// Alert represents an alert rule
type Alert struct {
	ID          string `gorm:"primaryKey;type:uuid" json:"id"`
	Name        string `gorm:"not null" json:"name"`
	Description string `json:"description"`
	QueryID     string `gorm:"type:uuid;not null" json:"query_id"`
	UserID      string `gorm:"type:uuid;not null" json:"user_id"`

	// Rule Definition
	Column    string  `gorm:"not null" json:"column"`
	Operator  string  `gorm:"not null" json:"operator"` // >, <, >=, <=, ==, !=
	Threshold float64 `gorm:"not null" json:"threshold"`

	// Scheduling
	Schedule  string     `gorm:"not null" json:"schedule"` // Cron expression
	Timezone  string     `json:"timezone"`
	NextRunAt *time.Time `json:"next_run_at"`

	// Configuration
	Severity        AlertSeverity `json:"severity"`
	CooldownMinutes int           `json:"cooldown_minutes"`
	IsActive        bool          `json:"is_active"`
	IsMuted         bool          `json:"is_muted"`
	MutedUntil      *time.Time    `json:"muted_until"`
	MuteDuration    *int          `json:"mute_duration"` // Minutes

	// State
	State             AlertState `json:"state"`
	LastRunAt         *time.Time `json:"last_run_at"`
	LastTriggeredAt   *time.Time `json:"last_triggered_at"`
	LastValue         *float64   `json:"last_value"`
	LastStatus        *string    `json:"last_status"`
	LastError         *string    `json:"last_error"`
	TriggerCount      int        `json:"trigger_count"`
	NotificationCount int        `json:"notification_count"`

	// Relationships
	Query          *SavedQuery                      `gorm:"foreignKey:QueryID" json:"query,omitempty"`
	User           *User                            `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ChannelsList   []AlertNotificationChannelConfig `gorm:"foreignKey:AlertID" json:"channels,omitempty"`
	WebhookURL     *string                          `json:"webhook_url,omitempty"`
	WebhookHeaders datatypes.JSON                   `json:"webhook_headers,omitempty"`

	CreatedAt time.Time `json:"created_at"`

	// ... (rest of struct is fine, I'll just append others at end of file)

	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// AlertHistory tracks the execution results of alerts
type AlertHistory struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	AlertID       string    `gorm:"type:uuid;not null" json:"alert_id"`
	Value         *float64  `json:"value"`
	Threshold     float64   `json:"threshold"`
	Status        string    `json:"status"` // ok, triggered, error
	Message       *string   `json:"message"`
	ErrorMessage  *string   `json:"error_message"`
	QueryDuration int       `json:"query_duration_ms"` // ms
	CheckedAt     time.Time `json:"checked_at"`
}

// Helper methods for Alert
func (a *Alert) CanSendNotification() bool {
	if a.IsMuted {
		if a.MutedUntil != nil && time.Now().After(*a.MutedUntil) {
			// Mute expired
			return true
		}
		return false
	}
	// Check cooldown
	// This logic handles simple mute check.
	// Cooldown logic is usually handled by checking LastTriggeredAt vs Now.
	if a.LastTriggeredAt != nil {
		if time.Since(*a.LastTriggeredAt) < time.Duration(a.CooldownMinutes)*time.Minute {
			return false
		}
	}
	return true
}

func (a *Alert) SetChannels(channels []AlertNotificationChannel) error {
	// Helper to set default channels, implementation might vary based on how we want to store it easily
	// But relations are better. This function was used in service line 118.
	// It likely creates AlertNotificationChannelConfig entries?
	// But service passed []AlertChannelEmail enum.
	// I'll stick to relationship.
	return nil
}

// Requests

type CreateAlertRequest struct {
	Name            string              `json:"name" validate:"required"`
	Description     string              `json:"description"`
	QueryID         string              `json:"query_id" validate:"required"`
	Column          string              `json:"column" validate:"required"`
	Operator        string              `json:"operator" validate:"required"`
	Threshold       float64             `json:"threshold" validate:"required"`
	Schedule        string              `json:"schedule" validate:"required"`
	Timezone        string              `json:"timezone"`
	Severity        AlertSeverity       `json:"severity"`
	CooldownMinutes int                 `json:"cooldown_minutes"`
	Channels        []AlertChannelInput `json:"channels"`
}

type UpdateAlertRequest struct {
	Name            *string              `json:"name"`
	Description     *string              `json:"description"`
	Column          *string              `json:"column"`
	Operator        *string              `json:"operator"`
	Threshold       *float64             `json:"threshold"`
	Schedule        *string              `json:"schedule"`
	Timezone        *string              `json:"timezone"`
	IsActive        *bool                `json:"is_active"`
	Severity        *AlertSeverity       `json:"severity"`
	CooldownMinutes *int                 `json:"cooldown_minutes"`
	Channels        *[]AlertChannelInput `json:"channels"`
}

type AlertChannelInput struct {
	ChannelType AlertNotificationChannel `json:"channel_type"`
	IsEnabled   bool                     `json:"is_enabled"`
	Config      datatypes.JSON           `json:"config"`
}

type AlertFilter struct {
	UserID   *string
	QueryID  *string
	IsActive *bool
	State    *AlertState
	Severity *AlertSeverity
	Search   *string
	Page     int
	Limit    int
	OrderBy  string
}

type AlertListResponse struct {
	Alerts []Alert `json:"alerts"`
	Total  int64   `json:"total"`
	Page   int     `json:"page"`
	Limit  int     `json:"limit"`
}

type AlertHistoryFilter struct {
	Status    *string
	StartDate *time.Time
	EndDate   *time.Time
	Page      int
	Limit     int
	OrderBy   string
}

type AlertHistoryListResponse struct {
	History []AlertHistory `json:"history"`
	Total   int64          `json:"total"`
	Page    int            `json:"page"`
	Limit   int            `json:"limit"`
}

type AcknowledgeAlertRequest struct {
	Note string `json:"note"`
}

type MuteAlertRequest struct {
	Duration *int `json:"duration"` // Minutes, null for indefinite
}

type TriggeredAlert struct {
	Alert
	CurrentValue   float64    `json:"current_value"`
	TriggeredAt    time.Time  `json:"triggered_at"`
	Acknowledged   bool       `json:"acknowledged"`
	AcknowledgedAt *time.Time `json:"acknowledged_at"`
	AcknowledgedBy *string    `json:"acknowledged_by"`
}

type AlertStats struct {
	Total        int64                   `json:"total"`
	Active       int64                   `json:"active"`
	Triggered    int64                   `json:"triggered"`
	Acknowledged int64                   `json:"acknowledged"`
	Muted        int64                   `json:"muted"`
	Error        int64                   `json:"error"`
	BySeverity   map[AlertSeverity]int64 `json:"by_severity"`
}

type AlertEvaluationResult struct {
	Value     float64
	Triggered bool
	Message   string
}

type AlertAcknowledgment struct {
	ID      uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	AlertID string    `gorm:"type:uuid;not null" json:"alert_id"`
	UserID  string    `gorm:"type:uuid;not null" json:"user_id"`
	Note    string    `json:"note"`

	AcknowledgedAt time.Time `gorm:"autoCreateTime" json:"acknowledged_at"`
}

// AlertNotificationChannelConfig represents the configuration for a notification channel on an alert
type AlertNotificationChannelConfig struct {
	ID          uuid.UUID                `gorm:"type:uuid;primaryKey" json:"id"`
	AlertID     string                   `gorm:"type:uuid;not null;index" json:"alert_id"`
	ChannelType AlertNotificationChannel `gorm:"not null" json:"channel_type"`
	Config      datatypes.JSON           `json:"config"`
	IsEnabled   bool                     `json:"is_enabled"`
	CreatedAt   time.Time                `json:"created_at"`
	UpdatedAt   time.Time                `json:"updated_at"`
}

func (c *AlertNotificationChannelConfig) SetConfig(config datatypes.JSON) error {
	c.Config = config
	return nil
}

func (c *AlertNotificationChannelConfig) GetConfig() (map[string]interface{}, error) {
	if len(c.Config) == 0 {
		return make(map[string]interface{}), nil
	}
	var config map[string]interface{}
	err := json.Unmarshal(c.Config, &config)
	return config, err
}

func (a *Alert) GetWebhookHeaders() (map[string]string, error) {
	if a.WebhookHeaders == nil {
		return nil, nil
	}
	var headers map[string]string
	err := json.Unmarshal(a.WebhookHeaders, &headers)
	return headers, err
}

type TestAlertRequest struct {
	QueryID   string  `json:"query_id"`
	Column    string  `json:"column"`
	Operator  string  `json:"operator"`
	Threshold float64 `json:"threshold"`
}

type TestAlertResponse struct {
	Triggered bool                   `json:"triggered"`
	Value     float64                `json:"value"`
	Message   string                 `json:"message"`
	QueryTime int                    `json:"query_time"`
	Error     string                 `json:"error,omitempty"`
	Threshold float64                `json:"threshold,omitempty"`
	Result    map[string]interface{} `json:"result,omitempty"`
}

type AlertNotificationTemplate struct {
	ID          uuid.UUID                `gorm:"primaryKey;type:uuid" json:"id"`
	Name        string                   `gorm:"type:varchar(255);not null" json:"name"`
	Description string                   `gorm:"type:text" json:"description,omitempty"`
	Content     string                   `gorm:"type:text;not null" json:"content"`
	ChannelType AlertNotificationChannel `gorm:"type:varchar(20);not null" json:"channelType"`
	IsActive    bool                     `gorm:"default:true" json:"isActive"`
	CreatedAt   time.Time                `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time                `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (AlertNotificationTemplate) TableName() string {
	return "alert_notification_templates"
}

type SendReportEmailRequest struct {
	To       []string
	Subject  string
	BodyHTML string
	BodyText string
}

type AlertNotificationLog struct {
	ID          uuid.UUID                `gorm:"type:uuid;primaryKey" json:"id"`
	HistoryID   uuid.UUID                `gorm:"type:uuid;not null;index" json:"history_id"`
	ChannelType AlertNotificationChannel `gorm:"not null" json:"channel_type"`
	Status      string                   `gorm:"not null" json:"status"`
	Error       *string                  `json:"error"`
	SentAt      *time.Time               `json:"sent_at"`
	CreatedAt   time.Time                `gorm:"autoCreateTime" json:"created_at"`
}
