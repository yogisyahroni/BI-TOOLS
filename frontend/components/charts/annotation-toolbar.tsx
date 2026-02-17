'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import {
    Popover,
    PopoverContent,
    PopoverTrigger,
} from '@/components/ui/popover';
import { 
    MapPin, 
    _MousePointer2, 
    Type, 
    MoveHorizontal,
    Palette,
    MessageSquare,
    X
} from 'lucide-react';
import type { _Annotation } from '@/types/comments';

type AnnotationType = 'point' | 'range' | 'text';

interface AnnotationToolbarProps {
    isAnnotationMode: boolean;
    onToggleMode: (enabled: boolean) => void;
    selectedType: AnnotationType;
    onSelectType: (type: AnnotationType) => void;
    selectedColor: string;
    onSelectColor: (color: string) => void;
    annotationCount: number;
}

const ANNOTATION_TYPES: { type: AnnotationType; label: string; icon: typeof MapPin; description: string }[] = [
    { 
        type: 'point', 
        label: 'Point', 
        icon: MapPin, 
        description: 'Mark a specific data point' 
    },
    { 
        type: 'range', 
        label: 'Range', 
        icon: MoveHorizontal, 
        description: 'Highlight a range of values' 
    },
    { 
        type: 'text', 
        label: 'Text', 
        icon: Type, 
        description: 'Add a text annotation' 
    },
];

const COLORS = [
    { name: 'Amber', value: '#F59E0B', class: 'bg-amber-500' },
    { name: 'Red', value: '#EF4444', class: 'bg-red-500' },
    { name: 'Green', value: '#10B981', class: 'bg-emerald-500' },
    { name: 'Blue', value: '#3B82F6', class: 'bg-blue-500' },
    { name: 'Purple', value: '#8B5CF6', class: 'bg-violet-500' },
    { name: 'Pink', value: '#EC4899', class: 'bg-pink-500' },
    { name: 'Cyan', value: '#06B6D4', class: 'bg-cyan-500' },
    { name: 'Gray', value: '#6B7280', class: 'bg-gray-500' },
];

export function AnnotationToolbar({
    isAnnotationMode,
    onToggleMode,
    selectedType,
    onSelectType,
    selectedColor,
    onSelectColor,
    annotationCount,
}: AnnotationToolbarProps) {
    const [showColorPicker, setShowColorPicker] = useState(false);

    const selectedTypeInfo = ANNOTATION_TYPES.find(t => t.type === selectedType);
    const selectedColorInfo = COLORS.find(c => c.value === selectedColor) || COLORS[0];

    return (
        <TooltipProvider delayDuration={300}>
            <div className="flex items-center gap-2 p-2 bg-background border rounded-lg shadow-sm">
                {/* Annotation Mode Toggle */}
                <Tooltip>
                    <TooltipTrigger asChild>
                        <Button
                            variant={isAnnotationMode ? "default" : "outline"}
                            size="sm"
                            className="gap-2"
                            onClick={() => onToggleMode(!isAnnotationMode)}
                        >
                            {isAnnotationMode ? (
                                <>
                                    <X className="w-4 h-4" />
                                    Exit
                                </>
                            ) : (
                                <>
                                    <MapPin className="w-4 h-4" />
                                    Annotate
                                    {annotationCount > 0 && (
                                        <Badge variant="secondary" className="ml-1 text-xs">
                                            {annotationCount}
                                        </Badge>
                                    )}
                                </>
                            )}
                        </Button>
                    </TooltipTrigger>
                    <TooltipContent>
                        <p>{isAnnotationMode ? 'Exit annotation mode' : 'Add chart annotations'}</p>
                    </TooltipContent>
                </Tooltip>

                {isAnnotationMode && (
                    <>
                        <div className="w-px h-6 bg-border mx-1" />

                        {/* Annotation Type Selector */}
                        <Popover>
                            <Tooltip>
                                <TooltipTrigger asChild>
                                    <PopoverTrigger asChild>
                                        <Button variant="outline" size="sm" className="gap-2">
                                            {selectedTypeInfo && (
                                                <selectedTypeInfo.icon className="w-4 h-4" />
                                            )}
                                            <span className="capitalize">{selectedType}</span>
                                        </Button>
                                    </PopoverTrigger>
                                </TooltipTrigger>
                                <TooltipContent>
                                    <p>Select annotation type</p>
                                </TooltipContent>
                            </Tooltip>
                            <PopoverContent className="w-48 p-2">
                                <div className="space-y-1">
                                    {ANNOTATION_TYPES.map(({ type, label, icon: Icon, description }) => (
                                        <button
                                            key={type}
                                            onClick={() => onSelectType(type)}
                                            className={`
                                                w-full flex items-start gap-3 p-2 rounded-md text-left
                                                transition-colors hover:bg-muted
                                                ${selectedType === type ? 'bg-muted' : ''}
                                            `}
                                        >
                                            <Icon className="w-4 h-4 mt-0.5 text-muted-foreground" />
                                            <div>
                                                <div className="text-sm font-medium">{label}</div>
                                                <div className="text-xs text-muted-foreground">{description}</div>
                                            </div>
                                        </button>
                                    ))}
                                </div>
                            </PopoverContent>
                        </Popover>

                        {/* Color Picker */}
                        <Popover open={showColorPicker} onOpenChange={setShowColorPicker}>
                            <Tooltip>
                                <TooltipTrigger asChild>
                                    <PopoverTrigger asChild>
                                        <Button variant="outline" size="sm" className="gap-2">
                                            <Palette className="w-4 h-4" />
                                            <div 
                                                className={`w-4 h-4 rounded-full ${selectedColorInfo.class}`}
                                            />
                                        </Button>
                                    </PopoverTrigger>
                                </TooltipTrigger>
                                <TooltipContent>
                                    <p>Choose annotation color</p>
                                </TooltipContent>
                            </Tooltip>
                            <PopoverContent className="w-56 p-3">
                                <div className="space-y-2">
                                    <div className="text-sm font-medium">Annotation Color</div>
                                    <div className="grid grid-cols-4 gap-2">
                                        {COLORS.map(({ name, value, class: colorClass }) => (
                                            <Tooltip key={value}>
                                                <TooltipTrigger asChild>
                                                    <button
                                                        onClick={() => {
                                                            onSelectColor(value);
                                                            setShowColorPicker(false);
                                                        }}
                                                        className={`
                                                            w-8 h-8 rounded-full border-2 transition-all
                                                            ${selectedColor === value 
                                                                ? 'border-foreground scale-110' 
                                                                : 'border-transparent hover:scale-105'
                                                            }
                                                            ${colorClass}
                                                        `}
                                                    />
                                                </TooltipTrigger>
                                                <TooltipContent>
                                                    <p>{name}</p>
                                                </TooltipContent>
                                            </Tooltip>
                                        ))}
                                    </div>
                                </div>
                            </PopoverContent>
                        </Popover>

                        <div className="w-px h-6 bg-border mx-1" />

                        {/* Instructions */}
                        <div className="text-xs text-muted-foreground hidden sm:block">
                            Click on the chart to add annotation
                        </div>
                    </>
                )}

                {!isAnnotationMode && annotationCount > 0 && (
                    <>
                        <div className="w-px h-6 bg-border mx-1" />
                        <div className="flex items-center gap-2 text-sm text-muted-foreground">
                            <MessageSquare className="w-4 h-4" />
                            <span>{annotationCount} annotation{annotationCount !== 1 ? 's' : ''}</span>
                        </div>
                    </>
                )}
            </div>
        </TooltipProvider>
    );
}

// Compact version for smaller spaces
interface CompactAnnotationToolbarProps {
    isAnnotationMode: boolean;
    onToggleMode: (enabled: boolean) => void;
    annotationCount: number;
}

export function CompactAnnotationToolbar({
    isAnnotationMode,
    onToggleMode,
    annotationCount,
}: CompactAnnotationToolbarProps) {
    return (
        <TooltipProvider delayDuration={300}>
            <Tooltip>
                <TooltipTrigger asChild>
                    <Button
                        variant={isAnnotationMode ? "default" : "ghost"}
                        size="sm"
                        className="relative"
                        onClick={() => onToggleMode(!isAnnotationMode)}
                    >
                        <MapPin className="w-4 h-4" />
                        {annotationCount > 0 && !isAnnotationMode && (
                            <Badge 
                                variant="secondary" 
                                className="absolute -top-1 -right-1 h-4 min-w-4 text-[10px] p-0 flex items-center justify-center"
                            >
                                {annotationCount}
                            </Badge>
                        )}
                    </Button>
                </TooltipTrigger>
                <TooltipContent>
                    <p>{isAnnotationMode ? 'Exit annotation mode' : 'Annotate chart'}</p>
                </TooltipContent>
            </Tooltip>
        </TooltipProvider>
    );
}
