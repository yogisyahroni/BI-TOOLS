// Scheduled Reports TypeScript Types
// Task: TASK-100

// Enums
export type ReportScheduleType = 'daily' | 'weekly' | 'monthly' | 'cron';
export type ReportFormat = 'pdf' | 'csv' | 'excel' | 'png';
export type ReportResourceType = 'dashboard' | 'query';
export type ReportRunStatus = 'pending' | 'running' | 'success' | 'failed' | 'cancelled';
export type RecipientType = 'to' | 'cc' | 'bcc';

// Interfaces
export interface ScheduledReport {
    id: string;
    name: string;
    description?: string;

    // Resource
    resourceType: ReportResourceType;
    resourceId: string;

    // Schedule
    scheduleType: ReportScheduleType;
    cronExpr?: string;
    timeOfDay?: string;
    dayOfWeek?: number;
    dayOfMonth?: number;
    timezone: string;

    // Recipients
    recipients: ScheduledReportRecipient[];

    // Format & Options
    format: ReportFormat;
    includeFilters: boolean;
    subject: string;
    message: string;

    // Status
    isActive: boolean;
    lastRunAt?: string;
    lastRunStatus?: 'success' | 'failed';
    lastRunError?: string;
    nextRunAt?: string;
    successCount: number;
    failureCount: number;
    consecutiveFail: number;

    // Options
    options?: Record<string, unknown>;

    // Ownership
    createdBy: string;
    createdAt: string;
    updatedAt: string;
}

export interface ScheduledReportRecipient {
    id: string;
    reportId: string;
    email: string;
    type: RecipientType;
    createdAt: string;
}

export interface ScheduledReportRun {
    id: string;
    reportId: string;
    startedAt: string;
    completedAt?: string;
    status: ReportRunStatus;
    errorMessage?: string;
    fileUrl?: string;
    fileSize?: number;
    fileType?: string;
    durationMs?: number;
    triggeredBy?: string;
    recipientCount: number;
    createdAt: string;
}

// Request types
export interface CreateScheduledReportRequest {
    name: string;
    description?: string;

    // Resource
    resourceType: ReportResourceType;
    resourceId: string;

    // Schedule
    scheduleType: ReportScheduleType;
    cronExpr?: string;
    timeOfDay?: string;
    dayOfWeek?: number;
    dayOfMonth?: number;
    timezone?: string;

    // Recipients
    recipients: RecipientInput[];

    // Format & Options
    format: ReportFormat;
    includeFilters?: boolean;
    subject?: string;
    message?: string;

    // Additional options
    options?: Record<string, unknown>;
}

export interface RecipientInput {
    email: string;
    type?: RecipientType;
}

export interface UpdateScheduledReportRequest {
    name?: string;
    description?: string;

    // Schedule
    scheduleType?: ReportScheduleType;
    cronExpr?: string;
    timeOfDay?: string;
    dayOfWeek?: number;
    dayOfMonth?: number;
    timezone?: string;

    // Recipients
    recipients?: RecipientInput[];

    // Format & Options
    format?: ReportFormat;
    includeFilters?: boolean;
    subject?: string;
    message?: string;
    options?: Record<string, unknown>;
}

export interface ReportPreviewRequest {
    resourceType: ReportResourceType;
    resourceId: string;
    format: ReportFormat;
    includeFilters?: boolean;
}

export interface TriggerReportRequest {
    triggerType?: 'manual' | 'api';
}

// Response types
export interface ScheduledReportResponse extends ScheduledReport {
    recipients: RecipientResponse[];
}

export interface RecipientResponse {
    id: string;
    email: string;
    type: RecipientType;
}

export interface ScheduledReportListResponse {
    reports: ScheduledReportResponse[];
    total: number;
    page: number;
    limit: number;
}

export interface ScheduledReportRunListResponse {
    runs: ScheduledReportRun[];
    total: number;
    page: number;
    limit: number;
}

export interface ReportPreviewResponse {
    previewUrl: string;
    fileSize: number;
    expiresAt: string;
}

export interface TriggerReportResponse {
    runId: string;
    status: ReportRunStatus;
    message: string;
    startedAt: string;
}

export interface ToggleReportResponse {
    id: string;
    isActive: boolean;
    message: string;
}

// Filter types
export interface ScheduledReportFilter {
    resourceType?: ReportResourceType;
    resourceId?: string;
    isActive?: boolean;
    scheduleType?: ReportScheduleType;
    search?: string;
    page?: number;
    limit?: number;
    orderBy?: string;
}

export interface ScheduledReportRunFilter {
    status?: ReportRunStatus;
    startDate?: string;
    endDate?: string;
    page?: number;
    limit?: number;
    orderBy?: string;
}

// UI Types
export interface TimezoneOption {
    value: string;
    label: string;
}

export interface ScheduleFormData {
    name: string;
    description: string;
    resourceType: ReportResourceType;
    resourceId: string;
    resourceName?: string;
    scheduleType: ReportScheduleType;
    cronExpr: string;
    timeOfDay: string;
    dayOfWeek: number;
    dayOfMonth: number;
    timezone: string;
    recipients: RecipientInput[];
    format: ReportFormat;
    includeFilters: boolean;
    subject: string;
    message: string;
}

export interface ReportRunHistoryItem {
    id: string;
    startedAt: string;
    completedAt?: string;
    status: ReportRunStatus;
    errorMessage?: string;
    fileUrl?: string;
    fileSize?: number;
    fileType?: string;
    durationMs?: number;
    triggeredBy?: string;
}

// Component Props Types
export interface ReportScheduleFormProps {
    initialData?: Partial<ScheduleFormData>;
    onSubmit: (data: CreateScheduledReportRequest) => Promise<void>;
    onCancel: () => void;
    isSubmitting?: boolean;
}

export interface ReportScheduleCardProps {
    report: ScheduledReportResponse;
    onEdit: (report: ScheduledReportResponse) => void;
    onDelete: (report: ScheduledReportResponse) => void;
    onToggle: (report: ScheduledReportResponse) => void;
    onRunNow: (report: ScheduledReportResponse) => void;
    onViewHistory: (report: ScheduledReportResponse) => void;
}

export interface ReportHistoryProps {
    runs: ScheduledReportRun[];
    loading?: boolean;
    hasMore?: boolean;
    onLoadMore?: () => void;
    onDownload?: (run: ScheduledReportRun) => void;
}

export interface RecipientManagerProps {
    recipients: RecipientInput[];
    onChange: (recipients: RecipientInput[]) => void;
    disabled?: boolean;
    error?: string;
}

export interface SchedulePickerProps {
    scheduleType: ReportScheduleType;
    cronExpr?: string;
    timeOfDay?: string;
    dayOfWeek?: number;
    dayOfMonth?: number;
    timezone?: string;
    onChange: (values: {
        scheduleType: ReportScheduleType;
        cronExpr?: string;
        timeOfDay?: string;
        dayOfWeek?: number;
        dayOfMonth?: number;
        timezone?: string;
    }) => void;
    disabled?: boolean;
    error?: string;
}

// Helper types
export type DayOfWeek = 0 | 1 | 2 | 3 | 4 | 5 | 6;

export const DAYS_OF_WEEK: { value: DayOfWeek; label: string }[] = [
    { value: 0, label: 'Sunday' },
    { value: 1, label: 'Monday' },
    { value: 2, label: 'Tuesday' },
    { value: 3, label: 'Wednesday' },
    { value: 4, label: 'Thursday' },
    { value: 5, label: 'Friday' },
    { value: 6, label: 'Saturday' },
];

export const REPORT_FORMATS: { value: ReportFormat; label: string; icon?: string }[] = [
    { value: 'pdf', label: 'PDF Document' },
    { value: 'csv', label: 'CSV Spreadsheet' },
    { value: 'excel', label: 'Excel Spreadsheet' },
    { value: 'png', label: 'PNG Image' },
];

export const SCHEDULE_TYPES: { value: ReportScheduleType; label: string; description: string }[] = [
    { value: 'daily', label: 'Daily', description: 'Send report every day at a specific time' },
    { value: 'weekly', label: 'Weekly', description: 'Send report on a specific day of the week' },
    { value: 'monthly', label: 'Monthly', description: 'Send report on a specific day of the month' },
    { value: 'cron', label: 'Custom (Cron)', description: 'Use a custom cron expression' },
];

export const RESOURCE_TYPES: { value: ReportResourceType; label: string }[] = [
    { value: 'dashboard', label: 'Dashboard' },
    { value: 'query', label: 'Query' },
];
