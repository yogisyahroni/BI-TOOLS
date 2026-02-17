'use client';

import React, { useMemo } from 'react';
import { cn } from '@/lib/utils';

/**
 * TASK-CHART-015: Word Cloud
 * Weighted text visualization using pure SVG to avoid external dependencies.
 */

interface WordCloudItem {
    text: string;
    weight: number;
    color?: string;
}

interface WordCloudChartProps {
    data: WordCloudItem[];
    title?: string;
    maxFontSize?: number;
    minFontSize?: number;
    colors?: string[];
    className?: string;
    onWordClick?: (word: string, weight: number) => void;
}

const CLOUD_COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899', '#06b6d4', '#f97316', '#84cc16', '#6366f1', '#14b8a6', '#a855f7'];

export function WordCloudChart({
    data,
    title,
    maxFontSize = 48,
    minFontSize = 12,
    colors = CLOUD_COLORS,
    className = 'h-full w-full min-h-[300px]',
    onWordClick,
}: WordCloudChartProps) {
    const processed = useMemo(() => {
        if (data.length === 0) return [];
        const sorted = [...data].sort((a, b) => b.weight - a.weight);
        const maxWeight = sorted[0].weight;
        const minWeight = sorted[sorted.length - 1].weight;
        const range = maxWeight - minWeight || 1;

        return sorted.map((item, i) => {
            const normalised = (item.weight - minWeight) / range;
            const fontSize = minFontSize + normalised * (maxFontSize - minFontSize);
            const color = item.color ?? colors[i % colors.length];
            const rotation = [0, 0, 0, -15, 15, -30, 30][i % 7];
            const opacity = 0.6 + normalised * 0.4;

            return {
                ...item,
                fontSize,
                color,
                rotation,
                opacity,
            };
        });
    }, [data, maxFontSize, minFontSize, colors]);

    // Spiral placement layout
    const positions = useMemo(() => {
        const placed: { x: number; y: number; w: number; h: number }[] = [];
        const containerW = 600;
        const containerH = 400;
        const centerX = containerW / 2;
        const centerY = containerH / 2;

        return processed.map((p, i) => {
            const estWidth = p.text.length * p.fontSize * 0.55;
            const estHeight = p.fontSize * 1.3;

            // Archimedean spiral placement
            let angle = i * 0.8;
            let radius = 0;
            let x = centerX;
            let y = centerY;
            let attempts = 0;

            while (attempts < 200) {
                x = centerX + radius * Math.cos(angle) - estWidth / 2;
                y = centerY + radius * Math.sin(angle) - estHeight / 2;

                // Bounds check
                if (x < 5 || x + estWidth > containerW - 5 || y < 5 || y + estHeight > containerH - 5) {
                    angle += 0.3;
                    radius += 2;
                    attempts++;
                    continue;
                }

                // Collision check
                const collides = placed.some(rect =>
                    x < rect.x + rect.w &&
                    x + estWidth > rect.x &&
                    y < rect.y + rect.h &&
                    y + estHeight > rect.y,
                );

                if (!collides) break;

                angle += 0.3;
                radius += 1.5;
                attempts++;
            }

            placed.push({ x, y, w: estWidth, h: estHeight });
            return { x, y };
        });
    }, [processed]);

    return (
        <div className={cn(className, 'flex flex-col')}>
            {title && <h3 className="text-sm font-semibold text-center mb-2">{title}</h3>}
            <div className="flex-1 relative">
                <svg viewBox="0 0 600 400" className="w-full h-full" preserveAspectRatio="xMidYMid meet">
                    {processed.map((word, i) => (
                        <text
                            key={`${word.text}-${i}`}
                            x={positions[i].x + (word.text.length * word.fontSize * 0.55) / 2}
                            y={positions[i].y + word.fontSize * 0.9}
                            fontSize={word.fontSize}
                            fill={word.color}
                            opacity={word.opacity}
                            fontWeight={word.fontSize > maxFontSize * 0.6 ? 700 : 500}
                            textAnchor="middle"
                            transform={`rotate(${word.rotation}, ${positions[i].x + (word.text.length * word.fontSize * 0.55) / 2}, ${positions[i].y + word.fontSize * 0.5})`}
                            className={cn(
                                'select-none transition-all duration-200',
                                onWordClick && 'cursor-pointer hover:opacity-100',
                            )}
                            onClick={() => onWordClick?.(word.text, word.weight)}
                        >
                            {word.text}
                        </text>
                    ))}
                </svg>
            </div>
        </div>
    );
}
