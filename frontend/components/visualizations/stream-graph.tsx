'use client';

import React from 'react';
import { EChartsWrapper } from './echarts-wrapper';
import type { EChartsOption } from 'echarts';

/**
 * TASK-CHART-011: Stream Graph (ThemeRiver)
 * Organic flow visualization for temporal composition data.
 */

interface StreamSeries {
    name: string;
    color?: string;
}

interface StreamDataPoint {
    date: string;
    name: string;
    value: number;
}

interface StreamGraphProps {
    data: StreamDataPoint[];
    series: StreamSeries[];
    title?: string;
    smooth?: boolean;
    className?: string;
    onAreaClick?: (seriesName: string, date: string, value: number) => void;
}

const STREAM_COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899', '#06b6d4', '#f97316', '#84cc16', '#6366f1'];

export function StreamGraph({
    data,
    series,
    title,
    _smooth = true,
    className = 'h-full w-full min-h-[400px]',
    onAreaClick,
}: StreamGraphProps) {
    // Convert to ECharts themeRiver format: [date, value, name]
    const themeRiverData = data.map(d => [d.date, d.value, d.name]);

    const option: EChartsOption = {
        title: title ? { text: title, left: 'center', textStyle: { fontSize: 14, fontWeight: 600 } } : undefined,
        tooltip: {
            trigger: 'axis',
            axisPointer: { type: 'line', lineStyle: { color: 'rgba(0,0,0,0.2)', width: 1 } },
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
            formatter: (params: any) => {
                if (!Array.isArray(params) || params.length === 0) return '';
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
                let html = `<strong>${params[0].axisValueLabel}</strong><br/>`;
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
                const sorted = [...params].sort((a: any, b: any) => (b.value?.[1] ?? 0) - (a.value?.[1] ?? 0));
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
                sorted.forEach((p: any) => {
                    if (p.value && p.value[1] !== undefined) {
                        html += `${p.marker} ${p.value[2] ?? p.seriesName}: <strong>${p.value[1].toLocaleString()}</strong><br/>`;
                    }
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
        singleAxis: {
            type: 'time',
            top: title ? '15%' : '8%',
            bottom: '14%',
            axisLabel: { fontSize: 10 },
        },
        series: [
            {
                type: 'themeRiver',
                data: themeRiverData,
                emphasis: {
                    itemStyle: {
                        shadowBlur: 10,
                        shadowColor: 'rgba(0,0,0,0.2)',
                    },
                },
                label: {
                    show: true,
                    fontSize: 10,
                    color: '#333',
                },
                itemStyle: {
                    borderWidth: 0,
                },
                color: series.map((s, i) => s.color ?? STREAM_COLORS[i % STREAM_COLORS.length]),
            },
        ],
        animationDuration: 800,
        animationEasing: 'cubicOut',
    };
        // eslint-disable-next-line @typescript-eslint/no-explicit-any

    const events = onAreaClick
        ? {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
            click: (params: any) => {
                if (params.data) {
                    onAreaClick(params.data[2], params.data[0], params.data[1]);
                }
            },
        }
        : undefined;

    return <EChartsWrapper options={option} className={className} onEvents={events} />;
}
