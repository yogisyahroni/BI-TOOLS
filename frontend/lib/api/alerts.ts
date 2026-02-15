import type {
    Alert,
    AlertListResponse,
    AlertHistory,
    AlertHistoryListResponse,
    AlertStats,
    TriggeredAlert,
    CreateAlertRequest,
    UpdateAlertRequest,
    MuteAlertRequest,
    AcknowledgeAlertRequest,
    TestAlertRequest,
    TestAlertResponse,
    AlertFilter,
    AlertHistoryFilter,
    TriggerAlertCheckResponse,
} from '@/types/alerts';

const API_BASE = '/api/go';

export const alertsApi = {
    // List all alerts
    list: async (filter?: AlertFilter): Promise<AlertListResponse> => {
        const params = new URLSearchParams();

        if (filter?.isActive !== undefined) {
            params.set('isActive', String(filter.isActive));
        }
        if (filter?.state) {
            params.set('state', filter.state);
        }
        if (filter?.severity) {
            params.set('severity', filter.severity);
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

        const res = await fetch(`${API_BASE}/alerts?${params.toString()}`, {
            credentials: 'include',
        });
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to fetch alerts');
        }
        return res.json();
    },

    // Get a single alert
    get: async (id: string): Promise<Alert> => {
        const res = await fetch(`${API_BASE}/alerts/${id}`, {
            credentials: 'include',
        });
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to fetch alert');
        }
        return res.json();
    },

    // Create a new alert
    create: async (data: CreateAlertRequest): Promise<Alert> => {
        const res = await fetch(`${API_BASE}/alerts`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            credentials: 'include',
            body: JSON.stringify(data),
        });
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to create alert');
        }
        return res.json();
    },

    // Update an alert
    update: async (id: string, data: UpdateAlertRequest): Promise<Alert> => {
        const res = await fetch(`${API_BASE}/alerts/${id}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            credentials: 'include',
            body: JSON.stringify(data),
        });
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to update alert');
        }
        return res.json();
    },

    // Delete an alert
    delete: async (id: string): Promise<void> => {
        const res = await fetch(`${API_BASE}/alerts/${id}`, {
            method: 'DELETE',
            credentials: 'include',
        });
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to delete alert');
        }
    },

    // Get alert history
    getHistory: async (id: string, filter?: AlertHistoryFilter): Promise<AlertHistoryListResponse> => {
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

        const res = await fetch(`${API_BASE}/alerts/${id}/history?${params.toString()}`, {
            credentials: 'include',
        });
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to fetch alert history');
        }
        return res.json();
    },

    // Acknowledge an alert
    acknowledge: async (id: string, data?: AcknowledgeAlertRequest): Promise<Alert> => {
        const res = await fetch(`${API_BASE}/alerts/${id}/acknowledge`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            credentials: 'include',
            body: JSON.stringify(data || {}),
        });
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to acknowledge alert');
        }
        return res.json();
    },

    // Mute an alert
    mute: async (id: string, data?: MuteAlertRequest): Promise<Alert> => {
        const res = await fetch(`${API_BASE}/alerts/${id}/mute`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            credentials: 'include',
            body: JSON.stringify(data || {}),
        });
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to mute alert');
        }
        return res.json();
    },

    // Unmute an alert
    unmute: async (id: string): Promise<Alert> => {
        const res = await fetch(`${API_BASE}/alerts/${id}/unmute`, {
            method: 'POST',
            credentials: 'include',
        });
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to unmute alert');
        }
        return res.json();
    },

    // Get triggered alerts
    getTriggered: async (): Promise<{ alerts: TriggeredAlert[] }> => {
        const res = await fetch(`${API_BASE}/alerts/triggered`, {
            credentials: 'include',
        });
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to fetch triggered alerts');
        }
        return res.json();
    },

    // Get alert stats
    getStats: async (): Promise<AlertStats> => {
        const res = await fetch(`${API_BASE}/alerts/stats`, {
            credentials: 'include',
        });
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to fetch alert stats');
        }
        return res.json();
    },

    // Trigger manual alert check
    triggerCheck: async (id: string): Promise<TriggerAlertCheckResponse> => {
        const res = await fetch(`${API_BASE}/alerts/${id}/execute`, {
            method: 'POST',
            credentials: 'include',
        });
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to trigger alert check');
        }
        return res.json();
    },

    // Test alert configuration
    test: async (data: TestAlertRequest): Promise<TestAlertResponse> => {
        const res = await fetch(`${API_BASE}/alerts/test`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            credentials: 'include',
            body: JSON.stringify(data),
        });
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to test alert');
        }
        return res.json();
    },
};

// React Query compatible API wrapper
export const alertsApiClient = {
    list: alertsApi.list,
    get: alertsApi.get,
    create: alertsApi.create,
    update: alertsApi.update,
    delete: alertsApi.delete,
    getHistory: alertsApi.getHistory,
    acknowledge: alertsApi.acknowledge,
    mute: alertsApi.mute,
    unmute: alertsApi.unmute,
    getTriggered: alertsApi.getTriggered,
    getStats: alertsApi.getStats,
    triggerCheck: alertsApi.triggerCheck,
    test: alertsApi.test,
};
