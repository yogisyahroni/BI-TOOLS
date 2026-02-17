'use client';

import React from 'react';
import { EChartsWrapper } from './echarts-wrapper';
import type { EChartsOption } from 'echarts';

/**
 * TASK-CHART-020: Nested Pie / Double Doughnut
 * Concentric ring chart showing hierarchical category breakdown at two levels.
 */

interface NestedPieRing {
    name: string;
    data: { name: string; value: number; color?: string }[];
    radius: [string | number, string | number];
}

interface NestedPieChartProps {
    rings: NestedPieRing[];
    title?: string;
    showLabels?: boolean;
    showLegend?: boolean;
    className?: string;
    onSegmentClick?: (ringName: string, segmentName: string, value: number) => void;
}

const NESTED_COLORS_OUTER = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899', '#06b6d4', '#f97316'];
const NESTED_COLORS_INNER = ['#60a5fa', '#34d399', '#fbbf24', '#f87171', '#a78bfa', '#f472b6', '#22d3ee', '#fb923c'];

export function NestedPieChart({
    rings,
    title,
    showLabels = true,
    showLegend = true,
    className = 'h-full w-full min-h-[400px]',
    onSegmentClick,
}: NestedPieChartProps) {
    const colorSets = [NESTED_COLORS_INNER, NESTED_COLORS_OUTER];

    const allLegendItems = rings.flatMap(r => r.data.map(d => d.name));
    const uniqueLegend = Array.from(new Set(allLegendItems));

    const option: EChartsOption = {
        title: title ? { text: title, left: 'center', textStyle: { fontSize: 14, fontWeight: 600 } } : undefined,
        tooltip: {
            trigger: 'item',
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
            formatter: (params: any) => {
                const ringName = params.seriesName ?? '';
                const total = rings
                    .find(r => r.name === ringName)
                    ?.data.reduce((s, d) => s + d.value, 0) ?? 1;
                const pct = ((params.value / total) * 100).toFixed(1);
                return `<strong>${ringName}</strong><br/>
                    ${params.name}: ${params.value.toLocaleString()} (${pct}%)`;
            },
        },
        legend: showLegend
            ? {
                bottom: 0,
                data: uniqueLegend,
                textStyle: { fontSize: 10 },
                itemWidth: 10,
                itemHeight: 10,
            }
            : undefined,
        series: rings.map((ring, ri) => ({
            name: ring.name,
            type: 'pie' as const,
            radius: ring.radius,
            center: ['50%', '48%'],
            avoidLabelOverlap: true,
            padAngle: ri === 0 ? 3 : 2,
            itemStyle: {
                borderRadius: ri === 0 ? 4 : 6,
                borderColor: 'rgba(255,255,255,0.3)',
                borderWidth: 2,
            },
            label: showLabels
                ? {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
                    show: ri === rings.length - 1,
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
                    formatter: (p: any) => {
                        const total = ring.data.reduce((s, d) => s + d.value, 0);
                        const pct = ((p.value / total) * 100).toFixed(0);
                        return `${p.name} ${pct}%`;
                    },
                    fontSize: 10,
                }
                : { show: false },
            labelLine: showLabels
                ? { show: ri === rings.length - 1, length: 12, length2: 6 }
                : { show: false },
            emphasis: {
                scaleSize: 5,
                label: { show: true, fontWeight: 'bold', fontSize: 12 },
            },
            data: ring.data.map((d, di) => ({
                name: d.name,
                value: d.value,
                itemStyle: {
                    color: d.color ?? (colorSets[ri % colorSets.length])[di % (colorSets[ri % colorSets.length]).length],
                },
            })),
        })),
        animationDuration: 700,
        animationEasing: 'cubicOut',
    };

        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const events = onSegmentClick
        ? {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
            click: (params: any) => {
                onSegmentClick(params.seriesName ?? '', params.name, params.value);
            },
        }
        : undefined;

    return <EChartsWrapper options={option} className={className} onEvents={events} />;
}
