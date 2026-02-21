"use client";

import { useState, useRef, useCallback, useEffect } from "react";
import { MessageSquare, X, Edit2, Trash2, MapPin, MoreHorizontal } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { formatDistanceToNow } from "date-fns";
import type { Comment, Annotation, AnnotationPosition } from "@/types/comments";

interface AnnotationMarkerProps {
  annotation: Annotation;
  comment: Comment;
  isSelected: boolean;
  onClick: () => void;
  onEdit: () => void;
  onDelete: () => void;
  currentUserId: string;
}

function AnnotationMarker({
  annotation,
  comment,
  isSelected,
  onClick,
  onEdit,
  onDelete,
  currentUserId,
}: AnnotationMarkerProps) {
  const [isHovered, setIsHovered] = useState(false);
  const isOwner = comment.userId === currentUserId;

  const getInitials = (name: string) => {
    return name
      .split(" ")
      .map((n) => n[0])
      .join("")
      .toUpperCase()
      .slice(0, 2);
  };

  const user = comment.user || { name: "Unknown" };

  return (
    <div
      className="absolute transform -translate-x-1/2 -translate-y-1/2 z-10"
      style={{
        left: annotation.position.x,
        top: annotation.position.y,
      }}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
    >
      <Popover open={isSelected} onOpenChange={(open) => !open && onClick()}>
        <PopoverTrigger asChild>
          <button
            onClick={onClick}
            className={`
                            relative flex items-center justify-center w-6 h-6 rounded-full
                            border-2 border-white shadow-md transition-all duration-200
                            hover:scale-110 focus:outline-none focus:ring-2 focus:ring-offset-2
                            ${isSelected ? "ring-2 ring-offset-2 ring-primary scale-110" : ""}
                        `}
            style={{ backgroundColor: annotation.color }}
          >
            <MapPin className="w-3 h-3 text-white" />

            {/* Hover tooltip */}
            {isHovered && !isSelected && (
              <div className="absolute bottom-full left-1/2 transform -translate-x-1/2 mb-2 px-2 py-1 bg-black text-white text-xs rounded whitespace-nowrap pointer-events-none">
                {user.name}&apos;s annotation
              </div>
            )}
          </button>
        </PopoverTrigger>

        <PopoverContent className="w-80 p-0" align="center" side="top" sideOffset={10}>
          <div className="p-3 border-b bg-muted/50 flex items-center justify-between">
            <div className="flex items-center gap-2">
              <div className="w-3 h-3 rounded-full" style={{ backgroundColor: annotation.color }} />
              <span className="text-sm font-medium">
                {annotation.type === "point"
                  ? "Point"
                  : annotation.type === "range"
                    ? "Range"
                    : "Text"}{" "}
                Annotation
              </span>
            </div>
            <div className="flex items-center gap-1">
              {isOwner && (
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="ghost" size="sm" className="h-7 w-7 p-0">
                      <MoreHorizontal className="w-4 h-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem onClick={onEdit}>
                      <Edit2 className="w-4 h-4 mr-2" />
                      Edit
                    </DropdownMenuItem>
                    <DropdownMenuItem onClick={onDelete} className="text-destructive">
                      <Trash2 className="w-4 h-4 mr-2" />
                      Delete
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              )}
              <Button variant="ghost" size="sm" className="h-7 w-7 p-0" onClick={() => onClick()}>
                <X className="w-4 h-4" />
              </Button>
            </div>
          </div>

          <div className="p-3">
            <div className="flex items-center gap-2 mb-2">
              <div className="w-6 h-6 rounded-full bg-primary/10 flex items-center justify-center text-xs font-medium">
                {getInitials(user.name)}
              </div>
              <div>
                <span className="text-sm font-medium">{user.name}</span>
                <span className="text-xs text-muted-foreground ml-2">
                  {formatDistanceToNow(new Date(comment.createdAt))} ago
                </span>
              </div>
            </div>

            <p className="text-sm whitespace-pre-wrap">{comment.content}</p>

            {(annotation.xValue !== undefined || annotation.xCategory) && (
              <div className="mt-3 pt-3 border-t text-xs text-muted-foreground">
                <div className="flex items-center gap-2">
                  <span>Position:</span>
                  {annotation.xValue !== undefined && (
                    <Badge variant="secondary" className="text-xs">
                      X: {annotation.xValue.toFixed(2)}
                    </Badge>
                  )}
                  {annotation.yValue !== undefined && (
                    <Badge variant="secondary" className="text-xs">
                      Y: {annotation.yValue.toFixed(2)}
                    </Badge>
                  )}
                  {annotation.xCategory && (
                    <Badge variant="secondary" className="text-xs">
                      {annotation.xCategory}
                    </Badge>
                  )}
                </div>
              </div>
            )}
          </div>
        </PopoverContent>
      </Popover>
    </div>
  );
}

interface ChartAnnotationsProps {
  chartId: string;
  comments: Comment[];
  currentUserId: string;
  isAnnotationMode: boolean;
  selectedAnnotationType: "point" | "range" | "text";
  selectedColor: string;
  onAnnotationClick: (position: AnnotationPosition, xValue?: number, yValue?: number) => void;
  onEditAnnotation: (comment: Comment) => void;
  onDeleteAnnotation: (annotationId: string) => void;
  selectedAnnotationId: string | null;
  onSelectAnnotation: (id: string | null) => void;
  children: React.ReactNode;
}

export function ChartAnnotations({
  chartId,
  comments,
  currentUserId,
  isAnnotationMode,
  selectedAnnotationType,
  selectedColor,
  onAnnotationClick,
  onEditAnnotation,
  onDeleteAnnotation,
  selectedAnnotationId,
  onSelectAnnotation,
  children,
}: ChartAnnotationsProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const [hoverPosition, setHoverPosition] = useState<AnnotationPosition | null>(null);

  // Filter comments that have annotations for this chart
  const annotations = comments.filter((c) => c.annotation && c.annotation.chartId === chartId);

  const handleChartClick = useCallback(
    (e: React.MouseEvent<HTMLDivElement>) => {
      if (!isAnnotationMode || !containerRef.current) return;

      const rect = containerRef.current.getBoundingClientRect();
      const x = e.clientX - rect.left;
      const y = e.clientY - rect.top;

      // Calculate relative position (0-1) for data value lookup
      const relativeX = x / rect.width;
      const relativeY = 1 - y / rect.height; // Invert Y for chart coordinates

      // TODO: In a real implementation, you'd convert these to actual data values
      // based on the chart's scale and domain
      const xValue = relativeX * 100;
      const yValue = relativeY * 100;

      onAnnotationClick({ x, y }, xValue, yValue);
    },
    [isAnnotationMode, onAnnotationClick],
  );

  const handleMouseMove = useCallback(
    (e: React.MouseEvent<HTMLDivElement>) => {
      if (!isAnnotationMode || !containerRef.current) return;

      const rect = containerRef.current.getBoundingClientRect();
      setHoverPosition({
        x: e.clientX - rect.left,
        y: e.clientY - rect.top,
      });
    },
    [isAnnotationMode],
  );

  const handleMouseLeave = useCallback(() => {
    setHoverPosition(null);
  }, []);

  return (
    <div
      ref={containerRef}
      className={`relative w-full h-full ${isAnnotationMode ? "cursor-crosshair" : ""}`}
      onClick={handleChartClick}
      onMouseMove={handleMouseMove}
      onMouseLeave={handleMouseLeave}
    >
      {/* Chart content */}
      <div className="w-full h-full">{children}</div>

      {/* Annotation Layer */}
      <div className="absolute inset-0 pointer-events-none">
        {/* Render existing annotations */}
        {annotations.map((comment) =>
          comment.annotation ? (
            <div key={comment.annotation.id} className="pointer-events-auto">
              <AnnotationMarker
                annotation={comment.annotation}
                comment={comment}
                isSelected={selectedAnnotationId === comment.annotation.id}
                onClick={() =>
                  onSelectAnnotation(
                    selectedAnnotationId === comment.annotation!.id ? null : comment.annotation!.id,
                  )
                }
                onEdit={() => onEditAnnotation(comment)}
                onDelete={() => onDeleteAnnotation(comment.annotation!.id)}
                currentUserId={currentUserId}
              />
            </div>
          ) : null,
        )}

        {/* Hover preview in annotation mode */}
        {isAnnotationMode && hoverPosition && (
          <div
            className="absolute pointer-events-none transform -translate-x-1/2 -translate-y-1/2"
            style={{
              left: hoverPosition.x,
              top: hoverPosition.y,
            }}
          >
            <div
              className="w-6 h-6 rounded-full border-2 border-white shadow-md opacity-50"
              style={{ backgroundColor: selectedColor }}
            />
            <div className="absolute top-full left-1/2 transform -translate-x-1/2 mt-1 px-2 py-1 bg-black text-white text-xs rounded whitespace-nowrap">
              Click to add {selectedAnnotationType} annotation
            </div>
          </div>
        )}
      </div>

      {/* Annotation count badge */}
      {annotations.length > 0 && !isAnnotationMode && (
        <div className="absolute top-2 right-2">
          <Badge variant="secondary" className="bg-white/90 shadow-sm">
            <MessageSquare className="w-3 h-3 mr-1" />
            {annotations.length} {annotations.length === 1 ? "annotation" : "annotations"}
          </Badge>
        </div>
      )}

      {/* Annotation mode indicator */}
      {isAnnotationMode && (
        <div className="absolute top-2 left-1/2 transform -translate-x-1/2">
          <Badge className="bg-primary text-white shadow-lg">
            <MapPin className="w-3 h-3 mr-1" />
            Annotation Mode: Click chart to add
          </Badge>
        </div>
      )}
    </div>
  );
}
