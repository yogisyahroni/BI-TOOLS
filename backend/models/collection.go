package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Collection struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Description *string   `json:"description"`
	UserID      string    `json:"userId" gorm:"not null"`
	WorkspaceID *string   `json:"workspaceId"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	// Relations
	Items []CollectionItem `json:"items,omitempty" gorm:"foreignKey:CollectionID"`
}

func (c *Collection) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return
}

type CollectionItem struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	CollectionID string    `json:"collectionId" gorm:"not null"`
	ItemType     string    `json:"itemType" gorm:"not null"` // 'pipeline' or 'dataflow'
	ItemID       string    `json:"itemId" gorm:"not null"`
	CreatedAt    time.Time `json:"createdAt"`
}

func (ci *CollectionItem) BeforeCreate(tx *gorm.DB) (err error) {
	if ci.ID == "" {
		ci.ID = uuid.New().String()
	}
	return
}
