package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// PulseChannelType defines the destination for the pulse
type PulseChannelType string

const (
	PulseChannelSlack PulseChannelType = "slack"
	PulseChannelTeams PulseChannelType = "teams"
	PulseChannelEmail PulseChannelType = "email"
)

// Pulse represents a scheduled screenshot delivery job
type Pulse struct {
	ID           uuid.UUID        `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Name         string           `gorm:"size:255;not null" json:"name"`
	DashboardID  uuid.UUID        `gorm:"type:uuid;not null;index" json:"dashboard_id"`
	UserID       uuid.UUID        `gorm:"type:uuid;not null;index" json:"user_id"` // Creator
	Schedule     string           `gorm:"size:100;not null" json:"schedule"`       // Cron expression
	ChannelType  PulseChannelType `gorm:"size:50;not null" json:"channel_type"`
	WebhookURL   string           `gorm:"type:text" json:"webhook_url"` // Encrypted? Ideally yes.
	Config       datatypes.JSON   `gorm:"type:jsonb" json:"config"`     // Width, height, filters
	IsActive     bool             `gorm:"default:true" json:"is_active"`
	LastRunAt    *time.Time       `json:"last_run_at"`
	NextRunAt    *time.Time       `json:"next_run_at"`
	FailureCount int              `gorm:"default:0" json:"failure_count"`
	LastError    string           `gorm:"type:text" json:"last_error"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
	DeletedAt    gorm.DeletedAt   `gorm:"index" json:"-"`
}

// TableName overrides the table name used by User to `pulses`
func (Pulse) TableName() string {
	return "pulses"
}
