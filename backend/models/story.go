package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Story represents a generated presentation story
type Story struct {
	ID          string         `gorm:"primaryKey" json:"id"`
	UserID      string         `gorm:"index;not null" json:"user_id"`
	DashboardID *string        `gorm:"index" json:"dashboard_id,omitempty"`
	Title       string         `gorm:"not null" json:"title"`
	Description string         `json:"description"`
	Content     datatypes.JSON `json:"content"` // Stores the JSON representation of SlideDeck
	ProviderID  string         `json:"provider_id"`
	Prompt      string         `json:"prompt"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`

	// Relationships
	User      User       `gorm:"foreignKey:UserID" json:"-"`
	Dashboard *Dashboard `gorm:"foreignKey:DashboardID" json:"-"`
}

// BeforeCreate hooks into GORM before creating a new record
func (s *Story) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return
}
