package handlers

import (
	"strconv"
	"time"

	"insight-engine-backend/models"
	"insight-engine-backend/pkg/validator"
	"insight-engine-backend/services"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// AdminUserHandler handles admin user management
type AdminUserHandler struct {
	db           *gorm.DB
	auditService *services.AuditService
}

// NewAdminUserHandler creates a new admin user handler
func NewAdminUserHandler(db *gorm.DB, auditService *services.AuditService) *AdminUserHandler {
	return &AdminUserHandler{
		db:           db,
		auditService: auditService,
	}
}

// RegisterRoutes registers all admin user routes
func (h *AdminUserHandler) RegisterRoutes(router fiber.Router, middlewares ...func(*fiber.Ctx) error) {
	users := router.Group("/admin/users")

	// Apply all provided middlewares (auth + admin check)

	for _, mw := range middlewares {

		users.Use(mw)

	}

	users.Get("/", h.ListUsers)
	users.Get("/stats", h.GetUserStats)
	users.Get("/:id", h.GetUser)
	users.Put("/:id/activate", h.ActivateUser)
	users.Put("/:id/deactivate", h.DeactivateUser)
	users.Put("/:id/role", h.UpdateUserRole)
	users.Post("/:id/impersonate", h.ImpersonateUser)
	users.Get("/:id/activity", h.GetUserActivity)
}

// UserListResponse represents user list item
type UserListResponse struct {
	ID            string     `json:"id"`
	Email         string     `json:"email"`
	Name          string     `json:"name"`
	Username      string     `json:"username"`
	Role          string     `json:"role"`
	Status        string     `json:"status"`
	EmailVerified bool       `json:"emailVerified"`
	Provider      string     `json:"provider,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	LastLoginAt   *time.Time `json:"lastLoginAt,omitempty"`
}

// ListUsers handles GET /admin/users
func (h *AdminUserHandler) ListUsers(c *fiber.Ctx) error {
	// Parse query parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(c.Query("pageSize", "20"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	search := c.Query("search")
	status := c.Query("status")
	role := c.Query("role")

	query := h.db.Model(&models.User{})

	// Apply filters
	if search != "" {
		query = query.Where("email ILIKE ? OR name ILIKE ? OR username ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if role != "" {
		query = query.Where("role = ?", role)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to count users",
		})
	}

	// Get paginated results
	var users []models.User
	offset := (page - 1) * pageSize
	if err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&users).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch users",
		})
	}

	// Convert to response format
	userList := make([]UserListResponse, len(users))
	for i, u := range users {
		userList[i] = UserListResponse{
			ID:            u.ID.String(),
			Email:         u.Email,
			Name:          u.Name,
			Username:      u.Username,
			Role:          u.Role,
			Status:        u.Status,
			EmailVerified: u.EmailVerified,
			Provider:      u.Provider,
			CreatedAt:     u.CreatedAt,
		}
	}

	// Calculate pagination metadata
	totalPages := (int(total) + pageSize - 1) / pageSize

	return c.JSON(fiber.Map{
		"data": userList,
		"pagination": fiber.Map{
			"page":       page,
			"pageSize":   pageSize,
			"total":      total,
			"totalPages": totalPages,
		},
	})
}

// GetUser handles GET /admin/users/:id
func (h *AdminUserHandler) GetUser(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User ID is required",
		})
	}

	var user models.User
	if err := h.db.Preload("Roles").Where("id = ?", id).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch user",
		})
	}

	// Clear sensitive fields
	user.Password = ""
	user.EmailVerificationToken = ""
	user.PasswordResetToken = ""
	user.ImpersonationToken = ""

	return c.JSON(user)
}

// ActivateUser handles PUT /admin/users/:id/activate
func (h *AdminUserHandler) ActivateUser(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User ID is required",
		})
	}

	// Get current admin user from context
	currentUser := c.Locals("user").(*models.User)
	if !currentUser.IsSuperAdmin() {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	var user models.User
	if err := h.db.Where("id = ?", id).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch user",
		})
	}

	// Update status
	if err := h.db.Model(&user).Updates(map[string]interface{}{
		"status":              models.UserStatusActive,
		"deactivated_at":      nil,
		"deactivated_by":      nil,
		"deactivation_reason": nil,
	}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to activate user",
		})
	}

	// Log audit
	if h.auditService != nil {
		userIDPtr := uint(0)
		// Logic disabled: UserID is now UUID, cannot fit in uint audit log
		/*if id, err := strconv.ParseUint(currentUser.ID.String(), 10, 32); err == nil {
			temp := uint(id)
			userIDPtr = temp
		}*/
		h.auditService.Log(&models.AuditLogEntry{
			UserID:       &userIDPtr,
			Username:     currentUser.Email,
			Action:       "UPDATE",
			ResourceType: "user",
			ResourceName: user.Email,
			Metadata: map[string]interface{}{
				"action": "activate",
				"userId": user.ID,
			},
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// DeactivateUserRequest defines payload for deactivating a user
type DeactivateUserRequest struct {
	Reason string `json:"reason" validate:"required,min=5,max=255"`
}

// DeactivateUser handles PUT /admin/users/:id/deactivate
func (h *AdminUserHandler) DeactivateUser(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User ID is required",
		})
	}

	// Get current admin user from context
	currentUser := c.Locals("user").(*models.User)
	if !currentUser.IsSuperAdmin() {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	var req DeactivateUserRequest
	if err := c.BodyParser(&req); err != nil {
		req.Reason = "Deactivated by admin"
	}

	// Optional: Validate reason if we want to enforce it
	if err := validator.GetValidator().ValidateStruct(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	var user models.User
	if err := h.db.Where("id = ?", id).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch user",
		})
	}

	// Cannot deactivate self
	if user.ID == currentUser.ID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot deactivate your own account",
		})
	}

	// Update status
	now := time.Now()
	if err := h.db.Model(&user).Updates(map[string]interface{}{
		"status":              models.UserStatusInactive,
		"deactivated_at":      &now,
		"deactivated_by":      &currentUser.ID,
		"deactivation_reason": &req.Reason,
	}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to deactivate user",
		})
	}

	// Log audit
	if h.auditService != nil {
		userIDPtr := uint(0)
		// Logic disabled: UserID is now UUID
		/*if id, err := strconv.ParseUint(currentUser.ID.String(), 10, 32); err == nil {
			temp := uint(id)
			userIDPtr = temp
		}*/
		h.auditService.Log(&models.AuditLogEntry{
			UserID:       &userIDPtr,
			Username:     currentUser.Email,
			Action:       "UPDATE",
			ResourceType: "user",
			ResourceName: user.Email,
			Metadata: map[string]interface{}{
				"action": "deactivate",
				"userId": user.ID,
				"reason": req.Reason,
			},
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// UpdateUserRoleRequest defines payload for updating user role
type UpdateUserRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=user admin"`
}

// UpdateUserRole handles PUT /admin/users/:id/role
func (h *AdminUserHandler) UpdateUserRole(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User ID is required",
		})
	}

	// Get current admin user from context
	currentUser := c.Locals("user").(*models.User)
	if !currentUser.IsSuperAdmin() {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	var req UpdateUserRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Use the validator package we imported in other files...
	// Note: We need to import the validator package in this file if not already present.
	// For now, manual validation as fallback if package not imported, but prefer strict.
	if req.Role != "user" && req.Role != "admin" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid role. Must be 'user' or 'admin'",
		})
	}

	var user models.User
	if err := h.db.Where("id = ?", id).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch user",
		})
	}

	// Cannot change own role
	if user.ID == currentUser.ID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot change your own role",
		})
	}

	// Update role
	if err := h.db.Model(&user).Update("role", req.Role).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update role",
		})
	}

	// Log audit
	if h.auditService != nil {
		userIDPtr := uint(0)
		/*if id, err := strconv.ParseUint(currentUser.ID.String(), 10, 32); err == nil {
			temp := uint(id)
			userIDPtr = temp
		}*/
		h.auditService.Log(&models.AuditLogEntry{
			UserID:       &userIDPtr,
			Username:     currentUser.Email,
			Action:       "UPDATE",
			ResourceType: "user",
			ResourceName: user.Email,
			Metadata: map[string]interface{}{
				"action":  "update_role",
				"userId":  user.ID,
				"newRole": req.Role,
			},
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ImpersonateUser handles POST /admin/users/:id/impersonate
func (h *AdminUserHandler) ImpersonateUser(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User ID is required",
		})
	}

	// Get current admin user from context
	currentUser := c.Locals("user").(*models.User)
	if !currentUser.IsSuperAdmin() {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	var user models.User
	if err := h.db.Where("id = ?", id).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch user",
		})
	}

	// Cannot impersonate self
	if user.ID == currentUser.ID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot impersonate yourself",
		})
	}

	// User must be active
	if !user.IsActive() {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot impersonate inactive user",
		})
	}

	// Generate impersonation token (JWT)
	token, err := services.GenerateJWT(user.ID.String(), user.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate token",
		})
	}

	// Store impersonation record
	expires := time.Now().Add(30 * time.Minute)
	adminID := currentUser.ID
	if err := h.db.Model(&user).Updates(map[string]interface{}{
		"impersonation_token":   token,
		"impersonation_expires": &expires,
		"impersonated_by":       &adminID,
	}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to store impersonation token",
		})
	}

	// Log audit
	if h.auditService != nil {
		userIDPtr := uint(0)
		// Logic disabled: UserID is now UUID
		/*if id, err := strconv.ParseUint(currentUser.ID.String(), 10, 32); err == nil {
			temp := uint(id)
			userIDPtr = temp
		}*/
		h.auditService.Log(&models.AuditLogEntry{
			UserID:       &userIDPtr,
			Username:     currentUser.Email,
			Action:       "EXECUTE",
			ResourceType: "user",
			ResourceName: user.Email,
			Metadata: map[string]interface{}{
				"action":          "impersonate",
				"targetUserId":    user.ID,
				"targetUserEmail": user.Email,
			},
		})
	}

	return c.JSON(fiber.Map{
		"token":     token,
		"expiresAt": expires,
		"user": fiber.Map{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
		},
	})
}

// GetUserActivity handles GET /admin/users/:id/activity
func (h *AdminUserHandler) GetUserActivity(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User ID is required",
		})
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(c.Query("pageSize", "20"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Get audit logs for this user
	var logs []models.AuditLog
	offset := (page - 1) * pageSize

	var total int64
	h.db.Model(&models.AuditLog{}).Where("user_id = ?", id).Count(&total)

	if err := h.db.
		Where("user_id = ?", id).
		Order("timestamp DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&logs).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch activity",
		})
	}

	totalPages := (int(total) + pageSize - 1) / pageSize

	return c.JSON(fiber.Map{
		"data": logs,
		"pagination": fiber.Map{
			"page":       page,
			"pageSize":   pageSize,
			"total":      total,
			"totalPages": totalPages,
		},
	})
}

// GetUserStats handles GET /admin/users/stats
func (h *AdminUserHandler) GetUserStats(c *fiber.Ctx) error {
	var stats struct {
		TotalUsers    int64 `json:"totalUsers"`
		ActiveUsers   int64 `json:"activeUsers"`
		InactiveUsers int64 `json:"inactiveUsers"`
		PendingUsers  int64 `json:"pendingUsers"`
		VerifiedUsers int64 `json:"verifiedUsers"`
		NewThisMonth  int64 `json:"newThisMonth"`
		OAuthUsers    int64 `json:"oauthUsers"`
	}

	// Total users
	h.db.Model(&models.User{}).Count(&stats.TotalUsers)

	// Active users
	h.db.Model(&models.User{}).Where("status = ?", models.UserStatusActive).Count(&stats.ActiveUsers)

	// Inactive users
	h.db.Model(&models.User{}).Where("status = ?", models.UserStatusInactive).Count(&stats.InactiveUsers)

	// Pending users
	h.db.Model(&models.User{}).Where("status = ?", models.UserStatusPending).Count(&stats.PendingUsers)

	// Verified users
	h.db.Model(&models.User{}).Where("email_verified = ?", true).Count(&stats.VerifiedUsers)

	// New this month
	firstDayOfMonth := time.Now().AddDate(0, 0, -time.Now().Day()+1)
	h.db.Model(&models.User{}).Where("created_at >= ?", firstDayOfMonth).Count(&stats.NewThisMonth)

	// OAuth users
	h.db.Model(&models.User{}).Where("provider IS NOT NULL AND provider != ''").Count(&stats.OAuthUsers)

	return c.JSON(stats)
}
