-- Migration: Create query_versions table for Query Versioning feature
-- Task: TASK-097 - Query Versioning Backend
-- Created: 2026-02-10

-- Query versions table: Store version history for queries
CREATE TABLE IF NOT EXISTS query_versions (
    id TEXT PRIMARY KEY,
    query_id TEXT NOT NULL REFERENCES saved_queries(id) ON DELETE CASCADE,
    version INTEGER NOT NULL,
    
    -- Snapshot data
    name TEXT NOT NULL,
    description TEXT,
    sql TEXT NOT NULL,
    ai_prompt TEXT,
    visualization_config JSONB,
    tags JSONB DEFAULT '[]',
    
    -- Metadata
    created_by TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    change_summary TEXT DEFAULT '',
    is_auto_save BOOLEAN DEFAULT FALSE,
    metadata JSONB DEFAULT '{}',
    
    -- Unique constraint: each query can only have one version with a specific version number
    UNIQUE(query_id, version)
);

-- Indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_query_versions_query_id ON query_versions(query_id);
CREATE INDEX IF NOT EXISTS idx_query_versions_created_by ON query_versions(created_by);
CREATE INDEX IF NOT EXISTS idx_query_versions_is_auto_save ON query_versions(is_auto_save);
CREATE INDEX IF NOT EXISTS idx_query_versions_created_at ON query_versions(created_at);

-- Composite index for filtering by query and date
CREATE INDEX IF NOT EXISTS idx_query_versions_query_created 
    ON query_versions(query_id, created_at DESC);

-- Index for finding auto-save versions per query
CREATE INDEX IF NOT EXISTS idx_query_versions_query_auto_save 
    ON query_versions(query_id, is_auto_save, created_at DESC);

-- Comments
COMMENT ON TABLE query_versions IS 'Query version history snapshots';
COMMENT ON COLUMN query_versions.query_id IS 'Reference to the saved query';
COMMENT ON COLUMN query_versions.version IS 'Version number (auto-increment per query)';
COMMENT ON COLUMN query_versions.sql IS 'SQL query snapshot at the time of version creation';
COMMENT ON COLUMN query_versions.visualization_config IS 'JSON visualization configuration';
COMMENT ON COLUMN query_versions.tags IS 'JSON array of tags';
COMMENT ON COLUMN query_versions.change_summary IS 'Human-readable summary of changes';
COMMENT ON COLUMN query_versions.is_auto_save IS 'Whether this version was auto-saved';
COMMENT ON COLUMN query_versions.metadata IS 'Additional metadata about the version including SQL diff summary';

-- Function to auto-increment version number
CREATE OR REPLACE FUNCTION increment_query_version()
RETURNS TRIGGER AS $$
BEGIN
    SELECT COALESCE(MAX(version), 0) + 1 INTO NEW.version
    FROM query_versions
    WHERE query_id = NEW.query_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to auto-increment version number on insert
DROP TRIGGER IF EXISTS tr_query_version_increment ON query_versions;
CREATE TRIGGER tr_query_version_increment
    BEFORE INSERT ON query_versions
    FOR EACH ROW
    EXECUTE FUNCTION increment_query_version();

-- Function to cleanup old auto-save versions (keep last 10)
CREATE OR REPLACE FUNCTION cleanup_query_auto_save_versions()
RETURNS TRIGGER AS $$
DECLARE
    delete_count INTEGER;
BEGIN
    IF NEW.is_auto_save THEN
        DELETE FROM query_versions
        WHERE id IN (
            SELECT id FROM query_versions
            WHERE query_id = NEW.query_id AND is_auto_save = TRUE
            ORDER BY created_at DESC
            OFFSET 10
        );
        
        GET DIAGNOSTICS delete_count = ROW_COUNT;
        
        IF delete_count > 0 THEN
            RAISE NOTICE 'Deleted % old auto-save versions for query %', delete_count, NEW.query_id;
        END IF;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to cleanup old auto-save versions after insert
DROP TRIGGER IF EXISTS tr_cleanup_query_auto_save ON query_versions;
CREATE TRIGGER tr_cleanup_query_auto_save
    AFTER INSERT ON query_versions
    FOR EACH ROW
    EXECUTE FUNCTION cleanup_query_auto_save_versions();
