package handlers

import (
	"encoding/json"
	"insight-engine-backend/database"
	"insight-engine-backend/models"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// DashboardHandler handles dashboard-related requests
type DashboardHandler struct{}

// NewDashboardHandler creates a new DashboardHandler
func NewDashboardHandler() *DashboardHandler {
	return &DashboardHandler{}
}

// GetDashboards retrieves all dashboards for the authenticated user
func (h *DashboardHandler) GetDashboards(c *fiber.Ctx) error {
	userID, _ := c.Locals("userId").(string)

	var dashboards []models.Dashboard
	result := database.DB.Where("user_id = ?", userID).
		Order("updated_at DESC").
		Find(&dashboards)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch dashboards",
			"error":   result.Error.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    dashboards,
		"count":   len(dashboards),
	})
}

// CreateDashboard creates a new dashboard
func (h *DashboardHandler) CreateDashboard(c *fiber.Ctx) error {
	log.Println("CreateDashboard: Handler started")
	userID, ok := c.Locals("userId").(string)
	log.Printf("CreateDashboard: UserID from locals: %v, ok: %v", userID, ok)
	if !ok {
		log.Println("CreateDashboard: UserID missing from locals")
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "User ID not found in context",
		})
	}

	type CreateDashboardRequest struct {
		Name         string  `json:"name"`
		Description  *string `json:"description"`
		CollectionID string  `json:"collectionId"`
		IsPublic     *bool   `json:"isPublic"`
	}

	req := new(CreateDashboardRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid input",
			"error":   err.Error(),
		})
	}

	// Validation
	if req.Name == "" || req.CollectionID == "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Dashboard name and collectionId are required",
		})
	}

	dashboard := models.Dashboard{
		ID:           uuid.New().String(),
		Name:         req.Name,
		Description:  req.Description,
		CollectionID: req.CollectionID,
		UserID:       userID,
		Layout:       &datatypes.JSON{}, // Default empty layout
		IsPublic:     false,
	}
	*dashboard.Layout = []byte("{}")

	if req.IsPublic != nil {
		dashboard.IsPublic = *req.IsPublic
	}

	if err := database.DB.Create(&dashboard).Error; err != nil {
		log.Printf("ERROR creating dashboard: %v\nDashboard Struct: %+v", err, dashboard) // Added logging
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to create dashboard",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    dashboard,
	})
}

// GetDashboard retrieves a single dashboard by ID
func (h *DashboardHandler) GetDashboard(c *fiber.Ctx) error {
	dashboardID := c.Params("id")
	userID, _ := c.Locals("userId").(string)

	var dashboard models.Dashboard
	result := database.DB.Where("id = ? AND user_id = ?", dashboardID, userID).
		Preload("Cards").
		First(&dashboard)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"success": false,
			"error":   "Dashboard not found",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    dashboard,
	})
}

// UpdateDashboard updates a dashboard (metadata or layout)
func (h *DashboardHandler) UpdateDashboard(c *fiber.Ctx) error {
	dashboardID := c.Params("id")
	userID, _ := c.Locals("userId").(string)

	// Check if dashboard exists and belongs to user
	var dashboard models.Dashboard
	if err := database.DB.Where("id = ? AND user_id = ?", dashboardID, userID).First(&dashboard).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"success": false,
			"error":   "Dashboard not found",
		})
	}

	// Parse request body
	var body map[string]interface{}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid input",
			"error":   err.Error(),
		})
	}

	// Check if this is a layout update (cards array present)
	if cards, hasCards := body["cards"].([]interface{}); hasCards {
		// Update metadata first if provided
		updates := make(map[string]interface{})
		if name, ok := body["name"].(string); ok {
			updates["name"] = name
		}
		if desc, ok := body["description"].(string); ok {
			updates["description"] = desc
		}
		if isPublic, ok := body["isPublic"].(bool); ok {
			updates["is_public"] = isPublic
		}
		if collID, ok := body["collectionId"].(string); ok {
			updates["collection_id"] = collID
		}
		if filters, ok := body["filters"]; ok {
			filtersJSON, _ := json.Marshal(filters)
			filtersStr := string(filtersJSON)
			updates["filters"] = filtersStr
		}

		if len(updates) > 0 {
			database.DB.Model(&dashboard).Updates(updates)
		}

		// Update layout (cards)
		// Delete existing cards
		database.DB.Where("dashboard_id = ?", dashboardID).Delete(&models.DashboardCard{})

		// Create new cards
		for _, cardData := range cards {
			cardMap := cardData.(map[string]interface{})

			var queryID *string
			if qid, ok := cardMap["queryId"].(string); ok && qid != "" {
				queryID = &qid
			}

			var title *string
			if t, ok := cardMap["title"].(string); ok && t != "" {
				title = &t
			}

			positionJSON, _ := json.Marshal(cardMap["position"])
			// positionStr := string(positionJSON) // Removed

			var vizConfigJSON datatypes.JSON
			if vizConfig, ok := cardMap["visualizationConfig"]; ok {
				vizBytes, _ := json.Marshal(vizConfig)
				vizConfigJSON = datatypes.JSON(vizBytes)
			}

			card := models.DashboardCard{
				DashboardID:         dashboardID,
				QueryID:             queryID,
				Title:               title,
				Position:            datatypes.JSON(positionJSON),
				VisualizationConfig: vizConfigJSON,
			}

			if cardID, ok := cardMap["id"].(string); ok && cardID != "" {
				card.ID = cardID
			}

			database.DB.Create(&card)
		}

		// Reload dashboard with cards
		database.DB.Where("id = ?", dashboardID).Preload("Cards").First(&dashboard)

	} else {
		// Metadata-only update
		updates := make(map[string]interface{})
		if name, ok := body["name"].(string); ok {
			updates["name"] = name
		}
		if desc, ok := body["description"].(string); ok {
			updates["description"] = desc
		}
		if isPublic, ok := body["isPublic"].(bool); ok {
			updates["is_public"] = isPublic
		}
		if collID, ok := body["collectionId"].(string); ok {
			updates["collection_id"] = collID
		}
		if filters, ok := body["filters"]; ok {
			filtersJSON, _ := json.Marshal(filters)
			filtersStr := string(filtersJSON)
			updates["filters"] = filtersStr
		}

		if err := database.DB.Model(&dashboard).Updates(updates).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to update dashboard",
				"error":   err.Error(),
			})
		}

		// Reload
		database.DB.Where("id = ?", dashboardID).First(&dashboard)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    dashboard,
	})
}

// DeleteDashboard deletes a dashboard and all its cards
func (h *DashboardHandler) DeleteDashboard(c *fiber.Ctx) error {
	dashboardID := c.Params("id")
	userID, _ := c.Locals("userId").(string)

	// Check if dashboard exists and belongs to user
	var dashboard models.Dashboard
	if err := database.DB.Where("id = ? AND user_id = ?", dashboardID, userID).First(&dashboard).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"success": false,
			"error":   "Dashboard not found",
		})
	}

	// Delete cards first (cascade)
	database.DB.Where("dashboard_id = ?", dashboardID).Delete(&models.DashboardCard{})

	// Delete dashboard
	if err := database.DB.Delete(&dashboard).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to delete dashboard",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Dashboard deleted successfully",
	})
}

// CertifyDashboard updates the certification status of a dashboard
func (h *DashboardHandler) CertifyDashboard(c *fiber.Ctx) error {
	dashboardID := c.Params("id")
	userID, _ := c.Locals("userId").(string)
	userRole, _ := c.Locals("userRole").(string) // Assuming middleware sets this

	// Only Admin or Editor can certify
	if userRole != "admin" && userRole != "editor" {
		// Fallback: Check if the user is the owner (optional, depending on requirements)
		// For now, let's strictly enforce role-based certification or allow owners to deprecate but not verify?
		// Let's stick to the plan: Admin or Editor.
		// If userRole is not available in Locals, we might need to fetch it.
		// Assuming AuthMiddleware puts it there. If not, we fetch user.
		var user models.User
		if err := database.DB.First(&user, "id = ?", userID).Error; err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "User not found"})
		}
		if user.Role != "admin" && user.Role != "editor" {
			return c.Status(403).JSON(fiber.Map{"error": "Unauthorized to certify dashboards"})
		}
	}

	type CertifyRequest struct {
		Status string `json:"status"` // "verified", "deprecated", "none"
	}

	var req CertifyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	if req.Status != "verified" && req.Status != "deprecated" && req.Status != "none" {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid status. Must be 'verified', 'deprecated', or 'none'"})
	}

	var dashboard models.Dashboard
	if err := database.DB.First(&dashboard, "id = ?", dashboardID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Dashboard not found"})
	}

	updates := map[string]interface{}{
		"certification_status": req.Status,
		"certified_by":         userID,
		"certified_at":         time.Now(),
	}

	if req.Status == "none" {
		updates["certified_by"] = nil
		updates["certified_at"] = nil
	}

	if err := database.DB.Model(&dashboard).Updates(updates).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update certification status"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    dashboard,
		"message": "Dashboard certification updated",
	})
}
