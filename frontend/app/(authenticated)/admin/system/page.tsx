'use client';

import { useEffect, useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { systemAdminApi } from '@/lib/api/admin';
import type { SystemHealth, SystemMetrics, DatabaseConnectionInfo } from '@/types/admin';
import { AlertCircle, CheckCircle, XCircle } from 'lucide-react';

export default function SystemHealthPage() {
    const [health, setHealth] = useState<SystemHealth | null>(null);
    const [metrics, setMetrics] = useState<SystemMetrics | null>(null);
    const [connections, setConnections] = useState<DatabaseConnectionInfo[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        loadData();
        // Refresh every 30 seconds
        const interval = setInterval(loadData, 30000);
        return () => clearInterval(interval);
    }, []);

    const loadData = async () => {
        try {
            setLoading(true);
            setError(null);

            const [healthData, metricsData, connectionsData] = await Promise.all([
                systemAdminApi.getHealth(),
                systemAdminApi.getMetrics(),
                systemAdminApi.getDatabaseConnections(),
            ]);

            setHealth(healthData);
            setMetrics(metricsData);
            setConnections(connectionsData.connections);
        } catch (err: unknown) {
            const errorMessage = err instanceof Error ? err.message : 'Failed to load system health data';
            setError(errorMessage);
            console.error('Error loading system health:', err);
        } finally {
            setLoading(false);
        }
    };

    const getStatusBadge = (status: string) => {
        const variants: Record<string, { variant: "default" | "secondary" | "destructive" | "outline" | "warning"; icon: React.ElementType }> = {
            healthy: { variant: 'default', icon: CheckCircle },
            up: { variant: 'default', icon: CheckCircle },
            degraded: { variant: 'warning', icon: AlertCircle },
            unhealthy: { variant: 'destructive', icon: XCircle },
            down: { variant: 'destructive', icon: XCircle },
        };

        const config = variants[status] || variants.degraded;
        const Icon = config.icon;

        return (
            <Badge variant={config.variant} className="gap-1">
                <Icon className="h-3 w-3" />
                {status.toUpperCase()}
            </Badge>
        );
    };

    if (loading && !health) {
        return <div className="p-6">Loading system health...</div>;
    }

    if (error) {
        return (
            <div className="p-6">
                <div className="bg-destructive/15 text-destructive p-4 rounded-md">
                    {error}
                </div>
            </div>
        );
    }

    return (
        <div className="container mx-auto p-6 space-y-6">
            <h1 className="text-3xl font-bold">System Health</h1>

            <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
                <Card>
                    <CardHeader>
                        <CardTitle className="text-sm font-medium">System Status</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold mb-2">
                            {health?.status ? getStatusBadge(health.status) : 'Unknown'}
                        </div>
                        <div className="text-xs text-muted-foreground">
                            Last active: {health?.lastActive ? new Date(health.lastActive).toLocaleString() : 'Never'}
                        </div>
                    </CardContent>
                </Card>

                {metrics && (
                    <>
                        <Card>
                            <CardHeader>
                                <CardTitle className="text-sm font-medium">Memory Usage</CardTitle>
                            </CardHeader>
                            <CardContent>
                                <div className="text-2xl font-bold">
                                    {Math.round(metrics.memory.heapUsed / 1024 / 1024)} MB
                                </div>
                                <div className="text-xs text-muted-foreground">
                                    of {Math.round(metrics.memory.heapTotal / 1024 / 1024)} MB Total
                                </div>
                            </CardContent>
                        </Card>

                        <Card>
                            <CardHeader>
                                <CardTitle className="text-sm font-medium">Uptime</CardTitle>
                            </CardHeader>
                            <CardContent>
                                <div className="text-2xl font-bold">
                                    {Math.floor(metrics.uptime / 3600)}h {Math.floor((metrics.uptime % 3600) / 60)}m
                                </div>
                                <div className="text-xs text-muted-foreground">
                                    Since restart
                                </div>
                            </CardContent>
                        </Card>
                    </>
                )}
            </div>

            <Card>
                <CardHeader>
                    <CardTitle>Database Connections</CardTitle>
                </CardHeader>
                <CardContent>
                    <div className="space-y-4">
                        {connections.map((conn, i) => (
                            <div key={i} className="flex items-center justify-between p-4 border rounded-lg">
                                <div>
                                    <div className="font-medium">{conn.name}</div>
                                    <div className="text-sm text-muted-foreground">{conn.type}</div>
                                </div>
                                {getStatusBadge(conn.status)}
                            </div>
                        ))}
                    </div>
                </CardContent>
            </Card>
        </div>
    );
}
