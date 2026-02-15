'use client';

export const dynamic = 'force-dynamic';

import { useState, useEffect } from 'react';
import { SidebarLayout } from '@/components/sidebar-layout';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { ArrowLeft, Plus, Calculator, Trash2, RefreshCw, Table as TableIcon, Eye, EyeOff } from 'lucide-react';
import { useRouter } from 'next/navigation';
import { toast } from 'sonner';
import { Skeleton } from '@/components/ui/skeleton';
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from '@/components/ui/table';
import { CreateCalculatedFieldDialog } from '@/components/semantic/create-calculated-field-dialog';
import { semanticApi } from '@/lib/api/semantic';
import { SemanticModel, SemanticMetric, SemanticDimension } from '@/types/semantic';

interface ModelEditorPageProps {
    params: Promise<{
        id: string;
    }>
}

export default function ModelEditorPage({ params }: ModelEditorPageProps) {
    const router = useRouter();
    const [model, setModel] = useState<SemanticModel | null>(null);
    const [loading, setLoading] = useState(true);
    const [showCreateMetricDialog, setShowCreateMetricDialog] = useState(false);
    const [modelId, setModelId] = useState<string>('');

    useEffect(() => {
        const loadParams = async () => {
            const { id } = await params;
            setModelId(id);
        };
        loadParams();
    }, [params]);

    useEffect(() => {
        if (modelId) {
            loadModelData();
        }
    }, [modelId]);

    const loadModelData = async () => {
        setLoading(true);
        try {
            const data = await semanticApi.getModel(modelId);
            setModel(data);
        } catch (error) {
            console.error(error);
            toast.error('Failed to load model data');
        } finally {
            setLoading(false);
        }
    };

    const handleCreateMetric = async (field: any) => {
        if (!model) return;

        // Current implementation requires full update to add a metric
        const newMetric = {
            name: field.name,
            formula: field.expression,
            description: field.description,
            format: field.dataType
        };

        try {
            const updatedModelRequest = {
                name: model.name,
                description: model.description,
                dataSourceId: model.dataSourceId,
                tableName: model.tableName,
                dimensions: model.dimensions.map(d => ({
                    name: d.name,
                    columnName: d.columnName,
                    dataType: d.dataType,
                    description: d.description,
                    isHidden: d.isHidden
                })),
                metrics: [
                    ...model.metrics.map(m => ({
                        name: m.name,
                        formula: m.formula,
                        description: m.description,
                        format: m.format
                    })),
                    newMetric
                ]
            };

            await semanticApi.updateModel(model.id, updatedModelRequest);
            toast.success('Metric created');
            setShowCreateMetricDialog(false);
            loadModelData(); // Reload to get updated IDs and state
        } catch (error: any) {
            toast.error(error.message || 'Failed to create metric');
        }
    };

    const handleDeleteMetric = async (metricName: string) => {
        if (!model || !confirm(`Are you sure you want to delete metric '${metricName}'?`)) return;

        try {
            const updatedModelRequest = {
                name: model.name,
                description: model.description,
                dataSourceId: model.dataSourceId,
                tableName: model.tableName,
                dimensions: model.dimensions.map(d => ({
                    name: d.name,
                    columnName: d.columnName,
                    dataType: d.dataType,
                    description: d.description,
                    isHidden: d.isHidden
                })),
                metrics: model.metrics.filter(m => m.name !== metricName).map(m => ({
                    name: m.name,
                    formula: m.formula,
                    description: m.description,
                    format: m.format
                }))
            };

            await semanticApi.updateModel(model.id, updatedModelRequest);
            toast.success('Metric deleted');
            loadModelData();
        } catch (error: any) {
            toast.error(error.message || 'Failed to delete metric');
        }
    };

    const toggleDimensionVisibility = async (dim: SemanticDimension) => {
        if (!model) return;

        try {
            const updatedModelRequest = {
                name: model.name,
                description: model.description,
                dataSourceId: model.dataSourceId,
                tableName: model.tableName,
                dimensions: model.dimensions.map(d => {
                    if (d.name === dim.name) {
                        return { ...d, isHidden: !d.isHidden };
                    }
                    return {
                        name: d.name,
                        columnName: d.columnName,
                        dataType: d.dataType,
                        description: d.description,
                        isHidden: d.isHidden
                    };
                }),
                metrics: model.metrics.map(m => ({
                    name: m.name,
                    formula: m.formula,
                    description: m.description,
                    format: m.format
                }))
            };

            await semanticApi.updateModel(model.id, updatedModelRequest);
            toast.success(`Dimension ${dim.isHidden ? 'shown' : 'hidden'}`);
            loadModelData();
        } catch (error: any) {
            toast.error(error.message || 'Failed to update dimension');
        }
    }


    if (loading) {
        return (
            <SidebarLayout>
                <div className="p-8 space-y-4">
                    <Skeleton className="h-12 w-1/3" />
                    <Skeleton className="h-64 w-full" />
                </div>
            </SidebarLayout>
        );
    }

    if (!model) {
        return (
            <SidebarLayout>
                <div className="p-8 text-center">
                    <h2 className="text-xl font-bold text-destructive">Model not found</h2>
                    <Button variant="ghost" onClick={() => router.back()} className="mt-4">
                        <ArrowLeft className="w-4 h-4 mr-2" /> Back
                    </Button>
                </div>
            </SidebarLayout>
        );
    }

    return (
        <SidebarLayout>
            <div className="flex flex-col h-full bg-background font-sans">
                {/* Header */}
                <div className="border-b border-border bg-card px-8 py-6 flex justify-between items-start">
                    <div>
                        <div className="flex items-center gap-2 mb-2">
                            <Button variant="ghost" size="sm" onClick={() => router.back()} className="h-6 px-0 hover:bg-transparent text-muted-foreground hover:text-foreground">
                                <ArrowLeft className="w-4 h-4 mr-1" /> Models
                            </Button>
                        </div>
                        <h1 className="text-2xl font-bold tracking-tight">{model.name}</h1>
                        <p className="text-muted-foreground text-sm font-mono mt-1 bg-muted/50 inline-block px-2 py-0.5 rounded">
                            {model.tableName}
                        </p>
                    </div>
                    <div className="flex gap-2">
                        <Button variant="outline" size="sm" onClick={loadModelData}>
                            <RefreshCw className="w-4 h-4 mr-2" /> Refresh
                        </Button>
                    </div>
                </div>

                <div className="flex-1 overflow-auto p-8 space-y-8">

                    {/* Dimensions Section */}
                    <Card>
                        <CardHeader className="flex flex-row items-center justify-between pb-2">
                            <div className="space-y-1">
                                <CardTitle className="text-lg font-bold flex items-center gap-2">
                                    <TableIcon className="w-5 h-5 text-secondary" />
                                    Dimensions
                                </CardTitle>
                                <CardDescription>
                                    Columns from the source table.
                                </CardDescription>
                            </div>
                        </CardHeader>
                        <CardContent>
                            {model.dimensions?.length === 0 ? (
                                <p className="text-muted-foreground text-sm">No dimensions found. Use 'Refresh' to sync with database schema.</p>
                            ) : (
                                <Table>
                                    <TableHeader>
                                        <TableRow>
                                            <TableHead>Name</TableHead>
                                            <TableHead>Column</TableHead>
                                            <TableHead>Type</TableHead>
                                            <TableHead className="w-[100px] text-right">Visibility</TableHead>
                                        </TableRow>
                                    </TableHeader>
                                    <TableBody>
                                        {model.dimensions?.map((dim) => (
                                            <TableRow key={dim.id} className={dim.isHidden ? 'opacity-50' : ''}>
                                                <TableCell className="font-medium">{dim.name}</TableCell>
                                                <TableCell className="font-mono text-xs">{dim.columnName}</TableCell>
                                                <TableCell className="text-xs text-muted-foreground">{dim.dataType}</TableCell>
                                                <TableCell className="text-right">
                                                    <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => toggleDimensionVisibility(dim)}>
                                                        {dim.isHidden ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                                                    </Button>
                                                </TableCell>
                                            </TableRow>
                                        ))}
                                    </TableBody>
                                </Table>
                            )}
                        </CardContent>
                    </Card>

                    {/* Metrics Section */}
                    <Card>
                        <CardHeader className="flex flex-row items-center justify-between pb-2">
                            <div className="space-y-1">
                                <CardTitle className="text-lg font-bold flex items-center gap-2">
                                    <Calculator className="w-5 h-5 text-primary" />
                                    Metrics
                                </CardTitle>
                                <CardDescription>
                                    Define calculations to be computed on the fly.
                                </CardDescription>
                            </div>
                            <Button size="sm" onClick={() => setShowCreateMetricDialog(true)}>
                                <Plus className="w-4 h-4 mr-2" />
                                Add Metric
                            </Button>
                        </CardHeader>
                        <CardContent>
                            {model.metrics?.length === 0 ? (
                                <div className="text-center py-12 text-muted-foreground bg-muted/10 rounded-xl border border-dashed">
                                    <p>No metrics defined for this model.</p>
                                    <Button variant="link" onClick={() => setShowCreateMetricDialog(true)}>Create one now</Button>
                                </div>
                            ) : (
                                <Table>
                                    <TableHeader>
                                        <TableRow>
                                            <TableHead>Name</TableHead>
                                            <TableHead>Formula</TableHead>
                                            <TableHead>Description</TableHead>
                                            <TableHead className="w-[100px] text-right">Actions</TableHead>
                                        </TableRow>
                                    </TableHeader>
                                    <TableBody>
                                        {model.metrics?.map((metric) => (
                                            <TableRow key={metric.id}>
                                                <TableCell className="font-medium text-primary">
                                                    {metric.name}
                                                </TableCell>
                                                <TableCell>
                                                    <code className="text-xs bg-muted px-2 py-1 rounded font-mono border border-border">
                                                        {metric.formula}
                                                    </code>
                                                </TableCell>
                                                <TableCell className="text-muted-foreground text-sm">
                                                    {metric.description || '-'}
                                                </TableCell>
                                                <TableCell className="text-right">
                                                    <Button
                                                        variant="ghost"
                                                        size="icon"
                                                        className="h-8 w-8 text-destructive hover:bg-destructive/10"
                                                        onClick={() => handleDeleteMetric(metric.name)}
                                                    >
                                                        <Trash2 className="w-4 h-4" />
                                                    </Button>
                                                </TableCell>
                                            </TableRow>
                                        ))}
                                    </TableBody>
                                </Table>
                            )}
                        </CardContent>
                    </Card>
                </div>
            </div>

            <CreateCalculatedFieldDialog
                open={showCreateMetricDialog}
                onOpenChange={setShowCreateMetricDialog}
                connectionId={model.dataSourceId}
                modelId={model.id}
                existingMetrics={model.metrics?.map(m => m.name) || []}
                onSave={handleCreateMetric}
            />
        </SidebarLayout>
    );
}
