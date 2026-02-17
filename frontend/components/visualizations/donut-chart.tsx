'use client';

import React from 'react';
import { EChartsWrapper } from './echarts-wrapper';
import type { EChartsOption } from 'echarts';

/**
 * TASK-CHART-007: Donut Chart
 * Ring-style pie chart with center label, interactive segments, and configurable radius.
 */

interface DonutDataItem {
    name: string;
    value: number;
    color?: string;
}

interface DonutChartProps {
    data: DonutDataItem[];
    title?: string;
    centerLabel?: string;
    centerValue?: string | number;
    innerRadius?: string | number;
    outerRadius?: string | number;
    roseType?: false | 'radius' | 'area';
    showLabels?: boolean;
    showLegend?: boolean;
    padAngle?: number;
    className?: string;
    onSegmentClick?: (item: DonutDataItem, percent: number) => void;
}

const DONUT_COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899', '#06b6d4', '#f97316', '#84cc16', '#6366f1'];

export function DonutChart({
    data,
    title,
    centerLabel,
    centerValue,
    innerRadius = '55%',
    outerRadius = '78%',
    roseType = false,
    showLabels = true,
    showLegend = true,
    padAngle = 2,
    className = 'h-full w-full min-h-[350px]',
    onSegmentClick,
}: DonutChartProps) {
    const total = data.reduce((s, d) => s + d.value, 0);

    const option: EChartsOption = {
        title: title
            ? { text: title, left: 'center', top: 0, textStyle: { fontSize: 14, fontWeight: 600 } }
            : undefined,
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
                orient: 'horizontal',
                bottom: 0,
                textStyle: { fontSize: 11 },
                itemWidth: 12,
                itemHeight: 12,
            }
            : undefined,
        graphic: (centerLabel || centerValue !== undefined)
            ? [
                ...(centerValue !== undefined
                    ? [{
                        type: 'text' as const,
                        left: 'center',
                        top: '42%',
                        style: {
                            text: String(centerValue),
                            fontSize: 28,
                            fontWeight: 'bold' as const,
                            fill: '#333',
                            textAlign: 'center' as const,
                        },
                    }]
                    : []),
                ...(centerLabel
                    ? [{
                        type: 'text' as const,
                        left: 'center',
                        top: centerValue !== undefined ? '53%' : '48%',
                        style: {
                            text: centerLabel,
                            fontSize: 12,
                            fill: '#999',
                            textAlign: 'center' as const,
                        },
                    }]
                    : []),
            ]
            : undefined,
        series: [
            {
                type: 'pie',
                radius: [innerRadius, outerRadius],
                center: ['50%', '48%'],
                avoidLabelOverlap: true,
                padAngle,
                roseType: roseType || undefined,
                itemStyle: {
                    borderRadius: 6,
                    borderColor: 'transparent',
                    borderWidth: 2,
                },
                label: showLabels
                    ? {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
                        show: true,
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
                        formatter: (p: any) => `${p.name}\n${((p.value / total) * 100).toFixed(1)}%`,
                        fontSize: 11,
                        lineHeight: 16,
                    }
                    : { show: false },
                labelLine: showLabels ? { show: true, length: 15, length2: 8 } : { show: false },
                emphasis: {
                    scaleSize: 6,
                    label: {
                        show: true,
                        fontSize: 13,
                        fontWeight: 'bold',
                    },
                },
                data: data.map((d, i) => ({
                    name: d.name,
                    value: d.value,
                    itemStyle: d.color ? { color: d.color } : { color: DONUT_COLORS[i % DONUT_COLORS.length] },
                })),
            },
        ],
        animationDuration: 700,
        animationEasing: 'cubicOut',
    };

        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const events = onSegmentClick
        ? {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
            click: (params: any) => {
                const item = data[params.dataIndex];
                const pct = (item.value / total) * 100;
                onSegmentClick(item, pct);
            },
        }
        : undefined;

    return <EChartsWrapper options={option} className={className} onEvents={events} />;
}
