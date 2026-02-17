import React from 'react';
import { _Card, _CardContent, _CardHeader, _CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { 
    ArrowUpRight, 
    ArrowDownRight, 
    AlertTriangle, 
    Info, 
    TrendingUp,
    Brain,
    Target,
    _Zap,
    BarChart3,
    _Lightbulb,
    _CheckCircle2
} from 'lucide-react';
import { type Insight } from '@/types/analytics';
import { cn } from '@/lib/utils';

interface AutoInsightsProps {
    insights: Insight[];
    isLoading?: boolean;
}

const insightTypeConfig = {
    trend: {
        icon: TrendingUp,
        label: 'Trend',
        color: 'text-blue-500',
        bgColor: 'bg-blue-500/10',
        borderColor: 'border-blue-500/20',
    },
    anomaly: {
        icon: AlertTriangle,
        label: 'Anomaly',
        color: 'text-amber-500',
        bgColor: 'bg-amber-500/10',
        borderColor: 'border-amber-500/20',
    },
    correlation: {
        icon: Target,
        label: 'Correlation',
        color: 'text-purple-500',
        bgColor: 'bg-purple-500/10',
        borderColor: 'border-purple-500/20',
    },
    forecast: {
        icon: Brain,
        label: 'Forecast',
        color: 'text-green-500',
        bgColor: 'bg-green-500/10',
        borderColor: 'border-green-500/20',
    },
    statistic: {
        icon: BarChart3,
        label: 'Statistic',
        color: 'text-cyan-500',
        bgColor: 'bg-cyan-500/10',
        borderColor: 'border-cyan-500/20',
    },
};

export function AutoInsights({ insights, isLoading }: AutoInsightsProps) {
    if (isLoading) {
        return (
            <div className="space-y-3">
                {[1, 2, 3].map((i) => (
                    <div key={i} className="h-28 animate-pulse rounded-xl bg-muted/60" />
                ))}
            </div>
        );
    }

    if (insights.length === 0) {
        return (
            <div className="flex flex-col items-center justify-center py-12 text-center">
                <div className="h-16 w-16 rounded-full bg-muted/50 flex items-center justify-center mb-4">
                    <Brain className="h-8 w-8 text-muted-foreground" />
                </div>
                <h3 className="font-semibold text-lg mb-2">No Insights Yet</h3>
                <p className="text-sm text-muted-foreground max-w-xs">
                    AI is analyzing your data. Insights will appear here once the analysis is complete.
                </p>
            </div>
        );
    }

    return (
        <div className="space-y-3">
            {insights.map((insight, index) => {
                const config = insightTypeConfig[insight.type as keyof typeof insightTypeConfig] || insightTypeConfig.statistic;
                const Icon = config.icon;
                
                return (
                    <div
                        key={insight.id || index}
                        className={cn(
                            "group relative flex flex-col gap-3 rounded-xl border p-4 transition-all duration-200",
                            "hover:shadow-md hover:-translate-y-0.5",
                            config.borderColor,
                            "bg-card/50 backdrop-blur-sm"
                        )}
                    >
                        {/* Header */}
                        <div className="flex items-start justify-between gap-3">
                            <div className="flex items-center gap-3">
                                <div className={cn(
                                    "h-10 w-10 rounded-lg flex items-center justify-center shrink-0",
                                    config.bgColor
                                )}>
                                    <Icon className={cn("h-5 w-5", config.color)} />
                                </div>
                                <div className="min-w-0 flex-1">
                                    <h4 className="font-semibold text-sm leading-tight line-clamp-2">
                                        {insight.title}
                                    </h4>
                                    <Badge 
                                        variant="secondary" 
                                        className="mt-1 text-xs font-normal"
                                    >
                                        {config.label}
                                    </Badge>
                                </div>
                            </div>
                            {getTrendIndicator(insight)}
                        </div>

                        {/* Description */}
                        <p className="text-sm text-muted-foreground leading-relaxed">
                            {insight.description}
                        </p>

                        {/* Value Display */}
                        {insight.value !== undefined && (
                            <div className="flex items-center gap-2 pt-2 border-t border-dashed">
                                <span className="text-xs text-muted-foreground">Value:</span>
                                <code className="text-xs font-mono bg-muted px-2 py-0.5 rounded">
                                    {typeof insight.value === 'number' 
                                        ? insight.value.toLocaleString(undefined, { maximumFractionDigits: 2 })
                                        : JSON.stringify(insight.value)
                                    }
                                </code>
                            </div>
                        )}

                        {/* Confidence Indicator */}
                        {insight.confidence !== undefined && (
                            <div className="flex items-center gap-2">
                                <div className="flex-1 h-1.5 bg-muted rounded-full overflow-hidden">
                                    <div 
                                        className={cn(
                                            "h-full rounded-full transition-all duration-500",
                                            insight.confidence > 0.8 ? "bg-green-500" :
                                            insight.confidence > 0.6 ? "bg-yellow-500" : "bg-red-500"
                                        )}
                                        style={{ width: `${insight.confidence * 100}%` }}
                                    />
                                </div>
                                <span className="text-xs text-muted-foreground">
                                    {Math.round(insight.confidence * 100)}%
                                </span>
                            </div>
                        )}
                    </div>
                );
            })}
        </div>
    );
}

function getTrendIndicator(insight: Insight) {
    const title = insight.title.toLowerCase();
    
    if (title.includes('up') || title.includes('increase') || title.includes('growth')) {
        return (
            <div className="flex items-center gap-1 text-green-500 shrink-0">
                <ArrowUpRight className="h-4 w-4" />
                <span className="text-xs font-medium">Positive</span>
            </div>
        );
    }
    
    if (title.includes('down') || title.includes('decrease') || title.includes('decline')) {
        return (
            <div className="flex items-center gap-1 text-red-500 shrink-0">
                <ArrowDownRight className="h-4 w-4" />
                <span className="text-xs font-medium">Negative</span>
            </div>
        );
    }
    
    if (insight.type === 'anomaly') {
        return (
            <div className="flex items-center gap-1 text-amber-500 shrink-0">
                <AlertTriangle className="h-4 w-4" />
                <span className="text-xs font-medium">Alert</span>
            </div>
        );
    }
    
    return (
        <div className="flex items-center gap-1 text-blue-500 shrink-0">
            <Info className="h-4 w-4" />
            <span className="text-xs font-medium">Info</span>
        </div>
    );
}
