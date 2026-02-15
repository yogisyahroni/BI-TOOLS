import {
    SemanticModel,
    CreateSemanticModelRequest,
    UpdateSemanticModelRequest,
    SemanticMetric,
    SemanticQueryRequest,
    SemanticQueryResponse
} from '@/types/semantic';

const BASE_URL = '/api/v1/semantic';

async function fetchAPI<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    const res = await fetch(`${BASE_URL}${endpoint}`, {
        ...options,
        headers: {
            'Content-Type': 'application/json',
            ...options.headers,
        },
    });

    if (!res.ok) {
        const errorData = await res.json().catch(() => ({}));
        throw new Error(errorData.error || `API Error: ${res.statusText}`);
    }

    return res.json();
}

export const semanticApi = {
    // Models
    getModels: () => fetchAPI<SemanticModel[]>('/models'),

    getModel: (id: string) => fetchAPI<SemanticModel>(`/models/${id}`),

    createModel: (data: CreateSemanticModelRequest) => fetchAPI<SemanticModel>('/models', {
        method: 'POST',
        body: JSON.stringify(data),
    }),

    updateModel: (id: string, data: UpdateSemanticModelRequest) => fetchAPI<SemanticModel>(`/models/${id}`, {
        method: 'PUT',
        body: JSON.stringify(data),
    }),

    deleteModel: (id: string) => fetchAPI<{ message: string }>(`/models/${id}`, {
        method: 'DELETE',
    }),

    // Metrics (Listing convenience)
    getMetrics: (modelId?: string) => {
        const query = modelId ? `?modelId=${modelId}` : '';
        return fetchAPI<SemanticMetric[]>(`/metrics${query}`);
    },

    // Query
    generateQuery: (data: SemanticQueryRequest) => fetchAPI<SemanticQueryResponse>('/query', {
        method: 'POST',
        body: JSON.stringify(data),
    }),
};
