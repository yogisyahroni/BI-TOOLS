package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// SavedQuery represents a saved SQL query
type SavedQuery struct {
	ID                  string    `gorm:"primaryKey;type:text" json:"id"`
	Name                string    `gorm:"type:text;not null" json:"name"`
	Description         *string   `gorm:"type:text" json:"description"`
	SQL                 string    `gorm:"type:text;not null" json:"sql"`
	AIPrompt            *string   `gorm:"type:text" json:"aiPrompt"`
	ConnectionID        string    `gorm:"type:text;not null" json:"connectionId"`
	CollectionID        string    `gorm:"type:text;not null" json:"collectionId"`
	UserID              string    `gorm:"type:text;not null" json:"userId"`
	VisualizationConfig []byte    `gorm:"type:jsonb" json:"visualizationConfig"` // JSON
	Tags                []string  `gorm:"type:text[]" json:"tags"`
	Pinned              bool      `gorm:"default:false" json:"pinned"`
	BusinessMetricID    *string   `gorm:"type:text" json:"businessMetricId"`
	CreatedAt           time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt           time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	// Relationships (optional for queries)
	Connection *Connection    `gorm:"foreignKey:ConnectionID" json:"connection,omitempty"`
	Versions   []QueryVersion `gorm:"foreignKey:QueryID" json:"versions,omitempty"`
}

// TableName overrides the table name
func (SavedQuery) TableName() string {
	return "saved_queries"
}

// QueryVersion represents a snapshot of a query at a specific point in time
type QueryVersion struct {
	ID      string `gorm:"primaryKey;type:text" json:"id"`
	QueryID string `gorm:"type:text;not null;index" json:"queryId"`
	Version int    `gorm:"not null" json:"version"` // Auto-increment per query

	// Snapshot data
	Name                string         `gorm:"type:text;not null" json:"name"`
	Description         *string        `gorm:"type:text" json:"description,omitempty"`
	SQL                 string         `gorm:"type:text;not null" json:"sql"`
	AIPrompt            *string        `gorm:"type:text" json:"aiPrompt,omitempty"`
	VisualizationConfig datatypes.JSON `gorm:"type:jsonb" json:"visualizationConfig,omitempty"`
	Tags                datatypes.JSON `gorm:"type:jsonb" json:"tags,omitempty"`

	// Metadata
	CreatedBy     string         `gorm:"type:text;not null" json:"createdBy"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	ChangeSummary string         `gorm:"type:text" json:"changeSummary"` // e.g., "Modified SQL query"
	IsAutoSave    bool           `gorm:"default:false" json:"isAutoSave"`
	Metadata      datatypes.JSON `gorm:"type:jsonb" json:"metadata,omitempty"` // Additional metadata

	// Relationships
	Query         *SavedQuery `gorm:"foreignKey:QueryID" json:"query,omitempty"`
	CreatedByUser *User       `gorm:"foreignKey:CreatedBy;references:ID" json:"createdByUser,omitempty"`
}

// TableName specifies the table name for QueryVersion
func (QueryVersion) TableName() string {
	return "query_versions"
}

// QueryVersionMetadata represents additional metadata for a query version
type QueryVersionMetadata struct {
	SQLChanged      bool     `json:"sqlChanged"`
	MetadataChanged bool     `json:"metadataChanged"`
	ConfigChanged   bool     `json:"configChanged"`
	TagsChanged     bool     `json:"tagsChanged"`
	SQLDiffSummary  string   `json:"sqlDiffSummary,omitempty"` // Brief summary of SQL changes
	TagsAdded       []string `json:"tagsAdded,omitempty"`
	TagsRemoved     []string `json:"tagsRemoved,omitempty"`
}

// QueryVersionFilter represents filter options for listing query versions
type QueryVersionFilter struct {
	QueryID    *string
	IsAutoSave *bool
	CreatedBy  *string
	Limit      int
	Offset     int
	OrderBy    string // "date_desc", "date_asc", "version_desc", "version_asc"
}

// QueryVersionCreateRequest represents a request to create a version
type QueryVersionCreateRequest struct {
	ChangeSummary string                 `json:"changeSummary,omitempty"`
	IsAutoSave    bool                   `json:"isAutoSave"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// QueryVersionDiff represents the differences between two query versions
type QueryVersionDiff struct {
	Version1ID           string                 `json:"version1Id"`
	Version2ID           string                 `json:"version2Id"`
	NameChanged          bool                   `json:"nameChanged"`
	NameFrom             *string                `json:"nameFrom,omitempty"`
	NameTo               *string                `json:"nameTo,omitempty"`
	DescChanged          bool                   `json:"descChanged"`
	DescFrom             *string                `json:"descFrom,omitempty"`
	DescTo               *string                `json:"descTo,omitempty"`
	SQLChanged           bool                   `json:"sqlChanged"`
	SQLFrom              *string                `json:"sqlFrom,omitempty"`
	SQLTo                *string                `json:"sqlTo,omitempty"`
	AIPromptChanged      bool                   `json:"aiPromptChanged"`
	AIPromptFrom         *string                `json:"aiPromptFrom,omitempty"`
	AIPromptTo           *string                `json:"aiPromptTo,omitempty"`
	VisualizationChanged bool                   `json:"visualizationChanged"`
	VisualizationFrom    map[string]interface{} `json:"visualizationFrom,omitempty"`
	VisualizationTo      map[string]interface{} `json:"visualizationTo,omitempty"`
	TagsChanged          bool                   `json:"tagsChanged"`
	TagsAdded            []string               `json:"tagsAdded"`
	TagsRemoved          []string               `json:"tagsRemoved"`
}

// QueryVersionRestoreResponse represents the response after restoring a query version
type QueryVersionRestoreResponse struct {
	Success           bool   `json:"success"`
	Message           string `json:"message"`
	QueryID           string `json:"queryId"`
	RestoredToVersion int    `json:"restoredToVersion"`
}

// GetMetadata parses and returns the version metadata
func (v *QueryVersion) GetMetadata() (*QueryVersionMetadata, error) {
	if v.Metadata == nil {
		return &QueryVersionMetadata{}, nil
	}

	var metadata QueryVersionMetadata
	if err := json.Unmarshal(v.Metadata, &metadata); err != nil {
		return nil, err
	}
	return &metadata, nil
}

// SetMetadata sets the version metadata
func (v *QueryVersion) SetMetadata(metadata *QueryVersionMetadata) error {
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

// GetTags parses and returns the tags
func (v *QueryVersion) GetTags() ([]string, error) {
	if v.Tags == nil {
		return []string{}, nil
	}

	var tags []string
	if err := json.Unmarshal(v.Tags, &tags); err != nil {
		return nil, err
	}
	return tags, nil
}

// GetVisualizationConfig parses and returns the visualization config
func (v *QueryVersion) GetVisualizationConfig() (map[string]interface{}, error) {
	if v.VisualizationConfig == nil {
		return nil, nil
	}

	var config map[string]interface{}
	if err := json.Unmarshal(v.VisualizationConfig, &config); err != nil {
		return nil, err
	}
	return config, nil
}

// NewQueryVersion creates a new version from a query
func NewQueryVersion(query *SavedQuery, userID string, changeSummary string, isAutoSave bool) (*QueryVersion, error) {
	// Serialize tags
	tagsData, err := json.Marshal(query.Tags)
	if err != nil {
		return nil, err
	}

	version := &QueryVersion{
		ID:            uuid.New().String(),
		QueryID:       query.ID,
		Name:          query.Name,
		Description:   query.Description,
		SQL:           query.SQL,
		AIPrompt:      query.AIPrompt,
		Tags:          tagsData,
		CreatedBy:     userID,
		ChangeSummary: changeSummary,
		IsAutoSave:    isAutoSave,
	}

	// Set visualization config if present
	if len(query.VisualizationConfig) > 0 {
		version.VisualizationConfig = query.VisualizationConfig
	}

	return version, nil
}
