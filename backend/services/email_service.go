package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"mime"
	"mime/multipart"
	"net/smtp"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"time"

	"insight-engine-backend/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EmailConfig holds email service configuration
type EmailConfig struct {
	Provider   string // "smtp", "console", "mock"
	SMTPHost   string
	SMTPPort   string
	SMTPUser   string
	SMTPPass   string
	FromEmail  string
	FromName   string
	AppBaseURL string // Frontend URL for links
	MaxRetries int
}

// EmailService handles email sending operations
type EmailService struct {
	config EmailConfig
	db     *gorm.DB
	Mock   bool // If true, emails are not sent but logged
}

// NewEmailService creates a new email service instance
func NewEmailService() *EmailService {
	config := EmailConfig{
		Provider:   getEnvOrDefault("EMAIL_PROVIDER", "console"),
		SMTPHost:   getEnvOrDefault("SMTP_HOST", "localhost"),
		SMTPPort:   getEnvOrDefault("SMTP_PORT", "587"),
		SMTPUser:   os.Getenv("SMTP_USER"),
		SMTPPass:   os.Getenv("SMTP_PASS"),
		FromEmail:  getEnvOrDefault("FROM_EMAIL", "noreply@insightengine.local"),
		FromName:   getEnvOrDefault("FROM_NAME", "InsightEngine"),
		AppBaseURL: getEnvOrDefault("APP_BASE_URL", "http://localhost:3000"),
		MaxRetries: 3,
	}

	return &EmailService{config: config}
}

// NewEmailServiceWithDB creates a new email service instance with database connection
func NewEmailServiceWithDB(db *gorm.DB) *EmailService {
	service := NewEmailService()
	service.db = db
	return service
}

// GetProvider returns the configured email provider
func (s *EmailService) GetProvider() string {
	return s.config.Provider
}

// getEnvOrDefault returns environment variable or default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GenerateVerificationToken creates a cryptographically secure verification token
func (s *EmailService) GenerateVerificationToken() string {
	return uuid.New().String()
}

// SendVerificationEmail sends email verification email to user
func (s *EmailService) SendVerificationEmail(toEmail, toName, token string) error {
	verificationURL := fmt.Sprintf("%s/auth/verify-email?token=%s", s.config.AppBaseURL, token)

	if s.Mock {
		return nil
	}

	subject := "Verify Your Email - InsightEngine"
	body := s.buildVerificationEmail(toName, verificationURL)

	return s.sendEmail(toEmail, subject, body)
}

// buildVerificationEmail creates HTML email content
func (s *EmailService) buildVerificationEmail(userName, verificationURL string) string {
	emailTemplate := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Verify Your Email</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f4f4f4;
        }
        .container {
            background-color: #ffffff;
            border-radius: 8px;
            padding: 40px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .header {
            text-align: center;
            margin-bottom: 30px;
        }
        .logo {
            font-size: 32px;
            font-weight: bold;
            color: #4F46E5;
            margin-bottom: 10px;
        }
        h1 {
            color: #1a1a1a;
            font-size: 24px;
            margin-bottom: 20px;
        }
        .button {
            display: inline-block;
            background-color: #4F46E5;
            color: #ffffff;
            text-decoration: none;
            padding: 12px 30px;
            border-radius: 6px;
            font-weight: 600;
            margin: 20px 0;
        }
        .button:hover {
            background-color: #4338ca;
        }
        .footer {
            margin-top: 30px;
            padding-top: 20px;
            border-top: 1px solid #e5e5e5;
            text-align: center;
            color: #666;
            font-size: 14px;
        }
        .expires {
            color: #dc2626;
            font-size: 14px;
            margin-top: 20px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="logo">InsightEngine</div>
        </div>
        
        <h1>Welcome, {{.UserName}}!</h1>
        
        <p>Thank you for creating an account with InsightEngine. To complete your registration and start using our AI-powered analytics platform, please verify your email address.</p>
        
        <div style="text-align: center;">
            <a href="{{.VerificationURL}}" class="button">Verify Email Address</a>
        </div>
        
        <p style="margin-top: 20px;">Or copy and paste this link into your browser:</p>
        <p style="word-break: break-all; color: #4F46E5;">{{.VerificationURL}}</p>
        
        <p class="expires">
            <strong>Note:</strong> This verification link will expire in 24 hours.
        </p>
        
        <div class="footer">
            <p>If you didn't create an account with InsightEngine, you can safely ignore this email.</p>
            <p style="margin-top: 10px;">&copy; 2026 InsightEngine. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`

	type TemplateData struct {
		UserName        string
		VerificationURL string
	}

	data := TemplateData{
		UserName:        userName,
		VerificationURL: verificationURL,
	}

	tmpl, err := template.New("verification").Parse(emailTemplate)
	if err != nil {
		return s.buildSimpleVerificationEmail(userName, verificationURL)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return s.buildSimpleVerificationEmail(userName, verificationURL)
	}

	return buf.String()
}

// buildSimpleVerificationEmail creates plain text fallback email
func (s *EmailService) buildSimpleVerificationEmail(userName, verificationURL string) string {
	return fmt.Sprintf(`Welcome to InsightEngine!

Hi %s,

Thank you for creating an account. Please verify your email address by clicking the link below:

%s

This link will expire in 24 hours.

If you didn't create an account, you can safely ignore this email.

Best regards,
The InsightEngine Team`, userName, verificationURL)
}

// sendEmail sends email based on configured provider
func (s *EmailService) sendEmail(toEmail, subject, body string) error {
	switch s.config.Provider {
	case "smtp":
		return s.sendViaSMTP(toEmail, subject, body, nil)
	case "console", "mock":
		return s.sendViaConsole(toEmail, subject, body)
	default:
		return s.sendViaConsole(toEmail, subject, body)
	}
}

// sendViaSMTP sends email using SMTP
func (s *EmailService) sendViaSMTP(toEmail, subject, body string, attachments []models.EmailAttachment) error {
	from := fmt.Sprintf("%s <%s>", s.config.FromName, s.config.FromEmail)

	var message bytes.Buffer

	// If there are attachments, use multipart/mixed
	if len(attachments) > 0 {
		writer := multipart.NewWriter(&message)

		// Headers
		message.WriteString(fmt.Sprintf("From: %s\r\n", from))
		message.WriteString(fmt.Sprintf("To: %s\r\n", toEmail))
		message.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
		message.WriteString("MIME-Version: 1.0\r\n")
		message.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n", writer.Boundary()))
		message.WriteString("\r\n")

		// HTML part
		htmlPart := textproto.MIMEHeader{}
		htmlPart.Set("Content-Type", "text/html; charset=UTF-8")
		htmlPart.Set("Content-Transfer-Encoding", "quoted-printable")
		part, _ := writer.CreatePart(htmlPart)
		part.Write([]byte(body))

		// Attachments
		for _, attachment := range attachments {
			attachmentPart := textproto.MIMEHeader{}
			contentType := attachment.ContentType
			if contentType == "" {
				contentType = "application/octet-stream"
			}
			attachmentPart.Set("Content-Type", contentType)
			attachmentPart.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", attachment.Filename))
			attachmentPart.Set("Content-Transfer-Encoding", "base64")

			part, _ := writer.CreatePart(attachmentPart)

			// Read file and encode to base64
			fileData, err := os.ReadFile(attachment.FilePath)
			if err != nil {
				LogError("email_attachment", "Failed to read attachment", map[string]interface{}{"file": attachment.FilePath, "error": err})
				continue
			}

			encoded := make([]byte, base64.StdEncoding.EncodedLen(len(fileData)))
			base64.StdEncoding.Encode(encoded, fileData)

			// Write in 76-character lines
			for i := 0; i < len(encoded); i += 76 {
				end := i + 76
				if end > len(encoded) {
					end = len(encoded)
				}
				part.Write(encoded[i:end])
				part.Write([]byte("\r\n"))
			}
		}

		writer.Close()
	} else {
		// Simple HTML email without attachments
		headers := make(map[string]string)
		headers["From"] = from
		headers["To"] = toEmail
		headers["Subject"] = subject
		headers["MIME-Version"] = "1.0"
		headers["Content-Type"] = "text/html; charset=UTF-8"

		for k, v := range headers {
			message.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
		}
		message.WriteString("\r\n")
		message.WriteString(body)
	}

	// Authentication
	auth := smtp.PlainAuth("", s.config.SMTPUser, s.config.SMTPPass, s.config.SMTPHost)

	// Send email
	addr := fmt.Sprintf("%s:%s", s.config.SMTPHost, s.config.SMTPPort)
	err := smtp.SendMail(addr, auth, s.config.FromEmail, []string{toEmail}, message.Bytes())
	if err != nil {
		return fmt.Errorf("failed to send email via SMTP: %w", err)
	}

	return nil
}

// sendViaConsole logs email to console (for development)
func (s *EmailService) sendViaConsole(toEmail, subject, body string) error {
	LogInfo("email_console_mode", "Email sent via console (dev mode)", map[string]interface{}{
		"to":      toEmail,
		"from":    fmt.Sprintf("%s <%s>", s.config.FromName, s.config.FromEmail),
		"subject": subject,
		"body":    body,
	})
	return nil
}

// GetVerificationExpiry returns the expiration time for verification tokens
func (s *EmailService) GetVerificationExpiry() time.Time {
	return time.Now().Add(24 * time.Hour) // 24 hours expiration
}

// SendPasswordResetEmail sends password reset email to user
func (s *EmailService) SendPasswordResetEmail(toEmail, toName, token string) error {
	resetURL := fmt.Sprintf("%s/auth/reset-password?token=%s", s.config.AppBaseURL, token)

	subject := "Reset Your Password - InsightEngine"
	body := s.buildPasswordResetEmail(toName, resetURL)

	return s.sendEmail(toEmail, subject, body)
}

// buildPasswordResetEmail creates HTML email content for password reset
func (s *EmailService) buildPasswordResetEmail(userName, resetURL string) string {
	emailTemplate := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Reset Your Password</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f4f4f4;
        }
        .container {
            background-color: #ffffff;
            border-radius: 8px;
            padding: 40px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .header {
            text-align: center;
            margin-bottom: 30px;
        }
        .logo {
            font-size: 32px;
            font-weight: bold;
            color: #4F46E5;
            margin-bottom: 10px;
        }
        h1 {
            color: #1a1a1a;
            font-size: 24px;
            margin-bottom: 20px;
        }
        .button {
            display: inline-block;
            background-color: #4F46E5;
            color: #ffffff;
            text-decoration: none;
            padding: 12px 30px;
            border-radius: 6px;
            font-weight: 600;
            margin: 20px 0;
        }
        .button:hover {
            background-color: #4338ca;
        }
        .footer {
            margin-top: 30px;
            padding-top: 20px;
            border-top: 1px solid #e5e5e5;
            text-align: center;
            color: #666;
            font-size: 14px;
        }
        .expires {
            color: #dc2626;
            font-size: 14px;
            margin-top: 20px;
        }
        .warning {
            background-color: #fef3c7;
            border-left: 4px solid #f59e0b;
            padding: 12px;
            margin: 20px 0;
            font-size: 14px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="logo">InsightEngine</div>
        </div>
        
        <h1>Password Reset Request</h1>
        
        <p>Hi {{.UserName}},</p>
        
        <p>We received a request to reset your password for your InsightEngine account. If you made this request, click the button below to reset your password:</p>
        
        <div style="text-align: center;">
            <a href="{{.ResetURL}}" class="button">Reset Password</a>
        </div>
        
        <p style="margin-top: 20px;">Or copy and paste this link into your browser:</p>
        <p style="word-break: break-all; color: #4F46E5;">{{.ResetURL}}</p>
        
        <div class="warning">
            <strong>Security Notice:</strong> This link will expire in 1 hour. If you didn't request a password reset, please ignore this email or contact support if you have concerns.
        </div>
        
        <div class="footer">
            <p>If you didn't request a password reset, you can safely ignore this email.</p>
            <p style="margin-top: 10px;">&copy; 2026 InsightEngine. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`

	type TemplateData struct {
		UserName string
		ResetURL string
	}

	data := TemplateData{
		UserName: userName,
		ResetURL: resetURL,
	}

	tmpl, err := template.New("password-reset").Parse(emailTemplate)
	if err != nil {
		return s.buildSimplePasswordResetEmail(userName, resetURL)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return s.buildSimplePasswordResetEmail(userName, resetURL)
	}

	return buf.String()
}

// buildSimplePasswordResetEmail creates plain text fallback email
func (s *EmailService) buildSimplePasswordResetEmail(userName, resetURL string) string {
	return fmt.Sprintf(`Password Reset Request - InsightEngine

Hi %s,

We received a request to reset your password. If you made this request, click the link below:

%s

This link will expire in 1 hour.

If you didn't request a password reset, please ignore this email.

Best regards,
The InsightEngine Team`, userName, resetURL)
}

// GetPasswordResetExpiry returns the expiration time for password reset tokens
func (s *EmailService) GetPasswordResetExpiry() time.Time {
	return time.Now().Add(1 * time.Hour) // 1 hour expiration
}

// ==================== TASK-098: Enhanced Email Service ====================

// SendReportEmailRequest represents a request to send a report email
type SendReportEmailRequest struct {
	To          []string
	Cc          []string
	Bcc         []string
	Subject     string
	BodyHTML    string
	BodyText    string
	Attachments []ReportAttachment
	TrackOpens  bool
	TrackClicks bool
}

// ReportAttachment represents a report file attachment
type ReportAttachment struct {
	Filename    string
	ContentType string
	FilePath    string
	FileSize    int64
}

// SendReportEmail sends an email with report attachments
func (s *EmailService) SendReportEmail(req *SendReportEmailRequest) error {
	if len(req.To) == 0 {
		return fmt.Errorf("at least one recipient is required")
	}

	// Convert ReportAttachment to EmailAttachment
	attachments := make([]models.EmailAttachment, len(req.Attachments))
	for i, att := range req.Attachments {
		attachments[i] = models.EmailAttachment{
			Filename:    att.Filename,
			ContentType: att.ContentType,
			FilePath:    att.FilePath,
			FileSize:    att.FileSize,
		}
	}

	// Build recipient string
	toStr := strings.Join(req.To, ", ")

	// Send via SMTP with attachments
	if s.config.Provider == "smtp" {
		return s.sendViaSMTP(toStr, req.Subject, req.BodyHTML, attachments)
	}

	// Console mode - log the email
	LogInfo("email_report_sent", "Report email sent", map[string]interface{}{
		"to":          req.To,
		"subject":     req.Subject,
		"attachments": len(req.Attachments),
	})
	return nil
}

// QueueEmail adds an email to the queue for async sending
func (s *EmailService) QueueEmail(req *SendReportEmailRequest, priority int) (*models.EmailQueue, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database connection required for email queuing")
	}

	// Convert attachments
	attachments := make([]models.EmailAttachment, len(req.Attachments))
	for i, att := range req.Attachments {
		attachments[i] = models.EmailAttachment{
			Filename:    att.Filename,
			ContentType: att.ContentType,
			FilePath:    att.FilePath,
			FileSize:    att.FileSize,
		}
	}

	// Build recipient strings
	toStr := strings.Join(req.To, ", ")
	var ccStr, bccStr string
	if len(req.Cc) > 0 {
		ccStr = strings.Join(req.Cc, ", ")
	}
	if len(req.Bcc) > 0 {
		bccStr = strings.Join(req.Bcc, ", ")
	}

	// Create queue entry
	emailQueue := &models.EmailQueue{
		ID:          uuid.New(),
		Status:      models.EmailStatusPending,
		Priority:    priority,
		To:          toStr,
		Cc:          &ccStr,
		Bcc:         &bccStr,
		FromEmail:   s.config.FromEmail,
		FromName:    s.config.FromName,
		Subject:     req.Subject,
		BodyHTML:    &req.BodyHTML,
		BodyText:    &req.BodyText,
		TrackOpens:  req.TrackOpens,
		TrackClicks: req.TrackClicks,
	}

	if err := emailQueue.SetAttachments(attachments); err != nil {
		return nil, fmt.Errorf("failed to set attachments: %w", err)
	}

	if err := s.db.Create(emailQueue).Error; err != nil {
		return nil, fmt.Errorf("failed to queue email: %w", err)
	}

	// Log the queue event
	log := &models.EmailLog{
		ID:           uuid.New(),
		EmailQueueID: emailQueue.ID,
		Event:        models.EmailEventQueued,
		Status:       models.EmailStatusPending,
		Message:      stringPtr("Email queued for sending"),
	}
	s.db.Create(log)

	return emailQueue, nil
}

// QueueBatchEmail queues a batch of emails to multiple recipients
func (s *EmailService) QueueBatchEmail(recipients []string, subject, bodyHTML, bodyText string, attachments []ReportAttachment, priority int) (*models.EmailBatch, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database connection required for batch email")
	}

	// Create batch
	batch := &models.EmailBatch{
		ID:           uuid.New(),
		Name:         fmt.Sprintf("Batch_%s", uuid.New().String()[:8]),
		TotalCount:   len(recipients),
		PendingCount: len(recipients),
		Status:       "pending",
	}

	if err := s.db.Create(batch).Error; err != nil {
		return nil, fmt.Errorf("failed to create batch: %w", err)
	}

	// Queue individual emails
	for _, recipient := range recipients {
		req := &SendReportEmailRequest{
			To:          []string{recipient},
			Subject:     subject,
			BodyHTML:    bodyHTML,
			BodyText:    bodyText,
			Attachments: attachments,
		}

		emailQueue, err := s.QueueEmail(req, priority)
		if err != nil {
			LogError("batch_email_queue", "Failed to queue email", map[string]interface{}{
				"recipient": recipient,
				"error":     err,
			})
			continue
		}

		// Update batch ID
		emailQueue.BatchID = &batch.ID
		emailQueue.IsBulk = true
		s.db.Save(emailQueue)
	}

	return batch, nil
}

// ProcessEmailQueue processes pending emails from the queue
func (s *EmailService) ProcessEmailQueue(ctx context.Context, batchSize int) error {
	if s.db == nil {
		return fmt.Errorf("database connection required")
	}

	var emails []models.EmailQueue
	if err := s.db.Where("status = ? AND (scheduled_at IS NULL OR scheduled_at <= ?)",
		models.EmailStatusPending, time.Now()).
		Order("priority ASC, created_at ASC").
		Limit(batchSize).
		Find(&emails).Error; err != nil {
		return fmt.Errorf("failed to fetch pending emails: %w", err)
	}

	for _, email := range emails {
		if err := s.processQueuedEmail(ctx, &email); err != nil {
			LogError("email_queue_process", "Failed to process queued email", map[string]interface{}{
				"email_id": email.ID,
				"error":    err,
			})
		}
	}

	return nil
}

// processQueuedEmail processes a single queued email
func (s *EmailService) processQueuedEmail(ctx context.Context, email *models.EmailQueue) error {
	// Update status to sending
	email.Status = models.EmailStatusSending
	s.db.Save(email)

	// Parse attachments
	attachments, err := email.GetAttachments()
	if err != nil {
		email.Status = models.EmailStatusFailed
		email.LastError = stringPtr(fmt.Sprintf("Failed to parse attachments: %v", err))
		s.db.Save(email)
		return err
	}

	// Build SendReportEmailRequest
	req := &SendReportEmailRequest{
		To:          []string{email.To},
		Subject:     email.Subject,
		BodyHTML:    *email.BodyHTML,
		BodyText:    *email.BodyText,
		Attachments: make([]ReportAttachment, len(attachments)),
	}

	for i, att := range attachments {
		req.Attachments[i] = ReportAttachment{
			Filename:    att.Filename,
			ContentType: att.ContentType,
			FilePath:    att.FilePath,
			FileSize:    att.FileSize,
		}
	}

	// Send the email
	if err := s.SendReportEmail(req); err != nil {
		email.RetryCount++

		if email.RetryCount >= email.MaxRetries {
			email.Status = models.EmailStatusFailed
			email.LastError = stringPtr(err.Error())

			// Log failure
			log := &models.EmailLog{
				ID:           uuid.New(),
				EmailQueueID: email.ID,
				Event:        models.EmailEventFailed,
				Status:       models.EmailStatusFailed,
				Message:      stringPtr(err.Error()),
			}
			s.db.Create(log)
		} else {
			email.Status = models.EmailStatusPending
			email.LastError = stringPtr(err.Error())

			// Log retry
			log := &models.EmailLog{
				ID:           uuid.New(),
				EmailQueueID: email.ID,
				Event:        models.EmailEventRetry,
				Status:       models.EmailStatusPending,
				Message:      stringPtr(fmt.Sprintf("Retry %d/%d: %v", email.RetryCount, email.MaxRetries, err)),
			}
			s.db.Create(log)
		}

		s.db.Save(email)
		return err
	}

	// Success
	now := time.Now()
	email.Status = models.EmailStatusSent
	email.SentAt = &now
	s.db.Save(email)

	// Log success
	log := &models.EmailLog{
		ID:           uuid.New(),
		EmailQueueID: email.ID,
		Event:        models.EmailEventSent,
		Status:       models.EmailStatusSent,
		Message:      stringPtr("Email sent successfully"),
	}
	s.db.Create(log)

	// Update batch stats if part of batch
	if email.BatchID != nil {
		s.updateBatchStats(*email.BatchID)
	}

	return nil
}

// updateBatchStats updates the statistics for an email batch
func (s *EmailService) updateBatchStats(batchID uuid.UUID) {
	var counts struct {
		Sent    int64
		Failed  int64
		Pending int64
	}

	s.db.Model(&models.EmailQueue{}).Where("batch_id = ? AND status = ?", batchID, models.EmailStatusSent).Count(&counts.Sent)
	s.db.Model(&models.EmailQueue{}).Where("batch_id = ? AND status = ?", batchID, models.EmailStatusFailed).Count(&counts.Failed)
	s.db.Model(&models.EmailQueue{}).Where("batch_id = ? AND status = ?", batchID, models.EmailStatusPending).Count(&counts.Pending)

	s.db.Model(&models.EmailBatch{}).Where("id = ?", batchID).Updates(map[string]interface{}{
		"sent_count":    counts.Sent,
		"failed_count":  counts.Failed,
		"pending_count": counts.Pending,
	})

	// Check if batch is complete
	if counts.Pending == 0 {
		now := time.Now()
		s.db.Model(&models.EmailBatch{}).Where("id = ?", batchID).Updates(map[string]interface{}{
			"status":       "completed",
			"completed_at": now,
		})
	}
}

// GetEmailTemplate retrieves an email template by name
func (s *EmailService) GetEmailTemplate(name string) (*models.EmailTemplate, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database connection required")
	}

	var template models.EmailTemplate
	if err := s.db.Where("name = ? AND is_active = ?", name, true).First(&template).Error; err != nil {
		return nil, err
	}

	return &template, nil
}

// RenderEmailTemplate renders an email template with data
func (s *EmailService) RenderEmailTemplate(tmpl *models.EmailTemplate, data map[string]interface{}) (subject, bodyHTML, bodyText string, err error) {
	// Render subject
	subjectTmpl, err := template.New("subject").Parse(tmpl.Subject)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to parse subject template: %w", err)
	}

	var subjectBuf bytes.Buffer
	if err := subjectTmpl.Execute(&subjectBuf, data); err != nil {
		return "", "", "", fmt.Errorf("failed to render subject: %w", err)
	}
	subject = subjectBuf.String()

	// Render HTML body
	if tmpl.BodyHTML != "" {
		htmlTmpl, err := template.New("html").Parse(tmpl.BodyHTML)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to parse HTML template: %w", err)
		}

		var htmlBuf bytes.Buffer
		if err := htmlTmpl.Execute(&htmlBuf, data); err != nil {
			return "", "", "", fmt.Errorf("failed to render HTML: %w", err)
		}
		bodyHTML = htmlBuf.String()
	}

	// Render text body
	if tmpl.BodyText != "" {
		textTmpl, err := template.New("text").Parse(tmpl.BodyText)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to parse text template: %w", err)
		}

		var textBuf bytes.Buffer
		if err := textTmpl.Execute(&textBuf, data); err != nil {
			return "", "", "", fmt.Errorf("failed to render text: %w", err)
		}
		bodyText = textBuf.String()
	}

	return subject, bodyHTML, bodyText, nil
}

// CreateTrackingPixel creates a tracking pixel for open tracking
func (s *EmailService) CreateTrackingPixel(emailQueueID uuid.UUID) (string, error) {
	if s.db == nil {
		return "", fmt.Errorf("database connection required")
	}

	token := uuid.New().String()
	pixel := &models.EmailTrackingPixel{
		ID:           uuid.New(),
		EmailQueueID: emailQueueID,
		Token:        token,
	}

	if err := s.db.Create(pixel).Error; err != nil {
		return "", err
	}

	// Return tracking pixel URL
	return fmt.Sprintf("%s/api/email/track/open?token=%s", s.config.AppBaseURL, token), nil
}

// RecordEmailOpen records when an email is opened
func (s *EmailService) RecordEmailOpen(token, ipAddress, userAgent string) error {
	if s.db == nil {
		return fmt.Errorf("database connection required")
	}

	var pixel models.EmailTrackingPixel
	if err := s.db.Where("token = ?", token).First(&pixel).Error; err != nil {
		return err
	}

	now := time.Now()

	updates := map[string]interface{}{
		"open_count":     gorm.Expr("open_count + 1"),
		"last_opened_at": now,
		"ip_address":     ipAddress,
		"user_agent":     userAgent,
	}

	if pixel.FirstOpenedAt == nil {
		updates["first_opened_at"] = now
	}

	if err := s.db.Model(&pixel).Updates(updates).Error; err != nil {
		return err
	}

	// Update email queue status
	s.db.Model(&models.EmailQueue{}).Where("id = ?", pixel.EmailQueueID).Update("status", models.EmailStatusOpened)

	// Log the event
	log := &models.EmailLog{
		ID:           uuid.New(),
		EmailQueueID: pixel.EmailQueueID,
		Event:        models.EmailEventOpened,
		Status:       models.EmailStatusOpened,
		IPAddress:    &ipAddress,
		UserAgent:    &userAgent,
	}

	return s.db.Create(log).Error
}

// GetEmailStats returns statistics for an email
func (s *EmailService) GetEmailStats(emailQueueID uuid.UUID) (map[string]interface{}, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database connection required")
	}

	var queue models.EmailQueue
	if err := s.db.First(&queue, "id = ?", emailQueueID).Error; err != nil {
		return nil, err
	}

	var logs []models.EmailLog
	if err := s.db.Where("email_queue_id = ?", emailQueueID).Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"emailId":     queue.ID,
		"status":      queue.Status,
		"sentAt":      queue.SentAt,
		"deliveredAt": queue.DeliveredAt,
		"openedAt":    queue.OpenedAt,
		"openCount":   queue.OpenCount,
		"clickCount":  queue.ClickCount,
		"events":      logs,
	}

	return stats, nil
}

// Helper function
func stringPtr(s string) *string {
	return &s
}

// ReadFileContent reads file content for attachment
func (s *EmailService) ReadFileContent(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return io.ReadAll(file)
}

// GetContentType determines the content type based on file extension
func (s *EmailService) GetContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".csv":
		return "text/csv"
	case ".xlsx", ".xls":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".txt":
		return "text/plain"
	case ".html", ".htm":
		return "text/html"
	case ".json":
		return "application/json"
	case ".xml":
		return "application/xml"
	case ".zip":
		return "application/zip"
	default:
		return mime.TypeByExtension(ext)
	}
}

// CleanupOldEmails removes old sent/failed emails from the queue
func (s *EmailService) CleanupOldEmails(olderThan time.Duration) error {
	if s.db == nil {
		return fmt.Errorf("database connection required")
	}

	cutoff := time.Now().Add(-olderThan)

	result := s.db.Where("(status IN ? OR status IN ?) AND updated_at < ?",
		[]models.EmailQueueStatus{models.EmailStatusSent, models.EmailStatusDelivered, models.EmailStatusOpened},
		[]models.EmailQueueStatus{models.EmailStatusFailed, models.EmailStatusBounced},
		cutoff).
		Delete(&models.EmailQueue{})

	if result.Error != nil {
		return result.Error
	}

	LogInfo("email_cleanup", "Cleaned up old emails", map[string]interface{}{
		"deleted": result.RowsAffected,
	})

	return nil
}
