'use client';

import React from 'react';
import { EChartsWrapper } from './echarts-wrapper';
import type { EChartsOption } from 'echarts';

/**
 * TASK-CHART-014: Treemap Enhanced
 * Enhanced treemap with breadcrumb drill-down, gradient coloring, and detailed tooltips.
 * Extends existing treemap-chart.tsx with more customization.
 */

interface TreemapEnhancedNode {
    name: string;
    value: number;
    children?: TreemapEnhancedNode[];
    itemStyle?: { color?: string; borderColor?: string };
}

interface TreemapEnhancedProps {
    data: TreemapEnhancedNode[];
    title?: string;
    showBreadcrumb?: boolean;
    showLabels?: boolean;
    leafDepth?: number;
    colorSaturation?: [number, number];
    className?: string;
    onNodeClick?: (node: TreemapEnhancedNode, path: string[]) => void;
}

const TREEMAP_COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899', '#06b6d4', '#f97316'];

export function TreemapEnhanced({
    data,
    title,
    showBreadcrumb = true,
    showLabels = true,
    leafDepth = 2,
    colorSaturation = [0.35, 0.65],
    className = 'h-full w-full min-h-[400px]',
    onNodeClick,
}: TreemapEnhancedProps) {
    // Assign colors to top-level nodes
    const coloredData = data.map((node, i) => ({
        ...node,
        itemStyle: node.itemStyle ?? { color: TREEMAP_COLORS[i % TREEMAP_COLORS.length] },
    }));

    const option: EChartsOption = {
        title: title ? { text: title, left: 'center', textStyle: { fontSize: 14, fontWeight: 600 } } : undefined,
        tooltip: {
            trigger: 'item',
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
            formatter: (params: any) => {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
                const path = params.treePathInfo?.map((p: any) => p.name).filter(Boolean).join(' > ');
                return `${path || params.name}<br/>Value: <strong>${(params.value ?? 0).toLocaleString()}</strong>`;
            },
        },
        series: [
            {
                type: 'treemap',
                data: coloredData,
                leafDepth,
                roam: false,
                nodeClick: 'zoomToNode',
                breadcrumb: {
                    show: showBreadcrumb,
                    top: title ? 35 : 8,
                    left: 'center',
                    itemStyle: {
                        borderRadius: 4,
                    },
                    textStyle: { fontSize: 11 },
                },
                label: {
                    show: showLabels,
                    formatter: '{b}',
                    fontSize: 11,
                    fontWeight: 500,
                    color: '#fff',
                    textShadowBlur: 2,
                    textShadowColor: 'rgba(0,0,0,0.3)',
                },
                upperLabel: {
                    show: true,
                    height: 28,
                    formatter: '{b}',
                    fontSize: 12,
                    fontWeight: 600,
                    color: '#fff',
                    textShadowBlur: 2,
                    textShadowColor: 'rgba(0,0,0,0.3)',
                },
                itemStyle: {
                    borderColor: 'rgba(255,255,255,0.5)',
                    borderWidth: 2,
                    gapWidth: 2,
                },
                emphasis: {
                    upperLabel: { show: true, fontWeight: 'bold' },
                    itemStyle: {
                        borderColor: '#fff',
                        borderWidth: 3,
                    },
                },
                levels: [
                    {
                        itemStyle: {
                            borderColor: 'rgba(255,255,255,0.6)',
                            borderWidth: 3,
                            gapWidth: 3,
                        },
                        upperLabel: { show: true },
                    },
                    {
                        itemStyle: {
                            borderColor: 'rgba(255,255,255,0.4)',
                            borderWidth: 2,
                            gapWidth: 2,
                        },
                        colorSaturation,
                    },
                    {
                        itemStyle: {
                            borderColor: 'rgba(255,255,255,0.2)',
                            borderWidth: 1,
                            gapWidth: 1,
                        },
                        colorSaturation: [colorSaturation[0] - 0.1, colorSaturation[1] + 0.1],
                    },
                ],
            },
        ],
        animationDuration: 600,
        animationEasing: 'cubicOut',
    };

        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const events = onNodeClick
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        ? {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
            click: (params: any) => {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
                const path = params.treePathInfo?.map((p: any) => p.name).filter(Boolean) ?? [];
                onNodeClick({ name: params.name, value: params.value }, path);
            },
        }
        : undefined;

    return <EChartsWrapper options={option} className={className} onEvents={events} />;
}
