'use client';

import React, { useMemo } from 'react';
import { EChartsWrapper } from './echarts-wrapper';
import type { EChartsOption } from 'echarts';

/**
 * TASK-CHART-004: Histogram
 * Distribution analysis with auto binning, density curve, and rug plot.
 */

interface HistogramChartProps {
    data: number[];
    title?: string;
    binCount?: number;
    showDensity?: boolean;
    showRug?: boolean;
    color?: string;
    densityColor?: string;
    xAxisLabel?: string;
    yAxisLabel?: string;
    className?: string;
    onBinClick?: (binRange: [number, number], count: number) => void;
}

function computeBins(data: number[], binCount: number): { start: number; end: number; count: number }[] {
    if (data.length === 0) return [];
    const sorted = [...data].sort((a, b) => a - b);
    const minVal = sorted[0];
    const maxVal = sorted[sorted.length - 1];
    const range = maxVal - minVal || 1;
    const binWidth = range / binCount;

    const bins = Array.from({ length: binCount }, (_, i) => ({
        start: minVal + i * binWidth,
        end: minVal + (i + 1) * binWidth,
        count: 0,
    }));

    for (const val of sorted) {
        let idx = Math.floor((val - minVal) / binWidth);
        if (idx >= binCount) idx = binCount - 1;
        bins[idx].count++;
    }

    return bins;
}

function gaussianKDE(data: number[], points: number[], bandwidth: number): number[] {
    const n = data.length;
    return points.map(x => {
        let sum = 0;
        for (const xi of data) {
            const u = (x - xi) / bandwidth;
            sum += Math.exp(-0.5 * u * u) / Math.sqrt(2 * Math.PI);
        }
        return sum / (n * bandwidth);
    });
}

export function HistogramChart({
    data,
    title,
    binCount,
    showDensity = false,
    showRug = false,
    color = '#3b82f6',
    densityColor = '#ef4444',
    xAxisLabel,
    yAxisLabel = 'Frequency',
    className = 'h-full w-full min-h-[400px]',
    onBinClick,
}: HistogramChartProps) {
    const effectiveBinCount = binCount ?? Math.max(Math.ceil(Math.sqrt(data.length)), 5);
    const bins = useMemo(() => computeBins(data, effectiveBinCount), [data, effectiveBinCount]);

    const densityCurve = useMemo(() => {
        if (!showDensity || data.length < 2) return null;
        const sorted = [...data].sort((a, b) => a - b);
        const minV = sorted[0];
        const maxV = sorted[sorted.length - 1];
        const range = maxV - minV || 1;
        const bandwidth = range / effectiveBinCount * 1.5;
        const steps = 50;
        const points = Array.from({ length: steps }, (_, i) => minV + (i / (steps - 1)) * range);
        const densities = gaussianKDE(data, points, bandwidth);
        const maxDensity = Math.max(...densities);
        const maxCount = Math.max(...bins.map(b => b.count));
        const scale = maxCount / (maxDensity || 1);
        return points.map((x, i) => [x, densities[i] * scale]);
    }, [data, showDensity, effectiveBinCount, bins]);

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
                const bin = bins[idx];
                return `Range: ${bin.start.toFixed(2)} – ${bin.end.toFixed(2)}<br/>Count: <strong>${bin.count}</strong>`;
            },
        },
        grid: {
            left: '10%',
            right: '5%',
            top: title ? '15%' : '8%',
            bottom: showRug ? '18%' : '12%',
        },
        xAxis: {
            type: 'category',
            data: bins.map(b => `${b.start.toFixed(1)}`),
            name: xAxisLabel,
            nameLocation: 'middle',
            nameGap: 30,
            axisLabel: { fontSize: 10, rotate: bins.length > 15 ? 45 : 0 },
        },
        yAxis: {
            type: 'value',
            name: yAxisLabel,
            nameLocation: 'middle',
            nameGap: 40,
            splitLine: { lineStyle: { opacity: 0.15 } },
        },
        series: [
            {
                name: 'Frequency',
                type: 'bar',
                data: bins.map(b => b.count),
                barWidth: '90%',
                itemStyle: {
                    color,
                    borderRadius: [3, 3, 0, 0],
                },
                emphasis: {
                    itemStyle: { color, opacity: 0.8 },
                },
            },
            ...(densityCurve
                ? [{
                    name: 'Density',
                    type: 'line' as const,
                    data: densityCurve,
                    smooth: true,
                    showSymbol: false,
                    lineStyle: { color: densityColor, width: 2 },
                    xAxisIndex: 0,
                    encode: { x: 0, y: 1 },
                    z: 5,
                }]
                : []),
        ],
        animationDuration: 500,
        animationEasing: 'cubicOut',
    };

    const events = onBinClick
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        ? {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
            click: (params: any) => {
                if (params.seriesName === 'Frequency') {
                    const bin = bins[params.dataIndex];
                    onBinClick([bin.start, bin.end], bin.count);
                }
            },
        }
        : undefined;

    return <EChartsWrapper options={option} className={className} onEvents={events} />;
}
