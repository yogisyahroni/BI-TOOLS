'use client';

export const dynamic = 'force-dynamic';

import { useState, useEffect } from 'react';
import { fetchWithAuth } from '@/lib/utils';
import { useWorkspace } from '@/contexts/workspace-context';
import { TablePicker } from '@/components/query-builder/table-picker';
import { ColumnSelector } from '@/components/query-builder/column-selector';
import { FilterBuilder } from '@/components/query-builder/filter-builder';
import { SortSelector } from '@/components/query-builder/sort-selector';
import { QueryPreview } from '@/components/query-builder/query-preview';
import { SaveQueryDialog } from '@/components/saved-queries/save-query-dialog';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select';
import { Separator } from '@/components/ui/separator';
import {
    type QueryBuilderState,
    createInitialState,
    type ColumnSelection,
    type FilterGroup,
    type SortRule,
} from '@/lib/query-builder/types';
import { toast } from 'sonner';
import { RotateCcw, Database, Sparkles } from 'lucide-react';
import { PageLayout } from '@/components/page-layout';
import { PageHeader, PageActions, PageContent } from '@/components/page-header';

export default function QueryBuilderPage() {
    const { workspace } = useWorkspace();
    const [connections, setConnections] = useState<any[]>([]);
    const [isLoadingConnections, setIsLoadingConnections] = useState(true);

    // State
    const [connectionId, setConnectionId] = useState<string>('');
    const [qbState, setQbState] = useState<QueryBuilderState | null>(null);
    const [isSaveDialogOpen, setIsSaveDialogOpen] = useState(false);
    const [generatedSql, setGeneratedSql] = useState('');

    // Load connections on mount
    useEffect(() => {
        fetchConnections();
    }, [workspace]);

    // Initial state when connection changes
    useEffect(() => {
        if (connectionId) {
            setQbState(createInitialState(connectionId));
        } else {
            setQbState(null);
        }
    }, [connectionId]);

    const fetchConnections = async () => {
        if (!workspace) return;
        try {
            setIsLoadingConnections(true);
            const res = await fetchWithAuth(`/api/go/connections?workspaceId=${workspace.id}`);
            if (res.ok) {
                const json = await res.json();
                // Handle both { data: [...] } and [...] formats
                const data = Array.isArray(json) ? json : (Array.isArray(json.data) ? json.data : []);

                setConnections(data);

                // Auto-select first connection if available and none selected
                if (data.length > 0 && !connectionId) {
                    setConnectionId(data[0].id);
                }
            }
        } catch (error) {
            console.error('Failed to fetch connections:', error);
            toast.error('Failed to load connections');
            setConnections([]); // Fallback to empty array
        } finally {
            setIsLoadingConnections(false);
        }
    };

    const handleTableSelect = (table: string) => {
        if (!qbState) return;
        // Reset columns, filters, sorts when table changes
        setQbState({
            ...qbState,
            table,
            columns: [],
            filters: {
                id: 'root',
                operator: 'AND',
                conditions: [],
            },
            sorts: [],
        });
    };

    const handleColumnsChange = (columns: ColumnSelection[]) => {
        if (!qbState) return;
        setQbState({ ...qbState, columns });
    };

    const handleFiltersChange = (filters: FilterGroup) => {
        if (!qbState) return;
        setQbState({ ...qbState, filters });
    };

    const handleSortsChange = (sorts: SortRule[]) => {
        if (!qbState) return;
        setQbState({ ...qbState, sorts });
    };

    const handleReset = () => {
        if (connectionId) {
            setQbState(createInitialState(connectionId));
            toast.success('Query builder reset');
        }
    };

    const handleSaveQuery = (sql: string) => {
        setGeneratedSql(sql);
        setIsSaveDialogOpen(true);
    };

    if (!workspace) {
        return (
            <PageLayout>
                <div className="flex h-[60vh] items-center justify-center">
                    <div className="text-center">
                        <div className="h-16 w-16 rounded-2xl bg-muted/50 flex items-center justify-center mx-auto mb-4">
                            <Database className="h-8 w-8 text-muted-foreground" />
                        </div>
                        <h3 className="text-lg font-semibold mb-2">Select a Workspace</h3>
                        <p className="text-muted-foreground">
                            Please select a workspace to start building queries
                        </p>
                    </div>
                </div>
            </PageLayout>
        );
    }

    return (
        <PageLayout className="p-4 lg:p-6">
            <PageHeader
                title="Query Builder"
                description="Visually build SQL queries without writing code"
                icon={Database}
                badge="Visual"
                badgeVariant="secondary"
                actions={
                    <PageActions>
                        {qbState?.table && (
                            <Button
                                variant="outline"
                                onClick={handleReset}
                            >
                                <RotateCcw className="mr-2 h-4 w-4" />
                                Reset
                            </Button>
                        )}
                    </PageActions>
                }
            />

            <PageContent>
                <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
                    {/* Left Panel: Configuration */}
                    <div className="lg:col-span-4 space-y-4">
                        <Card className="border-border/50">
                            <CardHeader className="pb-3">
                                <CardTitle className="text-base flex items-center gap-2">
                                    <Database className="h-4 w-4 text-primary" />
                                    Data Source
                                </CardTitle>
                            </CardHeader>
                            <CardContent className="space-y-4">
                                <div className="space-y-2">
                                    <label className="text-sm font-medium">Connection</label>
                                    <Select
                                        value={connectionId}
                                        onValueChange={setConnectionId}
                                        disabled={isLoadingConnections}
                                    >
                                        <SelectTrigger className="bg-muted/50">
                                            <SelectValue placeholder="Select connection..." />
                                        </SelectTrigger>
                                        <SelectContent>
                                            {connections.map((conn) => (
                                                <SelectItem key={conn.id} value={conn.id}>
                                                    {conn.name} ({conn.type})
                                                </SelectItem>
                                            ))}
                                        </SelectContent>
                                    </Select>
                                </div>

                                {connectionId && qbState && (
                                    <>
                                        <Separator />
                                        <TablePicker
                                            connectionId={connectionId}
                                            selectedTable={qbState.table}
                                            onTableSelect={handleTableSelect}
                                        />
                                    </>
                                )}
                            </CardContent>
                        </Card>

                        {qbState?.table && (
                            <ColumnSelector
                                connectionId={connectionId}
                                tableName={qbState.table}
                                selectedColumns={qbState.columns}
                                onColumnsChange={handleColumnsChange}
                            />
                        )}
                    </div>

                    {/* Center Panel: Filters & Logic */}
                    <div className="lg:col-span-4 space-y-4">
                        {qbState?.table ? (
                            <>
                                <FilterBuilder
                                    availableColumns={qbState.columns.map((c) => c.column)}
                                    filters={qbState.filters}
                                    onFiltersChange={handleFiltersChange}
                                />
                                <SortSelector
                                    availableColumns={qbState.columns.map((c) => c.column)}
                                    sorts={qbState.sorts}
                                    onSortsChange={handleSortsChange}
                                />
                            </>
                        ) : (
                            <Card className="h-full min-h-[300px] border-dashed border-2 bg-muted/20 flex items-center justify-center">
                                <CardContent className="text-center py-10">
                                    <div className="h-12 w-12 rounded-xl bg-muted/50 flex items-center justify-center mx-auto mb-4">
                                        <Sparkles className="h-6 w-6 text-muted-foreground" />
                                    </div>
                                    <p className="text-muted-foreground">
                                        Select a table to configure filters and sorting
                                    </p>
                                </CardContent>
                            </Card>
                        )}
                    </div>

                    {/* Right Panel: Preview & Results */}
                    <div className="lg:col-span-4 space-y-4">
                        {qbState ? (
                            <QueryPreview
                                state={qbState}
                                onSave={handleSaveQuery}
                                databaseType={connections.find(c => c.id === connectionId)?.type}
                            />
                        ) : (
                            <Card className="h-full min-h-[300px] border-dashed border-2 bg-muted/20 flex items-center justify-center">
                                <CardContent className="text-center py-10">
                                    <div className="h-12 w-12 rounded-xl bg-muted/50 flex items-center justify-center mx-auto mb-4">
                                        <Database className="h-6 w-6 text-muted-foreground" />
                                    </div>
                                    <p className="text-muted-foreground">
                                        Query preview will appear here
                                    </p>
                                </CardContent>
                            </Card>
                        )}
                    </div>
                </div>
            </PageContent>

            {/* Save Query Dialog */}
            <SaveQueryDialog
                open={isSaveDialogOpen}
                onOpenChange={setIsSaveDialogOpen}
                sql={generatedSql}
                connectionId={connectionId}
                aiPrompt="Created via Visual Query Builder"
                onSaveSuccess={() => {
                    // Nothing specific needed here
                }}
            />
        </PageLayout>
    );
}
