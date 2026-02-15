package handlers

import (
	"encoding/json"
	"fmt"
	"insight-engine-backend/models"
	"insight-engine-backend/pkg/validator"
	"insight-engine-backend/services"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// VisualQueryHandler handles visual query builder API endpoints
type VisualQueryHandler struct {
	db              *gorm.DB
	queryBuilder    *services.QueryBuilder
	queryExecutor   *services.QueryExecutor
	schemaDiscovery *services.SchemaDiscovery
	queryCache      *services.QueryCache
}

// NewVisualQueryHandler creates a new visual query handler
func NewVisualQueryHandler(
	db *gorm.DB,
	queryBuilder *services.QueryBuilder,
	queryExecutor *services.QueryExecutor,
	schemaDiscovery *services.SchemaDiscovery,
	queryCache *services.QueryCache,
) *VisualQueryHandler {
	return &VisualQueryHandler{
		db:              db,
		queryBuilder:    queryBuilder,
		queryExecutor:   queryExecutor,
		schemaDiscovery: schemaDiscovery,
		queryCache:      queryCache,
	}
}

// getUserContext helper to extract user context for RLS
func (h *VisualQueryHandler) getUserContext(c *fiber.Ctx) (string, string, *string) {
	// User ID from auth middleware
	userID := c.Locals("userID").(string)

	// Workspace ID from query param or header
	workspaceID := c.Query("workspaceId")
	if workspaceID == "" {
		workspaceID = c.Get("X-Workspace-ID")
	}

	// User Role
	var userRole *string
	if role, ok := c.Locals("role").(string); ok && role != "" {
		userRole = &role
	}

	return userID, workspaceID, userRole
}

// CreateVisualQueryRequest defines the schema for creating a visual query
type CreateVisualQueryRequest struct {
	Name         string                   `json:"name" validate:"required,min=3,max=100"`
	Description  *string                  `json:"description" validate:"omitempty,max=500"`
	ConnectionID string                   `json:"connectionId" validate:"required,uuid"`
	CollectionID string                   `json:"collectionId" validate:"required"`
	Config       models.VisualQueryConfig `json:"config" validate:"required"`
	Tags         []string                 `json:"tags" validate:"omitempty,max=10"`
	Pinned       bool                     `json:"pinned"`
}

// CreateVisualQuery creates a new visual query
// POST /api/visual-queries
func (h *VisualQueryHandler) CreateVisualQuery(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var req CreateVisualQueryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Strict Validation
	if err := validator.GetValidator().ValidateStruct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error(), "details": "Validation failed"})
	}

	// Get connection
	var conn models.Connection
	if err := h.db.Where("id = ? AND user_id = ?", req.ConnectionID, userID).First(&conn).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Connection not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch connection"})
	}

	// Get context for RLS
	userIDStr, workspaceID, userRole := h.getUserContext(c)

	// Generate SQL
	generatedSQL, _, err := h.queryBuilder.BuildSQL(c.UserContext(), &req.Config, &conn, userIDStr, workspaceID, userRole)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("Failed to generate SQL: %v", err)})
	}

	configJSON, err := json.Marshal(req.Config)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to marshal config"})
	}

	visualQuery := models.VisualQuery{
		ID:           uuid.New().String(),
		Name:         req.Name,
		Description:  req.Description,
		ConnectionID: req.ConnectionID,
		CollectionID: req.CollectionID,
		UserID:       userID.(string),
		Config:       configJSON,
		GeneratedSQL: &generatedSQL,
		Tags:         req.Tags,
		Pinned:       req.Pinned,
	}

	if err := h.db.Create(&visualQuery).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create visual query"})
	}

	return c.Status(fiber.StatusCreated).JSON(visualQuery.ToDTO())
}

// GetVisualQuery retrieves a visual query by ID
// GET /api/visual-queries/:id
func (h *VisualQueryHandler) GetVisualQuery(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	id := c.Params("id")

	var visualQuery models.VisualQuery
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&visualQuery).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Visual query not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch visual query"})
	}

	return c.JSON(visualQuery.ToDTO())
}

// UpdateVisualQueryRequest defines the schema for updating a visual query
type UpdateVisualQueryRequest struct {
	Name        string                   `json:"name" validate:"required,min=3,max=100"`
	Description *string                  `json:"description" validate:"omitempty,max=500"`
	Config      models.VisualQueryConfig `json:"config" validate:"required"`
	Tags        []string                 `json:"tags" validate:"omitempty,max=10"`
	Pinned      bool                     `json:"pinned"`
}

// UpdateVisualQuery updates a visual query
// PUT /api/visual-queries/:id
func (h *VisualQueryHandler) UpdateVisualQuery(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	id := c.Params("id")

	var req UpdateVisualQueryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := validator.GetValidator().ValidateStruct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error(), "details": "Validation failed"})
	}

	// Fetch existing visual query
	var visualQuery models.VisualQuery
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&visualQuery).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Visual query not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch visual query"})
	}

	// Get connection
	var conn models.Connection
	if err := h.db.Where("id = ?", visualQuery.ConnectionID).First(&conn).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch connection"})
	}

	// Get context for RLS
	userIDStr, workspaceID, userRole := h.getUserContext(c)

	// Generate SQL
	generatedSQL, _, err := h.queryBuilder.BuildSQL(c.UserContext(), &req.Config, &conn, userIDStr, workspaceID, userRole)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("Failed to generate SQL: %v", err)})
	}

	configJSON, err := json.Marshal(req.Config)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to marshal config"})
	}

	// Update fields
	visualQuery.Name = req.Name
	visualQuery.Description = req.Description
	visualQuery.Config = configJSON
	visualQuery.GeneratedSQL = &generatedSQL
	visualQuery.Tags = req.Tags
	visualQuery.Pinned = req.Pinned

	if err := h.db.Save(&visualQuery).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update visual query"})
	}

	// Invalidate cache
	if h.queryCache != nil {
		_ = h.queryCache.InvalidateQuery(c.Context(), id)
	}

	return c.JSON(visualQuery.ToDTO())
}

// DeleteVisualQuery deletes a visual query
// DELETE /api/visual-queries/:id
func (h *VisualQueryHandler) DeleteVisualQuery(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	id := c.Params("id")

	var visualQuery models.VisualQuery
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&visualQuery).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Visual query not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch visual query"})
	}

	if err := h.db.Delete(&visualQuery).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete visual query"})
	}

	if h.queryCache != nil {
		_ = h.queryCache.InvalidateQuery(c.Context(), id)
	}

	return c.JSON(fiber.Map{"message": "Visual query deleted successfully"})
}

// GenerateSQLRequest defines schema for SQL generation
type GenerateSQLRequest struct {
	ConnectionID string                   `json:"connectionId" validate:"required,uuid"`
	Config       models.VisualQueryConfig `json:"config" validate:"required"`
}

// GenerateSQL generates SQL from visual configuration
// POST /api/visual-queries/generate-sql
func (h *VisualQueryHandler) GenerateSQL(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var req GenerateSQLRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := validator.GetValidator().ValidateStruct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error(), "details": "Validation failed"})
	}

	var conn models.Connection
	if err := h.db.Where("id = ? AND user_id = ?", req.ConnectionID, userID).First(&conn).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Connection not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch connection"})
	}

	userIDStr, workspaceID, userRole := h.getUserContext(c)

	generatedSQL, params, err := h.queryBuilder.BuildSQL(c.UserContext(), &req.Config, &conn, userIDStr, workspaceID, userRole)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("Failed to generate SQL: %v", err)})
	}

	return c.JSON(fiber.Map{
		"sql":    generatedSQL,
		"params": params,
	})
}

// PreviewVisualQuery executes a visual query and returns preview results
// POST /api/visual-queries/:id/preview
func (h *VisualQueryHandler) PreviewVisualQuery(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	id := c.Params("id")

	var visualQuery models.VisualQuery
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&visualQuery).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Visual query not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch visual query"})
	}

	var conn models.Connection
	if err := h.db.Where("id = ?", visualQuery.ConnectionID).First(&conn).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch connection"})
	}

	var config models.VisualQueryConfig
	if err := json.Unmarshal(visualQuery.Config, &config); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse config"})
	}

	previewLimit := 100
	config.Limit = &previewLimit

	userIDStr, workspaceID, userRole := h.getUserContext(c)

	generatedSQL, _, err := h.queryBuilder.BuildSQL(c.UserContext(), &config, &conn, userIDStr, workspaceID, userRole)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("Failed to generate SQL: %v", err)})
	}

	result, err := h.queryBuilder.ExecuteQuery(c.UserContext(), &config, &conn, userIDStr, id, workspaceID, userRole)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Failed to execute query: %v", err)})
	}

	return c.JSON(fiber.Map{
		"sql":     generatedSQL,
		"results": result,
	})
}

// GetVisualQueries retrieves user's visual queries with pagination
// GET /api/visual-queries
func (h *VisualQueryHandler) GetVisualQueries(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	collectionID := c.Query("collectionId")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	query := h.db.Where("user_id = ?", userID)
	if collectionID != "" {
		query = query.Where("collection_id = ?", collectionID)
	}

	var total int64
	if err := query.Model(&models.VisualQuery{}).Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to count visual queries"})
	}

	var visualQueries []models.VisualQuery
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&visualQueries).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch visual queries"})
	}

	dtos := make([]models.VisualQueryDTO, len(visualQueries))
	for i, vq := range visualQueries {
		dtos[i] = vq.ToDTO()
	}

	return c.JSON(fiber.Map{
		"data":  dtos,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetCacheStats returns cache statistics
// GET /api/v1/visual-queries/cache/stats
func (h *VisualQueryHandler) GetCacheStats(c *fiber.Ctx) error {
	if h.queryCache == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error":   "Cache is not enabled",
			"message": "Redis cache is not configured or unavailable",
		})
	}

	stats, err := h.queryCache.GetStats(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to get cache stats: %v", err),
		})
	}

	return c.JSON(stats)
}

// JoinSuggestionsRequest defines schema for join suggestions
type JoinSuggestionsRequest struct {
	ConnectionID string   `json:"connectionId" validate:"required,uuid"`
	TableNames   []string `json:"tableNames" validate:"required,min=1,max=10"`
}

// GetJoinSuggestions suggests joins for selected tables
// POST /api/visual-queries/join-suggestions
func (h *VisualQueryHandler) GetJoinSuggestions(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var req JoinSuggestionsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := validator.GetValidator().ValidateStruct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error(), "details": "Validation failed"})
	}

	var conn models.Connection
	if err := h.db.Where("id = ? AND user_id = ?", req.ConnectionID, userID).First(&conn).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Connection not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch connection"})
	}

	suggestions, err := h.schemaDiscovery.GetJoinSuggestions(c.UserContext(), &conn, req.TableNames)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Failed to get join suggestions: %v", err)})
	}

	return c.JSON(fiber.Map{"suggestions": suggestions})
}
