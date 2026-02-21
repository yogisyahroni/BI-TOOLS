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
	ID           uuid.UUID       `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	Name         string          `gorm:"type:varchar(255);not null" json:"name"`
	Description  *string         `gorm:"type:text" json:"description"`
	CollectionID uuid.UUID       `gorm:"type:uuid;not null;index" json:"collectionId"`
	UserID       uuid.UUID       `gorm:"type:uuid;not null;index" json:"userId"`
	Filters      *string         `gorm:"type:jsonb" json:"filters"` // JSONB for filter configuration
	Layout       *datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"layout"`
	IsPublic     bool            `gorm:"default:false" json:"isPublic"`
	CreatedAt    time.Time       `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time       `gorm:"autoUpdateTime" json:"updatedAt"`

	// Certification (TASK-158)
	CertificationStatus string     `gorm:"type:varchar(50);default:'none'" json:"certificationStatus"` // none, verified, deprecated
	CertifiedBy         *uuid.UUID `gorm:"type:uuid" json:"certifiedBy,omitempty"`
	CertifiedAt         *time.Time `json:"certifiedAt,omitempty"`

	// Relationships (loaded on demand)
	Cards      []DashboardCard    `gorm:"foreignKey:DashboardID" json:"cards,omitempty"`
	Versions   []DashboardVersion `gorm:"foreignKey:DashboardID" json:"versions,omitempty"`
	Collection *Collection        `gorm:"foreignKey:CollectionID;references:ID" json:"collection,omitempty"`
	User       *User              `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
}

// TableName overrides the table name used by User to `dashboards`
func (Dashboard) TableName() string {
	return "dashboards"
}

// BeforeCreate hook
func (d *Dashboard) BeforeCreate(tx *gorm.DB) (err error) {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	if d.CreatedAt.IsZero() {
		d.CreatedAt = time.Now()
	}
	if d.UpdatedAt.IsZero() {
		d.UpdatedAt = time.Now()
	}
	return
}

// DashboardVersion represents a snapshot of a dashboard state
type DashboardVersion struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	DashboardID   uuid.UUID `gorm:"type:uuid;not null;index" json:"dashboardId"`
	Version       int       `gorm:"not null" json:"version"` // Incremental version number
	CreatedBy     uuid.UUID `gorm:"type:uuid;not null" json:"createdBy"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"createdAt"`
	ChangeSummary string    `gorm:"type:text" json:"changeSummary"`
	IsAutoSave    bool      `gorm:"default:false" json:"isAutoSave"`

	// Snapshot Data (JSONB)
	Name        string  `gorm:"type:varchar(255);not null" json:"name"`
	Description *string `gorm:"type:text" json:"description"`
	FiltersJSON *string `gorm:"type:jsonb" json:"filters"`
	CardsJSON   string  `gorm:"type:jsonb;not null" json:"cards"` // Serialized []DashboardVersionCard
	LayoutJSON  string  `gorm:"type:jsonb;default:'{}'" json:"layout"`

	// Metadata
	MetadataJSON string `gorm:"type:jsonb" json:"metadata"` // e.g. {"restored_from": 5}

	// Relationships
	Dashboard     *Dashboard `gorm:"foreignKey:DashboardID" json:"-"`
	CreatedByUser *User      `gorm:"foreignKey:CreatedBy" json:"createdByUser,omitempty"`
}

// TableName overrides the table name used by DashboardVersion to `dashboard_versions`
func (DashboardVersion) TableName() string {
	return "dashboard_versions"
}

// BeforeCreate hook
func (dv *DashboardVersion) BeforeCreate(tx *gorm.DB) (err error) {
	if dv.ID == uuid.Nil {
		dv.ID = uuid.New()
	}
	return
}

// SetMetadata sets the metadata from a struct
func (dv *DashboardVersion) SetMetadata(meta interface{}) error {
	bytes, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	dv.MetadataJSON = string(bytes)
	return nil
}

// GetMetadata retrieves the metadata into a struct
func (dv *DashboardVersion) GetMetadata(target interface{}) error {
	if dv.MetadataJSON == "" {
		return nil
	}
	return json.Unmarshal([]byte(dv.MetadataJSON), target)
}

// GetCards retrieves the cards from the snapshot
func (dv *DashboardVersion) GetCards() ([]DashboardVersionCard, error) {
	var cards []DashboardVersionCard
	if dv.CardsJSON == "" {
		return cards, nil
	}
	err := json.Unmarshal([]byte(dv.CardsJSON), &cards)
	return cards, err
}

// DashboardVersionCard represents a card snapshot within a version
// This struct is used for JSON serialization inside DashboardVersion
type DashboardVersionCard struct {
	ID                  string          `json:"id"`
	QueryID             *string         `json:"queryId,omitempty"`
	Title               *string         `json:"title,omitempty"`
	Type                string          `json:"type"`
	Position            json.RawMessage `json:"position"`
	VisualizationConfig json.RawMessage `json:"visualizationConfig,omitempty"`
	TextContent         *string         `json:"textContent,omitempty"`
}

// DashboardVersionMetadata represents metadata stored in a version
type DashboardVersionMetadata struct {
	CardCount   int `json:"cardCount"`
	FilterCount int `json:"filterCount"`
}

// NewDashboardVersion creates a version snapshot from a live dashboard
func NewDashboardVersion(dashboard *Dashboard, userID string, summary string, isAutoSave bool) (*DashboardVersion, error) {
	// Serialize cards to simplified version struct
	versionCards := make([]DashboardVersionCard, len(dashboard.Cards))
	for i, card := range dashboard.Cards {
		vc := DashboardVersionCard{
			ID:          card.ID.String(),
			Title:       card.Title,
			Type:        card.Type,
			Position:    json.RawMessage(card.Position),
			TextContent: card.TextContent,
		}
		if card.QueryID != nil {
			qid := card.QueryID.String()
			vc.QueryID = &qid
		}
		if len(card.VisualizationConfig) > 0 {
			vc.VisualizationConfig = json.RawMessage(card.VisualizationConfig)
		}
		versionCards[i] = vc
	}

	cardsBytes, err := json.Marshal(versionCards)
	if err != nil {
		return nil, err
	}

	layoutBytes := []byte("{}")
	if dashboard.Layout != nil {
		layoutBytes = []byte(*dashboard.Layout)
	}

	return &DashboardVersion{
		ID:            uuid.New(),
		DashboardID:   dashboard.ID,
		CreatedBy:     uuid.MustParse(userID),
		CreatedAt:     time.Now(),
		ChangeSummary: summary,
		IsAutoSave:    isAutoSave,
		Name:          dashboard.Name,
		Description:   dashboard.Description,
		FiltersJSON:   dashboard.Filters,
		CardsJSON:     string(cardsBytes),
		LayoutJSON:    string(layoutBytes),
	}, nil
}

// CreateRequest for API
type DashboardVersionCreateRequest struct {
	ChangeSummary string                 `json:"changeSummary"`
	IsAutoSave    bool                   `json:"isAutoSave"`
	Metadata      map[string]interface{} `json:"metadata"`
}

type DashboardVersionFilter struct {
	IsAutoSave *bool      `form:"isAutoSave"`
	CreatedBy  *uuid.UUID `form:"createdBy"`
	Limit      int        `form:"limit"`
	Offset     int        `form:"offset"`
	OrderBy    string     `form:"orderBy"` // date_desc, date_asc, version_desc
}

type DashboardVersionRestoreResponse struct {
	Success           bool   `json:"success"`
	Message           string `json:"message"`
	DashboardID       string `json:"dashboardId"`
	RestoredToVersion int    `json:"restoredToVersion"`
}

type DashboardVersionDiff struct {
	Version1ID     string `json:"version1Id"`
	Version2ID     string `json:"version2Id"`
	NameChanged    bool   `json:"nameChanged"`
	NameFrom       *string
	NameTo         *string
	DescChanged    bool `json:"descChanged"`
	DescFrom       *string
	DescTo         *string
	FiltersChanged bool `json:"filtersChanged"`
	FiltersFrom    *string
	FiltersTo      *string
	CardsDiff      DashboardCardsDiff `json:"cardsDiff"`
}

type DashboardCardsDiff struct {
	Added     []DashboardVersionCard `json:"added"`
	Removed   []DashboardVersionCard `json:"removed"`
	Modified  []DashboardCardChange  `json:"modified"`
	Unchanged []DashboardVersionCard `json:"unchanged"`
}

type DashboardCardChange struct {
	Before  DashboardVersionCard `json:"before"`
	After   DashboardVersionCard `json:"after"`
	Changes []string             `json:"changes"` // List of changed fields: title, query, position, visualization
}
