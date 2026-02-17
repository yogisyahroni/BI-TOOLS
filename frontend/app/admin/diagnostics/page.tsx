"use client";

import { useState, useEffect } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { RefreshCw, CheckCircle, XCircle, AlertTriangle } from "lucide-react";
import { ContextualHelp } from "@/components/help/contextual-help";

interface HealthStatus {
    service: string;
    status: "operational" | "degraded" | "down";
    latency: string;
    lastChecked: string;
    details?: string;
}

export default function DiagnosticsPage() {
    const [statuses, setStatuses] = useState<HealthStatus[]>([]);
    const [loading, setLoading] = useState(false);

    // Mock health check function
    const checkHealth = async () => {
        setLoading(true);
        // Simulate API call delay
        await new Promise((resolve) => setTimeout(resolve, 1500));

        const mockStatuses: HealthStatus[] = [
            {
                service: "Backend API",
                status: "operational",
                latency: "45ms",
                lastChecked: new Date().toLocaleTimeString(),
            },
            {
                service: "PostgreSQL Database",
                status: "operational",
                latency: "12ms",
                lastChecked: new Date().toLocaleTimeString(),
            },
            {
                service: "Redis Cache",
                status: "operational",
                latency: "5ms",
                lastChecked: new Date().toLocaleTimeString(),
            },
            {
                service: "Email Service",
                status: "degraded",
                latency: "350ms",
                lastChecked: new Date().toLocaleTimeString(),
                details: "High latency detected on SMTP connection",
            },
            {
                service: "AI Inference Engine",
                status: "down",
                latency: "-",
                lastChecked: new Date().toLocaleTimeString(),
                details: "Service unreachable via gRPC",
            },
        ];

        setStatuses(mockStatuses);
        setLoading(false);
    };

    useEffect(() => {
        checkHealth();
    }, []);

    const getStatusBadge = (status: string) => {
        switch (status) {
            case "operational":
                return <Badge className="bg-green-500">Operational</Badge>;
            case "degraded":
                return <Badge className="bg-yellow-500">Degraded</Badge>;
            case "down":
                return <Badge variant="destructive">Down</Badge>;
            default:
                return <Badge variant="secondary">Unknown</Badge>;
        }
    };

    const getStatusIcon = (status: string) => {
        switch (status) {
            case "operational":
                return <CheckCircle className="h-5 w-5 text-green-500" />;
            case "degraded":
                return <AlertTriangle className="h-5 w-5 text-yellow-500" />;
            case "down":
                return <XCircle className="h-5 w-5 text-red-500" />;
            default:
                return <RefreshCw className="h-5 w-5 animate-spin" />;
        }
    };

    return (
        <div className="container mx-auto py-10 space-y-8">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight">System Diagnostics</h1>
                    <p className="text-muted-foreground mt-2 flex items-center gap-2">
                        Monitor the health and performance of all system components.
                        <ContextualHelp content="Real-time health checks performed from the backend server." />
                    </p>
                </div>
                <Button onClick={checkHealth} disabled={loading}>
                    {loading ? (
                        <RefreshCw className="mr-2 h-4 w-4 animate-spin" />
                    ) : (
                        <RefreshCw className="mr-2 h-4 w-4" />
                    )}
                    Run Diagnostics
                </Button>
            </div>

            <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
                {statuses.map((status) => (
                    <Card key={status.service} className="shadow-sm">
                        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                            <CardTitle className="text-sm font-medium">
                                {status.service}
                            </CardTitle>
                            {getStatusIcon(status.status)}
                        </CardHeader>
                        <CardContent>
                            <div className="text-2xl font-bold flex items-center gap-2">
                                {getStatusBadge(status.status)}
                            </div>
                            <p className="text-xs text-muted-foreground mt-2">
                                Latency: {status.latency}
                            </p>
                            <p className="text-xs text-muted-foreground">
                                Checked: {status.lastChecked}
                            </p>
                            {status.details && (
                                <div className="mt-4 p-2 bg-muted rounded-md text-xs text-destructive">
                                    Error: {status.details}
                                </div>
                            )}
                        </CardContent>
                    </Card>
                ))}
            </div>
        </div>
    );
}
