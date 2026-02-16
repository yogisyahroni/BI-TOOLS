'use client';

import React, { useState } from 'react';
import type { TransformStep, TransformStepType } from '@/lib/types/batch2';
import {
    Filter,
    ArrowRightLeft,
    Type,
    Layers,
    BarChart3,
    Plus,
    X,
    GripVertical,
    ChevronDown,
} from 'lucide-react';

interface TransformStepBuilderProps {
    steps: TransformStep[];
    onChange: (steps: TransformStep[]) => void;
}

const STEP_TYPES: { type: TransformStepType; label: string; description: string; icon: React.ReactNode; color: string }[] = [
    {
        type: 'FILTER',
        label: 'Filter',
        description: 'Filter rows based on conditions',
        icon: <Filter className="w-4 h-4" />,
        color: 'text-blue-400 bg-blue-500/10 border-blue-500/20',
    },
    {
        type: 'RENAME',
        label: 'Rename Column',
        description: 'Rename a column to a new name',
        icon: <ArrowRightLeft className="w-4 h-4" />,
        color: 'text-purple-400 bg-purple-500/10 border-purple-500/20',
    },
    {
        type: 'CAST',
        label: 'Cast Type',
        description: 'Change column data type',
        icon: <Type className="w-4 h-4" />,
        color: 'text-amber-400 bg-amber-500/10 border-amber-500/20',
    },
    {
        type: 'DEDUPLICATE',
        label: 'Deduplicate',
        description: 'Remove duplicate rows by columns',
        icon: <Layers className="w-4 h-4" />,
        color: 'text-emerald-400 bg-emerald-500/10 border-emerald-500/20',
    },
    {
        type: 'AGGREGATE',
        label: 'Aggregate',
        description: 'Group and aggregate data',
        icon: <BarChart3 className="w-4 h-4" />,
        color: 'text-cyan-400 bg-cyan-500/10 border-cyan-500/20',
    },
];

function getStepMeta(type: TransformStepType) {
    return STEP_TYPES.find(s => s.type === type) || STEP_TYPES[0];
}

interface StepConfigEditorProps {
    step: TransformStep;
    onUpdate: (config: Record<string, string>) => void;
}

function StepConfigEditor({ step, onUpdate }: StepConfigEditorProps) {
    const configFields: Record<TransformStepType, { key: string; label: string; placeholder: string }[]> = {
        FILTER: [
            { key: 'column', label: 'Column', placeholder: 'e.g. status' },
            { key: 'operator', label: 'Operator', placeholder: 'e.g. =, !=, >, <, LIKE' },
            { key: 'value', label: 'Value', placeholder: 'e.g. active' },
        ],
        RENAME: [
            { key: 'from', label: 'From Column', placeholder: 'e.g. old_name' },
            { key: 'to', label: 'To Column', placeholder: 'e.g. new_name' },
        ],
        CAST: [
            { key: 'column', label: 'Column', placeholder: 'e.g. price' },
            { key: 'to_type', label: 'Target Type', placeholder: 'e.g. FLOAT, INTEGER, TEXT' },
        ],
        DEDUPLICATE: [
            { key: 'columns', label: 'Columns (comma-separated)', placeholder: 'e.g. email, phone' },
        ],
        AGGREGATE: [
            { key: 'group_by', label: 'Group By', placeholder: 'e.g. category, region' },
            { key: 'function', label: 'Function', placeholder: 'e.g. SUM, AVG, COUNT, MIN, MAX' },
            { key: 'column', label: 'Column', placeholder: 'e.g. revenue' },
        ],
    };

    const fields = configFields[step.type] || [];

    return (
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 mt-3">
            {fields.map(field => (
                <div key={field.key}>
                    <label className="block text-[10px] font-medium text-zinc-500 mb-1 uppercase tracking-wider">
                        {field.label}
                    </label>
                    <input
                        type="text"
                        value={step.config[field.key] || ''}
                        onChange={(e) => {
                            onUpdate({ ...step.config, [field.key]: e.target.value });
                        }}
                        placeholder={field.placeholder}
                        className="w-full px-3 py-1.5 rounded-md bg-black/30 border border-white/[0.06] text-xs text-white 
                            placeholder:text-zinc-600 focus:border-white/[0.15] focus:outline-none focus:ring-1 focus:ring-white/[0.08]
                            transition-all duration-200"
                    />
                </div>
            ))}
        </div>
    );
}

export function TransformStepBuilder({ steps, onChange }: TransformStepBuilderProps) {
    const [showPicker, setShowPicker] = useState(false);

    const addStep = (type: TransformStepType) => {
        const newStep: TransformStep = {
            type,
            config: {},
            order: steps.length + 1,
        };
        onChange([...steps, newStep]);
        setShowPicker(false);
    };

    const removeStep = (index: number) => {
        const updated = steps.filter((_, i) => i !== index).map((s, i) => ({ ...s, order: i + 1 }));
        onChange(updated);
    };

    const updateStepConfig = (index: number, config: Record<string, string>) => {
        const updated = steps.map((s, i) => (i === index ? { ...s, config } : s));
        onChange(updated);
    };

    const moveStep = (fromIndex: number, direction: 'up' | 'down') => {
        const toIndex = direction === 'up' ? fromIndex - 1 : fromIndex + 1;
        if (toIndex < 0 || toIndex >= steps.length) return;
        const updated = [...steps];
        [updated[fromIndex], updated[toIndex]] = [updated[toIndex], updated[fromIndex]];
        onChange(updated.map((s, i) => ({ ...s, order: i + 1 })));
    };

    return (
        <div className="space-y-3">
            <div className="flex items-center justify-between">
                <label className="text-xs font-semibold text-zinc-300 tracking-tight">
                    Transformation Steps
                </label>
                <span className="text-[10px] text-zinc-600">{steps.length} step{steps.length !== 1 ? 's' : ''}</span>
            </div>

            {/* Steps List */}
            {steps.length > 0 && (
                <div className="space-y-2">
                    {steps.map((step, index) => {
                        const meta = getStepMeta(step.type);
                        return (
                            <div
                                key={index}
                                className="rounded-lg border border-white/[0.06] bg-white/[0.02] overflow-hidden 
                                    transition-all duration-200 hover:border-white/[0.1]"
                            >
                                <div className="flex items-center gap-2 px-3 py-2.5">
                                    <GripVertical className="w-3.5 h-3.5 text-zinc-700 cursor-grab flex-shrink-0" />
                                    <span className={`inline-flex items-center gap-1.5 px-2 py-0.5 rounded-md text-[11px] font-medium border ${meta.color}`}>
                                        {meta.icon}
                                        {meta.label}
                                    </span>
                                    <span className="text-[10px] text-zinc-600 flex-1">Step {index + 1}</span>
                                    <div className="flex items-center gap-0.5">
                                        {index > 0 && (
                                            <button
                                                onClick={() => moveStep(index, 'up')}
                                                className="p-1 rounded hover:bg-white/[0.06] text-zinc-600 hover:text-zinc-400 transition-colors text-[10px]"
                                            >
                                                ↑
                                            </button>
                                        )}
                                        {index < steps.length - 1 && (
                                            <button
                                                onClick={() => moveStep(index, 'down')}
                                                className="p-1 rounded hover:bg-white/[0.06] text-zinc-600 hover:text-zinc-400 transition-colors text-[10px]"
                                            >
                                                ↓
                                            </button>
                                        )}
                                        <button
                                            onClick={() => removeStep(index)}
                                            className="p-1 rounded hover:bg-red-500/10 text-zinc-600 hover:text-red-400 transition-colors"
                                        >
                                            <X className="w-3 h-3" />
                                        </button>
                                    </div>
                                </div>
                                <div className="px-3 pb-3">
                                    <StepConfigEditor step={step} onUpdate={(config) => updateStepConfig(index, config)} />
                                </div>
                            </div>
                        );
                    })}
                </div>
            )}

            {/* Add Step */}
            <div className="relative">
                <button
                    onClick={() => setShowPicker(!showPicker)}
                    className="w-full flex items-center justify-center gap-2 px-4 py-2.5 rounded-lg 
                        border border-dashed border-white/[0.08] text-xs text-zinc-500 
                        hover:border-white/[0.15] hover:text-zinc-300 hover:bg-white/[0.02]
                        active:scale-[0.99] transition-all duration-200"
                >
                    <Plus className="w-3.5 h-3.5" />
                    Add Transform Step
                    <ChevronDown className={`w-3 h-3 transition-transform duration-200 ${showPicker ? 'rotate-180' : ''}`} />
                </button>

                {showPicker && (
                    <div className="absolute left-0 right-0 top-full mt-2 rounded-lg bg-zinc-900 border border-white/[0.08] shadow-2xl z-20 p-2
                        animate-in fade-in-0 slide-in-from-top-2 duration-200">
                        {STEP_TYPES.map(stepType => (
                            <button
                                key={stepType.type}
                                onClick={() => addStep(stepType.type)}
                                className="w-full flex items-center gap-3 px-3 py-2.5 rounded-md hover:bg-white/[0.06] transition-colors text-left"
                            >
                                <span className={`inline-flex items-center justify-center w-8 h-8 rounded-lg border ${stepType.color}`}>
                                    {stepType.icon}
                                </span>
                                <div>
                                    <p className="text-xs font-medium text-zinc-200">{stepType.label}</p>
                                    <p className="text-[10px] text-zinc-500">{stepType.description}</p>
                                </div>
                            </button>
                        ))}
                    </div>
                )}
            </div>
        </div>
    );
}
