'use client';

export const dynamic = 'force-dynamic';

import { useState, useEffect } from 'react';
import { SidebarLayout } from '@/components/sidebar-layout';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Plus, Database, ArrowRight, Table as TableIcon, Code, Layers } from 'lucide-react';
import { useConnections } from '@/hooks/use-connections';
import { Skeleton } from '@/components/ui/skeleton';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { useRouter } from 'next/navigation';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter, DialogDescription } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { toast } from 'sonner';
import { semanticApi } from '@/lib/api/semantic';
import { type SemanticModel } from '@/types/semantic';

export default function ModelingPage() {
    // TODO: Get real workspaceId context if needed, but API usually infers from header/token
    const { connections, isLoading: connectionsLoading } = useConnections();
    const [selectedConnId, setSelectedConnId] = useState<string>('');
    const [models, setModels] = useState<SemanticModel[]>([]);
    const [loadingModels, setLoadingModels] = useState(false);
    const router = useRouter();

    // New Model State
    const [isCreateOpen, setIsCreateOpen] = useState(false);
    const [newModelName, setNewModelName] = useState('');
    const [newModelDescription, setNewModelDescription] = useState('');
    const [newModelTable, setNewModelTable] = useState(''); // Simple table name for MVP
    const [creating, setCreating] = useState(false);

    useEffect(() => {
        if (selectedConnId) {
            fetchModels(selectedConnId);
        } else {
            setModels([]);
        }
    }, [selectedConnId]);

    const fetchModels = async (connId: string) => {
        setLoadingModels(true);
        try {
            const data = await semanticApi.getModels();
            // Filter by selected connection
            const relevantModels = data.filter((m) => m.dataSourceId === connId);
            setModels(relevantModels);
        } catch (error) {
            console.error('Failed to load models', error);
            toast.error('Failed to load models');
        } finally {
            setLoadingModels(false);
        }
    };

    const handleCreateModel = async () => {
        if (!selectedConnId || !newModelName || !newModelTable) return;
        setCreating(true);
        try {
            await semanticApi.createModel({
                name: newModelName,
                description: newModelDescription,
                dataSourceId: selectedConnId,
                tableName: newModelTable,
                dimensions: [], // Empty initially, added in detail view
                metrics: []    // Empty initially
            });

            toast.success('Model created');
            setIsCreateOpen(false);
            fetchModels(selectedConnId);
            setNewModelName('');
            setNewModelTable('');
            setNewModelDescription('');
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        } catch (error: any) {
            toast.error(error.message || 'Failed to create model');
        } finally {
            setCreating(false);
        }
    };

    return (
        <SidebarLayout>
            <div className="flex flex-col h-full bg-background overflow-hidden font-sans">
                {/* Header */}
                <div className="border-b border-border bg-card px-8 py-6">
                    <div className="flex items-center gap-2 mb-2">
                        <Layers className="w-6 h-6 text-primary" />
                        <h1 className="text-2xl font-bold tracking-tight bg-gradient-to-r from-primary to-primary/60 bg-clip-text text-transparent">Semantic Modeling</h1>
                    </div>
                    <p className="text-muted-foreground text-sm mt-0.5">
                        Define business logic, metrics, and relationships on top of your physical data tables.
                    </p>
                </div>

                <div className="flex-1 overflow-auto p-8 space-y-8 pb-20">
                    {/* Connection Selector */}
                    <Card className="border-primary/10 shadow-sm bg-card/50 backdrop-blur-sm">
                        <CardHeader className="pb-3">
                            <CardTitle className="text-base flex items-center gap-2">
                                <Database className="w-4 h-4 text-primary" />
                                Data Source
                            </CardTitle>
                        </CardHeader>
                        <CardContent>
                            {connectionsLoading ? (
                                <Skeleton className="h-10 w-full max-w-md rounded-lg" />
                            ) : (
                                <Select value={selectedConnId} onValueChange={setSelectedConnId}>
                                    <SelectTrigger className="bg-background/50 border-border/50 ring-offset-background transition-all focus:ring-1 focus:ring-primary w-full md:w-[400px]">
                                        <SelectValue placeholder="Select a connection..." />
                                    </SelectTrigger>
                                    <SelectContent>
                                        {connections.map(conn => (
                                            <SelectItem key={conn.id} value={conn.id}>
                                                {conn.name} <span className="text-muted-foreground ml-2 text-xs">({conn.type})</span>
                                            </SelectItem>
                                        ))}
                                    </SelectContent>
                                </Select>
                            )}
                        </CardContent>
                    </Card>

                    {/* Content Area */}
                    {!selectedConnId ? (
                        <div className="flex flex-col items-center justify-center py-20 text-center opacity-50">
                            <Database className="w-16 h-16 text-muted-foreground mb-4" />
                            <h3 className="text-lg font-semibold">Select a connection to manage models</h3>
                        </div>
                    ) : (
                        <div className="animate-in fade-in slide-in-from-bottom-2 duration-500">
                            <div className="flex justify-between items-center mb-6">
                                <h2 className="text-xl font-bold flex items-center gap-2">
                                    <TableIcon className="w-5 h-5 text-secondary" />
                                    Models
                                </h2>
                                <Button onClick={() => setIsCreateOpen(true)}>
                                    <Plus className="w-4 h-4 mr-2" />
                                    New Model
                                </Button>
                            </div>

                            {loadingModels ? (
                                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                                    <Skeleton className="h-40 rounded-xl" />
                                    <Skeleton className="h-40 rounded-xl" />
                                    <Skeleton className="h-40 rounded-xl" />
                                </div>
                            ) : models.length === 0 ? (
                                <Card className="border-dashed border-2 border-muted bg-muted/10">
                                    <CardContent className="flex flex-col items-center justify-center py-12">
                                        <Code className="w-12 h-12 text-muted-foreground mb-4 opacity-50" />
                                        <h3 className="text-lg font-medium">No Models Found</h3>
                                        <p className="text-muted-foreground text-sm max-w-sm text-center mt-2 mb-6">
                                            Create a semantic model to start defining metrics and dimensions for this data source.
                                        </p>
                                        <Button variant="outline" onClick={() => setIsCreateOpen(true)}>Create First Model</Button>
                                    </CardContent>
                                </Card>
                            ) : (
                                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                                    {models.map((model) => (
                                        <Card
                                            key={model.id}
                                            className="group hover:border-primary/50 transition-all cursor-pointer shadow-sm hover:shadow-md bg-card"
                                            onClick={() => router.push(`/modeling/${model.id}`)}
                                        >
                                            <CardHeader>
                                                <div className="flex justify-between items-start">
                                                    <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center text-primary mb-3 group-hover:scale-110 transition-transform">
                                                        <TableIcon className="w-5 h-5" />
                                                    </div>
                                                    <ArrowRight className="w-4 h-4 text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity" />
                                                </div>
                                                <CardTitle className="text-lg">{model.name}</CardTitle>
                                                <CardDescription className="line-clamp-2" title={model.description}>
                                                    {model.description || 'No description'}
                                                </CardDescription>
                                            </CardHeader>
                                            <CardContent>
                                                <div className="space-y-2">
                                                    <div className="text-xs font-mono bg-muted/50 px-2 py-1 rounded w-fit text-muted-foreground truncate max-w-full" title={model.tableName}>
                                                        {model.tableName}
                                                    </div>
                                                    <div className="flex gap-4 text-xs text-muted-foreground pt-2 border-t border-border/50">
                                                        <div className="flex items-center gap-1">
                                                            <span className="font-bold text-foreground">{model.dimensions?.length || 0}</span> Dims
                                                        </div>
                                                        <div className="flex items-center gap-1">
                                                            <span className="font-bold text-foreground">{model.metrics?.length || 0}</span> Metrics
                                                        </div>
                                                    </div>
                                                </div>
                                            </CardContent>
                                        </Card>
                                    ))}
                                </div>
                            )}
                        </div>
                    )}
                </div>
            </div>

            <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>Create Semantic Model</DialogTitle>
                        <DialogDescription>Map a database table to a semantic model.</DialogDescription>
                    </DialogHeader>
                    <div className="space-y-4 py-4">
                        <div className="space-y-2">
                            <Label>Model Name</Label>
                            <Input
                                placeholder="e.g. Sales Transactions"
                                value={newModelName}
                                onChange={e => setNewModelName(e.target.value)}
                            />
                        </div>
                        <div className="space-y-2">
                            <Label>Description</Label>
                            <Textarea
                                placeholder="Describe what this model represents..."
                                value={newModelDescription}
                                onChange={e => setNewModelDescription(e.target.value)}
                            />
                        </div>
                        <div className="space-y-2">
                            <Label>Source Table Name</Label>
                            <Input
                                placeholder="e.g. public.sales_fact"
                                value={newModelTable}
                                onChange={e => setNewModelTable(e.target.value)}
                            />
                            <p className="text-xs text-muted-foreground">Enter the exact physical table name from your database.</p>
                        </div>
                    </div>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setIsCreateOpen(false)}>Cancel</Button>
                        <Button onClick={handleCreateModel} disabled={creating || !newModelName || !newModelTable}>
                            {creating ? 'Creating...' : 'Create Model'}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </SidebarLayout>
    );
}
