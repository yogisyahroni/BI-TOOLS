package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"insight-engine-backend/models"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// AlertService handles alert operations
type AlertService struct {
	db                  *gorm.DB
	queryExecutor       *QueryExecutor
	notificationService *AlertNotificationService
}

// NewAlertService creates a new alert service
func NewAlertService(db *gorm.DB, queryExecutor *QueryExecutor, notificationService *AlertNotificationService) *AlertService {
	return &AlertService{
		db:                  db,
		queryExecutor:       queryExecutor,
		notificationService: notificationService,
	}
}

// CreateAlert creates a new alert
func (s *AlertService) CreateAlert(userID string, req *models.CreateAlertRequest) (*models.Alert, error) {
	// Validate query exists and user has access
	var query models.SavedQuery
	if err := s.db.Where("id = ?", req.QueryID).First(&query).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("query not found")
		}
		return nil, fmt.Errorf("failed to fetch query: %w", err)
	}

	// Set defaults
	severity := req.Severity
	if severity == "" {
		severity = models.AlertSeverityWarning
	}

	timezone := req.Timezone
	if timezone == "" {
		timezone = "UTC"
	}

	cooldown := req.CooldownMinutes
	if cooldown <= 0 {
		cooldown = 5 // Default 5 minutes
	}

	// Create alert
	alert := &models.Alert{
		ID:              uuid.New().String(),
		Name:            req.Name,
		Description:     req.Description,
		QueryID:         req.QueryID,
		UserID:          userID,
		Column:          req.Column,
		Operator:        req.Operator,
		Threshold:       req.Threshold,
		Schedule:        req.Schedule,
		Timezone:        timezone,
		Severity:        severity,
		CooldownMinutes: cooldown,
		IsActive:        true,
		State:           models.AlertStateOK,
	}

	// Calculate next run
	nextRun, err := s.CalculateNextRun(alert)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate next run: %w", err)
	}
	alert.NextRunAt = nextRun

	// Start transaction
	tx := s.db.Begin()

	// Create alert
	if err := tx.Create(alert).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create alert: %w", err)
	}

	// Create notification channels if provided
	if len(req.Channels) > 0 {
		for _, channelInput := range req.Channels {
			channel := &models.AlertNotificationChannelConfig{
				ID:          uuid.New(),
				AlertID:     alert.ID,
				ChannelType: channelInput.ChannelType,
				IsEnabled:   channelInput.IsEnabled,
			}
			if err := channel.SetConfig(channelInput.Config); err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to set channel config: %w", err)
			}
			if err := tx.Create(channel).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to create notification channel: %w", err)
			}
		}
	} else {
		// Create default email channel for backward compatibility
		defaultChannels := []models.AlertNotificationChannel{
			models.AlertChannelEmail,
		}
		if err := alert.SetChannels(defaultChannels); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to set default channels: %w", err)
		}
		tx.Save(alert)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Reload with relationships
	s.db.Preload("Query").Preload("ChannelsList").First(alert, "id = ?", alert.ID)

	return alert, nil
}

// GetAlert retrieves a single alert by ID
func (s *AlertService) GetAlert(alertID, userID string) (*models.Alert, error) {
	var alert models.Alert
	if err := s.db.Preload("Query").Preload("ChannelsList").Where("id = ? AND user_id = ?", alertID, userID).First(&alert).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("alert not found")
		}
		return nil, err
	}
	return &alert, nil
}

// UpdateAlert updates an existing alert
func (s *AlertService) UpdateAlert(alertID, userID string, req *models.UpdateAlertRequest) (*models.Alert, error) {
	// Get existing alert
	var alert models.Alert
	if err := s.db.Where("id = ? AND user_id = ?", alertID, userID).First(&alert).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("alert not found")
		}
		return nil, err
	}

	updates := map[string]interface{}{}

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Column != nil {
		updates["column"] = *req.Column
	}
	if req.Operator != nil {
		updates["operator"] = *req.Operator
	}
	if req.Threshold != nil {
		updates["threshold"] = *req.Threshold
	}
	if req.Schedule != nil {
		updates["schedule"] = *req.Schedule
	}
	if req.Timezone != nil {
		updates["timezone"] = *req.Timezone
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.Severity != nil {
		updates["severity"] = *req.Severity
	}
	if req.CooldownMinutes != nil {
		updates["cooldown_minutes"] = *req.CooldownMinutes
	}

	// Start transaction
	tx := s.db.Begin()

	// Update alert
	if len(updates) > 0 {
		if err := tx.Model(&alert).Updates(updates).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update alert: %w", err)
		}
	}

	// Update channels if provided
	if req.Channels != nil {
		// Delete existing channels
		if err := tx.Where("alert_id = ?", alertID).Delete(&models.AlertNotificationChannelConfig{}).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to delete existing channels: %w", err)
		}

		// Create new channels
		for _, channelInput := range *req.Channels {
			channel := &models.AlertNotificationChannelConfig{
				ID:          uuid.New(),
				AlertID:     alertID,
				ChannelType: channelInput.ChannelType,
				IsEnabled:   channelInput.IsEnabled,
			}
			if err := channel.SetConfig(channelInput.Config); err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to set channel config: %w", err)
			}
			if err := tx.Create(channel).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to create notification channel: %w", err)
			}
		}
	}

	// Recalculate next run if schedule or timezone changed
	if req.Schedule != nil || req.Timezone != nil {
		tx.First(&alert, "id = ?", alertID)
		nextRun, err := s.CalculateNextRun(&alert)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to recalculate next run: %w", err)
		}
		if err := tx.Model(&alert).Update("next_run_at", nextRun).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update next run: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Reload with relationships
	s.db.Preload("Query").Preload("ChannelsList").First(&alert, "id = ?", alertID)

	return &alert, nil
}

// DeleteAlert deletes an alert
func (s *AlertService) DeleteAlert(alertID, userID string) error {
	result := s.db.Where("id = ? AND user_id = ?", alertID, userID).Delete(&models.Alert{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("alert not found")
	}
	return nil
}

// ListAlerts retrieves alerts with filtering
func (s *AlertService) ListAlerts(filter *models.AlertFilter) (*models.AlertListResponse, error) {
	query := s.db.Model(&models.Alert{})

	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}
	if filter.QueryID != nil {
		query = query.Where("query_id = ?", *filter.QueryID)
	}
	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}
	if filter.State != nil {
		query = query.Where("state = ?", *filter.State)
	}
	if filter.Severity != nil {
		query = query.Where("severity = ?", *filter.Severity)
	}
	if filter.Search != nil && *filter.Search != "" {
		search := "%" + *filter.Search + "%"
		query = query.Where("name ILIKE ? OR description ILIKE ?", search, search)
	}

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// Apply pagination
	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	// Order by
	orderBy := filter.OrderBy
	if orderBy == "" {
		orderBy = "created_at DESC"
	}

	var alerts []models.Alert
	if err := query.Preload("Query").Preload("ChannelsList").Order(orderBy).Limit(limit).Offset(offset).Find(&alerts).Error; err != nil {
		return nil, err
	}

	return &models.AlertListResponse{
		Alerts: alerts,
		Total:  total,
		Page:   page,
		Limit:  limit,
	}, nil
}

// GetAlertHistory retrieves the check history for an alert
func (s *AlertService) GetAlertHistory(alertID, userID string, filter *models.AlertHistoryFilter) (*models.AlertHistoryListResponse, error) {
	// Verify ownership
	var alert models.Alert
	if err := s.db.Where("id = ? AND user_id = ?", alertID, userID).First(&alert).Error; err != nil {
		return nil, fmt.Errorf("alert not found")
	}

	query := s.db.Model(&models.AlertHistory{}).Where("alert_id = ?", alertID)

	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.StartDate != nil {
		query = query.Where("checked_at >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("checked_at <= ?", *filter.EndDate)
	}

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// Apply pagination
	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	// Order by
	orderBy := filter.OrderBy
	if orderBy == "" {
		orderBy = "checked_at DESC"
	}

	var history []models.AlertHistory
	if err := query.Order(orderBy).Limit(limit).Offset(offset).Find(&history).Error; err != nil {
		return nil, err
	}

	return &models.AlertHistoryListResponse{
		History: history,
		Total:   total,
		Page:    page,
		Limit:   limit,
	}, nil
}

// AcknowledgeAlert marks an alert as acknowledged
func (s *AlertService) AcknowledgeAlert(alertID, userID string, req *models.AcknowledgeAlertRequest) (*models.Alert, error) {
	var alert models.Alert
	if err := s.db.Where("id = ? AND user_id = ?", alertID, userID).First(&alert).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("alert not found")
		}
		return nil, err
	}

	// Update state
	alert.State = models.AlertStateAcknowledged

	// Create acknowledgment record
	ack := &models.AlertAcknowledgment{
		ID:      uuid.New(),
		AlertID: alertID,
		UserID:  userID,
		Note:    req.Note,
	}

	tx := s.db.Begin()

	if err := tx.Save(&alert).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update alert: %w", err)
	}

	if err := tx.Create(ack).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create acknowledgment: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &alert, nil
}

// MuteAlert mutes an alert temporarily or indefinitely
func (s *AlertService) MuteAlert(alertID, userID string, req *models.MuteAlertRequest) (*models.Alert, error) {
	var alert models.Alert
	if err := s.db.Where("id = ? AND user_id = ?", alertID, userID).First(&alert).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("alert not found")
		}
		return nil, err
	}

	alert.IsMuted = true
	alert.State = models.AlertStateMuted

	if req.Duration != nil && *req.Duration > 0 {
		mutedUntil := time.Now().Add(time.Duration(*req.Duration) * time.Minute)
		alert.MutedUntil = &mutedUntil
		alert.MuteDuration = req.Duration
	} else {
		alert.MutedUntil = nil
		alert.MuteDuration = nil
	}

	if err := s.db.Save(&alert).Error; err != nil {
		return nil, fmt.Errorf("failed to update alert: %w", err)
	}

	return &alert, nil
}

// UnmuteAlert unmutes an alert
func (s *AlertService) UnmuteAlert(alertID, userID string) (*models.Alert, error) {
	var alert models.Alert
	if err := s.db.Where("id = ? AND user_id = ?", alertID, userID).First(&alert).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("alert not found")
		}
		return nil, err
	}

	alert.IsMuted = false
	alert.MutedUntil = nil
	alert.MuteDuration = nil

	// Reset state based on last check
	if alert.LastStatus != nil && *alert.LastStatus == "TRIGGERED" {
		alert.State = models.AlertStateTriggered
	} else {
		alert.State = models.AlertStateOK
	}

	if err := s.db.Save(&alert).Error; err != nil {
		return nil, fmt.Errorf("failed to update alert: %w", err)
	}

	return &alert, nil
}

// GetTriggeredAlerts retrieves currently triggered alerts
func (s *AlertService) GetTriggeredAlerts(userID string) ([]models.TriggeredAlert, error) {
	var alerts []models.Alert
	query := s.db.Where("state = ? OR state = ?", models.AlertStateTriggered, models.AlertStateAcknowledged)
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.Preload("Query").Find(&alerts).Error; err != nil {
		return nil, err
	}

	triggeredAlerts := make([]models.TriggeredAlert, 0, len(alerts))
	for _, alert := range alerts {
		ta := models.TriggeredAlert{
			Alert: alert,
		}
		if alert.LastTriggeredAt != nil {
			ta.TriggeredAt = *alert.LastTriggeredAt
		}
		if alert.LastValue != nil {
			ta.CurrentValue = *alert.LastValue
		}

		// Check if acknowledged
		var ack models.AlertAcknowledgment
		if err := s.db.Where("alert_id = ?", alert.ID).Order("acknowledged_at DESC").First(&ack).Error; err == nil {
			ta.Acknowledged = true
			ta.AcknowledgedAt = &ack.AcknowledgedAt
			ta.AcknowledgedBy = &ack.UserID
		}

		triggeredAlerts = append(triggeredAlerts, ta)
	}

	return triggeredAlerts, nil
}

// GetAlertStats retrieves statistics for alerts
func (s *AlertService) GetAlertStats(userID string) (*models.AlertStats, error) {
	stats := &models.AlertStats{
		BySeverity: make(map[models.AlertSeverity]int64),
	}

	query := s.db.Model(&models.Alert{})
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	// Total count
	if err := query.Count(&stats.Total).Error; err != nil {
		return nil, err
	}

	// Active count
	if err := query.Where("is_active = ?", true).Count(&stats.Active).Error; err != nil {
		return nil, err
	}

	// By state
	var triggeredCount int64
	if err := query.Where("state = ?", models.AlertStateTriggered).Count(&triggeredCount).Error; err != nil {
		return nil, err
	}
	stats.Triggered = triggeredCount

	var acknowledgedCount int64
	if err := query.Where("state = ?", models.AlertStateAcknowledged).Count(&acknowledgedCount).Error; err != nil {
		return nil, err
	}
	stats.Acknowledged = acknowledgedCount

	var mutedCount int64
	if err := query.Where("state = ?", models.AlertStateMuted).Count(&mutedCount).Error; err != nil {
		return nil, err
	}
	stats.Muted = mutedCount

	var errorCount int64
	if err := query.Where("state = ?", models.AlertStateError).Count(&errorCount).Error; err != nil {
		return nil, err
	}
	stats.Error = errorCount

	// By severity
	for _, severity := range []models.AlertSeverity{models.AlertSeverityCritical, models.AlertSeverityWarning, models.AlertSeverityInfo} {
		var count int64
		if err := query.Where("severity = ?", severity).Count(&count).Error; err != nil {
			return nil, err
		}
		stats.BySeverity[severity] = count
	}

	return stats, nil
}

// ExecuteAlertCheck executes a single alert check
func (s *AlertService) ExecuteAlertCheck(ctx context.Context, alertID string) (*models.AlertHistory, error) {
	// Get alert with query
	var alert models.Alert
	if err := s.db.Preload("Query").Preload("ChannelsList").First(&alert, "id = ?", alertID).Error; err != nil {
		return nil, fmt.Errorf("alert not found: %w", err)
	}

	if !alert.IsActive {
		return nil, fmt.Errorf("alert is not active")
	}

	// Execute query
	startTime := time.Now()
	result, err := s.executeQueryForAlert(ctx, &alert)
	queryDuration := int(time.Since(startTime).Milliseconds())

	// Create history record
	history := &models.AlertHistory{
		ID:            uuid.New(),
		AlertID:       alertID,
		QueryDuration: queryDuration,
		CheckedAt:     time.Now(),
	}

	if err != nil {
		history.Status = "error"
		errMsg := err.Error()
		history.ErrorMessage = &errMsg
		history.Message = &errMsg

		// Update alert state
		alert.State = models.AlertStateError
		alert.LastError = &errMsg
		alert.LastRunAt = &history.CheckedAt
		s.db.Save(&alert)

		s.db.Create(history)
		return history, err
	}

	// Evaluate condition
	evalResult, err := s.EvaluateAlert(&alert, result)
	if err != nil {
		history.Status = "error"
		errMsg := err.Error()
		history.ErrorMessage = &errMsg
		history.Message = &errMsg

		alert.State = models.AlertStateError
		alert.LastError = &errMsg
		alert.LastRunAt = &history.CheckedAt
		s.db.Save(&alert)

		s.db.Create(history)
		return history, err
	}

	// Update history with evaluation result
	history.Value = &evalResult.Value
	history.Threshold = alert.Threshold

	if evalResult.Triggered {
		history.Status = "triggered"
		msg := fmt.Sprintf("Alert triggered: %s (value: %.2f, threshold: %.2f)", evalResult.Message, evalResult.Value, alert.Threshold)
		history.Message = &msg

		// Update alert state
		alert.State = models.AlertStateTriggered
		alert.LastStatus = strPtr("TRIGGERED")
		alert.LastValue = &evalResult.Value
		alert.LastTriggeredAt = &history.CheckedAt
		alert.TriggerCount++

		// Send notifications if not in cooldown and not muted
		if alert.CanSendNotification() {
			if err := s.sendNotifications(&alert, history, evalResult); err != nil {
				LogError("alert_notification", "Failed to send notifications", map[string]interface{}{
					"alert_id": alertID,
					"error":    err,
				})
			}
		}
	} else {
		history.Status = "ok"
		msg := fmt.Sprintf("Alert OK: value %.2f does not trigger condition", evalResult.Value)
		history.Message = &msg

		// Reset state if it was triggered
		if alert.State == models.AlertStateTriggered {
			alert.State = models.AlertStateOK
		}
		alert.LastStatus = strPtr("OK")
		alert.LastValue = &evalResult.Value
	}

	alert.LastRunAt = &history.CheckedAt
	alert.LastError = nil

	// Calculate next run
	nextRun, err := s.CalculateNextRun(&alert)
	if err == nil {
		alert.NextRunAt = nextRun
	}

	// Save alert and history
	tx := s.db.Begin()
	if err := tx.Save(&alert).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Create(history).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return history, nil
}

// executeQueryForAlert executes the query associated with an alert
func (s *AlertService) executeQueryForAlert(ctx context.Context, alert *models.Alert) (map[string]interface{}, error) {
	if alert.Query == nil {
		return nil, fmt.Errorf("query not found for alert")
	}

	// Get connection
	var conn models.Connection
	if err := s.db.Where("id = ?", alert.Query.ConnectionID).First(&conn).Error; err != nil {
		return nil, fmt.Errorf("connection not found: %w", err)
	}

	// Use the query executor to run the query
	limit := 1
	// Updated: include params (nil)
	result, err := s.queryExecutor.Execute(ctx, &conn, alert.Query.SQL, nil, &limit, nil)
	if err != nil {
		return nil, err
	}

	// Check for errors in result
	if result.Error != nil {
		return nil, fmt.Errorf("query execution failed: %s", *result.Error)
	}

	// Convert result to map
	if len(result.Rows) == 0 {
		return nil, fmt.Errorf("query returned no rows")
	}

	row := result.Rows[0]
	resultMap := make(map[string]interface{})
	for i, col := range result.Columns {
		if i < len(row) {
			resultMap[col] = row[i]
		}
	}

	return resultMap, nil
}

// EvaluateAlert evaluates if an alert condition is met
func (s *AlertService) EvaluateAlert(alert *models.Alert, result map[string]interface{}) (*models.AlertEvaluationResult, error) {
	value, ok := result[alert.Column]
	if !ok {
		return nil, fmt.Errorf("column %s not found in query result", alert.Column)
	}

	// Convert value to float64
	var numericValue float64
	switch v := value.(type) {
	case float64:
		numericValue = v
	case float32:
		numericValue = float64(v)
	case int:
		numericValue = float64(v)
	case int32:
		numericValue = float64(v)
	case int64:
		numericValue = float64(v)
	case string:
		// Try to parse string as number
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, fmt.Errorf("column %s value '%s' is not numeric", alert.Column, v)
		}
		numericValue = parsed
	case json.Number:
		parsed, err := v.Float64()
		if err != nil {
			return nil, fmt.Errorf("column %s value is not numeric", alert.Column)
		}
		numericValue = parsed
	default:
		return nil, fmt.Errorf("column %s is not numeric (type: %T)", alert.Column, value)
	}

	// Evaluate based on operator
	var triggered bool
	var message string

	switch alert.Operator {
	case ">":
		triggered = numericValue > alert.Threshold
		message = fmt.Sprintf("%s > %.2f", alert.Column, alert.Threshold)
	case "<":
		triggered = numericValue < alert.Threshold
		message = fmt.Sprintf("%s < %.2f", alert.Column, alert.Threshold)
	case "=", "==":
		triggered = numericValue == alert.Threshold
		message = fmt.Sprintf("%s = %.2f", alert.Column, alert.Threshold)
	case ">=":
		triggered = numericValue >= alert.Threshold
		message = fmt.Sprintf("%s >= %.2f", alert.Column, alert.Threshold)
	case "<=":
		triggered = numericValue <= alert.Threshold
		message = fmt.Sprintf("%s <= %.2f", alert.Column, alert.Threshold)
	case "!=":
		triggered = numericValue != alert.Threshold
		message = fmt.Sprintf("%s != %.2f", alert.Column, alert.Threshold)
	default:
		return nil, fmt.Errorf("unknown operator: %s", alert.Operator)
	}

	return &models.AlertEvaluationResult{
		Triggered: triggered,
		Value:     numericValue,
		Message:   message,
	}, nil
}

// TestAlert tests an alert configuration without saving
func (s *AlertService) TestAlert(ctx context.Context, userID string, req *models.TestAlertRequest) (*models.TestAlertResponse, error) {
	// Get query
	var query models.SavedQuery
	if err := s.db.Where("id = ?", req.QueryID).First(&query).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("query not found")
		}
		return nil, err
	}

	// Get connection
	var conn models.Connection
	if err := s.db.Where("id = ?", query.ConnectionID).First(&conn).Error; err != nil {
		return nil, fmt.Errorf("connection not found: %w", err)
	}

	// Execute query
	startTime := time.Now()
	limit := 1
	// Updated: include params (nil)
	result, err := s.queryExecutor.Execute(ctx, &conn, query.SQL, nil, &limit, nil)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	queryDuration := int(time.Since(startTime).Milliseconds())

	if result.Error != nil {
		return &models.TestAlertResponse{
			Triggered: false,
			Message:   fmt.Sprintf("Query error: %s", *result.Error),
			QueryTime: queryDuration,
		}, nil
	}

	if len(result.Rows) == 0 {
		return &models.TestAlertResponse{
			Triggered: false,
			Message:   "Query returned no rows",
			QueryTime: queryDuration,
		}, nil
	}

	// Convert result to map
	row := result.Rows[0]
	resultMap := make(map[string]interface{})
	for i, col := range result.Columns {
		if i < len(row) {
			resultMap[col] = row[i]
		}
	}

	// Create temporary alert for evaluation
	tempAlert := &models.Alert{
		Column:    req.Column,
		Operator:  req.Operator,
		Threshold: req.Threshold,
	}

	// Evaluate
	evalResult, err := s.EvaluateAlert(tempAlert, resultMap)
	if err != nil {
		return &models.TestAlertResponse{
			Triggered: false,
			Message:   fmt.Sprintf("Evaluation error: %s", err.Error()),
			QueryTime: queryDuration,
			Result:    resultMap,
		}, nil
	}

	return &models.TestAlertResponse{
		Triggered: evalResult.Triggered,
		Value:     evalResult.Value,
		Threshold: req.Threshold,
		Message:   evalResult.Message,
		QueryTime: queryDuration,
		Result:    resultMap,
	}, nil
}

// CalculateNextRun calculates the next run time for an alert
func (s *AlertService) CalculateNextRun(alert *models.Alert) (*time.Time, error) {
	loc, err := time.LoadLocation(alert.Timezone)
	if err != nil {
		loc = time.UTC
	}

	now := time.Now().In(loc)

	// Check if schedule is a cron expression
	if strings.Contains(alert.Schedule, " ") || strings.Contains(alert.Schedule, "*") {
		parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
		schedule, err := parser.Parse(alert.Schedule)
		if err != nil {
			return nil, fmt.Errorf("invalid cron expression: %w", err)
		}
		next := schedule.Next(now)
		return &next, nil
	}

	// Parse predefined schedules
	switch alert.Schedule {
	case "1m":
		next := now.Add(1 * time.Minute)
		return &next, nil
	case "5m":
		next := now.Add(5 * time.Minute)
		return &next, nil
	case "15m":
		next := now.Add(15 * time.Minute)
		return &next, nil
	case "30m":
		next := now.Add(30 * time.Minute)
		return &next, nil
	case "1h":
		next := now.Add(1 * time.Hour)
		return &next, nil
	default:
		// Try to parse as minutes
		if minutes, err := strconv.Atoi(alert.Schedule); err == nil && minutes > 0 {
			next := now.Add(time.Duration(minutes) * time.Minute)
			return &next, nil
		}
		return nil, fmt.Errorf("unsupported schedule format: %s", alert.Schedule)
	}
}

// ProcessAlerts processes all active alerts that are due to run
func (s *AlertService) ProcessAlerts(ctx context.Context) error {
	now := time.Now()

	var alerts []models.Alert
	if err := s.db.Where("is_active = ? AND (next_run_at IS NULL OR next_run_at <= ?)", true, now).Find(&alerts).Error; err != nil {
		return fmt.Errorf("failed to fetch due alerts: %w", err)
	}

	LogInfo("alert_processing", "Processing due alerts", map[string]interface{}{
		"count": len(alerts),
	})

	for _, alert := range alerts {
		if _, err := s.ExecuteAlertCheck(ctx, alert.ID); err != nil {
			LogError("alert_execution", "Failed to execute alert check", map[string]interface{}{
				"alert_id": alert.ID,
				"error":    err,
			})
		}
	}

	return nil
}

// sendNotifications sends notifications for a triggered alert
func (s *AlertService) sendNotifications(alert *models.Alert, history *models.AlertHistory, result *models.AlertEvaluationResult) error {
	if s.notificationService == nil {
		return fmt.Errorf("notification service not configured")
	}

	// Get channels
	channels := alert.ChannelsList
	if len(channels) == 0 {
		// Use legacy email configuration
		channels = []models.AlertNotificationChannelConfig{
			{
				ChannelType: models.AlertChannelEmail,
				IsEnabled:   true,
			},
		}
	}

	for _, channel := range channels {
		if !channel.IsEnabled {
			continue
		}

		err := s.notificationService.SendAlertNotification(alert, history, &channel, result)

		// Log notification
		log := &models.AlertNotificationLog{
			ID:          uuid.New(),
			HistoryID:   history.ID,
			ChannelType: channel.ChannelType,
			Status:      "sent",
			SentAt:      timePtr(time.Now()),
		}

		if err != nil {
			log.Status = "failed"
			errMsg := err.Error()
			log.Error = &errMsg
		}

		s.db.Create(log)
	}

	// Update notification count
	alert.NotificationCount += len(channels)
	s.db.Save(alert)

	return nil
}

// Helper functions
func strPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}

// Ensure AlertService implements the required interface
var _ AlertServiceInterface = (*AlertService)(nil)

// AlertServiceInterface defines the interface for alert service
type AlertServiceInterface interface {
	CreateAlert(userID string, req *models.CreateAlertRequest) (*models.Alert, error)
	GetAlert(alertID, userID string) (*models.Alert, error)
	UpdateAlert(alertID, userID string, req *models.UpdateAlertRequest) (*models.Alert, error)
	DeleteAlert(alertID, userID string) error
	ListAlerts(filter *models.AlertFilter) (*models.AlertListResponse, error)
	GetAlertHistory(alertID, userID string, filter *models.AlertHistoryFilter) (*models.AlertHistoryListResponse, error)
	AcknowledgeAlert(alertID, userID string, req *models.AcknowledgeAlertRequest) (*models.Alert, error)
	MuteAlert(alertID, userID string, req *models.MuteAlertRequest) (*models.Alert, error)
	UnmuteAlert(alertID, userID string) (*models.Alert, error)
	GetTriggeredAlerts(userID string) ([]models.TriggeredAlert, error)
	GetAlertStats(userID string) (*models.AlertStats, error)
	ExecuteAlertCheck(ctx context.Context, alertID string) (*models.AlertHistory, error)
	EvaluateAlert(alert *models.Alert, result map[string]interface{}) (*models.AlertEvaluationResult, error)
	TestAlert(ctx context.Context, userID string, req *models.TestAlertRequest) (*models.TestAlertResponse, error)
	CalculateNextRun(alert *models.Alert) (*time.Time, error)
	ProcessAlerts(ctx context.Context) error
}

// Global alert service instance (will be initialized by main)
var AlertServiceInstance *AlertService

// InitAlertService initializes the global alert service instance
func InitAlertService(db *gorm.DB, queryExecutor *QueryExecutor, notificationService *AlertNotificationService) {
	AlertServiceInstance = NewAlertService(db, queryExecutor, notificationService)
}

// GetAlertService returns the global alert service instance
func GetAlertService() *AlertService {
	return AlertServiceInstance
}
