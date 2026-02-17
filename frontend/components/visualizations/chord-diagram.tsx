'use client';

import React, { useMemo } from 'react';
import { cn } from '@/lib/utils';

/**
 * TASK-CHART-016: Chord Diagram
 * Inter-relationships between groups using a circular layout with arc ribbons.
 * Pure SVG implementation.
 */

interface ChordData {
    names: string[];
    matrix: number[][];
    colors?: string[];
}

interface ChordDiagramProps {
    data: ChordData;
    title?: string;
    innerRadius?: number;
    outerRadius?: number;
    padAngle?: number;
    className?: string;
    onChordClick?: (source: string, target: string, value: number) => void;
}

const CHORD_COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899', '#06b6d4', '#f97316', '#84cc16', '#6366f1'];

function polarToCartesian(cx: number, cy: number, r: number, angle: number) {
    return {
        x: cx + r * Math.cos(angle),
        y: cy + r * Math.sin(angle),
    };
}

function _arcPath(cx: number, cy: number, r: number, startAngle: number, endAngle: number) {
    const start = polarToCartesian(cx, cy, r, endAngle);
    const end = polarToCartesian(cx, cy, r, startAngle);
    const largeArcFlag = endAngle - startAngle > Math.PI ? 1 : 0;
    return `M ${start.x} ${start.y} A ${r} ${r} 0 ${largeArcFlag} 0 ${end.x} ${end.y}`;
}

export function ChordDiagram({
    data,
    title,
    innerRadius = 160,
    outerRadius = 180,
    padAngle = 0.04,
    className = 'h-full w-full min-h-[400px]',
    onChordClick,
}: ChordDiagramProps) {
    const { names, matrix, colors } = data;
    const n = names.length;
    const effectiveColors = colors ?? names.map((_, i) => CHORD_COLORS[i % CHORD_COLORS.length]);

    const layout = useMemo(() => {
        // Sum of all flows
        const groupTotals = matrix.map(row => row.reduce((s, v) => s + v, 0));
        const grandTotal = groupTotals.reduce((s, v) => s + v, 0);

        if (grandTotal === 0) return null;

        const totalAngle = 2 * Math.PI - n * padAngle;
        const groups: { startAngle: number; endAngle: number; total: number; index: number }[] = [];

        let currentAngle = 0;
        groupTotals.forEach((total, i) => {
            const angle = (total / grandTotal) * totalAngle;
            groups.push({
                startAngle: currentAngle,
                endAngle: currentAngle + angle,
                total,
                index: i,
            });
            currentAngle += angle + padAngle;
        });

        // Compute chords
        const chords: {
            source: { index: number; startAngle: number; endAngle: number };
            target: { index: number; startAngle: number; endAngle: number };
            value: number;
        }[] = [];

        const groupCurrentAngle = groups.map(g => g.startAngle);

        for (let i = 0; i < n; i++) {
            for (let j = 0; j < n; j++) {
                const val = matrix[i][j];
                if (val <= 0) continue;

                const sourceAngleSize = (val / groupTotals[i]) * (groups[i].endAngle - groups[i].startAngle);
                const targetAngleSize = (val / groupTotals[j]) * (groups[j].endAngle - groups[j].startAngle);

                chords.push({
                    source: {
                        index: i,
                        startAngle: groupCurrentAngle[i],
                        endAngle: groupCurrentAngle[i] + sourceAngleSize,
                    },
                    target: {
                        index: j,
                        startAngle: groupCurrentAngle[j],
                        endAngle: groupCurrentAngle[j] + targetAngleSize,
                    },
                    value: val,
                });

                groupCurrentAngle[i] += sourceAngleSize;
                groupCurrentAngle[j] += targetAngleSize;
            }
        }

        return { groups, chords };
    }, [matrix, n, padAngle]);

    if (!layout) {
        return <div className={cn(className, 'flex items-center justify-center text-muted-foreground')}>No data</div>;
    }

    const cx = 200;
    const cy = 200;
    const viewSize = 400;

    return (
        <div className={cn(className, 'flex flex-col')}>
            {title && <h3 className="text-sm font-semibold text-center mb-2">{title}</h3>}
            <div className="flex-1 relative">
                <svg viewBox={`0 0 ${viewSize} ${viewSize}`} className="w-full h-full" preserveAspectRatio="xMidYMid meet">
                    {/* Chords (ribbons) */}
                    {layout.chords.map((chord, i) => {
                        const s1 = polarToCartesian(cx, cy, innerRadius, chord.source.startAngle);
                        const s2 = polarToCartesian(cx, cy, innerRadius, chord.source.endAngle);
                        const t1 = polarToCartesian(cx, cy, innerRadius, chord.target.startAngle);
                        const t2 = polarToCartesian(cx, cy, innerRadius, chord.target.endAngle);

                        const sLargeArc = chord.source.endAngle - chord.source.startAngle > Math.PI ? 1 : 0;
                        const tLargeArc = chord.target.endAngle - chord.target.startAngle > Math.PI ? 1 : 0;

                        const d = `M ${s1.x} ${s1.y}
                            A ${innerRadius} ${innerRadius} 0 ${sLargeArc} 1 ${s2.x} ${s2.y}
                            Q ${cx} ${cy} ${t1.x} ${t1.y}
                            A ${innerRadius} ${innerRadius} 0 ${tLargeArc} 1 ${t2.x} ${t2.y}
                            Q ${cx} ${cy} ${s1.x} ${s1.y} Z`;

                        return (
                            <path
                                key={`chord-${i}`}
                                d={d}
                                fill={effectiveColors[chord.source.index]}
                                opacity={0.5}
                                className="transition-opacity duration-200 hover:opacity-80 cursor-pointer"
                                onClick={() => onChordClick?.(
                                    names[chord.source.index],
                                    names[chord.target.index],
                                    chord.value,
                                )}
                            >
                                <title>{`${names[chord.source.index]} → ${names[chord.target.index]}: ${chord.value.toLocaleString()}`}</title>
                            </path>
                        );
                    })}

                    {/* Group arcs */}
                    {layout.groups.map((group, i) => {
                        const s = polarToCartesian(cx, cy, outerRadius, group.startAngle);
                        const e = polarToCartesian(cx, cy, outerRadius, group.endAngle);
                        const si = polarToCartesian(cx, cy, innerRadius, group.startAngle);
                        const ei = polarToCartesian(cx, cy, innerRadius, group.endAngle);
                        const largeArc = group.endAngle - group.startAngle > Math.PI ? 1 : 0;

                        const d = `M ${si.x} ${si.y}
                            A ${innerRadius} ${innerRadius} 0 ${largeArc} 1 ${ei.x} ${ei.y}
                            L ${e.x} ${e.y}
                            A ${outerRadius} ${outerRadius} 0 ${largeArc} 0 ${s.x} ${s.y} Z`;

                        const midAngle = (group.startAngle + group.endAngle) / 2;
                        const labelPos = polarToCartesian(cx, cy, outerRadius + 14, midAngle);
                        const textAnchor = midAngle > Math.PI / 2 && midAngle < (3 * Math.PI) / 2 ? 'end' : 'start';

                        return (
                            <g key={`group-${i}`}>
                                <path
                                    d={d}
                                    fill={effectiveColors[i]}
                                    stroke="white"
                                    strokeWidth={1}
                                >
                                    <title>{`${names[i]}: ${group.total.toLocaleString()}`}</title>
                                </path>
                                {group.endAngle - group.startAngle > 0.15 && (
                                    <text
                                        x={labelPos.x}
                                        y={labelPos.y}
                                        textAnchor={textAnchor}
                                        dominantBaseline="middle"
                                        fontSize={9}
                                        fontWeight={500}
                                        fill="currentColor"
                                        className="text-foreground"
                                    >
                                        {names[i]}
                                    </text>
                                )}
                            </g>
                        );
                    })}
                </svg>
            </div>
        </div>
    );
}
