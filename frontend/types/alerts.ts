// Alert System TypeScript Types
// Task: TASK-102

// Enums
export type AlertSeverity = 'critical' | 'warning' | 'info';
export type AlertState = 'ok' | 'triggered' | 'acknowledged' | 'muted' | 'error';
export type AlertNotificationChannel = 'email' | 'webhook' | 'in_app' | 'slack' | 'teams';
export type AlertHistoryStatus = 'triggered' | 'ok' | 'error';
export type AlertOperator = '>' | '<' | '=' | '>=' | '<=' | '!=';

// Interfaces
export interface Alert {
    id: string;
    name: string;
    description?: string;
    queryId: string;
    userId: string;

    // Condition configuration
    column: string;
    operator: AlertOperator;
    threshold: number;

    // Schedule configuration
    schedule: string;
    timezone: string;

    // Status and metadata
    isActive: boolean;
    severity: AlertSeverity;
    state: AlertState;

    // Runtime tracking
    lastRunAt?: string;
    lastStatus?: string;
    lastValue?: number;
    lastError?: string;
    nextRunAt?: string;

    // Cooldown and throttling
    cooldownMinutes: number;
    lastTriggeredAt?: string;
    triggerCount: number;
    notificationCount: number;

    // Mute configuration
    isMuted: boolean;
    mutedUntil?: string;
    muteDuration?: number;

    // Channels
    channels?: AlertNotificationChannelConfig[];

    // Legacy fields
    email?: string;
    webhookUrl?: string;

    // Relationships
    query?: {
        id: string;
        name: string;
        sql: string;
    };
    user?: {
        id: string;
        name: string;
        email: string;
    };

    createdAt: string;
    updatedAt: string;
}

export interface AlertNotificationChannelConfig {
    id: string;
    alertId: string;
    channelType: AlertNotificationChannel;
    isEnabled: boolean;
    config?: Record<string, unknown>;
    createdAt: string;
    updatedAt: string;
}

export interface AlertHistory {
    id: string;
    alertId: string;
    status: AlertHistoryStatus;
    value?: number;
    threshold: number;
    message?: string;
    errorMessage?: string;
    queryDuration: number; // milliseconds
    checkedAt: string;
    notifications?: AlertNotificationLog[];
}

export interface AlertNotificationLog {
    id: string;
    historyId: string;
    channelType: AlertNotificationChannel;
    status: string; // sent, failed, pending
    recipient?: string;
    error?: string;
    sentAt?: string;
    createdAt: string;
}

export interface AlertAcknowledgment {
    id: string;
    alertId: string;
    userId: string;
    note?: string;
    acknowledgedAt: string;
    user?: {
        id: string;
        name: string;
    };
}

export interface TriggeredAlert {
    alert: Alert;
    triggeredAt: string;
    currentValue: number;
    acknowledged: boolean;
    acknowledgedAt?: string;
    acknowledgedBy?: string;
}

export interface AlertStats {
    total: number;
    active: number;
    triggered: number;
    acknowledged: number;
    muted: number;
    error: number;
    bySeverity: Record<AlertSeverity, number>;
}

// Request Types
export interface CreateAlertRequest {
    name: string;
    description?: string;
    queryId: string;
    column: string;
    operator: AlertOperator;
    threshold: number;
    schedule: string;
    timezone?: string;
    severity?: AlertSeverity;
    cooldownMinutes?: number;
    channels?: AlertChannelInput[];
}

export interface UpdateAlertRequest {
    name?: string;
    description?: string;
    column?: string;
    operator?: AlertOperator;
    threshold?: number;
    schedule?: string;
    timezone?: string;
    isActive?: boolean;
    severity?: AlertSeverity;
    cooldownMinutes?: number;
    channels?: AlertChannelInput[];
}

export interface AlertChannelInput {
    channelType: AlertNotificationChannel;
    isEnabled: boolean;
    config?: Record<string, unknown>;
}

export interface MuteAlertRequest {
    duration?: number; // minutes, undefined for indefinite
    reason?: string;
}

export interface AcknowledgeAlertRequest {
    note?: string;
}

export interface TestAlertRequest {
    name: string;
    description?: string;
    queryId: string;
    column: string;
    operator: AlertOperator;
    threshold: number;
}

export interface TestAlertResponse {
    triggered: boolean;
    value: number;
    threshold: number;
    message: string;
    queryTime: number; // milliseconds
    result?: Record<string, unknown>;
}

export interface AlertExecutionResult {
    triggered: boolean;
    value: number;
    message: string;
}

// Response Types
export interface AlertListResponse {
    alerts: Alert[];
    total: number;
    page: number;
    limit: number;
}

export interface AlertHistoryListResponse {
    history: AlertHistory[];
    total: number;
    page: number;
    limit: number;
}

export interface TriggerAlertCheckResponse {
    success: boolean;
    history: AlertHistory;
}

// Filter Types
export interface AlertFilter {
    queryId?: string;
    isActive?: boolean;
    state?: AlertState;
    severity?: AlertSeverity;
    search?: string;
    page?: number;
    limit?: number;
    orderBy?: string;
}

export interface AlertHistoryFilter {
    status?: AlertHistoryStatus;
    startDate?: string;
    endDate?: string;
    page?: number;
    limit?: number;
    orderBy?: string;
}

// Component Props Types
export interface AlertCardProps {
    alert: Alert;
    onEdit?: (alert: Alert) => void;
    onDelete?: (alert: Alert) => void;
    onAcknowledge?: (alert: Alert) => void;
    onMute?: (alert: Alert) => void;
    onUnmute?: (alert: Alert) => void;
}

export interface AlertListProps {
    alerts: Alert[];
    loading?: boolean;
    onEdit?: (alert: Alert) => void;
    onDelete?: (alert: Alert) => void;
    onAcknowledge?: (alert: Alert) => void;
    onMute?: (alert: Alert) => void;
    onUnmute?: (alert: Alert) => void;
}

export interface ConditionBuilderProps {
    column: string;
    operator: AlertOperator;
    threshold: number;
    availableColumns?: string[];
    onChange: (values: { column: string; operator: AlertOperator; threshold: number }) => void;
    disabled?: boolean;
}

export interface NotificationConfigProps {
    channels: AlertChannelInput[];
    onChange: (channels: AlertChannelInput[]) => void;
    disabled?: boolean;
}

export interface AlertHistoryProps {
    history: AlertHistory[];
    loading?: boolean;
    hasMore?: boolean;
    onLoadMore?: () => void;
}

export interface TriggeredAlertsProps {
    alerts: TriggeredAlert[];
    onAcknowledge?: (alertId: string) => void;
    onAcknowledgeAll?: () => void;
    loading?: boolean;
}

// Constants
export const ALERT_SEVERITIES: { value: AlertSeverity; label: string; color: string; icon: string }[] = [
    { value: 'critical', label: 'Critical', color: '#dc2626', icon: 'AlertCircle' },
    { value: 'warning', label: 'Warning', color: '#f59e0b', icon: 'AlertTriangle' },
    { value: 'info', label: 'Info', color: '#3b82f6', icon: 'Info' },
];

export const ALERT_STATES: { value: AlertState; label: string; color: string }[] = [
    { value: 'ok', label: 'OK', color: '#22c55e' },
    { value: 'triggered', label: 'Triggered', color: '#dc2626' },
    { value: 'acknowledged', label: 'Acknowledged', color: '#6366f1' },
    { value: 'muted', label: 'Muted', color: '#6b7280' },
    { value: 'error', label: 'Error', color: '#7c3aed' },
];

export const ALERT_OPERATORS: { value: AlertOperator; label: string }[] = [
    { value: '>', label: 'Greater than (>)' },
    { value: '<', label: 'Less than (<)' },
    { value: '=', label: 'Equal to (=)' },
    { value: '>=', label: 'Greater than or equal to (>=)' },
    { value: '<=', label: 'Less than or equal to (<=)' },
    { value: '!=', label: 'Not equal to (!=)' },
];

export const ALERT_CHANNELS: { value: AlertNotificationChannel; label: string; icon: string }[] = [
    { value: 'email', label: 'Email', icon: 'Mail' },
    { value: 'webhook', label: 'Webhook', icon: 'Webhook' },
    { value: 'in_app', label: 'In-App', icon: 'Bell' },
    { value: 'slack', label: 'Slack', icon: 'MessageSquare' },
    { value: 'teams', label: 'Microsoft Teams', icon: 'Users' },
];

export const SCHEDULE_OPTIONS: { value: string; label: string }[] = [
    { value: '1m', label: 'Every 1 minute' },
    { value: '5m', label: 'Every 5 minutes' },
    { value: '15m', label: 'Every 15 minutes' },
    { value: '30m', label: 'Every 30 minutes' },
    { value: '1h', label: 'Every 1 hour' },
    { value: '0 * * * *', label: 'Every hour (on the hour)' },
    { value: '0 */6 * * *', label: 'Every 6 hours' },
    { value: '0 0 * * *', label: 'Daily (midnight)' },
    { value: '0 9 * * 1', label: 'Weekly (Monday 9 AM)' },
];

export const COOLDOWN_OPTIONS: { value: number; label: string }[] = [
    { value: 0, label: 'No cooldown' },
    { value: 5, label: '5 minutes' },
    { value: 15, label: '15 minutes' },
    { value: 30, label: '30 minutes' },
    { value: 60, label: '1 hour' },
    { value: 180, label: '3 hours' },
    { value: 360, label: '6 hours' },
    { value: 720, label: '12 hours' },
    { value: 1440, label: '24 hours' },
];

export const MUTE_DURATION_OPTIONS: { value: number; label: string }[] = [
    { value: 30, label: '30 minutes' },
    { value: 60, label: '1 hour' },
    { value: 180, label: '3 hours' },
    { value: 360, label: '6 hours' },
    { value: 720, label: '12 hours' },
    { value: 1440, label: '24 hours' },
    { value: 10080, label: '7 days' },
];

// Helper Types
export interface AlertFormData {
    name: string;
    description: string;
    queryId: string;
    queryName?: string;
    column: string;
    operator: AlertOperator;
    threshold: number;
    schedule: string;
    timezone: string;
    severity: AlertSeverity;
    cooldownMinutes: number;
    channels: AlertChannelInput[];
}

export interface AlertWizardStep {
    id: number;
    title: string;
    description: string;
    isComplete: boolean;
}

export type AlertWizardStepType = 'basic' | 'query' | 'condition' | 'schedule' | 'notifications' | 'review';
