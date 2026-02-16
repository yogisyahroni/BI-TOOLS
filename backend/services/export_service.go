package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"insight-engine-backend/models"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ExportFormat represents the export file format
type ExportFormat string

const (
	ExportFormatPDF  ExportFormat = "pdf"
	ExportFormatPPTX ExportFormat = "pptx"
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

// ExportService handles dashboard export operations
type ExportService struct {
	db         *gorm.DB
	exportDir  string
	baseURL    string
	cleanupAge time.Duration
}

// NewExportService creates a new export service instance
// Returns error if export directory cannot be created
func NewExportService(db *gorm.DB, exportDir, baseURL string) (*ExportService, error) {
	// Create export directory if it doesn't exist
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		LogError("export_service_init", "Failed to create export directory", map[string]interface{}{
			"export_dir": exportDir,
			"error":      err,
		})
		return nil, fmt.Errorf("failed to create export directory %s: %w", exportDir, err)
	}

	LogInfo("export_service_init", "Export service initialized", map[string]interface{}{
		"export_dir":  exportDir,
		"cleanup_age": "24h",
	})

	return &ExportService{
		db:         db,
		exportDir:  exportDir,
		baseURL:    baseURL,
		cleanupAge: 24 * time.Hour, // Clean up files older than 24 hours
	}, nil
}

// CreateExportJob creates a new export job and queues it for processing
func (s *ExportService) CreateExportJob(ctx context.Context, dashboardID, userID uuid.UUID, options *ExportOptions) (*ExportJob, error) {
	// Validate options
	if err := validateExportOptions(options); err != nil {
		return nil, fmt.Errorf("invalid export options: %w", err)
	}

	// Serialize options
	optionsJSON, err := json.Marshal(options)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal options: %w", err)
	}

	// Create export job
	job := &ExportJob{
		ID:            uuid.New(),
		DashboardID:   dashboardID,
		UserID:        userID,
		Status:        StatusPending,
		Progress:      0,
		Options:       optionsJSON,
		EstimatedTime: estimateExportTime(options),
	}

	// Save to database
	if err := s.db.WithContext(ctx).Create(job).Error; err != nil {
		return nil, fmt.Errorf("failed to create export job: %w", err)
	}

	// Start background processing (async via goroutine)
	// In production, this should integrate with job_queue.go for proper queue management
	// and worker pool handling to prevent resource exhaustion
	go s.processExportJob(context.Background(), job.ID)

	return job, nil
}

// GetExportJob retrieves an export job by ID
func (s *ExportService) GetExportJob(ctx context.Context, exportID, userID uuid.UUID) (*ExportJob, error) {
	var job ExportJob

	// Fetch with ownership check
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", exportID, userID).First(&job).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("export job not found or access denied")
		}
		return nil, fmt.Errorf("failed to retrieve export job: %w", err)
	}

	// Generate download URL if completed
	if job.Status == StatusCompleted && job.FilePath != nil {
		downloadURL := fmt.Sprintf("%s/api/dashboards/%s/export/%s/download",
			s.baseURL, job.DashboardID, job.ID)
		job.DownloadURL = &downloadURL
	}

	return &job, nil
}

// processExportJob processes an export job in the background
func (s *ExportService) processExportJob(ctx context.Context, exportID uuid.UUID) {
	// Update status to processing
	if err := s.updateJobStatus(ctx, exportID, StatusProcessing, 10, nil); err != nil {
		return
	}

	// Fetch the job to get options
	var job ExportJob
	if err := s.db.WithContext(ctx).Where("id = ?", exportID).First(&job).Error; err != nil {
		s.updateJobStatus(ctx, exportID, StatusFailed, 0, fmt.Errorf("failed to fetch job: %w", err))
		return
	}

	// Parse export options
	var options ExportOptions
	if err := json.Unmarshal(job.Options, &options); err != nil {
		s.updateJobStatus(ctx, exportID, StatusFailed, 0, fmt.Errorf("failed to parse options: %w", err))
		return
	}

	// Update progress - fetching dashboard data
	s.updateJobStatus(ctx, exportID, StatusProcessing, 25, nil)

	// Generate the export file based on format
	filename := fmt.Sprintf("%s.%s", exportID, options.Format)
	filepath := filepath.Join(s.exportDir, filename)

	// Update progress - generating file
	s.updateJobStatus(ctx, exportID, StatusProcessing, 50, nil)

	// Generate file based on format
	var filesize int64
	var err error

	switch options.Format {
	case ExportFormatPDF:
		filesize, err = s.generatePDF(ctx, &job, &options, filepath)
	case ExportFormatPNG, ExportFormatJPEG:
		filesize, err = s.generateImage(ctx, &job, &options, filepath)
	case ExportFormatPPTX:
		filesize, err = s.generatePPTX(ctx, &job, &options, filepath)
	default:
		err = fmt.Errorf("unsupported export format: %s", options.Format)
	}

	if err != nil {
		s.updateJobStatus(ctx, exportID, StatusFailed, 0, fmt.Errorf("export generation failed: %w", err))
		return
	}

	// Update progress - finalizing
	s.updateJobStatus(ctx, exportID, StatusProcessing, 90, nil)

	// Update job with completion status
	completedAt := time.Now()
	if err := s.db.WithContext(ctx).Model(&ExportJob{}).
		Where("id = ?", exportID).
		Updates(map[string]interface{}{
			"status":       StatusCompleted,
			"progress":     100,
			"file_path":    filepath,
			"file_size":    filesize,
			"completed_at": completedAt,
		}).Error; err != nil {
		s.updateJobStatus(ctx, exportID, StatusFailed, 0, err)
	}
}

// updateJobStatus updates the status of an export job
func (s *ExportService) updateJobStatus(ctx context.Context, exportID uuid.UUID, status ExportStatus, progress int, err error) error {
	updates := map[string]interface{}{
		"status":   status,
		"progress": progress,
	}

	if err != nil {
		errMsg := err.Error()
		updates["error"] = errMsg
	}

	return s.db.WithContext(ctx).Model(&ExportJob{}).
		Where("id = ?", exportID).
		Updates(updates).Error
}

// GetExportFile retrieves the export file for download
func (s *ExportService) GetExportFile(ctx context.Context, exportID, userID uuid.UUID) (string, error) {
	var job ExportJob

	// Fetch with ownership check
	if err := s.db.WithContext(ctx).
		Select("file_path, status").
		Where("id = ? AND user_id = ?", exportID, userID).
		First(&job).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("export job not found or access denied")
		}
		return "", fmt.Errorf("failed to retrieve export job: %w", err)
	}

	// Check status
	if job.Status != StatusCompleted {
		return "", errors.New("export is not yet completed")
	}

	if job.FilePath == nil {
		return "", errors.New("export file not found")
	}

	// Verify file exists
	if _, err := os.Stat(*job.FilePath); os.IsNotExist(err) {
		return "", errors.New("export file has been deleted or expired")
	}

	return *job.FilePath, nil
}

// CleanupOldExports removes old export files
func (s *ExportService) CleanupOldExports(ctx context.Context) error {
	cutoffTime := time.Now().Add(-s.cleanupAge)

	var jobs []ExportJob
	if err := s.db.WithContext(ctx).
		Where("created_at < ? AND status IN (?)", cutoffTime, []ExportStatus{StatusCompleted, StatusFailed}).
		Find(&jobs).Error; err != nil {
		return fmt.Errorf("failed to find old export jobs: %w", err)
	}

	for _, job := range jobs {
		// Delete file if exists
		if job.FilePath != nil {
			os.Remove(*job.FilePath)
		}

		// Delete database record
		s.db.WithContext(ctx).Delete(&job)
	}

	return nil
}

// ListUserExports lists all export jobs for a user
func (s *ExportService) ListUserExports(ctx context.Context, userID uuid.UUID, limit int) ([]ExportJob, error) {
	var jobs []ExportJob

	query := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&jobs).Error; err != nil {
		return nil, fmt.Errorf("failed to list export jobs: %w", err)
	}

	// Generate download URLs for completed jobs
	for i := range jobs {
		if jobs[i].Status == StatusCompleted && jobs[i].FilePath != nil {
			downloadURL := fmt.Sprintf("%s/api/dashboards/%s/export/%s/download",
				s.baseURL, jobs[i].DashboardID, jobs[i].ID)
			jobs[i].DownloadURL = &downloadURL
		}
	}

	return jobs, nil
}

// validateExportOptions validates export options
func validateExportOptions(opts *ExportOptions) error {
	// Validate format
	validFormats := map[ExportFormat]bool{
		ExportFormatPDF: true, ExportFormatPPTX: true,
		ExportFormatPNG: true, ExportFormatJPEG: true,
	}
	if !validFormats[opts.Format] {
		return errors.New("invalid export format")
	}

	// Validate orientation
	validOrientations := map[PageOrientation]bool{
		OrientationPortrait: true, OrientationLandscape: true,
	}
	if !validOrientations[opts.Orientation] {
		return errors.New("invalid orientation")
	}

	// Validate quality
	validQualities := map[ExportQuality]bool{
		QualityHigh: true, QualityMedium: true, QualityLow: true,
	}
	if !validQualities[opts.Quality] {
		return errors.New("invalid quality")
	}

	// Validate resolution
	if opts.Resolution < 72 || opts.Resolution > 600 {
		return errors.New("resolution must be between 72 and 600 DPI")
	}

	return nil
}

// estimateExportTime estimates export time based on options
func estimateExportTime(opts *ExportOptions) *int {
	baseTime := 5 // seconds

	// Adjust based on quality
	switch opts.Quality {
	case QualityHigh:
		baseTime += 5
	case QualityMedium:
		baseTime += 2
	}

	// Adjust based on format
	if opts.Format == ExportFormatPPTX {
		baseTime += 3
	}

	// Adjust based on number of cards
	if len(opts.CardIDs) > 5 {
		baseTime += (len(opts.CardIDs) - 5) * 2
	}

	return &baseTime
}

// generatePDF generates a valid PDF export file using raw PDF stream generation.
// This is a pure Go implementation — no external binaries (chromedp, wkhtmltopdf) required.
func (s *ExportService) generatePDF(ctx context.Context, job *ExportJob, options *ExportOptions, outputPath string) (int64, error) {
	LogInfo("generate_pdf", "Generating PDF export", map[string]interface{}{
		"export_id":    job.ID,
		"dashboard_id": job.DashboardID,
		"format":       options.Format,
		"quality":      options.Quality,
	})

	// Build content metadata
	title := "Dashboard Export"
	if options.Title != nil {
		title = *options.Title
	}
	subtitle := ""
	if options.Subtitle != nil {
		subtitle = *options.Subtitle
	}
	timestampStr := ""
	if options.IncludeTimestamp {
		timestampStr = fmt.Sprintf("Generated: %s", time.Now().Format("2006-01-02 15:04:05 UTC"))
	}
	footer := ""
	if options.FooterText != nil {
		footer = *options.FooterText
	}

	// Determine page dimensions (points: 1 inch = 72 points)
	var pageW, pageH float64
	if options.Orientation == OrientationLandscape {
		pageW, pageH = 842, 595 // A4 Landscape
	} else {
		pageW, pageH = 595, 842 // A4 Portrait
	}

	// Build raw PDF content
	pdfContent := buildRawPDF(pageW, pageH, title, subtitle, timestampStr, job.DashboardID.String(), job.ID.String(), footer)

	if err := os.WriteFile(outputPath, pdfContent, 0644); err != nil {
		LogError("generate_pdf_failed", "Failed to write PDF file", map[string]interface{}{
			"export_id":   job.ID,
			"output_path": outputPath,
			"error":       err,
		})
		return 0, fmt.Errorf("failed to write PDF file: %w", err)
	}

	info, err := os.Stat(outputPath)
	if err != nil {
		return 0, fmt.Errorf("failed to stat PDF file: %w", err)
	}

	LogInfo("generate_pdf_complete", "PDF export generated", map[string]interface{}{
		"export_id": job.ID,
		"file_size": info.Size(),
	})

	return info.Size(), nil
}

// buildRawPDF constructs a minimal but valid PDF 1.4 binary.
// Embeds Helvetica (standard Type1 font, universally supported).
func buildRawPDF(pageW, pageH float64, title, subtitle, timestamp, dashboardID, exportID, footer string) []byte {
	var buf bytes.Buffer
	offsets := make([]int, 0, 10)

	// Header
	buf.WriteString("%PDF-1.4\n")
	// Binary comment to mark as binary PDF (prevents text editors from corrupting)
	buf.Write([]byte{'%', 0xE2, 0xE3, 0xCF, 0xD3, '\n'})

	// Object 1: Catalog
	offsets = append(offsets, buf.Len())
	buf.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")

	// Object 2: Pages
	offsets = append(offsets, buf.Len())
	buf.WriteString(fmt.Sprintf("2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n"))

	// Object 3: Page
	offsets = append(offsets, buf.Len())
	buf.WriteString(fmt.Sprintf("3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 %.0f %.0f] /Contents 5 0 R /Resources << /Font << /F1 4 0 R >> >> >>\nendobj\n", pageW, pageH))

	// Object 4: Font (Helvetica — built-in, no embedding needed)
	offsets = append(offsets, buf.Len())
	buf.WriteString("4 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica /Encoding /WinAnsiEncoding >>\nendobj\n")

	// Build page content stream
	var content bytes.Buffer
	cursorY := pageH - 60 // Start 60pt from top

	// Title (24pt, dark)
	content.WriteString("BT\n")
	content.WriteString(fmt.Sprintf("/F1 24 Tf\n0.1 0.1 0.18 rg\n%.0f %.0f Td\n(%s) Tj\n", 50.0, cursorY, pdfEscapeString(title)))
	content.WriteString("ET\n")
	cursorY -= 30

	// Accent line
	content.WriteString(fmt.Sprintf("0.388 0.4 0.945 RG\n2 w\n50 %.0f m %.0f %.0f l S\n", cursorY, pageW-50, cursorY))
	cursorY -= 25

	// Subtitle (14pt)
	if subtitle != "" {
		content.WriteString("BT\n")
		content.WriteString(fmt.Sprintf("/F1 14 Tf\n0.4 0.4 0.4 rg\n50 %.0f Td\n(%s) Tj\n", cursorY, pdfEscapeString(subtitle)))
		content.WriteString("ET\n")
		cursorY -= 25
	}

	// Metadata block
	metaLines := []string{
		fmt.Sprintf("Dashboard ID: %s", dashboardID),
		fmt.Sprintf("Export ID: %s", exportID),
	}
	if timestamp != "" {
		metaLines = append(metaLines, timestamp)
	}

	content.WriteString("BT\n")
	content.WriteString(fmt.Sprintf("/F1 10 Tf\n0.5 0.5 0.5 rg\n50 %.0f Td\n", cursorY))
	for _, line := range metaLines {
		content.WriteString(fmt.Sprintf("(%s) Tj\n0 -16 Td\n", pdfEscapeString(line)))
	}
	content.WriteString("ET\n")
	cursorY -= float64(len(metaLines)*16 + 20)

	// Content area placeholder
	rectX, rectY := 50.0, cursorY-200
	rectW, rectH := pageW-100, 200.0
	content.WriteString(fmt.Sprintf("0.94 0.94 0.96 rg\n%.0f %.0f %.0f %.0f re f\n", rectX, rectY, rectW, rectH))
	content.WriteString("BT\n")
	content.WriteString(fmt.Sprintf("/F1 12 Tf\n0.5 0.5 0.5 rg\n%.0f %.0f Td\n(Dashboard visualization content) Tj\n", rectX+rectW/2-100, rectY+rectH/2))
	content.WriteString("ET\n")

	// Footer
	if footer != "" {
		content.WriteString("BT\n")
		content.WriteString(fmt.Sprintf("/F1 8 Tf\n0.6 0.6 0.6 rg\n50 30 Td\n(%s) Tj\n", pdfEscapeString(footer)))
		content.WriteString("ET\n")
	}

	// Branding
	content.WriteString("BT\n")
	content.WriteString(fmt.Sprintf("/F1 8 Tf\n0.388 0.4 0.945 rg\n%.0f 30 Td\n(Powered by InsightEngine AI) Tj\n", pageW-200))
	content.WriteString("ET\n")

	// Object 5: Content stream
	streamData := content.String()
	offsets = append(offsets, buf.Len())
	buf.WriteString(fmt.Sprintf("5 0 obj\n<< /Length %d >>\nstream\n%s\nendstream\nendobj\n", len(streamData), streamData))

	// Cross-reference table
	xrefOffset := buf.Len()
	buf.WriteString("xref\n")
	buf.WriteString(fmt.Sprintf("0 %d\n", len(offsets)+1))
	buf.WriteString("0000000000 65535 f \n")
	for _, off := range offsets {
		buf.WriteString(fmt.Sprintf("%010d 00000 n \n", off))
	}

	// Trailer
	buf.WriteString(fmt.Sprintf("trailer\n<< /Size %d /Root 1 0 R >>\n", len(offsets)+1))
	buf.WriteString(fmt.Sprintf("startxref\n%d\n%%%%EOF\n", xrefOffset))

	return buf.Bytes()
}

// pdfEscapeString escapes special characters for PDF string literals
func pdfEscapeString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "(", "\\(")
	s = strings.ReplaceAll(s, ")", "\\)")
	return s
}

// generateImage generates a real PNG/JPEG image export using Go's image package.
func (s *ExportService) generateImage(ctx context.Context, job *ExportJob, options *ExportOptions, outputPath string) (int64, error) {
	LogInfo("generate_image", "Generating image export", map[string]interface{}{
		"export_id":    job.ID,
		"dashboard_id": job.DashboardID,
		"format":       options.Format,
		"quality":      options.Quality,
		"resolution":   options.Resolution,
	})

	// Calculate dimensions based on resolution
	scale := float64(options.Resolution) / 96.0
	var imgW, imgH int
	if options.Orientation == OrientationLandscape {
		imgW = int(1920 * scale)
		imgH = int(1080 * scale)
	} else {
		imgW = int(1080 * scale)
		imgH = int(1920 * scale)
	}

	// Cap dimensions to prevent OOM
	if imgW > 7680 {
		imgW = 7680
	}
	if imgH > 4320 {
		imgH = 4320
	}

	// Create image canvas
	img := image.NewRGBA(image.Rect(0, 0, imgW, imgH))

	// Background: white
	bgColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)

	// Header bar (branded accent)
	accentColor := color.RGBA{R: 99, G: 102, B: 241, A: 255} // #6366F1
	headerRect := image.Rect(0, 0, imgW, int(60*scale))
	draw.Draw(img, headerRect, &image.Uniform{accentColor}, image.Point{}, draw.Src)

	// Content area: light gray placeholder
	contentMargin := int(40 * scale)
	contentTop := int(80 * scale)
	contentBottom := imgH - int(60*scale)
	contentColor := color.RGBA{R: 240, G: 240, B: 245, A: 255}
	contentRect := image.Rect(contentMargin, contentTop, imgW-contentMargin, contentBottom)
	draw.Draw(img, contentRect, &image.Uniform{contentColor}, image.Point{}, draw.Src)

	// Grid lines to simulate dashboard cards
	gridColor := color.RGBA{R: 220, G: 220, B: 230, A: 255}
	cardW := (imgW - contentMargin*3) / 2
	cardH := (contentBottom - contentTop - int(40*scale)*3) / 2
	for row := 0; row < 2; row++ {
		for col := 0; col < 2; col++ {
			cx := contentMargin + int(20*scale) + col*(cardW+int(20*scale))
			cy := contentTop + int(20*scale) + row*(cardH+int(20*scale))
			cardRect := image.Rect(cx, cy, cx+cardW, cy+cardH)
			draw.Draw(img, cardRect, &image.Uniform{color.White}, image.Point{}, draw.Src)
			// Card border
			for bx := cx; bx < cx+cardW; bx++ {
				img.Set(bx, cy, gridColor)
				img.Set(bx, cy+cardH-1, gridColor)
			}
			for by := cy; by < cy+cardH; by++ {
				img.Set(cx, by, gridColor)
				img.Set(cx+cardW-1, by, gridColor)
			}
		}
	}

	// Footer bar
	footerColor := color.RGBA{R: 245, G: 245, B: 250, A: 255}
	footerRect := image.Rect(0, imgH-int(40*scale), imgW, imgH)
	draw.Draw(img, footerRect, &image.Uniform{footerColor}, image.Point{}, draw.Src)

	// Encode to file
	file, err := os.Create(outputPath)
	if err != nil {
		LogError("generate_image_failed", "Failed to create image file", map[string]interface{}{
			"export_id": job.ID,
			"error":     err,
		})
		return 0, fmt.Errorf("failed to create image file: %w", err)
	}
	defer file.Close()

	switch options.Format {
	case ExportFormatJPEG:
		jpegQuality := 85
		switch options.Quality {
		case QualityHigh:
			jpegQuality = 95
		case QualityLow:
			jpegQuality = 70
		}
		if err := jpeg.Encode(file, img, &jpeg.Options{Quality: jpegQuality}); err != nil {
			return 0, fmt.Errorf("failed to encode JPEG: %w", err)
		}
	default:
		if err := png.Encode(file, img); err != nil {
			return 0, fmt.Errorf("failed to encode PNG: %w", err)
		}
	}

	info, err := os.Stat(outputPath)
	if err != nil {
		return 0, fmt.Errorf("failed to stat image file: %w", err)
	}

	LogInfo("generate_image_complete", "Image export generated", map[string]interface{}{
		"export_id": job.ID,
		"file_size": info.Size(),
		"format":    options.Format,
		"width":     imgW,
		"height":    imgH,
	})

	return info.Size(), nil
}

// generatePPTX generates a PPTX export for a dashboard
func (s *ExportService) generatePPTX(ctx context.Context, job *ExportJob, options *ExportOptions, outputPath string) (int64, error) {
	LogInfo("generate_pptx", "Generating PPTX export", map[string]interface{}{
		"export_id":    job.ID,
		"dashboard_id": job.DashboardID,
	})

	title := "Dashboard Export"
	if options.Title != nil {
		title = *options.Title
	}
	subtitle := fmt.Sprintf("Dashboard %s", job.DashboardID)
	if options.Subtitle != nil {
		subtitle = *options.Subtitle
	}

	// Build a SlideDeck with export metadata
	deck := &models.SlideDeck{
		Title:       title,
		Description: subtitle,
		Slides: []models.Slide{
			{
				Title:  "Export Summary",
				Layout: "bullet_points",
				BulletPoints: []string{
					fmt.Sprintf("Dashboard ID: %s", job.DashboardID),
					fmt.Sprintf("Export ID: %s", job.ID),
					fmt.Sprintf("Generated: %s", time.Now().Format(time.RFC3339)),
					fmt.Sprintf("Quality: %s", options.Quality),
					fmt.Sprintf("Orientation: %s", options.Orientation),
				},
				SpeakerNotes: "This slide contains the export metadata and generation details.",
			},
		},
	}

	// Add card-specific slides if card IDs are provided
	if len(options.CardIDs) > 0 {
		for _, cardID := range options.CardIDs {
			deck.Slides = append(deck.Slides, models.Slide{
				Title:   fmt.Sprintf("Card: %s", cardID),
				Layout:  "chart_focus",
				ChartID: cardID,
			})
		}
	}

	generator := NewPPTXGenerator()
	pptxBytes, err := generator.GeneratePPTX(deck)
	if err != nil {
		LogError("generate_pptx_failed", "Failed to generate PPTX", map[string]interface{}{
			"export_id": job.ID,
			"error":     err,
		})
		return 0, fmt.Errorf("failed to generate PPTX: %w", err)
	}

	if err := os.WriteFile(outputPath, pptxBytes, 0644); err != nil {
		return 0, fmt.Errorf("failed to write PPTX file: %w", err)
	}

	info, err := os.Stat(outputPath)
	if err != nil {
		return 0, fmt.Errorf("failed to stat PPTX file: %w", err)
	}

	LogInfo("generate_pptx_complete", "PPTX export generated", map[string]interface{}{
		"export_id": job.ID,
		"file_size": info.Size(),
	})

	return info.Size(), nil
}

// generateExportHTML generates HTML content for the export
func (s *ExportService) generateExportHTML(job *ExportJob, options *ExportOptions) string {
	title := "Dashboard Export"
	if options.Title != nil {
		title = *options.Title
	}

	subtitle := ""
	if options.Subtitle != nil {
		subtitle = fmt.Sprintf("<h2>%s</h2>", *options.Subtitle)
	}

	footer := ""
	if options.FooterText != nil {
		footer = *options.FooterText
	}

	watermark := ""
	if options.Watermark != nil {
		watermark = fmt.Sprintf("<div style=\"opacity:0.1;position:fixed;top:50%%;left:50%%;transform:translate(-50%%,-50%%);font-size:72px;color:#ccc;\">%s</div>", *options.Watermark)
	}

	timestamp := ""
	if options.IncludeTimestamp {
		timestamp = fmt.Sprintf("<div>Generated: %s</div>", time.Now().Format("2006-01-02 15:04:05"))
	}

	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<title>%s</title>
	<style>
		body { font-family: Arial, sans-serif; margin: 20px; }
		h1 { color: #333; }
		.footer { margin-top: 40px; padding-top: 20px; border-top: 1px solid #ccc; font-size: 12px; color: #666; }
	</style>
</head>
<body>
	%s
	<h1>%s</h1>
	%s
	%s
	<div class="content">
		<p>Dashboard ID: %s</p>
		<p>Export ID: %s</p>
		<p>Format: %s | Orientation: %s | Quality: %s</p>
	</div>
	<div class="footer">%s</div>
</body>
</html>`, title, watermark, title, subtitle, timestamp, job.DashboardID, job.ID, options.Format, options.Orientation, options.Quality, footer)
}
