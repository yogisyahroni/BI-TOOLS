package handlers

import (
	"fmt"
	"time"

	"insight-engine-backend/models"
	"insight-engine-backend/services"

	"github.com/gofiber/fiber/v2"
)

type ReportingHandler struct {
	service *services.ReportingService
}

func NewReportingHandler(service *services.ReportingService) *ReportingHandler {
	return &ReportingHandler{service: service}
}

func (h *ReportingHandler) GenerateReport(c *fiber.Ctx) error {
	var req models.ReportRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Title == "" {
		req.Title = "Exported Report"
	}

	buf, err := h.service.GenerateExcelReport(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Failed to generate report: %v", err)})
	}

	filename := fmt.Sprintf("report_%s.xlsx", time.Now().Format("20060102_150405"))

	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	return c.SendStream(buf)
}
