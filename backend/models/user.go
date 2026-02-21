package models

import (
	"time"

	"github.com/google/uuid"
)

// User status constants
const (
	UserStatusActive   = "active"
	UserStatusInactive = "inactive"
	UserStatusPending  = "pending"
)

// User represents a system user
type User struct {
	ID                       uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Email                    string     `gorm:"uniqueIndex;not null" json:"email"`
	Username                 string     `gorm:"uniqueIndex;type:text" json:"username"`
	Name                     string     `gorm:"type:text" json:"name"`
	Password                 string     `gorm:"type:text" json:"-"`                   // Never return password, nullable for OAuth
	Role                     string     `gorm:"type:text;default:'user'" json:"role"` // user, admin
	EmailVerified            bool       `gorm:"column:email_verified;default:false" json:"emailVerified"`
	EmailVerifiedAt          *time.Time `gorm:"column:email_verified_at;type:timestamp" json:"emailVerifiedAt,omitempty"`
	EmailVerificationToken   string     `gorm:"column:email_verification_token;type:text;index" json:"-"`  // Never return token
	EmailVerificationExpires *time.Time `gorm:"column:email_verification_expires;type:timestamp" json:"-"` // Never return expiration
	PasswordResetToken       string     `gorm:"column:password_reset_token;type:text;index" json:"-"`      // Never return token
	PasswordResetExpires     *time.Time `gorm:"column:password_reset_expires;type:timestamp" json:"-"`     // Never return expiration
	// OAuth fields
	Provider   string    `gorm:"column:provider;type:text;index" json:"provider,omitempty"`      // e.g., "google", "github"
	ProviderID string    `gorm:"column:provider_id;type:text;index" json:"providerId,omitempty"` // OAuth provider user ID
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	// Status
	Status             string     `gorm:"type:text;default:'active'" json:"status"` // active, inactive, pending
	DeactivatedAt      *time.Time `gorm:"type:timestamp" json:"deactivatedAt,omitempty"`
	DeactivatedBy      *string    `gorm:"type:text" json:"deactivatedBy,omitempty"`
	DeactivationReason *string    `gorm:"type:text" json:"deactivationReason,omitempty"`

	// Impersonation
	ImpersonationToken   string     `gorm:"type:text;index" json:"-"`
	ImpersonationExpires *time.Time `gorm:"type:timestamp" json:"-"`
	ImpersonatedBy       *string    `gorm:"type:text" json:"-"`

	// RBAC
	Roles []Role `gorm:"many2many:user_roles;foreignKey:ID;joinForeignKey:UserID;References:ID;joinReferences:RoleID" json:"roles,omitempty"`
}

// TableName overrides the table name
func (User) TableName() string {
	return "users"
}

// IsSuperAdmin checks if the user has the 'admin' role (legacy) or 'Super Admin' role
func (u *User) IsSuperAdmin() bool {
	if u.Role == "admin" {
		return true
	}
	for _, r := range u.Roles {
		if r.Name == "Super Admin" || r.Name == "Admin" {
			return true
		}
	}
	return false
}

// IsActive checks if the user account is active
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// CanLogin checks if the user can login (active and either verified or OAuth)
func (u *User) CanLogin() bool {
	if !u.IsActive() {
		return false
	}
	// OAuth users can login without email verification
	if u.Provider != "" {
		return true
	}
	return u.EmailVerified
}
