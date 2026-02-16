package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Dashboard represents a user's analytics dashboard
type Dashboard struct {
	ID           string          `gorm:"primaryKey;type:varchar(255)" json:"id"`
	Name         string          `gorm:"type:varchar(255);not null" json:"name"`
	Description  *string         `gorm:"type:text" json:"description"`
	CollectionID string          `gorm:"column:collectionId;type:varchar(255);not null;index" json:"collectionId"`
	UserID       string          `gorm:"column:userId;type:varchar(255);not null;index" json:"userId"`
	Filters      *string         `gorm:"type:jsonb" json:"filters"` // JSONB for filter configuration
	Layout       *datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"layout"`
	IsPublic     bool            `gorm:"default:false" json:"isPublic"`
	CreatedAt    time.Time       `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time       `gorm:"column:updatedAt;autoUpdateTime" json:"updatedAt"`

	// Certification (TASK-158)
	CertificationStatus string     `gorm:"type:varchar(50);default:'none'" json:"certificationStatus"` // none, verified, deprecated
	CertifiedBy         *string    `gorm:"type:varchar(255)" json:"certifiedBy,omitempty"`
	CertifiedAt         *time.Time `json:"certifiedAt,omitempty"`

	// Redundant fields to satisfy duplicate DB columns (snake_case)
	CollectionIDSnake string `gorm:"column:collection_id;type:varchar(255)" json:"-"`
	UserIDSnake       string `gorm:"column:user_id;type:varchar(255)" json:"-"`

	// Relationships (loaded on demand)
	Cards      []DashboardCard    `gorm:"foreignKey:DashboardID" json:"cards,omitempty"`
	Versions   []DashboardVersion `gorm:"foreignKey:DashboardID" json:"versions,omitempty"`
	Collection *Collection        `gorm:"foreignKey:CollectionID;references:ID" json:"collection,omitempty"`
	User       *User              `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
}

// BeforeCreate hook to populate redundant snake_case fields
func (d *Dashboard) BeforeCreate(tx *gorm.DB) (err error) {
	d.CollectionIDSnake = d.CollectionID
	d.UserIDSnake = d.UserID
	if d.CreatedAt.IsZero() {
		d.CreatedAt = time.Now()
	}
	if d.UpdatedAt.IsZero() {
		d.UpdatedAt = time.Now()
	}
	return
}

// TableName specifies the table name for Dashboard
func (Dashboard) TableName() string {
	return "Dashboard"
}

// DashboardVersion represents a snapshot of a dashboard at a specific point in time
type DashboardVersion struct {
	ID          string `gorm:"primaryKey;type:text" json:"id"`
	DashboardID string `gorm:"type:text;not null;index" json:"dashboardId"`
	Version     int    `gorm:"not null" json:"version"` // Auto-increment per dashboard

	// Snapshot data (JSONB)
	Name        string  `gorm:"type:varchar(255);not null" json:"name"`
	Description *string `gorm:"type:text" json:"description,omitempty"`
	FiltersJSON *string `gorm:"type:jsonb" json:"filtersJson,omitempty"`
	CardsJSON   string  `gorm:"type:jsonb;not null" json:"cardsJson"` // Array of cards
	LayoutJSON  *string `gorm:"type:jsonb" json:"layoutJson,omitempty"`

	// Metadata
	CreatedBy     string         `gorm:"type:text;not null" json:"createdBy"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	ChangeSummary string         `gorm:"type:text" json:"changeSummary"` // e.g., "Added 2 cards, modified layout"
	IsAutoSave    bool           `gorm:"default:false" json:"isAutoSave"`
	Metadata      datatypes.JSON `gorm:"type:jsonb" json:"metadata,omitempty"` // Additional metadata

	// Relationships
	Dashboard     *Dashboard `gorm:"foreignKey:DashboardID" json:"dashboard,omitempty"`
	CreatedByUser *User      `gorm:"foreignKey:CreatedBy;references:ID" json:"createdByUser,omitempty"`
}

// TableName specifies the table name for DashboardVersion
func (DashboardVersion) TableName() string {
	return "dashboard_versions"
}

// DashboardVersionCard represents a card snapshot within a dashboard version
type DashboardVersionCard struct {
	ID                  string          `json:"id"`
	QueryID             *string         `json:"queryId,omitempty"`
	Title               *string         `json:"title,omitempty"`
	Position            json.RawMessage `json:"position"` // {x: 0, y: 0, w: 6, h: 4}
	VisualizationConfig json.RawMessage `json:"visualizationConfig,omitempty"`
}

// DashboardVersionMetadata represents additional metadata for a version
type DashboardVersionMetadata struct {
	CardCount      int      `json:"cardCount"`
	FilterCount    int      `json:"filterCount"`
	CardsAdded     []string `json:"cardsAdded,omitempty"`    // IDs of added cards
	CardsRemoved   []string `json:"cardsRemoved,omitempty"`  // IDs of removed cards
	CardsModified  []string `json:"cardsModified,omitempty"` // IDs of modified cards
	LayoutChanged  bool     `json:"layoutChanged"`
	FiltersChanged bool     `json:"filtersChanged"`
}

// DashboardVersionFilter represents filter options for listing dashboard versions
type DashboardVersionFilter struct {
	DashboardID *string
	IsAutoSave  *bool
	CreatedBy   *string
	Limit       int
	Offset      int
	OrderBy     string // "date_desc", "date_asc", "version_desc", "version_asc"
}

// DashboardVersionCreateRequest represents a request to create a version
type DashboardVersionCreateRequest struct {
	ChangeSummary string                 `json:"changeSummary,omitempty"`
	IsAutoSave    bool                   `json:"isAutoSave"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// DashboardVersionCompareRequest represents a request to compare two versions
type DashboardVersionCompareRequest struct {
	VersionID1 string `json:"versionId1" binding:"required"`
	VersionID2 string `json:"versionId2" binding:"required"`
}

// DashboardVersionDiff represents the differences between two versions
type DashboardVersionDiff struct {
	Version1ID     string             `json:"version1Id"`
	Version2ID     string             `json:"version2Id"`
	NameChanged    bool               `json:"nameChanged"`
	NameFrom       *string            `json:"nameFrom,omitempty"`
	NameTo         *string            `json:"nameTo,omitempty"`
	DescChanged    bool               `json:"descChanged"`
	DescFrom       *string            `json:"descFrom,omitempty"`
	DescTo         *string            `json:"descTo,omitempty"`
	FiltersChanged bool               `json:"filtersChanged"`
	FiltersFrom    *string            `json:"filtersFrom,omitempty"`
	FiltersTo      *string            `json:"filtersTo,omitempty"`
	LayoutChanged  bool               `json:"layoutChanged"`
	LayoutFrom     *string            `json:"layoutFrom,omitempty"`
	LayoutTo       *string            `json:"layoutTo,omitempty"`
	CardsDiff      DashboardCardsDiff `json:"cardsDiff"`
}

// DashboardCardsDiff represents card-level differences
type DashboardCardsDiff struct {
	Added     []DashboardVersionCard `json:"added"`
	Removed   []DashboardVersionCard `json:"removed"`
	Modified  []DashboardCardChange  `json:"modified"`
	Unchanged []DashboardVersionCard `json:"unchanged"`
}

// DashboardCardChange represents a modified card with before/after
type DashboardCardChange struct {
	Before  DashboardVersionCard `json:"before"`
	After   DashboardVersionCard `json:"after"`
	Changes []string             `json:"changes"` // Fields that changed: ["position", "title"]
}

// DashboardVersionRestoreResponse represents the response after restoring a version
type DashboardVersionRestoreResponse struct {
	Success           bool   `json:"success"`
	Message           string `json:"message"`
	DashboardID       string `json:"dashboardId"`
	RestoredToVersion int    `json:"restoredToVersion"`
}

// GetMetadata parses and returns the version metadata
func (v *DashboardVersion) GetMetadata() (*DashboardVersionMetadata, error) {
	if v.Metadata == nil {
		return &DashboardVersionMetadata{}, nil
	}

	var metadata DashboardVersionMetadata
	if err := json.Unmarshal(v.Metadata, &metadata); err != nil {
		return nil, err
	}
	return &metadata, nil
}

// SetMetadata sets the version metadata
func (v *DashboardVersion) SetMetadata(metadata *DashboardVersionMetadata) error {
	if metadata == nil {
		v.Metadata = nil
		return nil
	}

	data, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	v.Metadata = datatypes.JSON(data)
	return nil
}

// GetCards parses and returns the cards JSON
func (v *DashboardVersion) GetCards() ([]DashboardVersionCard, error) {
	var cards []DashboardVersionCard
	if err := json.Unmarshal([]byte(v.CardsJSON), &cards); err != nil {
		return nil, err
	}
	return cards, nil
}

// NewDashboardVersion creates a new version from a dashboard
func NewDashboardVersion(dashboard *Dashboard, userID string, changeSummary string, isAutoSave bool) (*DashboardVersion, error) {
	// Serialize cards
	cardsData, err := json.Marshal(dashboard.Cards)
	if err != nil {
		return nil, err
	}

	version := &DashboardVersion{
		ID:            uuid.New().String(),
		DashboardID:   dashboard.ID,
		Name:          dashboard.Name,
		Description:   dashboard.Description,
		FiltersJSON:   dashboard.Filters,
		CardsJSON:     string(cardsData),
		CreatedBy:     userID,
		ChangeSummary: changeSummary,
		IsAutoSave:    isAutoSave,
	}

	return version, nil
}
