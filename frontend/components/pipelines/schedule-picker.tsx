'use client';

import React, { useState } from 'react';
import { SCHEDULE_PRESETS } from '@/lib/types/batch2';
import {
    Clock,
    ChevronDown,
    Check,
    RefreshCcw,
    Calendar,
} from 'lucide-react';

interface SchedulePickerProps {
    value: string | null;
    onChange: (cron: string | null) => void;
}

export function SchedulePicker({ value, onChange }: SchedulePickerProps) {
    const [isOpen, setIsOpen] = useState(false);
    const [customCron, setCustomCron] = useState('');

    const selectedPreset = SCHEDULE_PRESETS.find(p => p.cron === value);
    const isCustom = value && !selectedPreset;

    const displayLabel = selectedPreset
        ? selectedPreset.label
        : isCustom
            ? `Custom: ${value}`
            : 'No schedule (manual)';

    const handleSelect = (cron: string) => {
        if (cron === '') {
            // Custom — don't close yet
            return;
        }
        onChange(cron);
        setIsOpen(false);
    };

    const handleClear = () => {
        onChange(null);
        setIsOpen(false);
    };

    const handleCustomSubmit = () => {
        if (customCron.trim()) {
            onChange(customCron.trim());
            setIsOpen(false);
            setCustomCron('');
        }
    };

    return (
        <div className="space-y-1.5">
            <label className="block text-xs font-semibold text-zinc-300 tracking-tight">
                Schedule
            </label>
            <div className="relative">
                <button
                    type="button"
                    onClick={() => setIsOpen(!isOpen)}
                    className="w-full flex items-center justify-between gap-2 px-3 py-2.5 rounded-lg 
                        bg-black/30 border border-white/[0.06] text-left
                        hover:border-white/[0.12] focus:border-white/[0.15] focus:outline-none focus:ring-1 focus:ring-white/[0.08]
                        transition-all duration-200"
                >
                    <div className="flex items-center gap-2 min-w-0">
                        <Clock className="w-4 h-4 text-zinc-500 flex-shrink-0" />
                        <div className="min-w-0">
                            <span className="text-xs text-white truncate block">{displayLabel}</span>
                            {selectedPreset && selectedPreset.description && (
                                <span className="text-[10px] text-zinc-500 truncate block">{selectedPreset.description}</span>
                            )}
                        </div>
                    </div>
                    <ChevronDown className={`w-4 h-4 text-zinc-500 transition-transform duration-200 flex-shrink-0 ${isOpen ? 'rotate-180' : ''}`} />
                </button>

                {isOpen && (
                    <div className="absolute left-0 right-0 top-full mt-1.5 rounded-lg bg-zinc-900 border border-white/[0.08] shadow-2xl z-30 overflow-hidden
                        animate-in fade-in-0 slide-in-from-top-2 duration-200">

                        {/* Clear option */}
                        <button
                            onClick={handleClear}
                            className="w-full flex items-center gap-3 px-3 py-2.5 text-xs hover:bg-white/[0.06] transition-colors border-b border-white/[0.04]"
                        >
                            <RefreshCcw className="w-3.5 h-3.5 text-zinc-500" />
                            <span className="text-zinc-400">No schedule (manual only)</span>
                            {!value && <Check className="w-3.5 h-3.5 text-emerald-400 ml-auto" />}
                        </button>

                        {/* Presets */}
                        <div className="max-h-60 overflow-y-auto py-1">
                            {SCHEDULE_PRESETS.filter(p => p.cron !== '').map(preset => (
                                <button
                                    key={preset.cron}
                                    onClick={() => handleSelect(preset.cron)}
                                    className="w-full flex items-center gap-3 px-3 py-2.5 text-xs hover:bg-white/[0.06] transition-colors"
                                >
                                    <Calendar className="w-3.5 h-3.5 text-zinc-600 flex-shrink-0" />
                                    <div className="flex-1 text-left min-w-0">
                                        <span className="text-zinc-200 block truncate">{preset.label}</span>
                                        <span className="text-[10px] text-zinc-600 block font-mono">{preset.cron}</span>
                                    </div>
                                    {value === preset.cron && <Check className="w-3.5 h-3.5 text-emerald-400 flex-shrink-0" />}
                                </button>
                            ))}
                        </div>

                        {/* Custom cron input */}
                        <div className="border-t border-white/[0.04] p-3">
                            <p className="text-[10px] font-medium text-zinc-500 uppercase tracking-wider mb-2">Custom Cron Expression</p>
                            <div className="flex gap-2">
                                <input
                                    type="text"
                                    value={customCron}
                                    onChange={(e) => setCustomCron(e.target.value)}
                                    placeholder="e.g. 0 */4 * * *"
                                    className="flex-1 px-3 py-1.5 rounded-md bg-black/30 border border-white/[0.06] text-xs text-white font-mono
                                        placeholder:text-zinc-700 focus:border-white/[0.15] focus:outline-none focus:ring-1 focus:ring-white/[0.08]
                                        transition-all duration-200"
                                    onKeyDown={(e) => {
                                        if (e.key === 'Enter') {
                                            e.preventDefault();
                                            handleCustomSubmit();
                                        }
                                    }}
                                />
                                <button
                                    onClick={handleCustomSubmit}
                                    disabled={!customCron.trim()}
                                    className="px-3 py-1.5 rounded-md bg-white/[0.06] border border-white/[0.06] text-xs text-zinc-300
                                        hover:bg-white/[0.1] disabled:opacity-40 disabled:cursor-not-allowed
                                        transition-all duration-200"
                                >
                                    Apply
                                </button>
                            </div>
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
}
