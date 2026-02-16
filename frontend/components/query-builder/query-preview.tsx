'use client';

import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Play, Save, Copy, CheckCircle2, AlertCircle, Loader2, Sparkles } from 'lucide-react';
import { SQLGenerator } from '@/lib/query-builder/sql-generator';
import { type QueryBuilderState } from '@/lib/query-builder/types';
import { type QueryResult } from '@/types/visual-query';
import { toast } from 'sonner';
import { fetchWithAuth } from '@/lib/utils';
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from '@/components/ui/dialog';
import { OptimizerPanel } from '@/components/query-optimizer/optimizer-panel';

interface QueryPreviewProps {
    state: QueryBuilderState;
    onSave?: (sql: string) => void;
    databaseType?: string;
}

export function QueryPreview({ state, onSave, databaseType }: QueryPreviewProps) {
    const [sql, setSql] = useState('');
    const [isExecuting, setIsExecuting] = useState(false);
    const [result, setResult] = useState<QueryResult | null>(null);
    const [error, setError] = useState<string | null>(null);
    const [copied, setCopied] = useState(false);
    const [showOptimizer, setShowOptimizer] = useState(false);

    useEffect(() => {
        try {
            if (state.table && state.columns.length > 0) {
                const generated = SQLGenerator.generate(state);
                const validation = SQLGenerator.validate(generated);

                if (validation.valid) {
                    setSql(generated);
                    setError(null);
                } else {
                    setSql(generated);
                    setError(validation.error || 'Invalid SQL');
                }
            } else {
                setSql('');
                setError(null);
            }
        } catch (err) {
            const message = err instanceof Error ? err.message : 'Failed to generate SQL';
            setError(message);
            setSql('');
        }
    }, [state]);

    const handleExecute = async (cursor?: string) => {
        if (!sql || error) return;

        try {
            setIsExecuting(true);
            setError(null);

            const res = await fetchWithAuth('/api/go/queries/execute', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    sql,
                    connectionId: state.connectionId,
                    page: 1, // Start page (for offset compatibility)
                    pageSize: 100, // Or whatever limit
                    cursor: cursor, // Send cursor if available
                    config: state, // Send visual query config for complexity analysis
                }),
            });

            const data = await res.json();

            if (data.success) {
                if (cursor) {
                    // Append mode
                    setResult((prev: QueryResult | null) => {
                        if (!prev) return data as QueryResult;
                        return {
                            ...data,
                            rows: [...prev.rows, ...data.rows],
                            rowCount: (prev.rowCount || 0) + (data.rowCount || 0),
                        } as QueryResult;
                    });
                } else {
                    // Replace mode
                    setResult(data as QueryResult);
                }
                toast.success(`Query executed successfully (${data.rowCount} rows)`);
            } else {
                setError(data.error || 'Query execution failed');
                toast.error(data.error || 'Query execution failed');
            }
        } catch (err) {
            const message = err instanceof Error ? err.message : 'Unknown error';
            setError(message);
            toast.error(message);
        } finally {
            setIsExecuting(false);
        }
    };

    const handleLoadMore = () => {
        if (result?.nextCursor) {
            handleExecute(result.nextCursor);
        }
    };

    const handleCopy = async () => {
        if (!sql) return;

        try {
            await navigator.clipboard.writeText(sql);
            setCopied(true);
            toast.success('SQL copied to clipboard');
            setTimeout(() => setCopied(false), 2000);
        } catch (err) {
            toast.error('Failed to copy SQL');
        }
    };

    return (
        <div className="space-y-4">
            {/* SQL Preview Card */}
            <Card>
                <CardHeader>
                    <div className="flex items-center justify-between">
                        <CardTitle className="text-base">SQL Preview</CardTitle>
                        <div className="flex items-center gap-2">
                            <Dialog open={showOptimizer} onOpenChange={setShowOptimizer}>
                                <DialogTrigger asChild>
                                    <Button
                                        variant="outline"
                                        size="sm"
                                        disabled={!sql || !!error || !databaseType}
                                        className="gap-2"
                                    >
                                        <Sparkles className="h-4 w-4 text-amber-500" />
                                        Optimize
                                    </Button>
                                </DialogTrigger>
                                <DialogContent className="max-w-3xl max-h-[90vh] overflow-auto">
                                    <DialogHeader>
                                        <DialogTitle>Optimize Query</DialogTitle>
                                    </DialogHeader>
                                    <OptimizerPanel
                                        query={sql}
                                        databaseType={databaseType || 'postgresql'}
                                        onApplySuggestion={(s) => {
                                            // If we wanted to apply suggestion to builder logic, it would be complex
                                            // For now, maybe just toast or copy optimized SQL
                                            if (s.optimized) {
                                                navigator.clipboard.writeText(s.optimized);
                                                toast.success("Optimized SQL copied to clipboard");
                                                setShowOptimizer(false);
                                            }
                                        }}
                                    />
                                </DialogContent>
                            </Dialog>

                            <Button
                                variant="outline"
                                size="sm"
                                onClick={() => onSave && sql && onSave(sql)}
                                disabled={!sql || !onSave}
                            >
                                <Save className="h-4 w-4 mr-2" />
                                Save
                            </Button>
                            <Button
                                variant="outline"
                                size="sm"
                                onClick={handleCopy}
                                disabled={!sql}
                            >
                                {copied ? (
                                    <CheckCircle2 className="h-4 w-4 mr-2" />
                                ) : (
                                    <Copy className="h-4 w-4 mr-2" />
                                )}
                                {copied ? 'Copied!' : 'Copy'}
                            </Button>
                            <Button
                                variant="default"
                                size="sm"
                                onClick={() => handleExecute()}
                                disabled={!sql || !!error || isExecuting}
                            >
                                {isExecuting ? (
                                    <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                                ) : (
                                    <Play className="h-4 w-4 mr-2" />
                                )}
                                Run Query
                            </Button>
                        </div>
                    </div>
                </CardHeader>
                <CardContent>
                    {!sql ? (
                        <p className="text-sm text-muted-foreground text-center py-8">
                            Select a table and columns to generate SQL
                        </p>
                    ) : (
                        <pre className="bg-muted p-4 rounded-md overflow-x-auto text-sm font-mono">
                            {sql}
                        </pre>
                    )}

                    {error && (
                        <Alert variant="destructive" className="mt-4">
                            <AlertCircle className="h-4 w-4" />
                            <AlertDescription>{error}</AlertDescription>
                        </Alert>
                    )}
                </CardContent>
            </Card>

            {/* Results Card */}
            {result && (
                <Card>
                    <CardHeader>
                        <CardTitle className="text-base flex justify-between items-center">
                            <span>
                                Results ({result.rows?.length || 0} rows loaded)
                            </span>
                            {result.executionTime && (
                                <span className="text-sm font-normal text-muted-foreground">
                                    {result.executionTime}ms
                                </span>
                            )}
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="overflow-x-auto max-h-[500px]">
                            <table className="w-full text-sm">
                                <thead className="sticky top-0 bg-background">
                                    <tr className="border-b">
                                        {result.columns?.map((col: string) => (
                                            <th
                                                key={col}
                                                className="text-left p-2 font-medium text-muted-foreground"
                                            >
                                                {col}
                                            </th>
                                        ))}
                                    </tr>
                                </thead>
                                <tbody>
                                    {result.rows?.map((row: any[], index: number) => (
                                        <tr key={index} className="border-b last:border-0 hover:bg-muted/50">
                                            {row.map((cell: any, cellIndex: number) => (
                                                <td key={cellIndex} className="p-2 whitespace-nowrap">
                                                    {cell !== null && cell !== undefined
                                                        ? String(cell)
                                                        : <span className="text-muted-foreground italic">null</span>}
                                                </td>
                                            ))}
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>

                        {result.nextCursor && (
                            <div className="mt-4 flex justify-center">
                                <Button
                                    variant="outline"
                                    onClick={handleLoadMore}
                                    disabled={isExecuting}
                                >
                                    {isExecuting ? (
                                        <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                                    ) : null}
                                    Load More
                                </Button>
                            </div>
                        )}

                        {!result.nextCursor && result.rows?.length > 0 && (
                            <p className="text-xs text-muted-foreground text-center mt-4">
                                End of results
                            </p>
                        )}
                    </CardContent>
                </Card>
            )}
        </div>
    );
}
