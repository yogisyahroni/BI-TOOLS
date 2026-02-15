package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"insight-engine-backend/models"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ShareService handles resource sharing operations
type ShareService struct {
	db           *gorm.DB
	auditService *AuditService
}

// NewShareService creates a new share service
func NewShareService(db *gorm.DB, auditService *AuditService) *ShareService {
	return &ShareService{
		db:           db,
		auditService: auditService,
	}
}

// CreateShare creates a new share for a resource
func (s *ShareService) CreateShare(req *models.ShareCreateRequest, sharedByUserID string) (*models.Share, error) {
	// Validate resource type
	resourceType, valid := models.ValidateResourceType(string(req.ResourceType))
	if !valid {
		return nil, errors.New("invalid resource type")
	}

	// Validate permission
	permission, valid := models.ValidateSharePermission(string(req.Permission))
	if !valid {
		return nil, errors.New("invalid permission level")
	}

	// Validate that either shared_with or shared_email is provided
	if req.SharedWith == nil && req.SharedEmail == nil {
		return nil, errors.New("either shared_with (user ID) or shared_email must be provided")
	}

	// Check if resource exists (based on type)
	if err := s.validateResourceExists(resourceType, req.ResourceID); err != nil {
		return nil, err
	}

	// Check if user has permission to share this resource
	hasPermission, err := s.checkUserCanShare(sharedByUserID, resourceType, req.ResourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, errors.New("you do not have permission to share this resource")
	}

	// Check for duplicate share (same resource, same recipient)
	var existingShare models.Share
	query := s.db.Where("resource_type = ? AND resource_id = ? AND shared_by = ? AND status = ?",
		resourceType, req.ResourceID, sharedByUserID, models.ShareStatusActive)

	if req.SharedWith != nil {
		query = query.Where("shared_with = ?", *req.SharedWith)
	} else if req.SharedEmail != nil {
		query = query.Where("shared_email = ?", *req.SharedEmail)
	}

	if err := query.First(&existingShare).Error; err == nil {
		return nil, errors.New("a share already exists for this recipient")
	}

	// Hash password if provided
	var passwordHash *string
	if req.Password != nil && *req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		hashStr := string(hash)
		passwordHash = &hashStr
	}

	// Determine status
	status := models.ShareStatusActive
	if req.SharedEmail != nil {
		status = models.ShareStatusPending // Email invites start as pending
	}

	share := models.Share{
		ResourceType: resourceType,
		ResourceID:   req.ResourceID,
		SharedBy:     sharedByUserID,
		SharedWith:   req.SharedWith,
		SharedEmail:  req.SharedEmail,
		Permission:   permission,
		PasswordHash: passwordHash,
		ExpiresAt:    req.ExpiresAt,
		Status:       status,
		Message:      req.Message,
	}

	if err := s.db.Create(&share).Error; err != nil {
		LogError("share_create_error", "Failed to create share", map[string]interface{}{
			"resource_type": resourceType,
			"resource_id":   req.ResourceID,
			"shared_by":     sharedByUserID,
			"error":         err.Error(),
		})
		return nil, fmt.Errorf("failed to create share: %w", err)
	}

	// Load relationships
	s.db.Preload("SharedByUser").Preload("SharedWithUser").First(&share, "id = ?", share.ID)

	LogInfo("share_created", "Share created successfully", map[string]interface{}{
		"share_id":      share.ID,
		"resource_type": resourceType,
		"resource_id":   req.ResourceID,
		"shared_by":     sharedByUserID,
		"permission":    permission,
	})

	return &share, nil
}

// GetSharesForResource retrieves all shares for a specific resource
func (s *ShareService) GetSharesForResource(resourceType models.ResourceType, resourceID string, userID string) ([]models.Share, error) {
	// Check if user has permission to view shares
	hasPermission, err := s.checkUserCanShare(userID, resourceType, resourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, errors.New("you do not have permission to view shares for this resource")
	}

	var shares []models.Share
	if err := s.db.Where("resource_type = ? AND resource_id = ?", resourceType, resourceID).
		Preload("SharedByUser").
		Preload("SharedWithUser").
		Order("created_at DESC").
		Find(&shares).Error; err != nil {
		LogError("shares_fetch_error", "Failed to fetch shares", map[string]interface{}{
			"resource_type": resourceType,
			"resource_id":   resourceID,
			"error":         err.Error(),
		})
		return nil, err
	}

	return shares, nil
}

// GetMyShares retrieves shares for the current user (both shared by and shared with)
func (s *ShareService) GetMyShares(userID string, filter *models.ShareFilter) ([]models.Share, error) {
	var shares []models.Share
	query := s.db.Model(&models.Share{})

	// Get shares where user is the recipient or the owner
	query = query.Where("(shared_with = ? OR shared_by = ?)", userID, userID)

	// Apply additional filters
	if filter != nil {
		if filter.ResourceType != nil {
			query = query.Where("resource_type = ?", *filter.ResourceType)
		}
		if filter.Status != nil {
			query = query.Where("status = ?", *filter.Status)
		}
		if !filter.IncludeExpired {
			query = query.Where("(expires_at IS NULL OR expires_at > ?)", time.Now())
		}
	} else {
		// Default: exclude expired shares
		query = query.Where("(expires_at IS NULL OR expires_at > ?)", time.Now())
	}

	if err := query.Preload("SharedByUser").
		Preload("SharedWithUser").
		Order("created_at DESC").
		Find(&shares).Error; err != nil {
		LogError("my_shares_fetch_error", "Failed to fetch user shares", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, err
	}

	return shares, nil
}

// GetShareByID retrieves a share by ID
func (s *ShareService) GetShareByID(shareID string) (*models.Share, error) {
	var share models.Share
	if err := s.db.Preload("SharedByUser").Preload("SharedWithUser").First(&share, "id = ?", shareID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("share not found")
		}
		return nil, err
	}

	// Check if expired and update status
	if share.IsExpired() && share.Status == models.ShareStatusActive {
		share.Status = models.ShareStatusExpired
		s.db.Save(&share)
	}

	return &share, nil
}

// UpdateShare updates a share
func (s *ShareService) UpdateShare(shareID string, req *models.ShareUpdateRequest, userID string) (*models.Share, error) {
	share, err := s.GetShareByID(shareID)
	if err != nil {
		return nil, err
	}

	// Only the owner or an admin can update the share
	if share.SharedBy != userID {
		isAdmin, err := s.isUserAdmin(userID)
		if err != nil || !isAdmin {
			return nil, errors.New("you do not have permission to update this share")
		}
	}

	// Check if share is already revoked
	if share.Status == models.ShareStatusRevoked {
		return nil, errors.New("cannot update a revoked share")
	}

	// Update permission if provided
	if req.Permission != nil {
		permission, valid := models.ValidateSharePermission(string(*req.Permission))
		if !valid {
			return nil, errors.New("invalid permission level")
		}
		share.Permission = permission
	}

	// Update password if provided
	if req.Password != nil {
		if *req.Password == "" {
			share.PasswordHash = nil
		} else {
			hash, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
			if err != nil {
				return nil, fmt.Errorf("failed to hash password: %w", err)
			}
			hashStr := string(hash)
			share.PasswordHash = &hashStr
		}
	}

	// Update expiration if provided
	if req.ExpiresAt != nil {
		share.ExpiresAt = req.ExpiresAt
	}

	// Update message if provided
	if req.Message != nil {
		share.Message = req.Message
	}

	if err := s.db.Save(share).Error; err != nil {
		LogError("share_update_error", "Failed to update share", map[string]interface{}{
			"share_id": shareID,
			"error":    err.Error(),
		})
		return nil, err
	}

	LogInfo("share_updated", "Share updated successfully", map[string]interface{}{
		"share_id":   shareID,
		"updated_by": userID,
	})

	return share, nil
}

// RevokeShare revokes a share
func (s *ShareService) RevokeShare(shareID string, userID string) error {
	share, err := s.GetShareByID(shareID)
	if err != nil {
		return err
	}

	// Only the owner or an admin can revoke the share
	if share.SharedBy != userID {
		isAdmin, err := s.isUserAdmin(userID)
		if err != nil || !isAdmin {
			return errors.New("you do not have permission to revoke this share")
		}
	}

	// Check if already revoked
	if share.Status == models.ShareStatusRevoked {
		return errors.New("share is already revoked")
	}

	share.Status = models.ShareStatusRevoked
	if err := s.db.Save(share).Error; err != nil {
		LogError("share_revoke_error", "Failed to revoke share", map[string]interface{}{
			"share_id": shareID,
			"error":    err.Error(),
		})
		return err
	}

	LogInfo("share_revoked", "Share revoked successfully", map[string]interface{}{
		"share_id":    shareID,
		"revoked_by":  userID,
		"resource_id": share.ResourceID,
	})

	return nil
}

// AcceptShare accepts a pending share invitation
func (s *ShareService) AcceptShare(shareID string, userID string, password *string) error {
	share, err := s.GetShareByID(shareID)
	if err != nil {
		return err
	}

	// Check if share is pending
	if share.Status != models.ShareStatusPending {
		return errors.New("share is not pending acceptance")
	}

	// Check if user is the intended recipient (if shared_with is set)
	if share.SharedWith != nil && *share.SharedWith != userID {
		return errors.New("you are not the intended recipient of this share")
	}

	// Validate password if required
	if share.RequiresPassword() {
		if password == nil || *password == "" {
			return errors.New("password is required to accept this share")
		}

		if err := bcrypt.CompareHashAndPassword([]byte(*share.PasswordHash), []byte(*password)); err != nil {
			return errors.New("invalid password")
		}
	}

	// Update share status
	share.Status = models.ShareStatusActive
	share.SharedWith = &userID
	share.SharedEmail = nil // Clear email once accepted
	now := time.Now()
	share.AcceptedAt = &now

	if err := s.db.Save(share).Error; err != nil {
		LogError("share_accept_error", "Failed to accept share", map[string]interface{}{
			"share_id": shareID,
			"user_id":  userID,
			"error":    err.Error(),
		})
		return err
	}

	LogInfo("share_accepted", "Share accepted successfully", map[string]interface{}{
		"share_id": shareID,
		"user_id":  userID,
	})

	return nil
}

// CheckAccess checks if a user has access to a resource via sharing
func (s *ShareService) CheckAccess(userID string, resourceType models.ResourceType, resourceID string) (*models.ShareAccessCheck, error) {
	// Look for active shares
	var share models.Share
	err := s.db.Where("resource_type = ? AND resource_id = ? AND shared_with = ? AND status = ?",
		resourceType, resourceID, userID, models.ShareStatusActive).
		Where("(expires_at IS NULL OR expires_at > ?)", time.Now()).
		First(&share).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &models.ShareAccessCheck{
				HasAccess: false,
			}, nil
		}
		return nil, err
	}

	return &models.ShareAccessCheck{
		HasAccess:        true,
		Permission:       share.Permission,
		ShareID:          share.ID,
		RequiresPassword: false, // Already authenticated via session
	}, nil
}

// ValidateShareAccess validates access to a share with optional password
func (s *ShareService) ValidateShareAccess(shareID string, password *string) (*models.Share, error) {
	share, err := s.GetShareByID(shareID)
	if err != nil {
		return nil, err
	}

	// Check if share is active
	if !share.IsActive() {
		return nil, errors.New("share is not active")
	}

	// Validate password if required
	if share.RequiresPassword() {
		if password == nil || *password == "" {
			return nil, errors.New("password is required")
		}

		if err := bcrypt.CompareHashAndPassword([]byte(*share.PasswordHash), []byte(*password)); err != nil {
			return nil, errors.New("invalid password")
		}
	}

	return share, nil
}

// CleanupExpiredShares marks expired shares as expired
func (s *ShareService) CleanupExpiredShares() (int64, error) {
	result := s.db.Model(&models.Share{}).
		Where("status = ? AND expires_at IS NOT NULL AND expires_at < ?",
			models.ShareStatusActive, time.Now()).
		Update("status", models.ShareStatusExpired)

	if result.Error != nil {
		LogError("share_cleanup_error", "Failed to cleanup expired shares", map[string]interface{}{
			"error": result.Error.Error(),
		})
		return 0, result.Error
	}

	if result.RowsAffected > 0 {
		LogInfo("shares_cleaned_up", "Expired shares cleaned up", map[string]interface{}{
			"count": result.RowsAffected,
		})
	}

	return result.RowsAffected, nil
}

// Helper methods

func (s *ShareService) validateResourceExists(resourceType models.ResourceType, resourceID string) error {
	switch resourceType {
	case models.ResourceTypeDashboard:
		var count int64
		if err := s.db.Model(&models.Dashboard{}).Where("id = ?", resourceID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return errors.New("dashboard not found")
		}
	case models.ResourceTypeQuery:
		var count int64
		if err := s.db.Model(&models.SavedQuery{}).Where("id = ?", resourceID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return errors.New("query not found")
		}
	default:
		return errors.New("unknown resource type")
	}
	return nil
}

func (s *ShareService) checkUserCanShare(userID string, resourceType models.ResourceType, resourceID string) (bool, error) {
	// Check if user owns the resource
	switch resourceType {
	case models.ResourceTypeDashboard:
		var dashboard models.Dashboard
		if err := s.db.Where("id = ?", resourceID).First(&dashboard).Error; err != nil {
			return false, err
		}
		if dashboard.UserID == userID {
			return true, nil
		}
	case models.ResourceTypeQuery:
		var query models.SavedQuery
		if err := s.db.Where("id = ?", resourceID).First(&query).Error; err != nil {
			return false, err
		}
		if query.UserID == userID {
			return true, nil
		}
	}

	// Check if user has admin/share permission
	var user models.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return false, err
	}

	return user.IsSuperAdmin(), nil
}

func (s *ShareService) isUserAdmin(userID string) (bool, error) {
	var user models.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return false, err
	}
	return user.IsSuperAdmin(), nil
}

// GenerateShareToken generates a unique token for share links
func GenerateShareToken() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
