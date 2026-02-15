package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// EmailQueueStatus represents the status of an email in the queue
type EmailQueueStatus string

const (
	EmailStatusPending   EmailQueueStatus = "pending"
	EmailStatusSending   EmailQueueStatus = "sending"
	EmailStatusSent      EmailQueueStatus = "sent"
	EmailStatusDelivered EmailQueueStatus = "delivered"
	EmailStatusOpened    EmailQueueStatus = "opened"
	EmailStatusFailed    EmailQueueStatus = "failed"
	EmailStatusBounced   EmailQueueStatus = "bounced"
)

// EmailQueue represents an email waiting to be sent
type EmailQueue struct {
	ID       uuid.UUID        `gorm:"type:uuid;primaryKey" json:"id"`
	Status   EmailQueueStatus `gorm:"type:varchar(20);not null;index" json:"status"`
	Priority int              `gorm:"default:5" json:"priority"` // 1-10, lower is higher priority

	// Recipient info
	To        string  `gorm:"type:text;not null" json:"to"`
	Cc        *string `gorm:"type:text" json:"cc,omitempty"`
	Bcc       *string `gorm:"type:text" json:"bcc,omitempty"`
	FromEmail string  `gorm:"type:text;not null" json:"fromEmail"`
	FromName  string  `gorm:"type:text" json:"fromName,omitempty"`

	// Content
	Subject      string         `gorm:"type:text;not null" json:"subject"`
	BodyHTML     *string        `gorm:"type:text" json:"bodyHtml,omitempty"`
	BodyText     *string        `gorm:"type:text" json:"bodyText,omitempty"`
	TemplateID   *uuid.UUID     `gorm:"type:uuid;index" json:"templateId,omitempty"`
	TemplateData datatypes.JSON `gorm:"type:jsonb" json:"templateData,omitempty"`

	// Attachments (stored as JSON array of file paths/metadata)
	Attachments datatypes.JSON `gorm:"type:jsonb" json:"attachments,omitempty"`

	// Metadata
	TrackOpens  bool       `gorm:"default:false" json:"trackOpens"`
	TrackClicks bool       `gorm:"default:false" json:"trackClicks"`
	IsBulk      bool       `gorm:"default:false" json:"isBulk"` // Part of a bulk send
	BatchID     *uuid.UUID `gorm:"type:uuid;index" json:"batchId,omitempty"`

	// Scheduling
	ScheduledAt *time.Time `gorm:"index" json:"scheduledAt,omitempty"`
	SendAfter   *time.Time `json:"sendAfter,omitempty"`
	ExpiresAt   *time.Time `json:"expiresAt,omitempty"`

	// Tracking
	SentAt      *time.Time `json:"sentAt,omitempty"`
	DeliveredAt *time.Time `json:"deliveredAt,omitempty"`
	OpenedAt    *time.Time `json:"openedAt,omitempty"`
	OpenCount   int        `gorm:"default:0" json:"openCount"`
	ClickCount  int        `gorm:"default:0" json:"clickCount"`
	LastError   *string    `gorm:"type:text" json:"lastError,omitempty"`
	RetryCount  int        `gorm:"default:0" json:"retryCount"`
	MaxRetries  int        `gorm:"default:3" json:"maxRetries"`

	// Relationships
	Template *EmailTemplate `gorm:"foreignKey:TemplateID" json:"template,omitempty"`
	Logs     []EmailLog     `gorm:"foreignKey:EmailQueueID" json:"logs,omitempty"`

	// Timestamps
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName specifies the table name for EmailQueue
func (EmailQueue) TableName() string {
	return "email_queue"
}

// EmailAttachment represents a file attachment for emails
type EmailAttachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	FilePath    string `json:"filePath"`
	FileSize    int64  `json:"fileSize"`
}

// GetAttachments parses the attachments JSON
func (e *EmailQueue) GetAttachments() ([]EmailAttachment, error) {
	if e.Attachments == nil {
		return []EmailAttachment{}, nil
	}

	var attachments []EmailAttachment
	if err := json.Unmarshal(e.Attachments, &attachments); err != nil {
		return nil, err
	}
	return attachments, nil
}

// SetAttachments serializes attachments to JSON
func (e *EmailQueue) SetAttachments(attachments []EmailAttachment) error {
	if len(attachments) == 0 {
		e.Attachments = nil
		return nil
	}

	data, err := json.Marshal(attachments)
	if err != nil {
		return err
	}
	e.Attachments = datatypes.JSON(data)
	return nil
}

// EmailLog represents a log entry for email operations
type EmailLog struct {
	ID           uuid.UUID        `gorm:"type:uuid;primaryKey" json:"id"`
	EmailQueueID uuid.UUID        `gorm:"type:uuid;not null;index" json:"emailQueueId"`
	Event        EmailLogEvent    `gorm:"type:varchar(50);not null;index" json:"event"`
	Status       EmailQueueStatus `gorm:"type:varchar(20);not null" json:"status"`
	Message      *string          `gorm:"type:text" json:"message,omitempty"`

	// Technical details
	IPAddress    *string `gorm:"type:varchar(45)" json:"ipAddress,omitempty"`
	UserAgent    *string `gorm:"type:text" json:"userAgent,omitempty"`
	SMTPCode     *int    `json:"smtpCode,omitempty"`
	SMTPResponse *string `gorm:"type:text" json:"smtpResponse,omitempty"`

	// Metadata
	Metadata datatypes.JSON `gorm:"type:jsonb" json:"metadata,omitempty"`

	// Timestamp
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

// EmailLogEvent represents different types of email events
type EmailLogEvent string

const (
	EmailEventQueued    EmailLogEvent = "queued"
	EmailEventSent      EmailLogEvent = "sent"
	EmailEventDelivered EmailLogEvent = "delivered"
	EmailEventOpened    EmailLogEvent = "opened"
	EmailEventClicked   EmailLogEvent = "clicked"
	EmailEventBounced   EmailLogEvent = "bounced"
	EmailEventFailed    EmailLogEvent = "failed"
	EmailEventRetry     EmailLogEvent = "retry"
	EmailEventExpired   EmailLogEvent = "expired"
	EmailEventCancelled EmailLogEvent = "cancelled"
)

// TableName specifies the table name for EmailLog
func (EmailLog) TableName() string {
	return "email_logs"
}

// EmailTemplate represents a reusable email template
type EmailTemplate struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(255);not null;uniqueIndex" json:"name"`
	Description *string   `gorm:"type:text" json:"description,omitempty"`

	// Template content
	Subject  string `gorm:"type:text;not null" json:"subject"`
	BodyHTML string `gorm:"type:text" json:"bodyHtml"`
	BodyText string `gorm:"type:text" json:"bodyText"`

	// Template variables (stored as JSON array of variable names)
	Variables datatypes.JSON `gorm:"type:jsonb" json:"variables,omitempty"`

	// Category for organization
	Category string `gorm:"type:varchar(100);index" json:"category"`

	// Status
	IsActive  bool `gorm:"default:true" json:"isActive"`
	IsDefault bool `gorm:"default:false" json:"isDefault"` // Default template for category

	// Usage tracking
	UsageCount int `gorm:"default:0" json:"usageCount"`

	// Ownership
	CreatedBy string    `gorm:"type:varchar(255);index" json:"createdBy"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName specifies the table name for EmailTemplate
func (EmailTemplate) TableName() string {
	return "email_templates"
}

// GetVariables parses the variables JSON
func (t *EmailTemplate) GetVariables() ([]string, error) {
	if t.Variables == nil {
		return []string{}, nil
	}

	var variables []string
	if err := json.Unmarshal(t.Variables, &variables); err != nil {
		return nil, err
	}
	return variables, nil
}

// SetVariables serializes variables to JSON
func (t *EmailTemplate) SetVariables(variables []string) error {
	if len(variables) == 0 {
		t.Variables = nil
		return nil
	}

	data, err := json.Marshal(variables)
	if err != nil {
		return err
	}
	t.Variables = datatypes.JSON(data)
	return nil
}

// EmailBatch represents a batch of emails sent together
type EmailBatch struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(255)" json:"name"`
	Description *string   `gorm:"type:text" json:"description,omitempty"`

	// Batch stats
	TotalCount   int `json:"totalCount"`
	SentCount    int `json:"sentCount"`
	FailedCount  int `json:"failedCount"`
	PendingCount int `json:"pendingCount"`

	// Template used for batch
	TemplateID *uuid.UUID `gorm:"type:uuid" json:"templateId,omitempty"`

	// Status
	Status string `gorm:"type:varchar(20);default:'pending'" json:"status"`

	// Timestamps
	StartedAt   *time.Time `json:"startedAt,omitempty"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`

	// Relationships
	Emails   []EmailQueue   `gorm:"foreignKey:BatchID" json:"emails,omitempty"`
	Template *EmailTemplate `gorm:"foreignKey:TemplateID" json:"template,omitempty"`
}

// TableName specifies the table name for EmailBatch
func (EmailBatch) TableName() string {
	return "email_batches"
}

// EmailTrackingPixel represents a tracking pixel for open tracking
type EmailTrackingPixel struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	EmailQueueID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"emailQueueId"`
	Token        string    `gorm:"type:varchar(255);not null;uniqueIndex" json:"token"`

	// Tracking data
	IPAddress     *string    `gorm:"type:varchar(45)" json:"ipAddress,omitempty"`
	UserAgent     *string    `gorm:"type:text" json:"userAgent,omitempty"`
	OpenCount     int        `gorm:"default:0" json:"openCount"`
	FirstOpenedAt *time.Time `json:"firstOpenedAt,omitempty"`
	LastOpenedAt  *time.Time `json:"lastOpenedAt,omitempty"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

// TableName specifies the table name for EmailTrackingPixel
func (EmailTrackingPixel) TableName() string {
	return "email_tracking_pixels"
}

// EmailClickLink represents a tracked link in an email
type EmailClickLink struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	EmailQueueID uuid.UUID `gorm:"type:uuid;not null;index" json:"emailQueueId"`
	Token        string    `gorm:"type:varchar(255);not null" json:"token"`
	OriginalURL  string    `gorm:"type:text;not null" json:"originalUrl"`

	// Tracking data
	ClickCount     int        `gorm:"default:0" json:"clickCount"`
	FirstClickedAt *time.Time `json:"firstClickedAt,omitempty"`
	LastClickedAt  *time.Time `json:"lastClickedAt,omitempty"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

// TableName specifies the table name for EmailClickLink
func (EmailClickLink) TableName() string {
	return "email_click_links"
}
