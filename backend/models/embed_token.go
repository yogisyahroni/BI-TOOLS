package models

import (
	"encoding/json"
	"time"
)

// EmbedToken represents a token for embedding resources
type EmbedToken struct {
	ID             string      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ResourceType   string      `gorm:"type:varchar(50);not null;index" json:"resource_type"`
	ResourceID     string      `gorm:"type:text;not null;index" json:"resource_id"`
	Token          string      `gorm:"type:text;not null;uniqueIndex" json:"token"`
	CreatedBy      string      `gorm:"type:text;not null;index" json:"created_by"`
	AllowedDomains StringArray `gorm:"type:jsonb;default:'[]'" json:"allowed_domains"`
	AllowedIPs     StringArray `gorm:"type:jsonb;default:'[]'" json:"allowed_ips"`
	ExpiresAt      *time.Time  `gorm:"type:timestamp" json:"expires_at,omitempty"`
	ViewCount      int64       `gorm:"default:0" json:"view_count"`
	LastViewedAt   *time.Time  `gorm:"type:timestamp" json:"last_viewed_at,omitempty"`
	IsRevoked      bool        `gorm:"default:false" json:"is_revoked"`
	RevokedAt      *time.Time  `gorm:"type:timestamp" json:"revoked_at,omitempty"`
	RevokedBy      *string     `gorm:"type:text" json:"revoked_by,omitempty"`
	Description    *string     `gorm:"type:text" json:"description,omitempty"`
	CreatedAt      time.Time   `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time   `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	Creator User `gorm:"foreignKey:CreatedBy;references:ID" json:"creator,omitempty"`
}

// TableName specifies the table name for EmbedToken
func (EmbedToken) TableName() string {
	return "embed_tokens"
}

// IsExpired checks if the token has expired
func (e *EmbedToken) IsExpired() bool {
	if e.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*e.ExpiresAt)
}

// IsActive checks if the token is active (not expired or revoked)
func (e *EmbedToken) IsActive() bool {
	if e.IsRevoked {
		return false
	}
	return !e.IsExpired()
}

// HasDomainRestriction checks if the token has domain restrictions
func (e *EmbedToken) HasDomainRestriction() bool {
	return len(e.AllowedDomains) > 0
}

// HasIPRestriction checks if the token has IP restrictions
func (e *EmbedToken) HasIPRestriction() bool {
	return len(e.AllowedIPs) > 0
}

// IsDomainAllowed checks if a domain is allowed by this token
func (e *EmbedToken) IsDomainAllowed(domain string) bool {
	if !e.HasDomainRestriction() {
		return true
	}

	for _, allowedDomain := range e.AllowedDomains {
		// Exact match
		if allowedDomain == domain {
			return true
		}
		// Wildcard subdomain match (e.g., *.example.com)
		if len(allowedDomain) > 2 && allowedDomain[:2] == "*." {
			suffix := allowedDomain[2:]
			if len(domain) > len(suffix) && domain[len(domain)-len(suffix):] == suffix {
				return true
			}
		}
	}
	return false
}

// IsIPAllowed checks if an IP is allowed by this token
func (e *EmbedToken) IsIPAllowed(ip string) bool {
	if !e.HasIPRestriction() {
		return true
	}

	for _, allowedIP := range e.AllowedIPs {
		if allowedIP == ip {
			return true
		}
	}
	return false
}

// IncrementViewCount increments the view count and updates last viewed time
func (e *EmbedToken) IncrementViewCount() {
	e.ViewCount++
	now := time.Now()
	e.LastViewedAt = &now
}

// StringArray represents an array of strings for JSONB storage
type StringArray []string

// Value implements the driver.Valuer interface
func (a StringArray) Value() (interface{}, error) {
	if a == nil {
		return []byte("[]"), nil
	}
	return a, nil
}

// Scan implements the sql.Scanner interface
func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = StringArray{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		// Try to unmarshal as JSON array
		var arr []string
		if err := json.Unmarshal(v, &arr); err == nil {
			*a = StringArray(arr)
			return nil
		}
	case []interface{}:
		arr := make([]string, len(v))
		for i, item := range v {
			if s, ok := item.(string); ok {
				arr[i] = s
			}
		}
		*a = StringArray(arr)
		return nil
	}

	*a = StringArray{}
	return nil
}

// EmbedTokenCreateRequest represents the request to create an embed token
type EmbedTokenCreateRequest struct {
	ResourceType   string     `json:"resource_type" validate:"required,oneof=dashboard query"`
	ResourceID     string     `json:"resource_id" validate:"required"`
	AllowedDomains []string   `json:"allowed_domains,omitempty"`
	AllowedIPs     []string   `json:"allowed_ips,omitempty"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	Description    *string    `json:"description,omitempty"`
}

// EmbedTokenUpdateRequest represents the request to update an embed token
type EmbedTokenUpdateRequest struct {
	AllowedDomains []string   `json:"allowed_domains,omitempty"`
	AllowedIPs     []string   `json:"allowed_ips,omitempty"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	Description    *string    `json:"description,omitempty"`
}

// EmbedTokenValidationResult represents the result of validating an embed token
type EmbedTokenValidationResult struct {
	IsValid      bool   `json:"is_valid"`
	TokenID      string `json:"token_id,omitempty"`
	ResourceType string `json:"resource_type,omitempty"`
	ResourceID   string `json:"resource_id,omitempty"`
	Error        string `json:"error,omitempty"`
}

// EmbedTokenFilter represents filter criteria for querying embed tokens
type EmbedTokenFilter struct {
	ResourceType   *string `json:"resource_type,omitempty"`
	ResourceID     *string `json:"resource_id,omitempty"`
	CreatedBy      *string `json:"created_by,omitempty"`
	IncludeExpired bool    `json:"include_expired,omitempty"`
	IncludeRevoked bool    `json:"include_revoked,omitempty"`
	Limit          int     `json:"limit,omitempty"`
	Offset         int     `json:"offset,omitempty"`
}

// EmbedTokenWithStats represents an embed token with usage statistics
type EmbedTokenWithStats struct {
	EmbedToken
	ResourceName string `json:"resource_name,omitempty"`
}
