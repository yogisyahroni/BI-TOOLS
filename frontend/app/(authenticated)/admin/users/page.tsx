'use client';

import { useEffect, useState, useCallback } from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
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

    const loadUsers = useCallback(async () => {
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
        } catch (error: unknown) {
            const errorMessage = error instanceof Error ? error.message : 'Failed to load users';
            toast({
                title: 'Error',
                description: errorMessage,
                variant: 'destructive',
            });
        } finally {
            setLoading(false);
        }
    }, [page, search, statusFilter, roleFilter, toast]);

    const loadStats = useCallback(async () => {
        try {
            const data = await userAdminApi.getStats();
            setStats(data);
        } catch (error) {
            console.error('Failed to load stats:', error);
        }
    }, []);

    useEffect(() => {
        loadUsers();
        loadStats();
    }, [loadUsers, loadStats]);

    const handleActivate = async (user: AdminUser) => {
        try {
            await userAdminApi.activate(user.id);
            toast({
                title: 'Success',
                description: `User ${user.email} has been activated`,
            });
            loadUsers();
            loadStats();
        } catch (error: unknown) {
            const errorMessage = error instanceof Error ? error.message : 'Failed to activate user';
            toast({
                title: 'Error',
                description: errorMessage,
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
        } catch (error: unknown) {
            const errorMessage = error instanceof Error ? error.message : 'Failed to deactivate user';
            toast({
                title: 'Error',
                description: errorMessage,
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
        } catch (error: unknown) {
            const errorMessage = error instanceof Error ? error.message : 'Failed to update role';
            toast({
                title: 'Error',
                description: errorMessage,
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
        } catch (error: unknown) {
            const errorMessage = error instanceof Error ? error.message : 'Failed to impersonate user';
            toast({
                title: 'Error',
                description: errorMessage,
                variant: 'destructive',
            });
        }
    };

    const getStatusBadge = (status: string) => {
        const config: Record<string, { variant: "default" | "secondary" | "outline" | "destructive"; icon: React.ElementType }> = {
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

    const renderTableBody = () => {
        if (loading) {
            return Array.from({ length: 5 }).map((_, i) => (
                <TableRow key={i}>
                    <TableCell><Skeleton className="h-4 w-[200px]" /></TableCell>
                    <TableCell><Skeleton className="h-4 w-[80px]" /></TableCell>
                    <TableCell><Skeleton className="h-4 w-[80px]" /></TableCell>
                    <TableCell><Skeleton className="h-4 w-[120px]" /></TableCell>
                    <TableCell><Skeleton className="h-4 w-[40px]" /></TableCell>
                </TableRow>
            ));
        }

        if (users.length === 0) {
            return (
                <TableRow>
                    <TableCell colSpan={5} className="text-center text-muted-foreground">
                        No users found
                    </TableCell>
                </TableRow>
            );
        }

        return users.map((user) => (
            <TableRow key={user.id}>
                <TableCell>
                    <div>
                        <div className="font-medium">{user.name || 'N/A'}</div>
                        <div className="text-sm text-muted-foreground">{user.email}</div>
                    </div>
                </TableCell>
                <TableCell>
                    <Badge variant="outline">{user.role}</Badge>
                </TableCell>
                <TableCell>
                    {getStatusBadge(user.status)}
                </TableCell>
                <TableCell className="text-muted-foreground">
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
                            <DropdownMenuItem onClick={() => handleImpersonate(user)}>
                                <Eye className="h-4 w-4 mr-2" />
                                Impersonate
                            </DropdownMenuItem>
                            <DropdownMenuSeparator />
                            <DropdownMenuItem onClick={() => setRoleDialog({ open: true, user, role: user.role })}>
                                <Shield className="h-4 w-4 mr-2" />
                                Update Role
                            </DropdownMenuItem>
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
                        </DropdownMenuContent>
                    </DropdownMenu>
                </TableCell>
            </TableRow>
        ));
    };

    return (
        <div className="container mx-auto p-6 space-y-6">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-3xl font-bold">User Management</h1>
                    <p className="text-muted-foreground">Manage users and roles</p>
                </div>
            </div>

            {/* Stats */}
            {stats && (
                <div className="grid gap-4 md:grid-cols-3">
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
                            <span className="text-2xl font-bold">{stats.activeUsers}</span>
                        </CardContent>
                    </Card>
                    <Card>
                        <CardHeader className="pb-2">
                            <CardTitle className="text-sm font-medium text-muted-foreground">
                                New (This Month)
                            </CardTitle>
                        </CardHeader>
                        <CardContent>
                            <span className="text-2xl font-bold">{stats.newUsersLastMonth}</span>
                        </CardContent>
                    </Card>
                </div>
            )}

            {/* Filters */}
            <Card>
                <CardContent className="pt-6">
                    <div className="flex flex-col md:flex-row gap-4">
                        <div className="relative flex-1">
                            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                            <Input
                                placeholder="Search users by name or email..."
                                value={search}
                                onChange={(e) => setSearch(e.target.value)}
                                className="pl-10"
                            />
                        </div>
                        <Select value={statusFilter} onValueChange={setStatusFilter}>
                            <SelectTrigger className="w-[180px]">
                                <SelectValue placeholder="Status" />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="all">All Status</SelectItem>
                                <SelectItem value="active">Active</SelectItem>
                                <SelectItem value="inactive">Inactive</SelectItem>
                                <SelectItem value="pending">Pending</SelectItem>
                            </SelectContent>
                        </Select>
                        <Select value={roleFilter} onValueChange={setRoleFilter}>
                            <SelectTrigger className="w-[180px]">
                                <SelectValue placeholder="Role" />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="all">All Roles</SelectItem>
                                <SelectItem value="user">User</SelectItem>
                                <SelectItem value="admin">Admin</SelectItem>
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
                                <TableHead>Role</TableHead>
                                <TableHead>Status</TableHead>
                                <TableHead>Joined</TableHead>
                                <TableHead className="text-right">Actions</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {renderTableBody()}
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
                    <div className="space-y-2">
                        <Label htmlFor="reason">Reason (Optional)</Label>
                        <Textarea
                            id="reason"
                            value={deactivateReason}
                            onChange={(e) => setDeactivateReason(e.target.value)}
                            placeholder="Why is this user being deactivated?"
                        />
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

            {/* Role Dialog */}
            <Dialog open={roleDialog.open} onOpenChange={(open) => setRoleDialog({ open, user: null, role: '' })}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>Update User Role</DialogTitle>
                        <DialogDescription>
                            Change the role for {roleDialog.user?.email}
                        </DialogDescription>
                    </DialogHeader>
                    <div className="space-y-2">
                        <Label>Role</Label>
                        <Select value={roleDialog.role} onValueChange={(role) => setRoleDialog(prev => ({ ...prev, role }))}>
                            <SelectTrigger>
                                <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="user">User</SelectItem>
                                <SelectItem value="admin">Admin</SelectItem>
                            </SelectContent>
                        </Select>
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
