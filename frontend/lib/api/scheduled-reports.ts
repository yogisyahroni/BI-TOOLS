import type {
    ScheduledReportResponse,
    ScheduledReportListResponse,
    ScheduledReportRun,
    ScheduledReportRunListResponse,
    CreateScheduledReportRequest,
    UpdateScheduledReportRequest,
    ReportPreviewRequest,
    ReportPreviewResponse,
    TriggerReportRequest,
    TriggerReportResponse,
    ToggleReportResponse,
    ScheduledReportFilter,
    ScheduledReportRunFilter,
    TimezoneOption,
} from '@/types/scheduled-reports';
import { fetchWithAuth } from '@/lib/utils';

const API_BASE = '/api/go';

export const scheduledReportsApi = {
    // List all scheduled reports
    list: async (filter?: ScheduledReportFilter): Promise<ScheduledReportListResponse> => {
        const params = new URLSearchParams();

        if (filter?.resourceType) {
            params.set('resourceType', filter.resourceType);
        }
        if (filter?.resourceId) {
            params.set('resourceId', filter.resourceId);
        }
        if (filter?.isActive !== undefined) {
            params.set('isActive', String(filter.isActive));
        }
        if (filter?.scheduleType) {
            params.set('scheduleType', filter.scheduleType);
        }
        if (filter?.search) {
            params.set('search', filter.search);
        }
        if (filter?.page) {
            params.set('page', String(filter.page));
        }
        if (filter?.limit) {
            params.set('limit', String(filter.limit));
        }
        if (filter?.orderBy) {
            params.set('orderBy', filter.orderBy);
        }

        const res = await fetchWithAuth(`${API_BASE}/scheduled-reports?${params.toString()}`);
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to fetch scheduled reports');
        }
        return res.json();
    },

    // Get a single scheduled report
    get: async (id: string): Promise<ScheduledReportResponse> => {
        const res = await fetchWithAuth(`${API_BASE}/scheduled-reports/${id}`);
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to fetch scheduled report');
        }
        return res.json();
    },

    // Create a new scheduled report
    create: async (data: CreateScheduledReportRequest): Promise<ScheduledReportResponse> => {
        const res = await fetchWithAuth(`${API_BASE}/scheduled-reports`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data),
        });
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to create scheduled report');
        }
        return res.json();
    },

    // Update a scheduled report
    update: async (id: string, data: UpdateScheduledReportRequest): Promise<ScheduledReportResponse> => {
        const res = await fetchWithAuth(`${API_BASE}/scheduled-reports/${id}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data),
        });
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to update scheduled report');
        }
        return res.json();
    },

    // Delete a scheduled report
    delete: async (id: string): Promise<void> => {
        const res = await fetchWithAuth(`${API_BASE}/scheduled-reports/${id}`, {
            method: 'DELETE',
        });
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to delete scheduled report');
        }
    },

    // Toggle active status
    toggleActive: async (id: string): Promise<ToggleReportResponse> => {
        const res = await fetchWithAuth(`${API_BASE}/scheduled-reports/${id}/toggle`, {
            method: 'POST',
        });
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to toggle report status');
        }
        return res.json();
    },

    // Trigger a report manually
    trigger: async (id: string, data?: TriggerReportRequest): Promise<TriggerReportResponse> => {
        const res = await fetchWithAuth(`${API_BASE}/scheduled-reports/${id}/trigger`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data || {}),
        });
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to trigger report');
        }
        return res.json();
    },

    // Get report run history
    getHistory: async (id: string, filter?: ScheduledReportRunFilter): Promise<ScheduledReportRunListResponse> => {
        const params = new URLSearchParams();

        if (filter?.status) {
            params.set('status', filter.status);
        }
        if (filter?.startDate) {
            params.set('startDate', filter.startDate);
        }
        if (filter?.endDate) {
            params.set('endDate', filter.endDate);
        }
        if (filter?.page) {
            params.set('page', String(filter.page));
        }
        if (filter?.limit) {
            params.set('limit', String(filter.limit));
        }
        if (filter?.orderBy) {
            params.set('orderBy', filter.orderBy);
        }

        const res = await fetchWithAuth(`${API_BASE}/scheduled-reports/${id}/history?${params.toString()}`);
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to fetch report history');
        }
        return res.json();
    },

    // Preview a report
    preview: async (data: ReportPreviewRequest): Promise<ReportPreviewResponse> => {
        const res = await fetchWithAuth(`${API_BASE}/scheduled-reports/preview`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data),
        });
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to generate preview');
        }
        return res.json();
    },

    // Preview from existing report
    previewFromReport: async (id: string): Promise<ReportPreviewResponse> => {
        const res = await fetchWithAuth(`${API_BASE}/scheduled-reports/${id}/preview`);
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to generate preview');
        }
        return res.json();
    },

    // Get download URL for a run
    getDownloadUrl: async (runId: string): Promise<{ downloadUrl: string }> => {
        const res = await fetchWithAuth(`${API_BASE}/scheduled-reports/runs/${runId}/download`);
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to get download URL');
        }
        return res.json();
    },

    // Get available timezones
    getTimezones: async (): Promise<{ timezones: TimezoneOption[] }> => {
        const res = await fetchWithAuth(`${API_BASE}/scheduled-reports/timezones`);
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to fetch timezones');
        }
        return res.json();
    },
};

// React Query compatible API wrapper
export const scheduledReportsApiClient = {
    list: scheduledReportsApi.list,
    get: scheduledReportsApi.get,
    create: scheduledReportsApi.create,
    update: scheduledReportsApi.update,
    delete: scheduledReportsApi.delete,
    toggleActive: scheduledReportsApi.toggleActive,
    trigger: scheduledReportsApi.trigger,
    getHistory: scheduledReportsApi.getHistory,
    preview: scheduledReportsApi.preview,
    previewFromReport: scheduledReportsApi.previewFromReport,
    getDownloadUrl: scheduledReportsApi.getDownloadUrl,
    getTimezones: scheduledReportsApi.getTimezones,
};
