'use client';

import React, { createContext, useContext, useMemo } from 'react';

/**
 * TASK-CHART-022: Advanced Tooltips
 * Reusable tooltip system with rich content, positioning, and theming.
 */

interface TooltipConfig {
    mode?: 'axis' | 'item' | 'cross';
    enterable?: boolean;
    confine?: boolean;
    showDelay?: number;
    hideDelay?: number;
    transitionDuration?: number;
    backgroundColor?: string;
    borderColor?: string;
    borderWidth?: number;
    borderRadius?: number;
    padding?: number | [number, number];
    textStyle?: {
        color?: string;
        fontSize?: number;
        fontFamily?: string;
    };
    extraCssText?: string;
}

const defaultTooltipConfig: TooltipConfig = {
    mode: 'item',
    enterable: false,
    confine: true,
    showDelay: 50,
    hideDelay: 100,
    transitionDuration: 0.15,
    backgroundColor: 'rgba(30,30,30,0.95)',
    borderColor: 'rgba(255,255,255,0.1)',
    borderWidth: 1,
    borderRadius: 8,
    padding: [10, 14],
    textStyle: {
        color: '#fff',
        fontSize: 12,
        fontFamily: 'Inter, system-ui, sans-serif',
    },
    extraCssText: 'box-shadow: 0 8px 24px rgba(0,0,0,0.25); backdrop-filter: blur(8px);',
};

const TooltipContext = createContext<TooltipConfig>(defaultTooltipConfig);

export function TooltipProvider({
    config,
    children,
}: {
    config?: Partial<TooltipConfig>;
    children: React.ReactNode;
}) {
    const merged = useMemo(
        () => ({ ...defaultTooltipConfig, ...config }),
        [config],
    );
    return (
        <TooltipContext.Provider value={merged}>
            {children}
        </TooltipContext.Provider>
    );
}

export function useAdvancedTooltip(overrides?: Partial<TooltipConfig>) {
    const ctx = useContext(TooltipContext);
    return useMemo(() => ({ ...ctx, ...overrides }), [ctx, overrides]);
}

/**
 * Build ECharts tooltip config from our TooltipConfig.
 */
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
export function buildEChartsTooltip(config: TooltipConfig, formatter?: string | ((params: any) => string)) {
    return {
        trigger: config.mode === 'cross' ? 'axis' : config.mode,
        enterable: config.enterable,
        confine: config.confine,
        showDelay: config.showDelay,
        hideDelay: config.hideDelay,
        transitionDuration: config.transitionDuration,
        backgroundColor: config.backgroundColor,
        borderColor: config.borderColor,
        borderWidth: config.borderWidth,
        borderRadius: config.borderRadius,
        padding: config.padding,
        textStyle: config.textStyle,
        extraCssText: config.extraCssText,
        formatter,
        ...(config.mode === 'cross'
            ? { axisPointer: { type: 'cross' as const, label: { backgroundColor: '#6a7985' } } }
            : config.mode === 'axis'
                ? { axisPointer: { type: 'shadow' as const } }
                : {}),
    };
}

/**
 * Rich tooltip formatter helpers
 */
export const tooltipFormatters = {
    /**
     * Currency formatter with locale
     */
    currency: (value: number, currency = 'USD', locale = 'en-US') =>
        new Intl.NumberFormat(locale, { style: 'currency', currency }).format(value),

    /**
     * Percentage from decimal
     */
    percent: (value: number, decimals = 1) => `${(value * 100).toFixed(decimals)}%`,

    /**
     * Compact number (1K, 1M, 1B)
     */
    compact: (value: number) => {
        if (Math.abs(value) >= 1e9) return `${(value / 1e9).toFixed(1)}B`;
        if (Math.abs(value) >= 1e6) return `${(value / 1e6).toFixed(1)}M`;
        if (Math.abs(value) >= 1e3) return `${(value / 1e3).toFixed(1)}K`;
        return value.toLocaleString();
    },

    /**
     * Builds a colored marker dot for tooltips
     */
    marker: (color: string) =>
        `<span style="display:inline-block;margin-right:4px;border-radius:50%;width:8px;height:8px;background:${color}"></span>`,

    /**
     * Multi-series tooltip with ranking
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
     */
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    rankedSeries: (params: any[], valueFormatter?: (v: number) => string) => {
        if (!Array.isArray(params) || params.length === 0) return '';
        const fmt = valueFormatter ?? ((v: number) => v.toLocaleString());
        const sorted = [...params].sort((a, b) => (b.value ?? 0) - (a.value ?? 0));
        let html = `<div style="font-weight:600;margin-bottom:6px">${params[0].axisValueLabel ?? ''}</div>`;
        sorted.forEach((p, i) => {
            const rank = i + 1;
            html += `<div style="display:flex;align-items:center;gap:4px;margin:2px 0">
                <span style="font-size:10px;color:#888;width:14px">${rank}.</span>
                ${p.marker ?? ''}
                <span style="flex:1">${p.seriesName}</span>
                <strong>${fmt(p.value ?? 0)}</strong>
            </div>`;
        });
        return html;
    },

    /**
     * Comparison tooltip: current vs. previous
     */
    comparison: (current: number, previous: number, label: string, valueFormatter?: (v: number) => string) => {
        const fmt = valueFormatter ?? ((v: number) => v.toLocaleString());
        const delta = previous !== 0 ? ((current - previous) / Math.abs(previous)) * 100 : 0;
        const deltaSign = delta >= 0 ? '+' : '';
        const deltaColor = delta >= 0 ? '#10b981' : '#ef4444';
        return `<div>
            <div style="font-weight:600;margin-bottom:4px">${label}</div>
            <div>Current: <strong>${fmt(current)}</strong></div>
            <div>Previous: ${fmt(previous)}</div>
            <div style="color:${deltaColor};font-weight:600">${deltaSign}${delta.toFixed(1)}%</div>
        </div>`;
    },
};
