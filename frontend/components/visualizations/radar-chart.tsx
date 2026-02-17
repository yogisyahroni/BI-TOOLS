'use client';

import React from 'react';
import { EChartsWrapper } from './echarts-wrapper';
import type { EChartsOption } from 'echarts';

/**
 * TASK-CHART-006: Radar / Spider Chart
 * Multi-axis comparison with configurable indicators, filled area, and multi-series overlay.
 */

interface RadarIndicator {
    name: string;
    max: number;
    min?: number;
}

interface RadarSeries {
    name: string;
    values: number[];
    color?: string;
    areaOpacity?: number;
}

interface RadarChartProps {
    indicators: RadarIndicator[];
    series: RadarSeries[];
    title?: string;
    shape?: 'polygon' | 'circle';
    showArea?: boolean;
    showValues?: boolean;
    radius?: string | number;
    className?: string;
    onPointClick?: (seriesName: string, indicatorName: string, value: number) => void;
}

const RADAR_COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899', '#06b6d4', '#f97316'];

export function RadarChart({
    indicators,
    series,
    title,
    shape = 'polygon',
    showArea = true,
    showValues = false,
    radius = '65%',
    className = 'h-full w-full min-h-[400px]',
    onPointClick,
}: RadarChartProps) {
    const option: EChartsOption = {
        title: title ? { text: title, left: 'center', textStyle: { fontSize: 14, fontWeight: 600 } } : undefined,
        tooltip: {
            trigger: 'item',
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
            formatter: (params: any) => {
                const data = params.data;
                let html = `<strong>${data.name}</strong><br/>`;
                indicators.forEach((ind, i) => {
                    const val = data.value[i];
                    const pct = ((val / ind.max) * 100).toFixed(0);
                    html += `${ind.name}: <strong>${val}</strong> / ${ind.max} (${pct}%)<br/>`;
                });
                return html;
            },
        },
        legend: {
            bottom: 0,
            data: series.map(s => s.name),
            textStyle: { fontSize: 11 },
        },
        radar: {
            indicator: indicators.map(ind => ({
                name: ind.name,
                max: ind.max,
                min: ind.min ?? 0,
            })),
            shape,
            radius,
            center: ['50%', '50%'],
            axisName: { fontSize: 11, color: '#888' },
            splitArea: {
                show: true,
                areaStyle: {
                    color: ['rgba(250,250,250,0.05)', 'rgba(200,200,200,0.05)'],
                },
            },
            splitLine: { lineStyle: { color: 'rgba(128,128,128,0.2)' } },
            axisLine: { lineStyle: { color: 'rgba(128,128,128,0.2)' } },
        },
        series: [
            {
                type: 'radar',
                data: series.map((s, i) => ({
                    name: s.name,
                    value: s.values,
                    symbolSize: 5,
                    lineStyle: {
                        color: s.color ?? RADAR_COLORS[i % RADAR_COLORS.length],
                        width: 2,
                    },
                    areaStyle: showArea
                        ? {
                            color: s.color ?? RADAR_COLORS[i % RADAR_COLORS.length],
                            opacity: s.areaOpacity ?? 0.15,
                        }
                        : undefined,
                    itemStyle: {
                        color: s.color ?? RADAR_COLORS[i % RADAR_COLORS.length],
                    },
                    label: showValues
                        ? {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
                            show: true,
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
                            formatter: (p: any) => p.value,
                            fontSize: 9,
                        }
                        : undefined,
                })),
            },
        ],
        animationDuration: 600,
        animationEasing: 'cubicOut',
    };

        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const events = onPointClick
        ? {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
            click: (params: any) => {
                const seriesData = params.data;
                const dimIdx = params.encode?.value?.[0] ?? 0;
                onPointClick(seriesData.name, indicators[dimIdx]?.name ?? '', seriesData.value[dimIdx]);
            },
        }
        : undefined;

    return <EChartsWrapper options={option} className={className} onEvents={events} />;
}
