import type {
    Organization,
    OrganizationMember,
    CreateOrganizationRequest,
    UpdateOrganizationRequest,
    AddOrganizationMemberRequest,
    UpdateOrganizationQuotaRequest,
    AdminUser,
    UserStats,
    DeactivateUserRequest,
    UpdateUserRoleRequest,
    ImpersonateUserResponse,
    SystemHealth,
    SystemMetrics,
    DatabaseConnectionInfo,
    QueryPerformanceInfo,
} from '@/types/admin';

// Backend API base URL
const API_BASE = process.env.NEXT_PUBLIC_BACKEND_URL || 'http://localhost:8080';

// Helper function for authenticated requests
const authFetch = async (url: string, options: RequestInit = {}) => {
    const token = typeof window !== 'undefined' ? localStorage.getItem('authToken') : null;
    const headers: Record<string, string> = {
        'Content-Type': 'application/json',
        ...(options.headers as Record<string, string>),
    };

    if (token) {
        headers.Authorization = `Bearer ${token}`;
    }

    const response = await fetch(`${API_BASE}${url}`, {
        ...options,
        headers,
        credentials: 'include',
    });

    if (!response.ok) {
        const error = await response.json().catch(() => ({ message: 'Request failed' }));
        throw new Error(error.message || `Request failed with status ${response.status}`);
    }

    // Handle 204 No Content
    if (response.status === 204) {
        return null;
    }

    return response.json();
};

// Organization Management API
export const organizationApi = {
    // List organizations with pagination
    async list(params?: {
        page?: number;
        pageSize?: number;
        search?: string;
        status?: string;
    }) {
        const queryParams = new URLSearchParams();
        if (params?.page) queryParams.set('page', params.page.toString());
        if (params?.pageSize) queryParams.set('pageSize', params.pageSize.toString());
        if (params?.search) queryParams.set('search', params.search);
        if (params?.status) queryParams.set('status', params.status);

        return authFetch(`/api/admin/organizations?${queryParams.toString()}`);
    },

    // Get organization by ID
    async get(id: string) {
        return authFetch(`/api/admin/organizations/${id}`) as Promise<Organization>;
    },

    // Create new organization
    async create(payload: CreateOrganizationRequest) {
        return authFetch(`/api/admin/organizations`, {
            method: 'POST',
            body: JSON.stringify(payload),
        }) as Promise<Organization>;
    },

    // Update organization
    async update(id: string, payload: UpdateOrganizationRequest) {
        return authFetch(`/api/admin/organizations/${id}`, {
            method: 'PUT',
            body: JSON.stringify(payload),
        }) as Promise<Organization>;
    },

    // Delete organization
    async delete(id: string) {
        return authFetch(`/api/admin/organizations/${id}`, { method: 'DELETE' });
    },

    // Get organization stats
    async getStats() {
        return authFetch('/api/admin/organizations/stats');
    },

    // Member management
    members: {
        async list(orgId: string) {
            return authFetch(`/api/admin/organizations/${orgId}/members`) as Promise<{ data: OrganizationMember[] }>;
        },

        async add(orgId: string, payload: AddOrganizationMemberRequest) {
            return authFetch(`/api/admin/organizations/${orgId}/members`, {
                method: 'POST',
                body: JSON.stringify(payload),
            }) as Promise<OrganizationMember>;
        },

        async remove(orgId: string, userId: string) {
            return authFetch(`/api/admin/organizations/${orgId}/members/${userId}`, {
                method: 'DELETE',
            });
        },

        async updateRole(orgId: string, userId: string, role: string) {
            return authFetch(`/api/admin/organizations/${orgId}/members/${userId}/role`, {
                method: 'PUT',
                body: JSON.stringify({ role }),
            }) as Promise<OrganizationMember>;
        },
    },

    // Quota management
    quotas: {
        async get(orgId: string) {
            return authFetch(`/api/admin/organizations/${orgId}/quotas`);
        },

        async update(orgId: string, payload: UpdateOrganizationQuotaRequest) {
            return authFetch(`/api/admin/organizations/${orgId}/quotas`, {
                method: 'PUT',
                body: JSON.stringify(payload),
            });
        },

        async refresh(orgId: string) {
            return authFetch(`/api/admin/organizations/${orgId}/quotas/refresh`, {
                method: 'POST',
            });
        },
    },
};

// User Management API
export const userAdminApi = {
    // List users with pagination and filters
    async list(params?: {
        page?: number;
        pageSize?: number;
        search?: string;
        status?: string;
        role?: string;
    }) {
        const queryParams = new URLSearchParams();
        if (params?.page) queryParams.set('page', params.page.toString());
        if (params?.pageSize) queryParams.set('pageSize', params.pageSize.toString());
        if (params?.search) queryParams.set('search', params.search);
        if (params?.status) queryParams.set('status', params.status);
        if (params?.role) queryParams.set('role', params.role);

        return authFetch(`/api/admin/users?${queryParams.toString()}`);
    },

    // Get user by ID
    async get(id: string) {
        return authFetch(`/api/admin/users/${id}`) as Promise<AdminUser>;
    },

    // Get user statistics
    async getStats() {
        return authFetch('/api/admin/users/stats') as Promise<UserStats>;
    },

    // Activate user
    async activate(id: string) {
        return authFetch(`/api/admin/users/${id}/activate`, { method: 'PUT' });
    },

    // Deactivate user
    async deactivate(id: string, payload?: DeactivateUserRequest) {
        return authFetch(`/api/admin/users/${id}/deactivate`, {
            method: 'PUT',
            body: payload ? JSON.stringify(payload) : undefined,
        });
    },

    // Update user role
    async updateRole(id: string, payload: UpdateUserRoleRequest) {
        return authFetch(`/api/admin/users/${id}/role`, {
            method: 'PUT',
            body: JSON.stringify(payload),
        });
    },

    // Impersonate user
    async impersonate(id: string) {
        return authFetch(`/api/admin/users/${id}/impersonate`, {
            method: 'POST',
        }) as Promise<ImpersonateUserResponse>;
    },

    // Get user activity logs
    async getActivity(id: string, params?: {
        page?: number;
        pageSize?: number;
    }) {
        const queryParams = new URLSearchParams();
        if (params?.page) queryParams.set('page', params.page.toString());
        if (params?.pageSize) queryParams.set('pageSize', params.pageSize.toString());

        return authFetch(`/api/admin/users/${id}/activity?${queryParams.toString()}`);
    },
};

// System Health & Monitoring API
export const systemAdminApi = {
    // Get system health status
    async getHealth() {
        return authFetch('/api/admin/system/health') as Promise<SystemHealth>;
    },

    // Get system metrics
    async getMetrics() {
        return authFetch('/api/admin/system/metrics') as Promise<SystemMetrics>;
    },

    // Get database connections
    async getDatabaseConnections() {
        return authFetch('/api/admin/system/database/connections') as Promise<{
            connections: DatabaseConnectionInfo[];
            total: number;
        }>;
    },

    // Get database performance metrics
    async getDatabasePerformance() {
        return authFetch('/api/admin/system/database/performance') as Promise<QueryPerformanceInfo>;
    },

    // Get cache statistics
    async getCacheStats() {
        return authFetch('/api/admin/system/cache/stats');
    },
};
