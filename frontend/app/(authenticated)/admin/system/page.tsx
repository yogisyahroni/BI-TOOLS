'use client';

export const dynamic = 'force-dynamic';

import { useEffect, useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { systemAdminApi } from '@/lib/api/admin';
import type { SystemHealth, SystemMetrics, DatabaseConnectionInfo } from '@/types/admin';
import { Activity, Database, Cpu, HardDrive, TrendingUp, AlertCircle, CheckCircle, XCircle } from 'lucide-react';

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
        } catch (err: any) {
            setError(err.message || 'Failed to load system health data');
            console.error('Error loading system health:', err);
        } finally {
            setLoading(false);
        }
    };

    const getStatusBadge = (status: string) => {
        const variants: Record<string, { variant: any; icon: any }> = {
            healthy: { variant: 'default', icon: CheckCircle },
            up: { variant: 'default', icon: CheckCircle },
            degraded: { variant: 'warning', icon: AlertCircle },
            unhealthy: { variant: 'destructive', icon: XCircle },
            down: { variant: 'destructive', icon: XCircle },
        };

        const config = variants[status] || variants.degraded;
        const Icon = config.icon;

        return (
            <Badge variant={config.variant as any} className="gap-1">
                <Icon className="h-3 w-3" />
                {status.toUpperCase()}
            </Badge>
        );
    };

    const formatBytes = (bytes: number) => {
        if (bytes === 0) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return `${(bytes / Math.pow(k, i)).toFixed(2)} ${sizes[i]}`;
    };

    const formatUptime = (seconds: number) => {
        const days = Math.floor(seconds / 86400);
        const hours = Math.floor((seconds % 86400) / 3600);
        const minutes = Math.floor((seconds % 3600) / 60);
        return `${days}d ${hours}h ${minutes}m`;
    };

    if (loading && !health) {
        return (
            <div className="container mx-auto p-6 space-y-6">
                <Skeleton className="h-12 w-64" />
                <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
                    {[1, 2, 3, 4].map((i) => (
                        <Skeleton key={i} className="h-32" />
                    ))}
                </div>
            </div>
        );
    }

    if (error) {
        return (
            <div className="container mx-auto p-6">
                <Card className="border-destructive">
                    <CardHeader>
                        <CardTitle className="text-destructive">Error</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <p>{error}</p>
                    </CardContent>
                </Card>
            </div>
        );
    }

    return (
        <div className="container mx-auto p-6 space-y-6">
            {/* Header */}
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-3xl font-bold">System Health</h1>
                    <p className="text-muted-foreground">
                        Monitor system performance and health metrics
                    </p>
                </div>
                {health && getStatusBadge(health.status)}
            </div>

            {/* Overall Status */}
            {health && (
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <Activity className="h-5 w-5" />
                            System Overview
                        </CardTitle>
                        <CardDescription>
                            Last updated: {new Date(health.timestamp).toLocaleString()}
                        </CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        <div className="grid gap-4 md:grid-cols-2">
                            <div>
                                <p className="text-sm text-muted-foreground">Version</p>
                                <p className="text-2xl font-bold">{health.version}</p>
                            </div>
                            <div>
                                <p className="text-sm text-muted-foreground">Uptime</p>
                                <p className="text-2xl font-bold">{formatUptime(health.uptime)}</p>
                            </div>
                        </div>
                    </CardContent>
                </Card>
            )}

            {/* Component Health */}
            {health && (
                <div className="grid gap-4 md:grid-cols-3">
                    {Object.entries(health.components).map(([name, component]) => (
                        <Card key={name}>
                            <CardHeader className="pb-3">
                                <div className="flex items-center justify-between">
                                    <CardTitle className="text-sm font-medium capitalize">
                                        {name}
                                    </CardTitle>
                                    {getStatusBadge(component.status)}
                                </div>
                            </CardHeader>
                            <CardContent>
                                {component.message && (
                                    <p className="text-sm text-muted-foreground mb-2">
                                        {component.message}
                                    </p>
                                )}
                                {component.details && (
                                    <div className="text-xs space-y-1">
                                        {Object.entries(component.details).map(([key, value]) => (
                                            <div key={key} className="flex justify-between">
                                                <span className="text-muted-foreground">{key}:</span>
                                                <span className="font-mono">{String(value)}</span>
                                            </div>
                                        ))}
                                    </div>
                                )}
                            </CardContent>
                        </Card>
                    ))}
                </div>
            )}

            {/* System Metrics */}
            {metrics && (
                <>
                    {/* Memory Metrics */}
                    <Card>
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2">
                                <Cpu className="h-5 w-5" />
                                Memory Usage
                            </CardTitle>
                        </CardHeader>
                        <CardContent>
                            <div className="space-y-4">
                                <div className="space-y-2">
                                    <div className="flex justify-between text-sm">
                                        <span className="text-muted-foreground">Usage</span>
                                        <span className="font-mono">
                                            {metrics.memory.usagePercent.toFixed(2)}%
                                        </span>
                                    </div>
                                    <div className="h-2 bg-secondary rounded-full overflow-hidden">
                                        <div
                                            className="h-full bg-primary transition-all"
                                            style={{ width: `${metrics.memory.usagePercent}%` }}
                                        />
                                    </div>
                                </div>
                                <div className="grid gap-4 md:grid-cols-4 text-sm">
                                    <div>
                                        <p className="text-muted-foreground">Allocated</p>
                                        <p className="font-mono">{formatBytes(metrics.memory.alloc)}</p>
                                    </div>
                                    <div>
                                        <p className="text-muted-foreground">System</p>
                                        <p className="font-mono">{formatBytes(metrics.memory.sys)}</p>
                                    </div>
                                    <div>
                                        <p className="text-muted-foreground">Total Allocated</p>
                                        <p className="font-mono">{formatBytes(metrics.memory.totalAlloc)}</p>
                                    </div>
                                    <div>
                                        <p className="text-muted-foreground">GC Runs</p>
                                        <p className="font-mono">{metrics.memory.numGC}</p>
                                    </div>
                                </div>
                                <div>
                                    <p className="text-muted-foreground text-sm">Goroutines</p>
                                    <p className="text-2xl font-bold">{metrics.goroutines}</p>
                                </div>
                            </div>
                        </CardContent>
                    </Card>

                    {/* Database Metrics */}
                    <Card>
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2">
                                <Database className="h-5 w-5" />
                                Database Connections
                            </CardTitle>
                        </CardHeader>
                        <CardContent>
                            <div className="grid gap-4 md:grid-cols-4 text-sm">
                                <div>
                                    <p className="text-muted-foreground">Open</p>
                                    <p className="text-2xl font-bold">{metrics.database.connectionCount}</p>
                                </div>
                                <div>
                                    <p className="text-muted-foreground">Active</p>
                                    <p className="text-2xl font-bold">{metrics.database.activeConnections}</p>
                                </div>
                                <div>
                                    <p className="text-muted-foreground">Idle</p>
                                    <p className="text-2xl font-bold">{metrics.database.idleConnections}</p>
                                </div>
                                <div>
                                    <p className="text-muted-foreground">Max</p>
                                    <p className="text-2xl font-bold">{metrics.database.maxConnections}</p>
                                </div>
                            </div>
                        </CardContent>
                    </Card>
                </>
            )}

            {/* Database Connections */}
            {connections.length > 0 && (
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <HardDrive className="h-5 w-5" />
                            Configured Connections
                        </CardTitle>
                        <CardDescription>
                            Status of configured database connections
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="space-y-2">
                            {connections.map((conn) => (
                                <div
                                    key={conn.id}
                                    className="flex items-center justify-between p-3 border rounded-lg"
                                >
                                    <div className="flex items-center gap-3">
                                        {getStatusBadge(conn.status)}
                                        <div>
                                            <p className="font-medium">{conn.name}</p>
                                            <p className="text-sm text-muted-foreground">{conn.type}</p>
                                        </div>
                                    </div>
                                    <div className="text-right">
                                        <p className="text-sm text-muted-foreground">Response Time</p>
                                        <p className="font-mono text-sm">{conn.responseTimeMs}ms</p>
                                    </div>
                                </div>
                            ))}
                        </div>
                    </CardContent>
                </Card>
            )}
        </div>
    );
}
