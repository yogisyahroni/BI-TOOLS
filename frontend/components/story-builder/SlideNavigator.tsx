"use client";

import { Slide } from "@/types/story";
import { Button } from "@/components/ui/button";
import { Plus, GripVertical, Trash2, ChevronUp, ChevronDown } from "lucide-react";

export function SlideNavigator({
  slides,
  activeIndex,
  onSelect,
  onAdd,
  onDelete,
  onMoveUp,
  onMoveDown,
}: {
  slides: Slide[];
  activeIndex: number;
  onSelect: (index: number) => void;
  onAdd: () => void;
  onDelete: (index: number) => void;
  onMoveUp: (index: number) => void;
  onMoveDown: (index: number) => void;
}) {
  return (
    <div className="w-64 border-r flex flex-col bg-background shrink-0">
      <div className="p-4 border-b flex items-center justify-between">
        <h3 className="font-semibold tracking-tight text-sm">Slides</h3>
        <Button size="icon" variant="ghost" className="h-7 w-7" onClick={onAdd}>
          <Plus className="h-4 w-4" />
        </Button>
      </div>
      <div className="flex-1 overflow-y-auto p-3 space-y-2">
        {slides.map((slide, index) => (
          <div
            key={index}
            onClick={() => onSelect(index)}
            className={`group relative p-3 pb-6 rounded-lg border-2 cursor-pointer transition-all duration-200 ${
              index === activeIndex
                ? "border-primary bg-primary/5 shadow-sm"
                : "border-transparent hover:border-border/50 bg-muted/40 hover:bg-muted/80"
            }`}
          >
            {/* Slide Thumbnail Preview Text */}
            <div className="text-[10px] font-bold text-muted-foreground mb-1 uppercase tracking-wider">
              {slide.layout.replace("_", " ")}
            </div>
            <div className="text-sm font-medium line-clamp-2 leading-snug">
              {slide.title || "Untitled Slide"}
            </div>

            {/* Hover Controls */}
            <div className="absolute top-2 right-2 flex flex-col gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
              <Button
                variant="secondary"
                size="icon"
                className="h-6 w-6 shadow-sm bg-background/80 hover:bg-background"
                disabled={index === 0}
                onClick={(e) => {
                  e.stopPropagation();
                  onMoveUp(index);
                }}
              >
                <ChevronUp className="h-3 w-3" />
              </Button>
              <Button
                variant="secondary"
                size="icon"
                className="h-6 w-6 shadow-sm bg-background/80 hover:bg-background flex items-center justify-center"
                disabled={index === slides.length - 1}
                onClick={(e) => {
                  e.stopPropagation();
                  onMoveDown(index);
                }}
              >
                <ChevronDown className="h-3 w-3" />
              </Button>
              <Button
                variant="destructive"
                size="icon"
                className="h-6 w-6 shadow-sm"
                onClick={(e) => {
                  e.stopPropagation();
                  onDelete(index);
                }}
              >
                <Trash2 className="h-3 w-3" />
              </Button>
            </div>

            {/* Slide Number */}
            <div className="absolute bottom-2 left-3 text-[10px] text-muted-foreground/60 font-mono">
              {String(index + 1).padStart(2, "0")}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
