package handlers

import (
	"encoding/json"
	"fmt"
	"insight-engine-backend/database"
	"insight-engine-backend/models"
	"insight-engine-backend/pkg/validator"
	"insight-engine-backend/services"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// GetPipelines returns all pipelines for a workspace (with optional pagination)
func GetPipelines(c *fiber.Ctx) error {
	userIDVal := c.Locals("userID")
	if userIDVal == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	userID, ok := userIDVal.(string)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid user session"})
	}

	workspaceID := c.Query("workspaceId")

	if workspaceID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "workspaceId query parameter required"})
	}

	// Verify workspace access
	var membership models.WorkspaceMember
	if err := database.DB.Where("workspace_id = ? AND user_id = ?", workspaceID, userID).First(&membership).Error; err != nil {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied to workspace"})
	}

	// Parse pagination params (0 = no pagination for backward compatibility)
	limit := c.QueryInt("limit", 0)
	offset := c.QueryInt("offset", 0)

	var pipelines []models.Pipeline
	query := database.DB.Where("workspace_id = ?", workspaceID).
		Order("created_at DESC") // Consistent ordering for pagination

	// Backward compatibility: If no pagination params, return old format
	if limit == 0 {
		if err := query.Find(&pipelines).Error; err != nil {
			fmt.Printf("Error fetching pipelines (no-limit): %v\n", err)
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(pipelines)
	}

	// Paginated response
	var total int64
	if err := database.DB.Model(&models.Pipeline{}).
		Where("workspace_id = ?", workspaceID).
		Count(&total).Error; err != nil {
		fmt.Printf("Error counting pipelines: %v\n", err)
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	if err := query.Limit(limit).Offset(offset).Find(&pipelines).Error; err != nil {
		fmt.Printf("Error fetching paginated pipelines: %v\n", err)
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"data": pipelines,
		"pagination": fiber.Map{
			"total":   total,
			"limit":   limit,
			"offset":  offset,
			"hasMore": offset+limit < int(total),
		},
	})
}

// GetPipeline returns a single pipeline by ID
func GetPipeline(c *fiber.Ctx) error {
	userIDVal := c.Locals("userID")
	if userIDVal == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	userID, ok := userIDVal.(string)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid user session"})
	}
	pipelineID := c.Params("id")

	var pipeline models.Pipeline
	if err := database.DB.Preload("QualityRules").First(&pipeline, "id = ?", pipelineID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Pipeline not found"})
	}

	// Verify workspace access
	var membership models.WorkspaceMember
	if err := database.DB.Where("workspace_id = ? AND user_id = ?", pipeline.WorkspaceID, userID).First(&membership).Error; err != nil {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	}

	return c.JSON(pipeline)
}

// CreatePipeline creates a new pipeline
func CreatePipeline(c *fiber.Ctx) error {
	userIDVal := c.Locals("userID")
	if userIDVal == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	userID, ok := userIDVal.(string)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid user session"})
	}

	var input struct {
		Name                string                 `json:"name" validate:"required"`
		Description         *string                `json:"description"`
		WorkspaceID         string                 `json:"workspaceId" validate:"required"`
		SourceType          string                 `json:"sourceType" validate:"required"`
		SourceConfig        map[string]interface{} `json:"sourceConfig" validate:"required"`
		ConnectionID        *string                `json:"connectionId"`
		SourceQuery         *string                `json:"sourceQuery"`
		DestinationType     string                 `json:"destinationType" validate:"required"`
		DestinationConfig   map[string]interface{} `json:"destinationConfig"`
		Mode                string                 `json:"mode" validate:"required,oneof=batch stream ETL ELT"`
		TransformationSteps []interface{}          `json:"transformationSteps"`
		QualityRules        []interface{}          `json:"qualityRules"`
		ScheduleCron        *string                `json:"scheduleCron"`
		RowLimit            int                    `json:"rowLimit"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := validator.GetValidator().ValidateStruct(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Verify workspace access (ADMIN, OWNER, EDITOR only)
	var membership models.WorkspaceMember
	if err := database.DB.Where("workspace_id = ? AND user_id = ?", input.WorkspaceID, userID).First(&membership).Error; err != nil {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied to workspace"})
	}

	if membership.Role != "ADMIN" && membership.Role != "OWNER" && membership.Role != "EDITOR" {
		return c.Status(403).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	// Convert maps to JSON strings
	sourceConfigJSON, err := json.Marshal(input.SourceConfig)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid source config"})
	}

	var destinationConfigStr *string
	if input.DestinationConfig != nil {
		destinationConfigJSON, err := json.Marshal(input.DestinationConfig)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid destination config"})
		}
		str := string(destinationConfigJSON)
		destinationConfigStr = &str
	}

	var transformationStepsStr *string
	if input.TransformationSteps != nil {
		transformationStepsJSON, err := json.Marshal(input.TransformationSteps)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid transformation steps"})
		}
		str := string(transformationStepsJSON)
		transformationStepsStr = &str
	}

	rowLimit := input.RowLimit
	if rowLimit <= 0 {
		rowLimit = 100000
	}

	pipeline := models.Pipeline{
		ID:                  uuid.New().String(),
		Name:                input.Name,
		Description:         input.Description,
		WorkspaceID:         input.WorkspaceID,
		SourceType:          input.SourceType,
		SourceConfig:        string(sourceConfigJSON),
		ConnectionID:        input.ConnectionID,
		SourceQuery:         input.SourceQuery,
		DestinationType:     input.DestinationType,
		DestinationConfig:   destinationConfigStr,
		Mode:                input.Mode,
		TransformationSteps: transformationStepsStr,
		ScheduleCron:        input.ScheduleCron,
		RowLimit:            rowLimit,
		IsActive:            true,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	if err := database.DB.Create(&pipeline).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(pipeline)
}

// UpdatePipeline updates an existing pipeline
func UpdatePipeline(c *fiber.Ctx) error {
	userIDVal := c.Locals("userID")
	if userIDVal == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	userID, ok := userIDVal.(string)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid user session"})
	}
	pipelineID := c.Params("id")

	var pipeline models.Pipeline
	if err := database.DB.First(&pipeline, "id = ?", pipelineID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Pipeline not found"})
	}

	// Verify workspace access (ADMIN, OWNER, EDITOR only)
	var membership models.WorkspaceMember
	if err := database.DB.Where("workspace_id = ? AND user_id = ?", pipeline.WorkspaceID, userID).First(&membership).Error; err != nil {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	}

	if membership.Role != "ADMIN" && membership.Role != "OWNER" && membership.Role != "EDITOR" {
		return c.Status(403).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	var input struct {
		Name                *string                `json:"name"`
		Description         *string                `json:"description"`
		SourceType          *string                `json:"sourceType"`
		SourceConfig        map[string]interface{} `json:"sourceConfig"`
		DestinationType     *string                `json:"destinationType"`
		DestinationConfig   map[string]interface{} `json:"destinationConfig"`
		Mode                *string                `json:"mode" validate:"omitempty,oneof=batch stream"`
		TransformationSteps []interface{}          `json:"transformationSteps"`
		ScheduleCron        *string                `json:"scheduleCron"`
		IsActive            *bool                  `json:"isActive"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := validator.GetValidator().ValidateStruct(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Update fields
	updates := make(map[string]interface{})
	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.Description != nil {
		updates["description"] = *input.Description
	}
	if input.SourceType != nil {
		updates["source_type"] = *input.SourceType
	}
	if input.SourceConfig != nil {
		sourceConfigJSON, _ := json.Marshal(input.SourceConfig)
		updates["source_config"] = string(sourceConfigJSON)
	}
	if input.DestinationType != nil {
		updates["destination_type"] = *input.DestinationType
	}
	if input.DestinationConfig != nil {
		destinationConfigJSON, _ := json.Marshal(input.DestinationConfig)
		updates["destination_config"] = string(destinationConfigJSON)
	}
	if input.Mode != nil {
		updates["mode"] = *input.Mode
	}
	if input.TransformationSteps != nil {
		transformationStepsJSON, _ := json.Marshal(input.TransformationSteps)
		updates["transformation_steps"] = string(transformationStepsJSON)
	}
	if input.ScheduleCron != nil {
		updates["schedule_cron"] = *input.ScheduleCron
	}
	if input.IsActive != nil {
		updates["is_active"] = *input.IsActive
	}
	updates["updated_at"] = time.Now()

	if err := database.DB.Model(&pipeline).Updates(updates).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Reload to get updated data
	database.DB.First(&pipeline, "id = ?", pipelineID)

	return c.JSON(pipeline)
}

// DeletePipeline deletes a pipeline
func DeletePipeline(c *fiber.Ctx) error {
	userIDVal := c.Locals("userID")
	if userIDVal == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	userID, ok := userIDVal.(string)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid user session"})
	}
	pipelineID := c.Params("id")

	var pipeline models.Pipeline
	if err := database.DB.First(&pipeline, "id = ?", pipelineID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Pipeline not found"})
	}

	// Verify workspace access (ADMIN, OWNER only)
	var membership models.WorkspaceMember
	if err := database.DB.Where("workspace_id = ? AND user_id = ?", pipeline.WorkspaceID, userID).First(&membership).Error; err != nil {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	}

	if membership.Role != "ADMIN" && membership.Role != "OWNER" {
		return c.Status(403).JSON(fiber.Map{"error": "Only workspace admins/owners can delete pipelines"})
	}

	// Delete pipeline (cascade will delete executions and quality rules)
	if err := database.DB.Delete(&pipeline).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.SendStatus(204)
}

// RunPipeline executes a pipeline (creates a job execution record)
func RunPipeline(c *fiber.Ctx) error {
	userIDVal := c.Locals("userID")
	if userIDVal == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	userID, ok := userIDVal.(string)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid user session"})
	}
	pipelineID := c.Params("id")

	var pipeline models.Pipeline
	if err := database.DB.First(&pipeline, "id = ?", pipelineID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Pipeline not found"})
	}

	// Verify workspace access
	var membership models.WorkspaceMember
	if err := database.DB.Where("workspace_id = ? AND user_id = ?", pipeline.WorkspaceID, userID).First(&membership).Error; err != nil {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	}

	// Create execution record
	execution := models.JobExecution{
		ID:            uuid.New().String(),
		PipelineID:    pipelineID,
		Status:        "PENDING",
		StartedAt:     time.Now(),
		RowsProcessed: 0,
	}

	if err := database.DB.Create(&execution).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Enqueue job for background processing
	services.GlobalJobQueue.Enqueue(services.Job{
		ID:        execution.ID,
		Type:      services.JobTypePipeline,
		EntityID:  pipelineID,
		CreatedAt: time.Now(),
		Retries:   0,
	})

	return c.Status(201).JSON(execution)
}

// GetPipelineStats returns pipeline statistics for a workspace
func GetPipelineStats(c *fiber.Ctx) error {
	userIDVal := c.Locals("userID")
	if userIDVal == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	userID, ok := userIDVal.(string)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid user session"})
	}
	workspaceID := c.Query("workspaceId")

	if workspaceID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "workspaceId query parameter required"})
	}

	// Verify workspace access
	var membership models.WorkspaceMember
	if err := database.DB.Where("workspace_id = ? AND user_id = ?", workspaceID, userID).First(&membership).Error; err != nil {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied to workspace"})
	}

	// 1. Pipeline Counts
	var totalPipelines int64
	database.DB.Model(&models.Pipeline{}).Where("workspace_id = ?", workspaceID).Count(&totalPipelines)

	var activePipelines int64
	database.DB.Model(&models.Pipeline{}).Where("workspace_id = ? AND is_active = ?", workspaceID, true).Count(&activePipelines)

	// 2. Recent Executions (Last 24h)
	oneDayAgo := time.Now().Add(-24 * time.Hour)

	var recentExecutions []models.JobExecution
	database.DB.Joins("JOIN \"Pipeline\" ON \"JobExecution\".\"pipelineId\" = \"Pipeline\".id").
		Where("\"Pipeline\".workspace_id = ? AND \"JobExecution\".started_at >= ?", workspaceID, oneDayAgo).
		Find(&recentExecutions)

	totalExecutions := len(recentExecutions)
	failedExecutions := 0
	successExecutions := 0
	totalRowsProcessed := 0

	for _, exec := range recentExecutions {
		if exec.Status == "FAILED" {
			failedExecutions++
		} else if exec.Status == "COMPLETED" {
			successExecutions++
		}
		totalRowsProcessed += exec.RowsProcessed
	}

	successRate := 0.0
	if totalExecutions > 0 {
		successRate = float64(successExecutions) / float64(totalExecutions) * 100
	}

	// 3. Recent Failures (Top 5)
	var recentFailures []struct {
		ID           string    `json:"id"`
		PipelineName string    `json:"pipelineName"`
		StartedAt    time.Time `json:"startedAt"`
		Error        *string   `json:"error"`
	}

	database.DB.Table("\"JobExecution\"").
		Select("\"JobExecution\".id, \"Pipeline\".name as pipeline_name, \"JobExecution\".started_at, \"JobExecution\".error").
		Joins("JOIN \"Pipeline\" ON \"JobExecution\".\"pipelineId\" = \"Pipeline\".id").
		Where("\"Pipeline\".workspace_id = ? AND \"JobExecution\".status = ?", workspaceID, "FAILED").
		Order("\"JobExecution\".started_at DESC").
		Limit(5).
		Scan(&recentFailures)

	// 4. All Pipelines for Heatmap
	var allPipelines []struct {
		ID         string     `json:"id"`
		Name       string     `json:"name"`
		LastStatus *string    `json:"lastStatus"`
		LastRunAt  *time.Time `json:"lastRunAt"`
	}

	database.DB.Table("\"Pipeline\"").
		Select("id, name, last_status, last_run_at").
		Where("workspace_id = ?", workspaceID).
		Order("last_run_at DESC").
		Scan(&allPipelines)

	return c.JSON(fiber.Map{
		"overview": fiber.Map{
			"totalPipelines":     totalPipelines,
			"activePipelines":    activePipelines,
			"successRate":        successRate,
			"totalRowsProcessed": totalRowsProcessed,
			"totalExecutions":    totalExecutions,
		},
		"pipelines":      allPipelines,
		"recentFailures": recentFailures,
	})
}

// GetPipelineExecutions returns execution history for a specific pipeline with per-pipeline success rate
func GetPipelineExecutions(c *fiber.Ctx) error {
	userIDVal := c.Locals("userID")
	if userIDVal == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	userID, ok := userIDVal.(string)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid user session"})
	}
	pipelineID := c.Params("id")
	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)

	// Load pipeline and verify workspace access
	var pipeline models.Pipeline
	if err := database.DB.First(&pipeline, "id = ?", pipelineID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Pipeline not found"})
	}

	var membership models.WorkspaceMember
	if err := database.DB.Where("workspace_id = ? AND user_id = ?", pipeline.WorkspaceID, userID).First(&membership).Error; err != nil {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	}

	// Get total count
	var total int64
	database.DB.Model(&models.JobExecution{}).Where("\"pipelineId\" = ?", pipelineID).Count(&total)

	// Get paginated executions
	var executions []models.JobExecution
	database.DB.Where("\"pipelineId\" = ?", pipelineID).
		Order("started_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&executions)

	// Calculate per-pipeline success rate (from ALL executions, not just this page)
	var allCount int64
	var successCount int64
	database.DB.Model(&models.JobExecution{}).Where("\"pipelineId\" = ?", pipelineID).Count(&allCount)
	database.DB.Model(&models.JobExecution{}).Where("\"pipelineId\" = ? AND status = ?", pipelineID, "COMPLETED").Count(&successCount)

	successRate := 0.0
	if allCount > 0 {
		successRate = float64(successCount) / float64(allCount) * 100
	}

	return c.JSON(fiber.Map{
		"executions":   executions,
		"total":        total,
		"limit":        limit,
		"offset":       offset,
		"successRate":  successRate,
		"successCount": successCount,
		"failedCount":  allCount - successCount,
	})
}
