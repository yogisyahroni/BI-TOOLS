package services

import (
	"context"
	"time"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// CronService manages scheduled tasks
type CronService struct {
	db                     *gorm.DB
	cron                   *cron.Cron
	scheduledReportService *ScheduledReportService
	alertService           *AlertService
}

// NewCronService creates a new cron service
func NewCronService(db *gorm.DB) *CronService {
	return &CronService{
		db:   db,
		cron: cron.New(),
	}
}

// NewCronServiceWithScheduledReports creates a new cron service with scheduled report support
func NewCronServiceWithScheduledReports(db *gorm.DB, scheduledReportService *ScheduledReportService) *CronService {
	return &CronService{
		db:                     db,
		cron:                   cron.New(),
		scheduledReportService: scheduledReportService,
	}
}

// NewCronServiceWithAlerts creates a new cron service with alert support
func NewCronServiceWithAlerts(db *gorm.DB, scheduledReportService *ScheduledReportService, alertService *AlertService) *CronService {
	return &CronService{
		db:                     db,
		cron:                   cron.New(),
		scheduledReportService: scheduledReportService,
		alertService:           alertService,
	}
}

// Start starts all cron jobs
func (s *CronService) Start() {
	// Budget reset - runs every hour
	_, err := s.cron.AddFunc("0 * * * *", func() {
		LogInfo("cron_budget_reset", "Starting budget reset job", nil)
		if err := s.resetBudgets(); err != nil {
			LogError("cron_budget_reset", "Budget reset failed", map[string]interface{}{"error": err})
		} else {
			LogInfo("cron_budget_reset", "Budget reset completed successfully", nil)
		}
	})
	if err != nil {
		LogError("cron_schedule", "Failed to schedule budget reset job", map[string]interface{}{"error": err})
	}

	// Materialized view refresh - runs every hour
	_, err = s.cron.AddFunc("0 * * * *", func() {
		LogInfo("cron_view_refresh", "Starting materialized view refresh job", nil)
		if err := s.refreshMaterializedViews(); err != nil {
			LogError("cron_view_refresh", "Materialized view refresh failed", map[string]interface{}{"error": err})
		} else {
			LogInfo("cron_view_refresh", "Materialized view refresh completed successfully", nil)
		}
	})
	if err != nil {
		LogError("cron_schedule", "Failed to schedule view refresh job", map[string]interface{}{"error": err})
	}

	// Expired shares cleanup - runs every hour
	_, err = s.cron.AddFunc("0 * * * *", func() {
		LogInfo("cron_shares_cleanup", "Starting expired shares cleanup job", nil)
		if count, err := s.cleanupExpiredShares(); err != nil {
			LogError("cron_shares_cleanup", "Expired shares cleanup failed", map[string]interface{}{"error": err})
		} else {
			LogInfo("cron_shares_cleanup", "Expired shares cleanup completed", map[string]interface{}{"count": count})
		}
	})
	if err != nil {
		LogError("cron_schedule", "Failed to schedule shares cleanup job", map[string]interface{}{"error": err})
	}

	// Expired embed tokens cleanup - runs every hour
	_, err = s.cron.AddFunc("0 * * * *", func() {
		LogInfo("cron_embed_tokens_cleanup", "Starting expired embed tokens cleanup job", nil)
		if count, err := s.cleanupExpiredEmbedTokens(); err != nil {
			LogError("cron_embed_tokens_cleanup", "Expired embed tokens cleanup failed", map[string]interface{}{"error": err})
		} else {
			LogInfo("cron_embed_tokens_cleanup", "Expired embed tokens cleanup completed", map[string]interface{}{"count": count})
		}
	})
	if err != nil {
		LogError("cron_schedule", "Failed to schedule embed tokens cleanup job", map[string]interface{}{"error": err})
	}

	// Scheduled Reports - runs every minute (TASK-099)
	_, err = s.cron.AddFunc("* * * * *", func() {
		LogInfo("cron_scheduled_reports", "Processing scheduled reports", nil)
		if s.scheduledReportService != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			if err := s.scheduledReportService.ProcessDueReports(ctx); err != nil {
				LogError("cron_scheduled_reports", "Failed to process scheduled reports", map[string]interface{}{"error": err})
			} else {
				LogInfo("cron_scheduled_reports", "Scheduled reports processing completed", nil)
			}
		}
	})
	if err != nil {
		LogError("cron_schedule", "Failed to schedule reports job", map[string]interface{}{"error": err})
	}

	// Email queue processing - runs every minute (TASK-098)
	_, err = s.cron.AddFunc("* * * * *", func() {
		LogInfo("cron_email_queue", "Processing email queue", nil)
		// Email service with DB connection would be needed here
		// This is a placeholder for the email queue processing
	})
	if err != nil {
		LogError("cron_schedule", "Failed to schedule email queue job", map[string]interface{}{"error": err})
	}

	// Alert checking - runs every minute (TASK-101)
	_, err = s.cron.AddFunc("* * * * *", func() {
		LogInfo("cron_alerts", "Processing alerts", nil)
		if s.alertService != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			if err := s.alertService.ProcessAlerts(ctx); err != nil {
				LogError("cron_alerts", "Failed to process alerts", map[string]interface{}{"error": err})
			} else {
				LogInfo("cron_alerts", "Alert processing completed", nil)
			}
		}
	})
	if err != nil {
		LogError("cron_schedule", "Failed to schedule alert checking job", map[string]interface{}{"error": err})
	}

	// Cleanup old report runs - runs daily at 3 AM
	_, err = s.cron.AddFunc("0 3 * * *", func() {
		LogInfo("cron_report_cleanup", "Cleaning up old report runs", nil)
		if s.scheduledReportService != nil {
			if err := s.scheduledReportService.CleanupOldReports(30 * 24 * time.Hour); err != nil {
				LogError("cron_report_cleanup", "Failed to cleanup old reports", map[string]interface{}{"error": err})
			} else {
				LogInfo("cron_report_cleanup", "Old reports cleanup completed", nil)
			}
		}
	})
	if err != nil {
		LogError("cron_schedule", "Failed to schedule report cleanup job", map[string]interface{}{"error": err})
	}

	s.cron.Start()
	LogInfo("cron_start", "Cron jobs started successfully", nil)
}

// Stop stops all cron jobs
func (s *CronService) Stop() {
	s.cron.Stop()
	LogInfo("cron_stop", "Cron jobs stopped", nil)
}

// resetBudgets calls the PostgreSQL function to reset budgets
func (s *CronService) resetBudgets() error {
	return s.db.Exec("SELECT reset_budgets()").Error
}

// refreshMaterializedViews refreshes all materialized views
func (s *CronService) refreshMaterializedViews() error {
	return s.db.Exec("SELECT refresh_ai_usage_stats()").Error
}

// RunBudgetResetNow manually triggers budget reset (for testing)
func (s *CronService) RunBudgetResetNow() error {
	LogInfo("cron_manual_budget_reset", "Manual budget reset triggered", nil)
	return s.resetBudgets()
}

// RunViewRefreshNow manually triggers view refresh (for testing)
func (s *CronService) RunViewRefreshNow() error {
	LogInfo("cron_manual_view_refresh", "Manual view refresh triggered", nil)
	return s.refreshMaterializedViews()
}

// cleanupExpiredShares marks expired shares as expired in the database
func (s *CronService) cleanupExpiredShares() (int64, error) {
	result := s.db.Exec(`
		UPDATE shares 
		SET status = 'expired', updated_at = NOW() 
		WHERE status = 'active' 
		AND expires_at IS NOT NULL 
		AND expires_at < NOW()
	`)
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

// cleanupExpiredEmbedTokens logs expired embed tokens (they remain in DB for audit)
func (s *CronService) cleanupExpiredEmbedTokens() (int64, error) {
	// Find expired tokens that haven't been logged yet
	var count int64
	err := s.db.Raw(`
		SELECT COUNT(*) 
		FROM embed_tokens 
		WHERE is_revoked = false 
		AND expires_at IS NOT NULL 
		AND expires_at < NOW()
	`).Scan(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

// RunSharesCleanupNow manually triggers expired shares cleanup (for testing)
func (s *CronService) RunSharesCleanupNow() (int64, error) {
	LogInfo("cron_manual_shares_cleanup", "Manual shares cleanup triggered", nil)
	return s.cleanupExpiredShares()
}

// RunEmbedTokensCleanupNow manually triggers embed tokens cleanup (for testing)
func (s *CronService) RunEmbedTokensCleanupNow() (int64, error) {
	LogInfo("cron_manual_embed_tokens_cleanup", "Manual embed tokens cleanup triggered", nil)
	return s.cleanupExpiredEmbedTokens()
}

// ProcessScheduledReportsNow manually triggers scheduled report processing (for testing)
func (s *CronService) ProcessScheduledReportsNow() error {
	LogInfo("cron_manual_scheduled_reports", "Manual scheduled reports processing triggered", nil)
	if s.scheduledReportService == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	return s.scheduledReportService.ProcessDueReports(ctx)
}

// ProcessAlertsNow manually triggers alert processing (for testing)
func (s *CronService) ProcessAlertsNow() error {
	LogInfo("cron_manual_alerts", "Manual alert processing triggered", nil)
	if s.alertService == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	return s.alertService.ProcessAlerts(ctx)
}
