import { fetchWithAuth } from '@/lib/utils';
import { type VisualQueryConfig } from '@/types/visual-query';

export interface VisualQuery {
    id: string;
    name: string;
    description?: string;
    connectionId: string;
    collectionId: string;
    userId: string;
    config: VisualQueryConfig;
    generatedSql?: string;
    tags?: string[];
    pinned: boolean;
    createdAt: string;
    updatedAt: string;
}

export const visualQueryApi = {
    create: async (data: Partial<VisualQuery>): Promise<VisualQuery> => {
        const response = await fetchWithAuth('/api/go/visual-queries', {
            method: 'POST',
            body: JSON.stringify(data),
        });
        if (!response.ok) throw new Error('Failed to create visual query');
        const json = await response.json();
        // Return data directly if wrapping or unwrapping is needed, adjust based on backend response
        return json.data;
    },

    get: async (id: string): Promise<VisualQuery> => {
        const response = await fetchWithAuth(`/api/go/visual-queries/${id}`);
        if (!response.ok) throw new Error('Failed to fetch visual query');
        const json = await response.json();
        return json.data;
    },

    update: async (id: string, data: Partial<VisualQuery>): Promise<VisualQuery> => {
        const response = await fetchWithAuth(`/api/go/visual-queries/${id}`, {
            method: 'PUT',
            body: JSON.stringify(data),
        });
        if (!response.ok) throw new Error('Failed to update visual query');
        const json = await response.json();
        return json.data;
    },

    delete: async (id: string): Promise<void> => {
        const response = await fetchWithAuth(`/api/go/visual-queries/${id}`, {
            method: 'DELETE',
        });
        if (!response.ok) throw new Error('Failed to delete visual query');
    },

    generateSql: async (config: VisualQueryConfig): Promise<{ sql: string, params: any[] }> => {
        const response = await fetchWithAuth(`/api/go/visual-queries/generate-sql`, {
            method: 'POST',
            body: JSON.stringify(config),
        });
        if (!response.ok) throw new Error('Failed to generate SQL');
        const json = await response.json();
        return json.data;
    },

    preview: async (id: string, config?: VisualQueryConfig): Promise<any[]> => {
        const response = await fetchWithAuth(`/api/go/visual-queries/${id}/preview`, { // Changed to specific ID preview or general preview
            method: 'POST',
            body: JSON.stringify(config), // If previewing changes before save
        });
        if (!response.ok) throw new Error('Failed to preview query');
        const json = await response.json();
        return json.data; // Assuming results are in data field
    }
};
