package models

import (
	"time"
)

// ResourceType represents the type of resource being shared
type ResourceType string

const (
	ResourceTypeDashboard ResourceType = "dashboard"
	ResourceTypeQuery     ResourceType = "query"
)

// SharePermission represents the permission level for a share
type SharePermission string

const (
	SharePermissionView  SharePermission = "view"
	SharePermissionEdit  SharePermission = "edit"
	SharePermissionAdmin SharePermission = "admin"
)

// ShareStatus represents the status of a share
type ShareStatus string

const (
	ShareStatusActive  ShareStatus = "active"
	ShareStatusRevoked ShareStatus = "revoked"
	ShareStatusExpired ShareStatus = "expired"
	ShareStatusPending ShareStatus = "pending"
)

// Share represents a resource share between users
type Share struct {
	ID           string          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ResourceType ResourceType    `gorm:"type:varchar(50);not null;index" json:"resource_type"`
	ResourceID   string          `gorm:"type:text;not null;index" json:"resource_id"`
	SharedBy     string          `gorm:"type:text;not null;index" json:"shared_by"`
	SharedWith   *string         `gorm:"type:text;index" json:"shared_with,omitempty"`  // User ID for internal shares
	SharedEmail  *string         `gorm:"type:text;index" json:"shared_email,omitempty"` // Email for external invites
	Permission   SharePermission `gorm:"type:varchar(20);not null;default:'view'" json:"permission"`
	PasswordHash *string         `gorm:"type:text" json:"-"` // Never return password hash
	ExpiresAt    *time.Time      `gorm:"type:timestamp" json:"expires_at,omitempty"`
	Status       ShareStatus     `gorm:"type:varchar(20);not null;default:'active';index" json:"status"`
	AcceptedAt   *time.Time      `gorm:"type:timestamp" json:"accepted_at,omitempty"`
	Message      *string         `gorm:"type:text" json:"message,omitempty"`
	CreatedAt    time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time       `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	SharedByUser   User  `gorm:"foreignKey:SharedBy;references:ID" json:"shared_by_user,omitempty"`
	SharedWithUser *User `gorm:"foreignKey:SharedWith;references:ID" json:"shared_with_user,omitempty"`
}

// TableName specifies the table name for Share
func (Share) TableName() string {
	return "shares"
}

// IsExpired checks if the share has expired
func (s *Share) IsExpired() bool {
	if s.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*s.ExpiresAt)
}

// IsActive checks if the share is active and not expired
func (s *Share) IsActive() bool {
	if s.Status != ShareStatusActive {
		return false
	}
	return !s.IsExpired()
}

// RequiresPassword checks if the share requires a password
func (s *Share) RequiresPassword() bool {
	return s.PasswordHash != nil && *s.PasswordHash != ""
}

// HasEditPermission checks if the share grants edit permission
func (s *Share) HasEditPermission() bool {
	return s.Permission == SharePermissionEdit || s.Permission == SharePermissionAdmin
}

// HasAdminPermission checks if the share grants admin permission
func (s *Share) HasAdminPermission() bool {
	return s.Permission == SharePermissionAdmin
}

// ShareFilter represents filter criteria for querying shares
type ShareFilter struct {
	ResourceType   *ResourceType    `json:"resource_type,omitempty"`
	ResourceID     *string          `json:"resource_id,omitempty"`
	SharedBy       *string          `json:"shared_by,omitempty"`
	SharedWith     *string          `json:"shared_with,omitempty"`
	SharedEmail    *string          `json:"shared_email,omitempty"`
	Status         *ShareStatus     `json:"status,omitempty"`
	Permission     *SharePermission `json:"permission,omitempty"`
	IncludeExpired bool             `json:"include_expired,omitempty"`
	Limit          int              `json:"limit,omitempty"`
	Offset         int              `json:"offset,omitempty"`
}

// ShareCreateRequest represents the request to create a share
type ShareCreateRequest struct {
	ResourceType ResourceType    `json:"resource_type" validate:"required,oneof=dashboard query"`
	ResourceID   string          `json:"resource_id" validate:"required"`
	SharedWith   *string         `json:"shared_with,omitempty"`  // User ID
	SharedEmail  *string         `json:"shared_email,omitempty"` // Email for external invite
	Permission   SharePermission `json:"permission" validate:"required,oneof=view edit admin"`
	Password     *string         `json:"password,omitempty"`
	ExpiresAt    *time.Time      `json:"expires_at,omitempty"`
	Message      *string         `json:"message,omitempty"`
}

// ShareUpdateRequest represents the request to update a share
type ShareUpdateRequest struct {
	Permission *SharePermission `json:"permission,omitempty" validate:"omitempty,oneof=view edit admin"`
	Password   *string          `json:"password,omitempty"`
	ExpiresAt  *time.Time       `json:"expires_at,omitempty"`
	Message    *string          `json:"message,omitempty"`
}

// ShareAccessCheck represents the result of checking share access
type ShareAccessCheck struct {
	HasAccess        bool            `json:"has_access"`
	Permission       SharePermission `json:"permission,omitempty"`
	ShareID          string          `json:"share_id,omitempty"`
	RequiresPassword bool            `json:"requires_password"`
}

// ShareWithDetails represents a share with user details
type ShareWithDetails struct {
	Share
	ResourceName string `json:"resource_name,omitempty"`
}

// ValidateSharePermission validates a share permission string
func ValidateSharePermission(permission string) (SharePermission, bool) {
	switch SharePermission(permission) {
	case SharePermissionView, SharePermissionEdit, SharePermissionAdmin:
		return SharePermission(permission), true
	default:
		return "", false
	}
}

// ValidateResourceType validates a resource type string
func ValidateResourceType(resourceType string) (ResourceType, bool) {
	switch ResourceType(resourceType) {
	case ResourceTypeDashboard, ResourceTypeQuery:
		return ResourceType(resourceType), true
	default:
		return "", false
	}
}
