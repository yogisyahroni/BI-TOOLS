'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import {
    AlertCircle,
    AlertTriangle,
    Info,
    Clock,
    Bell,
    Mail,
    Webhook,
    MessageSquare,
} from 'lucide-react';
import type { AlertSeverity, AlertOperator, AlertNotificationChannel } from '@/types/alerts';

interface AlertReviewProps {
    formData: {
        name: string;
        description?: string;
        severity: AlertSeverity;
        queryId: string;
        queryName?: string;
        column: string;
        operator: AlertOperator;
        threshold: number;
        schedule: string;
        cooldownMinutes: number;
        channels: Array<{ channelType: AlertNotificationChannel; isEnabled: boolean }>;
    };
}

export function AlertReview({ formData }: AlertReviewProps) {
    const getSeverityIcon = (severity: AlertSeverity) => {
        switch (severity) {
            case 'critical':
                return <AlertCircle className="h-5 w-5 text-red-500" />;
            case 'warning':
                return <AlertTriangle className="h-5 w-5 text-amber-500" />;
            case 'info':
                return <Info className="h-5 w-5 text-blue-500" />;
        }
    };

    const getChannelIcon = (type: AlertNotificationChannel) => {
        switch (type) {
            case 'email':
                return <Mail className="h-4 w-4" />;
            case 'webhook':
                return <Webhook className="h-4 w-4" />;
            case 'in_app':
                return <Bell className="h-4 w-4" />;
            case 'slack':
                return <MessageSquare className="h-4 w-4" />;
        }
    };

    const getScheduleLabel = (schedule: string) => {
        const scheduleMap: Record<string, string> = {
            '1m': 'Every 1 minute',
            '5m': 'Every 5 minutes',
            '15m': 'Every 15 minutes',
            '30m': 'Every 30 minutes',
            '1h': 'Every 1 hour',
        };
        return scheduleMap[schedule] || schedule;
    };

    return (
        <div className="space-y-4">
            <Card>
                <CardHeader>
                    <CardTitle className="text-base flex items-center gap-2">
                        {getSeverityIcon(formData.severity)}
                        {formData.name}
                    </CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                    {formData.description && (
                        <p className="text-sm text-gray-600">{formData.description}</p>
                    )}

                    <div className="grid grid-cols-2 gap-4 text-sm">
                        <div>
                            <span className="text-gray-500">Severity:</span>{' '}
                            <Badge variant="outline" className="capitalize">
                                {formData.severity}
                            </Badge>
                        </div>
                        <div>
                            <span className="text-gray-500">Query:</span>{' '}
                            {formData.queryName || formData.queryId}
                        </div>
                    </div>
                </CardContent>
            </Card>

            <Card>
                <CardHeader>
                    <CardTitle className="text-base">Condition</CardTitle>
                </CardHeader>
                <CardContent>
                    <div className="bg-gray-50 rounded-lg p-4 text-center">
                        <code className="text-lg font-mono">
                            {formData.column || '...'} {formData.operator} {formData.threshold}
                        </code>
                    </div>
                </CardContent>
            </Card>

            <Card>
                <CardHeader>
                    <CardTitle className="text-base flex items-center gap-2">
                        <Clock className="h-4 w-4" />
                        Schedule
                    </CardTitle>
                </CardHeader>
                <CardContent className="space-y-2">
                    <div className="text-sm">
                        <span className="text-gray-500">Check every:</span>{' '}
                        {getScheduleLabel(formData.schedule)}
                    </div>
                    <div className="text-sm">
                        <span className="text-gray-500">Cooldown:</span>{' '}
                        {formData.cooldownMinutes} minutes
                    </div>
                </CardContent>
            </Card>

            <Card>
                <CardHeader>
                    <CardTitle className="text-base">Notification Channels</CardTitle>
                </CardHeader>
                <CardContent>
                    {formData.channels.length === 0 ? (
                        <p className="text-sm text-gray-500">No channels configured</p>
                    ) : (
                        <div className="flex flex-wrap gap-2">
                            {formData.channels.map((channel, index) => (
                                <Badge
                                    key={index}
                                    variant={channel.isEnabled ? 'default' : 'secondary'}
                                    className="flex items-center gap-1"
                                >
                                    {getChannelIcon(channel.channelType)}
                                    <span className="capitalize">{channel.channelType}</span>
                                </Badge>
                            ))}
                        </div>
                    )}
                </CardContent>
            </Card>
        </div>
    );
}
