'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { formatDistanceToNow } from 'date-fns';
import {
    AlertCircle,
    AlertTriangle,
    Info,
    CheckCircle2,
    Clock,
    User,
} from 'lucide-react';
import type { TriggeredAlert, AlertSeverity } from '@/types/alerts';

interface TriggeredAlertsProps {
    alerts: TriggeredAlert[];
    onAcknowledge?: (alertId: string) => void;
    onAcknowledgeAll?: () => void;
    loading?: boolean;
}

export function TriggeredAlerts({
    alerts,
    onAcknowledge,
    onAcknowledgeAll,
    loading,
}: TriggeredAlertsProps) {
    // Group by severity
    const groupedBySeverity = alerts.reduce((acc, alert) => {
        const severity = alert.alert.severity;
        if (!acc[severity]) {
            acc[severity] = [];
        }
        acc[severity].push(alert);
        return acc;
    }, {} as Record<AlertSeverity, TriggeredAlert[]>);

    const getSeverityIcon = (severity: AlertSeverity) => {
        switch (severity) {
            case 'critical':
                return <AlertCircle className="h-5 w-5 text-red-500" />;
            case 'warning':
                return <AlertTriangle className="h-5 w-5 text-amber-500" />;
            case 'info':
                return <Info className="h-5 w-5 text-blue-500" />;
            default:
                return <AlertCircle className="h-5 w-5" />;
        }
    };

    const getSeverityOrder = (severity: AlertSeverity): number => {
        const order: Record<AlertSeverity, number> = {
            critical: 0,
            warning: 1,
            info: 2,
        };
        return order[severity] ?? 3;
    };

    if (loading) {
        return (
            <Card>
                <CardHeader>
                    <CardTitle className="text-lg">Triggered Alerts</CardTitle>
                </CardHeader>
                <CardContent>
                    <div className="space-y-2">
                        {[1, 2, 3].map((i) => (
                            <div key={i} className="h-16 bg-gray-100 rounded animate-pulse" />
                        ))}
                    </div>
                </CardContent>
            </Card>
        );
    }

    if (alerts.length === 0) {
        return (
            <Card>
                <CardHeader>
                    <CardTitle className="text-lg">Triggered Alerts</CardTitle>
                </CardHeader>
                <CardContent>
                    <div className="text-center py-8">
                        <CheckCircle2 className="h-12 w-12 text-green-500 mx-auto mb-3" />
                        <p className="text-gray-500">No triggered alerts</p>
                        <p className="text-sm text-gray-400 mt-1">
                            All alerts are currently OK
                        </p>
                    </div>
                </CardContent>
            </Card>
        );
    }

    return (
        <Card>
            <CardHeader className="flex flex-row items-center justify-between">
                <div className="flex items-center gap-2">
                    <AlertCircle className="h-5 w-5 text-red-500" />
                    <CardTitle className="text-lg">
                        Triggered Alerts ({alerts.length})
                    </CardTitle>
                </div>
                {onAcknowledgeAll && (
                    <Button variant="outline" size="sm" onClick={onAcknowledgeAll}>
                        <CheckCircle2 className="h-4 w-4 mr-2" />
                        Acknowledge All
                    </Button>
                )}
            </CardHeader>
            <CardContent>
                <ScrollArea className="h-[400px]">
                    <div className="space-y-4">
                        {(['critical', 'warning', 'info'] as AlertSeverity[]).map((severity) => {
                            const severityAlerts = groupedBySeverity[severity];
                            if (!severityAlerts || severityAlerts.length === 0) return null;

                            return (
                                <div key={severity} className="space-y-2">
                                    <h4 className="text-sm font-medium text-gray-500 flex items-center gap-2">
                                        {getSeverityIcon(severity)}
                                        {severity.charAt(0).toUpperCase() + severity.slice(1)} ({severityAlerts.length})
                                    </h4>
                                    <div className="space-y-2">
                                        {severityAlerts.map((triggeredAlert) => (
                                            <div
                                                key={triggeredAlert.alert.id}
                                                className={`p-3 rounded-lg border ${
                                                    triggeredAlert.acknowledged
                                                        ? 'bg-gray-50 border-gray-200'
                                                        : 'bg-red-50 border-red-200'
                                                }`}
                                            >
                                                <div className="flex items-start justify-between gap-2">
                                                    <div className="flex-1 min-w-0">
                                                        <div className="flex items-center gap-2">
                                                            <span className="font-medium truncate">
                                                                {triggeredAlert.alert.name}
                                                            </span>
                                                            {triggeredAlert.acknowledged && (
                                                                <Badge variant="outline" className="text-xs">
                                                                    <CheckCircle2 className="h-3 w-3 mr-1" />
                                                                    Acknowledged
                                                                </Badge>
                                                            )}
                                                        </div>
                                                        <div className="flex items-center gap-3 mt-1 text-sm text-gray-600">
                                                            <span className="flex items-center gap-1">
                                                                <Clock className="h-3 w-3" />
                                                                {formatDistanceToNow(
                                                                    new Date(triggeredAlert.triggeredAt),
                                                                    { addSuffix: true }
                                                                )}
                                                            </span>
                                                            {triggeredAlert.acknowledgedBy && (
                                                                <span className="flex items-center gap-1">
                                                                    <User className="h-3 w-3" />
                                                                    By {triggeredAlert.acknowledgedBy}
                                                                </span>
                                                            )}
                                                        </div>
                                                        <div className="mt-2 text-sm">
                                                            <span className="text-gray-500">Current value: </span>
                                                            <span className={`font-mono font-bold ${
                                                                triggeredAlert.acknowledged ? 'text-gray-600' : 'text-red-600'
                                                            }`}>
                                                                {triggeredAlert.currentValue.toFixed(2)}
                                                            </span>
                                                            <span className="text-gray-400 ml-2">
                                                                (threshold: {triggeredAlert.alert.threshold})
                                                            </span>
                                                        </div>
                                                    </div>
                                                    {!triggeredAlert.acknowledged && onAcknowledge && (
                                                        <Button
                                                            size="sm"
                                                            variant="outline"
                                                            onClick={() => onAcknowledge(triggeredAlert.alert.id)}
                                                        >
                                                            Acknowledge
                                                        </Button>
                                                    )}
                                                </div>
                                            </div>
                                        ))}
                                    </div>
                                </div>
                            );
                        })}
                    </div>
                </ScrollArea>
            </CardContent>
        </Card>
    );
}
