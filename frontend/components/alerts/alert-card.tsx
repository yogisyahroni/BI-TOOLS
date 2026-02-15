'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { formatDistanceToNow } from 'date-fns';
import {
    AlertCircle,
    AlertTriangle,
    Info,
    Bell,
    CheckCircle2,
    VolumeX,
} from 'lucide-react';
import type { Alert, AlertSeverity, AlertState } from '@/types/alerts';

interface AlertCardProps {
    alert: Alert;
    onAcknowledge?: () => void;
    onMute?: () => void;
    onUnmute?: () => void;
}

export function AlertCard({ alert, onAcknowledge, onMute, onUnmute }: AlertCardProps) {
    const getSeverityIcon = (severity: AlertSeverity) => {
        switch (severity) {
            case 'critical':
                return <AlertCircle className="h-5 w-5 text-red-500" />;
            case 'warning':
                return <AlertTriangle className="h-5 w-5 text-amber-500" />;
            case 'info':
                return <Info className="h-5 w-5 text-blue-500" />;
            default:
                return <Bell className="h-5 w-5" />;
        }
    };

    const getSeverityColor = (severity: AlertSeverity) => {
        switch (severity) {
            case 'critical':
                return 'bg-red-50 border-red-200';
            case 'warning':
                return 'bg-amber-50 border-amber-200';
            case 'info':
                return 'bg-blue-50 border-blue-200';
            default:
                return 'bg-gray-50 border-gray-200';
        }
    };

    const getStateIcon = (state: AlertState) => {
        switch (state) {
            case 'ok':
                return <CheckCircle2 className="h-4 w-4 text-green-500" />;
            case 'triggered':
                return <AlertCircle className="h-4 w-4 text-red-500" />;
            case 'muted':
                return <VolumeX className="h-4 w-4 text-gray-500" />;
            default:
                return <Bell className="h-4 w-4" />;
        }
    };

    return (
        <Card className={`border-2 ${getSeverityColor(alert.severity)}`}>
            <CardHeader className="pb-2">
                <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                        {getSeverityIcon(alert.severity)}
                        <CardTitle className="text-base">{alert.name}</CardTitle>
                    </div>
                    <Badge variant={alert.state === 'triggered' ? 'destructive' : 'secondary'}>
                        <span className="flex items-center gap-1">
                            {getStateIcon(alert.state)}
                            {alert.state}
                        </span>
                    </Badge>
                </div>
            </CardHeader>
            <CardContent className="pt-0">
                {alert.description && (
                    <p className="text-sm text-gray-600 mb-3">{alert.description}</p>
                )}

                {/* Current Value Display */}
                {alert.state === 'triggered' && alert.lastValue !== undefined && (
                    <div className="bg-white rounded-lg p-3 mb-3 border">
                        <div className="flex items-center justify-between">
                            <span className="text-sm text-gray-500">Current Value</span>
                            <span className="text-2xl font-bold text-red-600">
                                {alert.lastValue.toFixed(2)}
                            </span>
                        </div>
                        <div className="text-xs text-gray-400 mt-1">
                            Threshold: {alert.column} {alert.operator} {alert.threshold}
                        </div>
                    </div>
                )}

                {/* Alert Info */}
                <div className="flex items-center justify-between text-sm">
                    <div className="text-gray-500 space-y-1">
                        <div>
                            Last checked: {alert.lastRunAt
                                ? formatDistanceToNow(new Date(alert.lastRunAt), { addSuffix: true })
                                : 'Never'}
                        </div>
                        {alert.nextRunAt && (
                            <div className="text-xs">
                                Next check: {formatDistanceToNow(new Date(alert.nextRunAt), { addSuffix: true })}
                            </div>
                        )}
                    </div>

                    {/* Action Buttons */}
                    <div className="flex gap-2">
                        {alert.state === 'triggered' && onAcknowledge && (
                            <Button size="sm" variant="outline" onClick={onAcknowledge}>
                                <CheckCircle2 className="h-4 w-4 mr-1" />
                                Acknowledge
                            </Button>
                        )}
                        {alert.isMuted ? (
                            onUnmute && (
                                <Button size="sm" variant="outline" onClick={onUnmute}>
                                    Unmute
                                </Button>
                            )
                        ) : (
                            onMute && (
                                <Button size="sm" variant="outline" onClick={onMute}>
                                    <VolumeX className="h-4 w-4 mr-1" />
                                    Mute
                                </Button>
                            )
                        )}
                    </div>
                </div>
            </CardContent>
        </Card>
    );
}
