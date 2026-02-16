package models

import (
	"time"
)

// Pipeline represents a data pipeline configuration
type Pipeline struct {
	ID          string  `json:"id" gorm:"primaryKey;type:varchar(30)"`
	Name        string  `json:"name" gorm:"not null"`
	Description *string `json:"description"`
	WorkspaceID string  `json:"workspaceId" gorm:"not null;index"`

	// Source Configuration
	SourceType   string  `json:"sourceType" gorm:"not null"` // POSTGRES, MYSQL, CSV, REST_API
	SourceConfig string  `json:"sourceConfig" gorm:"type:jsonb;not null"`
	ConnectionID *string `json:"connectionId" gorm:"index"` // FK to Connection for DB sources
	SourceQuery  *string `json:"sourceQuery"`               // SQL query to execute on source

	// ELT vs ETL
	Mode                string  `json:"mode" gorm:"default:ELT"` // ETL | ELT
	TransformationSteps *string `json:"transformationSteps" gorm:"type:jsonb"`

	// Destination Configuration
	DestinationType   string  `json:"destinationType" gorm:"default:INTERNAL_RAW"`
	DestinationConfig *string `json:"destinationConfig" gorm:"type:jsonb"`

	// Schedule
	ScheduleCron *string `json:"scheduleCron"`
	IsActive     bool    `json:"isActive" gorm:"default:true;index"`

	// Safety
	RowLimit int `json:"rowLimit" gorm:"default:100000"` // Max rows per execution

	// Execution Tracking
	LastRunAt  *time.Time `json:"lastRunAt"`
	LastStatus *string    `json:"lastStatus"` // SUCCESS, FAILED

	// Relationships
	Executions   []JobExecution `json:"executions,omitempty" gorm:"foreignKey:PipelineID"`
	QualityRules []QualityRule  `json:"qualityRules,omitempty" gorm:"foreignKey:PipelineID"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// TableName specifies the table name for GORM
func (Pipeline) TableName() string {
	return "Pipeline"
}

// TransformStep defines a single transformation operation
type TransformStep struct {
	Type   string                 `json:"type"`   // FILTER, RENAME, CAST, AGGREGATE, DEDUPLICATE, VALIDATE
	Config map[string]interface{} `json:"config"` // Step-specific configuration
	Order  int                    `json:"order"`  // Execution order
}

// SourceConfig holds parsed source configuration
type SourceConfig struct {
	// For DB sources (connectionId takes precedence)
	Host     string `json:"host,omitempty"`
	Port     int    `json:"port,omitempty"`
	Database string `json:"database,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	SSLMode  string `json:"sslMode,omitempty"`
	Query    string `json:"query,omitempty"`

	// For CSV sources
	FilePath  string `json:"filePath,omitempty"`
	Delimiter string `json:"delimiter,omitempty"`
	HasHeader bool   `json:"hasHeader,omitempty"`

	// For REST API sources
	URL     string            `json:"url,omitempty"`
	Method  string            `json:"method,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
}

// DestConfig holds parsed destination configuration
type DestConfig struct {
	ConnectionID string `json:"connectionId,omitempty"` // For external DB destination
	TableName    string `json:"tableName,omitempty"`
	Schema       string `json:"schema,omitempty"`
	WriteMode    string `json:"writeMode,omitempty"` // APPEND, OVERWRITE, UPSERT
	UpsertKey    string `json:"upsertKey,omitempty"` // Column for UPSERT dedup
}
