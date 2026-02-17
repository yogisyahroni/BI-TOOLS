'use client';

import * as React from 'react';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Input } from '@/components/ui/input';
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select';
import {
    Tooltip,
    TooltipContent,
    TooltipProvider,
    TooltipTrigger,
} from '@/components/ui/tooltip';
import {
    Database,
    Link2,
    _Plus,
    Trash2,
    ZoomIn,
    ZoomOut,
    Maximize2,
    _Move,
    GripVertical,
    ArrowRight,
    ArrowLeftRight,
    Key,
    Columns,
    Search,
    Undo2,
} from 'lucide-react';
import { toast } from 'sonner';

// ============================================================
// Types
// ============================================================

interface TableColumn {
    name: string;
    type: string;
    isPrimary?: boolean;
    isForeignKey?: boolean;
}

interface TableNode {
    id: string;
    name: string;
    schema?: string;
    columns: TableColumn[];
    x: number;
    y: number;
}

interface Relationship {
    id: string;
    fromTable: string;
    fromColumn: string;
    toTable: string;
    toColumn: string;
    type: 'one-to-one' | 'one-to-many' | 'many-to-many';
}

interface RelationshipEditorProps {
    tables: {
        name: string;
        schema?: string;
        columns: { name: string; type: string; isPrimary?: boolean; isForeignKey?: boolean }[];
    }[];
    relationships?: Relationship[];
    onSaveRelationship?: (rel: Omit<Relationship, 'id'>) => Promise<void>;
    onDeleteRelationship?: (id: string) => Promise<void>;
    className?: string;
}

// ============================================================
// Layout Helpers
// ============================================================

function layoutTables(tables: RelationshipEditorProps['tables']): TableNode[] {
    const cols = Math.ceil(Math.sqrt(tables.length));
    const cardW = 220;
    const cardH = 200;
    const gapX = 60;
    const gapY = 40;
    const padX = 40;
    const padY = 40;

    return tables.map((t, i) => ({
        id: t.name,
        name: t.name,
        schema: t.schema,
        columns: t.columns,
        x: padX + (i % cols) * (cardW + gapX),
        y: padY + Math.floor(i / cols) * (cardH + gapY),
    }));
}

// Relationship type labels
const REL_TYPE_LABELS: Record<string, string> = {
    'one-to-one': '1 ↔ 1',
    'one-to-many': '1 → N',
    'many-to-many': 'N ↔ N',
};

// ============================================================
// Sub-Components
// ============================================================

interface TableCardProps {
    node: TableNode;
    isSelected: boolean;
    isDragTarget: boolean;
    selectedColumn: string | null;
    onMouseDown: (e: React.MouseEvent) => void;
    onColumnClick: (table: string, column: string) => void;
}

function TableCard({ node, isSelected, isDragTarget, selectedColumn, onMouseDown, onColumnClick }: TableCardProps) {
    return (
        <div
            className={cn(
                'absolute select-none rounded-xl border shadow-lg transition-shadow duration-200',
                'bg-card/90 backdrop-blur-md',
                isSelected && 'ring-2 ring-primary shadow-primary/20',
                isDragTarget && 'ring-2 ring-emerald-500/60 shadow-emerald-500/20',
                !isSelected && !isDragTarget && 'border-border/40 hover:shadow-xl',
            )}
            style={{
                left: node.x,
                top: node.y,
                width: 210,
                cursor: 'grab',
            }}
            onMouseDown={onMouseDown}
        >
            {/* Table Header */}
            <div className="flex items-center gap-2 px-3 py-2.5 border-b border-border/30 bg-gradient-to-r from-primary/5 to-transparent rounded-t-xl">
                <GripVertical className="w-3.5 h-3.5 text-muted-foreground/50" />
                <Database className="w-3.5 h-3.5 text-primary" />
                <span className="text-xs font-bold tracking-tight truncate flex-1">{node.name}</span>
                {node.schema && (
                    <Badge variant="outline" className="h-4 text-[8px] px-1 font-mono">{node.schema}</Badge>
                )}
            </div>

            {/* Columns */}
            <div className="py-1 max-h-[180px] overflow-auto">
                {node.columns.map(col => {
                    const isActive = selectedColumn === `${node.name}.${col.name}`;
                    return (
                        <button
                            key={col.name}
                            className={cn(
                                'w-full flex items-center gap-2 px-3 py-1 text-left',
                                'hover:bg-primary/10 transition-colors duration-100 cursor-pointer',
                                isActive && 'bg-primary/15',
                            )}
                            onClick={(e) => {
                                e.stopPropagation();
                                onColumnClick(node.name, col.name);
                            }}
                        >
                            {col.isPrimary ? (
                                <Key className="w-3 h-3 text-amber-400 flex-shrink-0" />
                            ) : col.isForeignKey ? (
                                <Link2 className="w-3 h-3 text-blue-400 flex-shrink-0" />
                            ) : (
                                <Columns className="w-3 h-3 text-muted-foreground/40 flex-shrink-0" />
                            )}
                            <span className="text-[11px] font-mono truncate flex-1">{col.name}</span>
                            <span className="text-[9px] text-muted-foreground/50 font-mono">{col.type}</span>
                        </button>
                    );
                })}
            </div>
        </div>
    );
}

// ============================================================
// SVG Relationship Lines
// ============================================================

function getConnectionPoint(node: TableNode, column: string, side: 'left' | 'right'): { x: number; y: number } {
    const colIdx = node.columns.findIndex(c => c.name === column);
    const headerH = 36;
    const rowH = 26;
    const y = node.y + headerH + (colIdx + 0.5) * rowH;
    const x = side === 'right' ? node.x + 210 : node.x;
    return { x, y };
}

interface RelLineProps {
    from: { x: number; y: number };
    to: { x: number; y: number };
    rel: Relationship;
    isSelected: boolean;
    onSelect: () => void;
}

function RelLine({ from, to, rel, isSelected, onSelect }: RelLineProps) {
    const midX = (from.x + to.x) / 2;
    const dx = Math.abs(to.x - from.x);
    const controlOffset = Math.max(60, dx * 0.4);

    const path = `M ${from.x} ${from.y} C ${from.x + controlOffset} ${from.y}, ${to.x - controlOffset} ${to.y}, ${to.x} ${to.y}`;

    return (
        <g onClick={(e) => { e.stopPropagation(); onSelect(); }} className="cursor-pointer">
            {/* Hit area */}
            <path d={path} fill="none" stroke="transparent" strokeWidth={14} />
            {/* Visible line */}
            <path
                d={path}
                fill="none"
                stroke={isSelected ? 'hsl(var(--primary))' : 'hsl(var(--muted-foreground) / 0.3)'}
                strokeWidth={isSelected ? 2.5 : 1.5}
                strokeDasharray={rel.type === 'many-to-many' ? '6 4' : 'none'}
                className="transition-all duration-200"
            />
            {/* Label */}
            <text
                x={midX}
                y={(from.y + to.y) / 2 - 8}
                textAnchor="middle"
                className="text-[9px] fill-muted-foreground/60 font-mono select-none"
            >
                {REL_TYPE_LABELS[rel.type] || rel.type}
            </text>
            {/* Dots at endpoints */}
            <circle cx={from.x} cy={from.y} r={3.5} fill={isSelected ? 'hsl(var(--primary))' : 'hsl(var(--muted-foreground) / 0.4)'} />
            <circle cx={to.x} cy={to.y} r={3.5} fill={isSelected ? 'hsl(var(--primary))' : 'hsl(var(--muted-foreground) / 0.4)'} />
        </g>
    );
}

// ============================================================
// Main Component
// ============================================================

export function RelationshipEditor({
    tables,
    relationships: initialRelationships = [],
    onSaveRelationship,
    onDeleteRelationship,
    className,
}: RelationshipEditorProps) {
    const [nodes, setNodes] = React.useState<TableNode[]>(() => layoutTables(tables));
    const [rels, setRels] = React.useState<Relationship[]>(initialRelationships);
    const [selectedRel, setSelectedRel] = React.useState<string | null>(null);
    const [zoom, setZoom] = React.useState(1);
    const [pan, setPan] = React.useState({ x: 0, y: 0 });
    const [draggingNode, setDraggingNode] = React.useState<string | null>(null);
    const [dragOffset, setDragOffset] = React.useState({ x: 0, y: 0 });
    const [isPanning, setIsPanning] = React.useState(false);
    const [panStart, setPanStart] = React.useState({ x: 0, y: 0 });
    const [linkingFrom, setLinkingFrom] = React.useState<{ table: string; column: string } | null>(null);
    const [newRelType, setNewRelType] = React.useState<'one-to-one' | 'one-to-many' | 'many-to-many'>('one-to-many');
    const [tableSearch, setTableSearch] = React.useState('');
    const containerRef = React.useRef<HTMLDivElement>(null);

    // Sync if tables change
    React.useEffect(() => {
        setNodes(layoutTables(tables));
    }, [tables]);

    React.useEffect(() => {
        setRels(initialRelationships);
    }, [initialRelationships]);

    // Drag node handlers
    const handleNodeMouseDown = (nodeId: string, e: React.MouseEvent) => {
        e.preventDefault();
        e.stopPropagation();
        const node = nodes.find(n => n.id === nodeId);
        if (!node) return;
        setDraggingNode(nodeId);
        setDragOffset({
            x: e.clientX / zoom - node.x - pan.x,
            y: e.clientY / zoom - node.y - pan.y,
        });
    };

    // Pan handlers
    const handleCanvasMouseDown = (e: React.MouseEvent) => {
        if (draggingNode) return;
        if (e.button !== 0) return;
        setIsPanning(true);
        setPanStart({ x: e.clientX - pan.x * zoom, y: e.clientY - pan.y * zoom });
        setSelectedRel(null);
        setLinkingFrom(null);
    };

    React.useEffect(() => {
        const handleMouseMove = (e: MouseEvent) => {
            if (draggingNode) {
                setNodes(prev => prev.map(n => {
                    if (n.id !== draggingNode) return n;
                    return {
                        ...n,
                        x: e.clientX / zoom - dragOffset.x - pan.x,
                        y: e.clientY / zoom - dragOffset.y - pan.y,
                    };
                }));
            }
            if (isPanning) {
                setPan({
                    x: (e.clientX - panStart.x) / zoom,
                    y: (e.clientY - panStart.y) / zoom,
                });
            }
        };

        const handleMouseUp = () => {
            setDraggingNode(null);
            setIsPanning(false);
        };

        window.addEventListener('mousemove', handleMouseMove);
        window.addEventListener('mouseup', handleMouseUp);
        return () => {
            window.removeEventListener('mousemove', handleMouseMove);
            window.removeEventListener('mouseup', handleMouseUp);
        };
    }, [draggingNode, dragOffset, zoom, pan, isPanning, panStart]);

    // Zoom
    const handleWheel = React.useCallback((e: React.WheelEvent) => {
        e.preventDefault();
        const delta = e.deltaY > 0 ? -0.05 : 0.05;
        setZoom(prev => Math.max(0.3, Math.min(2, prev + delta)));
    }, []);

    // Column click → start/finish linking
    const handleColumnClick = (table: string, column: string) => {
        if (!linkingFrom) {
            setLinkingFrom({ table, column });
            toast.info(`Select target column for ${table}.${column}`);
        } else {
            if (linkingFrom.table === table && linkingFrom.column === column) {
                setLinkingFrom(null);
                return;
            }
            // Create relationship
            const newRel: Relationship = {
                id: `rel-${Date.now()}`,
                fromTable: linkingFrom.table,
                fromColumn: linkingFrom.column,
                toTable: table,
                toColumn: column,
                type: newRelType,
            };

            const duplicate = rels.some(r =>
                (r.fromTable === newRel.fromTable && r.fromColumn === newRel.fromColumn &&
                    r.toTable === newRel.toTable && r.toColumn === newRel.toColumn) ||
                (r.fromTable === newRel.toTable && r.fromColumn === newRel.toColumn &&
                    r.toTable === newRel.fromTable && r.toColumn === newRel.fromColumn)
            );

            if (duplicate) {
                toast.error('Relationship already exists');
                setLinkingFrom(null);
                return;
            }

            setRels(prev => [...prev, newRel]);
            onSaveRelationship?.({
                fromTable: newRel.fromTable,
                fromColumn: newRel.fromColumn,
                toTable: newRel.toTable,
                toColumn: newRel.toColumn,
                type: newRel.type,
            });
            toast.success(`Linked ${linkingFrom.table}.${linkingFrom.column} → ${table}.${column}`);
            setLinkingFrom(null);
        }
    };

    const handleDeleteRelationship = async () => {
        if (!selectedRel) return;
        try {
            await onDeleteRelationship?.(selectedRel);
            setRels(prev => prev.filter(r => r.id !== selectedRel));
            setSelectedRel(null);
            toast.success('Relationship deleted');
        } catch {
            toast.error('Failed to delete relationship');
        }
    };

    // Fit to view
    const handleFitView = () => {
        if (nodes.length === 0) return;
        const minX = Math.min(...nodes.map(n => n.x));
        const minY = Math.min(...nodes.map(n => n.y));
        const maxX = Math.max(...nodes.map(n => n.x + 210));
        const maxY = Math.max(...nodes.map(n => n.y + 200));

        const container = containerRef.current;
        if (!container) return;
        const cw = container.clientWidth;
        const ch = container.clientHeight;
        const contentW = maxX - minX + 80;
        const contentH = maxY - minY + 80;
        const fitZoom = Math.min(cw / contentW, ch / contentH, 1.2);

        setZoom(fitZoom);
        setPan({ x: -minX + 40, y: -minY + 40 });
    };

    // Build SVG lines
    const lines = rels.map(rel => {
        const fromNode = nodes.find(n => n.id === rel.fromTable);
        const toNode = nodes.find(n => n.id === rel.toTable);
        if (!fromNode || !toNode) return null;

        const fromSide = fromNode.x < toNode.x ? 'right' : 'left';
        const toSide = fromNode.x < toNode.x ? 'left' : 'right';

        const from = getConnectionPoint(fromNode, rel.fromColumn, fromSide as 'left' | 'right');
        const to = getConnectionPoint(toNode, rel.toColumn, toSide as 'left' | 'right');

        return (
            <RelLine
                key={rel.id}
                from={from}
                to={to}
                rel={rel}
                isSelected={selectedRel === rel.id}
                onSelect={() => setSelectedRel(rel.id === selectedRel ? null : rel.id)}
            />
        );
    });

    const selectedColumn = linkingFrom ? `${linkingFrom.table}.${linkingFrom.column}` : null;

    const filteredNodes = tableSearch
        ? nodes.filter(n => n.name.toLowerCase().includes(tableSearch.toLowerCase()))
        : nodes;

    return (
        <Card className={cn('flex flex-col overflow-hidden bg-card/50 backdrop-blur-xl border-border/40', className)}>
            {/* Toolbar */}
            <div className="flex items-center gap-2 px-4 py-2.5 border-b border-border/30 bg-card/60">
                <div className="flex items-center gap-2">
                    <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-blue-500/20 to-indigo-600/20 flex items-center justify-center ring-1 ring-blue-500/30">
                        <Link2 className="w-4 h-4 text-blue-500" />
                    </div>
                    <div>
                        <h3 className="text-sm font-bold tracking-tight">Relationship Editor</h3>
                        <p className="text-[10px] text-muted-foreground">
                            {tables.length} tables · {rels.length} relationships
                        </p>
                    </div>
                </div>

                <div className="flex-1" />

                {/* Link Mode */}
                {linkingFrom && (
                    <Badge variant="default" className="h-6 text-[10px] bg-primary/20 text-primary border-primary/30 animate-pulse">
                        <Link2 className="w-3 h-3 mr-1" />
                        Linking: {linkingFrom.table}.{linkingFrom.column}
                    </Badge>
                )}

                {/* Relationship Type Selector */}
                <Select value={newRelType} onValueChange={(v) => setNewRelType(v as typeof newRelType)}>
                    <SelectTrigger className="w-[130px] h-7 text-[11px]">
                        <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value="one-to-one" className="text-xs">
                            <span className="flex items-center gap-1"><ArrowLeftRight className="w-3 h-3" /> One-to-One</span>
                        </SelectItem>
                        <SelectItem value="one-to-many" className="text-xs">
                            <span className="flex items-center gap-1"><ArrowRight className="w-3 h-3" /> One-to-Many</span>
                        </SelectItem>
                        <SelectItem value="many-to-many" className="text-xs">
                            <span className="flex items-center gap-1"><ArrowLeftRight className="w-3 h-3" /> Many-to-Many</span>
                        </SelectItem>
                    </SelectContent>
                </Select>

                <div className="w-px h-5 bg-border/30" />

                {/* Zoom Controls */}
                <TooltipProvider delayDuration={200}>
                    <Tooltip>
                        <TooltipTrigger asChild>
                            <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => setZoom(z => Math.min(2, z + 0.1))}>
                                <ZoomIn className="w-3.5 h-3.5" />
                            </Button>
                        </TooltipTrigger>
                        <TooltipContent className="text-xs">Zoom In</TooltipContent>
                    </Tooltip>
                    <Tooltip>
                        <TooltipTrigger asChild>
                            <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => setZoom(z => Math.max(0.3, z - 0.1))}>
                                <ZoomOut className="w-3.5 h-3.5" />
                            </Button>
                        </TooltipTrigger>
                        <TooltipContent className="text-xs">Zoom Out</TooltipContent>
                    </Tooltip>
                    <Tooltip>
                        <TooltipTrigger asChild>
                            <Button variant="ghost" size="icon" className="h-7 w-7" onClick={handleFitView}>
                                <Maximize2 className="w-3.5 h-3.5" />
                            </Button>
                        </TooltipTrigger>
                        <TooltipContent className="text-xs">Fit to View</TooltipContent>
                    </Tooltip>
                </TooltipProvider>

                <Badge variant="outline" className="h-5 text-[9px] px-1.5 font-mono">
                    {Math.round(zoom * 100)}%
                </Badge>

                {/* Delete Selected */}
                {selectedRel && (
                    <>
                        <div className="w-px h-5 bg-border/30" />
                        <Button variant="destructive" size="sm" className="h-7 text-[11px] px-2" onClick={handleDeleteRelationship}>
                            <Trash2 className="w-3 h-3 mr-1" />
                            Delete
                        </Button>
                    </>
                )}

                {linkingFrom && (
                    <Button variant="ghost" size="sm" className="h-7 text-[11px] px-2" onClick={() => setLinkingFrom(null)}>
                        <Undo2 className="w-3 h-3 mr-1" />
                        Cancel
                    </Button>
                )}
            </div>

            {/* Canvas */}
            <div className="flex flex-1 min-h-0">
                {/* Sidebar — Table List */}
                <div className="w-[180px] border-r border-border/20 flex flex-col bg-card/30">
                    <div className="p-2">
                        <div className="relative">
                            <Search className="absolute left-2 top-1/2 -translate-y-1/2 w-3 h-3 text-muted-foreground" />
                            <Input
                                placeholder="Search tables..."
                                value={tableSearch}
                                onChange={e => setTableSearch(e.target.value)}
                                className="h-7 text-[11px] pl-7"
                            />
                        </div>
                    </div>
                    <ScrollArea className="flex-1">
                        <div className="px-2 pb-2 space-y-0.5">
                            {(tableSearch
                                ? tables.filter(t => t.name.toLowerCase().includes(tableSearch.toLowerCase()))
                                : tables
                            ).map(t => {
                                const relCount = rels.filter(r => r.fromTable === t.name || r.toTable === t.name).length;
                                return (
                                    <div
                                        key={t.name}
                                        className="flex items-center gap-2 px-2 py-1.5 rounded-md hover:bg-primary/10 transition-colors cursor-default"
                                    >
                                        <Database className="w-3 h-3 text-muted-foreground/60 flex-shrink-0" />
                                        <span className="text-[11px] truncate flex-1 font-mono">{t.name}</span>
                                        {relCount > 0 && (
                                            <Badge variant="secondary" className="h-4 text-[8px] px-1">{relCount}</Badge>
                                        )}
                                    </div>
                                );
                            })}
                        </div>
                    </ScrollArea>

                    {/* Legend */}
                    <div className="p-3 border-t border-border/20 space-y-1.5">
                        <p className="text-[9px] text-muted-foreground/60 font-medium uppercase tracking-wider">Legend</p>
                        <div className="space-y-1">
                            <div className="flex items-center gap-2 text-[10px] text-muted-foreground">
                                <Key className="w-3 h-3 text-amber-400" />
                                Primary Key
                            </div>
                            <div className="flex items-center gap-2 text-[10px] text-muted-foreground">
                                <Link2 className="w-3 h-3 text-blue-400" />
                                Foreign Key
                            </div>
                            <div className="flex items-center gap-2 text-[10px] text-muted-foreground">
                                <div className="w-3 h-0.5 bg-muted-foreground/40" />
                                Solid = 1:1 / 1:N
                            </div>
                            <div className="flex items-center gap-2 text-[10px] text-muted-foreground">
                                <div className="w-3 h-0.5 bg-muted-foreground/40 border-b border-dashed" />
                                Dashed = N:N
                            </div>
                        </div>
                        <p className="text-[9px] text-muted-foreground/40 mt-2">
                            Click column → column to link
                        </p>
                    </div>
                </div>

                {/* Graph Canvas */}
                <div
                    ref={containerRef}
                    className="flex-1 relative overflow-hidden bg-gradient-to-br from-background to-card/30"
                    style={{ cursor: isPanning ? 'grabbing' : (draggingNode ? 'grabbing' : 'default') }}
                    onMouseDown={handleCanvasMouseDown}
                    onWheel={handleWheel}
                >
                    {/* Grid pattern */}
                    <svg className="absolute inset-0 w-full h-full pointer-events-none opacity-[0.04]">
                        <defs>
                            <pattern id="grid" x={pan.x * zoom} y={pan.y * zoom} width={40 * zoom} height={40 * zoom} patternUnits="userSpaceOnUse">
                                <circle cx={1} cy={1} r={1} fill="currentColor" />
                            </pattern>
                        </defs>
                        <rect width="100%" height="100%" fill="url(#grid)" />
                    </svg>

                    {/* Transformed content */}
                    <div
                        style={{
                            transform: `scale(${zoom}) translate(${pan.x}px, ${pan.y}px)`,
                            transformOrigin: '0 0',
                            position: 'absolute',
                            inset: 0,
                        }}
                    >
                        {/* SVG Lines */}
                        <svg className="absolute inset-0 w-[4000px] h-[4000px] pointer-events-none" style={{ zIndex: 1 }}>
                            <g style={{ pointerEvents: 'auto' }}>
                                {lines}
                            </g>
                        </svg>

                        {/* Table Cards */}
                        <div style={{ position: 'relative', zIndex: 2 }}>
                            {filteredNodes.map(node => (
                                <TableCard
                                    key={node.id}
                                    node={node}
                                    isSelected={draggingNode === node.id}
                                    isDragTarget={linkingFrom?.table !== node.name && linkingFrom !== null}
                                    selectedColumn={selectedColumn}
                                    onMouseDown={(e) => handleNodeMouseDown(node.id, e)}
                                    onColumnClick={handleColumnClick}
                                />
                            ))}
                        </div>
                    </div>

                    {/* Empty state */}
                    {tables.length === 0 && (
                        <div className="absolute inset-0 flex items-center justify-center">
                            <div className="text-center space-y-2">
                                <Database className="w-10 h-10 text-muted-foreground/20 mx-auto" />
                                <p className="text-sm text-muted-foreground/40">No tables loaded</p>
                                <p className="text-[11px] text-muted-foreground/30">Connect a data source to visualize relationships</p>
                            </div>
                        </div>
                    )}
                </div>
            </div>
        </Card>
    );
}
