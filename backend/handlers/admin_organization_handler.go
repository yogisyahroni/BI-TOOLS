package handlers

import (
	"context"
	"strconv"
	"time"

	"insight-engine-backend/services"

	"github.com/gofiber/fiber/v2"
)

// AdminOrganizationHandler handles admin organization management
type AdminOrganizationHandler struct {
	orgService *services.OrganizationService
}

// NewAdminOrganizationHandler creates a new admin organization handler
func NewAdminOrganizationHandler(orgService *services.OrganizationService) *AdminOrganizationHandler {
	return &AdminOrganizationHandler{
		orgService: orgService,
	}
}

// RegisterRoutes registers all admin organization routes
func (h *AdminOrganizationHandler) RegisterRoutes(router fiber.Router, middlewares ...func(*fiber.Ctx) error) {
	org := router.Group("/admin/organizations")

	

	// Apply all provided middlewares (auth + admin check)

	for _, mw := range middlewares {

		org.Use(mw)

	}

	org.Get("/", h.ListOrganizations)
	org.Post("/", h.CreateOrganization)
	org.Get("/stats", h.GetStats)
	org.Get("/:id", h.GetOrganization)
	org.Put("/:id", h.UpdateOrganization)
	org.Delete("/:id", h.DeleteOrganization)

	// Member management
	org.Get("/:id/members", h.ListMembers)
	org.Post("/:id/members", h.AddMember)
	org.Delete("/:id/members/:userId", h.RemoveMember)
	org.Put("/:id/members/:userId/role", h.UpdateMemberRole)

	// Quota management
	org.Get("/:id/quotas", h.GetQuotas)
	org.Put("/:id/quotas", h.UpdateQuotas)
	org.Post("/:id/quotas/refresh", h.RefreshQuotaUsage)
}

// ListOrganizations handles GET /admin/organizations
func (h *AdminOrganizationHandler) ListOrganizations(c *fiber.Ctx) error {
	// Parse query parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(c.Query("pageSize", "20"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	search := c.Query("search")

	ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
	defer cancel()
	filter := &services.GetOrganizationsFilter{
		Search: search,
		Limit:  pageSize,
		Offset: (page - 1) * pageSize,
	}

	orgs, total, err := h.orgService.GetOrganizations(ctx, filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve organizations",
		})
	}

	// Calculate pagination metadata
	totalPages := (int(total) + pageSize - 1) / pageSize

	return c.JSON(fiber.Map{
		"data": orgs,
		"pagination": fiber.Map{
			"page":       page,
			"pageSize":   pageSize,
			"total":      total,
			"totalPages": totalPages,
		},
	})
}

// GetOrganization handles GET /admin/organizations/:id
func (h *AdminOrganizationHandler) GetOrganization(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Organization ID is required",
		})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
	defer cancel()
	org, err := h.orgService.GetOrganizationByID(ctx, id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Organization not found",
		})
	}

	return c.JSON(org)
}

// CreateOrganization handles POST /admin/organizations
func (h *AdminOrganizationHandler) CreateOrganization(c *fiber.Ctx) error {
	var req services.CreateOrganizationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Organization name is required",
		})
	}

	if req.OwnerID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Owner ID is required",
		})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
	defer cancel()
	org, err := h.orgService.CreateOrganization(ctx, &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create organization",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(org)
}

// UpdateOrganization handles PUT /admin/organizations/:id
func (h *AdminOrganizationHandler) UpdateOrganization(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Organization ID is required",
		})
	}

	var req services.UpdateOrganizationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
	defer cancel()
	org, err := h.orgService.UpdateOrganization(ctx, id, &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update organization",
		})
	}

	return c.JSON(org)
}

// DeleteOrganization handles DELETE /admin/organizations/:id
func (h *AdminOrganizationHandler) DeleteOrganization(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Organization ID is required",
		})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
	defer cancel()
	if err := h.orgService.DeleteOrganization(ctx, id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete organization",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ListMembers handles GET /admin/organizations/:id/members
func (h *AdminOrganizationHandler) ListMembers(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Organization ID is required",
		})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
	defer cancel()
	members, err := h.orgService.GetOrganizationMembers(ctx, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve organization statistics",
		})
	}

	return c.JSON(fiber.Map{
		"data": members,
	})
}

// AddMember handles POST /admin/organizations/:id/members
func (h *AdminOrganizationHandler) AddMember(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Organization ID is required",
		})
	}

	var req services.AddOrganizationMemberRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.UserID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User ID is required",
		})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
	defer cancel()
	member, err := h.orgService.AddOrganizationMember(ctx, id, &req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to add member",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(member)
}

// RemoveMember handles DELETE /admin/organizations/:id/members/:userId
func (h *AdminOrganizationHandler) RemoveMember(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Params("userId")

	if id == "" || userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Organization ID and User ID are required",
		})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
	defer cancel()
	if err := h.orgService.RemoveOrganizationMember(ctx, id, userID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to remove member",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// UpdateMemberRole handles PUT /admin/organizations/:id/members/:userId/role
func (h *AdminOrganizationHandler) UpdateMemberRole(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Params("userId")

	if id == "" || userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Organization ID and User ID are required",
		})
	}

	var req services.UpdateMemberRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
	defer cancel()
	member, err := h.orgService.UpdateMemberRole(ctx, id, userID, &req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to update member role",
		})
	}

	return c.JSON(member)
}

// GetQuotas handles GET /admin/organizations/:id/quotas
func (h *AdminOrganizationHandler) GetQuotas(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Organization ID is required",
		})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
	defer cancel()
	quotas, err := h.orgService.GetOrganizationQuota(ctx, id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Failed to retrieve organization quota",
		})
	}

	return c.JSON(quotas)
}

// UpdateQuotas handles PUT /admin/organizations/:id/quotas
func (h *AdminOrganizationHandler) UpdateQuotas(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Organization ID is required",
		})
	}

	var req services.UpdateOrganizationQuotaRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
	defer cancel()
	quotas, err := h.orgService.UpdateOrganizationQuota(ctx, id, &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve quotas",
		})
	}

	return c.JSON(quotas)
}

// RefreshQuotaUsage handles POST /admin/organizations/:id/quotas/refresh
func (h *AdminOrganizationHandler) RefreshQuotaUsage(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Organization ID is required",
		})
	}

	// This would call a service method to recalculate usage
	// For now, just return success
	return c.SendStatus(fiber.StatusNoContent)
}

// GetStats handles GET /admin/organizations/stats
func (h *AdminOrganizationHandler) GetStats(c *fiber.Ctx) error {
	// For now, return empty stats. This would be implemented to aggregate stats
	return c.JSON(fiber.Map{
		"totalOrganizations":  0,
		"totalMembers":        0,
		"activeOrganizations": 0,
	})
}


