package handlers

import (
	"errors"
	"insight-engine-backend/models"
	"insight-engine-backend/services"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// VersionHandler handles dashboard versioning HTTP requests
type VersionHandler struct {
	db              *gorm.DB
	versionSvc      *services.VersionService
	notificationSvc *services.NotificationService
}

// NewVersionHandler creates a new version handler
func NewVersionHandler(db *gorm.DB, notificationSvc *services.NotificationService) *VersionHandler {
	return &VersionHandler{
		db:              db,
		versionSvc:      services.NewVersionService(db, notificationSvc),
		notificationSvc: notificationSvc,
	}
}

// ===========================================
// VERSION ENDPOINTS
// ===========================================

// CreateVersionRequest represents the request body for creating a version
type CreateVersionRequest struct {
	ChangeSummary string                 `json:"change_summary,omitempty"`
	IsAutoSave    bool                   `json:"is_auto_save"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// CreateVersion handles POST /api/dashboards/:id/versions
func (h *VersionHandler) CreateVersion(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	dashboardID := c.Params("id")

	var req CreateVersionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Convert to model request
	modelReq := &models.DashboardVersionCreateRequest{
		ChangeSummary: req.ChangeSummary,
		IsAutoSave:    req.IsAutoSave,
		Metadata:      req.Metadata,
	}

	version, err := h.versionSvc.CreateVersion(dashboardID, userID, modelReq)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    version,
	})
}

// GetVersions handles GET /api/dashboards/:id/versions
func (h *VersionHandler) GetVersions(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	dashboardID := c.Params("id")

	// Parse query parameters
	filter := &models.DashboardVersionFilter{}

	// Parse pagination
	limitStr := c.Query("limit", "20")
	offsetStr := c.Query("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	filter.Limit = limit
	filter.Offset = offset

	// Parse is_auto_save filter
	if autoSave := c.Query("is_auto_save"); autoSave != "" {
		isAutoSave := autoSave == "true"
		filter.IsAutoSave = &isAutoSave
	}

	// Parse order_by
	filter.OrderBy = c.Query("order_by", "date_desc")

	versions, total, err := h.versionSvc.GetVersions(dashboardID, userID, filter)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"versions": versions,
			"total":    total,
			"limit":    limit,
			"offset":   offset,
		},
	})
}

// GetVersion handles GET /api/versions/:id
func (h *VersionHandler) GetVersion(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	versionID := c.Params("id")

	version, err := h.versionSvc.GetVersion(versionID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Version not found",
			})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    version,
	})
}

// RestoreVersion handles POST /api/versions/:id/restore
func (h *VersionHandler) RestoreVersion(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	versionID := c.Params("id")

	response, err := h.versionSvc.RestoreVersion(versionID, userID)
	if err != nil {
		if err.Error() == "version not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// CompareVersionsRequest represents the request body for comparing versions
type CompareVersionsRequest struct {
	VersionID1 string `json:"version_id_1" validate:"required"`
	VersionID2 string `json:"version_id_2" validate:"required"`
}

// CompareVersions handles GET /api/versions/compare
func (h *VersionHandler) CompareVersions(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)

	versionID1 := c.Query("version_id_1")
	versionID2 := c.Query("version_id_2")

	if versionID1 == "" || versionID2 == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "version_id_1 and version_id_2 are required",
		})
	}

	diff, err := h.versionSvc.CompareVersions(versionID1, versionID2, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    diff,
	})
}

// DeleteVersion handles DELETE /api/versions/:id
func (h *VersionHandler) DeleteVersion(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	versionID := c.Params("id")

	if err := h.versionSvc.DeleteVersion(versionID, userID); err != nil {
		if err.Error() == "version not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Version deleted successfully",
	})
}

// AutoSaveVersion handles POST /api/dashboards/:id/versions/auto-save
func (h *VersionHandler) AutoSaveVersion(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	dashboardID := c.Params("id")

	version, err := h.versionSvc.AutoSaveVersion(dashboardID, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    version,
	})
}

// RegisterRoutes registers version routes with the router
func (h *VersionHandler) RegisterRoutes(router fiber.Router) {
	// Dashboard versions
	router.Post("/dashboards/:id/versions", h.CreateVersion)
	router.Get("/dashboards/:id/versions", h.GetVersions)
	router.Post("/dashboards/:id/versions/auto-save", h.AutoSaveVersion)

	// Individual versions
	router.Get("/versions/:id", h.GetVersion)
	router.Post("/versions/:id/restore", h.RestoreVersion)
	router.Delete("/versions/:id", h.DeleteVersion)
	router.Get("/versions/compare", h.CompareVersions)
}
