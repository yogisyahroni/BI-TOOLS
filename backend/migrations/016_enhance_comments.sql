-- Migration: Enhance comments table for threading, mentions, and annotations
-- Task: TASK-092 - Comment System Backend Enhancement
-- Task: TASK-094 - Chart Annotations
-- Created: 2026-02-10

-- First, create the annotations table
CREATE TABLE IF NOT EXISTS annotations (
    id TEXT PRIMARY KEY,
    comment_id TEXT NOT NULL REFERENCES comments(id) ON DELETE CASCADE,
    chart_id TEXT NOT NULL,
    x_value DOUBLE PRECISION,
    y_value DOUBLE PRECISION,
    x_category TEXT,
    y_category TEXT,
    position JSONB NOT NULL DEFAULT '{}',
    type VARCHAR(20) NOT NULL DEFAULT 'point',
    color VARCHAR(20) NOT NULL DEFAULT '#F59E0B',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Add new columns to comments table
ALTER TABLE comments 
    ADD COLUMN IF NOT EXISTS parent_id TEXT REFERENCES comments(id) ON DELETE CASCADE,
    ADD COLUMN IF NOT EXISTS is_resolved BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS mentions JSONB DEFAULT '[]';

-- Update existing comments entity_type to use new enum values
-- Note: Existing entity types (pipeline, dataflow, collection) remain valid

-- Create indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_comments_parent_id ON comments(parent_id) WHERE parent_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_comments_is_resolved ON comments(is_resolved);
CREATE INDEX IF NOT EXISTS idx_comments_entity_composite ON comments(entity_type, entity_id, is_resolved);
CREATE INDEX IF NOT EXISTS idx_comments_created_at ON comments(created_at DESC);

-- Index for mentions (using GIN for JSONB)
CREATE INDEX IF NOT EXISTS idx_comments_mentions ON comments USING GIN(mentions);

-- Indexes for annotations
CREATE INDEX IF NOT EXISTS idx_annotations_comment_id ON annotations(comment_id);
CREATE INDEX IF NOT EXISTS idx_annotations_chart_id ON annotations(chart_id);
CREATE INDEX IF NOT EXISTS idx_annotations_chart_composite ON annotations(chart_id, type);

-- Add comments for documentation
COMMENT ON TABLE annotations IS 'Chart annotations linked to comments';
COMMENT ON COLUMN annotations.comment_id IS 'Reference to the associated comment';
COMMENT ON COLUMN annotations.chart_id IS 'The chart this annotation belongs to';
COMMENT ON COLUMN annotations.x_value IS 'X coordinate for numeric axes';
COMMENT ON COLUMN annotations.y_value IS 'Y coordinate for numeric axes';
COMMENT ON COLUMN annotations.x_category IS 'X category for categorical axes';
COMMENT ON COLUMN annotations.y_category IS 'Y category for categorical axes';
COMMENT ON COLUMN annotations.position IS 'Pixel position as JSON {x, y}';
COMMENT ON COLUMN annotations.type IS 'Annotation type: point, range, or text';
COMMENT ON COLUMN annotations.color IS 'Annotation marker color';

COMMENT ON COLUMN comments.parent_id IS 'Reference to parent comment for threading/replies';
COMMENT ON COLUMN comments.is_resolved IS 'Whether this comment thread is resolved';
COMMENT ON COLUMN comments.mentions IS 'JSON array of mentioned user IDs';

-- Function to auto-update updated_at timestamp for annotations
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger for auto-updating annotations updated_at
DROP TRIGGER IF EXISTS update_annotations_updated_at ON annotations;
CREATE TRIGGER update_annotations_updated_at
    BEFORE UPDATE ON annotations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Trigger for auto-updating comments updated_at (if not already exists)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_comments_updated_at') THEN
        CREATE TRIGGER update_comments_updated_at
            BEFORE UPDATE ON comments
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at_column();
    END IF;
END $$;

-- Migration completed
