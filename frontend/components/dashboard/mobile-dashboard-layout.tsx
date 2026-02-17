'use client';

import * as React from 'react';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Skeleton } from '@/components/ui/skeleton';
import {
    Sheet,
    SheetContent,
    SheetHeader,
    SheetTitle,
    SheetTrigger,
} from '@/components/ui/sheet';
import {
    LayoutDashboard,
    BarChart3,
    Table2,
    Filter,
    RefreshCw,
    ChevronDown,
    ChevronUp,
    MoreVertical,
    Share2,
    Download,
    Maximize,
    _GripVertical,
    Eye,
    EyeOff,
    _Menu,
    Star,
    Clock,
    ArrowLeft,
} from 'lucide-react';
import { toast } from 'sonner';

// ============================================================
// Types
// ============================================================

interface MobileWidget {
    id: string;
    title: string;
    type: 'chart' | 'kpi' | 'table' | 'text';
    queryId?: string;
    isVisible: boolean;
    order: number;
}

interface KPIData {
    label: string;
    value: string | number;
    change?: number;
    changeLabel?: string;
    icon?: React.ReactNode;
}

interface MobileDashboardLayoutProps {
    dashboardId: string;
    dashboardTitle: string;
    widgets: MobileWidget[];
    kpis?: KPIData[];
    renderWidget: (widgetId: string) => React.ReactNode;
    onRefresh?: () => Promise<void>;
    onShare?: () => void;
    onExport?: () => void;
    onFilterOpen?: () => void;
    onBack?: () => void;
    isLoading?: boolean;
    lastUpdated?: Date;
    activeFilterCount?: number;
    className?: string;
}

// ============================================================
// Sub-Components
// ============================================================

function KPIStrip({ kpis }: { kpis: KPIData[] }) {
    const scrollRef = React.useRef<HTMLDivElement>(null);

    return (
        <div className="relative">
            <div
                ref={scrollRef}
                className="flex gap-3 overflow-x-auto pb-2 snap-x snap-mandatory scrollbar-hide px-4"
                style={{ WebkitOverflowScrolling: 'touch' }}
            >
                {kpis.map((kpi, idx) => (
                    <Card
                        key={idx}
                        className={cn(
                            'flex-shrink-0 snap-start p-3.5 min-w-[150px] max-w-[180px]',
                            'bg-gradient-to-br from-card/80 to-card/40 backdrop-blur-md border-border/30',
                            'shadow-sm hover:shadow-md transition-shadow duration-200',
                        )}
                    >
                        <div className="flex items-center gap-2 mb-1.5">
                            {kpi.icon}
                            <span className="text-[10px] text-muted-foreground font-medium truncate">{kpi.label}</span>
                        </div>
                        <div className="text-xl font-bold tracking-tight">{kpi.value}</div>
                        {kpi.change !== undefined && (
                            <div className={cn(
                                'text-[10px] font-medium mt-0.5',
                                kpi.change >= 0 ? 'text-emerald-500' : 'text-red-500',
                            )}>
                                {kpi.change >= 0 ? '↑' : '↓'} {Math.abs(kpi.change).toFixed(1)}%
                                {kpi.changeLabel && (
                                    <span className="text-muted-foreground/60 ml-1">{kpi.changeLabel}</span>
                                )}
                            </div>
                        )}
                    </Card>
                ))}
            </div>
            {/* Scroll indicator dots */}
            {kpis.length > 2 && (
                <div className="flex justify-center gap-1 mt-1">
                    {kpis.map((_, i) => (
                        <div key={i} className="w-1 h-1 rounded-full bg-muted-foreground/20" />
                    ))}
                </div>
            )}
        </div>
    );
}

function CollapsibleWidget({
    widget,
    renderWidget,
    _onToggleVisibility,
}: {
    widget: MobileWidget;
    renderWidget: (id: string) => React.ReactNode;
    onToggleVisibility: (id: string) => void;
}) {
    const [isCollapsed, setIsCollapsed] = React.useState(false);

    const typeIcons: Record<string, React.ReactNode> = {
        chart: <BarChart3 className="w-3.5 h-3.5 text-blue-400" />,
        kpi: <Star className="w-3.5 h-3.5 text-amber-400" />,
        table: <Table2 className="w-3.5 h-3.5 text-green-400" />,
        text: <LayoutDashboard className="w-3.5 h-3.5 text-purple-400" />,
    };

    if (!widget.isVisible) return null;

    return (
        <Card
            className={cn(
                'overflow-hidden border-border/30 bg-card/70 backdrop-blur-sm',
                'shadow-sm transition-all duration-200',
            )}
        >
            {/* Widget Header — Touch Friendly */}
            <button
                className="w-full flex items-center gap-2.5 px-4 py-3 active:bg-muted/30 transition-colors"
                onClick={() => setIsCollapsed(!isCollapsed)}
            >
                {typeIcons[widget.type]}
                <span className="text-sm font-semibold tracking-tight flex-1 text-left truncate">
                    {widget.title}
                </span>
                <Badge variant="outline" className="h-4 text-[8px] px-1 capitalize">
                    {widget.type}
                </Badge>
                {isCollapsed ? (
                    <ChevronDown className="w-4 h-4 text-muted-foreground" />
                ) : (
                    <ChevronUp className="w-4 h-4 text-muted-foreground" />
                )}
            </button>

            {/* Widget Content */}
            <div
                className={cn(
                    'transition-all duration-300 ease-in-out overflow-hidden',
                    isCollapsed ? 'max-h-0 opacity-0' : 'max-h-[600px] opacity-100',
                )}
            >
                <div className="px-4 pb-4">
                    <div className="rounded-lg overflow-hidden bg-background/30 min-h-[200px]">
                        {renderWidget(widget.id)}
                    </div>
                </div>
            </div>
        </Card>
    );
}

// ============================================================
// Main Component
// ============================================================

export function MobileDashboardLayout({
    _dashboardId,
    dashboardTitle,
    widgets,
    kpis = [],
    renderWidget,
    onRefresh,
    onShare,
    onExport,
    onFilterOpen,
    onBack,
    isLoading = false,
    lastUpdated,
    activeFilterCount = 0,
    className,
}: MobileDashboardLayoutProps) {
    const [visibleWidgets, setVisibleWidgets] = React.useState<Record<string, boolean>>(() => {
        const map: Record<string, boolean> = {};
        widgets.forEach(w => { map[w.id] = w.isVisible; });
        return map;
    });
    const [isRefreshing, setIsRefreshing] = React.useState(false);
    const [activeTab, setActiveTab] = React.useState<'dashboard' | 'charts' | 'tables'>('dashboard');

    React.useEffect(() => {
        const map: Record<string, boolean> = {};
        widgets.forEach(w => { map[w.id] = w.isVisible; });
        setVisibleWidgets(map);
    }, [widgets]);

    const handleRefresh = async () => {
        if (!onRefresh) return;
        setIsRefreshing(true);
        try {
            await onRefresh();
            toast.success('Dashboard refreshed');
        } catch {
            toast.error('Refresh failed');
        } finally {
            setIsRefreshing(false);
        }
    };

    const toggleWidgetVisibility = (id: string) => {
        setVisibleWidgets(prev => ({
            ...prev,
            [id]: !prev[id],
        }));
    };

    // Sort and filter widgets
    const sortedWidgets = [...widgets]
        .sort((a, b) => a.order - b.order)
        .map(w => ({ ...w, isVisible: visibleWidgets[w.id] ?? w.isVisible }));

    const filteredByTab = sortedWidgets.filter(w => {
        if (activeTab === 'charts') return w.type === 'chart' || w.type === 'kpi';
        if (activeTab === 'tables') return w.type === 'table';
        return true;
    });

    const visibleCount = sortedWidgets.filter(w => w.isVisible).length;

    return (
        <div className={cn('flex flex-col h-full bg-background', className)}>
            {/* Mobile Header — Sticky */}
            <header className="sticky top-0 z-40 px-4 py-3 bg-background/80 backdrop-blur-xl border-b border-border/30 safe-area-top">
                <div className="flex items-center gap-3">
                    {onBack && (
                        <Button variant="ghost" size="icon" className="h-8 w-8 -ml-1" onClick={onBack}>
                            <ArrowLeft className="w-4 h-4" />
                        </Button>
                    )}

                    <div className="flex-1 min-w-0">
                        <h1 className="text-base font-bold tracking-tight truncate">{dashboardTitle}</h1>
                        {lastUpdated && (
                            <div className="flex items-center gap-1 text-[10px] text-muted-foreground">
                                <Clock className="w-2.5 h-2.5" />
                                Updated {lastUpdated.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                            </div>
                        )}
                    </div>

                    {/* Action Buttons */}
                    <div className="flex items-center gap-1">
                        {onFilterOpen && (
                            <Button
                                variant="ghost"
                                size="icon"
                                className="h-8 w-8 relative"
                                onClick={onFilterOpen}
                            >
                                <Filter className="w-4 h-4" />
                                {activeFilterCount > 0 && (
                                    <span className="absolute -top-0.5 -right-0.5 w-4 h-4 rounded-full bg-primary text-[9px] text-primary-foreground flex items-center justify-center font-bold">
                                        {activeFilterCount}
                                    </span>
                                )}
                            </Button>
                        )}

                        <Button
                            variant="ghost"
                            size="icon"
                            className="h-8 w-8"
                            onClick={handleRefresh}
                            disabled={isRefreshing}
                        >
                            <RefreshCw className={cn('w-4 h-4', isRefreshing && 'animate-spin')} />
                        </Button>

                        {/* More Actions Sheet */}
                        <Sheet>
                            <SheetTrigger asChild>
                                <Button variant="ghost" size="icon" className="h-8 w-8">
                                    <MoreVertical className="w-4 h-4" />
                                </Button>
                            </SheetTrigger>
                            <SheetContent side="bottom" className="rounded-t-2xl">
                                <SheetHeader>
                                    <SheetTitle className="text-sm">Dashboard Actions</SheetTitle>
                                </SheetHeader>
                                <div className="grid gap-1 py-4">
                                    {onShare && (
                                        <Button variant="ghost" className="justify-start h-12 text-sm" onClick={onShare}>
                                            <Share2 className="w-4 h-4 mr-3" />
                                            Share Dashboard
                                        </Button>
                                    )}
                                    {onExport && (
                                        <Button variant="ghost" className="justify-start h-12 text-sm" onClick={onExport}>
                                            <Download className="w-4 h-4 mr-3" />
                                            Export as PDF
                                        </Button>
                                    )}
                                    <Button variant="ghost" className="justify-start h-12 text-sm" onClick={() => {
                                        // Enter fullscreen
                                        document.documentElement.requestFullscreen?.();
                                    }}>
                                        <Maximize className="w-4 h-4 mr-3" />
                                        Fullscreen Mode
                                    </Button>
                                </div>

                                {/* Widget Visibility Toggles */}
                                <div className="border-t border-border/30 pt-3 pb-2">
                                    <p className="text-xs font-medium text-muted-foreground px-2 mb-2">
                                        Widget Visibility ({visibleCount}/{sortedWidgets.length})
                                    </p>
                                    <div className="space-y-0.5">
                                        {sortedWidgets.map(w => (
                                            <Button
                                                key={w.id}
                                                variant="ghost"
                                                className="justify-start h-10 text-sm w-full"
                                                onClick={() => toggleWidgetVisibility(w.id)}
                                            >
                                                {visibleWidgets[w.id]
                                                    ? <Eye className="w-4 h-4 mr-3 text-emerald-400" />
                                                    : <EyeOff className="w-4 h-4 mr-3 text-muted-foreground" />
                                                }
                                                <span className={cn(!visibleWidgets[w.id] && 'text-muted-foreground line-through')}>
                                                    {w.title}
                                                </span>
                                            </Button>
                                        ))}
                                    </div>
                                </div>
                            </SheetContent>
                        </Sheet>
                    </div>
                </div>

                {/* Tab Bar — Segment Control */}
                <div className="flex gap-1 mt-2 p-0.5 bg-muted/30 rounded-lg">
                    {([
                        { key: 'dashboard', label: 'All', icon: LayoutDashboard },
                        { key: 'charts', label: 'Charts', icon: BarChart3 },
                        { key: 'tables', label: 'Tables', icon: Table2 },
                    ] as const).map(tab => (
                        <button
                            key={tab.key}
                            className={cn(
                                'flex-1 flex items-center justify-center gap-1.5 py-1.5 rounded-md text-xs font-medium transition-all duration-200',
                                activeTab === tab.key
                                    ? 'bg-background shadow-sm text-foreground'
                                    : 'text-muted-foreground hover:text-foreground',
                            )}
                            onClick={() => setActiveTab(tab.key)}
                        >
                            <tab.icon className="w-3.5 h-3.5" />
                            {tab.label}
                        </button>
                    ))}
                </div>
            </header>

            {/* Content Area */}
            <ScrollArea className="flex-1">
                <div className="pb-20">
                    {/* KPI Strip (only on dashboard tab) */}
                    {activeTab === 'dashboard' && kpis.length > 0 && (
                        <div className="pt-4 pb-2">
                            <KPIStrip kpis={kpis} />
                        </div>
                    )}

                    {/* Loading State */}
                    {isLoading ? (
                        <div className="px-4 pt-4 space-y-4">
                            {[1, 2, 3].map(i => (
                                <Card key={i} className="p-4 space-y-3 border-border/30">
                                    <div className="flex items-center gap-2">
                                        <Skeleton className="h-4 w-4 rounded" />
                                        <Skeleton className="h-4 w-32" />
                                    </div>
                                    <Skeleton className="h-[200px] w-full rounded-lg" />
                                </Card>
                            ))}
                        </div>
                    ) : (
                        /* Widget Stack */
                        <div className="px-4 pt-3 space-y-3">
                            {filteredByTab.map(widget => (
                                <CollapsibleWidget
                                    key={widget.id}
                                    widget={widget}
                                    renderWidget={renderWidget}
                                    onToggleVisibility={toggleWidgetVisibility}
                                />
                            ))}

                            {filteredByTab.filter(w => w.isVisible).length === 0 && (
                                <div className="flex flex-col items-center justify-center py-16 text-center">
                                    <LayoutDashboard className="w-10 h-10 text-muted-foreground/20 mb-3" />
                                    <p className="text-sm text-muted-foreground/50">No widgets visible</p>
                                    <p className="text-xs text-muted-foreground/30 mt-1">
                                        Use the menu to show/hide widgets
                                    </p>
                                </div>
                            )}
                        </div>
                    )}
                </div>
            </ScrollArea>

            {/* Pull-to-Refresh Indicator (visual only — actual hook requires touch event lib) */}
            {isRefreshing && (
                <div className="fixed top-14 left-1/2 -translate-x-1/2 z-50">
                    <div className="bg-card/90 backdrop-blur-md shadow-lg rounded-full px-4 py-1.5 flex items-center gap-2 border border-border/30">
                        <RefreshCw className="w-3 h-3 animate-spin text-primary" />
                        <span className="text-[11px] text-muted-foreground">Refreshing...</span>
                    </div>
                </div>
            )}
        </div>
    );
}
