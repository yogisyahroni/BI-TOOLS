package models

import (
	"time"
)

// JobExecution represents a pipeline execution record
type JobExecution struct {
	ID         string `json:"id" gorm:"primaryKey;type:varchar(30)"`
	PipelineID string `json:"pipelineId" gorm:"not null;index;column:pipelineId"`

	Status      string     `json:"status" gorm:"not null"` // PENDING, PROCESSING, EXTRACTING, TRANSFORMING, LOADING, COMPLETED, FAILED
	StartedAt   time.Time  `json:"startedAt"`
	CompletedAt *time.Time `json:"completedAt"`
	DurationMs  *int       `json:"durationMs"`

	// Execution metrics
	RowsProcessed     int   `json:"rowsProcessed" gorm:"default:0"`
	BytesProcessed    int64 `json:"bytesProcessed" gorm:"default:0"`
	QualityViolations int   `json:"qualityViolations" gorm:"default:0"`
	Progress          int   `json:"progress" gorm:"default:0"` // 0-100 percentage

	// Error and logging
	Error *string `json:"error"`
	Logs  *string `json:"logs" gorm:"type:jsonb"` // Array of structured log entries

	// Relationship
	Pipeline Pipeline `json:"pipeline,omitempty" gorm:"foreignKey:PipelineID"`
}

// TableName specifies the table name for GORM
func (JobExecution) TableName() string {
	return "JobExecution"
}

// ExecutionLog represents a single log entry during execution
type ExecutionLog struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"` // INFO, WARN, ERROR
	Message   string    `json:"message"`
	Step      string    `json:"step,omitempty"` // EXTRACT, TRANSFORM, LOAD, VALIDATE
	Details   string    `json:"details,omitempty"`
}
