package controllers

import (
	"insight-engine-backend/services"

	"github.com/gofiber/fiber/v2"
)

type LineageController struct {
	service *services.LineageService
}

func NewLineageController() *LineageController {
	return &LineageController{
		service: services.GetLineageService(),
	}
}

// GetLineage godoc
// @Summary Get Data Lineage Graph
// @Description Returns the graph of DataSources, Tables, Queries, and Dashboards
// @Tags lineage
// @Accept json
// @Produce json
// @Success 200 {object} services.LineageGraph
// @Router /lineage [get]
func (ctrl *LineageController) GetLineage(c *fiber.Ctx) error {
	graph, err := ctrl.service.GetLineageGraph()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(graph)
}
