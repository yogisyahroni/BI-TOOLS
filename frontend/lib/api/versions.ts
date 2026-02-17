/**
 * Version Control API Service
 * 
 * Service layer for Dashboard and Query Version Control API calls.
 * Provides type-safe methods for managing versions.
 */

import { logApiError, logger } from '../logger'
import { apiGet, apiPost, apiDelete } from './config'
import type {
  DashboardVersion,
  QueryVersion,
  CreateDashboardVersionRequest,
  CreateQueryVersionRequest,
  VersionFilterOptions,
  GetDashboardVersionsResponse,
  GetQueryVersionsResponse,
  DashboardVersionDiff,
  QueryVersionDiff,
  RestoreVersionResponse,
  _CompareVersionsRequest,
} from '@/types/versions'

// ============================================================================
// Dashboard Versions API
// ============================================================================

/**
 * Create a new dashboard version
 * 
 * @param dashboardId Dashboard ID
 * @param data Version creation data
 * @returns Newly created version
 */
export async function createDashboardVersion(
  dashboardId: string,
  data: CreateDashboardVersionRequest
): Promise<DashboardVersion> {
  try {
    logger.debug('dashboard_version_create_start', 'Creating new dashboard version', {
      dashboard_id: dashboardId,
      is_auto_save: data.isAutoSave,
      has_summary: !!data.changeSummary,
    })

    const response = await apiPost<DashboardVersion>(`/api/dashboards/${dashboardId}/versions`, data)

    logger.info('dashboard_version_created', 'Dashboard version created successfully', {
      version_id: response.id,
      dashboard_id: dashboardId,
      version_number: response.version,
    })

    return response
  } catch (error) {
    logApiError('dashboard_version_create_failed', error, {
      dashboard_id: dashboardId,
    })
    throw error
  }
}

/**
 * Get all versions for a dashboard
 * 
 * @param dashboardId Dashboard ID
 * @param filter Optional filter options
 * @returns List of versions
 */
export async function getDashboardVersions(
  dashboardId: string,
  filter?: VersionFilterOptions
): Promise<GetDashboardVersionsResponse> {
  try {
    logger.debug('dashboard_versions_fetch_start', 'Fetching dashboard versions', {
      dashboard_id: dashboardId,
      filter,
    })

    // Build query string
    const params = new URLSearchParams()
    if (filter?.limit) {
      params.append('limit', String(filter.limit))
    }
    if (filter?.offset !== undefined) {
      params.append('offset', String(filter.offset))
    }
    if (filter?.isAutoSave !== undefined) {
      params.append('is_auto_save', String(filter.isAutoSave))
    }
    if (filter?.orderBy) {
      params.append('order_by', filter.orderBy)
    }

    const queryString = params.toString()
    const url = queryString 
      ? `/api/dashboards/${dashboardId}/versions?${queryString}` 
      : `/api/dashboards/${dashboardId}/versions`

    const response = await apiGet<GetDashboardVersionsResponse>(url)

    logger.info('dashboard_versions_loaded', 'Dashboard versions loaded successfully', {
      dashboard_id: dashboardId,
      total: response.total,
    })

    return response
  } catch (error) {
    logApiError('dashboard_versions_fetch_failed', error, {
      dashboard_id: dashboardId,
    })
    throw error
  }
}

/**
 * Get a single dashboard version by ID
 * 
 * @param versionId Version ID
 * @returns Version details
 */
export async function getDashboardVersion(versionId: string): Promise<DashboardVersion> {
  try {
    logger.debug('dashboard_version_fetch_start', 'Fetching dashboard version', {
      version_id: versionId,
    })

    const response = await apiGet<DashboardVersion>(`/api/versions/${versionId}`)

    logger.info('dashboard_version_loaded', 'Dashboard version loaded successfully', {
      version_id: versionId,
      version_number: response.version,
    })

    return response
  } catch (error) {
    logApiError('dashboard_version_fetch_failed', error, {
      version_id: versionId,
    })
    throw error
  }
}

/**
 * Restore a dashboard to a specific version
 * 
 * @param versionId Version ID to restore
 * @returns Restore response
 */
export async function restoreDashboardVersion(versionId: string): Promise<RestoreVersionResponse> {
  try {
    logger.debug('dashboard_version_restore_start', 'Restoring dashboard version', {
      version_id: versionId,
    })

    const response = await apiPost<RestoreVersionResponse>(`/api/versions/${versionId}/restore`, {})

    logger.info('dashboard_version_restored', 'Dashboard version restored successfully', {
      version_id: versionId,
      restored_to_version: response.restoredToVersion,
    })

    return response
  } catch (error) {
    logApiError('dashboard_version_restore_failed', error, {
      version_id: versionId,
    })
    throw error
  }
}

/**
 * Delete a dashboard version
 * 
 * @param versionId Version ID to delete
 */
export async function deleteDashboardVersion(versionId: string): Promise<void> {
  try {
    logger.debug('dashboard_version_delete_start', 'Deleting dashboard version', {
      version_id: versionId,
    })

    await apiDelete<{ message: string }>(`/api/versions/${versionId}`)

    logger.info('dashboard_version_deleted', 'Dashboard version deleted successfully', {
      version_id: versionId,
    })
  } catch (error) {
    logApiError('dashboard_version_delete_failed', error, {
      version_id: versionId,
    })
    throw error
  }
}

/**
 * Compare two dashboard versions
 * 
 * @param versionId1 First version ID
 * @param versionId2 Second version ID
 * @returns Diff between versions
 */
export async function compareDashboardVersions(
  versionId1: string,
  versionId2: string
): Promise<DashboardVersionDiff> {
  try {
    logger.debug('dashboard_versions_compare_start', 'Comparing dashboard versions', {
      version_id_1: versionId1,
      version_id_2: versionId2,
    })

    const params = new URLSearchParams({
      version_id_1: versionId1,
      version_id_2: versionId2,
    })

    const response = await apiGet<DashboardVersionDiff>(`/api/versions/compare?${params.toString()}`)

    logger.info('dashboard_versions_compared', 'Dashboard versions compared successfully', {
      version_id_1: versionId1,
      version_id_2: versionId2,
    })

    return response
  } catch (error) {
    logApiError('dashboard_versions_compare_failed', error, {
      version_id_1: versionId1,
      version_id_2: versionId2,
    })
    throw error
  }
}

/**
 * Auto-save current dashboard state
 * 
 * @param dashboardId Dashboard ID
 * @returns Auto-saved version
 */
export async function autoSaveDashboard(dashboardId: string): Promise<DashboardVersion> {
  try {
    logger.debug('dashboard_auto_save_start', 'Auto-saving dashboard', {
      dashboard_id: dashboardId,
    })

    const response = await apiPost<DashboardVersion>(`/api/dashboards/${dashboardId}/versions/auto-save`, {})

    logger.debug('dashboard_auto_saved', 'Dashboard auto-saved successfully', {
      dashboard_id: dashboardId,
      version_id: response.id,
    })

    return response
  } catch (error) {
    logApiError('dashboard_auto_save_failed', error, {
      dashboard_id: dashboardId,
    })
    throw error
  }
}

// ============================================================================
// Query Versions API
// ============================================================================

/**
 * Create a new query version
 * 
 * @param queryId Query ID
 * @param data Version creation data
 * @returns Newly created version
 */
export async function createQueryVersion(
  queryId: string,
  data: CreateQueryVersionRequest
): Promise<QueryVersion> {
  try {
    logger.debug('query_version_create_start', 'Creating new query version', {
      query_id: queryId,
      is_auto_save: data.isAutoSave,
    })

    const response = await apiPost<QueryVersion>(`/api/queries/${queryId}/versions`, data)

    logger.info('query_version_created', 'Query version created successfully', {
      version_id: response.id,
      query_id: queryId,
      version_number: response.version,
    })

    return response
  } catch (error) {
    logApiError('query_version_create_failed', error, {
      query_id: queryId,
    })
    throw error
  }
}

/**
 * Get all versions for a query
 * 
 * @param queryId Query ID
 * @param filter Optional filter options
 * @returns List of versions
 */
export async function getQueryVersions(
  queryId: string,
  filter?: VersionFilterOptions
): Promise<GetQueryVersionsResponse> {
  try {
    logger.debug('query_versions_fetch_start', 'Fetching query versions', {
      query_id: queryId,
    })

    // Build query string
    const params = new URLSearchParams()
    if (filter?.limit) {
      params.append('limit', String(filter.limit))
    }
    if (filter?.offset !== undefined) {
      params.append('offset', String(filter.offset))
    }
    if (filter?.isAutoSave !== undefined) {
      params.append('is_auto_save', String(filter.isAutoSave))
    }

    const queryString = params.toString()
    const url = queryString 
      ? `/api/queries/${queryId}/versions?${queryString}` 
      : `/api/queries/${queryId}/versions`

    const response = await apiGet<GetQueryVersionsResponse>(url)

    logger.info('query_versions_loaded', 'Query versions loaded successfully', {
      query_id: queryId,
      total: response.total,
    })

    return response
  } catch (error) {
    logApiError('query_versions_fetch_failed', error, {
      query_id: queryId,
    })
    throw error
  }
}

/**
 * Get a single query version by ID
 * 
 * @param versionId Version ID
 * @returns Version details
 */
export async function getQueryVersion(versionId: string): Promise<QueryVersion> {
  try {
    logger.debug('query_version_fetch_start', 'Fetching query version', {
      version_id: versionId,
    })

    const response = await apiGet<QueryVersion>(`/api/query-versions/${versionId}`)

    logger.info('query_version_loaded', 'Query version loaded successfully', {
      version_id: versionId,
      version_number: response.version,
    })

    return response
  } catch (error) {
    logApiError('query_version_fetch_failed', error, {
      version_id: versionId,
    })
    throw error
  }
}

/**
 * Restore a query to a specific version
 * 
 * @param versionId Version ID to restore
 * @returns Restore response
 */
export async function restoreQueryVersion(versionId: string): Promise<RestoreVersionResponse> {
  try {
    logger.debug('query_version_restore_start', 'Restoring query version', {
      version_id: versionId,
    })

    const response = await apiPost<RestoreVersionResponse>(`/api/query-versions/${versionId}/restore`, {})

    logger.info('query_version_restored', 'Query version restored successfully', {
      version_id: versionId,
      restored_to_version: response.restoredToVersion,
    })

    return response
  } catch (error) {
    logApiError('query_version_restore_failed', error, {
      version_id: versionId,
    })
    throw error
  }
}

/**
 * Delete a query version
 * 
 * @param versionId Version ID to delete
 */
export async function deleteQueryVersion(versionId: string): Promise<void> {
  try {
    logger.debug('query_version_delete_start', 'Deleting query version', {
      version_id: versionId,
    })

    await apiDelete<{ message: string }>(`/api/query-versions/${versionId}`)

    logger.info('query_version_deleted', 'Query version deleted successfully', {
      version_id: versionId,
    })
  } catch (error) {
    logApiError('query_version_delete_failed', error, {
      version_id: versionId,
    })
    throw error
  }
}

/**
 * Compare two query versions
 * 
 * @param versionId1 First version ID
 * @param versionId2 Second version ID
 * @returns Diff between versions
 */
export async function compareQueryVersions(
  versionId1: string,
  versionId2: string
): Promise<QueryVersionDiff> {
  try {
    logger.debug('query_versions_compare_start', 'Comparing query versions', {
      version_id_1: versionId1,
      version_id_2: versionId2,
    })

    const params = new URLSearchParams({
      version_id_1: versionId1,
      version_id_2: versionId2,
    })

    const response = await apiGet<QueryVersionDiff>(`/api/query-versions/compare?${params.toString()}`)

    logger.info('query_versions_compared', 'Query versions compared successfully', {
      version_id_1: versionId1,
      version_id_2: versionId2,
    })

    return response
  } catch (error) {
    logApiError('query_versions_compare_failed', error, {
      version_id_1: versionId1,
      version_id_2: versionId2,
    })
    throw error
  }
}

// ============================================================================
// React Query compatible exports
// ============================================================================

export const versionApi = {
  // Dashboard versions
  createDashboardVersion,
  getDashboardVersions,
  getDashboardVersion,
  restoreDashboardVersion,
  deleteDashboardVersion,
  compareDashboardVersions,
  autoSaveDashboard,
  
  // Query versions
  createQueryVersion,
  getQueryVersions,
  getQueryVersion,
  restoreQueryVersion,
  deleteQueryVersion,
  compareQueryVersions,
}
