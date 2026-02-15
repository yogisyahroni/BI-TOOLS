'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Loader2, Sparkles, AlertTriangle } from 'lucide-react';
import { aiApi } from '@/lib/api/ai';
import { toast } from 'sonner';
import { QueryOptimizerSuggestions } from './suggestions';
import { AIOptimizationResponse } from '@/lib/types/ai';

interface OptimizerPanelProps {
    query: string;
    databaseType: string;
    schemaContext?: any;
    onApplySuggestion?: (suggestion: any) => void;
}

export function OptimizerPanel({ query, databaseType, schemaContext, onApplySuggestion }: OptimizerPanelProps) {
    const [isOptimizing, setIsOptimizing] = useState(false);
    const [result, setResult] = useState<AIOptimizationResponse | null>(null);
    const [mappedAnalysis, setMappedAnalysis] = useState<any | null>(null);

    const handleOptimize = async () => {
        if (!query.trim()) {
            toast.error('Please enter a query to optimize');
            return;
        }

        setIsOptimizing(true);
        setResult(null);
        setMappedAnalysis(null);

        try {
            const response = await aiApi.optimize({
                query,
                databaseType,
                schemaContext
            });

            setResult(response);

            // Map API response to the format expected by suggestions.tsx
            // This is an adapter to reuse the rich UI of suggestions.tsx
            const mapped = {
                query: query,
                staticAnalysis: {
                    query: query,
                    suggestions: response.suggestions.map(s => ({
                        type: s.type,
                        severity: s.confidence > 0.8 ? 'High' : s.confidence > 0.5 ? 'Medium' : 'Low',
                        title: s.title,
                        description: s.description,
                        original: 'See query', // Placeholder as backend didn't return exact snippet
                        optimized: s.sqlAction || 'See description',
                        impact: s.estimatedImpact,
                        example: s.rationale
                    })),
                    performanceScore: Math.round(response.suggestions.reduce((acc, s) => acc - (s.confidence * 10), 100)), // dynamic score
                    complexityLevel: 'Moderate', // Placeholder
                    estimatedImprovement: 'Variable'
                },
                planAnalysis: null, // We don't have execution plan analysis from this specific API yet
                explainAvailable: false
            };

            setMappedAnalysis(mapped);
            toast.success('Optimization complete');

        } catch (error: any) {
            toast.error('Optimization failed', {
                description: error.message
            });
        } finally {
            setIsOptimizing(false);
        }
    };

    return (
        <div className="space-y-4">
            <Card>
                <CardHeader className="pb-3">
                    <div className="flex items-center justify-between">
                        <div className="space-y-1">
                            <CardTitle className="text-base flex items-center gap-2">
                                <Sparkles className="w-4 h-4 text-primary" />
                                AI Query Optimizer
                            </CardTitle>
                            <CardDescription>
                                Analyze your query for performance improvements and best practices.
                            </CardDescription>
                        </div>
                        <Button
                            onClick={handleOptimize}
                            disabled={isOptimizing || !query.trim()}
                            className="gap-2"
                        >
                            {isOptimizing ? (
                                <>
                                    <Loader2 className="w-4 h-4 animate-spin" />
                                    Analyzing...
                                </>
                            ) : (
                                <>
                                    <Sparkles className="w-4 h-4" />
                                    Optimize Query
                                </>
                            )}
                        </Button>
                    </div>
                </CardHeader>
            </Card>

            {mappedAnalysis && (
                <div className="animate-in fade-in-50 slide-in-from-bottom-2 duration-500">
                    <QueryOptimizerSuggestions
                        analysis={mappedAnalysis}
                        onApplySuggestion={onApplySuggestion}
                    />
                </div>
            )}

            {!mappedAnalysis && !isOptimizing && (
                <div className="text-center py-10 text-muted-foreground bg-muted/20 rounded-lg border border-dashed">
                    <Sparkles className="w-10 h-10 mx-auto mb-3 opacity-20" />
                    <p>Click "Optimize Query" to get AI-powered suggestions.</p>
                </div>
            )}
        </div>
    );
}
