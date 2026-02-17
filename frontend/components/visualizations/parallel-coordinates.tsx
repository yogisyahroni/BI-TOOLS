'use client';

import React from 'react';
import { EChartsWrapper } from './echarts-wrapper';
import type { EChartsOption } from 'echarts';

/**
 * TASK-CHART-012: Parallel Coordinates
 * Multi-dimensional data analysis with brushable axes.
 */

interface ParallelAxis {
    name: string;
    type?: 'value' | 'category';
    min?: number;
    max?: number;
    categories?: string[];
    inverse?: boolean;
}

interface ParallelCoordinatesProps {
    axes: ParallelAxis[];
    data: (number | string)[][];
    title?: string;
    seriesColors?: string[];
    lineOpacity?: number;
    lineWidth?: number;
    smooth?: boolean;
    className?: string;
    onLineSelect?: (selectedIndices: number[]) => void;
}

const _PARALLEL_COLORS = ['#3b82f680', '#10b98180', '#f59e0b80', '#ef444480', '#8b5cf680'];

export function ParallelCoordinates({
    axes,
    data,
    title,
    lineOpacity = 0.35,
    lineWidth = 1.5,
    smooth = false,
    className = 'h-full w-full min-h-[400px]',
    _onLineSelect,
}: ParallelCoordinatesProps) {
    const parallelAxes = axes.map((axis, i) => ({
        dim: i,
        name: axis.name,
        type: axis.type ?? 'value',
        min: axis.min,
        max: axis.max,
        data: axis.categories,
        inverse: axis.inverse ?? false,
        nameTextStyle: { fontSize: 11, fontWeight: 500 },
        axisLabel: { fontSize: 9 },
    }));

    const option: EChartsOption = {
        title: title ? { text: title, left: 'center', textStyle: { fontSize: 14, fontWeight: 600 } } : undefined,
        parallelAxis: parallelAxes,
        parallel: {
            left: '8%',
            right: '8%',
            top: title ? '18%' : '12%',
            bottom: '12%',
            parallelAxisDefault: {
                type: 'value',
                nameLocation: 'start',
                nameGap: 20,
                axisLine: { lineStyle: { color: '#aaa' } },
                axisTick: { lineStyle: { color: '#aaa' } },
                splitLine: { show: false },
            },
        },
        tooltip: {
            trigger: 'item',
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
            formatter: (params: any) => {
                if (!params.data) return '';
                const row = params.data;
                let html = '';
                axes.forEach((ax, i) => {
                    html += `${ax.name}: <strong>${row[i]}</strong><br/>`;
                });
                return html;
            },
        },
        series: [
            {
                type: 'parallel',
                lineStyle: {
                    width: lineWidth,
                    opacity: lineOpacity,
                },
                emphasis: {
                    lineStyle: {
                        width: 3,
                        opacity: 1,
                    },
                },
                smooth,
                data,
                progressive: 500,
                progressiveThreshold: 3000,
            },
        ],
        animationDuration: 600,
    };

    return <EChartsWrapper options={option} className={className} />;
}
