'use client';

import React from 'react';
import { EChartsWrapper } from './echarts-wrapper';
import type { EChartsOption } from 'echarts';

/**
 * TASK-CHART-013: Polar Area / Nightingale Rose Chart
 * Florence Nightingale style rose diagram with variable radius sectors.
 */

interface PolarDataItem {
    name: string;
    value: number;
    color?: string;
}

interface PolarAreaChartProps {
    data: PolarDataItem[];
    title?: string;
    roseType?: 'radius' | 'area';
    showLabels?: boolean;
    showLegend?: boolean;
    innerRadius?: string | number;
    outerRadius?: string | number;
    className?: string;
    onSectorClick?: (item: PolarDataItem, percent: number) => void;
}

const POLAR_COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899', '#06b6d4', '#f97316', '#84cc16', '#6366f1'];

export function PolarAreaChart({
    data,
    title,
    roseType = 'radius',
    showLabels = true,
    showLegend = true,
    innerRadius = '15%',
    outerRadius = '75%',
    className = 'h-full w-full min-h-[400px]',
    onSectorClick,
}: PolarAreaChartProps) {
    const total = data.reduce((s, d) => s + d.value, 0);

    const option: EChartsOption = {
        title: title ? { text: title, left: 'center', textStyle: { fontSize: 14, fontWeight: 600 } } : undefined,
        tooltip: {
            trigger: 'item',
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
            formatter: (params: any) => {
                const pct = ((params.value / total) * 100).toFixed(1);
                return `<strong>${params.name}</strong><br/>${params.value.toLocaleString()} (${pct}%)`;
            },
        },
        legend: showLegend
            ? {
                bottom: 0,
                textStyle: { fontSize: 11 },
                itemWidth: 12,
                itemHeight: 12,
            }
            : undefined,
        series: [
            {
                type: 'pie',
                radius: [innerRadius, outerRadius],
                center: ['50%', '50%'],
                roseType,
                avoidLabelOverlap: true,
                itemStyle: {
                    borderRadius: 5,
                    borderColor: 'rgba(255,255,255,0.3)',
                    borderWidth: 2,
                },
                label: showLabels
                    ? {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
                        show: true,
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
                        formatter: (p: any) => `${p.name}\n${((p.value / total) * 100).toFixed(1)}%`,
                        fontSize: 10,
                        lineHeight: 14,
                    }
                    : { show: false },
                labelLine: showLabels ? { show: true, length: 10, length2: 8 } : { show: false },
                emphasis: {
                    label: { show: true, fontSize: 13, fontWeight: 'bold' },
                    itemStyle: { shadowBlur: 10, shadowOffsetX: 0, shadowColor: 'rgba(0,0,0,0.3)' },
                },
                data: data.map((d, i) => ({
                    name: d.name,
                    value: d.value,
                    itemStyle: { color: d.color ?? POLAR_COLORS[i % POLAR_COLORS.length] },
                })),
            },
        ],
        animationDuration: 700,
        animationEasing: 'cubicOut',
    };

        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const events = onSectorClick
        ? {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
            click: (params: any) => {
                const item = data[params.dataIndex];
                onSectorClick(item, (item.value / total) * 100);
            },
        }
        : undefined;

    return <EChartsWrapper options={option} className={className} onEvents={events} />;
}
