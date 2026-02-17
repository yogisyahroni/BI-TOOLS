'use client';

import React, { useMemo } from 'react';
import { cn } from '@/lib/utils';
import { Card } from '@/components/ui/card';
import { TrendingUp, TrendingDown, Minus, Target } from 'lucide-react';

/**
 * TASK-CHART-002: KPI Card / Big Number
 * Executive dashboard headline metrics with trend, sparkline, and goal tracking.
 */

interface KPICardProps {
    value: number | string;
    label: string;
    prefix?: string;
    suffix?: string;
    previousValue?: number;
    target?: number;
    sparklineData?: number[];
    format?: 'number' | 'currency' | 'percent' | 'compact';
    currencyCode?: string;
    decimals?: number;
    thresholds?: { danger: number; warning: number; success: number };
    invertThreshold?: boolean;
    className?: string;
    size?: 'sm' | 'md' | 'lg';
    onClick?: () => void;
}

function formatValue(
    value: number | string,
    format: string,
    currencyCode: string,
    decimals: number,
    prefix: string,
    suffix: string,
): string {
    if (typeof value === 'string') return `${prefix}${value}${suffix}`;
    const numVal = value as number;
    let formatted: string;

    switch (format) {
        case 'currency':
            formatted = new Intl.NumberFormat('en-US', {
                style: 'currency',
                currency: currencyCode,
                minimumFractionDigits: decimals,
                maximumFractionDigits: decimals,
            }).format(numVal);
            break;
        case 'percent':
            formatted = `${(numVal * 100).toFixed(decimals)}%`;
            break;
        case 'compact':
            if (Math.abs(numVal) >= 1e9) formatted = `${(numVal / 1e9).toFixed(1)}B`;
            else if (Math.abs(numVal) >= 1e6) formatted = `${(numVal / 1e6).toFixed(1)}M`;
            else if (Math.abs(numVal) >= 1e3) formatted = `${(numVal / 1e3).toFixed(1)}K`;
            else formatted = numVal.toFixed(decimals);
            break;
        default:
            formatted = numVal.toLocaleString('en-US', {
                minimumFractionDigits: decimals,
                maximumFractionDigits: decimals,
            });
    }
    return `${prefix}${formatted}${suffix}`;
}

function MiniSparkline({ data, color }: { data: number[]; color: string }) {
    if (data.length < 2) return null;
    const minVal = Math.min(...data);
    const maxVal = Math.max(...data);
    const range = maxVal - minVal || 1;
    const width = 80;
    const height = 28;
    const padding = 2;

    const points = data.map((val, i) => {
        const x = padding + (i / (data.length - 1)) * (width - padding * 2);
        const y = height - padding - ((val - minVal) / range) * (height - padding * 2);
        return `${x},${y}`;
    });

    return (
        <svg width={width} height={height} className="flex-shrink-0">
            <polyline
                points={points.join(' ')}
                fill="none"
                stroke={color}
                strokeWidth={1.5}
                strokeLinecap="round"
                strokeLinejoin="round"
            />
            {/* End dot */}
            {(() => {
                const lastPoint = points[points.length - 1].split(',');
                return (
                    <circle
                        cx={parseFloat(lastPoint[0])}
                        cy={parseFloat(lastPoint[1])}
                        r={2.5}
                        fill={color}
                    />
                );
            })()}
        </svg>
    );
}

function GoalProgress({ current, target }: { current: number; target: number }) {
    const progress = Math.min((current / target) * 100, 100);
    const isComplete = progress >= 100;

    return (
        <div className="mt-2 space-y-1">
            <div className="flex items-center justify-between text-[10px]">
                <div className="flex items-center gap-1 text-muted-foreground">
                    <Target className="w-3 h-3" />
                    <span>Goal: {target.toLocaleString()}</span>
                </div>
                <span className={cn(
                    'font-medium',
                    isComplete ? 'text-emerald-500' : 'text-muted-foreground',
                )}>
                    {progress.toFixed(0)}%
                </span>
            </div>
            <div className="h-1.5 bg-muted rounded-full overflow-hidden">
                <div
                    className={cn(
                        'h-full rounded-full transition-all duration-500',
                        isComplete ? 'bg-emerald-500' : 'bg-primary',
                    )}
                    style={{ width: `${progress}%` }}
                />
            </div>
        </div>
    );
}

export function KPICard({
    value,
    label,
    prefix = '',
    suffix = '',
    previousValue,
    target,
    sparklineData,
    format = 'number',
    currencyCode = 'USD',
    decimals = 0,
    thresholds,
    invertThreshold = false,
    className,
    size = 'md',
    onClick,
}: KPICardProps) {
    const delta = useMemo(() => {
        if (previousValue === undefined || typeof value !== 'number') return null;
        if (previousValue === 0) return value > 0 ? 100 : 0;
        return ((value - previousValue) / Math.abs(previousValue)) * 100;
    }, [value, previousValue]);

    const deltaColor = useMemo(() => {
        if (delta === null) return 'text-muted-foreground';
        if (delta === 0) return 'text-muted-foreground';
        const isPositive = delta > 0;
        if (invertThreshold) return isPositive ? 'text-red-500' : 'text-emerald-500';
        return isPositive ? 'text-emerald-500' : 'text-red-500';
    }, [delta, invertThreshold]);

    const thresholdColor = useMemo(() => {
        if (!thresholds || typeof value !== 'number') return undefined;
        const numVal = value as number;
        if (invertThreshold) {
            if (numVal >= thresholds.danger) return 'border-l-red-500';
            if (numVal >= thresholds.warning) return 'border-l-amber-500';
            return 'border-l-emerald-500';
        }
        if (numVal >= thresholds.success) return 'border-l-emerald-500';
        if (numVal >= thresholds.warning) return 'border-l-amber-500';
        return 'border-l-red-500';
    }, [value, thresholds, invertThreshold]);

    const sparkColor = delta !== null && delta >= 0 ? '#10b981' : '#ef4444';

    const sizeClasses = {
        sm: 'p-3',
        md: 'p-4',
        lg: 'p-6',
    };

    const valueSize = {
        sm: 'text-xl',
        md: 'text-3xl',
        lg: 'text-5xl',
    };

    const displayValue = formatValue(value, format, currencyCode, decimals, prefix, suffix);

    return (
        <Card
            className={cn(
                sizeClasses[size],
                'bg-card/70 backdrop-blur-sm border-border/30 transition-all duration-200',
                thresholdColor && 'border-l-4',
                thresholdColor,
                onClick && 'cursor-pointer hover:shadow-md hover:scale-[1.01] active:scale-[0.99]',
                className,
            )}
            onClick={onClick}
        >
            <div className="flex items-start justify-between gap-3">
                <div className="flex-1 min-w-0">
                    <p className="text-xs font-medium text-muted-foreground truncate mb-1">{label}</p>
                    <p className={cn(valueSize[size], 'font-bold tracking-tight leading-none')}>
                        {displayValue}
                    </p>

                    {/* Delta */}
                    {delta !== null && (
                        <div className={cn('flex items-center gap-1 mt-1.5', deltaColor)}>
                            {delta > 0 ? (
                                <TrendingUp className="w-3.5 h-3.5" />
                            ) : delta < 0 ? (
                                <TrendingDown className="w-3.5 h-3.5" />
                            ) : (
                                <Minus className="w-3.5 h-3.5" />
                            )}
                            <span className="text-xs font-semibold">
                                {delta > 0 ? '+' : ''}{delta.toFixed(1)}%
                            </span>
                            <span className="text-[10px] text-muted-foreground/60">vs prev</span>
                        </div>
                    )}
                </div>

                {/* Sparkline */}
                {sparklineData && sparklineData.length >= 2 && (
                    <MiniSparkline data={sparklineData} color={sparkColor} />
                )}
            </div>

            {/* Goal Progress */}
            {target !== undefined && typeof value === 'number' && (
                <GoalProgress current={value} target={target} />
            )}
        </Card>
    );
}
