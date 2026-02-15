package services

import (
	"context"
	"fmt"
	"time"

	"insight-engine-backend/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserAdminService handles user management operations for admins
type UserAdminService struct {
	db           *gorm.DB
	auditService *AuditService
}

// NewUserAdminService creates a new user admin service
func NewUserAdminService(db *gorm.DB, auditService *AuditService) *UserAdminService {
	return &UserAdminService{
		db:           db,
		auditService: auditService,
	}
}

// UserListResponse represents the response for user list
type UserListResponse struct {
	Users []models.User `json:"users"`
	Total int64         `json:"total"`
}

// UserFilter represents filter criteria for user list
type UserFilter struct {
	Search    string
	Role      string
	Status    string
	Provider  string
	SortBy    string
	SortOrder string
	Limit     int
	Offset    int
}

// GetUsers retrieves a paginated list of users with filters
func (s *UserAdminService) GetUsers(ctx context.Context, filter *UserFilter) (*UserListResponse, error) {
	query := s.db.Model(&models.User{})

	// Apply search filter
	if filter.Search != "" {
		search := "%" + filter.Search + "%"
		query = query.Where("LOWER(name) LIKE LOWER(?) OR LOWER(email) LIKE LOWER(?) OR LOWER(username) LIKE LOWER(?)", search, search, search)
	}

	// Apply role filter
	if filter.Role != "" {
		query = query.Where("role = ?", filter.Role)
	}

	// Apply status filter
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	// Apply provider filter
	if filter.Provider != "" {
		query = query.Where("provider = ?", filter.Provider)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	// Apply sorting
	sortBy := "created_at"
	if filter.SortBy != "" {
		sortBy = filter.SortBy
	}
	sortOrder := "DESC"
	if filter.SortOrder == "asc" {
		sortOrder = "ASC"
	}
	query = query.Order(sortBy + " " + sortOrder)

	// Apply pagination
	limit := filter.Limit
	if limit == 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	query = query.Limit(limit).Offset(filter.Offset)

	// Retrieve users
	var users []models.User
	if err := query.Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve users: %w", err)
	}

	// Clear sensitive data
	for i := range users {
		users[i].Password = ""
		users[i].EmailVerificationToken = ""
		users[i].PasswordResetToken = ""
		users[i].ImpersonationToken = ""
	}

	return &UserListResponse{
		Users: users,
		Total: total,
	}, nil
}

// GetUserByID retrieves a single user by ID
func (s *UserAdminService) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to retrieve user: %w", err)
	}

	// Load user roles
	if err := s.db.Model(&user).Association("Roles").Find(&user.Roles); err != nil {
		// Log error but don't fail
		LogWarn("user_roles_load_failed", "Failed to load user roles", map[string]interface{}{"user_id": userID, "error": err})
	}

	// Clear sensitive data
	user.Password = ""
	user.EmailVerificationToken = ""
	user.PasswordResetToken = ""
	user.ImpersonationToken = ""

	return &user, nil
}

// UpdateUserRequest represents update user request
type UpdateUserRequest struct {
	Name  string   `json:"name"`
	Email string   `json:"email"`
	Role  string   `json:"role"`
	Roles []string `json:"roles"` // Role IDs for RBAC
}

// UpdateUser updates user information
func (s *UserAdminService) UpdateUser(ctx context.Context, userID string, req *UpdateUserRequest) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to retrieve user: %w", err)
	}

	// Store old values for audit (can be used for audit logging)
	_ = map[string]interface{}{
		"name":  user.Name,
		"email": user.Email,
		"role":  user.Role,
	}

	updates := map[string]interface{}{}

	if req.Name != "" {
		updates["name"] = req.Name
	}

	if req.Email != "" && req.Email != user.Email {
		// Check email uniqueness
		var existing models.User
		if err := s.db.Where("email = ? AND id != ?", req.Email, userID).First(&existing).Error; err == nil {
			return nil, fmt.Errorf("email already in use")
		}
		updates["email"] = req.Email
	}

	if req.Role != "" {
		updates["role"] = req.Role
	}

	// Apply updates
	if len(updates) > 0 {
		if err := s.db.Model(&user).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}
	}

	// Update RBAC roles if provided
	if len(req.Roles) > 0 {
		// Clear existing roles and add new ones
		// This is a simplified version - in production, use proper role management
		LogInfo("user_roles_update", "Roles would be updated", map[string]interface{}{"user_id": userID, "roles": req.Roles})
	}

	// Reload user
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload user: %w", err)
	}

	// Clear sensitive data
	user.Password = ""
	user.EmailVerificationToken = ""
	user.PasswordResetToken = ""
	user.ImpersonationToken = ""

	return &user, nil
}

// ActivateUser activates a deactivated user account
func (s *UserAdminService) ActivateUser(ctx context.Context, userID string, adminID string) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to retrieve user: %w", err)
	}

	if user.Status == models.UserStatusActive {
		return nil, fmt.Errorf("user is already active")
	}

	updates := map[string]interface{}{
		"status":              models.UserStatusActive,
		"deactivated_at":      nil,
		"deactivated_by":      nil,
		"deactivation_reason": nil,
	}

	if err := s.db.Model(&user).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to activate user: %w", err)
	}

	// Reload user
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload user: %w", err)
	}

	// Clear sensitive data
	user.Password = ""
	user.EmailVerificationToken = ""
	user.PasswordResetToken = ""
	user.ImpersonationToken = ""

	return &user, nil
}

// DeactivateUserRequest represents deactivate user request
type DeactivateUserRequest struct {
	Reason string `json:"reason"`
}

// DeactivateUser deactivates a user account (soft delete)
func (s *UserAdminService) DeactivateUser(ctx context.Context, userID string, adminID string, reason string) (*models.User, error) {
	// Prevent self-deactivation
	if userID == adminID {
		return nil, fmt.Errorf("cannot deactivate your own account")
	}

	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to retrieve user: %w", err)
	}

	if user.Status == models.UserStatusInactive {
		return nil, fmt.Errorf("user is already inactive")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":              models.UserStatusInactive,
		"deactivated_at":      &now,
		"deactivated_by":      adminID,
		"deactivation_reason": reason,
	}

	if err := s.db.Model(&user).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to deactivate user: %w", err)
	}

	// Reload user
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload user: %w", err)
	}

	// Clear sensitive data
	user.Password = ""
	user.EmailVerificationToken = ""
	user.PasswordResetToken = ""
	user.ImpersonationToken = ""

	return &user, nil
}

// DeleteUser permanently deletes a user and their data
func (s *UserAdminService) DeleteUser(ctx context.Context, userID string, adminID string) error {
	// Prevent self-deletion
	if userID == adminID {
		return fmt.Errorf("cannot delete your own account")
	}

	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("user not found")
		}
		return fmt.Errorf("failed to retrieve user: %w", err)
	}

	// Start transaction for data cleanup
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Delete user's audit logs (or anonymize them)
	// In production, consider keeping anonymized audit logs for compliance
	if err := tx.Where("user_id = ?", userID).Delete(&models.AuditLog{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete user audit logs: %w", err)
	}

	// Delete user's roles associations
	if err := tx.Exec("DELETE FROM user_roles WHERE user_id = ?", userID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete user roles: %w", err)
	}

	// Delete user
	if err := tx.Delete(&user).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ImpersonationResponse represents impersonation response
type ImpersonationResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	UserID    string    `json:"user_id"`
}

// ImpersonateUser generates an impersonation token for a user
func (s *UserAdminService) ImpersonateUser(ctx context.Context, userID string, adminID string) (*ImpersonationResponse, error) {
	// Prevent self-impersonation
	if userID == adminID {
		return nil, fmt.Errorf("cannot impersonate yourself")
	}

	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to retrieve user: %w", err)
	}

	// Check if user can be impersonated
	if !user.IsActive() {
		return nil, fmt.Errorf("cannot impersonate inactive user")
	}

	// Generate impersonation token
	token := uuid.New().String()
	expiresAt := time.Now().Add(15 * time.Minute)

	updates := map[string]interface{}{
		"impersonation_token":   token,
		"impersonation_expires": &expiresAt,
		"impersonated_by":       adminID,
	}

	if err := s.db.Model(&user).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to create impersonation token: %w", err)
	}

	return &ImpersonationResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		UserID:    userID,
	}, nil
}

// ValidateImpersonationToken validates an impersonation token
func (s *UserAdminService) ValidateImpersonationToken(ctx context.Context, token string) (*models.User, error) {
	var user models.User
	if err := s.db.Where("impersonation_token = ?", token).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invalid impersonation token")
		}
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}

	// Check expiration
	if user.ImpersonationExpires == nil || time.Now().After(*user.ImpersonationExpires) {
		return nil, fmt.Errorf("impersonation token expired")
	}

	// Clear impersonation data after use
	s.db.Model(&user).Updates(map[string]interface{}{
		"impersonation_token":   "",
		"impersonation_expires": nil,
		"impersonated_by":       nil,
	})

	// Clear sensitive data
	user.Password = ""
	user.EmailVerificationToken = ""
	user.PasswordResetToken = ""
	user.ImpersonationToken = ""

	return &user, nil
}

// GetUserActivity retrieves a user's recent activity
func (s *UserAdminService) GetUserActivity(ctx context.Context, userID string, limit int) ([]models.AuditLog, error) {
	if limit == 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	// Convert userID to uint for audit log query
	// Note: This assumes user ID is stored as uint in audit logs
	// In production, adjust based on actual schema
	var logs []models.AuditLog
	err := s.db.Where("username = (SELECT username FROM users WHERE id = ?)", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve user activity: %w", err)
	}

	return logs, nil
}

// UserStats represents user statistics
type UserStats struct {
	TotalUsers       int64 `json:"total_users"`
	ActiveUsers      int64 `json:"active_users"`
	InactiveUsers    int64 `json:"inactive_users"`
	PendingUsers     int64 `json:"pending_users"`
	AdminUsers       int64 `json:"admin_users"`
	NewUsersToday    int64 `json:"new_users_today"`
	NewUsersThisWeek int64 `json:"new_users_this_week"`
}

// GetUserStats retrieves user statistics
func (s *UserAdminService) GetUserStats(ctx context.Context) (*UserStats, error) {
	stats := &UserStats{}

	// Total users
	if err := s.db.Model(&models.User{}).Count(&stats.TotalUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to count total users: %w", err)
	}

	// Active users
	if err := s.db.Model(&models.User{}).Where("status = ?", models.UserStatusActive).Count(&stats.ActiveUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to count active users: %w", err)
	}

	// Inactive users
	if err := s.db.Model(&models.User{}).Where("status = ?", models.UserStatusInactive).Count(&stats.InactiveUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to count inactive users: %w", err)
	}

	// Pending users
	if err := s.db.Model(&models.User{}).Where("status = ?", models.UserStatusPending).Count(&stats.PendingUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to count pending users: %w", err)
	}

	// Admin users
	if err := s.db.Model(&models.User{}).Where("role = ?", "admin").Count(&stats.AdminUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to count admin users: %w", err)
	}

	// New users today
	today := time.Now().Truncate(24 * time.Hour)
	if err := s.db.Model(&models.User{}).Where("created_at >= ?", today).Count(&stats.NewUsersToday).Error; err != nil {
		return nil, fmt.Errorf("failed to count new users today: %w", err)
	}

	// New users this week
	weekAgo := time.Now().AddDate(0, 0, -7)
	if err := s.db.Model(&models.User{}).Where("created_at >= ?", weekAgo).Count(&stats.NewUsersThisWeek).Error; err != nil {
		return nil, fmt.Errorf("failed to count new users this week: %w", err)
	}

	return stats, nil
}

// BulkUpdateRequest represents bulk update request
type BulkUpdateRequest struct {
	UserIDs []string `json:"user_ids"`
	Action  string   `json:"action"` // "activate", "deactivate", "delete"
	Reason  string   `json:"reason"`
}

// BulkUpdateUsers performs bulk operations on users
func (s *UserAdminService) BulkUpdateUsers(ctx context.Context, adminID string, req *BulkUpdateRequest) (int, error) {
	if len(req.UserIDs) == 0 {
		return 0, fmt.Errorf("no users specified")
	}

	if len(req.UserIDs) > 100 {
		return 0, fmt.Errorf("maximum 100 users can be processed at once")
	}

	processed := 0

	switch req.Action {
	case "activate":
		for _, userID := range req.UserIDs {
			if userID == adminID {
				continue // Skip self
			}
			_, err := s.ActivateUser(ctx, userID, adminID)
			if err == nil {
				processed++
			}
		}
	case "deactivate":
		for _, userID := range req.UserIDs {
			if userID == adminID {
				continue // Skip self
			}
			_, err := s.DeactivateUser(ctx, userID, adminID, req.Reason)
			if err == nil {
				processed++
			}
		}
	case "delete":
		for _, userID := range req.UserIDs {
			if userID == adminID {
				continue // Skip self
			}
			err := s.DeleteUser(ctx, userID, adminID)
			if err == nil {
				processed++
			}
		}
	default:
		return 0, fmt.Errorf("invalid action: %s", req.Action)
	}

	return processed, nil
}
