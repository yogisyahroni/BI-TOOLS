-- Migration 021: Create Organizations Tables
-- Purpose: Add multi-tenant organization support with quotas
-- Create organizations table
CREATE TABLE IF NOT EXISTS organizations (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    description TEXT,
    logo TEXT,
    status TEXT NOT NULL DEFAULT 'active',
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_organizations_status ON organizations(status);
CREATE INDEX IF NOT EXISTS idx_organizations_slug ON organizations(slug);
CREATE INDEX IF NOT EXISTS idx_organizations_created_at ON organizations(created_at DESC);
-- Create organization_members table
CREATE TABLE IF NOT EXISTS organization_members (
    id TEXT PRIMARY KEY,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT 'member',
    joined_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(organization_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_org_members_org_id ON organization_members(organization_id);
CREATE INDEX IF NOT EXISTS idx_org_members_user_id ON organization_members(user_id);
CREATE INDEX IF NOT EXISTS idx_org_members_role ON organization_members(role);
-- Create organization_quotas table
CREATE TABLE IF NOT EXISTS organization_quotas (
    id TEXT PRIMARY KEY,
    organization_id TEXT NOT NULL UNIQUE REFERENCES organizations(id) ON DELETE CASCADE,
    -- Quota limits
    max_users INTEGER NOT NULL DEFAULT 10,
    max_projects INTEGER NOT NULL DEFAULT 5,
    max_queries INTEGER NOT NULL DEFAULT 100,
    max_connections INTEGER NOT NULL DEFAULT 5,
    max_storage BIGINT NOT NULL DEFAULT 1073741824,
    -- 1GB
    -- Current usage
    current_users INTEGER NOT NULL DEFAULT 0,
    current_projects INTEGER NOT NULL DEFAULT 0,
    current_queries INTEGER NOT NULL DEFAULT 0,
    current_connections INTEGER NOT NULL DEFAULT 0,
    current_storage BIGINT NOT NULL DEFAULT 0,
    -- API limits
    api_requests_per_day INTEGER NOT NULL DEFAULT 10000,
    current_api_requests INTEGER NOT NULL DEFAULT 0,
    api_limit_reset_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_org_quotas_org_id ON organization_quotas(organization_id);
-- Add organization_id to existing tables for multi-tenancy support
-- Projects
ALTER TABLE IF EXISTS projects
ADD COLUMN IF NOT EXISTS organization_id TEXT REFERENCES organizations(id) ON DELETE
SET NULL;
CREATE INDEX IF NOT EXISTS idx_projects_org_id ON projects(organization_id);
-- Saved Queries (if table exists)
ALTER TABLE IF EXISTS saved_queries
ADD COLUMN IF NOT EXISTS organization_id TEXT REFERENCES organizations(id) ON DELETE
SET NULL;
CREATE INDEX IF NOT EXISTS idx_saved_queries_org_id ON saved_queries(organization_id);
-- Connections
ALTER TABLE IF NOT EXISTS connections
ADD COLUMN IF NOT EXISTS organization_id TEXT REFERENCES organizations(id) ON DELETE
SET NULL;
CREATE INDEX IF NOT EXISTS idx_connections_org_id ON connections(organization_id);
-- Dashboards
ALTER TABLE IF EXISTS dashboards
ADD COLUMN IF NOT EXISTS organization_id TEXT REFERENCES organizations(id) ON DELETE
SET NULL;
CREATE INDEX IF NOT EXISTS idx_dashboards_org_id ON dashboards(organization_id);
-- Create function to update organization quotas
CREATE OR REPLACE FUNCTION update_organization_usage() RETURNS TRIGGER AS $$ BEGIN -- This is a placeholder function that will be called by triggers
    -- Actual implementation will be done by the backend service
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- Create triggers (optional, can be handled by backend service)
-- These triggers will auto-update quotas when related entities change
COMMENT ON TABLE organizations IS 'Multi-tenant organizations for workspace isolation';
COMMENT ON TABLE organization_members IS 'User membership in organizations with roles';
COMMENT ON TABLE organization_quotas IS 'Resource quotas and usage tracking per organization';