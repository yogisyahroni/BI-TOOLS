-- Migration: 020_enhance_alerts.sql
-- Description: Enhance alerts table with new columns and create supporting tables for TASK-101
-- Author: InsightEngine Team
-- Date: 2026-02-10

-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================
-- Migrate existing Alert table to new structure
-- ============================================

-- First, rename existing table as backup
ALTER TABLE IF EXISTS "Alert" RENAME TO alerts_backup;

-- ============================================
-- Create Enhanced Alerts Table
-- ============================================
CREATE TABLE IF NOT EXISTS alerts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    query_id UUID NOT NULL,
    user_id UUID NOT NULL,
    
    -- Condition configuration
    column_name VARCHAR(255) NOT NULL,  -- Renamed from 'column' to avoid SQL keyword issues
    operator VARCHAR(10) NOT NULL CHECK (operator IN ('>', '<', '=', '>=', '<=', '!=')),
    threshold FLOAT NOT NULL,
    
    -- Schedule configuration
    schedule VARCHAR(50) NOT NULL,
    timezone VARCHAR(100) DEFAULT 'UTC',
    
    -- Status and metadata
    is_active BOOLEAN DEFAULT TRUE,
    severity VARCHAR(20) DEFAULT 'warning' CHECK (severity IN ('critical', 'warning', 'info')),
    state VARCHAR(20) DEFAULT 'ok' CHECK (state IN ('ok', 'triggered', 'acknowledged', 'muted', 'error')),
    
    -- Runtime tracking
    last_run_at TIMESTAMP WITH TIME ZONE,
    last_status VARCHAR(20),
    last_value FLOAT,
    last_error TEXT,
    next_run_at TIMESTAMP WITH TIME ZONE,
    
    -- Cooldown and throttling
    cooldown_minutes INTEGER DEFAULT 5,
    last_triggered_at TIMESTAMP WITH TIME ZONE,
    trigger_count INTEGER DEFAULT 0,
    notification_count INTEGER DEFAULT 0,
    
    -- Mute configuration
    is_muted BOOLEAN DEFAULT FALSE,
    muted_until TIMESTAMP WITH TIME ZONE,
    mute_duration INTEGER,
    
    -- Notification configuration
    channels JSONB DEFAULT '[]'::jsonb,
    
    -- Legacy fields for backward compatibility
    email VARCHAR(255),
    webhook_url TEXT,
    webhook_headers JSONB,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for alerts
CREATE INDEX IF NOT EXISTS idx_alerts_query_id ON alerts(query_id);
CREATE INDEX IF NOT EXISTS idx_alerts_user_id ON alerts(user_id);
CREATE INDEX IF NOT EXISTS idx_alerts_is_active ON alerts(is_active);
CREATE INDEX IF NOT EXISTS idx_alerts_state ON alerts(state);
CREATE INDEX IF NOT EXISTS idx_alerts_severity ON alerts(severity);
CREATE INDEX IF NOT EXISTS idx_alerts_next_run ON alerts(next_run_at);
CREATE INDEX IF NOT EXISTS idx_alerts_is_muted ON alerts(is_muted);

-- ============================================
-- Create Alert History Table
-- ============================================
CREATE TABLE IF NOT EXISTS alert_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    alert_id UUID NOT NULL REFERENCES alerts(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL CHECK (status IN ('triggered', 'ok', 'error')),
    value FLOAT,
    threshold FLOAT NOT NULL,
    message TEXT,
    error_message TEXT,
    query_duration INTEGER DEFAULT 0,  -- milliseconds
    checked_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for alert_history
CREATE INDEX IF NOT EXISTS idx_alert_history_alert_id ON alert_history(alert_id);
CREATE INDEX IF NOT EXISTS idx_alert_history_status ON alert_history(status);
CREATE INDEX IF NOT EXISTS idx_alert_history_checked_at ON alert_history(checked_at);

-- ============================================
-- Create Alert Notification Channels Table
-- ============================================
CREATE TABLE IF NOT EXISTS alert_notification_channels (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    alert_id UUID NOT NULL REFERENCES alerts(id) ON DELETE CASCADE,
    channel_type VARCHAR(20) NOT NULL CHECK (channel_type IN ('email', 'webhook', 'in_app', 'slack')),
    is_enabled BOOLEAN DEFAULT TRUE,
    config JSONB DEFAULT '{}'::jsonb,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for alert_notification_channels
CREATE INDEX IF NOT EXISTS idx_alert_channels_alert_id ON alert_notification_channels(alert_id);
CREATE INDEX IF NOT EXISTS idx_alert_channels_type ON alert_notification_channels(channel_type);
CREATE INDEX IF NOT EXISTS idx_alert_channels_enabled ON alert_notification_channels(is_enabled);

-- ============================================
-- Create Alert Notification Logs Table
-- ============================================
CREATE TABLE IF NOT EXISTS alert_notification_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    history_id UUID NOT NULL REFERENCES alert_history(id) ON DELETE CASCADE,
    channel_type VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL,
    recipient VARCHAR(255),
    error TEXT,
    sent_at TIMESTAMP WITH TIME ZONE,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for alert_notification_logs
CREATE INDEX IF NOT EXISTS idx_alert_logs_history_id ON alert_notification_logs(history_id);
CREATE INDEX IF NOT EXISTS idx_alert_logs_status ON alert_notification_logs(status);

-- ============================================
-- Create Alert Acknowledgments Table
-- ============================================
CREATE TABLE IF NOT EXISTS alert_acknowledgments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    alert_id UUID NOT NULL REFERENCES alerts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    note TEXT,
    acknowledged_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for alert_acknowledgments
CREATE INDEX IF NOT EXISTS idx_alert_ack_alert_id ON alert_acknowledgments(alert_id);
CREATE INDEX IF NOT EXISTS idx_alert_ack_user_id ON alert_acknowledgments(user_id);
CREATE INDEX IF NOT EXISTS idx_alert_ack_time ON alert_acknowledgments(acknowledged_at);

-- ============================================
-- Create Alert Conditions Table (for multiple conditions)
-- ============================================
CREATE TABLE IF NOT EXISTS alert_conditions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    alert_id UUID NOT NULL REFERENCES alerts(id) ON DELETE CASCADE,
    column_name VARCHAR(255) NOT NULL,
    operator VARCHAR(10) NOT NULL,
    threshold FLOAT NOT NULL,
    logic VARCHAR(10) DEFAULT 'AND' CHECK (logic IN ('AND', 'OR')),
    sort_order INTEGER DEFAULT 0,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for alert_conditions
CREATE INDEX IF NOT EXISTS idx_alert_conditions_alert_id ON alert_conditions(alert_id);

-- ============================================
-- Create Alert Notification Templates Table
-- ============================================
CREATE TABLE IF NOT EXISTS alert_notification_templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    content TEXT NOT NULL,
    channel_type VARCHAR(20) NOT NULL CHECK (channel_type IN ('email', 'webhook', 'in_app', 'slack')),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for alert_notification_templates
CREATE INDEX IF NOT EXISTS idx_alert_templates_name ON alert_notification_templates(name);
CREATE INDEX IF NOT EXISTS idx_alert_templates_channel ON alert_notification_templates(channel_type);
CREATE INDEX IF NOT EXISTS idx_alert_templates_active ON alert_notification_templates(is_active);

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
DROP TRIGGER IF EXISTS update_alerts_updated_at ON alerts;
CREATE TRIGGER update_alerts_updated_at
    BEFORE UPDATE ON alerts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_alert_channels_updated_at ON alert_notification_channels;
CREATE TRIGGER update_alert_channels_updated_at
    BEFORE UPDATE ON alert_notification_channels
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_alert_templates_updated_at ON alert_notification_templates;
CREATE TRIGGER update_alert_templates_updated_at
    BEFORE UPDATE ON alert_notification_templates
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_alert_conditions_updated_at ON alert_conditions;
CREATE TRIGGER update_alert_conditions_updated_at
    BEFORE UPDATE ON alert_conditions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- Insert Default Templates
-- ============================================

-- Default Email Template
INSERT INTO alert_notification_templates (name, description, content, channel_type, is_active)
VALUES (
    'default_email',
    'Default email template for alert notifications',
    '<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Alert Notification</title>
</head>
<body>
    <h1>{{AlertName}}</h1>
    <p>Severity: {{AlertSeverity}}</p>
    <p>Condition: {{Column}} {{Operator}} {{Threshold}}</p>
    <p>Current Value: {{Value}}</p>
    <p>Time: {{Timestamp}}</p>
</body>
</html>',
    'email',
    TRUE
)
ON CONFLICT (name) DO NOTHING;

-- Default Webhook Template
INSERT INTO alert_notification_templates (name, description, content, channel_type, is_active)
VALUES (
    'default_webhook',
    'Default webhook template for alert notifications',
    '{"alert": {"name": "{{AlertName}}", "severity": "{{AlertSeverity}}", "value": {{Value}}, "threshold": {{Threshold}}}}',
    'webhook',
    TRUE
)
ON CONFLICT (name) DO NOTHING;

-- Default Slack Template
INSERT INTO alert_notification_templates (name, description, content, channel_type, is_active)
VALUES (
    'default_slack',
    'Default Slack template for alert notifications',
    '{"text": "Alert: {{AlertName}}\nSeverity: {{AlertSeverity}}\nValue: {{Value}}\nThreshold: {{Threshold}}"}',
    'slack',
    TRUE
)
ON CONFLICT (name) DO NOTHING;

-- ============================================
-- Migrate data from backup table if exists
-- ============================================

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'alerts_backup') THEN
        INSERT INTO alerts (
            id,
            name,
            description,
            query_id,
            user_id,
            column_name,
            operator,
            threshold,
            schedule,
            email,
            webhook_url,
            webhook_headers,
            is_active,
            last_run_at,
            last_status,
            created_at,
            updated_at
        )
        SELECT 
            COALESCE("ID"::uuid, uuid_generate_v4()),
            "Name",
            "Description",
            "QueryId"::uuid,
            "UserId"::uuid,
            "Column",
            "Operator",
            "Threshold",
            "Schedule",
            "Email",
            "WebhookUrl",
            "WebhookHeaders",
            "IsActive",
            "LastRunAt",
            "LastStatus",
            "CreatedAt",
            "UpdatedAt"
        FROM alerts_backup
        ON CONFLICT (id) DO NOTHING;
        
        -- Drop backup table after successful migration
        DROP TABLE alerts_backup;
    END IF;
END $$;

-- ============================================
-- Grant Permissions
-- ============================================
-- GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO insightengine_app;
-- GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO insightengine_app;

-- ============================================
-- Comments
-- ============================================

COMMENT ON TABLE alerts IS 'Enhanced alerts table with state management, cooldown, and notification channels';
COMMENT ON TABLE alert_history IS 'History of alert evaluations';
COMMENT ON TABLE alert_notification_channels IS 'Notification channel configurations per alert';
COMMENT ON TABLE alert_notification_logs IS 'Log of sent notifications';
COMMENT ON TABLE alert_acknowledgments IS 'User acknowledgments of triggered alerts';
COMMENT ON TABLE alert_conditions IS 'Multiple condition support for complex alerts';
COMMENT ON TABLE alert_notification_templates IS 'Reusable notification templates';
