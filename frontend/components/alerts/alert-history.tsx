'use client';

import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Button } from '@/components/ui/button';
import { formatDistanceToNow, format } from 'date-fns';
import {
    CheckCircle2,
    AlertCircle,
    XCircle,
    Clock,
    ChevronDown,
    ChevronUp,
} from 'lucide-react';
import { useState } from 'react';
import type { AlertHistory, AlertHistoryStatus } from '@/types/alerts';

interface AlertHistoryProps {
    history: AlertHistory[];
    loading?: boolean;
    hasMore?: boolean;
    onLoadMore?: () => void;
}

export function AlertHistory({ history, loading, hasMore, onLoadMore }: AlertHistoryProps) {
    const [expandedItems, setExpandedItems] = useState<Set<string>>(new Set());

    const toggleExpanded = (id: string) => {
        const newExpanded = new Set(expandedItems);
        if (newExpanded.has(id)) {
            newExpanded.delete(id);
        } else {
            newExpanded.add(id);
        }
        setExpandedItems(newExpanded);
    };

    const getStatusIcon = (status: AlertHistoryStatus) => {
        switch (status) {
            case 'ok':
                return <CheckCircle2 className="h-4 w-4 text-green-500" />;
            case 'triggered':
                return <AlertCircle className="h-4 w-4 text-red-500" />;
            case 'error':
                return <XCircle className="h-4 w-4 text-purple-500" />;
        }
    };

    const getStatusColor = (status: AlertHistoryStatus) => {
        switch (status) {
            case 'ok':
                return 'bg-green-100 text-green-800 border-green-200';
            case 'triggered':
                return 'bg-red-100 text-red-800 border-red-200';
            case 'error':
                return 'bg-purple-100 text-purple-800 border-purple-200';
        }
    };

    if (loading) {
        return (
            <div className="space-y-2">
                {[1, 2, 3].map((i) => (
                    <Card key={i}>
                        <CardContent className="p-4">
                            <div className="h-4 bg-gray-200 rounded w-1/4 mb-2"></div>
                            <div className="h-3 bg-gray-200 rounded w-3/4"></div>
                        </CardContent>
                    </Card>
                ))}
            </div>
        );
    }

    if (history.length === 0) {
        return (
            <div className="text-center py-8 text-gray-500">
                <Clock className="h-12 w-12 mx-auto mb-3 text-gray-300" />
                <p>No history available</p>
                <p className="text-sm">Alert check history will appear here</p>
            </div>
        );
    }

    return (
        <div className="space-y-2">
            <ScrollArea className="h-[400px]">
                <div className="space-y-2">
                    {history.map((item) => (
                        <Card key={item.id}>
                            <CardContent className="p-4">
                                <div className="flex items-start justify-between gap-4">
                                    <div className="flex-1 min-w-0">
                                        <div className="flex items-center gap-2">
                                            {getStatusIcon(item.status)}
                                            <Badge
                                                variant="outline"
                                                className={getStatusColor(item.status)}
                                            >
                                                {item.status}
                                            </Badge>
                                            <span className="text-sm text-gray-500">
                                                {formatDistanceToNow(new Date(item.checkedAt), {
                                                    addSuffix: true,
                                                })}
                                            </span>
                                        </div>

                                        {item.message && (
                                            <p className="text-sm mt-2">{item.message}</p>
                                        )}

                                        {item.value !== undefined && (
                                            <div className="flex items-center gap-4 mt-2 text-sm">
                                                <span>
                                                    Value:{' '}
                                                    <strong
                                                        className={
                                                            item.status === 'triggered'
                                                                ? 'text-red-600'
                                                                : 'text-green-600'
                                                        }
                                                    >
                                                        {item.value.toFixed(2)}
                                                    </strong>
                                                </span>
                                                <span className="text-gray-400">
                                                    Threshold: {item.threshold}
                                                </span>
                                            </div>
                                        )}

                                        {/* Expanded details */}
                                        {expandedItems.has(item.id) && (
                                            <div className="mt-3 pt-3 border-t text-sm space-y-2">
                                                <div className="text-gray-500">
                                                    <span className="font-medium">Time:</span>{' '}
                                                    {format(new Date(item.checkedAt), 'PPpp')}
                                                </div>
                                                <div className="text-gray-500">
                                                    <span className="font-medium">Query Duration:</span>{' '}
                                                    {item.queryDuration}ms
                                                </div>
                                                {item.errorMessage && (
                                                    <div className="text-red-600 bg-red-50 p-2 rounded">
                                                        <span className="font-medium">Error:</span>{' '}
                                                        {item.errorMessage}
                                                    </div>
                                                )}
                                                {item.notifications && item.notifications.length > 0 && (
                                                    <div>
                                                        <span className="font-medium text-gray-500">
                                                            Notifications Sent:
                                                        </span>
                                                        <div className="mt-1 space-y-1">
                                                            {item.notifications.map((notif, idx) => (
                                                                <div
                                                                    key={idx}
                                                                    className="flex items-center gap-2 text-xs"
                                                                >
                                                                    <Badge
                                                                        variant={
                                                                            notif.status === 'sent'
                                                                                ? 'default'
                                                                                : 'destructive'
                                                                        }
                                                                    >
                                                                        {notif.channelType}
                                                                    </Badge>
                                                                    <span>{notif.status}</span>
                                                                </div>
                                                            ))}
                                                        </div>
                                                    </div>
                                                )}
                                            </div>
                                        )}
                                    </div>

                                    <Button
                                        variant="ghost"
                                        size="sm"
                                        onClick={() => toggleExpanded(item.id)}
                                    >
                                        {expandedItems.has(item.id) ? (
                                            <ChevronUp className="h-4 w-4" />
                                        ) : (
                                            <ChevronDown className="h-4 w-4" />
                                        )}
                                    </Button>
                                </div>
                            </CardContent>
                        </Card>
                    ))}
                </div>
            </ScrollArea>

            {hasMore && onLoadMore && (
                <div className="text-center pt-4">
                    <Button variant="outline" onClick={onLoadMore} disabled={loading}>
                        Load More
                    </Button>
                </div>
            )}
        </div>
    );
}
