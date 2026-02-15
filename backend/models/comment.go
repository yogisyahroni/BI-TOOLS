package models

import (
	"encoding/json"
	"time"

	"gorm.io/datatypes"
)

// EntityType represents the type of entity a comment can be attached to
type CommentEntityType string

const (
	EntityTypeDashboard  CommentEntityType = "dashboard"
	EntityTypeQuery      CommentEntityType = "query"
	EntityTypeChart      CommentEntityType = "chart"
	EntityTypePipeline   CommentEntityType = "pipeline"
	EntityTypeDataflow   CommentEntityType = "dataflow"
	EntityTypeCollection CommentEntityType = "collection"
)

// Annotation represents a chart annotation with position data
type Annotation struct {
	ID        string         `json:"id" gorm:"primaryKey;type:text"`
	CommentID string         `json:"commentId" gorm:"not null;type:text;index"`
	ChartID   string         `json:"chartId" gorm:"not null;type:text;index"`
	XValue    *float64       `json:"xValue,omitempty" gorm:"type:double precision"`
	YValue    *float64       `json:"yValue,omitempty" gorm:"type:double precision"`
	XCategory *string        `json:"xCategory,omitempty" gorm:"type:text"`
	YCategory *string        `json:"yCategory,omitempty" gorm:"type:text"`
	Position  datatypes.JSON `json:"position" gorm:"type:jsonb"`           // {x: number, y: number} pixel position
	Type      string         `json:"type" gorm:"not null;default:'point'"` // "point", "range", "text"
	Color     string         `json:"color" gorm:"not null;default:'#F59E0B'"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

// TableName specifies the table name for Annotation
func (Annotation) TableName() string {
	return "annotations"
}

// Comment represents a comment on any entity with threading support
type Comment struct {
	ID         string            `json:"id" gorm:"primaryKey;type:text"`
	EntityType CommentEntityType `json:"entityType" gorm:"not null;type:text;index"`
	EntityID   string            `json:"entityId" gorm:"not null;type:text;index"`
	UserID     string            `json:"userId" gorm:"not null;type:text;index"`
	Content    string            `json:"content" gorm:"type:text;not null"`
	ParentID   *string           `json:"parentId,omitempty" gorm:"type:text;index"` // For threading/replies
	IsResolved bool              `json:"isResolved" gorm:"not null;default:false;index"`
	Mentions   datatypes.JSON    `json:"mentions,omitempty" gorm:"type:jsonb"` // Array of mentioned user IDs
	CreatedAt  time.Time         `json:"createdAt"`
	UpdatedAt  time.Time         `json:"updatedAt"`

	// Relationships
	User       User        `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID"`
	Parent     *Comment    `json:"parent,omitempty" gorm:"foreignKey:ParentID;references:ID"`
	Replies    []Comment   `json:"replies,omitempty" gorm:"foreignKey:ParentID;references:ID"`
	Annotation *Annotation `json:"annotation,omitempty" gorm:"foreignKey:CommentID;references:ID"`
}

// TableName specifies the table name for Comment
func (Comment) TableName() string {
	return "comments"
}

// IsReply returns true if this comment is a reply to another comment
func (c *Comment) IsReply() bool {
	return c.ParentID != nil && *c.ParentID != ""
}

// GetMentionedUserIDs extracts mentioned user IDs from the Mentions JSON field
func (c *Comment) GetMentionedUserIDs() []string {
	if c.Mentions == nil {
		return []string{}
	}

	var mentions []string
	if err := json.Unmarshal(c.Mentions, &mentions); err != nil {
		return []string{}
	}
	return mentions
}

// SetMentionedUserIDs sets the mentioned user IDs in the Mentions JSON field
func (c *Comment) SetMentionedUserIDs(userIDs []string) error {
	if userIDs == nil {
		userIDs = []string{}
	}
	mentions, err := json.Marshal(userIDs)
	if err != nil {
		return err
	}
	c.Mentions = mentions
	return nil
}

// CommentCreateRequest represents a request to create a comment
type CommentCreateRequest struct {
	EntityType string  `json:"entityType" binding:"required"`
	EntityID   string  `json:"entityId" binding:"required"`
	Content    string  `json:"content" binding:"required"`
	ParentID   *string `json:"parentId,omitempty"`
}

// CommentUpdateRequest represents a request to update a comment
type CommentUpdateRequest struct {
	Content string `json:"content" binding:"required"`
}

// CommentResolveRequest represents a request to resolve/unresolve a comment
type CommentResolveRequest struct {
	IsResolved bool `json:"isResolved"`
}

// CommentFilterRequest represents query parameters for filtering comments
type CommentFilterRequest struct {
	EntityType string `query:"entityType" binding:"required"`
	EntityID   string `query:"entityId" binding:"required"`
	ParentID   string `query:"parentId"` // "root" for top-level only, "*" for all
	IsResolved *bool  `query:"isResolved"`
	SortBy     string `query:"sortBy"`    // "date", "popular"
	SortOrder  string `query:"sortOrder"` // "asc", "desc"
	Limit      int    `query:"limit"`
	Offset     int    `query:"offset"`
}

// CommentMention represents a mention extracted from comment content
type CommentMention struct {
	Username string `json:"username"`
	UserID   string `json:"userId"`
	Position int    `json:"position"` // Position in the text where mention starts
}

// CommentWithDetails extends Comment with additional computed fields
type CommentWithDetails struct {
	Comment
	ReplyCount     int    `json:"replyCount"`
	MentionedUsers []User `json:"mentionedUsers,omitempty" gorm:"-"` // Not stored, computed on fetch
}

// CommentFilter provides filtering options for listing comments
type CommentFilter struct {
	EntityType *CommentEntityType
	EntityID   *string
	UserID     *string
	ParentID   *string // nil for top-level only, "*" for all, specific ID for replies
	IsResolved *bool
	Limit      int
	Offset     int
	SortBy     string // "date", "popular"
	SortOrder  string // "asc", "desc"
}

// AnnotationPosition represents the pixel position for annotations
type AnnotationPosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// AnnotationCreateRequest represents a request to create an annotation
type AnnotationCreateRequest struct {
	ChartID   string             `json:"chartId" binding:"required"`
	XValue    *float64           `json:"xValue,omitempty"`
	YValue    *float64           `json:"yValue,omitempty"`
	XCategory *string            `json:"xCategory,omitempty"`
	YCategory *string            `json:"yCategory,omitempty"`
	Position  AnnotationPosition `json:"position" binding:"required"`
	Type      string             `json:"type" binding:"required"` // "point", "range", "text"
	Color     string             `json:"color"`
	Content   string             `json:"content" binding:"required"` // Comment content
}

// ValidateEntityType validates if the entity type is supported
func ValidateEntityType(entityType string) (CommentEntityType, bool) {
	switch CommentEntityType(entityType) {
	case EntityTypeDashboard, EntityTypeQuery, EntityTypeChart,
		EntityTypePipeline, EntityTypeDataflow, EntityTypeCollection:
		return CommentEntityType(entityType), true
	default:
		return "", false
	}
}
