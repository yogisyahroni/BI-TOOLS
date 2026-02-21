package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Collection struct {
	ID          uuid.UUID `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Name        string    `json:"name" gorm:"not null"`
	Description *string   `json:"description"`
	UserID      uuid.UUID `json:"userId" gorm:"type:uuid;not null"`
	WorkspaceID *string   `json:"workspaceId"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	// Relations
	Items []CollectionItem `json:"items,omitempty" gorm:"foreignKey:CollectionID"`
}

func (c *Collection) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return
}

type CollectionItem struct {
	ID           uuid.UUID `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	CollectionID uuid.UUID `json:"collectionId" gorm:"type:uuid;not null"`
	ItemType     string    `json:"itemType" gorm:"not null"` // 'pipeline' or 'dataflow'
	ItemID       string    `json:"itemId" gorm:"not null"`
	CreatedAt    time.Time `json:"createdAt"`
}

func (ci *CollectionItem) BeforeCreate(tx *gorm.DB) (err error) {
	if ci.ID == uuid.Nil {
		ci.ID = uuid.New()
	}
	return
}
