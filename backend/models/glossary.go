package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BusinessTerm represents a definition in the business glossary
type BusinessTerm struct {
	ID          string         `gorm:"primaryKey" json:"id"`
	WorkspaceID string         `gorm:"index;not null" json:"workspace_id"`
	Name        string         `gorm:"not null" json:"name"`
	Definition  string         `gorm:"type:text;not null" json:"definition"`
	Synonyms    StringArray    `gorm:"type:text[]" json:"synonyms"` // Postgres array
	OwnerID     string         `json:"owner_id"`
	Status      string         `json:"status"` // draft, approved, deprecated
	Tags        StringArray    `gorm:"type:text[]" json:"tags"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	RelatedColumns []TermColumnMapping `json:"related_columns,omitempty"`
}

// TermColumnMapping links a business term to a physical column or semantic metric
type TermColumnMapping struct {
	ID           string    `gorm:"primaryKey" json:"id"`
	TermID       string    `gorm:"index;not null" json:"term_id"`
	DataSourceID string    `json:"data_source_id"` // Optional: if linking to physical table
	TableName    string    `json:"table_name"`
	ColumnName   string    `json:"column_name"`
	MetricID     *string   `json:"metric_id,omitempty"` // Optional: if linking to semantic metric
	CreatedAt    time.Time `json:"created_at"`
}

// BeforeCreate hooks
func (t *BusinessTerm) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return
}

func (m *TermColumnMapping) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return
}
