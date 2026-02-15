'use client';

export const dynamic = 'force-dynamic';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { organizationApi, userAdminApi, systemAdminApi } from '@/lib/api/admin';
import { Building2, Users, Activity, TrendingUp, ArrowRight, AlertCircle } from 'lucide-react';
import { Badge } from '@/components/ui/badge';

export default function AdminDashboardPage() {
    const [loading, setLoading] = useState(true);
    const [data, setData] = useState<{
        orgStats: any;
        userStats: any;
        systemHealth: any;
    } | null>(null);

    useEffect(() => {
        loadDashboardData();
    }, []);

    const loadDashboardData = async () => {
        try {
            setLoading(true);
            const [orgStats, userStats, systemHealth] = await Promise.all([
                organizationApi.getStats().catch(() => null),
                userAdminApi.getStats().catch(() => null),
                systemAdminApi.getHealth().catch(() => null),
            ]);

            setData({ orgStats, userStats, systemHealth });
        } catch (error) {
            console.error('Failed to load dashboard data:', error);
        } finally {
            setLoading(false);
        }
    };

    if (loading) {
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

    return (
        <div className="container mx-auto p-6 space-y-6">
            {/* Header */}
            <div>
                <h1 className="text-3xl font-bold">Admin Dashboard</h1>
                <p className="text-muted-foreground">
                    Overview of system status and key metrics
                </p>
            </div>

            {/* System Status Alert */}
            {data?.systemHealth && data.systemHealth.status !== 'healthy' && (
                <Card className="border-destructive">
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2 text-destructive">
                            <AlertCircle className="h-5 w-5" />
                            System Status: {data.systemHealth.status.toUpperCase()}
                        </CardTitle>
                        <CardDescription>
                            Some system components require attention
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <Button asChild variant="destructive">
                            <Link href="/admin/system">
                                View System Health
                                <ArrowRight className="h-4 w-4 ml-2" />
                            </Link>
                        </Button>
                    </CardContent>
                </Card>
            )}

            {/* Quick Stats */}
            <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
                {/* Organizations */}
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between pb-2">
                        <CardTitle className="text-sm font-medium text-muted-foreground">
                            Organizations
                        </CardTitle>
                        <Building2 className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">
                            {data?.orgStats?.totalOrganizations || 0}
                        </div>
                        <p className="text-xs text-muted-foreground mt-1">
                            {data?.orgStats?.activeOrganizations || 0} active
                        </p>
                        <Button asChild variant="link" className="mt-2 p-0 h-auto">
                            <Link href="/admin/organizations">
                                View all <ArrowRight className="h-3 w-3 ml-1" />
                            </Link>
                        </Button>
                    </CardContent>
                </Card>

                {/* Users */}
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between pb-2">
                        <CardTitle className="text-sm font-medium text-muted-foreground">
                            Total Users
                        </CardTitle>
                        <Users className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">
                            {data?.userStats?.totalUsers || 0}
                        </div>
                        <p className="text-xs text-muted-foreground mt-1">
                            {data?.userStats?.activeUsers || 0} active
                        </p>
                        <Button asChild variant="link" className="mt-2 p-0 h-auto">
                            <Link href="/admin/users">
                                Manage users <ArrowRight className="h-3 w-3 ml-1" />
                            </Link>
                        </Button>
                    </CardContent>
                </Card>

                {/* New Users */}
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between pb-2">
                        <CardTitle className="text-sm font-medium text-muted-foreground">
                            New This Month
                        </CardTitle>
                        <TrendingUp className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">
                            {data?.userStats?.newThisMonth || 0}
                        </div>
                        <p className="text-xs text-muted-foreground mt-1">
                            New user registrations
                        </p>
                    </CardContent>
                </Card>

                {/* System Health */}
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between pb-2">
                        <CardTitle className="text-sm font-medium text-muted-foreground">
                            System Status
                        </CardTitle>
                        <Activity className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        {data?.systemHealth ? (
                            <>
                                <Badge
                                    variant={
                                        data.systemHealth.status === 'healthy'
                                            ? 'default'
                                            : data.systemHealth.status === 'degraded'
                                                ? 'warning'
                                                : 'destructive'
                                    }
                                    className="text-sm"
                                >
                                    {data.systemHealth.status.toUpperCase()}
                                </Badge>
                                <Button asChild variant="link" className="mt-2 p-0 h-auto">
                                    <Link href="/admin/system">
                                        View details <ArrowRight className="h-3 w-3 ml-1" />
                                    </Link>
                                </Button>
                            </>
                        ) : (
                            <p className="text-sm text-muted-foreground">Loading...</p>
                        )}
                    </CardContent>
                </Card>
            </div>

            {/* Quick Actions */}
            <div className="grid gap-6 md:grid-cols-3">
                <Card>
                    <CardHeader>
                        <CardTitle className="text-lg flex items-center gap-2">
                            <Building2 className="h-5 w-5" />
                            Organizations
                        </CardTitle>
                        <CardDescription>
                            Manage organizations and workspaces
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="space-y-2">
                            <p className="text-sm text-muted-foreground">
                                Total: {data?.orgStats?.totalOrganizations || 0}
                            </p>
                            <p className="text-sm text-muted-foreground">
                                Members: {data?.orgStats?.totalMembers || 0}
                            </p>
                        </div>
                        <Button asChild className="mt-4 w-full">
                            <Link href="/admin/organizations">
                                Manage Organizations
                            </Link>
                        </Button>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader>
                        <CardTitle className="text-lg flex items-center gap-2">
                            <Users className="h-5 w-5" />
                            User Management
                        </CardTitle>
                        <CardDescription>
                            Manage users, roles, and permissions
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="space-y-2">
                            <p className="text-sm text-muted-foreground">
                                Active: {data?.userStats?.activeUsers || 0}
                            </p>
                            <p className="text-sm text-muted-foreground">
                                Verified: {data?.userStats?.verifiedUsers || 0}
                            </p>
                        </div>
                        <Button asChild className="mt-4 w-full">
                            <Link href="/admin/users">
                                Manage Users
                            </Link>
                        </Button>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader>
                        <CardTitle className="text-lg flex items-center gap-2">
                            <Activity className="h-5 w-5" />
                            System Health
                        </CardTitle>
                        <CardDescription>
                            Monitor system performance and health
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="space-y-2">
                            {data?.systemHealth && (
                                <>
                                    <div className="flex items-center justify-between">
                                        <span className="text-sm text-muted-foreground">Status:</span>
                                        <Badge
                                            variant={
                                                data.systemHealth.status === 'healthy'
                                                    ? 'default'
                                                    : 'destructive'
                                            }
                                        >
                                            {data.systemHealth.status}
                                        </Badge>
                                    </div>
                                    <p className="text-sm text-muted-foreground">
                                        Version: {data.systemHealth.version}
                                    </p>
                                </>
                            )}
                        </div>
                        <Button asChild className="mt-4 w-full">
                            <Link href="/admin/system">
                                View System Health
                            </Link>
                        </Button>
                    </CardContent>
                </Card>
            </div>

            {/* User Statistics */}
            {data?.userStats && (
                <Card>
                    <CardHeader>
                        <CardTitle>User Statistics</CardTitle>
                        <CardDescription>
                            Breakdown of user status and verification
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="grid gap-4 md:grid-cols-5">
                            <div>
                                <p className="text-sm text-muted-foreground">Total Users</p>
                                <p className="text-2xl font-bold">{data.userStats.totalUsers}</p>
                            </div>
                            <div>
                                <p className="text-sm text-muted-foreground">Active</p>
                                <p className="text-2xl font-bold text-green-600">{data.userStats.activeUsers}</p>
                            </div>
                            <div>
                                <p className="text-sm text-muted-foreground">Inactive</p>
                                <p className="text-2xl font-bold text-red-600">{data.userStats.inactiveUsers}</p>
                            </div>
                            <div>
                                <p className="text-sm text-muted-foreground">Pending</p>
                                <p className="text-2xl font-bold text-yellow-600">{data.userStats.pendingUsers}</p>
                            </div>
                            <div>
                                <p className="text-sm text-muted-foreground">Verified</p>
                                <p className="text-2xl font-bold">{data.userStats.verifiedUsers}</p>
                            </div>
                        </div>
                    </CardContent>
                </Card>
            )}
        </div>
    );
}
