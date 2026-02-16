'use client';

export const dynamic = 'force-dynamic';

import { useState } from 'react';
import Link from 'next/link';
import { useDashboards } from '@/hooks/use-dashboards';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
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
import { LayoutDashboard, Plus, Search, Loader2, FileBarChart, Calendar, MoreVertical, Trash2, Copy, Sparkles } from 'lucide-react';
import { PageLayout } from '@/components/page-layout';
import { PageHeader, PageActions, PageContent } from '@/components/page-header';
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
      collectionId: '',
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
    e.preventDefault();
    e.stopPropagation();
    // eslint-disable-next-line no-alert
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
    <PageLayout>
      <PageHeader
        title="Dashboards"
        description="Manage and view your analytics dashboards"
        icon={LayoutDashboard}
        badge="Beta"
        badgeVariant="secondary"
        actions={
          <PageActions>
            <div className="relative hidden sm:block">
              <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
              <Input
                type="search"
                placeholder="Search dashboards..."
                className="w-64 pl-8"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
              />
            </div>

            <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
              <DialogTrigger asChild>
                <Button className="gap-2">
                  <Plus className="h-4 w-4" />
                  <span className="hidden sm:inline">New Dashboard</span>
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
          </PageActions>
        }
      />

      <PageContent>
        {/* Mobile Search */}
        <div className="sm:hidden">
          <div className="relative">
            <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input
              type="search"
              placeholder="Search dashboards..."
              className="w-full pl-8"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
            />
          </div>
        </div>

        {error && (
          <div className="bg-destructive/10 text-destructive p-4 rounded-xl flex items-center gap-2">
            <span>Error: {error}</span>
            <Button variant="outline" size="sm" onClick={() => window.location.reload()} className="ml-auto">Retry</Button>
          </div>
        )}

        {isLoading ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {[1, 2, 3, 4].map((i) => (
              <Card key={i} className="h-[220px] animate-pulse bg-muted/50" />
            ))}
          </div>
        ) : filteredDashboards.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-20 text-center border-2 border-dashed border-muted rounded-2xl bg-muted/5">
            <div className="h-16 w-16 rounded-2xl bg-gradient-to-br from-primary/20 to-primary/5 flex items-center justify-center mb-6">
              <FileBarChart className="h-8 w-8 text-primary" />
            </div>
            <h3 className="text-xl font-semibold mb-2">
              {searchQuery ? "No dashboards found" : "Create your first dashboard"}
            </h3>
            <p className="text-muted-foreground mb-6 max-w-md">
              {searchQuery 
                ? "Try adjusting your search query" 
                : "Dashboards help you visualize and share your data insights with your team"}
            </p>
            {!searchQuery && (
              <Button onClick={() => setIsCreateOpen(true)} size="lg" className="gap-2">
                <Plus className="h-5 w-5" />
                Create Dashboard
              </Button>
            )}
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {filteredDashboards.map((dashboard) => (
              <Link key={dashboard.id} href={`/dashboards/${dashboard.id}`} className="block group">
                <Card className="h-full card-hover border-border/50 overflow-hidden">
                  <CardHeader className="pb-3">
                    <div className="flex items-start justify-between">
                      <div className="space-y-1 flex-1 min-w-0">
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
                          <Button 
                            variant="ghost" 
                            size="icon" 
                            className="h-8 w-8 -mt-1 -mr-2 opacity-0 group-hover:opacity-100 transition-opacity"
                            onClick={(e) => e.preventDefault()}
                          >
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
                  <CardContent className="pb-3">
                    <div className="flex items-center gap-2 text-xs text-muted-foreground">
                      <span className="inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-medium bg-secondary">
                        {dashboard.cards?.length || 0} Widgets
                      </span>
                      {dashboard.isPublic && (
                        <span className="inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-medium bg-green-500/10 text-green-600 border-green-500/20">
                          Public
                        </span>
                      )}
                    </div>
                  </CardContent>
                  <div className="border-t bg-muted/30 px-6 py-4">
                    <div className="flex items-center text-xs text-muted-foreground">
                      <Calendar className="h-3.5 w-3.5 mr-1.5" />
                      <span>Updated {formatDistanceToNow(new Date(dashboard.updatedAt), { addSuffix: true })}</span>
                    </div>
                  </div>
                </Card>
              </Link>
            ))}
          </div>
        )}
      </PageContent>
    </PageLayout>
  );
}
