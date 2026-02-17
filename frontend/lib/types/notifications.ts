// Notification Types
export interface Notification {
    id: string;
    userId: string;
    title: string;
    message: string;
    type: 'info' | 'success' | 'warning' | 'error' | 'system';
    isRead: boolean;
    createdAt: Date;
}

export interface CreateNotificationInput {
    userId: string;
    title: string;
    message: string;
    type?: 'info' | 'success' | 'warning' | 'error' | 'system';
}

export interface NotificationResponse {
    notifications: Notification[];
    total: number;
    limit: number;
    offset: number;
}

export interface UnreadCountResponse {
    count: number;
}

// Activity Log Types
export interface ActivityLog {
    id: string;
    userId: string;
    workspaceId?: string;
    action: string;
    entityType: string;
    entityId?: string;
    description: string; // Human-readable description of the activity
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    metadata?: Record<string, any>;
    createdAt: Date;
}

export interface ActivityFeedResponse {
    activities: ActivityLog[];
    total: number;
    limit: number;
    offset: number;
}

export interface LogActivityInput {
    userId: string;
    workspaceId?: string;
    action: string;
    entityType: string;
    entityId?: string;
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    description: string;
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    metadata?: Record<string, any>;
}

// Scheduler Job Types
export interface SchedulerJob {
    id: string;
    name: string;
    schedule: string;
    status: 'active' | 'paused' | 'error';
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    enabled?: boolean;
    jobType?: string;
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    payload?: Record<string, any>;
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    lastRun?: Date;
    nextRun?: Date;
    lastError?: string;
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    config?: Record<string, any>;
    createdAt: Date;
    updatedAt: Date;
}

export interface CreateSchedulerJobInput {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    name: string;
    schedule: string;
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    status?: 'active' | 'paused';
    jobType?: string;
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    payload?: Record<string, any>;
    enabled?: boolean;
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    config?: Record<string, any>;
}

export interface UpdateSchedulerJobInput {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    name?: string;
    schedule?: string;
    status?: 'active' | 'paused' | 'error';
    jobType?: string;
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    payload?: Record<string, any>;
    enabled?: boolean;
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    config?: Record<string, any>;
}

// Alias for compatibility
export type CreateSchedulerJobRequest = CreateSchedulerJobInput;

export interface UpdateSchedulerJobInput {
    name?: string;
    schedule?: string;
    status?: 'active' | 'paused' | 'error';
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    config?: Record<string, any>;
}
        // eslint-disable-next-line @typescript-eslint/no-explicit-any

export interface SchedulerJobsResponse {
    jobs: SchedulerJob[];
}

// WebSocket Message Types
export interface WebSocketMessage {
    type: 'notification' | 'activity' | 'system';
    userId: string;
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    payload: any;
}

export interface NotificationWebSocketPayload {
    notification: Notification;
}

export interface ActivityWebSocketPayload {
    activity: ActivityLog;
}

export interface SystemWebSocketPayload {
    message: string;
    level: 'info' | 'warning' | 'error';
}

// WebSocket Connection State
export interface WebSocketState {
    connected: boolean;
    connecting: boolean;
    error?: string;
}

// WebSocket Stats
export interface WebSocketStats {
    connectedUsers: string[];
    totalConnections: number;
    timestamp: Date;
}
