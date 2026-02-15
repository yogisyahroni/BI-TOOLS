'use client';

export const dynamic = 'force-dynamic';

import { useState } from 'react';
import Link from 'next/link';
import { useDashboards } from '@/hooks/use-dashboards';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { LayoutDashboard, Plus, Search, Loader2, FileBarChart, Calendar, MoreVertical, Trash2, Copy } from 'lucide-react';
import { SidebarLayout } from '@/components/sidebar-layout';
import { formatDistanceToNow } from 'date-fns';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { toast } from 'sonner';
import { VerificationBadge } from '@/components/catalog/verification-badge';

export default function DashboardsPage() {
  const {
    dashboards,
    isLoading,
    error,
    createDashboard,
    deleteDashboard,
    duplicateDashboard
  } = useDashboards({ autoFetch: true });

  const [searchQuery, setSearchQuery] = useState('');
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [newDashboardName, setNewDashboardName] = useState('');
  const [isCreating, setIsCreating] = useState(false);

  const filteredDashboards = dashboards.filter(d =>
    d.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    (d.description && d.description.toLowerCase().includes(searchQuery.toLowerCase()))
  );

  const handleCreate = async () => {
    if (!newDashboardName.trim()) return;

    setIsCreating(true);
    const result = await createDashboard({
      name: newDashboardName,
      description: 'New dashboard created via Explorer',
      collectionId: '', // Default collection? or handle backend default
    });

    setIsCreating(false);

    if (result.success) {
      toast.success('Dashboard created successfully');
      setIsCreateOpen(false);
      setNewDashboardName('');
    } else {
      toast.error(result.error || 'Failed to create dashboard');
    }
  };

  const handleDelete = async (id: string, e: React.MouseEvent) => {
    e.preventDefault(); // Prevent link navigation
    e.stopPropagation();
    if (confirm('Are you sure you want to delete this dashboard?')) {
      const result = await deleteDashboard(id);
      if (result.success) {
        toast.success('Dashboard deleted');
      } else {
        toast.error(result.error);
      }
    }
  };

  const handleDuplicate = async (id: string, e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    const result = await duplicateDashboard(id);
    if (result.success) {
      toast.success('Dashboard duplicated');
    } else {
      toast.error(result.error);
    }
  };

  return (
    <SidebarLayout>
      <div className="container py-8 max-w-7xl mx-auto">
        <div className="flex flex-col md:flex-row items-start md:items-center justify-between mb-8 gap-4">
          <div>
            <h1 className="text-3xl font-bold flex items-center gap-3">
              <LayoutDashboard className="h-8 w-8 text-primary" />
              Dashboards
            </h1>
            <p className="text-muted-foreground mt-2">
              Manage and view your analytics dashboards
            </p>
          </div>

          <div className="flex items-center gap-2 w-full md:w-auto">
            <div className="relative flex-1 md:w-64">
              <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
              <Input
                type="search"
                placeholder="Search dashboards..."
                className="pl-8"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
              />
            </div>

            <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
              <DialogTrigger asChild>
                <Button className="gap-2">
                  <Plus className="h-4 w-4" />
                  New Dashboard
                </Button>
              </DialogTrigger>
              <DialogContent>
                <DialogHeader>
                  <DialogTitle>Create New Dashboard</DialogTitle>
                  <DialogDescription>
                    Enter a name for your new dashboard.
                  </DialogDescription>
                </DialogHeader>
                <div className="grid gap-4 py-4">
                  <Input
                    placeholder="Dashboard Name"
                    value={newDashboardName}
                    onChange={(e) => setNewDashboardName(e.target.value)}
                    onKeyDown={(e) => e.key === 'Enter' && handleCreate()}
                  />
                </div>
                <DialogFooter>
                  <Button variant="outline" onClick={() => setIsCreateOpen(false)}>Cancel</Button>
                  <Button onClick={handleCreate} disabled={!newDashboardName.trim() || isCreating}>
                    {isCreating ? <Loader2 className="h-4 w-4 animate-spin mr-2" /> : null}
                    Create
                  </Button>
                </DialogFooter>
              </DialogContent>
            </Dialog>
          </div>
        </div>

        {error && (
          <div className="bg-destructive/10 text-destructive p-4 rounded-lg mb-6 flex items-center gap-2">
            <span>Error: {error}</span>
            <Button variant="outline" size="sm" onClick={() => window.location.reload()} className="ml-auto">Retry</Button>
          </div>
        )}

        {isLoading ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {[1, 2, 3].map((i) => (
              <Card key={i} className="h-[200px] animate-pulse bg-muted/50" />
            ))}
          </div>
        ) : filteredDashboards.length === 0 ? (
          <div className="text-center py-20 border-2 border-dashed rounded-lg bg-muted/10">
            <FileBarChart className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
            <h3 className="text-lg font-medium">No dashboards found</h3>
            <p className="text-muted-foreground mb-6">
              {searchQuery ? "Try adjusting your search query" : "Create your first dashboard to get started"}
            </p>
            {!searchQuery && (
              <Button onClick={() => setIsCreateOpen(true)}>
                <Plus className="h-4 w-4 mr-2" />
                Create Dashboard
              </Button>
            )}
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {filteredDashboards.map((dashboard) => (
              <Link key={dashboard.id} href={`/dashboards/${dashboard.id}`} className="block group h-full">
                <Card className="h-full hover:shadow-md transition-all duration-200 border-border/50 hover:border-primary/50 flex flex-col">
                  <CardHeader className="pb-3">
                    <div className="flex items-start justify-between">
                      <div className="space-y-1">
                        <CardTitle className="truncate text-base flex items-center gap-2" title={dashboard.name}>
                          <span className="truncate">{dashboard.name}</span>
                          <VerificationBadge status={dashboard.certificationStatus || 'none'} />
                        </CardTitle>
                        <CardDescription className="line-clamp-2 text-xs">
                          {dashboard.description || 'No description'}
                        </CardDescription>
                      </div>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="icon" className="h-8 w-8 -mt-1 -mr-2 opacity-0 group-hover:opacity-100 transition-opacity" onClick={(e) => e.preventDefault()}>
                            <MoreVertical className="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem onClick={(e) => handleDuplicate(dashboard.id, e)}>
                            <Copy className="h-4 w-4 mr-2" /> Duplicate
                          </DropdownMenuItem>
                          <DropdownMenuItem className="text-destructive focus:text-destructive" onClick={(e) => handleDelete(dashboard.id, e)}>
                            <Trash2 className="h-4 w-4 mr-2" /> Delete
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </div>
                  </CardHeader>
                  <CardContent className="flex-1 pb-3">
                    <div className="flex items-center gap-2 text-xs text-muted-foreground mb-2">
                      <span className="inline-flex items-center rounded-full border px-2 py-0.5 text-xs font-semibold transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 border-transparent bg-secondary text-secondary-foreground hover:bg-secondary/80">
                        {dashboard.cards?.length || 0} Widgets
                      </span>
                      {dashboard.isPublic && (
                        <span className="inline-flex items-center rounded-full border px-2 py-0.5 text-xs font-semibold border-transparent bg-green-500/10 text-green-500">
                          Public
                        </span>
                      )}
                    </div>
                  </CardContent>
                  <CardFooter className="pt-0 border-t bg-muted/5 mt-auto p-4">
                    <div className="flex items-center text-xs text-muted-foreground w-full">
                      <Calendar className="h-3 w-3 mr-1" />
                      <span>Updated {formatDistanceToNow(new Date(dashboard.updatedAt), { addSuffix: true })}</span>
                    </div>
                  </CardFooter>
                </Card>
              </Link>
            ))}
          </div>
        )}
      </div>
    </SidebarLayout>
  );
}
