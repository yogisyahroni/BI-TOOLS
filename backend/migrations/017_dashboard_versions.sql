-- Migration: Create dashboard_versions table for Dashboard Versioning feature
-- Task: TASK-095 - Dashboard Versioning Backend
-- Created: 2026-02-10

-- Dashboard versions table: Store version history for dashboards
CREATE TABLE IF NOT EXISTS dashboard_versions (
    id TEXT PRIMARY KEY,
    dashboard_id TEXT NOT NULL REFERENCES "Dashboard"(id) ON DELETE CASCADE,
    version INTEGER NOT NULL,
    
    -- Snapshot data (JSONB)
    name VARCHAR(255) NOT NULL,
    description TEXT,
    filters_json JSONB,
    cards_json JSONB NOT NULL,
    layout_json JSONB,
    
    -- Metadata
    created_by TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    change_summary TEXT DEFAULT '',
    is_auto_save BOOLEAN DEFAULT FALSE,
    metadata JSONB DEFAULT '{}',
    
    -- Unique constraint: each dashboard can only have one version with a specific version number
    UNIQUE(dashboard_id, version)
);

-- Indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_dashboard_versions_dashboard_id ON dashboard_versions(dashboard_id);
CREATE INDEX IF NOT EXISTS idx_dashboard_versions_created_by ON dashboard_versions(created_by);
CREATE INDEX IF NOT EXISTS idx_dashboard_versions_is_auto_save ON dashboard_versions(is_auto_save);
CREATE INDEX IF NOT EXISTS idx_dashboard_versions_created_at ON dashboard_versions(created_at);

-- Composite index for filtering by dashboard and date
CREATE INDEX IF NOT EXISTS idx_dashboard_versions_dashboard_created 
    ON dashboard_versions(dashboard_id, created_at DESC);

-- Index for finding auto-save versions per dashboard
CREATE INDEX IF NOT EXISTS idx_dashboard_versions_dashboard_auto_save 
    ON dashboard_versions(dashboard_id, is_auto_save, created_at DESC);

-- Comments
COMMENT ON TABLE dashboard_versions IS 'Dashboard version history snapshots';
COMMENT ON COLUMN dashboard_versions.dashboard_id IS 'Reference to the dashboard';
COMMENT ON COLUMN dashboard_versions.version IS 'Version number (auto-increment per dashboard)';
COMMENT ON COLUMN dashboard_versions.cards_json IS 'JSON array of cards at the time of version creation';
COMMENT ON COLUMN dashboard_versions.filters_json IS 'JSON array of filter configurations';
COMMENT ON COLUMN dashboard_versions.change_summary IS 'Human-readable summary of changes';
COMMENT ON COLUMN dashboard_versions.is_auto_save IS 'Whether this version was auto-saved';
COMMENT ON COLUMN dashboard_versions.metadata IS 'Additional metadata about the version';

-- Function to auto-increment version number
CREATE OR REPLACE FUNCTION increment_dashboard_version()
RETURNS TRIGGER AS $$
BEGIN
    SELECT COALESCE(MAX(version), 0) + 1 INTO NEW.version
    FROM dashboard_versions
    WHERE dashboard_id = NEW.dashboard_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to auto-increment version number on insert
DROP TRIGGER IF EXISTS tr_dashboard_version_increment ON dashboard_versions;
CREATE TRIGGER tr_dashboard_version_increment
    BEFORE INSERT ON dashboard_versions
    FOR EACH ROW
    EXECUTE FUNCTION increment_dashboard_version();

-- Function to cleanup old auto-save versions (keep last 10)
CREATE OR REPLACE FUNCTION cleanup_dashboard_auto_save_versions()
RETURNS TRIGGER AS $$
DECLARE
    delete_count INTEGER;
BEGIN
    IF NEW.is_auto_save THEN
        DELETE FROM dashboard_versions
        WHERE id IN (
            SELECT id FROM dashboard_versions
            WHERE dashboard_id = NEW.dashboard_id AND is_auto_save = TRUE
            ORDER BY created_at DESC
            OFFSET 10
        );
        
        GET DIAGNOSTICS delete_count = ROW_COUNT;
        
        IF delete_count > 0 THEN
            RAISE NOTICE 'Deleted % old auto-save versions for dashboard %', delete_count, NEW.dashboard_id;
        END IF;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to cleanup old auto-save versions after insert
DROP TRIGGER IF EXISTS tr_cleanup_dashboard_auto_save ON dashboard_versions;
CREATE TRIGGER tr_cleanup_dashboard_auto_save
    AFTER INSERT ON dashboard_versions
    FOR EACH ROW
    EXECUTE FUNCTION cleanup_dashboard_auto_save_versions();
