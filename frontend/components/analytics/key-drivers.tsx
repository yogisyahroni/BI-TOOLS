import React from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import { CorrelationResult } from '@/types/analytics';
import { cn } from '@/lib/utils';
import { 
    ArrowRight, 
    TrendingUp, 
    TrendingDown, 
    Minus,
    Target,
    Zap,
    BarChart3
} from 'lucide-react';

interface KeyDriversProps {
    correlations: CorrelationResult[];
    isLoading?: boolean;
}

const strengthConfig = {
    Strong: {
        color: 'text-green-600',
        bgColor: 'bg-green-500',
        lightBg: 'bg-green-50',
        borderColor: 'border-green-200',
        label: 'Strong Relationship',
    },
    Moderate: {
        color: 'text-blue-600',
        bgColor: 'bg-blue-500',
        lightBg: 'bg-blue-50',
        borderColor: 'border-blue-200',
        label: 'Moderate Relationship',
    },
    Weak: {
        color: 'text-yellow-600',
        bgColor: 'bg-yellow-500',
        lightBg: 'bg-yellow-50',
        borderColor: 'border-yellow-200',
        label: 'Weak Relationship',
    },
};

export function KeyDrivers({ correlations, isLoading }: KeyDriversProps) {
    if (isLoading) {
        return (
            <div className="space-y-4">
                {[1, 2, 3, 4].map((i) => (
                    <div key={i} className="h-20 animate-pulse rounded-xl bg-muted/60" />
                ))}
            </div>
        );
    }

    // Filter for meaningful correlations and sort by strength
    const significant = correlations
        .filter(c => Math.abs(c.coefficient) > 0.15)
        .sort((a, b) => Math.abs(b.coefficient) - Math.abs(a.coefficient))
        .slice(0, 8); // Limit to top 8

    if (significant.length === 0) {
        return (
            <div className="flex flex-col items-center justify-center py-12 text-center">
                <div className="h-16 w-16 rounded-full bg-muted/50 flex items-center justify-center mb-4">
                    <BarChart3 className="h-8 w-8 text-muted-foreground" />
                </div>
                <h3 className="font-semibold text-lg mb-2">No Significant Correlations</h3>
                <p className="text-sm text-muted-foreground max-w-xs">
                    No strong relationships found between your metrics. Try analyzing different data ranges.
                </p>
            </div>
        );
    }

    return (
        <div className="space-y-4">
            {significant.map((correlation, index) => {
                const absCoefficient = Math.abs(correlation.coefficient);
                const strength = absCoefficient > 0.7 ? 'Strong' : absCoefficient > 0.4 ? 'Moderate' : 'Weak';
                const config = strengthConfig[strength];
                const isPositive = correlation.coefficient > 0;

                return (
                    <div
                        key={index}
                        className={cn(
                            "group relative rounded-xl border p-4 transition-all duration-200",
                            "hover:shadow-md hover:-translate-y-0.5",
                            config.borderColor,
                            "bg-card/50"
                        )}
                    >
                        {/* Header */}
                        <div className="flex items-center justify-between mb-3">
                            <div className="flex items-center gap-3">
                                <div className={cn(
                                    "h-10 w-10 rounded-lg flex items-center justify-center",
                                    config.lightBg
                                )}>
                                    {isPositive ? (
                                        <TrendingUp className={cn("h-5 w-5", config.color)} />
                                    ) : (
                                        <TrendingDown className={cn("h-5 w-5", config.color)} />
                                    )}
                                </div>
                                <div>
                                    <div className="flex items-center gap-2">
                                        <span className="font-semibold text-sm">{correlation.variableA}</span>
                                        <ArrowRight className="h-3 w-3 text-muted-foreground" />
                                        <span className="font-semibold text-sm">{correlation.variableB}</span>
                                    </div>
                                    <Badge 
                                        variant="outline" 
                                        className={cn("mt-1 text-xs", config.color)}
                                    >
                                        {config.label}
                                    </Badge>
                                </div>
                            </div>
                            <div className="text-right">
                                <div className={cn("text-2xl font-bold", config.color)}>
                                    {correlation.coefficient > 0 ? '+' : ''}{correlation.coefficient.toFixed(2)}
                                </div>
                                <div className="text-xs text-muted-foreground">
                                    correlation
                                </div>
                            </div>
                        </div>

                        {/* Progress Bar */}
                        <div className="space-y-2">
                            <div className="flex items-center justify-between text-xs">
                                <span className="text-muted-foreground">Relationship Strength</span>
                                <span className={cn("font-medium", config.color)}>
                                    {Math.round(absCoefficient * 100)}%
                                </span>
                            </div>
                            <div className="h-2.5 bg-muted rounded-full overflow-hidden">
                                <div
                                    className={cn(
                                        "h-full rounded-full transition-all duration-700 ease-out",
                                        config.bgColor
                                    )}
                                    style={{ width: `${absCoefficient * 100}%` }}
                                />
                            </div>
                        </div>

                        {/* Interpretation */}
                        <div className="mt-3 pt-3 border-t border-dashed">
                            <p className="text-xs text-muted-foreground">
                                {isPositive 
                                    ? `When ${correlation.variableA} increases, ${correlation.variableB} tends to increase as well.`
                                    : `When ${correlation.variableA} increases, ${correlation.variableB} tends to decrease.`
                                }
                            </p>
                        </div>
                    </div>
                );
            })}
        </div>
    );
}
