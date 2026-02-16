'use client';

import React, { useState } from 'react';
import type { JobExecution, ExecutionLog } from '@/lib/types/batch2';
import { usePipelineExecutions } from '@/hooks/use-pipelines';

import {
    CheckCircle2,
    XCircle,
    Clock,
    Loader2,
    AlertTriangle,
    ChevronDown,
    ChevronUp,
    ArrowRight,
    Database,
    FileText,
    Activity,
    BarChart3,
    RefreshCcw,
} from 'lucide-react';

interface PipelineHistoryProps {
    pipelineId: string;
    pipelineName?: string;
}

const STATUS_STYLES: Record<string, { color: string; bg: string; icon: React.ReactNode }> = {
    COMPLETED: {
        color: 'text-emerald-400',
        bg: 'bg-emerald-500/10 border-emerald-500/20',
        icon: <CheckCircle2 className="w-4 h-4" />,
    },
    SUCCESS: {
        color: 'text-emerald-400',
        bg: 'bg-emerald-500/10 border-emerald-500/20',
        icon: <CheckCircle2 className="w-4 h-4" />,
    },
    FAILED: {
        color: 'text-red-400',
        bg: 'bg-red-500/10 border-red-500/20',
        icon: <XCircle className="w-4 h-4" />,
    },
    PROCESSING: {
        color: 'text-blue-400',
        bg: 'bg-blue-500/10 border-blue-500/20',
        icon: <Loader2 className="w-4 h-4 animate-spin" />,
    },
    EXTRACTING: {
        color: 'text-cyan-400',
        bg: 'bg-cyan-500/10 border-cyan-500/20',
        icon: <Database className="w-4 h-4 animate-pulse" />,
    },
    TRANSFORMING: {
        color: 'text-amber-400',
        bg: 'bg-amber-500/10 border-amber-500/20',
        icon: <Activity className="w-4 h-4 animate-pulse" />,
    },
    LOADING: {
        color: 'text-violet-400',
        bg: 'bg-violet-500/10 border-violet-500/20',
        icon: <ArrowRight className="w-4 h-4 animate-pulse" />,
    },
    PENDING: {
        color: 'text-zinc-400',
        bg: 'bg-zinc-500/10 border-zinc-500/20',
        icon: <Clock className="w-4 h-4" />,
    },
};

function getStatusStyle(status: string) {
    return STATUS_STYLES[status] || STATUS_STYLES.PENDING;
}

function formatDuration(ms: number | null | undefined): string {
    if (!ms) return '—';
    if (ms < 1000) return `${ms}ms`;
    const seconds = Math.floor(ms / 1000);
    if (seconds < 60) return `${seconds}s`;
    const minutes = Math.floor(seconds / 60);
    const remainingSeconds = seconds % 60;
    return `${minutes}m ${remainingSeconds}s`;
}

function formatBytes(bytes: number | null | undefined): string {
    if (!bytes) return '—';
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

function formatTimestamp(dateStr: string): string {
    const date = new Date(dateStr);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffHours = diffMs / 3600000;

    if (diffHours < 24) {
        return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    }
    return date.toLocaleDateString([], { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
}

interface ExecutionRowProps {
    execution: JobExecution;
}

function ExecutionRow({ execution }: ExecutionRowProps) {
    const [expanded, setExpanded] = useState(false);
    const statusStyle = getStatusStyle(execution.status);

    let parsedLogs: ExecutionLog[] = [];
    if (execution.logs) {
        try {
            parsedLogs = typeof execution.logs === 'string' ? JSON.parse(execution.logs) : [];
        } catch {
            parsedLogs = [];
        }
    }

    const hasQualityIssues = (execution.qualityViolations || 0) > 0;

    return (
        <div className="rounded-lg border border-white/[0.06] bg-white/[0.02] overflow-hidden transition-all duration-200 hover:border-white/[0.1]">
            {/* Main Row */}
            <button
                onClick={() => setExpanded(!expanded)}
                className="w-full flex items-center gap-3 px-4 py-3 text-left"
            >
                {/* Status */}
                <span className={`inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-[10px] font-medium border ${statusStyle.bg} ${statusStyle.color}`}>
                    {statusStyle.icon}
                    {execution.status}
                </span>

                {/* Timestamp */}
                <span className="text-[11px] text-zinc-500 min-w-[80px]">
                    {formatTimestamp(execution.startedAt)}
                </span>

                {/* Metrics */}
                <div className="flex items-center gap-4 flex-1 text-[11px]">
                    <span className="flex items-center gap-1 text-zinc-400">
                        <BarChart3 className="w-3 h-3 text-zinc-600" />
                        {execution.rowsProcessed?.toLocaleString() || 0} rows
                    </span>
                    <span className="text-zinc-700">|</span>
                    <span className="text-zinc-500">{formatBytes(execution.bytesProcessed)}</span>
                    <span className="text-zinc-700">|</span>
                    <span className="text-zinc-500">{formatDuration(execution.durationMs)}</span>
                    {hasQualityIssues && (
                        <>
                            <span className="text-zinc-700">|</span>
                            <span className="flex items-center gap-1 text-amber-400">
                                <AlertTriangle className="w-3 h-3" />
                                {execution.qualityViolations} violations
                            </span>
                        </>
                    )}
                </div>

                {/* Progress */}
                {execution.status === 'PROCESSING' || execution.status === 'EXTRACTING' || execution.status === 'TRANSFORMING' || execution.status === 'LOADING' ? (
                    <div className="flex items-center gap-2 min-w-[80px]">
                        <div className="flex-1 h-1.5 bg-zinc-800 rounded-full overflow-hidden">
                            <div
                                className="h-full bg-gradient-to-r from-blue-500 to-cyan-400 rounded-full transition-all duration-500"
                                style={{ width: `${execution.progress || 0}%` }}
                            />
                        </div>
                        <span className="text-[10px] text-zinc-500 w-8 text-right">{execution.progress || 0}%</span>
                    </div>
                ) : null}

                {/* Expand Toggle */}
                <div className="text-zinc-600">
                    {expanded ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
                </div>
            </button>

            {/* Expanded Details */}
            {expanded && (
                <div className="border-t border-white/[0.04] px-4 py-3 space-y-3 bg-black/20">
                    {/* Error */}
                    {execution.error && (
                        <div className="flex items-start gap-2 p-3 rounded-md bg-red-500/5 border border-red-500/20">
                            <XCircle className="w-4 h-4 text-red-400 flex-shrink-0 mt-0.5" />
                            <div>
                                <p className="text-[10px] font-medium text-red-400 uppercase tracking-wider mb-1">Error</p>
                                <p className="text-xs text-red-300 font-mono whitespace-pre-wrap break-all">{execution.error}</p>
                            </div>
                        </div>
                    )}

                    {/* Execution Logs */}
                    {parsedLogs.length > 0 && (
                        <div>
                            <p className="text-[10px] font-medium text-zinc-500 uppercase tracking-wider mb-2">Execution Log</p>
                            <div className="rounded-md bg-black/40 border border-white/[0.04] overflow-hidden max-h-60 overflow-y-auto">
                                {parsedLogs.map((log, i) => (
                                    <div
                                        key={i}
                                        className="flex items-start gap-2 px-3 py-1.5 border-b border-white/[0.02] last:border-b-0 font-mono"
                                    >
                                        <span className={`text-[9px] font-bold uppercase w-12 flex-shrink-0 pt-0.5
                                            ${log.level === 'ERROR' ? 'text-red-400' : log.level === 'WARN' ? 'text-amber-400' : 'text-zinc-600'}`}>
                                            {log.level}
                                        </span>
                                        <span className="text-[10px] text-zinc-700 w-20 flex-shrink-0 pt-0.5">{log.phase}</span>
                                        <span className="text-[10px] text-zinc-400 flex-1">{log.message}</span>
                                    </div>
                                ))}
                            </div>
                        </div>
                    )}

                    {/* Metrics Grid */}
                    <div className="grid grid-cols-4 gap-3">
                        <div className="p-2.5 rounded-md bg-white/[0.02] border border-white/[0.04]">
                            <p className="text-[9px] text-zinc-600 uppercase tracking-wider mb-0.5">Rows</p>
                            <p className="text-sm font-semibold text-white">{(execution.rowsProcessed || 0).toLocaleString()}</p>
                        </div>
                        <div className="p-2.5 rounded-md bg-white/[0.02] border border-white/[0.04]">
                            <p className="text-[9px] text-zinc-600 uppercase tracking-wider mb-0.5">Data</p>
                            <p className="text-sm font-semibold text-white">{formatBytes(execution.bytesProcessed)}</p>
                        </div>
                        <div className="p-2.5 rounded-md bg-white/[0.02] border border-white/[0.04]">
                            <p className="text-[9px] text-zinc-600 uppercase tracking-wider mb-0.5">Duration</p>
                            <p className="text-sm font-semibold text-white">{formatDuration(execution.durationMs)}</p>
                        </div>
                        <div className="p-2.5 rounded-md bg-white/[0.02] border border-white/[0.04]">
                            <p className="text-[9px] text-zinc-600 uppercase tracking-wider mb-0.5">Quality</p>
                            <p className={`text-sm font-semibold ${hasQualityIssues ? 'text-amber-400' : 'text-emerald-400'}`}>
                                {execution.qualityViolations || 0} issues
                            </p>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}

export function PipelineHistory({ pipelineId, pipelineName }: PipelineHistoryProps) {
    const {
        data,
        isLoading,
        refetch: fetchExecutions
    } = usePipelineExecutions(pipelineId, { limit: 50 });

    const executions = data?.executions || [];
    const successRate = data?.successRate ?? null;


    if (isLoading) {
        return (
            <div className="space-y-3">
                {Array.from({ length: 5 }).map((_, i) => (
                    <div key={i} className="h-14 rounded-lg bg-zinc-900/50 border border-white/[0.06] animate-pulse" />
                ))}
            </div>
        );
    }

    return (
        <div className="space-y-4">
            {/* Header */}
            <div className="flex items-center justify-between">
                <div>
                    <h3 className="text-sm font-semibold text-white tracking-tight">
                        Execution History
                        {pipelineName && <span className="text-zinc-500 font-normal ml-2">— {pipelineName}</span>}
                    </h3>
                    {successRate !== null && (
                        <p className="text-[11px] text-zinc-500 mt-0.5">
                            Success rate: <span className={successRate >= 90 ? 'text-emerald-400' : successRate >= 50 ? 'text-amber-400' : 'text-red-400'}>
                                {successRate.toFixed(1)}%
                            </span>
                        </p>
                    )}
                </div>
                <button
                    onClick={() => fetchExecutions()}
                    className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-[11px] text-zinc-500
                        hover:text-zinc-300 hover:bg-white/[0.06] transition-all duration-200"
                >
                    <RefreshCcw className="w-3 h-3" />
                    Refresh
                </button>
            </div>

            {/* Executions List */}
            {executions.length === 0 ? (
                <div className="flex flex-col items-center justify-center py-16 rounded-xl border border-dashed border-white/[0.06]">
                    <FileText className="w-8 h-8 text-zinc-700 mb-3" />
                    <p className="text-xs text-zinc-500 mb-1">No execution history yet</p>
                    <p className="text-[10px] text-zinc-600">Run this pipeline to see results here.</p>
                </div>
            ) : (
                <div className="space-y-2">
                    {executions.map(execution => (
                        <ExecutionRow key={execution.id} execution={execution} />
                    ))}
                </div>
            )}
        </div>
    );
}
