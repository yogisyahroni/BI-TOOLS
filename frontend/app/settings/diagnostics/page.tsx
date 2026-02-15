'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { RefreshCcw, CheckCircle2, XCircle } from 'lucide-react';
import { useState, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useSession } from 'next-auth/react';

interface HealthStatus {
    database: boolean;
    api: boolean;
    auth: boolean;
    version: string;
}

export default function DiagnosticsPage() {
    const { data: session } = useSession();
    const [lastChecked, setLastChecked] = useState<Date | null>(null);

    // Simulated health check (replace with real API call)
    const checkHealth = async (): Promise<HealthStatus> => {
        // Simulate API latency
        await new Promise((resolve) => setTimeout(resolve, 1000));

        // In a real app, this would fetch /api/health
        return {
            database: true,
            api: true,
            auth: !!session,
            version: '1.0.0-beta',
        };
    };

    const { data: health, refetch, isFetching } = useQuery({
        queryKey: ['system-health'],
        queryFn: checkHealth,
    });

    useEffect(() => {
        if (health) setLastChecked(new Date());
    }, [health]);

    const StatusIcon = ({ status }: { status: boolean }) => (
        status ? <CheckCircle2 className="h-5 w-5 text-green-500" /> : <XCircle className="h-5 w-5 text-red-500" />
    );

    return (
        <div className="space-y-6">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-3xl font-bold">System Diagnostics</h1>
                    <p className="text-muted-foreground">Monitor the health of InsightEngine services.</p>
                </div>
                <Button onClick={() => refetch()} disabled={isFetching}>
                    <RefreshCcw className={`mr-2 h-4 w-4 ${isFetching ? 'animate-spin' : ''}`} />
                    Refresh Status
                </Button>
            </div>

            <div className="grid gap-6 md:grid-cols-3">
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <StatusIcon status={health?.database ?? false} />
                            Database Connector
                        </CardTitle>
                        <CardDescription>Connection to primary data store.</CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{health?.database ? 'Operational' : 'Offline'}</div>
                        <p className="text-xs text-muted-foreground">Latency: 12ms</p>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <StatusIcon status={health?.api ?? false} />
                            API Gateway
                        </CardTitle>
                        <CardDescription>Backend API availability.</CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{health?.api ? 'Operational' : 'Offline'}</div>
                        <p className="text-xs text-muted-foreground">Uptime: 99.9%</p>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <StatusIcon status={health?.auth ?? false} />
                            Authentication
                        </CardTitle>
                        <CardDescription>Identity provider status.</CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{health?.auth ? 'Active' : 'Inactive'}</div>
                        <p className="text-xs text-muted-foreground">Session: {session?.user?.email || 'None'}</p>
                    </CardContent>
                </Card>
            </div>

            <Card>
                <CardHeader>
                    <CardTitle>System Information</CardTitle>
                </CardHeader>
                <CardContent className="space-y-2">
                    <div className="flex justify-between border-b pb-2">
                        <span className="text-muted-foreground">Version</span>
                        <span>{health?.version || 'Unknown'}</span>
                    </div>
                    <div className="flex justify-between border-b pb-2">
                        <span className="text-muted-foreground">Environment</span>
                        <span>Production</span>
                    </div>
                    <div className="flex justify-between pt-2">
                        <span className="text-muted-foreground">Last Checked</span>
                        <span>{lastChecked?.toLocaleTimeString() || 'Never'}</span>
                    </div>
                </CardContent>
            </Card>
        </div>
    );
}
