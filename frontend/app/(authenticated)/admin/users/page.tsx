'use client';

export const dynamic = 'force-dynamic';

import { useEffect, useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from '@/components/ui/table';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Skeleton } from '@/components/ui/skeleton';
import { useToast } from '@/hooks/use-toast';
import { userAdminApi } from '@/lib/api/admin';
import type { AdminUser, UserStats } from '@/types/admin';
import {
    Search,
    UserCheck,
    UserX,
    Shield,
    Eye,
    MoreVertical,
    CheckCircle,
    XCircle,
    Clock,
    Users,
} from 'lucide-react';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';

export default function UserManagementPage() {
    const { toast } = useToast();
    const [users, setUsers] = useState<AdminUser[]>([]);
    const [stats, setStats] = useState<UserStats | null>(null);
    const [loading, setLoading] = useState(true);
    const [page, setPage] = useState(1);
    const [total, setTotal] = useState(0);
    const [search, setSearch] = useState('');
    const [statusFilter, setStatusFilter] = useState('');
    const [roleFilter, setRoleFilter] = useState('');

    // Dialogs
    const [deactivateDialog, setDeactivateDialog] = useState<{ open: boolean; user: AdminUser | null }>({
        open: false,
        user: null,
    });
    const [roleDialog, setRoleDialog] = useState<{ open: boolean; user: AdminUser | null; role: string }>({
        open: false,
        user: null,
        role: '',
    });
    const [deactivateReason, setDeactivateReason] = useState('');

    useEffect(() => {
        loadUsers();
        loadStats();
    }, [page, search, statusFilter, roleFilter]);

    const loadUsers = async () => {
        try {
            setLoading(true);
            const data = await userAdminApi.list({
                page,
                pageSize: 20,
                search: search || undefined,
                status: statusFilter || undefined,
                role: roleFilter || undefined,
            });
            setUsers(data.data);
            setTotal(data.pagination.total);
        } catch (error: any) {
            toast({
                title: 'Error',
                description: error.message || 'Failed to load users',
                variant: 'destructive',
            });
        } finally {
            setLoading(false);
        }
    };

    const loadStats = async () => {
        try {
            const data = await userAdminApi.getStats();
            setStats(data);
        } catch (error) {
            console.error('Failed to load stats:', error);
        }
    };

    const handleActivate = async (user: AdminUser) => {
        try {
            await userAdminApi.activate(user.id);
            toast({
                title: 'Success',
                description: `User ${user.email} has been activated`,
            });
            loadUsers();
            loadStats();
        } catch (error: any) {
            toast({
                title: 'Error',
                description: error.message || 'Failed to activate user',
                variant: 'destructive',
            });
        }
    };

    const handleDeactivate = async () => {
        if (!deactivateDialog.user) return;

        try {
            await userAdminApi.deactivate(deactivateDialog.user.id, {
                reason: deactivateReason || undefined,
            });
            toast({
                title: 'Success',
                description: `User ${deactivateDialog.user.email} has been deactivated`,
            });
            setDeactivateDialog({ open: false, user: null });
            setDeactivateReason('');
            loadUsers();
            loadStats();
        } catch (error: any) {
            toast({
                title: 'Error',
                description: error.message || 'Failed to deactivate user',
                variant: 'destructive',
            });
        }
    };

    const handleUpdateRole = async () => {
        if (!roleDialog.user || !roleDialog.role) return;

        try {
            await userAdminApi.updateRole(roleDialog.user.id, { role: roleDialog.role });
            toast({
                title: 'Success',
                description: `User role updated to ${roleDialog.role}`,
            });
            setRoleDialog({ open: false, user: null, role: '' });
            loadUsers();
        } catch (error: any) {
            toast({
                title: 'Error',
                description: error.message || 'Failed to update role',
                variant: 'destructive',
            });
        }
    };

    const handleImpersonate = async (user: AdminUser) => {
        try {
            const result = await userAdminApi.impersonate(user.id);

            // Store the impersonation token and redirect
            localStorage.setItem('authToken', result.token);
            toast({
                title: 'Success',
                description: `Now impersonating ${user.email}`,
            });

            // Reload the page to apply the new token
            window.location.href = '/';
        } catch (error: any) {
            toast({
                title: 'Error',
                description: error.message || 'Failed to impersonate user',
                variant: 'destructive',
            });
        }
    };

    const getStatusBadge = (status: string) => {
        const config: Record<string, { variant: any; icon: any }> = {
            active: { variant: 'default', icon: CheckCircle },
            inactive: { variant: 'secondary', icon: XCircle },
            pending: { variant: 'outline', icon: Clock },
        };

        const { variant, icon: Icon } = config[status] || config.pending;

        return (
            <Badge variant={variant} className="gap-1">
                <Icon className="h-3 w-3" />
                {status}
            </Badge>
        );
    };

    return (
        <div className="container mx-auto p-6 space-y-6">
            {/* Header */}
            <div>
                <h1 className="text-3xl font-bold">User Management</h1>
                <p className="text-muted-foreground">
                    Manage users, roles, and permissions
                </p>
            </div>

            {/* Stats */}
            {stats && (
                <div className="grid gap-4 md:grid-cols-4">
                    <Card>
                        <CardHeader className="pb-2">
                            <CardTitle className="text-sm font-medium text-muted-foreground">
                                Total Users
                            </CardTitle>
                        </CardHeader>
                        <CardContent>
                            <div className="flex items-center gap-2">
                                <Users className="h-4 w-4 text-muted-foreground" />
                                <span className="text-2xl font-bold">{stats.totalUsers}</span>
                            </div>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader className="pb-2">
                            <CardTitle className="text-sm font-medium text-muted-foreground">
                                Active Users
                            </CardTitle>
                        </CardHeader>
                        <CardContent>
                            <div className="flex items-center gap-2">
                                <UserCheck className="h-4 w-4 text-green-600" />
                                <span className="text-2xl font-bold">{stats.activeUsers}</span>
                            </div>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader className="pb-2">
                            <CardTitle className="text-sm font-medium text-muted-foreground">
                                Inactive Users
                            </CardTitle>
                        </CardHeader>
                        <CardContent>
                            <div className="flex items-center gap-2">
                                <UserX className="h-4 w-4 text-red-600" />
                                <span className="text-2xl font-bold">{stats.inactiveUsers}</span>
                            </div>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader className="pb-2">
                            <CardTitle className="text-sm font-medium text-muted-foreground">
                                New This Month
                            </CardTitle>
                        </CardHeader>
                        <CardContent>
                            <span className="text-2xl font-bold">{stats.newThisMonth}</span>
                        </CardContent>
                    </Card>
                </div>
            )}

            {/* Filters */}
            <Card>
                <CardContent className="pt-6">
                    <div className="flex flex-col md:flex-row gap-4">
                        <div className="flex-1 relative">
                            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                            <Input
                                placeholder="Search by email, name, or username..."
                                value={search}
                                onChange={(e) => setSearch(e.target.value)}
                                className="pl-10"
                            />
                        </div>

                        <Select value={statusFilter} onValueChange={setStatusFilter}>
                            <SelectTrigger className="w-full md:w-[180px]">
                                <SelectValue placeholder="Filter by status" />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="">All Statuses</SelectItem>
                                <SelectItem value="active">Active</SelectItem>
                                <SelectItem value="inactive">Inactive</SelectItem>
                                <SelectItem value="pending">Pending</SelectItem>
                            </SelectContent>
                        </Select>

                        <Select value={roleFilter} onValueChange={setRoleFilter}>
                            <SelectTrigger className="w-full md:w-[180px]">
                                <SelectValue placeholder="Filter by role" />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="">All Roles</SelectItem>
                                <SelectItem value="admin">Admin</SelectItem>
                                <SelectItem value="user">User</SelectItem>
                            </SelectContent>
                        </Select>
                    </div>
                </CardContent>
            </Card>

            {/* Users Table */}
            <Card>
                <CardHeader>
                    <CardTitle>Users</CardTitle>
                    <CardDescription>
                        {total} user{total !== 1 ? 's' : ''} total
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>User</TableHead>
                                <TableHead>Status</TableHead>
                                <TableHead>Role</TableHead>
                                <TableHead>Verified</TableHead>
                                <TableHead>Created</TableHead>
                                <TableHead className="text-right">Actions</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {loading ? (
                                Array.from({ length: 5 }).map((_, i) => (
                                    <TableRow key={i}>
                                        <TableCell><Skeleton className="h-4 w-[200px]" /></TableCell>
                                        <TableCell><Skeleton className="h-4 w-[80px]" /></TableCell>
                                        <TableCell><Skeleton className="h-4 w-[60px]" /></TableCell>
                                        <TableCell><Skeleton className="h-4 w-[60px]" /></TableCell>
                                        <TableCell><Skeleton className="h-4 w-[100px]" /></TableCell>
                                        <TableCell><Skeleton className="h-4 w-[40px]" /></TableCell>
                                    </TableRow>
                                ))
                            ) : users.length === 0 ? (
                                <TableRow>
                                    <TableCell colSpan={6} className="text-center text-muted-foreground">
                                        No users found
                                    </TableCell>
                                </TableRow>
                            ) : (
                                users.map((user) => (
                                    <TableRow key={user.id}>
                                        <TableCell>
                                            <div>
                                                <p className="font-medium">{user.name || user.username}</p>
                                                <p className="text-sm text-muted-foreground">{user.email}</p>
                                            </div>
                                        </TableCell>
                                        <TableCell>{getStatusBadge(user.status)}</TableCell>
                                        <TableCell>
                                            <Badge variant="outline">{user.role}</Badge>
                                        </TableCell>
                                        <TableCell>
                                            {user.emailVerified ? (
                                                <CheckCircle className="h-4 w-4 text-green-600" />
                                            ) : (
                                                <XCircle className="h-4 w-4 text-muted-foreground" />
                                            )}
                                        </TableCell>
                                        <TableCell className="text-sm text-muted-foreground">
                                            {new Date(user.createdAt).toLocaleDateString()}
                                        </TableCell>
                                        <TableCell className="text-right">
                                            <DropdownMenu>
                                                <DropdownMenuTrigger asChild>
                                                    <Button variant="ghost" size="sm">
                                                        <MoreVertical className="h-4 w-4" />
                                                    </Button>
                                                </DropdownMenuTrigger>
                                                <DropdownMenuContent align="end">
                                                    {user.status === 'active' ? (
                                                        <DropdownMenuItem
                                                            onClick={() => setDeactivateDialog({ open: true, user })}
                                                            className="text-destructive"
                                                        >
                                                            <UserX className="h-4 w-4 mr-2" />
                                                            Deactivate
                                                        </DropdownMenuItem>
                                                    ) : (
                                                        <DropdownMenuItem onClick={() => handleActivate(user)}>
                                                            <UserCheck className="h-4 w-4 mr-2" />
                                                            Activate
                                                        </DropdownMenuItem>
                                                    )}
                                                    <DropdownMenuItem
                                                        onClick={() => setRoleDialog({ open: true, user, role: user.role })}
                                                    >
                                                        <Shield className="h-4 w-4 mr-2" />
                                                        Change Role
                                                    </DropdownMenuItem>
                                                    <DropdownMenuSeparator />
                                                    <DropdownMenuItem onClick={() => handleImpersonate(user)}>
                                                        <Eye className="h-4 w-4 mr-2" />
                                                        Impersonate
                                                    </DropdownMenuItem>
                                                </DropdownMenuContent>
                                            </DropdownMenu>
                                        </TableCell>
                                    </TableRow>
                                ))
                            )}
                        </TableBody>
                    </Table>

                    {/* Pagination */}
                    {total > 20 && (
                        <div className="flex items-center justify-between mt-4">
                            <p className="text-sm text-muted-foreground">
                                Showing {(page - 1) * 20 + 1} to {Math.min(page * 20, total)} of {total}
                            </p>
                            <div className="flex gap-2">
                                <Button
                                    variant="outline"
                                    size="sm"
                                    disabled={page === 1}
                                    onClick={() => setPage(page - 1)}
                                >
                                    Previous
                                </Button>
                                <Button
                                    variant="outline"
                                    size="sm"
                                    disabled={page * 20 >= total}
                                    onClick={() => setPage(page + 1)}
                                >
                                    Next
                                </Button>
                            </div>
                        </div>
                    )}
                </CardContent>
            </Card>

            {/* Deactivate Dialog */}
            <Dialog open={deactivateDialog.open} onOpenChange={(open) => setDeactivateDialog({ open, user: null })}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>Deactivate User</DialogTitle>
                        <DialogDescription>
                            Are you sure you want to deactivate {deactivateDialog.user?.email}?
                        </DialogDescription>
                    </DialogHeader>
                    <div className="space-y-4">
                        <div>
                            <Label htmlFor="reason">Reason (optional)</Label>
                            <Textarea
                                id="reason"
                                value={deactivateReason}
                                onChange={(e) => setDeactivateReason(e.target.value)}
                                placeholder="Enter reason for deactivation..."
                            />
                        </div>
                    </div>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setDeactivateDialog({ open: false, user: null })}>
                            Cancel
                        </Button>
                        <Button variant="destructive" onClick={handleDeactivate}>
                            Deactivate
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>

            {/* Role Change Dialog */}
            <Dialog open={roleDialog.open} onOpenChange={(open) => setRoleDialog({ open, user: null, role: '' })}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>Change User Role</DialogTitle>
                        <DialogDescription>
                            Update the role for {roleDialog.user?.email}
                        </DialogDescription>
                    </DialogHeader>
                    <div className="space-y-4">
                        <div>
                            <Label htmlFor="role">Role</Label>
                            <Select value={roleDialog.role} onValueChange={(role) => setRoleDialog({ ...roleDialog, role })}>
                                <SelectTrigger>
                                    <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="user">User</SelectItem>
                                    <SelectItem value="admin">Admin</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                    </div>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setRoleDialog({ open: false, user: null, role: '' })}>
                            Cancel
                        </Button>
                        <Button onClick={handleUpdateRole}>
                            Update Role
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </div>
    );
}
