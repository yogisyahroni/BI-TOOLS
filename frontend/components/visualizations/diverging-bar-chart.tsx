'use client';

import React from 'react';
import { EChartsWrapper } from './echarts-wrapper';
import type { EChartsOption } from 'echarts';

/**
 * TASK-CHART-019: Diverging Bar Chart
 * Bi-directional bars showing positive/negative deviations from a baseline.
 */

interface DivergingDataItem {
    label: string;
    value: number;
    color?: string;
}

interface DivergingBarChartProps {
    data: DivergingDataItem[];
    title?: string;
    baselineLabel?: string;
    positiveColor?: string;
    negativeColor?: string;
    orientation?: 'horizontal' | 'vertical';
    showValues?: boolean;
    className?: string;
    onBarClick?: (item: DivergingDataItem, index: number) => void;
}

export function DivergingBarChart({
    data,
    title,
    _baselineLabel = '0',
    positiveColor = '#10b981',
    negativeColor = '#ef4444',
    orientation = 'horizontal',
    showValues = true,
    className = 'h-full w-full min-h-[400px]',
    onBarClick,
}: DivergingBarChartProps) {
    const isHorizontal = orientation === 'horizontal';
    const sortedData = [...data].sort((a, b) => b.value - a.value);

    const positiveData = sortedData.map(d => d.value >= 0 ? d.value : null);
    const negativeData = sortedData.map(d => d.value < 0 ? d.value : null);

    const categoryAxis = {
        type: 'category' as const,
        data: sortedData.map(d => d.label),
        axisLine: { show: false },
        axisTick: { show: false },
        axisLabel: { fontSize: 11, fontWeight: 500 },
    };

    const valueAxis = {
        type: 'value' as const,
        splitLine: { lineStyle: { opacity: 0.15 } },
        axisLabel: { fontSize: 10 },
    };

    const option: EChartsOption = {
        title: title ? { text: title, left: 'center', textStyle: { fontSize: 14, fontWeight: 600 } } : undefined,
        tooltip: {
            trigger: 'axis',
            axisPointer: { type: 'shadow' },
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
            formatter: (params: any) => {
                if (!Array.isArray(params) || params.length === 0) return '';
                const idx = params[0].dataIndex;
                const item = sortedData[idx];
                const sign = item.value >= 0 ? '+' : '';
                return `<strong>${item.label}</strong><br/>Value: ${sign}${item.value.toLocaleString()}`;
            },
        },
        grid: {
            left: isHorizontal ? '18%' : '10%',
            right: '8%',
            top: title ? '15%' : '8%',
            bottom: '8%',
        },
        [isHorizontal ? 'yAxis' : 'xAxis']: categoryAxis,
        [isHorizontal ? 'xAxis' : 'yAxis']: valueAxis,
        series: [
            {
                name: 'Positive',
                type: 'bar',
                stack: 'total',
                data: positiveData,
                itemStyle: {
                    color: positiveColor,
                    borderRadius: isHorizontal ? [0, 4, 4, 0] : [4, 4, 0, 0],
                },
                label: showValues
                    ? {
                        show: true,
                        position: isHorizontal ? 'right' : 'top',
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
                        fontSize: 10,
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
                        formatter: (p: any) => p.value !== null ? `+${p.value.toLocaleString()}` : '',
                    }
                    : undefined,
            },
            {
                name: 'Negative',
                type: 'bar',
                stack: 'total',
                data: negativeData,
                itemStyle: {
                    color: negativeColor,
                    borderRadius: isHorizontal ? [4, 0, 0, 4] : [0, 0, 4, 4],
                },
                label: showValues
                    ? {
                        show: true,
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
                        position: isHorizontal ? 'left' : 'bottom',
                        fontSize: 10,
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
                        formatter: (p: any) => p.value !== null ? p.value.toLocaleString() : '',
                    }
                    : undefined,
            },
        ],
        animationDuration: 600,
        animationEasing: 'cubicOut',
    };
        // eslint-disable-next-line @typescript-eslint/no-explicit-any

    // Add baseline markLine
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    if (isHorizontal) {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        (option as any).xAxis.axisLine = { show: true, lineStyle: { color: '#666', width: 2 } };
    } else {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        (option as any).yAxis.axisLine = { show: true, lineStyle: { color: '#666', width: 2 } };
    }

    const events = onBarClick
        ? {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
            click: (params: any) => {
                onBarClick(sortedData[params.dataIndex], params.dataIndex);
            },
        }
        : undefined;

    return <EChartsWrapper options={option} className={className} onEvents={events} />;
}
