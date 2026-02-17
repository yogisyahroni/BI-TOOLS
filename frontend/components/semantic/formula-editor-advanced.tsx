'use client';

import * as React from 'react';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
    Tooltip,
    TooltipContent,
    TooltipProvider,
    TooltipTrigger,
} from '@/components/ui/tooltip';
import {
    Calculator,
    Search,
    BookOpen,
    AlertCircle,
    CheckCircle,
    Play,
    Copy,
    Check,
    Save,
    Undo2,
    Redo2,
    Brackets,
    Hash,
    Type,
    ToggleLeft,
    Clock,
    Sigma,
} from 'lucide-react';
import { toast } from 'sonner';

// ============================================================
// Formula Function Reference Catalog
// ============================================================

interface FunctionDef {
    name: string;
    category: string;
    syntax: string;
    description: string;
    example: string;
    returns: string;
}

const FORMULA_FUNCTIONS: FunctionDef[] = [
    // Aggregation
    { name: 'SUM', category: 'Aggregation', syntax: 'SUM(range)', description: 'Adds all values in a range', example: 'SUM(A1:A10)', returns: 'number' },
    { name: 'AVG', category: 'Aggregation', syntax: 'AVG(range)', description: 'Calculates the arithmetic mean', example: 'AVG(B1:B50)', returns: 'number' },
    { name: 'COUNT', category: 'Aggregation', syntax: 'COUNT(range)', description: 'Counts non-empty values', example: 'COUNT(C1:C100)', returns: 'number' },
    { name: 'MIN', category: 'Aggregation', syntax: 'MIN(range)', description: 'Returns the smallest value', example: 'MIN(D1:D20)', returns: 'number' },
    { name: 'MAX', category: 'Aggregation', syntax: 'MAX(range)', description: 'Returns the largest value', example: 'MAX(D1:D20)', returns: 'number' },
    { name: 'MEDIAN', category: 'Aggregation', syntax: 'MEDIAN(range)', description: 'Returns the median value', example: 'MEDIAN(E1:E50)', returns: 'number' },
    { name: 'STDDEV', category: 'Aggregation', syntax: 'STDDEV(range)', description: 'Standard deviation', example: 'STDDEV(F1:F30)', returns: 'number' },
    // Math
    { name: 'ROUND', category: 'Math', syntax: 'ROUND(value, decimals)', description: 'Rounds to N decimal places', example: 'ROUND(3.14159, 2)', returns: 'number' },
    { name: 'ABS', category: 'Math', syntax: 'ABS(value)', description: 'Returns absolute value', example: 'ABS(-42)', returns: 'number' },
    { name: 'POWER', category: 'Math', syntax: 'POWER(base, exp)', description: 'Raises base to exponent', example: 'POWER(2, 10)', returns: 'number' },
    { name: 'SQRT', category: 'Math', syntax: 'SQRT(value)', description: 'Square root', example: 'SQRT(144)', returns: 'number' },
    { name: 'LOG', category: 'Math', syntax: 'LOG(value, base?)', description: 'Logarithm (default base 10)', example: 'LOG(100)', returns: 'number' },
    { name: 'FLOOR', category: 'Math', syntax: 'FLOOR(value)', description: 'Rounds down to nearest integer', example: 'FLOOR(3.7)', returns: 'number' },
    { name: 'CEIL', category: 'Math', syntax: 'CEIL(value)', description: 'Rounds up to nearest integer', example: 'CEIL(3.2)', returns: 'number' },
    // Logic
    { name: 'IF', category: 'Logic', syntax: 'IF(condition, true_val, false_val)', description: 'Conditional evaluation', example: 'IF(A1>100, "High", "Low")', returns: 'any' },
    { name: 'AND', category: 'Logic', syntax: 'AND(cond1, cond2, ...)', description: 'True if all conditions are true', example: 'AND(A1>0, B1<100)', returns: 'boolean' },
    { name: 'OR', category: 'Logic', syntax: 'OR(cond1, cond2, ...)', description: 'True if any condition is true', example: 'OR(A1="Yes", B1="Yes")', returns: 'boolean' },
    { name: 'NOT', category: 'Logic', syntax: 'NOT(condition)', description: 'Negates a boolean', example: 'NOT(A1>50)', returns: 'boolean' },
    { name: 'IFERROR', category: 'Logic', syntax: 'IFERROR(value, fallback)', description: 'Returns fallback on error', example: 'IFERROR(A1/B1, 0)', returns: 'any' },
    // Text
    { name: 'CONCAT', category: 'Text', syntax: 'CONCAT(str1, str2, ...)', description: 'Joins strings together', example: 'CONCAT(A1, " ", B1)', returns: 'string' },
    { name: 'UPPER', category: 'Text', syntax: 'UPPER(text)', description: 'Converts to uppercase', example: 'UPPER("hello")', returns: 'string' },
    { name: 'LOWER', category: 'Text', syntax: 'LOWER(text)', description: 'Converts to lowercase', example: 'LOWER("HELLO")', returns: 'string' },
    { name: 'LEFT', category: 'Text', syntax: 'LEFT(text, count)', description: 'Extracts left N characters', example: 'LEFT("Hello", 3)', returns: 'string' },
    { name: 'RIGHT', category: 'Text', syntax: 'RIGHT(text, count)', description: 'Extracts right N characters', example: 'RIGHT("Hello", 2)', returns: 'string' },
    { name: 'LEN', category: 'Text', syntax: 'LEN(text)', description: 'Returns string length', example: 'LEN("Hello")', returns: 'number' },
    // Date
    { name: 'NOW', category: 'Date', syntax: 'NOW()', description: 'Current date and time', example: 'NOW()', returns: 'datetime' },
    { name: 'TODAY', category: 'Date', syntax: 'TODAY()', description: 'Current date only', example: 'TODAY()', returns: 'date' },
    { name: 'YEAR', category: 'Date', syntax: 'YEAR(date)', description: 'Extracts year from date', example: 'YEAR(A1)', returns: 'number' },
    { name: 'MONTH', category: 'Date', syntax: 'MONTH(date)', description: 'Extracts month from date', example: 'MONTH(A1)', returns: 'number' },
    { name: 'DATEDIFF', category: 'Date', syntax: 'DATEDIFF(start, end, unit)', description: 'Difference between dates', example: 'DATEDIFF(A1, B1, "days")', returns: 'number' },
];

const CATEGORIES = ['All', 'Aggregation', 'Math', 'Logic', 'Text', 'Date'] as const;

const CATEGORY_ICONS: Record<string, React.ReactNode> = {
    All: <BookOpen className="w-3 h-3" />,
    Aggregation: <Sigma className="w-3 h-3" />,
    Math: <Hash className="w-3 h-3" />,
    Logic: <ToggleLeft className="w-3 h-3" />,
    Text: <Type className="w-3 h-3" />,
    Date: <Clock className="w-3 h-3" />,
};

// ============================================================
// Bracket Matching & Validation
// ============================================================

interface FormulaError {
    line: number;
    col: number;
    message: string;
    severity: 'error' | 'warning';
}

function validateFormula(formula: string): FormulaError[] {
    const errors: FormulaError[] = [];
    let parenDepth = 0;
    let inString = false;
    let stringChar = '';

    for (let i = 0; i < formula.length; i++) {
        const ch = formula[i];

        if (inString) {
            if (ch === stringChar) {
                inString = false;
            }
            continue;
        }

        if (ch === '"' || ch === "'") {
            inString = true;
            stringChar = ch;
            continue;
        }

        if (ch === '(') parenDepth++;
        if (ch === ')') {
            parenDepth--;
            if (parenDepth < 0) {
                errors.push({ line: 1, col: i, message: `Unmatched ')' at position ${i}`, severity: 'error' });
            }
        }
    }

    if (inString) {
        errors.push({ line: 1, col: formula.length, message: 'Unterminated string literal', severity: 'error' });
    }

    if (parenDepth > 0) {
        errors.push({ line: 1, col: formula.length, message: `${parenDepth} unclosed parenthesis(es)`, severity: 'error' });
    }

    // Check for empty function calls like FUNC()
    const emptyCallPattern = /\b([A-Z]+)\(\s*\)/g;
    let match;
    while ((match = emptyCallPattern.exec(formula)) !== null) {
        const funcName = match[1];
        if (funcName !== 'NOW' && funcName !== 'TODAY') {
            errors.push({
                line: 1,
                col: match.index,
                message: `${funcName}() called without arguments`,
                severity: 'warning',
            });
        }
    }

    // Check for double operators
    const doubleOpPattern = /[+\-*/]{2,}/g;
    while ((match = doubleOpPattern.exec(formula)) !== null) {
        if (match[0] !== '--') { // Allow double negative
            errors.push({
                line: 1,
                col: match.index,
                message: `Consecutive operators '${match[0]}'`,
                severity: 'error',
            });
        }
    }

    return errors;
}

// ============================================================
// Syntax Highlighter (inline)
// ============================================================

function highlightFormula(formula: string): React.ReactNode[] {
    const tokens: React.ReactNode[] = [];
    let i = 0;
    const len = formula.length;

    while (i < len) {
        // Function name
        const funcMatch = formula.slice(i).match(/^([A-Z_][A-Z_0-9]*)\s*(?=\()/);
        if (funcMatch) {
            tokens.push(
                <span key={`f-${i}`} className="text-purple-400 font-semibold">{funcMatch[1]}</span>
            );
            i += funcMatch[1].length;
            continue;
        }

        // Number
        const numMatch = formula.slice(i).match(/^(\d+\.?\d*)/);
        if (numMatch) {
            tokens.push(
                <span key={`n-${i}`} className="text-amber-400">{numMatch[1]}</span>
            );
            i += numMatch[1].length;
            continue;
        }

        // String literal
        if (formula[i] === '"' || formula[i] === "'") {
            const quote = formula[i];
            let j = i + 1;
            while (j < len && formula[j] !== quote) j++;
            const str = formula.slice(i, j + 1);
            tokens.push(
                <span key={`s-${i}`} className="text-green-400">{str}</span>
            );
            i = j + 1;
            continue;
        }

        // Cell reference (e.g., A1, B2:C5)
        const cellMatch = formula.slice(i).match(/^([A-Z]+\d+(?::[A-Z]+\d+)?)/);
        if (cellMatch) {
            tokens.push(
                <span key={`c-${i}`} className="text-cyan-400">{cellMatch[1]}</span>
            );
            i += cellMatch[1].length;
            continue;
        }

        // Operators
        if ('+-*/^%&=<>!'.includes(formula[i])) {
            tokens.push(
                <span key={`o-${i}`} className="text-rose-400 font-bold">{formula[i]}</span>
            );
            i++;
            continue;
        }

        // Parentheses
        if (formula[i] === '(' || formula[i] === ')') {
            tokens.push(
                <span key={`p-${i}`} className="text-yellow-300">{formula[i]}</span>
            );
            i++;
            continue;
        }

        // Comma
        if (formula[i] === ',') {
            tokens.push(
                <span key={`cm-${i}`} className="text-muted-foreground">{formula[i]}</span>
            );
            i++;
            continue;
        }

        // Default
        tokens.push(
            <span key={`d-${i}`}>{formula[i]}</span>
        );
        i++;
    }

    return tokens;
}

// ============================================================
// Main Component
// ============================================================

interface FormulaEditorAdvancedProps {
    initialFormula?: string;
    columns?: string[];
    onSave?: (formula: string, name: string, description: string) => void;
    onExecute?: (formula: string) => void;
    className?: string;
}

export function FormulaEditorAdvanced({
    initialFormula = '',
    columns = [],
    onSave,
    onExecute,
    className,
}: FormulaEditorAdvancedProps) {
    const [formula, setFormula] = React.useState(initialFormula);
    const [formulaName, setFormulaName] = React.useState('');
    const [formulaDesc, setFormulaDesc] = React.useState('');
    const [errors, setErrors] = React.useState<FormulaError[]>([]);
    const [fnSearch, setFnSearch] = React.useState('');
    const [selectedCategory, setSelectedCategory] = React.useState<string>('All');
    const [copied, setCopied] = React.useState(false);
    const [history, setHistory] = React.useState<string[]>([initialFormula]);
    const [historyIdx, setHistoryIdx] = React.useState(0);
    const textareaRef = React.useRef<HTMLTextAreaElement>(null);

    // Live validation
    React.useEffect(() => {
        const timer = setTimeout(() => {
            setErrors(validateFormula(formula));
        }, 300);
        return () => clearTimeout(timer);
    }, [formula]);

    // Push to undo history on change
    const pushHistory = React.useCallback((newFormula: string) => {
        setHistory(prev => {
            const trimmed = prev.slice(0, historyIdx + 1);
            return [...trimmed, newFormula];
        });
        setHistoryIdx(prev => prev + 1);
    }, [historyIdx]);

    const handleFormulaChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
        const val = e.target.value;
        setFormula(val);
        pushHistory(val);
    };

    const handleUndo = () => {
        if (historyIdx > 0) {
            const newIdx = historyIdx - 1;
            setHistoryIdx(newIdx);
            setFormula(history[newIdx]);
        }
    };

    const handleRedo = () => {
        if (historyIdx < history.length - 1) {
            const newIdx = historyIdx + 1;
            setHistoryIdx(newIdx);
            setFormula(history[newIdx]);
        }
    };

    const handleInsertFunction = (fn: FunctionDef) => {
        const insertion = `${fn.name}()`;
        const textarea = textareaRef.current;
        if (textarea) {
            const start = textarea.selectionStart;
            const end = textarea.selectionEnd;
            const newFormula =
                formula.slice(0, start) + insertion + formula.slice(end);
            setFormula(newFormula);
            pushHistory(newFormula);
            // Position cursor inside parens
            setTimeout(() => {
                textarea.focus();
                textarea.setSelectionRange(start + fn.name.length + 1, start + fn.name.length + 1);
            }, 0);
        } else {
            const newFormula = formula + insertion;
            setFormula(newFormula);
            pushHistory(newFormula);
        }
    };

    const handleInsertColumn = (col: string) => {
        const textarea = textareaRef.current;
        if (textarea) {
            const start = textarea.selectionStart;
            const end = textarea.selectionEnd;
            const newFormula = formula.slice(0, start) + col + formula.slice(end);
            setFormula(newFormula);
            pushHistory(newFormula);
            setTimeout(() => {
                textarea.focus();
                textarea.setSelectionRange(start + col.length, start + col.length);
            }, 0);
        } else {
            const newFormula = formula + col;
            setFormula(newFormula);
            pushHistory(newFormula);
        }
    };

    const handleCopy = async () => {
        if (!formula) return;
        try {
            await navigator.clipboard.writeText(formula);
            setCopied(true);
            toast.success('Formula copied');
            setTimeout(() => setCopied(false), 2000);
        } catch {
            toast.error('Copy failed');
        }
    };

    const handleSave = () => {
        if (!formula.trim() || !formulaName.trim()) {
            toast.error('Formula and name are required');
            return;
        }
        if (errors.some(e => e.severity === 'error')) {
            toast.error('Fix errors before saving');
            return;
        }
        onSave?.(formula, formulaName, formulaDesc);
        toast.success('Formula saved');
    };

    const handleExecute = () => {
        if (!formula.trim()) return;
        if (errors.some(e => e.severity === 'error')) {
            toast.error('Fix errors before running');
            return;
        }
        onExecute?.(formula);
    };

    // Filtered function list
    const filteredFunctions = FORMULA_FUNCTIONS.filter(fn => {
        const matchesCategory = selectedCategory === 'All' || fn.category === selectedCategory;
        const matchesSearch = !fnSearch || fn.name.toLowerCase().includes(fnSearch.toLowerCase()) ||
            fn.description.toLowerCase().includes(fnSearch.toLowerCase());
        return matchesCategory && matchesSearch;
    });

    const isValid = errors.length === 0 && formula.trim().length > 0;
    const errorCount = errors.filter(e => e.severity === 'error').length;
    const warningCount = errors.filter(e => e.severity === 'warning').length;

    return (
        <Card
            className={cn(
                'flex flex-col bg-gradient-to-br from-card/60 to-card/30 backdrop-blur-xl border-border/40 shadow-2xl overflow-hidden',
                className,
            )}
        >
            {/* Header */}
            <div className="flex items-center justify-between p-4 border-b border-border/30 bg-card/40">
                <div className="flex items-center gap-3">
                    <div className="w-9 h-9 rounded-xl bg-gradient-to-br from-amber-500/20 to-orange-600/20 flex items-center justify-center ring-1 ring-amber-500/30">
                        <Calculator className="w-4.5 h-4.5 text-amber-500" />
                    </div>
                    <div>
                        <h3 className="text-sm font-bold tracking-tight">Formula Editor</h3>
                        <p className="text-[10px] text-muted-foreground">
                            Build, validate, and preview formulas
                        </p>
                    </div>
                </div>

                {/* Status Badges */}
                <div className="flex items-center gap-2">
                    {formula.trim() && (
                        <>
                            {isValid ? (
                                <Badge variant="default" className="h-5 text-[10px] bg-emerald-500/20 text-emerald-400 border-emerald-500/30">
                                    <CheckCircle className="w-3 h-3 mr-1" />
                                    Valid
                                </Badge>
                            ) : (
                                <>
                                    {errorCount > 0 && (
                                        <Badge variant="destructive" className="h-5 text-[10px]">
                                            <AlertCircle className="w-3 h-3 mr-1" />
                                            {errorCount} error{errorCount !== 1 ? 's' : ''}
                                        </Badge>
                                    )}
                                    {warningCount > 0 && (
                                        <Badge variant="outline" className="h-5 text-[10px] text-amber-400 border-amber-400/30">
                                            {warningCount} warning{warningCount !== 1 ? 's' : ''}
                                        </Badge>
                                    )}
                                </>
                            )}
                        </>
                    )}
                </div>
            </div>

            {/* Body - Split: Editor + Reference */}
            <div className="flex flex-1 min-h-0">
                {/* Left: Editor */}
                <div className="flex-1 flex flex-col border-r border-border/20">
                    {/* Toolbar */}
                    <div className="flex items-center gap-1 px-3 py-2 border-b border-border/20 bg-card/20">
                        <TooltipProvider delayDuration={200}>
                            <Tooltip>
                                <TooltipTrigger asChild>
                                    <Button variant="ghost" size="icon" className="h-7 w-7" onClick={handleUndo} disabled={historyIdx <= 0}>
                                        <Undo2 className="w-3.5 h-3.5" />
                                    </Button>
                                </TooltipTrigger>
                                <TooltipContent side="bottom" className="text-xs">Undo</TooltipContent>
                            </Tooltip>
                            <Tooltip>
                                <TooltipTrigger asChild>
                                    <Button variant="ghost" size="icon" className="h-7 w-7" onClick={handleRedo} disabled={historyIdx >= history.length - 1}>
                                        <Redo2 className="w-3.5 h-3.5" />
                                    </Button>
                                </TooltipTrigger>
                                <TooltipContent side="bottom" className="text-xs">Redo</TooltipContent>
                            </Tooltip>
                        </TooltipProvider>

                        <div className="w-px h-4 bg-border/40 mx-1" />

                        <TooltipProvider delayDuration={200}>
                            <Tooltip>
                                <TooltipTrigger asChild>
                                    <Button variant="ghost" size="icon" className="h-7 w-7" onClick={handleCopy} disabled={!formula}>
                                        {copied ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
                                    </Button>
                                </TooltipTrigger>
                                <TooltipContent side="bottom" className="text-xs">Copy</TooltipContent>
                            </Tooltip>
                        </TooltipProvider>

                        <div className="flex-1" />

                        {/* Quick Insert Operators */}
                        <div className="flex gap-0.5">
                            {['+', '-', '*', '/', '(', ')', ','].map(op => (
                                <Button
                                    key={op}
                                    variant="ghost"
                                    size="sm"
                                    className="h-6 w-6 p-0 text-[11px] font-mono text-muted-foreground hover:text-foreground"
                                    onClick={() => {
                                        const newF = formula + op;
                                        setFormula(newF);
                                        pushHistory(newF);
                                    }}
                                >
                                    {op}
                                </Button>
                            ))}
                        </div>
                    </div>

                    {/* Textarea Editor */}
                    <div className="flex-1 relative">
                        <textarea
                            ref={textareaRef}
                            value={formula}
                            onChange={handleFormulaChange}
                            placeholder="Enter your formula here... e.g. SUM(revenue) / COUNT(orders)"
                            spellCheck={false}
                            className={cn(
                                'w-full h-full resize-none p-4 bg-transparent font-mono text-sm',
                                'focus:outline-none focus:ring-0 border-0',
                                'placeholder:text-muted-foreground/40',
                                errors.some(e => e.severity === 'error') && 'text-destructive/90',
                            )}
                            style={{ minHeight: 120 }}
                        />
                    </div>

                    {/* Live Preview */}
                    {formula.trim() && (
                        <div className="border-t border-border/20 bg-background/60">
                            <div className="px-3 py-1.5 flex items-center gap-2">
                                <Brackets className="w-3 h-3 text-muted-foreground" />
                                <span className="text-[10px] text-muted-foreground font-medium">Preview</span>
                            </div>
                            <div className="px-4 pb-3 font-mono text-sm leading-relaxed">
                                {highlightFormula(formula)}
                            </div>
                        </div>
                    )}

                    {/* Errors Panel */}
                    {errors.length > 0 && (
                        <div className="border-t border-border/20 max-h-24 overflow-auto">
                            {errors.map((err, idx) => (
                                <div
                                    key={idx}
                                    className={cn(
                                        'flex items-center gap-2 px-4 py-1.5 text-[11px]',
                                        err.severity === 'error'
                                            ? 'text-red-400 bg-red-500/5'
                                            : 'text-amber-400 bg-amber-500/5',
                                    )}
                                >
                                    <AlertCircle className="w-3 h-3 flex-shrink-0" />
                                    <span>Col {err.col}: {err.message}</span>
                                </div>
                            ))}
                        </div>
                    )}

                    {/* Save Form + Actions */}
                    <div className="p-3 border-t border-border/20 space-y-2">
                        <div className="flex gap-2">
                            <Input
                                placeholder="Metric name"
                                value={formulaName}
                                onChange={e => setFormulaName(e.target.value)}
                                className="h-8 text-xs flex-1"
                            />
                            <Input
                                placeholder="Description (optional)"
                                value={formulaDesc}
                                onChange={e => setFormulaDesc(e.target.value)}
                                className="h-8 text-xs flex-1"
                            />
                        </div>
                        <div className="flex gap-2">
                            {onExecute && (
                                <Button
                                    variant="outline"
                                    size="sm"
                                    className="flex-1 h-8"
                                    onClick={handleExecute}
                                    disabled={!formula.trim() || errorCount > 0}
                                >
                                    <Play className="w-3 h-3 mr-1" />
                                    Run Preview
                                </Button>
                            )}
                            <Button
                                size="sm"
                                className="flex-1 h-8"
                                onClick={handleSave}
                                disabled={!isValid || !formulaName.trim()}
                            >
                                <Save className="w-3 h-3 mr-1" />
                                Save Metric
                            </Button>
                        </div>
                    </div>
                </div>

                {/* Right: Function Reference & Columns */}
                <div className="w-[260px] flex flex-col bg-card/20">
                    <Tabs defaultValue="functions" className="flex flex-col flex-1">
                        <TabsList className="h-8 rounded-none border-b border-border/20 bg-transparent px-2">
                            <TabsTrigger value="functions" className="h-6 text-[10px] px-2 data-[state=active]:bg-primary/10">
                                Functions
                            </TabsTrigger>
                            {columns.length > 0 && (
                                <TabsTrigger value="columns" className="h-6 text-[10px] px-2 data-[state=active]:bg-primary/10">
                                    Columns
                                </TabsTrigger>
                            )}
                        </TabsList>

                        <TabsContent value="functions" className="flex-1 flex flex-col m-0 mt-0">
                            {/* Search */}
                            <div className="px-2 pt-2 pb-1">
                                <div className="relative">
                                    <Search className="absolute left-2 top-1/2 -translate-y-1/2 w-3 h-3 text-muted-foreground" />
                                    <Input
                                        placeholder="Search functions..."
                                        value={fnSearch}
                                        onChange={e => setFnSearch(e.target.value)}
                                        className="h-7 text-[11px] pl-7"
                                    />
                                </div>
                            </div>

                            {/* Category Filter */}
                            <div className="flex flex-wrap gap-1 px-2 pb-2">
                                {CATEGORIES.map(cat => (
                                    <Button
                                        key={cat}
                                        variant={selectedCategory === cat ? 'secondary' : 'ghost'}
                                        size="sm"
                                        className={cn(
                                            'h-5 px-1.5 text-[9px] gap-1',
                                            selectedCategory === cat && 'bg-primary/15 text-primary',
                                        )}
                                        onClick={() => setSelectedCategory(cat)}
                                    >
                                        {CATEGORY_ICONS[cat]}
                                        {cat}
                                    </Button>
                                ))}
                            </div>

                            {/* Function List */}
                            <ScrollArea className="flex-1">
                                <div className="px-2 pb-2 space-y-0.5">
                                    {filteredFunctions.map(fn => (
                                        <button
                                            key={fn.name}
                                            onClick={() => handleInsertFunction(fn)}
                                            className={cn(
                                                'w-full text-left p-2 rounded-md',
                                                'hover:bg-primary/10 transition-colors duration-150',
                                                'group cursor-pointer',
                                            )}
                                        >
                                            <div className="flex items-center gap-2">
                                                <code className="text-[11px] font-bold text-purple-400">{fn.name}</code>
                                                <Badge variant="outline" className="h-3.5 text-[8px] px-1 ml-auto opacity-60">
                                                    {fn.returns}
                                                </Badge>
                                            </div>
                                            <p className="text-[10px] text-muted-foreground mt-0.5 line-clamp-1">{fn.description}</p>
                                            <p className="text-[9px] text-muted-foreground/60 font-mono mt-0.5 opacity-0 group-hover:opacity-100 transition-opacity">
                                                {fn.syntax}
                                            </p>
                                        </button>
                                    ))}
                                    {filteredFunctions.length === 0 && (
                                        <p className="text-[11px] text-muted-foreground/50 text-center py-6">No functions found</p>
                                    )}
                                </div>
                            </ScrollArea>
                        </TabsContent>

                        {columns.length > 0 && (
                            <TabsContent value="columns" className="flex-1 m-0 mt-0">
                                <ScrollArea className="h-full">
                                    <div className="px-2 py-2 space-y-0.5">
                                        {columns.map(col => (
                                            <button
                                                key={col}
                                                onClick={() => handleInsertColumn(col)}
                                                className="w-full text-left px-2 py-1.5 rounded-md hover:bg-primary/10 transition-colors text-xs font-mono text-cyan-400"
                                            >
                                                {col}
                                            </button>
                                        ))}
                                    </div>
                                </ScrollArea>
                            </TabsContent>
                        )}
                    </Tabs>
                </div>
            </div>
        </Card>
    );
}
