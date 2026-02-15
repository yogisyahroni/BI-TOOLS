package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// ReportScheduleType represents the type of schedule
type ReportScheduleType string

const (
	ReportScheduleDaily   ReportScheduleType = "daily"
	ReportScheduleWeekly  ReportScheduleType = "weekly"
	ReportScheduleMonthly ReportScheduleType = "monthly"
	ReportScheduleCron    ReportScheduleType = "cron"
)

// ReportFormat represents the format of the report
type ReportFormat string

const (
	ReportFormatPDF   ReportFormat = "pdf"
	ReportFormatCSV   ReportFormat = "csv"
	ReportFormatExcel ReportFormat = "excel"
	ReportFormatPNG   ReportFormat = "png"
)

// ReportResourceType represents the type of resource being reported
type ReportResourceType string

const (
	ReportResourceDashboard ReportResourceType = "dashboard"
	ReportResourceQuery     ReportResourceType = "query"
)

// ReportRunStatus represents the status of a report run
type ReportRunStatus string

const (
	ReportRunPending   ReportRunStatus = "pending"
	ReportRunRunning   ReportRunStatus = "running"
	ReportRunSuccess   ReportRunStatus = "success"
	ReportRunFailed    ReportRunStatus = "failed"
	ReportRunCancelled ReportRunStatus = "cancelled"
)

// ScheduledReport represents a scheduled report configuration
type ScheduledReport struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(255);not null" json:"name"`
	Description *string   `gorm:"type:text" json:"description,omitempty"`

	// Resource
	ResourceType ReportResourceType `gorm:"type:varchar(50);not null;index" json:"resourceType"`
	ResourceID   string             `gorm:"type:varchar(255);not null;index" json:"resourceId"`

	// Schedule
	ScheduleType ReportScheduleType `gorm:"type:varchar(50);not null" json:"scheduleType"`
	CronExpr     string             `gorm:"type:varchar(255)" json:"cronExpr,omitempty"`     // For custom cron
	TimeOfDay    string             `gorm:"type:varchar(10)" json:"timeOfDay,omitempty"`     // HH:MM format
	DayOfWeek    *int               `json:"dayOfWeek,omitempty"`                             // 0-6 for weekly
	DayOfMonth   *int               `json:"dayOfMonth,omitempty"`                            // 1-31 for monthly
	Timezone     string             `gorm:"type:varchar(100);default:'UTC'" json:"timezone"` // e.g., "America/New_York"

	// Recipients
	Recipients []ScheduledReportRecipient `gorm:"foreignKey:ReportID" json:"recipients"`

	// Format & Options
	Format         ReportFormat `gorm:"type:varchar(20);not null" json:"format"`
	IncludeFilters bool         `gorm:"default:false" json:"includeFilters"`
	Subject        string       `gorm:"type:text" json:"subject"`
	Message        string       `gorm:"type:text" json:"message"`

	// Status
	IsActive        bool       `gorm:"default:true;index" json:"isActive"`
	LastRunAt       *time.Time `json:"lastRunAt,omitempty"`
	LastRunStatus   *string    `gorm:"type:varchar(20)" json:"lastRunStatus,omitempty"` // "success", "failed"
	LastRunError    *string    `gorm:"type:text" json:"lastRunError,omitempty"`
	NextRunAt       *time.Time `gorm:"index" json:"nextRunAt,omitempty"`
	SuccessCount    int        `gorm:"default:0" json:"successCount"`
	FailureCount    int        `gorm:"default:0" json:"failureCount"`
	ConsecutiveFail int        `gorm:"default:0" json:"consecutiveFail"`

	// Additional options stored as JSON
	Options datatypes.JSON `gorm:"type:jsonb" json:"options,omitempty"`

	// Ownership
	CreatedBy string    `gorm:"type:varchar(255);not null;index" json:"createdBy"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	// Relationships
	Runs []ScheduledReportRun `gorm:"foreignKey:ReportID" json:"runs,omitempty"`
}

// TableName specifies the table name for ScheduledReport
func (ScheduledReport) TableName() string {
	return "scheduled_reports"
}

// GetOptions parses and returns the options
func (r *ScheduledReport) GetOptions() (map[string]interface{}, error) {
	if r.Options == nil {
		return map[string]interface{}{}, nil
	}

	var options map[string]interface{}
	if err := json.Unmarshal(r.Options, &options); err != nil {
		return nil, err
	}
	return options, nil
}

// SetOptions serializes options to JSON
func (r *ScheduledReport) SetOptions(options map[string]interface{}) error {
	if options == nil || len(options) == 0 {
		r.Options = nil
		return nil
	}

	data, err := json.Marshal(options)
	if err != nil {
		return err
	}
	r.Options = datatypes.JSON(data)
	return nil
}

// ScheduledReportRecipient represents a recipient of a scheduled report
type ScheduledReportRecipient struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ReportID  uuid.UUID `gorm:"type:varchar(255);not null;index" json:"reportId"`
	Email     string    `gorm:"type:varchar(255);not null" json:"email"`
	Type      string    `gorm:"type:varchar(10);default:'to'" json:"type"` // "to", "cc", "bcc"
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

// TableName specifies the table name for ScheduledReportRecipient
func (ScheduledReportRecipient) TableName() string {
	return "scheduled_report_recipients"
}

// ScheduledReportRun represents a single run of a scheduled report
type ScheduledReportRun struct {
	ID           uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	ReportID     uuid.UUID       `gorm:"type:varchar(255);not null;index" json:"reportId"`
	StartedAt    time.Time       `json:"startedAt"`
	CompletedAt  *time.Time      `json:"completedAt,omitempty"`
	Status       ReportRunStatus `gorm:"type:varchar(20);not null;index" json:"status"`
	ErrorMessage *string         `gorm:"type:text" json:"errorMessage,omitempty"`
	FileURL      *string         `gorm:"type:text" json:"fileUrl,omitempty"`
	FilePath     *string         `gorm:"type:text" json:"-"`
	FileSize     *int64          `json:"fileSize,omitempty"`
	FileType     *string         `gorm:"type:varchar(50)" json:"fileType,omitempty"`
	SentTo       datatypes.JSON  `gorm:"type:jsonb" json:"-"`                    // Array of recipient emails
	SendStatus   datatypes.JSON  `gorm:"type:jsonb" json:"sendStatus,omitempty"` // Per-recipient status
	DurationMs   *int64          `json:"durationMs,omitempty"`                   // Execution time in milliseconds

	// Metadata
	TriggeredBy *string `gorm:"type:varchar(255)" json:"triggeredBy,omitempty"` // "schedule", "manual", "api"
	IPAddress   *string `gorm:"type:varchar(45)" json:"ipAddress,omitempty"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	// Relationships
	Report *ScheduledReport `gorm:"foreignKey:ReportID" json:"report,omitempty"`
}

// TableName specifies the table name for ScheduledReportRun
func (ScheduledReportRun) TableName() string {
	return "scheduled_report_runs"
}

// GetSentTo returns the list of recipients the report was sent to
func (r *ScheduledReportRun) GetSentTo() ([]string, error) {
	if r.SentTo == nil {
		return []string{}, nil
	}

	var sentTo []string
	if err := json.Unmarshal(r.SentTo, &sentTo); err != nil {
		return nil, err
	}
	return sentTo, nil
}

// SetSentTo sets the list of recipients
func (r *ScheduledReportRun) SetSentTo(sentTo []string) error {
	if len(sentTo) == 0 {
		r.SentTo = nil
		return nil
	}

	data, err := json.Marshal(sentTo)
	if err != nil {
		return err
	}
	r.SentTo = datatypes.JSON(data)
	return nil
}

// GetSendStatus returns the per-recipient send status
func (r *ScheduledReportRun) GetSendStatus() (map[string]string, error) {
	if r.SendStatus == nil {
		return map[string]string{}, nil
	}

	var status map[string]string
	if err := json.Unmarshal(r.SendStatus, &status); err != nil {
		return nil, err
	}
	return status, nil
}

// SetSendStatus sets the per-recipient send status
func (r *ScheduledReportRun) SetSendStatus(status map[string]string) error {
	if status == nil || len(status) == 0 {
		r.SendStatus = nil
		return nil
	}

	data, err := json.Marshal(status)
	if err != nil {
		return err
	}
	r.SendStatus = datatypes.JSON(data)
	return nil
}

// ReportPreviewRequest represents a request to preview a report
type ReportPreviewRequest struct {
	ResourceType   ReportResourceType `json:"resourceType" binding:"required"`
	ResourceID     string             `json:"resourceId" binding:"required"`
	Format         ReportFormat       `json:"format" binding:"required"`
	IncludeFilters bool               `json:"includeFilters"`
}

// ReportPreviewResponse represents the response from a preview request
type ReportPreviewResponse struct {
	PreviewURL string `json:"previewUrl"`
	FileSize   int64  `json:"fileSize"`
	ExpiresAt  string `json:"expiresAt"`
}

// CreateScheduledReportRequest represents a request to create a scheduled report
type CreateScheduledReportRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description,omitempty"`

	// Resource
	ResourceType ReportResourceType `json:"resourceType" binding:"required"`
	ResourceID   string             `json:"resourceId" binding:"required"`

	// Schedule
	ScheduleType ReportScheduleType `json:"scheduleType" binding:"required"`
	CronExpr     string             `json:"cronExpr,omitempty"`
	TimeOfDay    string             `json:"timeOfDay,omitempty"`
	DayOfWeek    *int               `json:"dayOfWeek,omitempty"`
	DayOfMonth   *int               `json:"dayOfMonth,omitempty"`
	Timezone     string             `json:"timezone"`

	// Recipients
	Recipients []RecipientInput `json:"recipients"`

	// Format & Options
	Format         ReportFormat `json:"format" binding:"required"`
	IncludeFilters bool         `json:"includeFilters"`
	Subject        string       `json:"subject"`
	Message        string       `json:"message"`

	// Additional options
	Options map[string]interface{} `json:"options,omitempty"`
}

// RecipientInput represents a recipient input
type RecipientInput struct {
	Email string `json:"email" binding:"required,email"`
	Type  string `json:"type"` // "to", "cc", "bcc"
}

// UpdateScheduledReportRequest represents a request to update a scheduled report
type UpdateScheduledReportRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`

	// Schedule
	ScheduleType *ReportScheduleType `json:"scheduleType,omitempty"`
	CronExpr     *string             `json:"cronExpr,omitempty"`
	TimeOfDay    *string             `json:"timeOfDay,omitempty"`
	DayOfWeek    *int                `json:"dayOfWeek,omitempty"`
	DayOfMonth   *int                `json:"dayOfMonth,omitempty"`
	Timezone     *string             `json:"timezone,omitempty"`

	// Recipients
	Recipients *[]RecipientInput `json:"recipients,omitempty"`

	// Format & Options
	Format         *ReportFormat          `json:"format,omitempty"`
	IncludeFilters *bool                  `json:"includeFilters,omitempty"`
	Subject        *string                `json:"subject,omitempty"`
	Message        *string                `json:"message,omitempty"`
	Options        map[string]interface{} `json:"options,omitempty"`
}

// ScheduledReportResponse represents a scheduled report response
type ScheduledReportResponse struct {
	ScheduledReport
	Recipients []RecipientResponse `json:"recipients"`
}

// RecipientResponse represents a recipient in a response
type RecipientResponse struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
	Type  string    `json:"type"`
}

// ScheduledReportRunResponse represents a scheduled report run response
type ScheduledReportRunResponse struct {
	ID             uuid.UUID       `json:"id"`
	ReportID       uuid.UUID       `json:"reportId"`
	StartedAt      string          `json:"startedAt"`
	CompletedAt    *string         `json:"completedAt,omitempty"`
	Status         ReportRunStatus `json:"status"`
	ErrorMessage   *string         `json:"errorMessage,omitempty"`
	FileURL        *string         `json:"fileUrl,omitempty"`
	FileSize       *int64          `json:"fileSize,omitempty"`
	FileType       *string         `json:"fileType,omitempty"`
	DurationMs     *int64          `json:"durationMs,omitempty"`
	TriggeredBy    *string         `json:"triggeredBy,omitempty"`
	RecipientCount int             `json:"recipientCount"`
	CreatedAt      string          `json:"createdAt"`
}

// ScheduledReportListResponse represents a list response
type ScheduledReportListResponse struct {
	Reports []ScheduledReportResponse `json:"reports"`
	Total   int64                     `json:"total"`
	Page    int                       `json:"page"`
	Limit   int                       `json:"limit"`
}

// ScheduledReportRunListResponse represents a run list response
type ScheduledReportRunListResponse struct {
	Runs  []ScheduledReportRunResponse `json:"runs"`
	Total int64                        `json:"total"`
	Page  int                          `json:"page"`
	Limit int                          `json:"limit"`
}

// TriggerReportRequest represents a request to manually trigger a report
type TriggerReportRequest struct {
	TriggerType string `json:"triggerType"` // "manual", "api"
}

// TriggerReportResponse represents the response from triggering a report
type TriggerReportResponse struct {
	RunID     uuid.UUID       `json:"runId"`
	Status    ReportRunStatus `json:"status"`
	Message   string          `json:"message"`
	StartedAt time.Time       `json:"startedAt"`
}

// ScheduledReportFilter represents filter options for listing reports
type ScheduledReportFilter struct {
	UserID       *string
	ResourceType *ReportResourceType
	ResourceID   *string
	IsActive     *bool
	ScheduleType *ReportScheduleType
	Search       *string
	Page         int
	Limit        int
	OrderBy      string
}

// ScheduledReportRunFilter represents filter options for listing runs
type ScheduledReportRunFilter struct {
	ReportID  *string
	Status    *ReportRunStatus
	StartDate *time.Time
	EndDate   *time.Time
	Page      int
	Limit     int
	OrderBy   string
}
