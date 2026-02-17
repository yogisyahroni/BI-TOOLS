import { type Dashboard, type DashboardCard, _VisualizationConfig } from '@/lib/types';
import { fetchWithAuth } from '@/lib/utils';

export interface CreateDashboardInput {
    name: string;
    description?: string;
    collectionId: string;
    tags?: string[];
}

export interface UpdateDashboardInput {
    name?: string;
    description?: string;
    isPublic?: boolean;
    collectionId?: string;
    tags?: string[];
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    filters?: any[];
    cards?: DashboardCard[]; // Full sync of layout
}

export const dashboardService = {
    async getAll(): Promise<Dashboard[]> {
        const response = await fetchWithAuth('/api/go/dashboards', {
            credentials: 'include',
        });
        if (!response.ok) throw new Error('Failed to fetch dashboards');
        const json = await response.json();
        return json.data;
    },

    async getById(id: string): Promise<Dashboard> {
        const response = await fetchWithAuth(`/api/go/dashboards/${id}`, {
            credentials: 'include',
        });
        if (!response.ok) throw new Error('Failed to fetch dashboard');
        const json = await response.json();
        return json.data;
    },

    async create(input: CreateDashboardInput): Promise<Dashboard> {
        const response = await fetchWithAuth('/api/go/dashboards', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            credentials: 'include',
            body: JSON.stringify(input),
        });
        if (!response.ok) throw new Error('Failed to create dashboard');
        const json = await response.json();
        return json.data;
    },

    async update(id: string, updates: UpdateDashboardInput): Promise<Dashboard> {
        const response = await fetchWithAuth(`/api/go/dashboards/${id}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            credentials: 'include',
            body: JSON.stringify(updates),
        });
        if (!response.ok) throw new Error('Failed to update dashboard');
        const json = await response.json();
        return json.data;
    },

    async delete(id: string): Promise<void> {
        const response = await fetchWithAuth(`/api/go/dashboards/${id}`, {
            method: 'DELETE',
            credentials: 'include',
        });
        if (!response.ok) throw new Error('Failed to delete dashboard');
    }
};
