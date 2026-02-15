package handlers

import (
	"errors"
	"insight-engine-backend/models"
	"insight-engine-backend/services"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// ShareHandler handles resource sharing requests
type ShareHandler struct {
	db           *gorm.DB
	shareService *services.ShareService
	auditService *services.AuditService
}

// NewShareHandler creates a new share handler
func NewShareHandler(db *gorm.DB, auditService *services.AuditService) *ShareHandler {
	return &ShareHandler{
		db:           db,
		shareService: services.NewShareService(db, auditService),
		auditService: auditService,
	}
}

// ============================
// SHARE ENDPOINTS
// ============================

// CreateShareRequest represents the request body for creating a share
type CreateShareRequest struct {
	ResourceType string  `json:"resource_type" validate:"required,oneof=dashboard query"`
	ResourceID   string  `json:"resource_id" validate:"required"`
	SharedWith   *string `json:"shared_with,omitempty"`
	SharedEmail  *string `json:"shared_email,omitempty"`
	Permission   string  `json:"permission" validate:"required,oneof=view edit admin"`
	Password     *string `json:"password,omitempty"`
	ExpiresAt    *string `json:"expires_at,omitempty"`
	Message      *string `json:"message,omitempty"`
}

// CreateShare handles POST /api/shares
func (h *ShareHandler) CreateShare(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)

	var req CreateShareRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validation
	if req.SharedWith == nil && req.SharedEmail == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Either shared_with (user ID) or shared_email must be provided",
		})
	}

	// Parse expiration date if provided
	var expiresAtTime *time.Time
	if req.ExpiresAt != nil && *req.ExpiresAt != "" {
		if parsed, err := time.Parse(time.RFC3339, *req.ExpiresAt); err == nil {
			expiresAtTime = &parsed
		}
	}

	// Convert to model request
	modelReq := &models.ShareCreateRequest{
		ResourceType: models.ResourceType(req.ResourceType),
		ResourceID:   req.ResourceID,
		SharedWith:   req.SharedWith,
		SharedEmail:  req.SharedEmail,
		Permission:   models.SharePermission(req.Permission),
		Password:     req.Password,
		Message:      req.Message,
		ExpiresAt:    expiresAtTime,
	}

	share, err := h.shareService.CreateShare(modelReq, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Log audit event
	if h.auditService != nil {
		username := c.Locals("username").(string)
		h.auditService.LogCreate(c, nil, username, "share", nil, share.ID, map[string]interface{}{
			"resource_type": req.ResourceType,
			"resource_id":   req.ResourceID,
			"permission":    req.Permission,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(share)
}

// GetSharesForResource handles GET /api/shares/resource/:type/:id
func (h *ShareHandler) GetSharesForResource(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	resourceType := c.Params("type")
	resourceID := c.Params("id")

	// Validate resource type
	resType, valid := models.ValidateResourceType(resourceType)
	if !valid {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid resource type. Must be 'dashboard' or 'query'",
		})
	}

	shares, err := h.shareService.GetSharesForResource(resType, resourceID, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"shares": shares,
		"total":  len(shares),
	})
}

// GetMyShares handles GET /api/shares/my
func (h *ShareHandler) GetMyShares(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)

	// Parse query parameters for filtering
	filter := &models.ShareFilter{}

	if resourceType := c.Query("resource_type"); resourceType != "" {
		if rt, valid := models.ValidateResourceType(resourceType); valid {
			filter.ResourceType = &rt
		}
	}

	if status := c.Query("status"); status != "" {
		s := models.ShareStatus(status)
		filter.Status = &s
	}

	filter.IncludeExpired = c.Query("include_expired") == "true"

	shares, err := h.shareService.GetMyShares(userID, filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch shares",
		})
	}

	return c.JSON(fiber.Map{
		"shares": shares,
		"total":  len(shares),
	})
}

// GetShareByID handles GET /api/shares/:id
func (h *ShareHandler) GetShareByID(c *fiber.Ctx) error {
	shareID := c.Params("id")

	share, err := h.shareService.GetShareByID(shareID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Share not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(share)
}

// UpdateShareRequest represents the request body for updating a share
type UpdateShareRequest struct {
	Permission *string `json:"permission,omitempty"`
	Password   *string `json:"password,omitempty"`
	ExpiresAt  *string `json:"expires_at,omitempty"`
	Message    *string `json:"message,omitempty"`
}

// UpdateShare handles PUT /api/shares/:id
func (h *ShareHandler) UpdateShare(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	shareID := c.Params("id")

	var req UpdateShareRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Parse expiration date if provided
	var expiresAtTime *time.Time
	if req.ExpiresAt != nil && *req.ExpiresAt != "" {
		if parsed, err := time.Parse(time.RFC3339, *req.ExpiresAt); err == nil {
			expiresAtTime = &parsed
		}
	}

	// Convert to model request
	modelReq := &models.ShareUpdateRequest{}

	if req.Permission != nil {
		perm := models.SharePermission(*req.Permission)
		modelReq.Permission = &perm
	}

	if req.Password != nil {
		modelReq.Password = req.Password
	}

	modelReq.ExpiresAt = expiresAtTime

	if req.Message != nil {
		modelReq.Message = req.Message
	}

	share, err := h.shareService.UpdateShare(shareID, modelReq, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Log audit event
	if h.auditService != nil {
		username := c.Locals("username").(string)
		h.auditService.LogUpdate(c, nil, username, "share", nil, share.ID, nil, map[string]interface{}{
			"permission": req.Permission,
			"expires_at": req.ExpiresAt,
		})
	}

	return c.JSON(share)
}

// DeleteShare handles DELETE /api/shares/:id
func (h *ShareHandler) DeleteShare(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	shareID := c.Params("id")

	if err := h.shareService.RevokeShare(shareID, userID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Log audit event
	if h.auditService != nil {
		username := c.Locals("username").(string)
		h.auditService.LogDelete(c, nil, username, "share", nil, shareID, nil)
	}

	return c.JSON(fiber.Map{
		"message": "Share revoked successfully",
	})
}

// AcceptShareRequest represents the request body for accepting a share
type AcceptShareRequest struct {
	Password *string `json:"password,omitempty"`
}

// AcceptShare handles POST /api/shares/:id/accept
func (h *ShareHandler) AcceptShare(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	shareID := c.Params("id")

	var req AcceptShareRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.shareService.AcceptShare(shareID, userID, req.Password); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Share accepted successfully",
	})
}

// CheckShareAccessRequest represents the request body for checking share access
type CheckShareAccessRequest struct {
	ResourceType string `json:"resource_type" validate:"required"`
	ResourceID   string `json:"resource_id" validate:"required"`
}

// CheckShareAccess handles GET /api/shares/check
func (h *ShareHandler) CheckShareAccess(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	resourceType := c.Query("resource_type")
	resourceID := c.Query("resource_id")

	if resourceType == "" || resourceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "resource_type and resource_id are required",
		})
	}

	// Validate resource type
	resType, valid := models.ValidateResourceType(resourceType)
	if !valid {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid resource type",
		})
	}

	access, err := h.shareService.CheckAccess(userID, resType, resourceID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(access)
}

// ValidateShareAccessRequest represents the request body for validating share access with password
type ValidateShareAccessRequest struct {
	ShareID  string  `json:"share_id" validate:"required"`
	Password *string `json:"password,omitempty"`
}

// ValidateShareAccess handles POST /api/shares/validate
func (h *ShareHandler) ValidateShareAccess(c *fiber.Ctx) error {
	var req ValidateShareAccessRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	share, err := h.shareService.ValidateShareAccess(req.ShareID, req.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"valid":         true,
		"share_id":      share.ID,
		"resource_type": share.ResourceType,
		"resource_id":   share.ResourceID,
		"permission":    share.Permission,
	})
}
