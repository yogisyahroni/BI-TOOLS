/**
 * Share API Service
 * 
 * Service layer for Advanced Sharing API calls.
 * Provides type-safe methods for managing shares and embed tokens.
 */

import { logApiError, logger } from '../logger'
import { apiGet, apiPost, apiPut, apiDelete } from './config'
import type {
    // Share types
    Share,
    ShareWithDetails,
    CreateShareRequest,
    UpdateShareRequest,
    AcceptShareRequest,
    ValidateShareAccessRequest,
    ShareAccessCheck,
    ValidateShareAccessResponse,
    ShareFilter,
    GetSharesResponse,
    ResourceType,
    // Embed token types
    EmbedToken,
    EmbedTokenWithStats,
    CreateEmbedTokenRequest,
    UpdateEmbedTokenRequest,
    ValidateEmbedTokenResponse,
    EmbedTokenFilter,
    GetEmbedTokensResponse,
} from '@/types/share'

// ============================================================================
// Share API
// ============================================================================

/**
 * Create a new share for a resource
 * 
 * @param data Share creation data
 * @returns Newly created share
 */
export async function createShare(data: CreateShareRequest): Promise<Share> {
    try {
        logger.debug('share_create_start', 'Creating new share', {
            resource_type: data.resource_type,
            resource_id: data.resource_id,
            permission: data.permission,
            has_password: !!data.password,
            has_expiration: !!data.expires_at,
        })

        const response = await apiPost<Share>('/api/shares', data)

        logger.info('share_created', 'Share created successfully', {
            share_id: response.id,
            resource_type: data.resource_type,
            resource_id: data.resource_id,
        })

        return response
    } catch (error) {
        logApiError('share_create_failed', error, {
            resource_type: data.resource_type,
            resource_id: data.resource_id,
        })
        throw error
    }
}

/**
 * Get all shares for a specific resource
 * 
 * @param resourceType Type of resource (dashboard or query)
 * @param resourceId Resource ID
 * @returns List of shares for the resource
 */
export async function getSharesForResource(
    resourceType: ResourceType,
    resourceId: string
): Promise<Share[]> {
    try {
        logger.debug('shares_fetch_start', 'Fetching shares for resource', {
            resource_type: resourceType,
            resource_id: resourceId,
        })

        const response = await apiGet<GetSharesResponse>(`/api/shares/resource/${resourceType}/${resourceId}`)

        logger.info('shares_loaded', 'Shares loaded successfully', {
            resource_type: resourceType,
            resource_id: resourceId,
            count: response.total,
        })

        return response.shares || []
    } catch (error) {
        logApiError('shares_fetch_failed', error, {
            resource_type: resourceType,
            resource_id: resourceId,
        })
        throw error
    }
}

/**
 * Get shares for the current user (both shared by and shared with)
 * 
 * @param filter Optional filter criteria
 * @returns List of user's shares
 */
export async function getMyShares(filter?: ShareFilter): Promise<Share[]> {
    try {
        logger.debug('my_shares_fetch_start', 'Fetching user shares', { filter })

        // Build query string
        const params = new URLSearchParams()
        if (filter?.resource_type) {
            params.append('resource_type', filter.resource_type)
        }
        if (filter?.status) {
            params.append('status', filter.status)
        }
        if (filter?.include_expired) {
            params.append('include_expired', 'true')
        }

        const queryString = params.toString()
        const url = queryString ? `/api/shares/my?${queryString}` : '/api/shares/my'

        const response = await apiGet<GetSharesResponse>(url)

        logger.info('my_shares_loaded', 'User shares loaded successfully', {
            count: response.total,
        })

        return response.shares || []
    } catch (error) {
        logApiError('my_shares_fetch_failed', error, {})
        throw error
    }
}

/**
 * Get a single share by ID
 * 
 * @param shareId Share ID
 * @returns Share details
 */
export async function getShareById(shareId: string): Promise<Share> {
    try {
        logger.debug('share_fetch_start', 'Fetching share by ID', { share_id: shareId })

        const response = await apiGet<Share>(`/api/shares/${shareId}`)

        logger.info('share_loaded', 'Share loaded successfully', {
            share_id: shareId,
            status: response.status,
        })

        return response
    } catch (error) {
        logApiError('share_fetch_failed', error, { share_id: shareId })
        throw error
    }
}

/**
 * Update a share
 * 
 * @param shareId Share ID to update
 * @param data Update data
 * @returns Updated share
 */
export async function updateShare(shareId: string, data: UpdateShareRequest): Promise<Share> {
    try {
        logger.debug('share_update_start', 'Updating share', {
            share_id: shareId,
            updates: Object.keys(data),
        })

        const response = await apiPut<Share>(`/api/shares/${shareId}`, data)

        logger.info('share_updated', 'Share updated successfully', {
            share_id: shareId,
        })

        return response
    } catch (error) {
        logApiError('share_update_failed', error, { share_id: shareId })
        throw error
    }
}

/**
 * Revoke (delete) a share
 * 
 * @param shareId Share ID to revoke
 */
export async function revokeShare(shareId: string): Promise<void> {
    try {
        logger.debug('share_revoke_start', 'Revoking share', { share_id: shareId })

        await apiDelete<{ message: string }>(`/api/shares/${shareId}`)

        logger.info('share_revoked', 'Share revoked successfully', {
            share_id: shareId,
        })
    } catch (error) {
        logApiError('share_revoke_failed', error, { share_id: shareId })
        throw error
    }
}

/**
 * Accept a pending share invitation
 * 
 * @param shareId Share ID to accept
 * @param data Optional password for password-protected shares
 */
export async function acceptShare(shareId: string, data?: AcceptShareRequest): Promise<void> {
    try {
        logger.debug('share_accept_start', 'Accepting share', { share_id: shareId })

        await apiPost<{ message: string }>(`/api/shares/${shareId}/accept`, data || {})

        logger.info('share_accepted', 'Share accepted successfully', {
            share_id: shareId,
        })
    } catch (error) {
        logApiError('share_accept_failed', error, { share_id: shareId })
        throw error
    }
}

/**
 * Check if current user has access to a resource via sharing
 * 
 * @param resourceType Type of resource
 * @param resourceId Resource ID
 * @returns Access check result
 */
export async function checkShareAccess(
    resourceType: ResourceType,
    resourceId: string
): Promise<ShareAccessCheck> {
    try {
        logger.debug('share_access_check_start', 'Checking share access', {
            resource_type: resourceType,
            resource_id: resourceId,
        })

        const response = await apiGet<ShareAccessCheck>(
            `/api/shares/check?resource_type=${resourceType}&resource_id=${resourceId}`
        )

        logger.debug('share_access_checked', 'Share access check completed', {
            resource_type: resourceType,
            resource_id: resourceId,
            has_access: response.has_access,
        })

        return response
    } catch (error) {
        logApiError('share_access_check_failed', error, {
            resource_type: resourceType,
            resource_id: resourceId,
        })
        throw error
    }
}

/**
 * Validate share access with password
 * 
 * @param data Validation request with share ID and optional password
 * @returns Validation result
 */
export async function validateShareAccess(
    data: ValidateShareAccessRequest
): Promise<ValidateShareAccessResponse> {
    try {
        logger.debug('share_validate_start', 'Validating share access', { share_id: data.share_id })

        const response = await apiPost<ValidateShareAccessResponse>('/api/shares/validate', data)

        logger.info('share_validated', 'Share access validated successfully', {
            share_id: data.share_id,
            valid: response.valid,
        })

        return response
    } catch (error) {
        logApiError('share_validate_failed', error, { share_id: data.share_id })
        throw error
    }
}

// ============================================================================
// Embed Token API
// ============================================================================

/**
 * Create a new embed token
 * 
 * @param data Embed token creation data
 * @returns Newly created embed token
 */
export async function createEmbedToken(data: CreateEmbedTokenRequest): Promise<EmbedToken> {
    try {
        logger.debug('embed_token_create_start', 'Creating new embed token', {
            resource_type: data.resource_type,
            resource_id: data.resource_id,
            has_domain_restriction: !!data.allowed_domains?.length,
            has_ip_restriction: !!data.allowed_ips?.length,
        })

        const response = await apiPost<EmbedToken>('/api/embed-tokens', data)

        logger.info('embed_token_created', 'Embed token created successfully', {
            token_id: response.id,
            resource_type: data.resource_type,
            resource_id: data.resource_id,
        })

        return response
    } catch (error) {
        logApiError('embed_token_create_failed', error, {
            resource_type: data.resource_type,
            resource_id: data.resource_id,
        })
        throw error
    }
}

/**
 * Get all embed tokens for the current user
 * 
 * @param filter Optional filter criteria
 * @returns List of embed tokens
 */
export async function getMyEmbedTokens(filter?: EmbedTokenFilter): Promise<EmbedToken[]> {
    try {
        logger.debug('embed_tokens_fetch_start', 'Fetching user embed tokens', { filter })

        // Build query string
        const params = new URLSearchParams()
        if (filter?.resource_type) {
            params.append('resource_type', filter.resource_type)
        }
        if (filter?.resource_id) {
            params.append('resource_id', filter.resource_id)
        }
        if (filter?.include_expired) {
            params.append('include_expired', 'true')
        }
        if (filter?.include_revoked) {
            params.append('include_revoked', 'true')
        }

        const queryString = params.toString()
        const url = queryString ? `/api/embed-tokens?${queryString}` : '/api/embed-tokens'

        const response = await apiGet<GetEmbedTokensResponse>(url)

        logger.info('embed_tokens_loaded', 'Embed tokens loaded successfully', {
            count: response.total,
        })

        return response.tokens || []
    } catch (error) {
        logApiError('embed_tokens_fetch_failed', error, {})
        throw error
    }
}

/**
 * Get embed tokens for a specific resource
 * 
 * @param resourceType Type of resource
 * @param resourceId Resource ID
 * @returns List of embed tokens
 */
export async function getEmbedTokensForResource(
    resourceType: ResourceType,
    resourceId: string
): Promise<EmbedToken[]> {
    try {
        logger.debug('embed_tokens_resource_fetch_start', 'Fetching embed tokens for resource', {
            resource_type: resourceType,
            resource_id: resourceId,
        })

        const response = await apiGet<GetEmbedTokensResponse>(
            `/api/embed-tokens/resource/${resourceType}/${resourceId}`
        )

        logger.info('embed_tokens_resource_loaded', 'Embed tokens loaded successfully', {
            resource_type: resourceType,
            resource_id: resourceId,
            count: response.total,
        })

        return response.tokens || []
    } catch (error) {
        logApiError('embed_tokens_resource_fetch_failed', error, {
            resource_type: resourceType,
            resource_id: resourceId,
        })
        throw error
    }
}

/**
 * Get a single embed token by ID
 * 
 * @param tokenId Token ID
 * @returns Embed token details
 */
export async function getEmbedTokenById(tokenId: string): Promise<EmbedToken> {
    try {
        logger.debug('embed_token_fetch_start', 'Fetching embed token by ID', { token_id: tokenId })

        const response = await apiGet<EmbedToken>(`/api/embed-tokens/${tokenId}`)

        logger.info('embed_token_loaded', 'Embed token loaded successfully', {
            token_id: tokenId,
        })

        return response
    } catch (error) {
        logApiError('embed_token_fetch_failed', error, { token_id: tokenId })
        throw error
    }
}

/**
 * Update an embed token
 * 
 * @param tokenId Token ID to update
 * @param data Update data
 * @returns Updated embed token
 */
export async function updateEmbedToken(tokenId: string, data: UpdateEmbedTokenRequest): Promise<EmbedToken> {
    try {
        logger.debug('embed_token_update_start', 'Updating embed token', {
            token_id: tokenId,
            updates: Object.keys(data),
        })

        const response = await apiPut<EmbedToken>(`/api/embed-tokens/${tokenId}`, data)

        logger.info('embed_token_updated', 'Embed token updated successfully', {
            token_id: tokenId,
        })

        return response
    } catch (error) {
        logApiError('embed_token_update_failed', error, { token_id: tokenId })
        throw error
    }
}

/**
 * Revoke (delete) an embed token
 * 
 * @param tokenId Token ID to revoke
 */
export async function revokeEmbedToken(tokenId: string): Promise<void> {
    try {
        logger.debug('embed_token_revoke_start', 'Revoking embed token', { token_id: tokenId })

        await apiDelete<{ message: string }>(`/api/embed-tokens/${tokenId}`)

        logger.info('embed_token_revoked', 'Embed token revoked successfully', {
            token_id: tokenId,
        })
    } catch (error) {
        logApiError('embed_token_revoke_failed', error, { token_id: tokenId })
        throw error
    }
}

/**
 * Validate an embed token
 * 
 * @param token Token string
 * @param domain Current domain (for domain restriction check)
 * @returns Validation result with resource info if valid
 */
export async function validateEmbedToken(
    token: string,
    domain?: string
): Promise<ValidateEmbedTokenResponse> {
    try {
        logger.debug('embed_token_validate_start', 'Validating embed token', {
            token_preview: token.substring(0, 8) + '...',
            domain,
        })

        const queryParams = domain ? `?domain=${encodeURIComponent(domain)}` : ''
        const response = await apiGet<ValidateEmbedTokenResponse>(
            `/api/embed-tokens/${token}/validate${queryParams}`
        )

        logger.info('embed_token_validated', 'Embed token validated successfully', {
            valid: response.valid,
            resource_type: response.resource_type,
        })

        return response
    } catch (error) {
        logApiError('embed_token_validate_failed', error, {
            token_preview: token.substring(0, 8) + '...',
        })
        throw error
    }
}

/**
 * Get embed token statistics
 * 
 * @param tokenId Token ID
 * @returns Token with usage statistics
 */
export async function getEmbedTokenStats(tokenId: string): Promise<EmbedTokenWithStats> {
    try {
        logger.debug('embed_token_stats_fetch_start', 'Fetching embed token stats', { token_id: tokenId })

        const response = await apiGet<EmbedTokenWithStats>(`/api/embed-tokens/${tokenId}/stats`)

        logger.info('embed_token_stats_loaded', 'Embed token stats loaded successfully', {
            token_id: tokenId,
            view_count: response.view_count,
        })

        return response
    } catch (error) {
        logApiError('embed_token_stats_fetch_failed', error, { token_id: tokenId })
        throw error
    }
}
