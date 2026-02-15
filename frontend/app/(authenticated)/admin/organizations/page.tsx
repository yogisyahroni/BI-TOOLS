'use client';

export const dynamic = 'force-dynamic';

import { useEffect, useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
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
import { useToast } from '@/hooks/use-toast';
import { organizationApi } from '@/lib/api/admin';
import type { Organization } from '@/types/admin';
import {
    Building2,
    Users,
    Search,
    Plus,
    Settings,
    Trash2,
} from 'lucide-react';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { MoreVertical } from 'lucide-react';

export default function OrganizationManagementPage() {
    const { toast } = useToast();
    const [organizations, setOrganizations] = useState<Organization[]>([]);
    const [loading, setLoading] = useState(true);
    const [page, setPage] = useState(1);
    const [total, setTotal] = useState(0);
    const [search, setSearch] = useState('');
    const [stats, setStats] = useState<any>(null);

    // Dialogs
    const [createDialog, setCreateDialog] = useState(false);
    const [deleteDialog, setDeleteDialog] = useState<{ open: boolean; org: Organization | null }>({
        open: false,
        org: null,
    });

    const [formData, setFormData] = useState({
        name: '',
        description: '',
        ownerId: '',
    });

    useEffect(() => {
        loadOrganizations();
        loadStats();
    }, [page, search]);

    const loadOrganizations = async () => {
        try {
            setLoading(true);
            const data = await organizationApi.list({
                page,
                pageSize: 20,
                search: search || undefined,
            });
            setOrganizations(data.data);
            setTotal(data.pagination.total);
        } catch (error: any) {
            toast({
                title: 'Error',
                description: error.message || 'Failed to load organizations',
                variant: 'destructive',
            });
        } finally {
            setLoading(false);
        }
    };

    const loadStats = async () => {
        try {
            const data = await organizationApi.getStats();
            setStats(data);
        } catch (error) {
            console.error('Failed to load stats:', error);
        }
    };

    const handleCreate = async () => {
        try {
            await organizationApi.create(formData);
            toast({
                title: 'Success',
                description: 'Organization created successfully',
            });
            setCreateDialog(false);
            setFormData({ name: '', description: '', ownerId: '' });
            loadOrganizations();
            loadStats();
        } catch (error: any) {
            toast({
                title: 'Error',
                description: error.message || 'Failed to create organization',
                variant: 'destructive',
            });
        }
    };

    const handleDelete = async () => {
        if (!deleteDialog.org) return;

        try {
            await organizationApi.delete(deleteDialog.org.id);
            toast({
                title: 'Success',
                description: 'Organization deleted successfully',
            });
            setDeleteDialog({ open: false, org: null });
            loadOrganizations();
            loadStats();
        } catch (error: any) {
            toast({
                title: 'Error',
                description: error.message || 'Failed to delete organization',
                variant: 'destructive',
            });
        }
    };

    const formatQuotaPercentage = (current: number, max: number) => {
        if (max === 0) return 0;
        return Math.round((current / max) * 100);
    };

    return (
        <div className="container mx-auto p-6 space-y-6">
            {/* Header */}
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-3xl font-bold">Organization Management</h1>
                    <p className="text-muted-foreground">
                        Manage organizations and workspaces
                    </p>
                </div>
                <Button onClick={() => setCreateDialog(true)}>
                    <Plus className="h-4 w-4 mr-2" />
                    New Organization
                </Button>
            </div>

            {/* Stats */}
            {stats && (
                <div className="grid gap-4 md:grid-cols-3">
                    <Card>
                        <CardHeader className="pb-2">
                            <CardTitle className="text-sm font-medium text-muted-foreground">
                                Total Organizations
                            </CardTitle>
                        </CardHeader>
                        <CardContent>
                            <div className="flex items-center gap-2">
                                <Building2 className="h-4 w-4 text-muted-foreground" />
                                <span className="text-2xl font-bold">{stats.totalOrganizations}</span>
                            </div>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader className="pb-2">
                            <CardTitle className="text-sm font-medium text-muted-foreground">
                                Active Organizations
                            </CardTitle>
                        </CardHeader>
                        <CardContent>
                            <span className="text-2xl font-bold">{stats.activeOrganizations}</span>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader className="pb-2">
                            <CardTitle className="text-sm font-medium text-muted-foreground">
                                Total Members
                            </CardTitle>
                        </CardHeader>
                        <CardContent>
                            <div className="flex items-center gap-2">
                                <Users className="h-4 w-4 text-muted-foreground" />
                                <span className="text-2xl font-bold">{stats.totalMembers}</span>
                            </div>
                        </CardContent>
                    </Card>
                </div>
            )}

            {/* Search */}
            <Card>
                <CardContent className="pt-6">
                    <div className="relative">
                        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                        <Input
                            placeholder="Search organizations..."
                            value={search}
                            onChange={(e) => setSearch(e.target.value)}
                            className="pl-10"
                        />
                    </div>
                </CardContent>
            </Card>

            {/* Organizations Table */}
            <Card>
                <CardHeader>
                    <CardTitle>Organizations</CardTitle>
                    <CardDescription>
                        {total} organization{total !== 1 ? 's' : ''} total
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>Name</TableHead>
                                <TableHead>Members</TableHead>
                                <TableHead>Quota Usage</TableHead>
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
                                        <TableCell><Skeleton className="h-4 w-[120px]" /></TableCell>
                                        <TableCell><Skeleton className="h-4 w-[100px]" /></TableCell>
                                        <TableCell><Skeleton className="h-4 w-[40px]" /></TableCell>
                                    </TableRow>
                                ))
                            ) : organizations.length === 0 ? (
                                <TableRow>
                                    <TableCell colSpan={5} className="text-center text-muted-foreground">
                                        No organizations found
                                    </TableCell>
                                </TableRow>
                            ) : (
                                organizations.map((org) => (
                                    <TableRow key={org.id}>
                                        <TableCell>
                                            <div>
                                                <p className="font-medium">{org.name}</p>
                                                {org.description && (
                                                    <p className="text-sm text-muted-foreground">{org.description}</p>
                                                )}
                                            </div>
                                        </TableCell>
                                        <TableCell>
                                            <div className="flex items-center gap-1">
                                                <Users className="h-4 w-4 text-muted-foreground" />
                                                <span>{org.memberCount}</span>
                                            </div>
                                        </TableCell>
                                        <TableCell>
                                            <div className="space-y-1">
                                                <div className="text-sm">
                                                    <span className="text-muted-foreground">Users: </span>
                                                    <span className="font-medium">
                                                        {org.usage.users}/{org.quota.maxUsers}
                                                    </span>
                                                </div>
                                                <div className="h-1.5 bg-secondary rounded-full overflow-hidden">
                                                    <div
                                                        className="h-full bg-primary transition-all"
                                                        style={{ width: `${formatQuotaPercentage(org.usage.users, org.quota.maxUsers)}%` }}
                                                    />
                                                </div>
                                            </div>
                                        </TableCell>
                                        <TableCell className="text-sm text-muted-foreground">
                                            {new Date(org.createdAt).toLocaleDateString()}
                                        </TableCell>
                                        <TableCell className="text-right">
                                            <DropdownMenu>
                                                <DropdownMenuTrigger asChild>
                                                    <Button variant="ghost" size="sm">
                                                        <MoreVertical className="h-4 w-4" />
                                                    </Button>
                                                </DropdownMenuTrigger>
                                                <DropdownMenuContent align="end">
                                                    <DropdownMenuItem>
                                                        <Settings className="h-4 w-4 mr-2" />
                                                        Settings
                                                    </DropdownMenuItem>
                                                    <DropdownMenuItem>
                                                        <Users className="h-4 w-4 mr-2" />
                                                        Manage Members
                                                    </DropdownMenuItem>
                                                    <DropdownMenuItem
                                                        onClick={() => setDeleteDialog({ open: true, org })}
                                                        className="text-destructive"
                                                    >
                                                        <Trash2 className="h-4 w-4 mr-2" />
                                                        Delete
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

            {/* Create Dialog */}
            <Dialog open={createDialog} onOpenChange={setCreateDialog}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>Create New Organization</DialogTitle>
                        <DialogDescription>
                            Add a new organization to the system
                        </DialogDescription>
                    </DialogHeader>
                    <div className="space-y-4">
                        <div>
                            <Label htmlFor="name">Name</Label>
                            <Input
                                id="name"
                                value={formData.name}
                                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                                placeholder="Enter organization name..."
                            />
                        </div>
                        <div>
                            <Label htmlFor="description">Description (optional)</Label>
                            <Input
                                id="description"
                                value={formData.description}
                                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                                placeholder="Enter description..."
                            />
                        </div>
                        <div>
                            <Label htmlFor="ownerId">Owner User ID</Label>
                            <Input
                                id="ownerId"
                                value={formData.ownerId}
                                onChange={(e) => setFormData({ ...formData, ownerId: e.target.value })}
                                placeholder="Enter owner user ID..."
                            />
                        </div>
                    </div>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setCreateDialog(false)}>
                            Cancel
                        </Button>
                        <Button onClick={handleCreate} disabled={!formData.name || !formData.ownerId}>
                            Create
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>

            {/* Delete Dialog */}
            <Dialog open={deleteDialog.open} onOpenChange={(open) => setDeleteDialog({ open, org: null })}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>Delete Organization</DialogTitle>
                        <DialogDescription>
                            Are you sure you want to delete {deleteDialog.org?.name}? This action cannot be undone.
                        </DialogDescription>
                    </DialogHeader>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setDeleteDialog({ open: false, org: null })}>
                            Cancel
                        </Button>
                        <Button variant="destructive" onClick={handleDelete}>
                            Delete
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </div>
    );
}
