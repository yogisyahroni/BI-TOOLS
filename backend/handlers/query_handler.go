package handlers

import (
	"context"
	"insight-engine-backend/database"
	"insight-engine-backend/models"
	"insight-engine-backend/pkg/validator"
	"insight-engine-backend/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type QueryHandler struct {
	queryExecutor     *services.QueryExecutor
	encryptionService *services.EncryptionService
}

func NewQueryHandler(qe *services.QueryExecutor) *QueryHandler {
	// Initialize encryption service (fail gracefully if not configured)
	encryptionService, err := services.NewEncryptionService()
	if err != nil {
		encryptionService = nil
	}

	return &QueryHandler{
		queryExecutor:     qe,
		encryptionService: encryptionService,
	}
}

// GetQueries returns a list of saved queries
// @Summary List saved queries
// @Description Returns a list of saved queries for the authenticated user.
// @Tags Query
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /query [get]
func (h *QueryHandler) GetQueries(c *fiber.Ctx) error {
	// Get user ID from auth middleware
	userID, ok := c.Locals("userId").(string)
	if !ok {
		return c.Status(401).JSON(fiber.Map{
			"status":  "error",
			"message": "Unauthorized",
		})
	}

	var queries []models.SavedQuery
	result := database.DB.Where("user_id = ?", userID).
		Preload("Connection").
		Order("updated_at DESC").
		Find(&queries)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not fetch queries",
			"error":   result.Error.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    queries,
	})
}

// GetQuery returns a single query by ID
// @Summary Get a saved query
// @Description Returns a single saved query by ID.
// @Tags Query
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Query ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /query/{id} [get]
func (h *QueryHandler) GetQuery(c *fiber.Ctx) error {
	queryID := c.Params("id")
	userID, _ := c.Locals("userId").(string)

	var query models.SavedQuery
	result := database.DB.Where("id = ? AND user_id = ?", queryID, userID).
		Preload("Connection").
		First(&query)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Query not found",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    query,
	})
}

// CreateQuery creates a new saved query
// @Summary Create a saved query
// @Description Creates a new saved query.
// @Tags Query
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.SavedQuery true "Saved Query"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /query [post]
func (h *QueryHandler) CreateQuery(c *fiber.Ctx) error {
	userID, _ := c.Locals("userId").(string)

	var input struct {
		Name         string `json:"name" validate:"required"`
		SQL          string `json:"sql" validate:"required"`
		Description  string `json:"description"`
		ConnectionID string `json:"connectionId" validate:"required"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid input",
			"error":   err.Error(),
		})
	}

	if err := validator.GetValidator().ValidateStruct(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	query := models.SavedQuery{
		ID:           uuid.New().String(),
		UserID:       userID,
		Name:         input.Name,
		SQL:          input.SQL,
		Description:  &input.Description,
		ConnectionID: input.ConnectionID,
	}

	result := database.DB.Create(&query)
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not create query",
			"error":   result.Error.Error(),
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"success": true,
		"data":    query,
	})
}

// UpdateQuery updates an existing query
// @Summary Update a saved query
// @Description Updates an existing saved query.
// @Tags Query
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Query ID"
// @Param request body models.SavedQuery true "Saved Query Updates"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /query/{id} [put]
func (h *QueryHandler) UpdateQuery(c *fiber.Ctx) error {
	queryID := c.Params("id")
	userID, _ := c.Locals("userId").(string)

	// Check ownership
	var existing models.SavedQuery
	if err := database.DB.Where("id = ? AND user_id = ?", queryID, userID).First(&existing).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Query not found",
		})
	}

	var input struct {
		Name         string `json:"name"`
		SQL          string `json:"sql"`
		Description  string `json:"description"`
		ConnectionID string `json:"connectionId"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid input",
			"error":   err.Error(),
		})
	}

	// Update fields if provided
	if input.Name != "" {
		existing.Name = input.Name
	}
	if input.SQL != "" {
		existing.SQL = input.SQL
	}
	if input.Description != "" {
		existing.Description = &input.Description
	}
	if input.ConnectionID != "" {
		existing.ConnectionID = input.ConnectionID
	}

	if err := database.DB.Save(&existing).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not update query",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    existing,
	})
}

// DeleteQuery deletes a query
// @Summary Delete a saved query
// @Description Deletes a saved query by ID.
// @Tags Query
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Query ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /query/{id} [delete]
func (h *QueryHandler) DeleteQuery(c *fiber.Ctx) error {
	queryID := c.Params("id")
	userID, _ := c.Locals("userId").(string)

	result := database.DB.Where("id = ? AND user_id = ?", queryID, userID).Delete(&models.SavedQuery{})
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not delete query",
			"error":   result.Error.Error(),
		})
	}

	if result.RowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Query not found",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Query deleted",
	})
}

// RunQuery executes a saved query
// @Summary Run a saved query
// @Description Executes a saved query by ID with optional limit and offset.
// @Tags Query
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Query ID"
// @Param request body map[string]interface{} false "Run Parameters (limit, offset)"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /query/{id}/run [post]
func (h *QueryHandler) RunQuery(c *fiber.Ctx) error {
	queryID := c.Params("id")
	userID, _ := c.Locals("userId").(string)

	// Fetch query
	var query models.SavedQuery
	if err := database.DB.Where("id = ? AND user_id = ?", queryID, userID).
		Preload("Connection").
		First(&query).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Query not found",
		})
	}

	// Decrypt password
	if h.encryptionService != nil && query.Connection.Password != nil && *query.Connection.Password != "" {
		decryptedPassword, err := h.encryptionService.Decrypt(*query.Connection.Password)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to decrypt password",
				"error":   err.Error(),
			})
		}
		query.Connection.Password = &decryptedPassword
	}

	// Parse request body for limit/offset
	type RunParams struct {
		Limit  *int `json:"limit" validate:"omitempty,min=0"`
		Offset *int `json:"offset" validate:"omitempty,min=0"`
	}
	params := new(RunParams)
	if err := c.BodyParser(params); err != nil {
		// Ignore body parser error as params are optional
	}

	if err := validator.GetValidator().ValidateStruct(params); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// Execute query
	ctx := context.Background()
	// Updated: include params (nil)
	result, err := h.queryExecutor.Execute(ctx, query.Connection, query.SQL, nil, params.Limit, params.Offset)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Query execution failed",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// ExecuteAdHocQuery executes a query without saving it
// @Summary Execute ad-hoc query
// @Description Executes a raw SQL query on a specific connection.
// @Tags Query
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.QueryExecutionRequest true "Query Request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /query/execute [post]
func (h *QueryHandler) ExecuteAdHocQuery(c *fiber.Ctx) error {
	userID, _ := c.Locals("userId").(string)

	var req struct {
		ConnectionID string `json:"connectionId" validate:"required"`
		SQL          string `json:"sql" validate:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid input",
			"error":   err.Error(),
		})
	}

	if err := validator.GetValidator().ValidateStruct(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// Fetch connection
	var conn models.Connection
	if err := database.DB.Where("id = ? AND user_id = ?", req.ConnectionID, userID).First(&conn).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Connection not found",
		})
	}

	// Decrypt password
	if h.encryptionService != nil && conn.Password != nil && *conn.Password != "" {
		decryptedPassword, err := h.encryptionService.Decrypt(*conn.Password)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to decrypt password",
				"error":   err.Error(),
			})
		}
		conn.Password = &decryptedPassword
	}

	// Execute query
	ctx := c.UserContext()
	// Updated: include params (nil)
	result, err := h.queryExecutor.Execute(ctx, &conn, req.SQL, nil, nil, nil)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Query execution failed",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}
