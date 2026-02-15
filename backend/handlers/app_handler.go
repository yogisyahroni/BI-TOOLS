package handlers

import (
	"insight-engine-backend/database"
	"insight-engine-backend/models"
	"insight-engine-backend/pkg/validator"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ==================== APP HANDLERS ====================

// GetApps returns all apps for the authenticated user
// @Summary List apps
// @Description Returns a list of apps for the authenticated user.
// @Tags App
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} []models.App
// @Failure 500 {object} map[string]interface{}
// @Router /app [get]
func GetApps(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var apps []models.App
	if err := database.DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&apps).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(apps)
}

// CreateAppRequest defines payload for creating an app
type CreateAppRequest struct {
	Name        string  `json:"name" validate:"required,min=3,max=100"`
	Description *string `json:"description"`
	WorkspaceID *string `json:"workspaceId"`
}

// CreateApp creates a new app
// @Summary Create an app
// @Description Creates a new application.
// @Tags App
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateAppRequest true "App Details"
// @Success 201 {object} models.App
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /app [post]
func CreateApp(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var req CreateAppRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := validator.GetValidator().ValidateStruct(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	app := models.App{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		UserID:      userID,
		WorkspaceID: req.WorkspaceID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := database.DB.Create(&app).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(app)
}

// GetApp returns a single app by ID
// @Summary Get an app
// @Description Returns a single app by ID.
// @Tags App
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "App ID"
// @Success 200 {object} models.App
// @Failure 404 {object} map[string]interface{}
// @Router /app/{id} [get]
func GetApp(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("userID").(string)

	var app models.App
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&app).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "App not found"})
	}

	return c.JSON(app)
}

// UpdateAppRequest defines payload for updating an app
type UpdateAppRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=3,max=100"`
	Description *string `json:"description"`
}

// UpdateApp updates an app (owner only)
// @Summary Update an app
// @Description Updates an existing application.
// @Tags App
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "App ID"
// @Param request body UpdateAppRequest true "App Updates"
// @Success 200 {object} models.App
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /app/{id} [put]
func UpdateApp(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("userID").(string)

	var app models.App
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&app).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "App not found"})
	}

	var req UpdateAppRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := validator.GetValidator().ValidateStruct(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Explicit field mapping
	if req.Name != nil {
		app.Name = *req.Name
	}
	if req.Description != nil {
		app.Description = req.Description
	}

	app.UpdatedAt = time.Now()

	if err := database.DB.Save(&app).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(app)
}

// DeleteApp deletes an app (owner only, cascade delete canvases & widgets)
// @Summary Delete an app
// @Description Deletes an application by ID.
// @Tags App
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "App ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /app/{id} [delete]
func DeleteApp(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("userID").(string)

	var app models.App
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&app).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "App not found"})
	}

	// Delete all canvases and widgets (cascade)
	var canvases []models.Canvas
	database.DB.Where("app_id = ?", id).Find(&canvases)
	for _, canvas := range canvases {
		database.DB.Where("canvas_id = ?", canvas.ID).Delete(&models.Widget{})
	}
	database.DB.Where("app_id = ?", id).Delete(&models.Canvas{})

	// Delete app
	if err := database.DB.Delete(&app).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "App deleted successfully"})
}

// ==================== CANVAS HANDLERS ====================

// GetCanvases returns all canvases for an app
// @Summary List canvases
// @Description Returns all canvases for a specific app.
// @Tags Canvas
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param appId query string true "App ID"
// @Success 200 {object} []models.Canvas
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /canvas [get]
func GetCanvases(c *fiber.Ctx) error {
	appID := c.Query("appId")
	userID := c.Locals("userID").(string)

	if appID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "appId is required"})
	}

	// Verify app ownership
	var app models.App
	if err := database.DB.Where("id = ? AND user_id = ?", appID, userID).First(&app).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "App not found"})
	}

	var canvases []models.Canvas
	if err := database.DB.Where("app_id = ?", appID).
		Order("created_at DESC").
		Find(&canvases).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(canvases)
}

// CreateCanvasRequest defines payload for creating a canvas
type CreateCanvasRequest struct {
	AppID  string       `json:"appId" validate:"required,uuid"`
	Name   string       `json:"name" validate:"required,min=3,max=100"`
	Config models.JSONB `json:"config"`
}

// CreateCanvas creates a new canvas
// @Summary Create a canvas
// @Description Creates a new canvas within an app.
// @Tags Canvas
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateCanvasRequest true "Canvas Details"
// @Success 201 {object} models.Canvas
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /canvas [post]
func CreateCanvas(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var req CreateCanvasRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := validator.GetValidator().ValidateStruct(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Verify app ownership
	var app models.App
	if err := database.DB.Where("id = ? AND user_id = ?", req.AppID, userID).First(&app).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "App not found"})
	}

	canvas := models.Canvas{
		ID:        uuid.New().String(),
		AppID:     req.AppID,
		Name:      req.Name,
		Config:    req.Config,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := database.DB.Create(&canvas).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(canvas)
}

// GetCanvas returns a single canvas by ID
// @Summary Get a canvas
// @Description Returns a single canvas by ID.
// @Tags Canvas
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Canvas ID"
// @Success 200 {object} models.Canvas
// @Failure 404 {object} map[string]interface{}
// @Router /canvas/{id} [get]
func GetCanvas(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("userID").(string)

	var canvas models.Canvas
	if err := database.DB.First(&canvas, "id = ?", id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Canvas not found"})
	}

	// Verify app ownership
	var app models.App
	if err := database.DB.Where("id = ? AND user_id = ?", canvas.AppID, userID).First(&app).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Access denied"})
	}

	return c.JSON(canvas)
}

// UpdateCanvasRequest defines payload for updating a canvas
type UpdateCanvasRequest struct {
	Name   *string      `json:"name" validate:"omitempty,min=3,max=100"`
	Config models.JSONB `json:"config"`
}

// UpdateCanvas updates a canvas
// @Summary Update a canvas
// @Description Updates an existing canvas.
// @Tags Canvas
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Canvas ID"
// @Param request body UpdateCanvasRequest true "Canvas Updates"
// @Success 200 {object} models.Canvas
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /canvas/{id} [put]
func UpdateCanvas(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("userID").(string)

	var canvas models.Canvas
	if err := database.DB.First(&canvas, "id = ?", id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Canvas not found"})
	}

	// Verify app ownership
	var app models.App
	if err := database.DB.Where("id = ? AND user_id = ?", canvas.AppID, userID).First(&app).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Access denied"})
	}

	var req UpdateCanvasRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := validator.GetValidator().ValidateStruct(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Explicit field mapping
	if req.Name != nil {
		canvas.Name = *req.Name
	}
	if req.Config != nil {
		canvas.Config = req.Config
	}

	canvas.UpdatedAt = time.Now()

	if err := database.DB.Save(&canvas).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(canvas)
}

// DeleteCanvas deletes a canvas (cascade delete widgets)
// @Summary Delete a canvas
// @Description Deletes a canvas by ID.
// @Tags Canvas
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Canvas ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /canvas/{id} [delete]
func DeleteCanvas(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("userID").(string)

	var canvas models.Canvas
	if err := database.DB.First(&canvas, "id = ?", id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Canvas not found"})
	}

	// Verify app ownership
	var app models.App
	if err := database.DB.Where("id = ? AND user_id = ?", canvas.AppID, userID).First(&app).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Access denied"})
	}

	// Delete all widgets (cascade)
	database.DB.Where("canvas_id = ?", id).Delete(&models.Widget{})

	// Delete canvas
	if err := database.DB.Delete(&canvas).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Canvas deleted successfully"})
}

// ==================== WIDGET HANDLERS ====================

// GetWidgets returns all widgets for a canvas
// @Summary List widgets
// @Description Returns all widgets for a specific canvas.
// @Tags Widget
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param canvasId query string true "Canvas ID"
// @Success 200 {object} []models.Widget
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /widget [get]
func GetWidgets(c *fiber.Ctx) error {
	canvasID := c.Query("canvasId")
	userID := c.Locals("userID").(string)

	if canvasID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "canvasId is required"})
	}

	// Verify canvas ownership (through app)
	var canvas models.Canvas
	if err := database.DB.First(&canvas, "id = ?", canvasID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Canvas not found"})
	}

	var app models.App
	if err := database.DB.Where("id = ? AND user_id = ?", canvas.AppID, userID).First(&app).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Access denied"})
	}

	var widgets []models.Widget
	if err := database.DB.Where("canvas_id = ?", canvasID).
		Order("created_at ASC").
		Find(&widgets).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(widgets)
}

// CreateWidgetRequest defines payload for creating a widget
type CreateWidgetRequest struct {
	CanvasID string       `json:"canvasId" validate:"required,uuid"`
	Type     string       `json:"type" validate:"required,oneof=chart table metric text image spacer"`
	Config   models.JSONB `json:"config"`
	Position models.JSONB `json:"position"`
}

// CreateWidget creates a new widget
// @Summary Create a widget
// @Description Creates a new widget on a canvas.
// @Tags Widget
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateWidgetRequest true "Widget Details"
// @Success 201 {object} models.Widget
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /widget [post]
func CreateWidget(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var req CreateWidgetRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := validator.GetValidator().ValidateStruct(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Verify canvas ownership (through app)
	var canvas models.Canvas
	if err := database.DB.First(&canvas, "id = ?", req.CanvasID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Canvas not found"})
	}

	var app models.App
	if err := database.DB.Where("id = ? AND user_id = ?", canvas.AppID, userID).First(&app).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Access denied"})
	}

	widget := models.Widget{
		ID:        uuid.New().String(),
		CanvasID:  req.CanvasID,
		Type:      req.Type,
		Config:    req.Config,
		Position:  req.Position,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := database.DB.Create(&widget).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(widget)
}

// UpdateWidgetRequest defines payload for updating a widget
type UpdateWidgetRequest struct {
	Type     *string      `json:"type" validate:"omitempty,oneof=chart table metric text image spacer"`
	Config   models.JSONB `json:"config"`
	Position models.JSONB `json:"position"`
}

// UpdateWidget updates a widget
// @Summary Update a widget
// @Description Updates an existing widget.
// @Tags Widget
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Widget ID"
// @Param request body UpdateWidgetRequest true "Widget Updates"
// @Success 200 {object} models.Widget
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /widget/{id} [put]
func UpdateWidget(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("userID").(string)

	var widget models.Widget
	if err := database.DB.First(&widget, "id = ?", id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Widget not found"})
	}

	// Verify canvas ownership (through app)
	var canvas models.Canvas
	if err := database.DB.First(&canvas, "id = ?", widget.CanvasID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Canvas not found"})
	}

	var app models.App
	if err := database.DB.Where("id = ? AND user_id = ?", canvas.AppID, userID).First(&app).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Access denied"})
	}

	var req UpdateWidgetRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := validator.GetValidator().ValidateStruct(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Explicit field mapping
	if req.Type != nil {
		widget.Type = *req.Type
	}
	if req.Config != nil {
		widget.Config = req.Config
	}
	if req.Position != nil {
		widget.Position = req.Position
	}

	widget.UpdatedAt = time.Now()

	if err := database.DB.Save(&widget).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(widget)
}

// DeleteWidget deletes a widget
// @Summary Delete a widget
// @Description Deletes a widget by ID.
// @Tags Widget
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Widget ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /widget/{id} [delete]
func DeleteWidget(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("userID").(string)

	var widget models.Widget
	if err := database.DB.First(&widget, "id = ?", id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Widget not found"})
	}

	// Verify canvas ownership (through app)
	var canvas models.Canvas
	if err := database.DB.First(&canvas, "id = ?", widget.CanvasID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Canvas not found"})
	}

	var app models.App
	if err := database.DB.Where("id = ? AND user_id = ?", canvas.AppID, userID).First(&app).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Access denied"})
	}

	if err := database.DB.Delete(&widget).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Widget deleted successfully"})
}
