'use client';

export const dynamic = 'force-dynamic';

import { useState, useCallback } from 'react';
import { LayoutDashboard, Save, Edit2, Share2, Plus, ArrowLeft, Loader2, MoreVertical, ShieldCheck, Presentation } from 'lucide-react';
import { Button } from '@/components/ui/button';
import Link from 'next/link';
import { SidebarLayout } from '@/components/sidebar-layout';
import { useDashboard } from '@/hooks/use-dashboard';
import { DashboardGrid } from '@/components/dashboard/dashboard-grid';
import { AddCardDialog } from '@/components/dashboard/add-card-dialog';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { toast } from 'sonner';
import { VerificationBadge } from '@/components/catalog/verification-badge';

export default function DashboardDetailPage({ params }: { params: { id: string } }) {
    const {
        dashboard,
        isLoading,
        error,
        isEditing,
        setIsEditing,
        updateLayout,
        addCard,
        removeCard,
        updateCard,
        saveDashboard,
        certifyDashboard
    } = useDashboard(params.id);

    const [isAddCardOpen, setIsAddCardOpen] = useState(false);
    const [isSaving, setIsSaving] = useState(false);

    const handleSave = async () => {
        setIsSaving(true);
        try {
            await saveDashboard();
            toast.success('Dashboard saved successfully');
        } catch (err) {
            toast.error('Failed to save dashboard');
        } finally {
            setIsSaving(false);
        }
    };

    const handleCertify = async (status: 'verified' | 'deprecated' | 'none') => {
        const result = await certifyDashboard(status);
        if (result.success) {
            toast.success(`Dashboard marked as ${status}`);
        } else {
            toast.error(result.error || 'Failed to update certification');
        }
    };

    const handleAddQuery = useCallback((queryId: string, name: string) => {
        addCard({
            title: name,
            type: 'visualization',
            queryId,
            visualizationConfig: {
                type: 'bar', // Default, should effectively be loaded from query config if available
                xAxis: '',
                yAxis: [],
            }
        });
        toast.success('Widget added');
    }, [addCard]);

    const handleAddText = useCallback((title: string, content: string) => {
        addCard({
            title,
            type: 'text',
            textContent: content
        });
        toast.success('Text widget added');
    }, [addCard]);

    if (isLoading) {
        return (
            <SidebarLayout>
                <div className="container py-8 flex items-center justify-center h-[calc(100vh-4rem)]">
                    <div className="flex flex-col items-center gap-4">
                        <Loader2 className="h-8 w-8 animate-spin text-primary" />
                        <p className="text-muted-foreground">Loading dashboard...</p>
                    </div>
                </div>
            </SidebarLayout>
        );
    }

    if (error || !dashboard) {
        return (
            <SidebarLayout>
                <div className="container py-8">
                    <div className="bg-destructive/10 text-destructive p-6 rounded-lg text-center">
                        <h2 className="text-lg font-semibold mb-2">Error Loading Dashboard</h2>
                        <p className="mb-4">{error || 'Dashboard not found'}</p>
                        <Link href="/dashboards">
                            <Button variant="outline">Return to Dashboards</Button>
                        </Link>
                    </div>
                </div>
            </SidebarLayout>
        );
    }

    return (
        <SidebarLayout>
            <div className="flex flex-col min-h-screen">
                {/* Header */}
                <div className="border-b bg-background sticky top-0 z-10 px-6 py-3 flex items-center justify-between shadow-sm">
                    <div className="flex items-center gap-4">
                        <Link href="/dashboards">
                            <Button variant="ghost" size="icon" className="h-8 w-8">
                                <ArrowLeft className="h-4 w-4" />
                            </Button>
                        </Link>
                        <div>
                            <h1 className="text-xl font-bold flex items-center gap-2">
                                <LayoutDashboard className="h-5 w-5 text-primary" />
                                {dashboard.name}
                                <VerificationBadge status={dashboard.certificationStatus || 'none'} />
                                {dashboard.isPublic && (
                                    <span className="text-xs bg-green-500/10 text-green-500 px-2 py-0.5 rounded-full border border-green-500/20 font-medium">
                                        Public
                                    </span>
                                )}
                            </h1>
                            {dashboard.description && (
                                <p className="text-xs text-muted-foreground mt-0.5 hidden md:block">
                                    {dashboard.description}
                                </p>
                            )}
                        </div>
                    </div>

                    <div className="flex items-center gap-2">
                        {isEditing ? (
                            <>
                                <Button variant="outline" size="sm" onClick={() => setIsAddCardOpen(true)}>
                                    <Plus className="h-4 w-4 mr-2" />
                                    Add Widget
                                </Button>
                                <Button size="sm" onClick={handleSave} disabled={isSaving}>
                                    {isSaving ? <Loader2 className="h-4 w-4 animate-spin mr-2" /> : <Save className="h-4 w-4 mr-2" />}
                                    Save Layout
                                </Button>
                                <Button variant="ghost" size="sm" onClick={() => setIsEditing(false)}>
                                    Cancel
                                </Button>
                            </>
                        ) : (
                            <>
                                <Button variant="outline" size="sm" onClick={() => setIsEditing(true)}>
                                    <Edit2 className="h-4 w-4 mr-2" />
                                    Edit
                                </Button>
                                <DropdownMenu>
                                    <DropdownMenuTrigger asChild>
                                        <Link href={`/stories/draft?dashboardId=${dashboard.id}`}>
                                            <Button>
                                                <Presentation className="mr-2 h-4 w-4" />
                                                AI Presentation (Story Builder)
                                            </Button>
                                        </Link>
                                    </DropdownMenuTrigger>
                                    <DropdownMenuContent align="end">
                                        <DropdownMenuItem>
                                            <Share2 className="h-4 w-4 mr-2" /> Share
                                        </DropdownMenuItem>
                                        <DropdownMenuItem onClick={() => handleCertify('verified')}>
                                            <ShieldCheck className="h-4 w-4 mr-2" /> Verify Dashboard
                                        </DropdownMenuItem>
                                    </DropdownMenuContent>
                                </DropdownMenu>
                            </>
                        )}
                        <Link href={`/presentation/${dashboard?.id}`}>
                            <Button variant="outline" size="sm">
                                <Presentation className="h-4 w-4 mr-2" />
                                AI Presentation
                            </Button>
                        </Link>
                    </div>
                </div>

                {/* Main Grid Content */}
                <div className="flex-1 p-6 bg-slate-50/50 dark:bg-slate-950/20 overflow-auto">
                    <DashboardGrid
                        cards={dashboard.cards}
                        isEditing={isEditing}
                        onLayoutChange={updateLayout}
                        onRemoveCard={removeCard}
                        onUpdateCard={updateCard}
                    // queriesData can be fetched inside dashboard-grid per card or centrally.
                    // For the initial restore, we'll assume cards fetch their own data or use `useDashboardData` hook logic if it was implemented that way.
                    // Re-checking useDashboard, it doesn't return data.
                    // However, DashboardCard component often handles data fetching if queryId is present.
                    // Let's rely on DashboardCard internal fetching for now.
                    />
                </div>
            </div>

            <AddCardDialog
                open={isAddCardOpen}
                onOpenChange={setIsAddCardOpen}
                onAddQuery={handleAddQuery}
                onAddText={handleAddText}
            />
        </SidebarLayout>
    );
}
