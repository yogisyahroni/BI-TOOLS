import { toast } from 'sonner';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

type RequestMethod = 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';

interface ApiRequestOptions extends RequestInit {
    params?: Record<string, string>;
}

async function request<T>(endpoint: string, method: RequestMethod, data?: any, options?: ApiRequestOptions): Promise<{ data: T; status: number; statusText: string }> {
    const url = new URL(`${API_BASE_URL}${endpoint}`);

    if (options?.params) {
        Object.entries(options.params).forEach(([key, value]) => {
            url.searchParams.append(key, value);
        });
    }

    const headers: HeadersInit = {
        'Content-Type': 'application/json',
        ...options?.headers,
    };

    // Add auth token if available (example)
    // const token = localStorage.getItem('token');
    // if (token) {
    //     headers['Authorization'] = `Bearer ${token}`;
    // }

    const config: RequestInit = {
        method,
        headers,
        ...options,
    };

    if (data) {
        config.body = JSON.stringify(data);
    }

    try {
        const response = await fetch(url.toString(), config);

        let responseData: any;
        const contentType = response.headers.get('content-type');
        if (contentType && contentType.includes('application/json')) {
            responseData = await response.json();
        } else {
            responseData = await response.text();
        }

        if (!response.ok) {
            const errorMessage = responseData?.error || responseData?.message || `API Error: ${response.statusText}`;
            // toast.error(errorMessage); // Optional: global error handling
            throw new Error(errorMessage);
        }

        return {
            data: responseData,
            status: response.status,
            statusText: response.statusText,
        };
    } catch (error: any) {
        console.error('API Request Failed:', error);
        throw error;
    }
}

export const api = {
    get: <T>(endpoint: string, options?: ApiRequestOptions) => request<T>(endpoint, 'GET', undefined, options),
    post: <T>(endpoint: string, data?: any, options?: ApiRequestOptions) => request<T>(endpoint, 'POST', data, options),
    put: <T>(endpoint: string, data?: any, options?: ApiRequestOptions) => request<T>(endpoint, 'PUT', data, options),
    delete: <T>(endpoint: string, options?: ApiRequestOptions) => request<T>(endpoint, 'DELETE', undefined, options),
    patch: <T>(endpoint: string, data?: any, options?: ApiRequestOptions) => request<T>(endpoint, 'PATCH', data, options),
};

export default api;
