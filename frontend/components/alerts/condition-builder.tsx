'use client';

import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select';
import { Button } from '@/components/ui/button';
import { Play, CheckCircle2, XCircle } from 'lucide-react';
import type { AlertOperator } from '@/types/alerts';
import { ALERT_OPERATORS } from '@/types/alerts';
import { useState } from 'react';

interface ConditionBuilderProps {
    column: string;
    operator: AlertOperator;
    threshold: number;
    availableColumns?: string[];
    onChange: (values: { column: string; operator: AlertOperator; threshold: number }) => void;
    disabled?: boolean;
    onTest?: () => Promise<{ success: boolean; message: string; value?: number }>;
}

export function ConditionBuilder({
    column,
    operator,
    threshold,
    availableColumns,
    onChange,
    disabled,
    onTest,
}: ConditionBuilderProps) {
    const [testResult, setTestResult] = useState<{
        success: boolean;
        message: string;
        value?: number;
    } | null>(null);
    const [isTesting, setIsTesting] = useState(false);

    const handleTest = async () => {
        if (!onTest) return;
        setIsTesting(true);
        try {
            const result = await onTest();
            setTestResult(result);
        } catch (error) {
            setTestResult({
                success: false,
                message: error instanceof Error ? error.message : 'Test failed',
            });
        } finally {
            setIsTesting(false);
        }
    };

    return (
        <div className="space-y-4">
            <div className="grid grid-cols-3 gap-4">
                {/* Column Selector */}
                <div className="space-y-2">
                    <Label htmlFor="column">Column</Label>
                    {availableColumns && availableColumns.length > 0 ? (
                        <Select
                            value={column}
                            onValueChange={(v) => onChange({ column: v, operator, threshold })}
                            disabled={disabled}
                        >
                            <SelectTrigger id="column">
                                <SelectValue placeholder="Select column" />
                            </SelectTrigger>
                            <SelectContent>
                                {availableColumns.map((col) => (
                                    <SelectItem key={col} value={col}>
                                        {col}
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        </Select>
                    ) : (
                        <Input
                            id="column"
                            value={column}
                            onChange={(e) =>
                                onChange({ column: e.target.value, operator, threshold })
                            }
                            placeholder="Enter column name"
                            disabled={disabled}
                        />
                    )}
                </div>

                {/* Operator Selector */}
                <div className="space-y-2">
                    <Label htmlFor="operator">Operator</Label>
                    <Select
                        value={operator}
                        onValueChange={(v) =>
                            onChange({ column, operator: v as AlertOperator, threshold })
                        }
                        disabled={disabled}
                    >
                        <SelectTrigger id="operator">
                            <SelectValue placeholder="Select operator" />
                        </SelectTrigger>
                        <SelectContent>
                            {ALERT_OPERATORS.map((op) => (
                                <SelectItem key={op.value} value={op.value}>
                                    {op.label}
                                </SelectItem>
                            ))}
                        </SelectContent>
                    </Select>
                </div>

                {/* Threshold Input */}
                <div className="space-y-2">
                    <Label htmlFor="threshold">Threshold</Label>
                    <Input
                        id="threshold"
                        type="number"
                        step="any"
                        value={threshold}
                        onChange={(e) =>
                            onChange({
                                column,
                                operator,
                                threshold: parseFloat(e.target.value) || 0,
                            })
                        }
                        placeholder="Enter threshold value"
                        disabled={disabled}
                    />
                </div>
            </div>

            {/* Preview */}
            <div className="bg-gray-50 rounded-lg p-3 border">
                <div className="flex items-center justify-between">
                    <div className="text-sm">
                        <span className="text-gray-500">Condition:</span>{' '}
                        <span className="font-mono font-medium">
                            {column || '...'} {operator} {threshold}
                        </span>
                    </div>
                    {onTest && (
                        <Button
                            variant="outline"
                            size="sm"
                            onClick={handleTest}
                            disabled={isTesting || disabled || !column}
                        >
                            {isTesting ? (
                                <span className="animate-spin mr-2">⟳</span>
                            ) : (
                                <Play className="h-3 w-3 mr-2" />
                            )}
                            Test
                        </Button>
                    )}
                </div>
            </div>

            {/* Test Result */}
            {testResult && (
                <div
                    className={`rounded-lg p-3 flex items-center gap-2 ${
                        testResult.success
                            ? 'bg-green-50 border border-green-200 text-green-800'
                            : 'bg-red-50 border border-red-200 text-red-800'
                    }`}
                >
                    {testResult.success ? (
                        <CheckCircle2 className="h-4 w-4" />
                    ) : (
                        <XCircle className="h-4 w-4" />
                    )}
                    <div className="flex-1">
                        <p className="text-sm">{testResult.message}</p>
                        {testResult.value !== undefined && (
                            <p className="text-xs mt-1">
                                Current value: <strong>{testResult.value.toFixed(2)}</strong>
                            </p>
                        )}
                    </div>
                </div>
            )}
        </div>
    );
}
