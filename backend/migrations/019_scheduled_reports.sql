-- Migration: 019_scheduled_reports.sql
-- Description: Create scheduled reports tables for TASK-099
-- Author: InsightEngine Team
-- Date: 2026-02-10

-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================
-- Scheduled Reports Table
-- ============================================
CREATE TABLE IF NOT EXISTS scheduled_reports (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Resource
    resource_type VARCHAR(50) NOT NULL,
    resource_id VARCHAR(255) NOT NULL,
    
    -- Schedule
    schedule_type VARCHAR(50) NOT NULL,
    cron_expr VARCHAR(255),
    time_of_day VARCHAR(10),
    day_of_week INTEGER,
    day_of_month INTEGER,
    timezone VARCHAR(100) DEFAULT 'UTC',
    
    -- Format & Options
    format VARCHAR(20) NOT NULL,
    include_filters BOOLEAN DEFAULT FALSE,
    subject TEXT,
    message TEXT,
    options JSONB,
    
    -- Status
    is_active BOOLEAN DEFAULT TRUE,
    last_run_at TIMESTAMP WITH TIME ZONE,
    last_run_status VARCHAR(20),
    last_run_error TEXT,
    next_run_at TIMESTAMP WITH TIME ZONE,
    success_count INTEGER DEFAULT 0,
    failure_count INTEGER DEFAULT 0,
    consecutive_fail INTEGER DEFAULT 0,
    
    -- Ownership
    created_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for scheduled_reports
CREATE INDEX IF NOT EXISTS idx_scheduled_reports_resource ON scheduled_reports(resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_scheduled_reports_created_by ON scheduled_reports(created_by);
CREATE INDEX IF NOT EXISTS idx_scheduled_reports_is_active ON scheduled_reports(is_active);
CREATE INDEX IF NOT EXISTS idx_scheduled_reports_next_run ON scheduled_reports(next_run_at);
CREATE INDEX IF NOT EXISTS idx_scheduled_reports_schedule_type ON scheduled_reports(schedule_type);

-- ============================================
-- Scheduled Report Recipients Table
-- ============================================
CREATE TABLE IF NOT EXISTS scheduled_report_recipients (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    report_id UUID NOT NULL REFERENCES scheduled_reports(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    type VARCHAR(10) DEFAULT 'to',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for scheduled_report_recipients
CREATE INDEX IF NOT EXISTS idx_scheduled_report_recipients_report ON scheduled_report_recipients(report_id);
CREATE INDEX IF NOT EXISTS idx_scheduled_report_recipients_email ON scheduled_report_recipients(email);

-- ============================================
-- Scheduled Report Runs Table (History)
-- ============================================
CREATE TABLE IF NOT EXISTS scheduled_report_runs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    report_id UUID NOT NULL REFERENCES scheduled_reports(id) ON DELETE CASCADE,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL,
    completed_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) NOT NULL,
    error_message TEXT,
    file_url TEXT,
    file_path TEXT,
    file_size BIGINT,
    file_type VARCHAR(50),
    sent_to JSONB,
    send_status JSONB,
    duration_ms BIGINT,
    triggered_by VARCHAR(255),
    ip_address VARCHAR(45),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for scheduled_report_runs
CREATE INDEX IF NOT EXISTS idx_scheduled_report_runs_report ON scheduled_report_runs(report_id);
CREATE INDEX IF NOT EXISTS idx_scheduled_report_runs_status ON scheduled_report_runs(status);
CREATE INDEX IF NOT EXISTS idx_scheduled_report_runs_started ON scheduled_report_runs(started_at);

-- ============================================
-- Email Queue Table (TASK-098)
-- ============================================
CREATE TABLE IF NOT EXISTS email_queue (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    status VARCHAR(20) NOT NULL,
    priority INTEGER DEFAULT 5,
    
    -- Recipient info
    to_email TEXT NOT NULL,
    cc TEXT,
    bcc TEXT,
    from_email TEXT NOT NULL,
    from_name TEXT,
    
    -- Content
    subject TEXT NOT NULL,
    body_html TEXT,
    body_text TEXT,
    template_id UUID,
    template_data JSONB,
    attachments JSONB,
    
    -- Metadata
    track_opens BOOLEAN DEFAULT FALSE,
    track_clicks BOOLEAN DEFAULT FALSE,
    is_bulk BOOLEAN DEFAULT FALSE,
    batch_id UUID,
    
    -- Scheduling
    scheduled_at TIMESTAMP WITH TIME ZONE,
    send_after TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    
    -- Tracking
    sent_at TIMESTAMP WITH TIME ZONE,
    delivered_at TIMESTAMP WITH TIME ZONE,
    opened_at TIMESTAMP WITH TIME ZONE,
    open_count INTEGER DEFAULT 0,
    click_count INTEGER DEFAULT 0,
    last_error TEXT,
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for email_queue
CREATE INDEX IF NOT EXISTS idx_email_queue_status ON email_queue(status);
CREATE INDEX IF NOT EXISTS idx_email_queue_priority ON email_queue(priority);
CREATE INDEX IF NOT EXISTS idx_email_queue_scheduled ON email_queue(scheduled_at);
CREATE INDEX IF NOT EXISTS idx_email_queue_batch ON email_queue(batch_id);

-- ============================================
-- Email Logs Table
-- ============================================
CREATE TABLE IF NOT EXISTS email_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email_queue_id UUID NOT NULL REFERENCES email_queue(id) ON DELETE CASCADE,
    event VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL,
    message TEXT,
    ip_address VARCHAR(45),
    user_agent TEXT,
    smtp_code INTEGER,
    smtp_response TEXT,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for email_logs
CREATE INDEX IF NOT EXISTS idx_email_logs_queue ON email_logs(email_queue_id);
CREATE INDEX IF NOT EXISTS idx_email_logs_event ON email_logs(event);

-- ============================================
-- Email Templates Table
-- ============================================
CREATE TABLE IF NOT EXISTS email_templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    subject TEXT NOT NULL,
    body_html TEXT,
    body_text TEXT,
    variables JSONB,
    category VARCHAR(100),
    is_active BOOLEAN DEFAULT TRUE,
    is_default BOOLEAN DEFAULT FALSE,
    usage_count INTEGER DEFAULT 0,
    created_by VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for email_templates
CREATE INDEX IF NOT EXISTS idx_email_templates_category ON email_templates(category);
CREATE INDEX IF NOT EXISTS idx_email_templates_name ON email_templates(name);

-- ============================================
-- Email Batches Table
-- ============================================
CREATE TABLE IF NOT EXISTS email_batches (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255),
    description TEXT,
    total_count INTEGER DEFAULT 0,
    sent_count INTEGER DEFAULT 0,
    failed_count INTEGER DEFAULT 0,
    pending_count INTEGER DEFAULT 0,
    template_id UUID,
    status VARCHAR(20) DEFAULT 'pending',
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for email_batches
CREATE INDEX IF NOT EXISTS idx_email_batches_status ON email_batches(status);

-- ============================================
-- Email Tracking Pixels Table
-- ============================================
CREATE TABLE IF NOT EXISTS email_tracking_pixels (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email_queue_id UUID NOT NULL UNIQUE REFERENCES email_queue(id) ON DELETE CASCADE,
    token VARCHAR(255) NOT NULL UNIQUE,
    ip_address VARCHAR(45),
    user_agent TEXT,
    open_count INTEGER DEFAULT 0,
    first_opened_at TIMESTAMP WITH TIME ZONE,
    last_opened_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for email_tracking_pixels
CREATE INDEX IF NOT EXISTS idx_email_tracking_pixels_token ON email_tracking_pixels(token);

-- ============================================
-- Email Click Links Table
-- ============================================
CREATE TABLE IF NOT EXISTS email_click_links (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email_queue_id UUID NOT NULL REFERENCES email_queue(id) ON DELETE CASCADE,
    token VARCHAR(255) NOT NULL,
    original_url TEXT NOT NULL,
    click_count INTEGER DEFAULT 0,
    first_clicked_at TIMESTAMP WITH TIME ZONE,
    last_clicked_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for email_click_links
CREATE INDEX IF NOT EXISTS idx_email_click_links_queue ON email_click_links(email_queue_id);
CREATE INDEX IF NOT EXISTS idx_email_click_links_token ON email_click_links(token);

-- ============================================
-- Functions
-- ============================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers to update updated_at
DROP TRIGGER IF EXISTS update_scheduled_reports_updated_at ON scheduled_reports;
CREATE TRIGGER update_scheduled_reports_updated_at
    BEFORE UPDATE ON scheduled_reports
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_scheduled_report_runs_updated_at ON scheduled_report_runs;
CREATE TRIGGER update_scheduled_report_runs_updated_at
    BEFORE UPDATE ON scheduled_report_runs
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_email_queue_updated_at ON email_queue;
CREATE TRIGGER update_email_queue_updated_at
    BEFORE UPDATE ON email_queue
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_email_templates_updated_at ON email_templates;
CREATE TRIGGER update_email_templates_updated_at
    BEFORE UPDATE ON email_templates
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_email_batches_updated_at ON email_batches;
CREATE TRIGGER update_email_batches_updated_at
    BEFORE UPDATE ON email_batches
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- Insert Default Email Templates
-- ============================================

-- Scheduled Report Template
INSERT INTO email_templates (name, description, subject, body_html, body_text, category, is_active, is_default, variables)
VALUES (
    'scheduled_report',
    'Default template for scheduled reports',
    '[Scheduled Report] {{.ReportName}}',
    '<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, ''Segoe UI'', Roboto, sans-serif; line-height: 1.6; color: #333; }
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
            <h1 style="margin: 0; font-size: 20px;">{{.ReportName}}</h1>
        </div>
        <div class="content">
            <p>{{.Message}}</p>
            <p style="margin-top: 20px;">
                <a href="{{.DownloadURL}}" class="button">Download Report</a>
            </p>
            <p style="font-size: 12px; color: #6b7280; margin-top: 20px;">
                This is an automated report from InsightEngine.<br>
                Generated at: {{.GeneratedAt}}
            </p>
        </div>
        <div class="footer">
            <p>&copy; 2026 InsightEngine. All rights reserved.</p>
        </div>
    </div>
</body>
</html>',
    'Report: {{.ReportName}}

{{.Message}}

Download: {{.DownloadURL}}

Generated at: {{.GeneratedAt}}

---
InsightEngine
© 2026 InsightEngine. All rights reserved.',
    'reports',
    TRUE,
    TRUE,
    '["ReportName", "Message", "DownloadURL", "GeneratedAt"]'::jsonb
)
ON CONFLICT (name) DO NOTHING;

-- ============================================
-- Grant Permissions
-- ============================================

-- Grant permissions to application user (adjust username as needed)
-- GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO insightengine_app;
-- GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO insightengine_app;

-- ============================================
-- Comments
-- ============================================

COMMENT ON TABLE scheduled_reports IS 'Scheduled reports configuration table';
COMMENT ON TABLE scheduled_report_recipients IS 'Recipients for scheduled reports';
COMMENT ON TABLE scheduled_report_runs IS 'Execution history for scheduled reports';
COMMENT ON TABLE email_queue IS 'Queue for outgoing emails';
COMMENT ON TABLE email_logs IS 'Event log for email operations';
COMMENT ON TABLE email_templates IS 'Reusable email templates';
COMMENT ON TABLE email_batches IS 'Batch email tracking';
COMMENT ON TABLE email_tracking_pixels IS 'Email open tracking pixels';
COMMENT ON TABLE email_click_links IS 'Email click tracking links';
