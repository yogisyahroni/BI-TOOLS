package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// WebhookRequest represents a webhook creation/update request
type WebhookRequest struct {
	Name        string            `json:"name" validate:"required"`
	URL         string            `json:"url" validate:"required,url"`
	Events      []string          `json:"events" validate:"required,min=1"`
	IsActive    bool              `json:"is_active"`
	Headers     map[string]string `json:"headers"`
	Description string            `json:"description"`
}

// Webhook represents a registered webhook
type Webhook struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Name            string         `gorm:"size:255;not null" json:"name"`
	URL             string         `gorm:"not null" json:"url"`
	Events          datatypes.JSON `gorm:"type:jsonb" json:"events"`        // Array of event strings
	Secret          string         `gorm:"size:255;not null" json:"secret"` // HMAC secret
	IsActive        bool           `gorm:"default:true" json:"is_active"`
	Headers         datatypes.JSON `gorm:"type:jsonb" json:"headers"` // Custom headers map[string]string
	Description     string         `gorm:"type:text" json:"description"`
	LastTriggeredAt *time.Time     `json:"last_triggered_at"`
	FailureCount    int            `gorm:"default:0" json:"failure_count"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
	UserID          uuid.UUID      `gorm:"type:uuid;not null" json:"user_id"`
}

// Prepare sets default values before saving
func (w *Webhook) BeforeCreate(tx *gorm.DB) (err error) {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	return
}

// WebhookLog represents a log of a webhook dispatch attempt
type WebhookLog struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	WebhookID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"webhook_id"`
	Webhook        Webhook        `gorm:"foreignKey:WebhookID" json:"-"`
	EventType      string         `gorm:"size:255;not null" json:"event_type"`
	RequestPayload datatypes.JSON `gorm:"type:jsonb" json:"request_payload"`
	ResponseStatus int            `json:"response_status"`
	ResponseBody   string         `gorm:"type:text" json:"response_body"`
	DurationMs     int64          `json:"duration_ms"`
	Status         string         `gorm:"size:50;not null" json:"status"` // "success", "failure"
	ErrorMessage   string         `gorm:"type:text" json:"error_message,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
}

func (l *WebhookLog) BeforeCreate(tx *gorm.DB) (err error) {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return
}
