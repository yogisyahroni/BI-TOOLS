/**
 * Share and Embed Token Type Definitions
 * 
 * TypeScript interfaces for Advanced Sharing feature (TASK-088 to TASK-091)
 * Matches backend models in backend/models/share.go and backend/models/embed_token.go
 */

// ============================================================================
// Core Entity Types
// ============================================================================

/**
 * Resource types that can be shared
 */
export type ResourceType = 'dashboard' | 'query'

/**
 * Permission levels for shares
 */
export type SharePermission = 'view' | 'edit' | 'admin'

/**
 * Share statuses
 */
export type ShareStatus = 'active' | 'revoked' | 'expired' | 'pending'

/**
 * User information for shares
 */
export interface ShareUser {
    id: string
    username: string
    email: string
    name?: string
}

/**
 * Share entity representing a resource share
 */
export interface Share {
    id: string
    resource_type: ResourceType
    resource_id: string
    shared_by: string
    shared_with?: string
    shared_email?: string
    permission: SharePermission
    expires_at?: string
    status: ShareStatus
    accepted_at?: string
    message?: string
    created_at: string
    updated_at: string

    // Relationships (populated when fetching)
    shared_by_user?: ShareUser
    shared_with_user?: ShareUser
}

/**
 * Share with resource details
 */
export interface ShareWithDetails extends Share {
    resource_name?: string
}

// ============================================================================
// Share API Request Types
// ============================================================================

/**
 * Request to create a new share
 */
export interface CreateShareRequest {
    resource_type: ResourceType
    resource_id: string
    shared_with?: string
    shared_email?: string
    permission: SharePermission
    password?: string
    expires_at?: string
    message?: string
}

/**
 * Request to update a share
 */
export interface UpdateShareRequest {
    permission?: SharePermission
    password?: string
    expires_at?: string
    message?: string
}

/**
 * Request to accept a pending share
 */
export interface AcceptShareRequest {
    password?: string
}

/**
 * Request to validate share access with password
 */
export interface ValidateShareAccessRequest {
    share_id: string
    password?: string
}

// ============================================================================
// Share API Response Types
// ============================================================================

/**
 * Response for share access check
 */
export interface ShareAccessCheck {
    has_access: boolean
    permission?: SharePermission
    share_id?: string
    requires_password: boolean
}

/**
 * Response for validating share access
 */
export interface ValidateShareAccessResponse {
    valid: boolean
    share_id: string
    resource_type: ResourceType
    resource_id: string
    permission: SharePermission
}

/**
 * Filter for querying shares
 */
export interface ShareFilter {
    resource_type?: ResourceType
    status?: ShareStatus
    include_expired?: boolean
}

/**
 * Response for list of shares
 */
export interface GetSharesResponse {
    shares: Share[]
    total: number
}

// ============================================================================
// Embed Token Types
// ============================================================================

/**
 * Embed token entity for external embedding
 */
export interface EmbedToken {
    id: string
    resource_type: ResourceType
    resource_id: string
    token: string
    created_by: string
    allowed_domains: string[]
    allowed_ips: string[]
    expires_at?: string
    view_count: number
    last_viewed_at?: string
    is_revoked: boolean
    revoked_at?: string
    revoked_by?: string
    description?: string
    created_at: string
    updated_at: string

    // Relationships
    creator?: ShareUser
}

/**
 * Embed token with resource statistics
 */
export interface EmbedTokenWithStats extends EmbedToken {
    resource_name?: string
}

// ============================================================================
// Embed Token API Request Types
// ============================================================================

/**
 * Request to create a new embed token
 */
export interface CreateEmbedTokenRequest {
    resource_type: ResourceType
    resource_id: string
    allowed_domains?: string[]
    allowed_ips?: string[]
    expires_at?: string
    description?: string
}

/**
 * Request to update an embed token
 */
export interface UpdateEmbedTokenRequest {
    allowed_domains?: string[]
    allowed_ips?: string[]
    expires_at?: string
    description?: string
}

// ============================================================================
// Embed Token API Response Types
// ============================================================================

/**
 * Response for embed token validation
 */
export interface EmbedTokenValidationResult {
    is_valid: boolean
    token_id?: string
    resource_type?: ResourceType
    resource_id?: string
    error?: string
}

/**
 * Response for validating embed token
 */
export interface ValidateEmbedTokenResponse {
    valid: boolean
    token_id: string
    resource_type: ResourceType
    resource_id: string
    view_count: number
}

/**
 * Filter for querying embed tokens
 */
export interface EmbedTokenFilter {
    resource_type?: ResourceType
    resource_id?: string
    include_expired?: boolean
    include_revoked?: boolean
}

/**
 * Response for list of embed tokens
 */
export interface GetEmbedTokensResponse {
    tokens: EmbedToken[]
    total: number
}

// ============================================================================
// UI State Types
// ============================================================================

/**
 * Share dialog state
 */
export interface ShareDialogState {
    isOpen: boolean
    resourceType: ResourceType
    resourceId: string
    resourceName: string
    shares: Share[]
    isLoading: boolean
    error: string | null
}

/**
 * Embed dialog state
 */
export interface EmbedDialogState {
    isOpen: boolean
    resourceType: ResourceType
    resourceId: string
    resourceName: string
    tokens: EmbedToken[]
    isLoading: boolean
    error: string | null
}

/**
 * New share form state
 */
export interface NewShareFormState {
    recipientType: 'user' | 'email'
    recipient: string
    permission: SharePermission
    requirePassword: boolean
    password: string
    setExpiration: boolean
    expirationDate: string
    message: string
}

/**
 * New embed token form state
 */
export interface NewEmbedTokenFormState {
    allowedDomains: string
    allowedIPs: string
    setExpiration: boolean
    expirationDate: string
    description: string
}

/**
 * Permission option for UI
 */
export interface PermissionOption {
    value: SharePermission
    label: string
    description: string
    icon: string
}
