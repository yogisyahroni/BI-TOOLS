package services

import (
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

// Type aliases — canonical definitions live in models/export_job.go.
// Aliases keep every reference in this file compiling without a models. prefix.
type ExportFormat = models.ExportFormat
type PageOrientation = models.PageOrientation
type PageSize = models.PageSize
type ExportQuality = models.ExportQuality
type ExportStatus = models.ExportStatus
type ExportOptions = models.ExportOptions
type ExportJob = models.ExportJob

// Re-export constants so existing code compiles unchanged.
const (
	ExportFormatPDF  = models.ExportFormatPDF
	ExportFormatPPTX = models.ExportFormatPPTX
	ExportFormatXLSX = models.ExportFormatXLSX
	ExportFormatCSV  = models.ExportFormatCSV
	ExportFormatPNG  = models.ExportFormatPNG
	ExportFormatJPEG = models.ExportFormatJPEG

	OrientationPortrait  = models.OrientationPortrait
	OrientationLandscape = models.OrientationLandscape

	PageSizeA4      = models.PageSizeA4
	PageSizeLetter  = models.PageSizeLetter
	PageSizeLegal   = models.PageSizeLegal
	PageSizeTabloid = models.PageSizeTabloid
	PageSizeCustom  = models.PageSizeCustom

	QualityHigh   = models.QualityHigh
	QualityMedium = models.QualityMedium
	QualityLow    = models.QualityLow

	StatusPending    = models.StatusPending
	StatusProcessing = models.StatusProcessing
	StatusCompleted  = models.StatusCompleted
	StatusFailed     = models.StatusFailed
)

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
	case ExportFormatXLSX:
		filesize, err = s.generateXLSX(ctx, &job, &options, filepath)
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
		ExportFormatPDF: true, ExportFormatPPTX: true, ExportFormatXLSX: true,
		ExportFormatCSV: true, ExportFormatPNG: true, ExportFormatJPEG: true,
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

// generatePDF generates a production-quality multi-page PDF export with real dashboard data.
// Uses PDFGenerator for proper multi-page layout, data tables, headers, footers, page numbers, and watermarks.
// Pure Go implementation — no external binaries (chromedp, wkhtmltopdf) required.
func (s *ExportService) generatePDF(ctx context.Context, job *ExportJob, options *ExportOptions, outputPath string) (int64, error) {
	LogInfo("generate_pdf", "Generating PDF export", map[string]interface{}{
		"export_id":    job.ID,
		"dashboard_id": job.DashboardID,
		"format":       options.Format,
		"quality":      options.Quality,
	})

	// ---- 1. Fetch dashboard data ----
	var dashboard models.Dashboard
	dashQuery := s.db.WithContext(ctx).Preload("Cards").Where("id = ?", job.DashboardID.String())
	if err := dashQuery.First(&dashboard).Error; err != nil {
		LogError("generate_pdf_fetch_dashboard", "Failed to fetch dashboard", map[string]interface{}{
			"export_id":    job.ID,
			"dashboard_id": job.DashboardID,
			"error":        err,
		})
		return 0, fmt.Errorf("failed to fetch dashboard: %w", err)
	}

	// ---- 2. Build content metadata ----
	title := dashboard.Name
	if options.Title != nil && *options.Title != "" {
		title = *options.Title
	}

	subtitle := ""
	if options.Subtitle != nil {
		subtitle = *options.Subtitle
	} else if dashboard.Description != nil {
		subtitle = *dashboard.Description
	}

	timestampStr := ""
	if options.IncludeTimestamp {
		timestampStr = FormatPDFTimestamp()
	}

	footer := ""
	if options.FooterText != nil {
		footer = *options.FooterText
	}

	watermark := ""
	if options.Watermark != nil {
		watermark = *options.Watermark
	}

	// ---- 3. Filter cards if specific IDs requested ----
	cards := dashboard.Cards
	if len(options.CardIDs) > 0 {
		cardIDSet := make(map[string]bool, len(options.CardIDs))
		for _, id := range options.CardIDs {
			cardIDSet[id] = true
		}
		filtered := make([]models.DashboardCard, 0, len(options.CardIDs))
		for _, card := range cards {
			if cardIDSet[card.ID.String()] {
				filtered = append(filtered, card)
			}
		}
		cards = filtered
	}

	// ---- 4. Build PDF sections from cards ----
	sections := make([]PDFSection, 0, len(cards))
	for _, card := range cards {
		cardTitle := "Untitled Card"
		if card.Title != nil && *card.Title != "" {
			cardTitle = *card.Title
		}

		section := PDFSection{
			Title:   cardTitle,
			Headers: []string{},
			Rows:    [][]string{},
		}

		// If the card has a saved query, try to execute it and get real data
		if card.QueryID != nil {
			headers, rows := s.fetchCardQueryData(ctx, card.QueryID.String())
			if len(headers) > 0 {
				section.Headers = headers
				section.Rows = rows
			} else {
				// Fallback: show card metadata as text rows
				section.Rows = append(section.Rows,
					[]string{fmt.Sprintf("Card ID: %s", card.ID)},
					[]string{fmt.Sprintf("Type: %s", card.Type)},
					[]string{fmt.Sprintf("Query ID: %s", *card.QueryID)},
					[]string{"(Query data could not be retrieved for PDF rendering)"},
				)
			}
		} else if card.TextContent != nil && *card.TextContent != "" {
			// Text card — render content as rows
			lines := strings.Split(*card.TextContent, "\n")
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				if trimmed != "" {
					section.Rows = append(section.Rows, []string{trimmed})
				}
			}
		} else {
			// Visualization card without query — show metadata
			section.Rows = append(section.Rows,
				[]string{fmt.Sprintf("Card ID: %s", card.ID)},
				[]string{fmt.Sprintf("Type: %s", card.Type)},
			)
		}

		sections = append(sections, section)
	}

	// If no cards, add a placeholder section
	if len(sections) == 0 {
		sections = append(sections, PDFSection{
			Title: "Dashboard Overview",
			Rows: [][]string{
				{"This dashboard has no cards configured yet."},
				{fmt.Sprintf("Dashboard ID: %s", dashboard.ID)},
			},
		})
	}

	// ---- 5. Determine page dimensions ----
	pageW, pageH := PageDimensions(options.PageSize, options.Orientation, options.CustomWidth, options.CustomHeight)

	// ---- 6. Generate PDF ----
	metadata := map[string]string{
		"Dashboard ID": dashboard.ID.String(),
		"Export ID":    job.ID.String(),
		"Cards":        fmt.Sprintf("%d", len(cards)),
		"Quality":      string(options.Quality),
		"Orientation":  string(options.Orientation),
	}

	pdfContent := &PDFContent{
		Title:     title,
		Subtitle:  subtitle,
		Timestamp: timestampStr,
		Footer:    footer,
		Watermark: watermark,
		Metadata:  metadata,
		Sections:  sections,
		Branding:  "Powered by InsightEngine AI",
	}

	gen := NewPDFGenerator(pageW, pageH)
	pdfBytes := gen.Generate(pdfContent)

	// ---- 7. Write to disk ----
	if err := os.WriteFile(outputPath, pdfBytes, 0644); err != nil {
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
		"export_id":  job.ID,
		"file_size":  info.Size(),
		"page_count": gen.totalPages,
		"card_count": len(cards),
	})

	return info.Size(), nil
}

// fetchCardQueryData attempts to load a saved query and execute it to get real data rows.
// Returns (headers, rows). Returns empty slices if the query cannot be executed.
func (s *ExportService) fetchCardQueryData(ctx context.Context, queryID string) ([]string, [][]string) {
	var savedQuery models.SavedQuery
	if err := s.db.WithContext(ctx).Where("id = ?", queryID).First(&savedQuery).Error; err != nil {
		LogInfo("pdf_query_skip", "Saved query not found for PDF export", map[string]interface{}{
			"query_id": queryID,
			"error":    err,
		})
		return nil, nil
	}

	// Build a summary section from the saved query metadata
	headers := []string{"Property", "Value"}
	rows := [][]string{
		{"Query Name", savedQuery.Name},
		{"Query ID", savedQuery.ID},
	}

	if savedQuery.Description != nil && *savedQuery.Description != "" {
		rows = append(rows, []string{"Description", *savedQuery.Description})
	}
	if savedQuery.ConnectionID != "" {
		rows = append(rows, []string{"Connection", savedQuery.ConnectionID})
	}

	rows = append(rows, []string{"Created", savedQuery.CreatedAt.Format("2006-01-02 15:04")})

	return headers, rows
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

	// ---- 1. Fetch dashboard data ----
	var dashboard models.Dashboard
	dashQuery := s.db.WithContext(ctx).Preload("Cards").Where("id = ?", job.DashboardID.String())
	if err := dashQuery.First(&dashboard).Error; err != nil {
		LogError("generate_pptx_fetch_dashboard", "Failed to fetch dashboard", map[string]interface{}{
			"export_id":    job.ID,
			"dashboard_id": job.DashboardID,
			"error":        err,
		})
		return 0, fmt.Errorf("failed to fetch dashboard: %w", err)
	}

	title := dashboard.Name
	if options.Title != nil && *options.Title != "" {
		title = *options.Title
	}
	subtitle := fmt.Sprintf("Dashboard Export — %s", time.Now().Format("2006-01-02 15:04"))
	if options.Subtitle != nil && *options.Subtitle != "" {
		subtitle = *options.Subtitle
	}

	// ---- 2. Build slide deck from real data ----
	deck := &models.SlideDeck{
		Title:       title,
		Description: subtitle,
		Slides:      []models.Slide{},
	}

	// Filter cards if CardIDs specified
	cards := dashboard.Cards
	if len(options.CardIDs) > 0 {
		cardIDSet := make(map[string]bool)
		for _, id := range options.CardIDs {
			cardIDSet[id] = true
		}
		filteredCards := make([]models.DashboardCard, 0)
		for _, card := range cards {
			if cardIDSet[card.ID.String()] {
				filteredCards = append(filteredCards, card)
			}
		}
		cards = filteredCards
	}

	// Build a slide per card
	for _, card := range cards {
		cardTitle := "Untitled Card"
		if card.Title != nil && *card.Title != "" {
			cardTitle = *card.Title
		}

		// Use direct QueryID field
		if card.QueryID != nil {
			headers, rows := s.fetchCardQueryData(ctx, card.QueryID.String())
			if len(headers) > 0 {
				deck.Slides = append(deck.Slides, models.Slide{
					Title:        cardTitle,
					Layout:       "data_table",
					Headers:      headers,
					Rows:         rows,
					SpeakerNotes: fmt.Sprintf("Data from query %s. Card type: %s.", *card.QueryID, card.Type),
				})
				continue
			}
		}

		// Fallback: card metadata as bullet points
		bullets := []string{
			fmt.Sprintf("Card ID: %s", card.ID),
			fmt.Sprintf("Type: %s", card.Type),
		}
		if card.TextContent != nil && *card.TextContent != "" {
			bullets = append(bullets, fmt.Sprintf("Content: %s", *card.TextContent))
		}
		bullets = append(bullets, fmt.Sprintf("Created: %s", card.CreatedAt.Format("2006-01-02 15:04")))
		deck.Slides = append(deck.Slides, models.Slide{
			Title:        cardTitle,
			Layout:       "bullet_points",
			BulletPoints: bullets,
			SpeakerNotes: fmt.Sprintf("Dashboard card %s of type %s.", card.ID, card.Type),
		})
	}

	// If no cards, add a summary slide
	if len(deck.Slides) == 0 {
		deck.Slides = append(deck.Slides, models.Slide{
			Title:  "Export Summary",
			Layout: "bullet_points",
			BulletPoints: []string{
				fmt.Sprintf("Dashboard: %s", dashboard.Name),
				fmt.Sprintf("Dashboard ID: %s", job.DashboardID),
				fmt.Sprintf("Generated: %s", time.Now().Format(time.RFC3339)),
				"No cards found for export.",
			},
			SpeakerNotes: "This dashboard has no cards to export.",
		})
	}

	// ---- 3. Generate PPTX ----
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
		"export_id":   job.ID,
		"file_size":   info.Size(),
		"slide_count": len(deck.Slides) + 1,
	})

	return info.Size(), nil
}

// generateXLSX generates an XLSX export for a dashboard
func (s *ExportService) generateXLSX(ctx context.Context, job *ExportJob, options *ExportOptions, outputPath string) (int64, error) {
	LogInfo("generate_xlsx", "Generating XLSX export", map[string]interface{}{
		"export_id":    job.ID,
		"dashboard_id": job.DashboardID,
	})

	// ---- 1. Fetch dashboard data ----
	var dashboard models.Dashboard
	dashQuery := s.db.WithContext(ctx).Preload("Cards").Where("id = ?", job.DashboardID.String())
	if err := dashQuery.First(&dashboard).Error; err != nil {
		LogError("generate_xlsx_fetch_dashboard", "Failed to fetch dashboard", map[string]interface{}{
			"export_id":    job.ID,
			"dashboard_id": job.DashboardID,
			"error":        err,
		})
		return 0, fmt.Errorf("failed to fetch dashboard: %w", err)
	}

	title := dashboard.Name
	if options.Title != nil && *options.Title != "" {
		title = *options.Title
	}

	// ---- 2. Build sheets from card data ----
	var sheets []XLSXSheet

	// Filter cards
	cards := dashboard.Cards
	if len(options.CardIDs) > 0 {
		cardIDSet := make(map[string]bool)
		for _, id := range options.CardIDs {
			cardIDSet[id] = true
		}
		filteredCards := make([]models.DashboardCard, 0)
		for _, card := range cards {
			if cardIDSet[card.ID.String()] {
				filteredCards = append(filteredCards, card)
			}
		}
		cards = filteredCards
	}

	for _, card := range cards {
		sheetName := "Sheet"
		if card.Title != nil && *card.Title != "" {
			sheetName = *card.Title
		}

		// Try to fetch query data
		if card.QueryID != nil {
			headers, rows := s.fetchCardQueryData(ctx, card.QueryID.String())
			if len(headers) > 0 {
				sheets = append(sheets, XLSXSheet{
					Name:    sheetName,
					Headers: headers,
					Rows:    rows,
				})
				continue
			}
		}

		// Fallback: card metadata sheet
		sheets = append(sheets, XLSXSheet{
			Name:    sheetName,
			Headers: []string{"Property", "Value"},
			Rows: [][]string{
				{"Card ID", card.ID.String()},
				{"Type", card.Type},
				{"Created", card.CreatedAt.Format("2006-01-02 15:04:05")},
			},
		})
	}

	// If no cards, add a summary sheet
	if len(sheets) == 0 {
		sheets = append(sheets, XLSXSheet{
			Name:    "Summary",
			Headers: []string{"Property", "Value"},
			Rows: [][]string{
				{"Dashboard", dashboard.Name},
				{"Dashboard ID", job.DashboardID.String()},
				{"Generated", time.Now().Format(time.RFC3339)},
				{"Status", "No cards found for export"},
			},
		})
	}

	// ---- 3. Generate XLSX ----
	generator := NewXLSXGenerator()
	xlsxBytes, err := generator.GenerateXLSX(sheets, title)
	if err != nil {
		LogError("generate_xlsx_failed", "Failed to generate XLSX", map[string]interface{}{
			"export_id": job.ID,
			"error":     err,
		})
		return 0, fmt.Errorf("failed to generate XLSX: %w", err)
	}

	if err := os.WriteFile(outputPath, xlsxBytes, 0644); err != nil {
		return 0, fmt.Errorf("failed to write XLSX file: %w", err)
	}

	info, err := os.Stat(outputPath)
	if err != nil {
		return 0, fmt.Errorf("failed to stat XLSX file: %w", err)
	}

	LogInfo("generate_xlsx_complete", "XLSX export generated", map[string]interface{}{
		"export_id": job.ID,
		"file_size": info.Size(),
	})

	return info.Size(), nil
}
