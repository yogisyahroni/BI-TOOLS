package models

import (
	"time"
)

// Organization status constants
const (
	OrgStatusActive    = "active"
	OrgStatusInactive  = "inactive"
	OrgStatusSuspended = "suspended"
)

// Organization role constants
const (
	OrgRoleOwner  = "owner"
	OrgRoleAdmin  = "admin"
	OrgRoleMember = "member"
	OrgRoleViewer = "viewer"
)

// Organization represents a tenant organization
type Organization struct {
	ID          string                 `gorm:"primaryKey;type:text" json:"id"`
	Name        string                 `gorm:"type:text;not null" json:"name"`
	Slug        string                 `gorm:"uniqueIndex;type:text;not null" json:"slug"`
	Description string                 `gorm:"type:text" json:"description,omitempty"`
	Logo        string                 `gorm:"type:text" json:"logo,omitempty"`
	Status      string                 `gorm:"type:text;default:'active'" json:"status"` // active, inactive, suspended
	Settings    map[string]interface{} `gorm:"type:jsonb;default:'{}'" json:"settings"`
	CreatedAt   time.Time              `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time              `gorm:"autoUpdateTime" json:"updatedAt"`

	// Relationships
	Members []OrganizationMember `gorm:"foreignKey:OrganizationID" json:"members,omitempty"`
	Quotas  *OrganizationQuota   `gorm:"foreignKey:OrganizationID" json:"quotas,omitempty"`
}

// TableName overrides the table name
func (Organization) TableName() string {
	return "organizations"
}

// OrganizationMember represents a user's membership in an organization
type OrganizationMember struct {
	ID             string    `gorm:"primaryKey;type:text" json:"id"`
	OrganizationID string    `gorm:"type:text;not null;index" json:"organizationId"`
	UserID         string    `gorm:"type:text;not null;index" json:"userId"`
	Role           string    `gorm:"type:text;not null" json:"role"` // owner, admin, member, viewer
	JoinedAt       time.Time `gorm:"autoCreateTime" json:"joinedAt"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	// Relationships
	User         *User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Organization *Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
}

// TableName overrides the table name
func (OrganizationMember) TableName() string {
	return "organization_members"
}

// OrganizationQuota represents usage quotas for an organization
type OrganizationQuota struct {
	ID             string `gorm:"primaryKey;type:text" json:"id"`
	OrganizationID string `gorm:"uniqueIndex;type:text;not null" json:"organizationId"`

	// Quota limits
	MaxUsers       int   `gorm:"default:10" json:"maxUsers"`
	MaxProjects    int   `gorm:"default:5" json:"maxProjects"`
	MaxQueries     int   `gorm:"default:100" json:"maxQueries"`
	MaxConnections int   `gorm:"default:5" json:"maxConnections"`
	MaxStorage     int64 `gorm:"default:1073741824" json:"maxStorage"` // 1GB in bytes

	// Current usage (updated regularly)
	CurrentUsers       int   `gorm:"default:0" json:"currentUsers"`
	CurrentProjects    int   `gorm:"default:0" json:"currentProjects"`
	CurrentQueries     int   `gorm:"default:0" json:"currentQueries"`
	CurrentConnections int   `gorm:"default:0" json:"currentConnections"`
	CurrentStorage     int64 `gorm:"default:0" json:"currentStorage"`

	// API limits
	ApiRequestsPerDay  int       `gorm:"default:10000" json:"apiRequestsPerDay"`
	CurrentApiRequests int       `gorm:"default:0" json:"currentApiRequests"`
	ApiLimitResetAt    time.Time `json:"apiLimitResetAt"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	// Relationships
	Organization *Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
}

// TableName overrides the table name
func (OrganizationQuota) TableName() string {
	return "organization_quotas"
}

// IsAtLimit checks if a specific quota limit has been reached
func (q *OrganizationQuota) IsAtLimit(resource string) bool {
	switch resource {
	case "users":
		return q.CurrentUsers >= q.MaxUsers
	case "projects":
		return q.CurrentProjects >= q.MaxProjects
	case "queries":
		return q.CurrentQueries >= q.MaxQueries
	case "connections":
		return q.CurrentConnections >= q.MaxConnections
	case "storage":
		return q.CurrentStorage >= q.MaxStorage
	case "api":
		return q.CurrentApiRequests >= q.ApiRequestsPerDay
	default:
		return false
	}
}

// GetUsagePercentage returns the usage percentage for a resource
func (q *OrganizationQuota) GetUsagePercentage(resource string) float64 {
	switch resource {
	case "users":
		if q.MaxUsers == 0 {
			return 0
		}
		return float64(q.CurrentUsers) / float64(q.MaxUsers) * 100
	case "projects":
		if q.MaxProjects == 0 {
			return 0
		}
		return float64(q.CurrentProjects) / float64(q.MaxProjects) * 100
	case "queries":
		if q.MaxQueries == 0 {
			return 0
		}
		return float64(q.CurrentQueries) / float64(q.MaxQueries) * 100
	case "connections":
		if q.MaxConnections == 0 {
			return 0
		}
		return float64(q.CurrentConnections) / float64(q.MaxConnections) * 100
	case "storage":
		if q.MaxStorage == 0 {
			return 0
		}
		return float64(q.CurrentStorage) / float64(q.MaxStorage) * 100
	case "api":
		if q.ApiRequestsPerDay == 0 {
			return 0
		}
		return float64(q.CurrentApiRequests) / float64(q.ApiRequestsPerDay) * 100
	default:
		return 0
	}
}
