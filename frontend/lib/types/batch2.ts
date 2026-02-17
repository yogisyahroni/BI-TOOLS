/**
 * Batch 2 Types: Pipelines, Dataflows, and Ingestion
 * These types match the Go backend models
 */

// ============================================================================
// PIPELINE TYPES
// ============================================================================

export type TransformStepType = 'FILTER' | 'RENAME' | 'CAST' | 'DEDUPLICATE' | 'AGGREGATE';

export interface TransformStep {
    type: TransformStepType;
    config: Record<string, string>;
    order: number;
}

export type PipelineSourceType = 'POSTGRES' | 'MYSQL' | 'CSV' | 'REST_API';
export type PipelineMode = 'ETL' | 'ELT' | 'batch' | 'stream';
export type PipelineDestinationType = 'INTERNAL_RAW' | 'POSTGRES' | 'MYSQL';

export type ExecutionStatus =
    | 'PENDING'
    | 'PROCESSING'
    | 'EXTRACTING'
    | 'TRANSFORMING'
    | 'LOADING'
    | 'COMPLETED'
    | 'FAILED';

export interface Pipeline {
    id: string;
    name: string;
    description: string | null;
    workspaceId: string;
    sourceType: PipelineSourceType | string;
    sourceConfig: string; // JSON string from JSONB
    connectionId: string | null;
    sourceQuery: string | null;
    destinationType: PipelineDestinationType | string;
    destinationConfig: string | null; // JSON string from JSONB
    mode: PipelineMode | string;
    transformationSteps: string | null; // JSON string from JSONB
    scheduleCron: string | null;
    isActive: boolean;
    rowLimit: number;
    lastRunAt: string | null;
    lastStatus: string | null;
    executions?: JobExecution[];
    qualityRules?: QualityRule[];
    createdAt: string;
    updatedAt: string;
}

export interface ExecutionLog {
    timestamp: string;
    level: 'INFO' | 'WARN' | 'ERROR';
    phase: string;
    message: string;
    detail: string;
}

export interface JobExecution {
    id: string;
    pipelineId: string;
    status: ExecutionStatus;
    startedAt: string;
    completedAt: string | null;
    durationMs: number | null;
    rowsProcessed: number;
    bytesProcessed: number;
    qualityViolations: number;
    progress: number; // 0-100
    error: string | null;
    logs: string | null; // JSON string of ExecutionLog[]
}

export interface QualityRule {
    id: string;
    pipelineId: string;
    ruleType: string;
    config: string; // JSON string
    isActive: boolean;
    createdAt: string;
}

export interface PipelineWithRules extends Pipeline {
    qualityRules: QualityRule[];
}

export interface PipelineStats {
    totalPipelines: number;
    activePipelines: number;
    totalExecutions: number;
    successRate: number;
    totalRowsProcessed: number;
    recentFailures: JobExecution[];
}

export interface PipelineExecutionsResponse {
    executions: JobExecution[];
    total: number;
    limit: number;
    offset: number;
    successRate: number;
    successCount: number;
    failedCount: number;
}

/** SSE progress event payload */
export interface SSEProgressEvent {
    pipelineId: string;
    executionId?: string;
    status: string;
    progress: number;
    elapsedMs?: number;
}

/** Human-readable schedule presets */
export interface SchedulePreset {
    label: string;
    cron: string;
    description: string;
}

export const SCHEDULE_PRESETS: SchedulePreset[] = [
    { label: 'Every 15 minutes', cron: '*/15 * * * *', description: 'Runs every 15 minutes' },
    { label: 'Every 30 minutes', cron: '*/30 * * * *', description: 'Runs every 30 minutes' },
    { label: 'Every hour', cron: '0 * * * *', description: 'Runs at the top of every hour' },
    { label: 'Every 6 hours', cron: '0 */6 * * *', description: 'Runs every 6 hours' },
    { label: 'Every 12 hours', cron: '0 */12 * * *', description: 'Runs twice a day' },
    { label: 'Daily at midnight', cron: '0 0 * * *', description: 'Runs once a day at 00:00 UTC' },
    { label: 'Daily at 6 AM', cron: '0 6 * * *', description: 'Runs daily at 06:00 UTC' },
    { label: 'Weekly on Monday', cron: '0 0 * * 1', description: 'Runs every Monday at midnight' },
    { label: 'Monthly on 1st', cron: '0 0 1 * *', description: 'Runs on the 1st of each month' },
    { label: 'Custom', cron: '', description: 'Enter a custom cron expression' },
];

// ============================================================================
// DATAFLOW TYPES
// ============================================================================

export interface Dataflow {
    id: string;
    name: string;
    description: string | null;
    userId: string;
    schedule: string | null;
    isActive: boolean;
    createdAt: string;
    updatedAt: string;
}

export interface DataflowStep {
    id: string;
    dataflowId: string;
    stepOrder: number;
    stepType: string;
    config: string; // JSON string
    createdAt: string;
}

export interface DataflowRun {
    id: string;
    dataflowId: string;
    status: 'PENDING' | 'RUNNING' | 'COMPLETED' | 'FAILED';
    startedAt: string;
    completedAt: string | null;
    error: string | null;
    createdAt: string;
    updatedAt: string;
}

// ============================================================================
// INGESTION TYPES
// ============================================================================

export interface IngestionRequest {
    workspaceId: string;
    sourceType: 'CSV' | 'JSON' | 'API' | 'DATABASE';
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    sourceConfig: Record<string, any>;
    targetTable: string;
    mode: 'OVERWRITE' | 'APPEND';
}

export interface IngestionPreviewRequest {
    workspaceId: string;
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    sourceType: 'CSV' | 'JSON' | 'API' | 'DATABASE';
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    sourceConfig: Record<string, any>;
    limit?: number;
}

export interface IngestionPreview {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    sourceType: string;
    columns: string[];
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    rows: Record<string, any>[];
    totalRows: number;
    previewLimit: number;
    detectedTypes: Record<string, string>;
}

export interface IngestionResult {
    status: 'PENDING' | 'PROCESSING' | 'COMPLETED' | 'FAILED';
    sourceType: string;
    targetTable: string;
    mode: string;
    rowsIngested: number;
    message: string;
}

// ============================================================================
// REQUEST/RESPONSE TYPES
// ============================================================================

export interface CreatePipelineRequest {
    name: string;
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    description?: string;
    workspaceId: string;
    sourceType: string;
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    sourceConfig: Record<string, any>;
    connectionId?: string;
    sourceQuery?: string;
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    destinationType: string;
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    destinationConfig?: Record<string, any>;
    mode: string;
    transformationSteps?: TransformStep[];
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    qualityRules?: any[];
    scheduleCron?: string;
    rowLimit?: number;
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
}

export interface UpdatePipelineRequest {
    name?: string;
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    description?: string;
    sourceType?: string;
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    sourceConfig?: Record<string, any>;
    connectionId?: string;
    sourceQuery?: string;
    destinationType?: string;
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    destinationConfig?: Record<string, any>;
    mode?: string;
    transformationSteps?: TransformStep[];
    scheduleCron?: string;
    rowLimit?: number;
    isActive?: boolean;
}

export interface CreateDataflowRequest {
    name: string;
    description?: string;
    schedule?: string;
    isActive?: boolean;
}

// ============================================================================
// PAGINATION TYPES
// ============================================================================

export interface PaginationParams {
    limit?: number;
    offset?: number;
}

export interface PaginationMeta {
    total: number;
    limit: number;
    offset: number;
    hasMore: boolean;
}

export interface PaginatedResponse<T> {
    data: T[];
    pagination: PaginationMeta;
}
