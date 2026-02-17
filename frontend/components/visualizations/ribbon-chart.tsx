'use client';

import React from 'react';
import { EChartsWrapper } from './echarts-wrapper';
import type { EChartsOption } from 'echarts';

/**
 * TASK-CHART-010: Ribbon Chart
 * Stacked area chart with flowing transitions (Power BI-style).
 * Shows how categories contribute over time with ribbon-like areas.
 */

interface RibbonSeries {
    name: string;
    data: number[];
    color?: string;
}

interface RibbonChartProps {
    categories: string[];
    series: RibbonSeries[];
    title?: string;
    smooth?: boolean;
    showPercentage?: boolean;
    className?: string;
    onAreaClick?: (seriesName: string, categoryIndex: number) => void;
}

const RIBBON_COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899', '#06b6d4', '#f97316', '#84cc16', '#6366f1'];

export function RibbonChart({
    categories,
    series,
    title,
    smooth = true,
    showPercentage = false,
    className = 'h-full w-full min-h-[400px]',
    onAreaClick,
}: RibbonChartProps) {
    // If percentage mode, normalize each category column to 100%
    const normalizedSeries = showPercentage
        ? series.map(s => ({
            ...s,
            data: s.data.map((val, ci) => {
                const colTotal = series.reduce((sum, sr) => sum + sr.data[ci], 0);
                return colTotal > 0 ? (val / colTotal) * 100 : 0;
            }),
        }))
        : series;

    const option: EChartsOption = {
        title: title ? { text: title, left: 'center', textStyle: { fontSize: 14, fontWeight: 600 } } : undefined,
        tooltip: {
            trigger: 'axis',
            axisPointer: { type: 'cross', label: { show: false } },
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
            formatter: (params: any) => {
                if (!Array.isArray(params) || params.length === 0) return '';
                let html = `<strong>${params[0].axisValueLabel}</strong><br/>`;
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
                let total = 0;
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
                params.forEach((p: any) => { total += p.value; });
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
                params.forEach((p: any) => {
                    const pct = total > 0 ? ((p.value / total) * 100).toFixed(1) : '0';
                    const marker = p.marker;
                    html += `${marker} ${p.seriesName}: <strong>${showPercentage ? p.value.toFixed(1) + '%' : p.value.toLocaleString()}</strong> (${pct}%)<br/>`;
                });
                return html;
            },
        },
        legend: {
            bottom: 0,
            textStyle: { fontSize: 11 },
            itemWidth: 12,
            itemHeight: 12,
        },
        grid: {
            left: '8%',
            right: '5%',
            top: title ? '15%' : '8%',
            bottom: '14%',
        },
        xAxis: {
            type: 'category',
            data: categories,
            boundaryGap: false,
            axisLabel: { fontSize: 10 },
        },
        yAxis: {
            type: 'value',
            max: showPercentage ? 100 : undefined,
            axisLabel: {
                fontSize: 10,
                formatter: showPercentage ? '{value}%' : undefined,
            },
            splitLine: { lineStyle: { opacity: 0.15 } },
        },
        series: normalizedSeries.map((s, i) => ({
            name: s.name,
            type: 'line' as const,
            stack: 'total',
            areaStyle: {
                opacity: 0.7,
                color: s.color ?? RIBBON_COLORS[i % RIBBON_COLORS.length],
            },
            lineStyle: {
                width: 1,
                color: s.color ?? RIBBON_COLORS[i % RIBBON_COLORS.length],
            },
            itemStyle: {
                color: s.color ?? RIBBON_COLORS[i % RIBBON_COLORS.length],
            },
            data: s.data,
            smooth: smooth ? 0.4 : false,
            symbol: 'none',
            emphasis: {
                focus: 'series',
                areaStyle: { opacity: 0.9 },
            },
        })),
        animationDuration: 800,
        animationEasing: 'cubicOut',
    };
        // eslint-disable-next-line @typescript-eslint/no-explicit-any

    const events = onAreaClick
        ? {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
            click: (params: any) => {
                onAreaClick(params.seriesName, params.dataIndex);
            },
        }
        : undefined;

    return <EChartsWrapper options={option} className={className} onEvents={events} />;
}
