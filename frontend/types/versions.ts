/**
 * Version Control Type Definitions
 * 
 * TypeScript interfaces for Version Control System feature (TASK-095 to TASK-097)
 * Matches backend models in backend/models/dashboard.go
 */

// ============================================================================
// Core Entity Types
// ============================================================================

/**
 * Resource types that support versioning
 */
export type VersionResourceType = 'dashboard' | 'query'

/**
 * Dashboard version entity
 */
export interface DashboardVersion {
  id: string
  dashboardId: string
  version: number
  
  // Snapshot data
  name: string
  description?: string
  filtersJson?: string
  cardsJson: string
  layoutJson?: string
  
  // Metadata
  createdBy: string
  createdAt: string
  changeSummary: string
  isAutoSave: boolean
  metadata?: Record<string, unknown>
  
  // Relationships
  dashboard?: {
    id: string
    name: string
  }
  createdByUser?: VersionUser
}

/**
 * Query version entity
 */
export interface QueryVersion {
  id: string
  queryId: string
  version: number
  
  // Snapshot data
  name: string
  description?: string
  sql: string
  aiPrompt?: string
  visualizationConfig?: Record<string, unknown>
  tags: string[]
  
  // Metadata
  createdBy: string
  createdAt: string
  changeSummary: string
  isAutoSave: boolean
  
  // Relationships
  query?: {
    id: string
    name: string
  }
  createdByUser?: VersionUser
}

/**
 * User information in versions
 */
export interface VersionUser {
  id: string
  name: string
  email: string
  username?: string
  avatar?: string
  image?: string
}

// ============================================================================
// Card Types
// ============================================================================

/**
 * Dashboard card within a version
 */
export interface VersionCard {
  id: string
  queryId?: string
  title?: string
  position: CardPosition
  visualizationConfig?: Record<string, unknown>
}

/**
 * Card position in grid layout
 */
export interface CardPosition {
  x: number
  y: number
  w: number
  h: number
  i?: string
  moved?: boolean
  static?: boolean
}

/**
 * Filter configuration within a version
 */
export interface VersionFilter {
  id: string
  field: string
  operator: string
  value: unknown
  label?: string
}

// ============================================================================
// Version Metadata Types
// ============================================================================

/**
 * Dashboard version metadata
 */
export interface DashboardVersionMetadata {
  cardCount: number
  filterCount: number
  cardsAdded?: string[]
  cardsRemoved?: string[]
  cardsModified?: string[]
  layoutChanged?: boolean
  filtersChanged?: boolean
}

/**
 * Query version metadata
 */
export interface QueryVersionMetadata {
  sqlChanged: boolean
  metadataChanged: boolean
  configChanged: boolean
  tagsChanged: boolean
}

// ============================================================================
// API Request Types
// ============================================================================

/**
 * Request to create a new dashboard version
 */
export interface CreateDashboardVersionRequest {
  changeSummary?: string
  isAutoSave: boolean
  metadata?: Record<string, unknown>
}

/**
 * Request to create a new query version
 */
export interface CreateQueryVersionRequest {
  changeSummary?: string
  isAutoSave: boolean
  metadata?: Record<string, unknown>
}

/**
 * Filter options for listing versions
 */
export interface VersionFilterOptions {
  isAutoSave?: boolean
  createdBy?: string
  limit?: number
  offset?: number
  orderBy?: 'date_desc' | 'date_asc' | 'version_desc' | 'version_asc'
}

/**
 * Request to compare two versions
 */
export interface CompareVersionsRequest {
  versionId1: string
  versionId2: string
}

// ============================================================================
// API Response Types
// ============================================================================

/**
 * Response for listing dashboard versions
 */
export interface GetDashboardVersionsResponse {
  versions: DashboardVersion[]
  total: number
  limit: number
  offset: number
}

/**
 * Response for listing query versions
 */
export interface GetQueryVersionsResponse {
  versions: QueryVersion[]
  total: number
  limit: number
  offset: number
}

/**
 * Response after restoring a version
 */
export interface RestoreVersionResponse {
  success: boolean
  message: string
  dashboardId?: string
  queryId?: string
  restoredToVersion: number
}

// ============================================================================
// Diff/Comparison Types
// ============================================================================

/**
 * Dashboard version diff
 */
export interface DashboardVersionDiff {
  version1Id: string
  version2Id: string
  nameChanged: boolean
  nameFrom?: string
  nameTo?: string
  descChanged: boolean
  descFrom?: string
  descTo?: string
  filtersChanged: boolean
  filtersFrom?: string
  filtersTo?: string
  layoutChanged: boolean
  layoutFrom?: string
  layoutTo?: string
  cardsDiff: DashboardCardsDiff
}

/**
 * Query version diff
 */
export interface QueryVersionDiff {
  version1Id: string
  version2Id: string
  nameChanged: boolean
  nameFrom?: string
  nameTo?: string
  descChanged: boolean
  descFrom?: string
  descTo?: string
  sqlChanged: boolean
  sqlFrom?: string
  sqlTo?: string
  aiPromptChanged: boolean
  aiPromptFrom?: string
  aiPromptTo?: string
  visualizationChanged: boolean
  visualizationFrom?: Record<string, unknown>
  visualizationTo?: Record<string, unknown>
  tagsChanged: boolean
  tagsAdded: string[]
  tagsRemoved: string[]
}

/**
 * Cards diff within dashboard versions
 */
export interface DashboardCardsDiff {
  added: VersionCard[]
  removed: VersionCard[]
  modified: DashboardCardChange[]
  unchanged: VersionCard[]
}

/**
 * Individual card change
 */
export interface DashboardCardChange {
  before: VersionCard
  after: VersionCard
  changes: string[] // Fields that changed: ["position", "title", "query", "visualization"]
}

/**
 * Diff type for UI
 */
export type DiffType = 'added' | 'removed' | 'modified' | 'unchanged'

// ============================================================================
// UI State Types
// ============================================================================

/**
 * Version history dialog state
 */
export interface VersionHistoryDialogState {
  isOpen: boolean
  resourceType: VersionResourceType
  resourceId: string
  resourceName: string
  versions: DashboardVersion[] | QueryVersion[]
  isLoading: boolean
  error: string | null
  selectedVersions: string[] // IDs of selected versions for comparison
  comparingVersions: boolean
}

/**
 * Version card display props
 */
export interface VersionCardDisplayProps {
  version: DashboardVersion | QueryVersion
  isSelected: boolean
  isAutoSave: boolean
  onSelect: (id: string) => void
  onRestore: (version: DashboardVersion | QueryVersion) => void
  onPreview: (version: DashboardVersion | QueryVersion) => void
  onCompare: (version: DashboardVersion | QueryVersion) => void
}

/**
 * Version timeline grouping
 */
export interface VersionTimelineGroup {
  label: string
  versions: (DashboardVersion | QueryVersion)[]
}

/**
 * Restore dialog state
 */
export interface RestoreDialogState {
  isOpen: boolean
  version: DashboardVersion | QueryVersion | null
  isRestoring: boolean
  error: string | null
}

/**
 * Diff view state
 */
export interface DiffViewState {
  isOpen: boolean
  version1: DashboardVersion | QueryVersion | null
  version2: DashboardVersion | QueryVersion | null
  diff: DashboardVersionDiff | QueryVersionDiff | null
  isLoading: boolean
}

// ============================================================================
// Constants
// ============================================================================

/**
 * Maximum number of versions to select for comparison
 */
export const MAX_COMPARE_VERSIONS = 2

/**
 * Number of auto-save versions to keep
 */
export const MAX_AUTO_SAVE_VERSIONS = 10

/**
 * Auto-save interval in minutes
 */
export const AUTO_SAVE_INTERVAL_MINUTES = 5
