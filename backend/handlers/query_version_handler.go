package handlers

import (
	"errors"
	"insight-engine-backend/models"
	"insight-engine-backend/services"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// QueryVersionHandler handles query versioning HTTP requests
type QueryVersionHandler struct {
	db              *gorm.DB
	queryVersionSvc *services.QueryVersionService
	notificationSvc *services.NotificationService
}

// NewQueryVersionHandler creates a new query version handler
func NewQueryVersionHandler(db *gorm.DB, notificationSvc *services.NotificationService) *QueryVersionHandler {
	return &QueryVersionHandler{
		db:              db,
		queryVersionSvc: services.NewQueryVersionService(db, notificationSvc),
		notificationSvc: notificationSvc,
	}
}

// ===========================================
// QUERY VERSION ENDPOINTS
// ===========================================

// CreateQueryVersionRequest represents the request body for creating a query version
type CreateQueryVersionRequest struct {
	ChangeSummary string                 `json:"change_summary,omitempty"`
	IsAutoSave    bool                   `json:"is_auto_save"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// CreateQueryVersion handles POST /api/queries/:id/versions
func (h *QueryVersionHandler) CreateQueryVersion(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	queryID := c.Params("id")

	var req CreateQueryVersionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Convert to model request
	modelReq := &models.QueryVersionCreateRequest{
		ChangeSummary: req.ChangeSummary,
		IsAutoSave:    req.IsAutoSave,
		Metadata:      req.Metadata,
	}

	version, err := h.queryVersionSvc.CreateVersion(queryID, userID, modelReq)
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

// GetQueryVersions handles GET /api/queries/:id/versions
func (h *QueryVersionHandler) GetQueryVersions(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	queryID := c.Params("id")

	// Parse query parameters
	filter := &models.QueryVersionFilter{}

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

	versions, total, err := h.queryVersionSvc.GetVersions(queryID, userID, filter)
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

// GetQueryVersion handles GET /api/query-versions/:id
func (h *QueryVersionHandler) GetQueryVersion(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	versionID := c.Params("id")

	version, err := h.queryVersionSvc.GetVersion(versionID, userID)
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

// RestoreQueryVersion handles POST /api/query-versions/:id/restore
func (h *QueryVersionHandler) RestoreQueryVersion(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	versionID := c.Params("id")

	response, err := h.queryVersionSvc.RestoreVersion(versionID, userID)
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

// CompareQueryVersions handles GET /api/query-versions/compare
func (h *QueryVersionHandler) CompareQueryVersions(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)

	versionID1 := c.Query("version_id_1")
	versionID2 := c.Query("version_id_2")

	if versionID1 == "" || versionID2 == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "version_id_1 and version_id_2 are required",
		})
	}

	diff, err := h.queryVersionSvc.CompareVersions(versionID1, versionID2, userID)
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

// DeleteQueryVersion handles DELETE /api/query-versions/:id
func (h *QueryVersionHandler) DeleteQueryVersion(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	versionID := c.Params("id")

	if err := h.queryVersionSvc.DeleteVersion(versionID, userID); err != nil {
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
		"message": "Query version deleted successfully",
	})
}

// AutoSaveQueryVersion handles POST /api/queries/:id/versions/auto-save
func (h *QueryVersionHandler) AutoSaveQueryVersion(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	queryID := c.Params("id")

	version, err := h.queryVersionSvc.AutoSaveVersion(queryID, userID)
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

// RegisterRoutes registers query version routes with the router
func (h *QueryVersionHandler) RegisterRoutes(router fiber.Router) {
	// Query versions
	router.Post("/queries/:id/versions", h.CreateQueryVersion)
	router.Get("/queries/:id/versions", h.GetQueryVersions)
	router.Post("/queries/:id/versions/auto-save", h.AutoSaveQueryVersion)

	// Individual versions
	router.Get("/query-versions/:id", h.GetQueryVersion)
	router.Post("/query-versions/:id/restore", h.RestoreQueryVersion)
	router.Delete("/query-versions/:id", h.DeleteQueryVersion)
	router.Get("/query-versions/compare", h.CompareQueryVersions)
}
