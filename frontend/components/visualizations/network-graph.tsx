'use client';

import React from 'react';
import { EChartsWrapper } from './echarts-wrapper';
import type { EChartsOption } from 'echarts';

/**
 * TASK-CHART-018: Network / Force-Directed Graph
 * Node-link diagram for relationship and dependency visualization.
 */

interface NetworkNode {
    id: string;
    name: string;
    value?: number;
    category?: number;
    symbolSize?: number;
    color?: string;
}

interface NetworkEdge {
    source: string;
    target: string;
    value?: number;
    label?: string;
}

interface NetworkCategory {
    name: string;
    color?: string;
}

interface NetworkGraphProps {
    nodes: NetworkNode[];
    edges: NetworkEdge[];
    categories?: NetworkCategory[];
    title?: string;
    layout?: 'force' | 'circular';
    repulsion?: number;
    gravity?: number;
    edgeLength?: number | [number, number];
    draggable?: boolean;
    showLabels?: boolean;
    className?: string;
    onNodeClick?: (node: NetworkNode) => void;
    onEdgeClick?: (edge: NetworkEdge) => void;
}

const NETWORK_COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899', '#06b6d4', '#f97316'];

export function NetworkGraph({
    nodes,
    edges,
    categories,
    title,
    layout = 'force',
    repulsion = 200,
    gravity = 0.1,
    edgeLength = [50, 200],
    draggable = true,
    showLabels = true,
    className = 'h-full w-full min-h-[400px]',
    onNodeClick,
    onEdgeClick,
}: NetworkGraphProps) {
    const ecCategories = categories?.map((c, i) => ({
        name: c.name,
        itemStyle: { color: c.color ?? NETWORK_COLORS[i % NETWORK_COLORS.length] },
    })) ?? [];

    // Compute node sizes based on connections if not provided
    const connectionCounts = new Map<string, number>();
    edges.forEach(e => {
        connectionCounts.set(e.source, (connectionCounts.get(e.source) ?? 0) + 1);
        connectionCounts.set(e.target, (connectionCounts.get(e.target) ?? 0) + 1);
    });

    const ecNodes = nodes.map(n => {
        const connections = connectionCounts.get(n.id) ?? 0;
        const autoSize = Math.max(10, Math.min(40, 8 + connections * 4));

        return {
            id: n.id,
            name: n.name,
            value: n.value ?? connections,
            category: n.category,
            symbolSize: n.symbolSize ?? autoSize,
            itemStyle: n.color ? { color: n.color } : undefined,
            draggable,
            label: {
                show: showLabels && (n.symbolSize ?? autoSize) > 15,
                fontSize: 10,
                position: 'right' as const,
            },
        };
    });

    const ecEdges = edges.map(e => ({
        source: e.source,
        target: e.target,
        value: e.value,
        lineStyle: {
            width: e.value ? Math.max(1, Math.min(5, e.value)) : 1,
            curveness: 0.15,
            opacity: 0.5,
        },
        label: e.label
            ? { show: true, formatter: e.label, fontSize: 9 }
            : undefined,
    }));

    const option: EChartsOption = {
        title: title ? { text: title, left: 'center', textStyle: { fontSize: 14, fontWeight: 600 } } : undefined,
        tooltip: {
            trigger: 'item',
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
            formatter: (params: any) => {
                if (params.dataType === 'edge') {
                    const source = nodes.find(n => n.id === params.data.source)?.name ?? params.data.source;
                    const target = nodes.find(n => n.id === params.data.target)?.name ?? params.data.target;
                    return `${source} → ${target}${params.data.value ? `<br/>Weight: ${params.data.value}` : ''}`;
                }
                return `<strong>${params.data.name}</strong><br/>Connections: ${params.data.value ?? 0}`;
            },
        },
        legend: ecCategories.length > 0
            ? {
                bottom: 0,
                data: ecCategories.map(c => c.name),
                textStyle: { fontSize: 11 },
            }
            : undefined,
        series: [
            {
                type: 'graph',
                layout,
                data: ecNodes,
                links: ecEdges,
                categories: ecCategories.length > 0 ? ecCategories : undefined,
                roam: true,
                force: layout === 'force'
                    ? {
                        repulsion,
                        gravity,
                        edgeLength,
                        layoutAnimation: true,
                    }
                    : undefined,
                circular: layout === 'circular'
                    ? { rotateLabel: true }
                    : undefined,
                emphasis: {
                    focus: 'adjacency',
                    lineStyle: { width: 3, opacity: 1 },
                    itemStyle: { borderWidth: 2, borderColor: '#fff' },
                },
                label: {
                    show: showLabels,
                    position: 'right',
                    fontSize: 10,
                },
                lineStyle: {
                    color: 'source',
                    curveness: 0.15,
                },
                edgeSymbol: ['circle', 'arrow'],
                edgeSymbolSize: [4, 8],
            },
        ],
        animationDuration: 1000,
        animationEasingUpdate: 'quinticInOut',
    };
        // eslint-disable-next-line @typescript-eslint/no-explicit-any

        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const events: Record<string, (...args: any[]) => any> = {};
    if (onNodeClick || onEdgeClick) {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        events.click = (params: any) => {
            if (params.dataType === 'node' && onNodeClick) {
                const node = nodes.find(n => n.id === params.data.id);
                if (node) onNodeClick(node);
            } else if (params.dataType === 'edge' && onEdgeClick) {
                const edge = edges.find(e => e.source === params.data.source && e.target === params.data.target);
                if (edge) onEdgeClick(edge);
            }
        };
    }

    return <EChartsWrapper options={option} className={className} onEvents={Object.keys(events).length > 0 ? events : undefined} />;
}
