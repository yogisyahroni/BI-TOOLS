package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WebhookChannelType represents the type of external notification channel
type WebhookChannelType string

const (
	WebhookChannelSlack WebhookChannelType = "slack"
	WebhookChannelTeams WebhookChannelType = "teams"
)

// WebhookConfig stores external webhook configurations for Slack/Teams notifications
type WebhookConfig struct {
	ID          uuid.UUID          `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      uuid.UUID          `json:"userId" gorm:"type:uuid;not null;index"`
	Name        string             `json:"name" gorm:"type:varchar(255);not null"`
	ChannelType WebhookChannelType `json:"channelType" gorm:"type:varchar(20);not null"`
	WebhookURL  string             `json:"webhookUrl" gorm:"type:text;not null"`
	IsActive    bool               `json:"isActive" gorm:"default:true"`
	Description string             `json:"description,omitempty" gorm:"type:text"`
	CreatedAt   time.Time          `json:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt     `json:"-" gorm:"index"`
}

// TableName specifies the table name for GORM
func (WebhookConfig) TableName() string {
	return "webhook_configs"
}
