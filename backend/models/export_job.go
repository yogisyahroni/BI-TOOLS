package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ExportFormat represents the export file format
type ExportFormat string

const (
	ExportFormatPDF  ExportFormat = "pdf"
	ExportFormatPPTX ExportFormat = "pptx"
	ExportFormatXLSX ExportFormat = "xlsx"
	ExportFormatCSV  ExportFormat = "csv"
	ExportFormatPNG  ExportFormat = "png"
	ExportFormatJPEG ExportFormat = "jpeg"
)

// PageOrientation represents page orientation
type PageOrientation string

const (
	OrientationPortrait  PageOrientation = "portrait"
	OrientationLandscape PageOrientation = "landscape"
)

// PageSize represents page size presets
type PageSize string

const (
	PageSizeA4      PageSize = "A4"
	PageSizeLetter  PageSize = "Letter"
	PageSizeLegal   PageSize = "Legal"
	PageSizeTabloid PageSize = "Tabloid"
	PageSizeCustom  PageSize = "Custom"
)

// ExportQuality represents export quality
type ExportQuality string

const (
	QualityHigh   ExportQuality = "high"
	QualityMedium ExportQuality = "medium"
	QualityLow    ExportQuality = "low"
)

// ExportStatus represents the export job status
type ExportStatus string

const (
	StatusPending    ExportStatus = "pending"
	StatusProcessing ExportStatus = "processing"
	StatusCompleted  ExportStatus = "completed"
	StatusFailed     ExportStatus = "failed"
)

// ExportOptions holds all export configuration
type ExportOptions struct {
	Format            ExportFormat    `json:"format"`
	Orientation       PageOrientation `json:"orientation"`
	PageSize          PageSize        `json:"pageSize"`
	CustomWidth       *int            `json:"customWidth,omitempty"`
	CustomHeight      *int            `json:"customHeight,omitempty"`
	Quality           ExportQuality   `json:"quality"`
	IncludeFilters    bool            `json:"includeFilters"`
	IncludeTimestamp  bool            `json:"includeTimestamp"`
	IncludeDataTables bool            `json:"includeDataTables"`
	Title             *string         `json:"title,omitempty"`
	Subtitle          *string         `json:"subtitle,omitempty"`
	FooterText        *string         `json:"footerText,omitempty"`
	Watermark         *string         `json:"watermark,omitempty"`
	Resolution        int             `json:"resolution"`
	CardIDs           []string        `json:"cardIds,omitempty"`
	CurrentTabOnly    bool            `json:"currentTabOnly,omitempty"`
}

// ExportJob represents an export job
type ExportJob struct {
	ID            uuid.UUID       `json:"exportId" gorm:"type:uuid;primaryKey"`
	DashboardID   uuid.UUID       `json:"dashboardId" gorm:"type:uuid;not null;index"`
	UserID        uuid.UUID       `json:"userId" gorm:"type:uuid;not null;index"`
	Status        ExportStatus    `json:"status" gorm:"type:varchar(20);not null;index"`
	Progress      int             `json:"progress" gorm:"default:0"`
	Options       json.RawMessage `json:"options" gorm:"type:jsonb"`
	DownloadURL   *string         `json:"downloadUrl,omitempty" gorm:"type:text"`
	FilePath      *string         `json:"-" gorm:"type:text"`
	FileSize      *int64          `json:"fileSize,omitempty"`
	Error         *string         `json:"error,omitempty" gorm:"type:text"`
	EstimatedTime *int            `json:"estimatedTime,omitempty"`
	CreatedAt     time.Time       `json:"createdAt"`
	UpdatedAt     time.Time       `json:"updatedAt"`
	CompletedAt   *time.Time      `json:"completedAt,omitempty"`
}

// TableName specifies the table name
func (ExportJob) TableName() string {
	return "export_jobs"
}
