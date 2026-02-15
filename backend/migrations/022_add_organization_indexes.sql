-- Migration: Add indexes for organization tables
-- Created: 2026-02-10
-- TASK-P06: Database performance optimization
-- Add indexes on organization_members table
CREATE INDEX IF NOT EXISTS idx_org_members_user ON organization_members(user_id);
CREATE INDEX IF NOT EXISTS idx_org_members_workspace ON organization_members(workspace_id);
CREATE INDEX IF NOT EXISTS idx_org_members_workspace_role ON organization_members(workspace_id, role);
-- Add index on organization_quotas table
CREATE INDEX IF NOT EXISTS idx_org_quotas_workspace ON organization_quotas(workspace_id);
-- Add indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_org_members_invited_at ON organization_members(invited_at DESC);
CREATE INDEX IF NOT EXISTS idx_workspaces_created_at ON workspaces(created_at DESC);