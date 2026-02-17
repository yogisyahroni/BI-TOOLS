'use client';

import React, { useEffect, useRef, useState, useCallback } from 'react';
import type { Pipeline, SSEProgressEvent } from '@/lib/types/batch2';
import { pipelineApi } from '@/lib/api/pipelines';
import {
    Play,
    _Pause,
    Trash2,
    Settings,
    Clock,
    CheckCircle2,
    XCircle,
    Loader2,
    Database,
    ArrowRight,
    MoreVertical,
    Zap,
    Activity,
    ChevronRight,
} from 'lucide-react';

interface PipelineCardProps {
    pipeline: Pipeline;
    onEdit: (pipeline: Pipeline) => void;
    onDelete: (id: string) => void;
    onRun: (id: string) => void;
    onViewHistory: (pipeline: Pipeline) => void;
}

const STATUS_CONFIG: Record<string, { color: string; bgColor: string; icon: React.ReactNode; label: string }> = {
    SUCCESS: {
        color: 'text-emerald-400',
        bgColor: 'bg-emerald-500/10 border-emerald-500/20',
        icon: <CheckCircle2 className="w-3.5 h-3.5" />,
        label: 'Healthy',
    },
    FAILED: {
        color: 'text-red-400',
        bgColor: 'bg-red-500/10 border-red-500/20',
        icon: <XCircle className="w-3.5 h-3.5" />,
        label: 'Failed',
    },
    PROCESSING: {
        color: 'text-blue-400',
        bgColor: 'bg-blue-500/10 border-blue-500/20',
        icon: <Loader2 className="w-3.5 h-3.5 animate-spin" />,
        label: 'Running',
    },
    IDLE: {
        color: 'text-zinc-500',
        bgColor: 'bg-zinc-500/10 border-zinc-500/20',
        icon: <Clock className="w-3.5 h-3.5" />,
        label: 'Idle',
    },
};

const SOURCE_ICONS: Record<string, string> = {
    POSTGRES: '🐘',
    MYSQL: '🐬',
    CSV: '📄',
    REST_API: '🌐',
};

function getStatusConfig(status: string | null) {
    if (!status) return STATUS_CONFIG.IDLE;
    return STATUS_CONFIG[status] || STATUS_CONFIG.IDLE;
}

function formatRelativeTime(dateStr: string | null): string {
    if (!dateStr) return 'Never';
    const date = new Date(dateStr);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMinutes = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);

    if (diffMinutes < 1) return 'Just now';
    if (diffMinutes < 60) return `${diffMinutes}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    if (diffDays < 7) return `${diffDays}d ago`;
    return date.toLocaleDateString();
}

function formatCron(cron: string | null): string {
    if (!cron) return 'Manual';
    const presets: Record<string, string> = {
        '*/15 * * * *': 'Every 15m',
        '*/30 * * * *': 'Every 30m',
        '0 * * * *': 'Hourly',
        '0 */6 * * *': 'Every 6h',
        '0 */12 * * *': 'Every 12h',
        '0 0 * * *': 'Daily',
        '0 6 * * *': 'Daily 6AM',
        '0 0 * * 1': 'Weekly',
        '0 0 1 * *': 'Monthly',
    };
    return presets[cron] || cron;
}

export function PipelineCard({ pipeline, onEdit, onDelete, onRun, onViewHistory }: PipelineCardProps) {
    const [isRunning, setIsRunning] = useState(false);
    const [progress, setProgress] = useState(0);
    const [menuOpen, setMenuOpen] = useState(false);
    const menuRef = useRef<HTMLDivElement>(null);
    const eventSourceRef = useRef<EventSource | null>(null);

    const statusCfg = getStatusConfig(pipeline.lastStatus);
    const sourceEmoji = SOURCE_ICONS[pipeline.sourceType] || '📊';

    // Close menu on outside click
    useEffect(() => {
        function handleClick(e: MouseEvent) {
            if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
                setMenuOpen(false);
            }
        }
        document.addEventListener('mousedown', handleClick);
        return () => document.removeEventListener('mousedown', handleClick);
    }, []);

    // Cleanup SSE on unmount
    useEffect(() => {
        return () => {
            if (eventSourceRef.current) {
                eventSourceRef.current.close();
            }
        };
    }, []);

    const handleRun = useCallback(async () => {
        setIsRunning(true);
        setProgress(0);
        try {
            const execution = await pipelineApi.run(pipeline.id);

            // Start SSE stream
            eventSourceRef.current = pipelineApi.streamExecution(
                pipeline.id,
                execution.id,
                (event: SSEProgressEvent) => {
                    setProgress(event.progress);
                },
                () => {
                    setIsRunning(false);
                    setProgress(100);
                    eventSourceRef.current = null;
                    onRun(pipeline.id);
                },
                () => {
                    setIsRunning(false);
                    eventSourceRef.current = null;
                }
            );
        } catch {
            setIsRunning(false);
        }
    }, [pipeline.id, onRun]);

    return (
        <div
            className="group relative rounded-xl border border-white/[0.06] bg-gradient-to-br from-zinc-900/80 to-zinc-950/90 
            backdrop-blur-xl shadow-lg hover:shadow-xl hover:border-white/[0.12] 
            transition-all duration-300 ease-out hover:translate-y-[-2px] overflow-hidden"
        >
            {/* Progress bar overlay */}
            {isRunning && (
                <div className="absolute top-0 left-0 right-0 h-0.5 bg-zinc-800 overflow-hidden z-10">
                    <div
                        className="h-full bg-gradient-to-r from-blue-500 to-cyan-400 transition-all duration-500 ease-out"
                        style={{ width: `${progress}%` }}
                    />
                </div>
            )}

            {/* Header */}
            <div className="p-5 pb-3">
                <div className="flex items-start justify-between gap-3">
                    <div className="flex items-center gap-3 min-w-0">
                        <div className="flex-shrink-0 w-9 h-9 rounded-lg bg-white/[0.04] border border-white/[0.06] flex items-center justify-center text-lg">
                            {sourceEmoji}
                        </div>
                        <div className="min-w-0">
                            <h3 className="text-sm font-semibold text-white truncate tracking-tight">
                                {pipeline.name}
                            </h3>
                            <p className="text-[11px] text-zinc-500 mt-0.5 truncate">
                                {pipeline.description || `${pipeline.sourceType} → ${pipeline.destinationType}`}
                            </p>
                        </div>
                    </div>
                    <div className="flex items-center gap-1.5 flex-shrink-0">
                        {/* Status Badge */}
                        <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-medium border ${statusCfg.bgColor} ${statusCfg.color}`}>
                            {statusCfg.icon}
                            {statusCfg.label}
                        </span>
                        {/* Menu */}
                        <div className="relative" ref={menuRef}>
                            <button
                                onClick={() => setMenuOpen(!menuOpen)}
                                className="p-1 rounded-md hover:bg-white/[0.06] text-zinc-500 hover:text-zinc-300 transition-colors"
                            >
                                <MoreVertical className="w-4 h-4" />
                            </button>
                            {menuOpen && (
                                <div className="absolute right-0 top-7 w-40 rounded-lg bg-zinc-900 border border-white/[0.08] shadow-2xl z-50 py-1
                                    animate-in fade-in-0 zoom-in-95 duration-150">
                                    <button
                                        onClick={() => { setMenuOpen(false); onEdit(pipeline); }}
                                        className="w-full flex items-center gap-2 px-3 py-2 text-xs text-zinc-300 hover:bg-white/[0.06] transition-colors"
                                    >
                                        <Settings className="w-3.5 h-3.5" /> Edit Pipeline
                                    </button>
                                    <button
                                        onClick={() => { setMenuOpen(false); onViewHistory(pipeline); }}
                                        className="w-full flex items-center gap-2 px-3 py-2 text-xs text-zinc-300 hover:bg-white/[0.06] transition-colors"
                                    >
                                        <Activity className="w-3.5 h-3.5" /> View History
                                    </button>
                                    <div className="border-t border-white/[0.06] my-1" />
                                    <button
                                        onClick={() => { setMenuOpen(false); onDelete(pipeline.id); }}
                                        className="w-full flex items-center gap-2 px-3 py-2 text-xs text-red-400 hover:bg-red-500/10 transition-colors"
                                    >
                                        <Trash2 className="w-3.5 h-3.5" /> Delete
                                    </button>
                                </div>
                            )}
                        </div>
                    </div>
                </div>
            </div>

            {/* Pipeline Flow Visualization */}
            <div className="px-5 py-2.5">
                <div className="flex items-center gap-2 text-[11px]">
                    <span className="flex items-center gap-1 px-2 py-1 rounded-md bg-white/[0.03] border border-white/[0.06] text-zinc-400">
                        <Database className="w-3 h-3" />
                        {pipeline.sourceType}
                    </span>
                    <ArrowRight className="w-3 h-3 text-zinc-600" />
                    <span className="flex items-center gap-1 px-2 py-1 rounded-md bg-white/[0.03] border border-white/[0.06] text-zinc-400">
                        <Zap className="w-3 h-3" />
                        {pipeline.mode}
                    </span>
                    <ArrowRight className="w-3 h-3 text-zinc-600" />
                    <span className="flex items-center gap-1 px-2 py-1 rounded-md bg-white/[0.03] border border-white/[0.06] text-zinc-400">
                        <Database className="w-3 h-3" />
                        {pipeline.destinationType === 'INTERNAL_RAW' ? 'Internal' : pipeline.destinationType}
                    </span>
                </div>
            </div>

            {/* Footer */}
            <div className="px-5 py-3 border-t border-white/[0.04] flex items-center justify-between">
                <div className="flex items-center gap-3 text-[11px] text-zinc-500">
                    <span className="flex items-center gap-1">
                        <Clock className="w-3 h-3" />
                        {formatRelativeTime(pipeline.lastRunAt)}
                    </span>
                    <span className="text-zinc-700">•</span>
                    <span>{formatCron(pipeline.scheduleCron)}</span>
                </div>
                <div className="flex items-center gap-1.5">
                    {isRunning ? (
                        <button
                            disabled
                            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-[11px] font-medium 
                                bg-blue-500/10 text-blue-400 border border-blue-500/20 cursor-not-allowed"
                        >
                            <Loader2 className="w-3 h-3 animate-spin" />
                            {progress}%
                        </button>
                    ) : (
                        <button
                            onClick={handleRun}
                            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-[11px] font-medium 
                                bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 
                                hover:bg-emerald-500/20 active:scale-[0.97] transition-all duration-200"
                        >
                            <Play className="w-3 h-3" />
                            Run
                        </button>
                    )}
                    <button
                        onClick={() => onViewHistory(pipeline)}
                        className="p-1.5 rounded-lg text-zinc-500 hover:text-zinc-300 hover:bg-white/[0.06] transition-all duration-200"
                    >
                        <ChevronRight className="w-4 h-4" />
                    </button>
                </div>
            </div>
        </div>
    );
}
