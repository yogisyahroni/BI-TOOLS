'use client';

import React from 'react';
import { EChartsWrapper } from './echarts-wrapper';
import type { EChartsOption } from 'echarts';

/**
 * TASK-CHART-008: Sunburst Chart
 * Hierarchical circular visualization for multi-level categorical breakdowns.
 */

interface SunburstNode {
    name: string;
    value?: number;
    children?: SunburstNode[];
    itemStyle?: { color?: string };
}

interface SunburstChartProps {
    data: SunburstNode[];
    title?: string;
    innerRadius?: string | number;
    outerRadius?: string | number;
    levels?: number;
    showLabels?: boolean;
    highlightAncestor?: boolean;
    className?: string;
    onNodeClick?: (node: SunburstNode, depth: number) => void;
}

const SUNBURST_COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899', '#06b6d4', '#f97316'];

export function SunburstChart({
    data,
    title,
    innerRadius = '15%',
    outerRadius = '80%',
    levels: maxLevels,
    showLabels = true,
    highlightAncestor = true,
    className = 'h-full w-full min-h-[400px]',
    onNodeClick,
}: SunburstChartProps) {
    // Assign colors to top-level nodes if not provided
    const coloredData = data.map((node, i) => ({
        ...node,
        itemStyle: node.itemStyle ?? { color: SUNBURST_COLORS[i % SUNBURST_COLORS.length] },
    }));

        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const levelConfigs: any[] = [
        { r0: innerRadius, r: typeof outerRadius === 'string' ? '28%' : (outerRadius as number) * 0.35 },
    ];

    const depth = getMaxDepth(data);
    const effectiveLevels = maxLevels ?? depth;

    for (let i = 1; i <= effectiveLevels; i++) {
        const startPct = 28 + ((i - 1) / effectiveLevels) * 52;
        const endPct = 28 + (i / effectiveLevels) * 52;
        levelConfigs.push({
            r0: `${startPct}%`,
            r: `${endPct}%`,
            label: {
                show: showLabels && i <= 2,
                rotate: i > 1 ? 'radial' : 0,
                fontSize: Math.max(12 - i * 1.5, 8),
            },
            itemStyle: {
                borderWidth: Math.max(3 - i, 1),
                borderColor: 'rgba(255,255,255,0.5)',
            },
        });
    }

    const option: EChartsOption = {
        title: title ? { text: title, left: 'center', textStyle: { fontSize: 14, fontWeight: 600 } } : undefined,
        tooltip: {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
            trigger: 'item',
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
                type: 'sunburst',
                data: coloredData,
                radius: [innerRadius, outerRadius],
                center: ['50%', '50%'],
                nodeClick: 'rootToNode',
                sort: undefined,
                emphasis: {
                    focus: highlightAncestor ? 'ancestor' : 'self',
                },
                label: {
                    show: showLabels,
                    rotate: 'radial',
                    fontSize: 11,
                    minAngle: 10,
                },
                itemStyle: {
                    borderRadius: 4,
                    borderWidth: 2,
                    borderColor: 'rgba(255,255,255,0.3)',
                },
                levels: levelConfigs,
            },
        ],
        animationDuration: 700,
        animationEasing: 'cubicOut',
    };
        // eslint-disable-next-line @typescript-eslint/no-explicit-any

    const events = onNodeClick
        ? {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
            click: (params: any) => {
                const depth = params.treePathInfo?.length ?? 0;
                onNodeClick({ name: params.name, value: params.value }, depth);
            },
        }
        : undefined;

    return <EChartsWrapper options={option} className={className} onEvents={events} />;
}

function getMaxDepth(nodes: SunburstNode[]): number {
    let max = 0;
    for (const node of nodes) {
        if (node.children && node.children.length > 0) {
            const childDepth = getMaxDepth(node.children);
            max = Math.max(max, childDepth + 1);
        }
    }
    return max;
}
