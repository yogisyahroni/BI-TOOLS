package services

import (
	"context"
	"fmt"
	"time"

	"insight-engine-backend/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OrganizationService handles organization/workspace management
type OrganizationService struct {
	db           *gorm.DB
	auditService *AuditService
}

// NewOrganizationService creates a new organization service
func NewOrganizationService(db *gorm.DB, auditService *AuditService) *OrganizationService {
	return &OrganizationService{
		db:           db,
		auditService: auditService,
	}
}

// Organization represents a workspace/organization with admin metadata
type Organization struct {
	models.Workspace
	MemberCount int64             `json:"memberCount"`
	Quota       OrganizationQuota `json:"quota"`
	Usage       OrganizationUsage `json:"usage"`
	OwnerEmail  string            `json:"ownerEmail,omitempty"`
}

// OrganizationQuota represents quota limits for an organization
type OrganizationQuota struct {
	MaxUsers      int   `json:"maxUsers"`
	MaxQueries    int   `json:"maxQueries"`
	MaxDashboards int   `json:"maxDashboards"`
	MaxStorageMB  int64 `json:"maxStorageMB"`
	MaxWorkspaces int   `json:"maxWorkspaces"`
}

// OrganizationUsage represents current resource usage
type OrganizationUsage struct {
	Users      int   `json:"users"`
	Queries    int   `json:"queries"`
	Dashboards int   `json:"dashboards"`
	StorageMB  int64 `json:"storageMB"`
	Workspaces int   `json:"workspaces"`
}

// OrganizationMember represents a member with user details
type OrganizationMember struct {
	ID          string       `json:"id"`
	UserID      string       `json:"userId"`
	WorkspaceID string       `json:"workspaceId"`
	Role        string       `json:"role"`
	InvitedAt   time.Time    `json:"invitedAt"`
	JoinedAt    *time.Time   `json:"joinedAt,omitempty"`
	User        *models.User `json:"user,omitempty"`
}

// GetOrganizationsFilter represents filter for organizations
type GetOrganizationsFilter struct {
	Search string
	Limit  int
	Offset int
}

// GetOrganizations retrieves a list of organizations with stats
func (s *OrganizationService) GetOrganizations(ctx context.Context, filter *GetOrganizationsFilter) ([]Organization, int64, error) {
	query := s.db.Model(&models.Workspace{})

	// Apply search filter
	if filter.Search != "" {
		search := "%" + filter.Search + "%"
		query = query.Where("name ILIKE ? OR description ILIKE ?", search, search)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count organizations: %w", err)
	}

	// Apply pagination
	limit := filter.Limit
	if limit == 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// Retrieve workspaces
	var workspaces []models.Workspace
	if err := query.Order("created_at DESC").Limit(limit).Offset(filter.Offset).Find(&workspaces).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve organizations: %w", err)
	}

	// Build organization list with stats
	var orgs []Organization
	for _, ws := range workspaces {
		org := Organization{
			Workspace: ws,
			Quota: OrganizationQuota{
				MaxUsers:      100, // Default quotas
				MaxQueries:    10000,
				MaxDashboards: 100,
				MaxStorageMB:  1024,
				MaxWorkspaces: 10,
			},
		}

		// Get member count
		var memberCount int64
		s.db.Model(&models.WorkspaceMember{}).Where("workspace_id = ?", ws.ID).Count(&memberCount)
		org.MemberCount = memberCount
		org.Usage.Users = int(memberCount)

		// Get owner email
		var owner models.User
		if err := s.db.Select("email").Where("id = ?", ws.OwnerID).First(&owner).Error; err == nil {
			org.OwnerEmail = owner.Email
		}

		orgs = append(orgs, org)
	}

	return orgs, total, nil
}

// GetOrganizationByID retrieves a single organization with details
func (s *OrganizationService) GetOrganizationByID(ctx context.Context, orgID string) (*Organization, error) {
	var ws models.Workspace
	if err := s.db.First(&ws, "id = ?", orgID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("organization not found")
		}
		return nil, fmt.Errorf("failed to retrieve organization: %w", err)
	}

	org := &Organization{
		Workspace: ws,
		Quota: OrganizationQuota{
			MaxUsers:      100,
			MaxQueries:    10000,
			MaxDashboards: 100,
			MaxStorageMB:  1024,
			MaxWorkspaces: 10,
		},
	}

	// Get member count
	var memberCount int64
	s.db.Model(&models.WorkspaceMember{}).Where("workspace_id = ?", orgID).Count(&memberCount)
	org.MemberCount = memberCount
	org.Usage.Users = int(memberCount)

	// Get owner email
	var owner models.User
	if err := s.db.Select("email").Where("id = ?", ws.OwnerID).First(&owner).Error; err == nil {
		org.OwnerEmail = owner.Email
	}

	return org, nil
}

// CreateOrganizationRequest represents create organization request
type CreateOrganizationRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	OwnerID     string `json:"ownerId"`
}

// CreateOrganization creates a new organization
func (s *OrganizationService) CreateOrganization(ctx context.Context, req *CreateOrganizationRequest) (*Organization, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("organization name is required")
	}

	if req.OwnerID == "" {
		return nil, fmt.Errorf("owner ID is required")
	}

	// Verify owner exists
	var owner models.User
	if err := s.db.First(&owner, "id = ?", req.OwnerID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("owner not found")
		}
		return nil, fmt.Errorf("failed to verify owner: %w", err)
	}

	// Create workspace
	ws := &models.Workspace{
		ID:        uuid.New().String(),
		Name:      req.Name,
		OwnerID:   req.OwnerID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if req.Description != "" {
		ws.Description = &req.Description
	}

	if err := s.db.Create(ws).Error; err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	// Add owner as member
	member := &models.WorkspaceMember{
		ID:          uuid.New().String(),
		WorkspaceID: ws.ID,
		UserID:      req.OwnerID,
		Role:        models.RoleOwner,
		InvitedAt:   time.Now(),
		JoinedAt:    &[]time.Time{time.Now()}[0],
	}

	if err := s.db.Create(member).Error; err != nil {
		// Rollback workspace creation
		s.db.Delete(ws)
		return nil, fmt.Errorf("failed to add owner as member: %w", err)
	}

	org := &Organization{
		Workspace:   *ws,
		MemberCount: 1,
		Quota: OrganizationQuota{
			MaxUsers:      100,
			MaxQueries:    10000,
			MaxDashboards: 100,
			MaxStorageMB:  1024,
			MaxWorkspaces: 10,
		},
		Usage: OrganizationUsage{
			Users: 1,
		},
		OwnerEmail: owner.Email,
	}

	return org, nil
}

// UpdateOrganizationRequest represents update organization request
type UpdateOrganizationRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpdateOrganization updates organization details
func (s *OrganizationService) UpdateOrganization(ctx context.Context, orgID string, req *UpdateOrganizationRequest) (*Organization, error) {
	var ws models.Workspace
	if err := s.db.First(&ws, "id = ?", orgID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("organization not found")
		}
		return nil, fmt.Errorf("failed to retrieve organization: %w", err)
	}

	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if req.Name != "" {
		updates["name"] = req.Name
	}

	if req.Description != "" {
		updates["description"] = req.Description
	}

	if len(updates) > 1 { // More than just updated_at
		if err := s.db.Model(&ws).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update organization: %w", err)
		}
	}

	return s.GetOrganizationByID(ctx, orgID)
}

// DeleteOrganization deletes an organization and all its data
func (s *OrganizationService) DeleteOrganization(ctx context.Context, orgID string) error {
	var ws models.Workspace
	if err := s.db.First(&ws, "id = ?", orgID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("organization not found")
		}
		return fmt.Errorf("failed to retrieve organization: %w", err)
	}

	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Delete all members
	if err := tx.Where("workspace_id = ?", orgID).Delete(&models.WorkspaceMember{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete members: %w", err)
	}

	// Delete workspace
	if err := tx.Delete(&ws).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete organization: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetOrganizationMembers retrieves all members of an organization
func (s *OrganizationService) GetOrganizationMembers(ctx context.Context, orgID string) ([]OrganizationMember, error) {
	var members []models.WorkspaceMember
	if err := s.db.Where("workspace_id = ?", orgID).Find(&members).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve members: %w", err)
	}

	var result []OrganizationMember
	for _, m := range members {
		member := OrganizationMember{
			ID:          m.ID,
			UserID:      m.UserID,
			WorkspaceID: m.WorkspaceID,
			Role:        m.Role,
			InvitedAt:   m.InvitedAt,
			JoinedAt:    m.JoinedAt,
		}

		// Get user details
		var user models.User
		if err := s.db.Select("id, email, name, username").Where("id = ?", m.UserID).First(&user).Error; err == nil {
			// Clear sensitive data
			user.Password = ""
			user.EmailVerificationToken = ""
			user.PasswordResetToken = ""
			member.User = &user
		}

		result = append(result, member)
	}

	return result, nil
}

// AddOrganizationMemberRequest represents add member request
type AddOrganizationMemberRequest struct {
	UserID string `json:"userId"`
	Role   string `json:"role"`
}

// AddOrganizationMember adds a member to an organization
func (s *OrganizationService) AddOrganizationMember(ctx context.Context, orgID string, req *AddOrganizationMemberRequest) (*OrganizationMember, error) {
	// Verify organization exists
	var ws models.Workspace
	if err := s.db.First(&ws, "id = ?", orgID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("organization not found")
		}
		return nil, fmt.Errorf("failed to retrieve organization: %w", err)
	}

	// Verify user exists
	var user models.User
	if err := s.db.First(&user, "id = ?", req.UserID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to verify user: %w", err)
	}

	// Check if user is already a member
	var existing models.WorkspaceMember
	if err := s.db.Where("workspace_id = ? AND user_id = ?", orgID, req.UserID).First(&existing).Error; err == nil {
		return nil, fmt.Errorf("user is already a member")
	}

	// Validate role
	validRoles := map[string]bool{
		models.RoleOwner:  true,
		models.RoleAdmin:  true,
		models.RoleEditor: true,
		models.RoleViewer: true,
	}
	if !validRoles[req.Role] {
		req.Role = models.RoleViewer // Default to viewer
	}

	// Create member
	member := &models.WorkspaceMember{
		ID:          uuid.New().String(),
		WorkspaceID: orgID,
		UserID:      req.UserID,
		Role:        req.Role,
		InvitedAt:   time.Now(),
	}

	if err := s.db.Create(member).Error; err != nil {
		return nil, fmt.Errorf("failed to add member: %w", err)
	}

	return &OrganizationMember{
		ID:          member.ID,
		UserID:      member.UserID,
		WorkspaceID: member.WorkspaceID,
		Role:        member.Role,
		InvitedAt:   member.InvitedAt,
		JoinedAt:    member.JoinedAt,
		User: &models.User{
			ID:       user.ID,
			Email:    user.Email,
			Name:     user.Name,
			Username: user.Username,
		},
	}, nil
}

// RemoveOrganizationMember removes a member from an organization
func (s *OrganizationService) RemoveOrganizationMember(ctx context.Context, orgID, userID string) error {
	// Verify organization exists
	var ws models.Workspace
	if err := s.db.First(&ws, "id = ?", orgID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("organization not found")
		}
		return fmt.Errorf("failed to retrieve organization: %w", err)
	}

	// Cannot remove owner
	if ws.OwnerID == userID {
		return fmt.Errorf("cannot remove organization owner")
	}

	// Delete member
	result := s.db.Where("workspace_id = ? AND user_id = ?", orgID, userID).Delete(&models.WorkspaceMember{})
	if result.Error != nil {
		return fmt.Errorf("failed to remove member: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("member not found")
	}

	return nil
}

// UpdateMemberRoleRequest represents update role request
type UpdateMemberRoleRequest struct {
	Role string `json:"role"`
}

// UpdateMemberRole updates a member's role
func (s *OrganizationService) UpdateMemberRole(ctx context.Context, orgID, userID string, req *UpdateMemberRoleRequest) (*OrganizationMember, error) {
	// Validate role
	validRoles := map[string]bool{
		models.RoleOwner:  true,
		models.RoleAdmin:  true,
		models.RoleEditor: true,
		models.RoleViewer: true,
	}
	if !validRoles[req.Role] {
		return nil, fmt.Errorf("invalid role")
	}

	// Verify organization exists
	var ws models.Workspace
	if err := s.db.First(&ws, "id = ?", orgID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("organization not found")
		}
		return nil, fmt.Errorf("failed to retrieve organization: %w", err)
	}

	// Update member role
	result := s.db.Model(&models.WorkspaceMember{}).
		Where("workspace_id = ? AND user_id = ?", orgID, userID).
		Update("role", req.Role)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to update member role: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("member not found")
	}

	// Reload member
	var member models.WorkspaceMember
	if err := s.db.Where("workspace_id = ? AND user_id = ?", orgID, userID).First(&member).Error; err != nil {
		return nil, fmt.Errorf("failed to reload member: %w", err)
	}

	return &OrganizationMember{
		ID:          member.ID,
		UserID:      member.UserID,
		WorkspaceID: member.WorkspaceID,
		Role:        member.Role,
		InvitedAt:   member.InvitedAt,
		JoinedAt:    member.JoinedAt,
	}, nil
}

// OrganizationStats represents organization statistics
type OrganizationStats struct {
	TotalOrganizations int64            `json:"totalOrganizations"`
	TotalMembers       int64            `json:"totalMembers"`
	OrganizationsByDay map[string]int64 `json:"organizationsByDay"`
}

// GetOrganizationStats retrieves organization statistics
func (s *OrganizationService) GetOrganizationStats(ctx context.Context, orgID string) (*OrganizationStats, error) {
	stats := &OrganizationStats{
		OrganizationsByDay: make(map[string]int64),
	}

	// Total organizations
	if err := s.db.Model(&models.Workspace{}).Count(&stats.TotalOrganizations).Error; err != nil {
		return nil, fmt.Errorf("failed to count organizations: %w", err)
	}

	// Total members
	if err := s.db.Model(&models.WorkspaceMember{}).Count(&stats.TotalMembers).Error; err != nil {
		return nil, fmt.Errorf("failed to count members: %w", err)
	}

	return stats, nil
}

// GetOrganizationQuota retrieves an organization's quota
func (s *OrganizationService) GetOrganizationQuota(ctx context.Context, orgID string) (*OrganizationQuota, error) {
	// Verify organization exists
	var ws models.Workspace
	if err := s.db.First(&ws, "id = ?", orgID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("organization not found")
		}
		return nil, fmt.Errorf("failed to retrieve organization: %w", err)
	}

	// In a real implementation, quotas would be stored in a separate table
	// For now, return default quotas
	quota := &OrganizationQuota{
		MaxUsers:      100,
		MaxQueries:    10000,
		MaxDashboards: 100,
		MaxStorageMB:  1024,
		MaxWorkspaces: 10,
	}

	return quota, nil
}

// UpdateOrganizationQuotaRequest represents update quota request
type UpdateOrganizationQuotaRequest struct {
	MaxUsers      int   `json:"maxUsers"`
	MaxQueries    int   `json:"maxQueries"`
	MaxDashboards int   `json:"maxDashboards"`
	MaxStorageMB  int64 `json:"maxStorageMB"`
	MaxWorkspaces int   `json:"maxWorkspaces"`
}

// UpdateOrganizationQuota updates an organization's quota
func (s *OrganizationService) UpdateOrganizationQuota(ctx context.Context, orgID string, req *UpdateOrganizationQuotaRequest) (*OrganizationQuota, error) {
	// Verify organization exists
	var ws models.Workspace
	if err := s.db.First(&ws, "id = ?", orgID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("organization not found")
		}
		return nil, fmt.Errorf("failed to retrieve organization: %w", err)
	}

	// In a real implementation, save quotas to database
	// For now, return the updated quotas
	quota := &OrganizationQuota{
		MaxUsers:      req.MaxUsers,
		MaxQueries:    req.MaxQueries,
		MaxDashboards: req.MaxDashboards,
		MaxStorageMB:  req.MaxStorageMB,
		MaxWorkspaces: req.MaxWorkspaces,
	}

	return quota, nil
}
