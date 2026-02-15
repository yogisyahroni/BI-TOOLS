-- Migration: Create shares table for Advanced Sharing feature
-- Task: TASK-088 - Granular Sharing Permissions
-- Created: 2026-02-10

-- Shares table: Store resource sharing information
CREATE TABLE IF NOT EXISTS shares (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource_type VARCHAR(50) NOT NULL, -- 'dashboard', 'query'
    resource_id TEXT NOT NULL,
    shared_by TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    shared_with TEXT REFERENCES users(id) ON DELETE CASCADE, -- NULL for email invites
    shared_email TEXT, -- For external email invites
    permission VARCHAR(20) NOT NULL DEFAULT 'view', -- 'view', 'edit', 'admin'
    password_hash TEXT, -- NULL if no password protection
    expires_at TIMESTAMP WITH TIME ZONE, -- NULL if no expiration
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- 'active', 'revoked', 'expired', 'pending'
    accepted_at TIMESTAMP WITH TIME ZONE, -- When invite was accepted
    message TEXT, -- Optional message to recipient
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Embed tokens table: Store embed tokens with domain/IP restrictions
CREATE TABLE IF NOT EXISTS embed_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource_type VARCHAR(50) NOT NULL, -- 'dashboard', 'query'
    resource_id TEXT NOT NULL,
    token TEXT NOT NULL UNIQUE,
    created_by TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    allowed_domains JSONB DEFAULT '[]', -- Array of allowed domains
    allowed_ips JSONB DEFAULT '[]', -- Array of allowed IPs
    expires_at TIMESTAMP WITH TIME ZONE, -- NULL if no expiration
    view_count BIGINT DEFAULT 0,
    last_viewed_at TIMESTAMP WITH TIME ZONE,
    is_revoked BOOLEAN DEFAULT FALSE,
    revoked_at TIMESTAMP WITH TIME ZONE,
    revoked_by TEXT REFERENCES users(id) ON DELETE SET NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for shares table
CREATE INDEX IF NOT EXISTS idx_shares_resource ON shares(resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_shares_shared_by ON shares(shared_by);
CREATE INDEX IF NOT EXISTS idx_shares_shared_with ON shares(shared_with);
CREATE INDEX IF NOT EXISTS idx_shares_shared_email ON shares(shared_email);
CREATE INDEX IF NOT EXISTS idx_shares_status ON shares(status);
CREATE INDEX IF NOT EXISTS idx_shares_expires_at ON shares(expires_at) WHERE expires_at IS NOT NULL;

-- Indexes for embed_tokens table
CREATE INDEX IF NOT EXISTS idx_embed_tokens_resource ON embed_tokens(resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_embed_tokens_created_by ON embed_tokens(created_by);
CREATE INDEX IF NOT EXISTS idx_embed_tokens_token ON embed_tokens(token);
CREATE INDEX IF NOT EXISTS idx_embed_tokens_expires_at ON embed_tokens(expires_at) WHERE expires_at IS NOT NULL;

-- Comments
COMMENT ON TABLE shares IS 'Resource shares with granular permissions (view/edit/admin)';
COMMENT ON COLUMN shares.resource_type IS 'Type of resource being shared: dashboard or query';
COMMENT ON COLUMN shares.permission IS 'Permission level: view, edit, or admin';
COMMENT ON COLUMN shares.status IS 'Share status: active, revoked, expired, or pending';

COMMENT ON TABLE embed_tokens IS 'Embed tokens for external embedding with domain/IP restrictions';
COMMENT ON COLUMN embed_tokens.allowed_domains IS 'JSON array of allowed domains (e.g., ["example.com", "*.example.com"])';
COMMENT ON COLUMN embed_tokens.allowed_ips IS 'JSON array of allowed IP addresses';

-- Function to auto-update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers for auto-updating updated_at
DROP TRIGGER IF EXISTS update_shares_updated_at ON shares;
CREATE TRIGGER update_shares_updated_at
    BEFORE UPDATE ON shares
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_embed_tokens_updated_at ON embed_tokens;
CREATE TRIGGER update_embed_tokens_updated_at
    BEFORE UPDATE ON embed_tokens
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
