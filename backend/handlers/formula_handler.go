package handlers

import (
	"fmt"
	"insight-engine-backend/services/formula_engine"

	"github.com/gofiber/fiber/v2"
)

// FormulaHandler handles formula-related requests
type FormulaHandler struct {
	engine *formula_engine.FormulaEngine
}

// NewFormulaHandler creates a new FormulaHandler
func NewFormulaHandler(engine *formula_engine.FormulaEngine) *FormulaHandler {
	return &FormulaHandler{
		engine: engine,
	}
}

// ValidateRequest represents the request body for validation
type ValidateRequest struct {
	Formula string `json:"formula"`
}

// EvaluateRequest represents the request body for evaluation
type EvaluateRequest struct {
	Formula string                 `json:"formula"`
	Context map[string]interface{} `json:"context"` // Variables for the formula
}

// Validate checks if a formula is syntactically valid
func (h *FormulaHandler) Validate(c *fiber.Ctx) error {
	fmt.Println("FormulaHandler.Validate called")
	var req ValidateRequest
	if err := c.BodyParser(&req); err != nil {
		fmt.Println("BodyParser error:", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	fmt.Printf("Formula: %s\n", req.Formula)

	if req.Formula == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "formula is required"})
	}

	if err := h.engine.Validate(req.Formula); err != nil {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"valid": false,
			"error": err.Error(),
		})
	}

	// Extract references to show what columns/cells are used
	refs, _ := h.engine.ExtractReferences(req.Formula)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"valid":      true,
		"references": refs,
	})
}

// Evaluate evaluates a formula with provided context
func (h *FormulaHandler) Evaluate(c *fiber.Ctx) error {
	var req EvaluateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if req.Formula == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "formula is required"})
	}

	// Build context
	ctx := &formula_engine.FormulaContext{
		FieldValues: req.Context,
		CellValues:  make(map[string]interface{}), // Can be expanded if needed
	}

	result, err := h.engine.Evaluate(req.Formula, ctx)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"result": result,
	})
}

// RegisterRoutes registers the formula routes
func (h *FormulaHandler) RegisterRoutes(router fiber.Router) {
	formulas := router.Group("/formulas")
	formulas.Post("/validate", h.Validate)
	formulas.Post("/evaluate", h.Evaluate)
}
