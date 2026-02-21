package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// DashboardCard represents a visualization or text card on a dashboard
type DashboardCard struct {
	ID                  uuid.UUID      `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	DashboardID         uuid.UUID      `gorm:"type:uuid;not null;index" json:"dashboardId"`
	QueryID             *uuid.UUID     `gorm:"type:uuid" json:"queryId"`
	Type                string         `gorm:"type:varchar(50);default:'visualization'" json:"type"` // visualization | text
	Title               *string        `gorm:"type:varchar(255)" json:"title"`
	TextContent         *string        `gorm:"type:text" json:"textContent"`
	Position            datatypes.JSON `gorm:"type:jsonb;not null" json:"position"` // {x, y, w, h}
	VisualizationConfig datatypes.JSON `gorm:"type:jsonb" json:"visualizationConfig"`
	CalculatedFields    datatypes.JSON `gorm:"type:jsonb" json:"calculatedFields"` // [{name: "Profit", formula: "[Sales]-[Cost]"}]
	CreatedAt           time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt           time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`

	// Relationships
	Dashboard *Dashboard  `gorm:"foreignKey:DashboardID" json:"dashboard,omitempty"`
	Query     *SavedQuery `gorm:"foreignKey:QueryID" json:"query,omitempty"`
}

// TableName specifies the table name for DashboardCard
func (DashboardCard) TableName() string {
	return "dashboard_cards"
}
