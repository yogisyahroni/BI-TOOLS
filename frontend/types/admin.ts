// Organization types for admin management
export interface Organization {
    id: string;
    name: string;
    description?: string;
    ownerId: string;
    ownerEmail?: string;
    createdAt: string;
    updatedAt: string;
    memberCount: number;
    quota: OrganizationQuota;
    usage: OrganizationUsage;
}

export interface OrganizationQuota {
    maxUsers: number;
    maxQueries: number;
    maxDashboards: number;
    maxStorageMB: number;
    maxWorkspaces: number;
}

export interface OrganizationUsage {
    users: number;
    queries: number;
    dashboards: number;
    storageMB: number;
    workspaces: number;
}

export interface OrganizationMember {
    id: string;
    userId: string;
    workspaceId: string;
    role: string;
    invitedAt: string;
    joinedAt?: string;
    user?: {
        id: string;
        email: string;
        name: string;
        username: string;
    };
}

export interface CreateOrganizationRequest {
    name: string;
    description?: string;
    ownerId: string;
}

export interface UpdateOrganizationRequest {
    name?: string;
    description?: string;
}

export interface AddOrganizationMemberRequest {
    userId: string;
    role: string;
}

export interface UpdateOrganizationQuotaRequest {
    maxUsers?: number;
    maxQueries?: number;
    maxDashboards?: number;
    maxStorageMB?: number;
    maxWorkspaces?: number;
}

// User types for admin management
export interface AdminUser {
    id: string;
    email: string;
    name: string;
    username: string;
    role: string;
    status: string;
    emailVerified: boolean;
    provider?: string;
    createdAt: string;
    lastLoginAt?: string;
    deactivatedAt?: string;
    deactivatedBy?: string;
    deactivationReason?: string;
}

export interface UserStats {
    totalUsers: number;
    activeUsers: number;
    inactiveUsers: number;
    pendingUsers: number;
    verifiedUsers: number;
    newThisMonth: number;
    oauthUsers: number;
}

export interface DeactivateUserRequest {
    reason?: string;
}

export interface UpdateUserRoleRequest {
    role: string;
}

export interface ImpersonateUserResponse {
    token: string;
    expiresAt: string;
    user: {
        id: string;
        email: string;
        name: string;
    };
}

// System health types
export interface SystemHealth {
    status: 'healthy' | 'degraded' | 'unhealthy';
    timestamp: string;
    uptime: number;
    version: string;
    components: {
        [key: string]: ComponentHealth;
    };
}

export interface ComponentHealth {
    status: 'up' | 'down' | 'degraded';
    message?: string;
    details?: Record<string, any>;
}

export interface SystemMetrics {
    timestamp: string;
    memory: MemoryMetrics;
    goroutines: number;
    database: DatabaseMetrics;
    cache: CacheMetrics;
    api: APIMetrics;
}

export interface MemoryMetrics {
    alloc: number;
    totalAlloc: number;
    sys: number;
    numGC: number;
    usagePercent: number;
}

export interface DatabaseMetrics {
    connectionCount: number;
    maxConnections: number;
    idleConnections: number;
    activeConnections: number;
    avgQueryTimeMs: number;
    slowQueries: number;
    totalQueries: number;
}

export interface CacheMetrics {
    hitRate: number;
    missRate: number;
    totalKeys: number;
    memoryUsage: number;
}

export interface APIMetrics {
    requestsPerSecond: number;
    avgResponseTimeMs: number;
    errorRate: number;
    totalRequests: number;
}

export interface DatabaseConnectionInfo {
    id: string;
    name: string;
    type: string;
    status: string;
    lastChecked: string;
    responseTimeMs: number;
}

export interface QueryPerformanceInfo {
    totalQueries: number;
    avgTimeMs: number;
    slowQueries: number;
    failedQueries: number;
    topQueries: TopQuery[];
}

export interface TopQuery {
    query: string;
    count: number;
    avgTimeMs: number;
    lastRun: string;
}
