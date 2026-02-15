package handlers

import (
	"insight-engine-backend/models"
	"insight-engine-backend/services"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type DataGovernanceHandler struct {
	service *services.DataGovernanceService
}

func NewDataGovernanceHandler(service *services.DataGovernanceService) *DataGovernanceHandler {
	return &DataGovernanceHandler{service: service}
}

// GetClassifications returns all data classifications
func (h *DataGovernanceHandler) GetClassifications(c *fiber.Ctx) error {
	classifications, err := h.service.GetDataClassifications()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(classifications)
}

// GetColumnMetadata returns metadata for a specific datasource and table
func (h *DataGovernanceHandler) GetColumnMetadata(c *fiber.Ctx) error {
	datasourceID := c.Query("datasource_id")
	tableName := c.Query("table_name")

	if datasourceID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "datasource_id is required"})
	}

	metadata, err := h.service.GetColumnMetadata(datasourceID, tableName)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(metadata)
}

// UpdateColumnMetadata updates or creates metadata for a column
func (h *DataGovernanceHandler) UpdateColumnMetadata(c *fiber.Ctx) error {
	var body map[string]interface{}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	// Validate required fields
	datasourceID, _ := body["datasource_id"].(string)
	tableName, _ := body["table_name"].(string)
	columnName, _ := body["column_name"].(string)

	if datasourceID == "" || tableName == "" || columnName == "" {
		return c.Status(400).JSON(fiber.Map{"error": "datasource_id, table_name, and column_name are required"})
	}

	if err := h.service.UpdateColumnMetadata(datasourceID, tableName, columnName, body); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "data": body})
}

// GetColumnPermissions returns permissions for a specific role
func (h *DataGovernanceHandler) GetColumnPermissions(c *fiber.Ctx) error {
	roleIDStr := c.Query("role_id")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid role_id"})
	}

	permissions, err := h.service.GetColumnPermissions(uint(roleID))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(permissions)
}

// SetColumnPermission sets permission for a column and role
func (h *DataGovernanceHandler) SetColumnPermission(c *fiber.Ctx) error {
	var perm models.ColumnPermission
	if err := c.BodyParser(&perm); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	if err := h.service.SetColumnPermission(&perm); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true})
}
