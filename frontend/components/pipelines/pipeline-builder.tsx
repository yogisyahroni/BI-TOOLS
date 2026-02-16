'use client';

import React, { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { toast } from 'sonner';
import {
    ArrowLeft,
    Database,
    Globe,
    FileText,
    Zap,
    Layers,
    Shield,
    Loader2,
    ChevronDown,
    Info,
    AlertTriangle,
} from 'lucide-react';
import { usePipelines } from '@/hooks/use-pipelines';
import { fetchWithAuth } from '@/lib/utils';
import type { TransformStep } from '@/lib/types/batch2';
import { TransformStepBuilder } from './transform-step-builder';
import { SchedulePicker } from './schedule-picker';

// === Schema ===
const pipelineFormSchema = z.object({
    name: z.string().min(2, 'Pipeline name is required (min 2 chars)'),
    description: z.string().optional(),
    sourceType: z.enum(['POSTGRES', 'MYSQL', 'CSV', 'REST_API']),
    sourceConfig: z.string().default('{}'),
    connectionId: z.string().optional(),
    sourceQuery: z.string().optional(),
    destinationType: z.string().default('INTERNAL_RAW'),
    mode: z.enum(['ELT', 'ETL']),
    transformationSteps: z.string().optional(),
    scheduleCron: z.string().nullable().optional(),
    rowLimit: z.number().min(1).max(10000000).default(100000),
});

type PipelineFormValues = z.infer<typeof pipelineFormSchema>;

interface PipelineBuilderProps {
    workspaceId: string;
}

interface ConnectionOption {
    id: string;
    name: string;
    type: string;
    host: string;
    databaseName: string;
}

const SOURCE_TYPES = [
    { value: 'POSTGRES', label: 'PostgreSQL', icon: <Database className="w-4 h-4" />, emoji: '🐘', description: 'Connect to PostgreSQL database' },
    { value: 'MYSQL', label: 'MySQL', icon: <Database className="w-4 h-4" />, emoji: '🐬', description: 'Connect to MySQL database' },
    { value: 'CSV', label: 'CSV Upload', icon: <FileText className="w-4 h-4" />, emoji: '📄', description: 'Upload a CSV file' },
    { value: 'REST_API', label: 'REST API', icon: <Globe className="w-4 h-4" />, emoji: '🌐', description: 'Fetch from REST endpoint' },
];

const MODE_OPTIONS = [
    { value: 'ETL', label: 'ETL', description: 'Extract → Transform → Load. Best for cleaning/filtering before storage.' },
    { value: 'ELT', label: 'ELT', description: 'Extract → Load → Transform. Best for large datasets with in-DB transforms.' },
];

export function PipelineBuilder({ workspaceId }: PipelineBuilderProps) {
    const router = useRouter();
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [connections, setConnections] = useState<ConnectionOption[]>([]);
    const [transformSteps, setTransformSteps] = useState<TransformStep[]>([]);
    const [activeSection, setActiveSection] = useState(0);
    const { createPipeline } = usePipelines({ workspaceId });

    const form = useForm<PipelineFormValues>({
        resolver: zodResolver(pipelineFormSchema),
        defaultValues: {
            name: '',
            description: '',
            sourceType: 'POSTGRES',
            sourceConfig: '{}',
            connectionId: '',
            sourceQuery: '',
            destinationType: 'INTERNAL_RAW',
            mode: 'ETL',
            transformationSteps: '[]',
            scheduleCron: null,
            rowLimit: 100000,
        },
    });

    const sourceType = form.watch('sourceType');
    const selectedConnectionId = form.watch('connectionId');
    const mode = form.watch('mode');

    // Fetch available connections
    useEffect(() => {
        (async () => {
            try {
                const res = await fetchWithAuth('/api/go/connections');
                if (res.ok) {
                    const data = await res.json();
                    const items = Array.isArray(data) ? data : (data.data || []);
                    setConnections(items.map((c: any) => ({
                        id: c.id,
                        name: c.name,
                        type: (c.type || '').toUpperCase(),
                        host: c.host || '',
                        databaseName: c.database || c.databaseName || '',
                    })));
                }
            } catch {
                // Non-critical: connections may not exist yet
            }
        })();
    }, []);

    const filteredConnections = connections.filter(c => {
        if (sourceType === 'POSTGRES') return c.type === 'POSTGRES' || c.type === 'POSTGRESQL';
        if (sourceType === 'MYSQL') return c.type === 'MYSQL';
        return false;
    });

    async function onSubmit(data: PipelineFormValues) {
        setIsSubmitting(true);
        try {
            const payload: any = {
                name: data.name,
                description: data.description || undefined,
                workspaceId,
                sourceType: data.sourceType,
                sourceConfig: data.sourceConfig,
                destinationType: data.destinationType,
                mode: data.mode,
                scheduleCron: data.scheduleCron || undefined,
                rowLimit: data.rowLimit,
            };

            if (data.connectionId) {
                payload.connectionId = data.connectionId;
            }
            if (data.sourceQuery) {
                payload.sourceQuery = data.sourceQuery;
            }
            if (transformSteps.length > 0) {
                payload.transformationSteps = JSON.stringify(transformSteps);
            }

            const result = await createPipeline(payload);
            if (!result.success) {
                throw new Error(result.error || 'Failed to create pipeline');
            }

            toast.success('Pipeline created successfully');
            router.push(`/workspace/${workspaceId}/pipelines`);
        } catch (error: any) {
            toast.error(error.message || 'Failed to create pipeline');
        } finally {
            setIsSubmitting(false);
        }
    }

    const sections = [
        { label: 'Basics', icon: <Info className="w-3.5 h-3.5" /> },
        { label: 'Source', icon: <Database className="w-3.5 h-3.5" /> },
        { label: 'Transform', icon: <Zap className="w-3.5 h-3.5" /> },
        { label: 'Schedule', icon: <Layers className="w-3.5 h-3.5" /> },
    ];

    return (
        <div className="min-h-screen bg-gradient-to-b from-zinc-950 to-black">
            {/* Header */}
            <div className="sticky top-0 z-20 bg-zinc-950/80 backdrop-blur-xl border-b border-white/[0.04]">
                <div className="max-w-3xl mx-auto px-6 py-4 flex items-center justify-between">
                    <div className="flex items-center gap-3">
                        <button
                            type="button"
                            onClick={() => router.back()}
                            className="p-2 rounded-lg hover:bg-white/[0.06] text-zinc-400 hover:text-white transition-all duration-200"
                        >
                            <ArrowLeft className="w-4 h-4" />
                        </button>
                        <div>
                            <h1 className="text-sm font-semibold text-white tracking-tight">Create Pipeline</h1>
                            <p className="text-[11px] text-zinc-500">Configure and deploy a new data pipeline</p>
                        </div>
                    </div>
                    <div className="flex items-center gap-2">
                        <button
                            type="button"
                            onClick={() => router.back()}
                            className="px-4 py-2 rounded-lg text-xs text-zinc-400 hover:text-zinc-200 hover:bg-white/[0.06] transition-all duration-200"
                        >
                            Cancel
                        </button>
                        <button
                            type="button"
                            onClick={form.handleSubmit(onSubmit)}
                            disabled={isSubmitting}
                            className="flex items-center gap-2 px-5 py-2 rounded-lg text-xs font-semibold 
                                bg-gradient-to-r from-violet-600 to-blue-600 text-white
                                hover:from-violet-500 hover:to-blue-500 
                                disabled:opacity-50 disabled:cursor-not-allowed
                                active:scale-[0.97] shadow-lg shadow-violet-600/20 transition-all duration-200"
                        >
                            {isSubmitting && <Loader2 className="w-3.5 h-3.5 animate-spin" />}
                            {isSubmitting ? 'Creating...' : 'Create Pipeline'}
                        </button>
                    </div>
                </div>
            </div>

            {/* Section Nav */}
            <div className="max-w-3xl mx-auto px-6 pt-6">
                <div className="flex items-center gap-1 p-1 rounded-xl bg-black/30 border border-white/[0.04] mb-6">
                    {sections.map((sec, i) => (
                        <button
                            key={sec.label}
                            onClick={() => setActiveSection(i)}
                            className={`flex items-center gap-1.5 px-4 py-2 rounded-lg text-xs font-medium transition-all duration-200 flex-1 justify-center
                                ${activeSection === i
                                    ? 'bg-white/[0.08] text-white shadow-sm'
                                    : 'text-zinc-500 hover:text-zinc-300'
                                }`}
                        >
                            {sec.icon}
                            {sec.label}
                        </button>
                    ))}
                </div>
            </div>

            {/* Form Content */}
            <form onSubmit={form.handleSubmit(onSubmit)} className="max-w-3xl mx-auto px-6 pb-20 space-y-6">

                {/* Section 0: Basics */}
                {activeSection === 0 && (
                    <div className="rounded-xl border border-white/[0.06] bg-zinc-900/50 overflow-hidden">
                        <div className="p-6 space-y-5">
                            <div>
                                <label className="block text-xs font-semibold text-zinc-300 mb-2 tracking-tight">Pipeline Name</label>
                                <input
                                    {...form.register('name')}
                                    placeholder="e.g. Sync Users Daily"
                                    className="w-full px-4 py-3 rounded-lg bg-black/30 border border-white/[0.06] text-sm text-white
                                        placeholder:text-zinc-600 focus:border-white/[0.15] focus:outline-none focus:ring-1 focus:ring-white/[0.08]
                                        transition-all duration-200"
                                />
                                {form.formState.errors.name && (
                                    <p className="text-[11px] text-red-400 mt-1">{form.formState.errors.name.message}</p>
                                )}
                            </div>

                            <div>
                                <label className="block text-xs font-semibold text-zinc-300 mb-2 tracking-tight">Description</label>
                                <textarea
                                    {...form.register('description')}
                                    placeholder="Optional description of what this pipeline does..."
                                    rows={2}
                                    className="w-full px-4 py-3 rounded-lg bg-black/30 border border-white/[0.06] text-sm text-white resize-none
                                        placeholder:text-zinc-600 focus:border-white/[0.15] focus:outline-none focus:ring-1 focus:ring-white/[0.08]
                                        transition-all duration-200"
                                />
                            </div>

                            {/* Mode Selector */}
                            <div>
                                <label className="block text-xs font-semibold text-zinc-300 mb-2 tracking-tight">Processing Mode</label>
                                <Controller
                                    control={form.control}
                                    name="mode"
                                    render={({ field }) => (
                                        <div className="grid grid-cols-2 gap-3">
                                            {MODE_OPTIONS.map(opt => (
                                                <button
                                                    key={opt.value}
                                                    type="button"
                                                    onClick={() => field.onChange(opt.value)}
                                                    className={`p-4 rounded-lg border text-left transition-all duration-200
                                                        ${field.value === opt.value
                                                            ? 'border-violet-500/40 bg-violet-500/5 ring-1 ring-violet-500/20'
                                                            : 'border-white/[0.06] bg-black/20 hover:border-white/[0.12]'
                                                        }`}
                                                >
                                                    <p className={`text-xs font-semibold mb-1 ${field.value === opt.value ? 'text-violet-300' : 'text-zinc-300'}`}>
                                                        {opt.label}
                                                    </p>
                                                    <p className="text-[10px] text-zinc-500 leading-relaxed">{opt.description}</p>
                                                </button>
                                            ))}
                                        </div>
                                    )}
                                />
                            </div>

                            {/* Row Limit */}
                            <div>
                                <label className="block text-xs font-semibold text-zinc-300 mb-2 tracking-tight">Row Limit</label>
                                <input
                                    type="number"
                                    {...form.register('rowLimit', { valueAsNumber: true })}
                                    className="w-full px-4 py-3 rounded-lg bg-black/30 border border-white/[0.06] text-sm text-white
                                        placeholder:text-zinc-600 focus:border-white/[0.15] focus:outline-none focus:ring-1 focus:ring-white/[0.08]
                                        transition-all duration-200"
                                />
                                <p className="text-[10px] text-zinc-600 mt-1">Maximum rows per execution (safety limit)</p>
                            </div>
                        </div>
                    </div>
                )}

                {/* Section 1: Source Configuration */}
                {activeSection === 1 && (
                    <div className="space-y-4">
                        {/* Source Type Selector */}
                        <div className="rounded-xl border border-white/[0.06] bg-zinc-900/50 p-6">
                            <label className="block text-xs font-semibold text-zinc-300 mb-3 tracking-tight">Source Type</label>
                            <Controller
                                control={form.control}
                                name="sourceType"
                                render={({ field }) => (
                                    <div className="grid grid-cols-2 gap-3">
                                        {SOURCE_TYPES.map(src => (
                                            <button
                                                key={src.value}
                                                type="button"
                                                onClick={() => field.onChange(src.value)}
                                                className={`flex items-center gap-3 p-4 rounded-lg border text-left transition-all duration-200
                                                    ${field.value === src.value
                                                        ? 'border-violet-500/40 bg-violet-500/5 ring-1 ring-violet-500/20'
                                                        : 'border-white/[0.06] bg-black/20 hover:border-white/[0.12]'
                                                    }`}
                                            >
                                                <span className="text-xl">{src.emoji}</span>
                                                <div>
                                                    <p className={`text-xs font-semibold ${field.value === src.value ? 'text-violet-300' : 'text-zinc-300'}`}>
                                                        {src.label}
                                                    </p>
                                                    <p className="text-[10px] text-zinc-600">{src.description}</p>
                                                </div>
                                            </button>
                                        ))}
                                    </div>
                                )}
                            />
                        </div>

                        {/* Connection Selector (for DB sources) */}
                        {(sourceType === 'POSTGRES' || sourceType === 'MYSQL') && (
                            <div className="rounded-xl border border-white/[0.06] bg-zinc-900/50 p-6 space-y-5">
                                <div>
                                    <label className="block text-xs font-semibold text-zinc-300 mb-2 tracking-tight">Connection</label>
                                    {filteredConnections.length > 0 ? (
                                        <Controller
                                            control={form.control}
                                            name="connectionId"
                                            render={({ field }) => (
                                                <div className="space-y-2">
                                                    {filteredConnections.map(conn => (
                                                        <button
                                                            key={conn.id}
                                                            type="button"
                                                            onClick={() => field.onChange(conn.id)}
                                                            className={`w-full flex items-center gap-3 p-3 rounded-lg border text-left transition-all duration-200
                                                                ${field.value === conn.id
                                                                    ? 'border-emerald-500/40 bg-emerald-500/5'
                                                                    : 'border-white/[0.06] bg-black/20 hover:border-white/[0.12]'
                                                                }`}
                                                        >
                                                            <Database className={`w-4 h-4 flex-shrink-0 ${field.value === conn.id ? 'text-emerald-400' : 'text-zinc-500'}`} />
                                                            <div className="min-w-0 flex-1">
                                                                <p className="text-xs font-medium text-zinc-200 truncate">{conn.name}</p>
                                                                <p className="text-[10px] text-zinc-600 truncate">{conn.host} / {conn.databaseName}</p>
                                                            </div>
                                                        </button>
                                                    ))}
                                                </div>
                                            )}
                                        />
                                    ) : (
                                        <div className="flex items-center gap-2 p-3 rounded-lg bg-amber-500/5 border border-amber-500/20 text-[11px] text-amber-400">
                                            <AlertTriangle className="w-3.5 h-3.5 flex-shrink-0" />
                                            No {sourceType} connections found. Create one in Connections first.
                                        </div>
                                    )}
                                </div>

                                {/* SQL Query Editor */}
                                <div>
                                    <label className="block text-xs font-semibold text-zinc-300 mb-2 tracking-tight">Source Query (SQL)</label>
                                    <textarea
                                        {...form.register('sourceQuery')}
                                        placeholder="SELECT * FROM public.users WHERE active = true"
                                        rows={4}
                                        className="w-full px-4 py-3 rounded-lg bg-black/40 border border-white/[0.06] text-xs text-white font-mono resize-y
                                            placeholder:text-zinc-700 focus:border-white/[0.15] focus:outline-none focus:ring-1 focus:ring-white/[0.08]
                                            transition-all duration-200 leading-relaxed"
                                    />
                                    <p className="text-[10px] text-zinc-600 mt-1">
                                        SQL query to execute against the source. Leave empty to extract full table.
                                    </p>
                                </div>
                            </div>
                        )}

                        {/* REST API Config */}
                        {sourceType === 'REST_API' && (
                            <div className="rounded-xl border border-white/[0.06] bg-zinc-900/50 p-6">
                                <label className="block text-xs font-semibold text-zinc-300 mb-2 tracking-tight">API Endpoint</label>
                                <input
                                    type="text"
                                    placeholder="https://api.example.com/users"
                                    onChange={(e) => form.setValue('sourceConfig', JSON.stringify({ endpoint: e.target.value }))}
                                    className="w-full px-4 py-3 rounded-lg bg-black/30 border border-white/[0.06] text-sm text-white font-mono
                                        placeholder:text-zinc-600 focus:border-white/[0.15] focus:outline-none focus:ring-1 focus:ring-white/[0.08]
                                        transition-all duration-200"
                                />
                            </div>
                        )}

                        {/* Destination Config */}
                        <div className="rounded-xl border border-white/[0.06] bg-zinc-900/50 p-6">
                            <label className="block text-xs font-semibold text-zinc-300 mb-2 tracking-tight">Destination</label>
                            <Controller
                                control={form.control}
                                name="destinationType"
                                render={({ field }) => (
                                    <div className="grid grid-cols-2 gap-3">
                                        <button
                                            type="button"
                                            onClick={() => field.onChange('INTERNAL_RAW')}
                                            className={`p-4 rounded-lg border text-left transition-all duration-200
                                                ${field.value === 'INTERNAL_RAW'
                                                    ? 'border-emerald-500/40 bg-emerald-500/5 ring-1 ring-emerald-500/20'
                                                    : 'border-white/[0.06] bg-black/20 hover:border-white/[0.12]'
                                                }`}
                                        >
                                            <p className={`text-xs font-semibold mb-1 ${field.value === 'INTERNAL_RAW' ? 'text-emerald-300' : 'text-zinc-300'}`}>
                                                Internal Storage
                                            </p>
                                            <p className="text-[10px] text-zinc-500">Store in local data warehouse</p>
                                        </button>
                                        <button
                                            type="button"
                                            onClick={() => field.onChange('EXTERNAL')}
                                            className={`p-4 rounded-lg border text-left transition-all duration-200
                                                ${field.value === 'EXTERNAL'
                                                    ? 'border-emerald-500/40 bg-emerald-500/5 ring-1 ring-emerald-500/20'
                                                    : 'border-white/[0.06] bg-black/20 hover:border-white/[0.12]'
                                                }`}
                                        >
                                            <p className={`text-xs font-semibold mb-1 ${field.value === 'EXTERNAL' ? 'text-emerald-300' : 'text-zinc-300'}`}>
                                                External Database
                                            </p>
                                            <p className="text-[10px] text-zinc-500">Push to external DB via connection</p>
                                        </button>
                                    </div>
                                )}
                            />
                        </div>
                    </div>
                )}

                {/* Section 2: Transformation */}
                {activeSection === 2 && (
                    <div className="rounded-xl border border-white/[0.06] bg-zinc-900/50 p-6">
                        <div className="flex items-center gap-2 mb-4">
                            <Zap className="w-4 h-4 text-amber-400" />
                            <span className="text-xs font-semibold text-zinc-300">Transform Pipeline</span>
                            {mode === 'ELT' && (
                                <span className="ml-auto text-[10px] px-2 py-0.5 rounded-full bg-amber-500/10 text-amber-400 border border-amber-500/20">
                                    ELT: Post-load SQL transforms
                                </span>
                            )}
                        </div>
                        <TransformStepBuilder steps={transformSteps} onChange={setTransformSteps} />
                    </div>
                )}

                {/* Section 3: Schedule */}
                {activeSection === 3 && (
                    <div className="rounded-xl border border-white/[0.06] bg-zinc-900/50 p-6">
                        <Controller
                            control={form.control}
                            name="scheduleCron"
                            render={({ field }) => (
                                <SchedulePicker
                                    value={field.value || null}
                                    onChange={(cron) => field.onChange(cron)}
                                />
                            )}
                        />
                    </div>
                )}
            </form>
        </div>
    );
}
