package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"insight-engine-backend/models"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// ScheduledReportService handles scheduled report operations
type ScheduledReportService struct {
	db           *gorm.DB
	emailService *EmailService
	exportDir    string
	baseURL      string
}

// NewScheduledReportService creates a new scheduled report service
func NewScheduledReportService(db *gorm.DB, emailService *EmailService, exportDir, baseURL string) (*ScheduledReportService, error) {
	// Create export directory if it doesn't exist
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create export directory: %w", err)
	}

	return &ScheduledReportService{
		db:           db,
		emailService: emailService,
		exportDir:    exportDir,
		baseURL:      baseURL,
	}, nil
}

// CreateScheduledReport creates a new scheduled report
func (s *ScheduledReportService) CreateScheduledReport(userID string, req *models.CreateScheduledReportRequest) (*models.ScheduledReport, error) {
	// Validate schedule
	if err := s.validateSchedule(req.ScheduleType, req.CronExpr, req.TimeOfDay, req.DayOfWeek, req.DayOfMonth); err != nil {
		return nil, err
	}

	// Set default timezone
	timezone := req.Timezone
	if timezone == "" {
		timezone = "UTC"
	}

	// Create report
	report := &models.ScheduledReport{
		ID:             uuid.New(),
		Name:           req.Name,
		Description:    req.Description,
		ResourceType:   req.ResourceType,
		ResourceID:     req.ResourceID,
		ScheduleType:   req.ScheduleType,
		CronExpr:       req.CronExpr,
		TimeOfDay:      req.TimeOfDay,
		DayOfWeek:      req.DayOfWeek,
		DayOfMonth:     req.DayOfMonth,
		Timezone:       timezone,
		Format:         req.Format,
		IncludeFilters: req.IncludeFilters,
		Subject:        req.Subject,
		Message:        req.Message,
		IsActive:       true,
		CreatedBy:      userID,
	}

	// Set options
	if err := report.SetOptions(req.Options); err != nil {
		return nil, fmt.Errorf("failed to set options: %w", err)
	}

	// Calculate next run
	nextRun, err := s.CalculateNextRun(report)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate next run: %w", err)
	}
	report.NextRunAt = nextRun

	// Start transaction
	tx := s.db.Begin()

	// Create report
	if err := tx.Create(report).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create scheduled report: %w", err)
	}

	// Create recipients
	for _, recipient := range req.Recipients {
		r := &models.ScheduledReportRecipient{
			ID:       uuid.New(),
			ReportID: report.ID,
			Email:    recipient.Email,
			Type:     recipient.Type,
		}
		if r.Type == "" {
			r.Type = "to"
		}
		if err := tx.Create(r).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create recipient: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return report, nil
}

// GetScheduledReport retrieves a scheduled report by ID
func (s *ScheduledReportService) GetScheduledReport(reportID, userID string) (*models.ScheduledReportResponse, error) {
	var report models.ScheduledReport
	if err := s.db.Preload("Recipients").Where("id = ? AND created_by = ?", reportID, userID).First(&report).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("scheduled report not found")
		}
		return nil, err
	}

	return s.toReportResponse(&report), nil
}

// UpdateScheduledReport updates a scheduled report
func (s *ScheduledReportService) UpdateScheduledReport(reportID, userID string, req *models.UpdateScheduledReportRequest) (*models.ScheduledReport, error) {
	// Get existing report
	var report models.ScheduledReport
	if err := s.db.Where("id = ? AND created_by = ?", reportID, userID).First(&report).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("scheduled report not found")
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
	if req.ScheduleType != nil {
		updates["schedule_type"] = *req.ScheduleType
	}
	if req.CronExpr != nil {
		updates["cron_expr"] = *req.CronExpr
	}
	if req.TimeOfDay != nil {
		updates["time_of_day"] = *req.TimeOfDay
	}
	if req.DayOfWeek != nil {
		updates["day_of_week"] = *req.DayOfWeek
	}
	if req.DayOfMonth != nil {
		updates["day_of_month"] = *req.DayOfMonth
	}
	if req.Timezone != nil {
		updates["timezone"] = *req.Timezone
	}
	if req.Format != nil {
		updates["format"] = *req.Format
	}
	if req.IncludeFilters != nil {
		updates["include_filters"] = *req.IncludeFilters
	}
	if req.Subject != nil {
		updates["subject"] = *req.Subject
	}
	if req.Message != nil {
		updates["message"] = *req.Message
	}
	if req.Options != nil {
		if err := report.SetOptions(req.Options); err != nil {
			return nil, err
		}
		updates["options"] = report.Options
	}

	// Start transaction
	tx := s.db.Begin()

	// Update report
	if err := tx.Model(&report).Updates(updates).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update report: %w", err)
	}

	// Update recipients if provided
	if req.Recipients != nil {
		// Delete existing recipients
		if err := tx.Where("report_id = ?", reportID).Delete(&models.ScheduledReportRecipient{}).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to delete existing recipients: %w", err)
		}

		// Create new recipients
		for _, recipient := range *req.Recipients {
			r := &models.ScheduledReportRecipient{
				ID:       uuid.New(),
				ReportID: report.ID,
				Email:    recipient.Email,
				Type:     recipient.Type,
			}
			if r.Type == "" {
				r.Type = "to"
			}
			if err := tx.Create(r).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to create recipient: %w", err)
			}
		}
	}

	// Recalculate next run if schedule changed
	if req.ScheduleType != nil || req.CronExpr != nil || req.TimeOfDay != nil ||
		req.DayOfWeek != nil || req.DayOfMonth != nil || req.Timezone != nil {
		tx.First(&report, "id = ?", reportID)
		nextRun, err := s.CalculateNextRun(&report)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to recalculate next run: %w", err)
		}
		if err := tx.Model(&report).Update("next_run_at", nextRun).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update next run: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Reload report with recipients
	s.db.Preload("Recipients").First(&report, "id = ?", reportID)

	return &report, nil
}

// DeleteScheduledReport deletes a scheduled report
func (s *ScheduledReportService) DeleteScheduledReport(reportID, userID string) error {
	result := s.db.Where("id = ? AND created_by = ?", reportID, userID).Delete(&models.ScheduledReport{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("scheduled report not found")
	}
	return nil
}

// GetScheduledReports lists scheduled reports with filtering
func (s *ScheduledReportService) GetScheduledReports(userID string, filter *models.ScheduledReportFilter) (*models.ScheduledReportListResponse, error) {
	query := s.db.Model(&models.ScheduledReport{}).Where("created_by = ?", userID)

	if filter.ResourceType != nil {
		query = query.Where("resource_type = ?", *filter.ResourceType)
	}
	if filter.ResourceID != nil {
		query = query.Where("resource_id = ?", *filter.ResourceID)
	}
	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}
	if filter.ScheduleType != nil {
		query = query.Where("schedule_type = ?", *filter.ScheduleType)
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

	var reports []models.ScheduledReport
	if err := query.Preload("Recipients").Order(orderBy).Limit(limit).Offset(offset).Find(&reports).Error; err != nil {
		return nil, err
	}

	// Convert to response
	responseReports := make([]models.ScheduledReportResponse, len(reports))
	for i, report := range reports {
		responseReports[i] = *s.toReportResponse(&report)
	}

	return &models.ScheduledReportListResponse{
		Reports: responseReports,
		Total:   total,
		Page:    page,
		Limit:   limit,
	}, nil
}

// GetScheduledReportRuns retrieves run history for a report
func (s *ScheduledReportService) GetScheduledReportRuns(reportID, userID string, filter *models.ScheduledReportRunFilter) (*models.ScheduledReportRunListResponse, error) {
	// Verify ownership
	var report models.ScheduledReport
	if err := s.db.Where("id = ? AND created_by = ?", reportID, userID).First(&report).Error; err != nil {
		return nil, fmt.Errorf("scheduled report not found")
	}

	query := s.db.Model(&models.ScheduledReportRun{}).Where("report_id = ?", reportID)

	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.StartDate != nil {
		query = query.Where("started_at >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("started_at <= ?", *filter.EndDate)
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
		orderBy = "started_at DESC"
	}

	var runs []models.ScheduledReportRun
	if err := query.Order(orderBy).Limit(limit).Offset(offset).Find(&runs).Error; err != nil {
		return nil, err
	}

	// Convert to response
	responseRuns := make([]models.ScheduledReportRunResponse, len(runs))
	for i, run := range runs {
		responseRuns[i] = *s.toRunResponse(&run)
	}

	return &models.ScheduledReportRunListResponse{
		Runs:  responseRuns,
		Total: total,
		Page:  page,
		Limit: limit,
	}, nil
}

// ExecuteScheduledReport executes a scheduled report and sends email
func (s *ScheduledReportService) ExecuteScheduledReport(ctx context.Context, reportID, triggeredBy string) (*models.TriggerReportResponse, error) {
	// Get report with recipients
	var report models.ScheduledReport
	if err := s.db.Preload("Recipients").First(&report, "id = ?", reportID).Error; err != nil {
		return nil, fmt.Errorf("scheduled report not found: %w", err)
	}

	if !report.IsActive {
		return nil, fmt.Errorf("scheduled report is not active")
	}

	// Create run record
	run := &models.ScheduledReportRun{
		ID:          uuid.New(),
		ReportID:    report.ID,
		StartedAt:   time.Now(),
		Status:      models.ReportRunRunning,
		TriggeredBy: &triggeredBy,
	}

	if err := s.db.Create(run).Error; err != nil {
		return nil, fmt.Errorf("failed to create run record: %w", err)
	}

	// Generate report in background
	go s.generateAndSendReport(report, run, triggeredBy)

	return &models.TriggerReportResponse{
		RunID:     run.ID,
		Status:    run.Status,
		Message:   "Report generation started",
		StartedAt: run.StartedAt,
	}, nil
}

// generateAndSendReport generates the report file and sends email
func (s *ScheduledReportService) generateAndSendReport(report models.ScheduledReport, run *models.ScheduledReportRun, triggeredBy string) {
	startTime := time.Now()

	defer func() {
		// Update run record
		completedAt := time.Now()
		duration := completedAt.Sub(startTime).Milliseconds()
		run.CompletedAt = &completedAt
		run.DurationMs = &duration
		s.db.Save(run)

		// Update report stats
		if run.Status == models.ReportRunSuccess {
			s.db.Model(&report).Updates(map[string]interface{}{
				"last_run_at":      time.Now(),
				"last_run_status":  "success",
				"success_count":    gorm.Expr("success_count + 1"),
				"consecutive_fail": 0,
			})
		} else {
			s.db.Model(&report).Updates(map[string]interface{}{
				"last_run_at":      time.Now(),
				"last_run_status":  "failed",
				"last_run_error":   run.ErrorMessage,
				"failure_count":    gorm.Expr("failure_count + 1"),
				"consecutive_fail": gorm.Expr("consecutive_fail + 1"),
			})
		}

		// Calculate next run
		nextRun, err := s.CalculateNextRun(&report)
		if err == nil && nextRun != nil {
			s.db.Model(&report).Update("next_run_at", nextRun)
		}
	}()

	// Generate report file
	filePath, fileSize, fileType, err := s.GenerateReport(&report)
	if err != nil {
		run.Status = models.ReportRunFailed
		errMsg := err.Error()
		run.ErrorMessage = &errMsg
		LogError("scheduled_report_generate", "Failed to generate report", map[string]interface{}{
			"report_id": report.ID,
			"run_id":    run.ID,
			"error":     err,
		})
		return
	}

	run.FilePath = &filePath
	run.FileSize = &fileSize
	run.FileType = &fileType

	// Build download URL
	downloadURL := fmt.Sprintf("%s/api/scheduled-reports/runs/%s/download", s.baseURL, run.ID)
	run.FileURL = &downloadURL

	// Get all recipients
	var toRecipients, ccRecipients, bccRecipients []string
	for _, r := range report.Recipients {
		switch r.Type {
		case "cc":
			ccRecipients = append(ccRecipients, r.Email)
		case "bcc":
			bccRecipients = append(bccRecipients, r.Email)
		default:
			toRecipients = append(toRecipients, r.Email)
		}
	}

	if len(toRecipients) == 0 {
		run.Status = models.ReportRunFailed
		errMsg := "No recipients configured"
		run.ErrorMessage = &errMsg
		return
	}

	// Build email content
	subject := report.Subject
	if subject == "" {
		subject = fmt.Sprintf("[Scheduled Report] %s", report.Name)
	}

	message := report.Message
	if message == "" {
		message = fmt.Sprintf("Please find attached the scheduled report: %s", report.Name)
	}

	// Build HTML body
	bodyHTML := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #4F46E5; color: white; padding: 20px; border-radius: 8px 8px 0 0; }
        .content { background: #f9fafb; padding: 20px; border-radius: 0 0 8px 8px; }
        .footer { margin-top: 20px; font-size: 12px; color: #6b7280; }
        .button { display: inline-block; background: #4F46E5; color: white; padding: 10px 20px; text-decoration: none; border-radius: 6px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1 style="margin: 0; font-size: 20px;">%s</h1>
        </div>
        <div class="content">
            <p>%s</p>
            <p style="margin-top: 20px;">
                <a href="%s" class="button">Download Report</a>
            </p>
            <p style="font-size: 12px; color: #6b7280; margin-top: 20px;">
                This is an automated report from InsightEngine.<br>
                Generated at: %s
            </p>
        </div>
        <div class="footer">
            <p>&copy; 2026 InsightEngine. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`, report.Name, message, downloadURL, time.Now().Format("2006-01-02 15:04:05"))

	// Prepare attachment
	attachment := ReportAttachment{
		Filename:    fmt.Sprintf("%s.%s", report.Name, fileType),
		ContentType: s.emailService.GetContentType(filePath),
		FilePath:    filePath,
		FileSize:    fileSize,
	}

	// Send email using email service
	emailReq := &SendReportEmailRequest{
		To:          toRecipients,
		Cc:          ccRecipients,
		Bcc:         bccRecipients,
		Subject:     subject,
		BodyHTML:    bodyHTML,
		BodyText:    message,
		Attachments: []ReportAttachment{attachment},
		TrackOpens:  true,
	}

	if err := s.emailService.SendReportEmail(emailReq); err != nil {
		run.Status = models.ReportRunFailed
		errMsg := fmt.Sprintf("Failed to send email: %v", err)
		run.ErrorMessage = &errMsg
		LogError("scheduled_report_email", "Failed to send report email", map[string]interface{}{
			"report_id": report.ID,
			"run_id":    run.ID,
			"error":     err,
		})
		return
	}

	// Update run with success
	run.Status = models.ReportRunSuccess

	// Set sent to recipients
	sentTo := make([]string, 0, len(report.Recipients))
	for _, r := range report.Recipients {
		sentTo = append(sentTo, r.Email)
	}
	run.SetSentTo(sentTo)

	// Set per-recipient status
	sendStatus := make(map[string]string)
	for _, email := range sentTo {
		sendStatus[email] = "sent"
	}
	run.SetSendStatus(sendStatus)

	LogInfo("scheduled_report_success", "Report generated and sent successfully", map[string]interface{}{
		"report_id":  report.ID,
		"run_id":     run.ID,
		"recipients": len(sentTo),
	})
}

// GenerateReport generates the report file
func (s *ScheduledReportService) GenerateReport(report *models.ScheduledReport) (filePath string, fileSize int64, fileType string, err error) {
	// Generate filename
	runID := uuid.New().String()

	switch report.Format {
	case models.ReportFormatPDF:
		fileType = "pdf"
	case models.ReportFormatCSV:
		fileType = "csv"
	case models.ReportFormatExcel:
		fileType = "xlsx"
	case models.ReportFormatPNG:
		fileType = "png"
	default:
		fileType = "pdf"
	}

	filename := fmt.Sprintf("report_%s.%s", runID[:8], fileType)
	filePath = filepath.Join(s.exportDir, filename)

	// Generate report based on resource type
	switch report.ResourceType {
	case models.ReportResourceDashboard:
		return s.generateDashboardReport(report, filePath, fileType)
	case models.ReportResourceQuery:
		return s.generateQueryReport(report, filePath, fileType)
	default:
		return "", 0, "", fmt.Errorf("unsupported resource type: %s", report.ResourceType)
	}
}

// generateDashboardReport generates a report from a dashboard
func (s *ScheduledReportService) generateDashboardReport(report *models.ScheduledReport, filePath, fileType string) (string, int64, string, error) {
	// This is a placeholder implementation
	// In production, this would:
	// 1. Load the dashboard
	// 2. Execute all queries
	// 3. Generate the output in the requested format using the export service

	// For now, create a placeholder file
	content := fmt.Sprintf("Dashboard Report: %s\nResource ID: %s\nGenerated: %s\nFormat: %s",
		report.Name, report.ResourceID, time.Now().Format(time.RFC3339), fileType)

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return "", 0, "", fmt.Errorf("failed to write report file: %w", err)
	}

	// Get file size
	info, err := os.Stat(filePath)
	if err != nil {
		return "", 0, "", fmt.Errorf("failed to stat report file: %w", err)
	}

	return filePath, info.Size(), fileType, nil
}

// generateQueryReport generates a report from a query
func (s *ScheduledReportService) generateQueryReport(report *models.ScheduledReport, filePath, fileType string) (string, int64, string, error) {
	// This is a placeholder implementation
	// In production, this would:
	// 1. Load the query
	// 2. Execute the query
	// 3. Generate the output in the requested format

	// For now, create a placeholder file
	content := fmt.Sprintf("Query Report: %s\nResource ID: %s\nGenerated: %s\nFormat: %s",
		report.Name, report.ResourceID, time.Now().Format(time.RFC3339), fileType)

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return "", 0, "", fmt.Errorf("failed to write report file: %w", err)
	}

	// Get file size
	info, err := os.Stat(filePath)
	if err != nil {
		return "", 0, "", fmt.Errorf("failed to stat report file: %w", err)
	}

	return filePath, info.Size(), fileType, nil
}

// CalculateNextRun calculates the next run time for a scheduled report
func (s *ScheduledReportService) CalculateNextRun(report *models.ScheduledReport) (*time.Time, error) {
	// Load timezone
	loc, err := time.LoadLocation(report.Timezone)
	if err != nil {
		loc = time.UTC
	}

	now := time.Now().In(loc)

	switch report.ScheduleType {
	case models.ReportScheduleDaily:
		return s.calculateDailyNextRun(report, now, loc)
	case models.ReportScheduleWeekly:
		return s.calculateWeeklyNextRun(report, now, loc)
	case models.ReportScheduleMonthly:
		return s.calculateMonthlyNextRun(report, now, loc)
	case models.ReportScheduleCron:
		return s.calculateCronNextRun(report, now, loc)
	default:
		return nil, fmt.Errorf("unsupported schedule type: %s", report.ScheduleType)
	}
}

// calculateDailyNextRun calculates the next run for daily schedules
func (s *ScheduledReportService) calculateDailyNextRun(report *models.ScheduledReport, now time.Time, loc *time.Location) (*time.Time, error) {
	// Parse time of day
	hour, minute, err := s.parseTimeOfDay(report.TimeOfDay)
	if err != nil {
		return nil, err
	}

	// Calculate next run
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, loc)
	if !next.After(now) {
		next = next.AddDate(0, 0, 1)
	}

	return &next, nil
}

// calculateWeeklyNextRun calculates the next run for weekly schedules
func (s *ScheduledReportService) calculateWeeklyNextRun(report *models.ScheduledReport, now time.Time, loc *time.Location) (*time.Time, error) {
	if report.DayOfWeek == nil {
		return nil, fmt.Errorf("day_of_week is required for weekly schedules")
	}

	hour, minute, err := s.parseTimeOfDay(report.TimeOfDay)
	if err != nil {
		return nil, err
	}

	// Calculate next run
	targetDay := *report.DayOfWeek
	currentDay := int(now.Weekday())
	daysUntilTarget := targetDay - currentDay

	timeHasPassed := now.Hour() > hour || (now.Hour() == hour && now.Minute() >= minute)
	if daysUntilTarget < 0 || (daysUntilTarget == 0 && timeHasPassed) {
		daysUntilTarget += 7
	}

	next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, loc)
	next = next.AddDate(0, 0, daysUntilTarget)

	return &next, nil
}

// calculateMonthlyNextRun calculates the next run for monthly schedules
func (s *ScheduledReportService) calculateMonthlyNextRun(report *models.ScheduledReport, now time.Time, loc *time.Location) (*time.Time, error) {
	if report.DayOfMonth == nil {
		return nil, fmt.Errorf("day_of_month is required for monthly schedules")
	}

	hour, minute, err := s.parseTimeOfDay(report.TimeOfDay)
	if err != nil {
		return nil, err
	}

	targetDay := *report.DayOfMonth

	// Adjust for months with fewer days
	lastDayOfMonth := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, loc).Day()
	if targetDay > lastDayOfMonth {
		targetDay = lastDayOfMonth
	}

	next := time.Date(now.Year(), now.Month(), targetDay, hour, minute, 0, 0, loc)
	if !next.After(now) {
		// Move to next month
		next = time.Date(now.Year(), now.Month()+1, 1, hour, minute, 0, 0, loc)
		lastDayOfNextMonth := time.Date(next.Year(), next.Month()+1, 0, 0, 0, 0, 0, loc).Day()
		if targetDay > lastDayOfNextMonth {
			next = time.Date(next.Year(), next.Month(), lastDayOfNextMonth, hour, minute, 0, 0, loc)
		} else {
			next = time.Date(next.Year(), next.Month(), targetDay, hour, minute, 0, 0, loc)
		}
	}

	return &next, nil
}

// calculateCronNextRun calculates the next run for cron schedules
func (s *ScheduledReportService) calculateCronNextRun(report *models.ScheduledReport, now time.Time, loc *time.Location) (*time.Time, error) {
	if report.CronExpr == "" {
		return nil, fmt.Errorf("cron_expr is required for cron schedules")
	}

	// Parse cron expression
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(report.CronExpr)
	if err != nil {
		return nil, fmt.Errorf("invalid cron expression: %w", err)
	}

	next := schedule.Next(now)
	return &next, nil
}

// parseTimeOfDay parses time of day in HH:MM format
func (s *ScheduledReportService) parseTimeOfDay(timeStr string) (hour, minute int, err error) {
	if timeStr == "" {
		return 9, 0, nil // Default to 9:00 AM
	}

	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid time format, expected HH:MM")
	}

	hour, err = strconv.Atoi(parts[0])
	if err != nil || hour < 0 || hour > 23 {
		return 0, 0, fmt.Errorf("invalid hour: %s", parts[0])
	}

	minute, err = strconv.Atoi(parts[1])
	if err != nil || minute < 0 || minute > 59 {
		return 0, 0, fmt.Errorf("invalid minute: %s", parts[1])
	}

	return hour, minute, nil
}

// ProcessDueReports processes all reports that are due to run
func (s *ScheduledReportService) ProcessDueReports(ctx context.Context) error {
	now := time.Now()

	var reports []models.ScheduledReport
	if err := s.db.Preload("Recipients").
		Where("is_active = ? AND next_run_at <= ?", true, now).
		Find(&reports).Error; err != nil {
		return fmt.Errorf("failed to fetch due reports: %w", err)
	}

	LogInfo("scheduled_report_due", "Processing due reports", map[string]interface{}{
		"count": len(reports),
	})

	for _, report := range reports {
		if _, err := s.ExecuteScheduledReport(ctx, report.ID.String(), "schedule"); err != nil {
			LogError("scheduled_report_process", "Failed to execute scheduled report", map[string]interface{}{
				"report_id": report.ID,
				"error":     err,
			})
		}
	}

	return nil
}

// PreviewReport generates a preview of a report
func (s *ScheduledReportService) PreviewReport(userID string, req *models.ReportPreviewRequest) (*models.ReportPreviewResponse, error) {
	// Create a temporary scheduled report for preview
	tempReport := &models.ScheduledReport{
		ID:             uuid.New(),
		Name:           "Preview",
		ResourceType:   req.ResourceType,
		ResourceID:     req.ResourceID,
		Format:         req.Format,
		IncludeFilters: req.IncludeFilters,
		CreatedBy:      userID,
	}

	// Generate report
	filePath, fileSize, _, err := s.GenerateReport(tempReport)
	if err != nil {
		return nil, fmt.Errorf("failed to generate preview: %w", err)
	}

	// Generate preview URL (expires in 1 hour)
	previewID := uuid.New().String()
	previewURL := fmt.Sprintf("%s/api/scheduled-reports/preview/%s?file=%s", s.baseURL, previewID, filePath)

	return &models.ReportPreviewResponse{
		PreviewURL: previewURL,
		FileSize:   fileSize,
		ExpiresAt:  time.Now().Add(1 * time.Hour).Format(time.RFC3339),
	}, nil
}

// ToggleReportActive toggles the active status of a report
func (s *ScheduledReportService) ToggleReportActive(reportID, userID string) (*models.ScheduledReport, error) {
	var report models.ScheduledReport
	if err := s.db.Where("id = ? AND created_by = ?", reportID, userID).First(&report).Error; err != nil {
		return nil, fmt.Errorf("scheduled report not found")
	}

	report.IsActive = !report.IsActive

	// If activating, recalculate next run
	if report.IsActive {
		nextRun, err := s.CalculateNextRun(&report)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate next run: %w", err)
		}
		report.NextRunAt = nextRun
	} else {
		report.NextRunAt = nil
	}

	if err := s.db.Save(&report).Error; err != nil {
		return nil, fmt.Errorf("failed to update report: %w", err)
	}

	return &report, nil
}

// GetRunDownloadURL generates a download URL for a report run
func (s *ScheduledReportService) GetRunDownloadURL(runID, userID string) (string, error) {
	var run models.ScheduledReportRun
	if err := s.db.First(&run, "id = ?", runID).Error; err != nil {
		return "", fmt.Errorf("run not found")
	}

	// Verify ownership through report
	var report models.ScheduledReport
	if err := s.db.Where("id = ? AND created_by = ?", run.ReportID, userID).First(&report).Error; err != nil {
		return "", fmt.Errorf("access denied")
	}

	if run.Status != models.ReportRunSuccess || run.FilePath == nil {
		return "", fmt.Errorf("report not available for download")
	}

	// Check if file exists
	if _, err := os.Stat(*run.FilePath); os.IsNotExist(err) {
		return "", fmt.Errorf("report file has expired")
	}

	return fmt.Sprintf("%s/api/scheduled-reports/runs/%s/download", s.baseURL, runID), nil
}

// CleanupOldReports removes old report runs and files
func (s *ScheduledReportService) CleanupOldReports(olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)

	var runs []models.ScheduledReportRun
	if err := s.db.Where("created_at < ? AND status IN ?",
		cutoff, []models.ReportRunStatus{models.ReportRunSuccess, models.ReportRunFailed}).
		Find(&runs).Error; err != nil {
		return err
	}

	for _, run := range runs {
		// Delete file if exists
		if run.FilePath != nil {
			os.Remove(*run.FilePath)
		}

		// Delete run record
		s.db.Delete(&run)
	}

	LogInfo("scheduled_report_cleanup", "Cleaned up old report runs", map[string]interface{}{
		"deleted": len(runs),
	})

	return nil
}

// validateSchedule validates schedule configuration
func (s *ScheduledReportService) validateSchedule(scheduleType models.ReportScheduleType, cronExpr, timeOfDay string, dayOfWeek, dayOfMonth *int) error {
	switch scheduleType {
	case models.ReportScheduleDaily:
		if _, _, err := s.parseTimeOfDay(timeOfDay); err != nil {
			return fmt.Errorf("invalid time_of_day: %w", err)
		}

	case models.ReportScheduleWeekly:
		if dayOfWeek == nil {
			return fmt.Errorf("day_of_week is required for weekly schedules")
		}
		if *dayOfWeek < 0 || *dayOfWeek > 6 {
			return fmt.Errorf("day_of_week must be between 0 (Sunday) and 6 (Saturday)")
		}
		if _, _, err := s.parseTimeOfDay(timeOfDay); err != nil {
			return fmt.Errorf("invalid time_of_day: %w", err)
		}

	case models.ReportScheduleMonthly:
		if dayOfMonth == nil {
			return fmt.Errorf("day_of_month is required for monthly schedules")
		}
		if *dayOfMonth < 1 || *dayOfMonth > 31 {
			return fmt.Errorf("day_of_month must be between 1 and 31")
		}
		if _, _, err := s.parseTimeOfDay(timeOfDay); err != nil {
			return fmt.Errorf("invalid time_of_day: %w", err)
		}

	case models.ReportScheduleCron:
		if cronExpr == "" {
			return fmt.Errorf("cron_expr is required for cron schedules")
		}
		parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
		if _, err := parser.Parse(cronExpr); err != nil {
			return fmt.Errorf("invalid cron expression: %w", err)
		}

	default:
		return fmt.Errorf("unsupported schedule type: %s", scheduleType)
	}

	return nil
}

// toReportResponse converts a ScheduledReport to ScheduledReportResponse
func (s *ScheduledReportService) toReportResponse(report *models.ScheduledReport) *models.ScheduledReportResponse {
	recipients := make([]models.RecipientResponse, len(report.Recipients))
	for i, r := range report.Recipients {
		recipients[i] = models.RecipientResponse{
			ID:    r.ID,
			Email: r.Email,
			Type:  r.Type,
		}
	}

	return &models.ScheduledReportResponse{
		ScheduledReport: *report,
		Recipients:      recipients,
	}
}

// toRunResponse converts a ScheduledReportRun to ScheduledReportRunResponse
func (s *ScheduledReportService) toRunResponse(run *models.ScheduledReportRun) *models.ScheduledReportRunResponse {
	completedAt := ""
	if run.CompletedAt != nil {
		completedAt = run.CompletedAt.Format(time.RFC3339)
	}

	recipientCount := 0
	if sentTo, err := run.GetSentTo(); err == nil {
		recipientCount = len(sentTo)
	}

	return &models.ScheduledReportRunResponse{
		ID:             run.ID,
		ReportID:       run.ReportID,
		StartedAt:      run.StartedAt.Format(time.RFC3339),
		CompletedAt:    &completedAt,
		Status:         run.Status,
		ErrorMessage:   run.ErrorMessage,
		FileURL:        run.FileURL,
		FileSize:       run.FileSize,
		FileType:       run.FileType,
		DurationMs:     run.DurationMs,
		TriggeredBy:    run.TriggeredBy,
		RecipientCount: recipientCount,
		CreatedAt:      run.CreatedAt.Format(time.RFC3339),
	}
}
