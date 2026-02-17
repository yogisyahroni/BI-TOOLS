'use client';

import React, { useMemo } from 'react';
import { EChartsWrapper } from './echarts-wrapper';
import type { EChartsOption } from 'echarts';

/**
 * TASK-CHART-005: Box Plot (Box & Whisker)
 * Statistical distribution with outliers, mean, and multi-series comparison.
 */

interface BoxPlotSeries {
    name: string;
    data: number[];
    color?: string;
}

interface BoxPlotChartProps {
    series: BoxPlotSeries[];
    title?: string;
    orientation?: 'horizontal' | 'vertical';
    showMean?: boolean;
    showOutliers?: boolean;
    outlierMultiplier?: number;
    className?: string;
    onBoxClick?: (seriesName: string, stats: BoxStats) => void;
}

interface BoxStats {
    min: number;
    q1: number;
    median: number;
    q3: number;
    max: number;
    mean: number;
    outliers: number[];
}

function computeBoxStats(data: number[], multiplier: number): BoxStats {
    const sorted = [...data].sort((a, b) => a - b);
    const n = sorted.length;

    const q1Idx = Math.floor(n * 0.25);
    const medIdx = Math.floor(n * 0.5);
    const q3Idx = Math.floor(n * 0.75);

    const q1 = sorted[q1Idx];
    const median = n % 2 === 0 ? (sorted[medIdx - 1] + sorted[medIdx]) / 2 : sorted[medIdx];
    const q3 = sorted[q3Idx];
    const iqr = q3 - q1;

    const lowerFence = q1 - multiplier * iqr;
    const upperFence = q3 + multiplier * iqr;

    const outliers = sorted.filter(v => v < lowerFence || v > upperFence);
    const inRange = sorted.filter(v => v >= lowerFence && v <= upperFence);

    const mean = data.reduce((s, v) => s + v, 0) / n;

    return {
        min: inRange.length > 0 ? inRange[0] : sorted[0],
        q1,
        median,
        q3,
        max: inRange.length > 0 ? inRange[inRange.length - 1] : sorted[n - 1],
        mean,
        outliers,
    };
}

const DEFAULT_COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899'];

export function BoxPlotChart({
    series,
    title,
    orientation = 'vertical',
    showMean = true,
    showOutliers = true,
    outlierMultiplier = 1.5,
    className = 'h-full w-full min-h-[400px]',
    onBoxClick,
}: BoxPlotChartProps) {
    const isVertical = orientation === 'vertical';

    const stats = useMemo(
        () => series.map(s => computeBoxStats(s.data, outlierMultiplier)),
        [series, outlierMultiplier],
    );

    const categoryAxis = {
        type: 'category' as const,
        data: series.map(s => s.name),
        axisLabel: { fontSize: 11 },
        axisLine: { show: false },
        axisTick: { show: false },
    };

    const valueAxis = {
        type: 'value' as const,
        splitLine: { lineStyle: { opacity: 0.15 } },
    };

        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const ecSeries: any[] = [
        {
            name: 'BoxPlot',
            type: 'boxplot',
            data: stats.map((st, i) => ({
                value: [st.min, st.q1, st.median, st.q3, st.max],
                itemStyle: {
                    color: 'transparent',
                    borderColor: series[i].color ?? DEFAULT_COLORS[i % DEFAULT_COLORS.length],
                    borderWidth: 2,
                },
            })),
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
            tooltip: {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
                formatter: (params: any) => {
                    const idx = params.dataIndex;
                    const st = stats[idx];
                    return `<strong>${series[idx].name}</strong><br/>
                        Max: ${st.max.toFixed(2)}<br/>
                        Q3: ${st.q3.toFixed(2)}<br/>
                        Median: <strong>${st.median.toFixed(2)}</strong><br/>
                        Q1: ${st.q1.toFixed(2)}<br/>
                        Min: ${st.min.toFixed(2)}<br/>
                        Mean: ${st.mean.toFixed(2)}<br/>
                        Outliers: ${st.outliers.length}`;
                },
            },
        },
    ];

        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    // Outlier scatter
    if (showOutliers) {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const outlierData: any[] = [];
        stats.forEach((st, i) => {
            st.outliers.forEach(val => {
                outlierData.push({
                    value: isVertical ? [i, val] : [val, i],
                    itemStyle: { color: series[i].color ?? DEFAULT_COLORS[i % DEFAULT_COLORS.length] },
                });
            });
        });

        if (outlierData.length > 0) {
            ecSeries.push({
                name: 'Outliers',
                type: 'scatter',
                data: outlierData,
                symbolSize: 6,
                z: 3,
            });
        }
    }

    // Mean markers
    if (showMean) {
        ecSeries.push({
            name: 'Mean',
            type: 'scatter',
            data: stats.map((st, i) => ({
                value: isVertical ? [i, st.mean] : [st.mean, i],
                itemStyle: { color: '#f59e0b' },
            })),
            symbol: 'diamond',
            symbolSize: 8,
            z: 4,
        });
    }

    const option: EChartsOption = {
        title: title ? { text: title, left: 'center', textStyle: { fontSize: 14, fontWeight: 600 } } : undefined,
        tooltip: { trigger: 'item' },
        legend: {
            bottom: 0,
            data: showMean ? ['BoxPlot', 'Outliers', 'Mean'] : ['BoxPlot', 'Outliers'],
            textStyle: { fontSize: 11 },
        },
        grid: {
            left: '12%',
            right: '8%',
            top: title ? '15%' : '10%',
            bottom: '15%',
        },
        xAxis: isVertical ? categoryAxis : valueAxis,
        yAxis: isVertical ? valueAxis : categoryAxis,
        series: ecSeries,
        animationDuration: 500,
    };
        // eslint-disable-next-line @typescript-eslint/no-explicit-any

    const events = onBoxClick
        ? {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
            click: (params: any) => {
                if (params.seriesName === 'BoxPlot') {
                    onBoxClick(series[params.dataIndex].name, stats[params.dataIndex]);
                }
            },
        }
        : undefined;

    return <EChartsWrapper options={option} className={className} onEvents={events} />;
}
