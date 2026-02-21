"use client";

import { Slide } from "@/types/story";
import { LayoutTemplate, AlignLeft, Image as ImageIcon, BarChart3 } from "lucide-react";
import { ChartVisualization } from "@/components/chart-visualization";

export function SlideCanvas({
  slide,
  isPresentMode = false,
}: {
  slide: Slide;
  isPresentMode?: boolean;
}) {
  if (!slide) {
    return (
      <div className="flex-1 flex flex-col items-center justify-center bg-muted/20">
        <LayoutTemplate className="h-12 w-12 text-muted-foreground/30 mb-4" />
        <p className="text-muted-foreground">Select a slide to edit</p>
      </div>
    );
  }

  const LayoutIcon = () => {
    switch (slide.layout) {
      case "title":
        return <AlignLeft className="h-16 w-16 text-primary/20 mb-6" />;
      case "bullet_points":
        return <AlignLeft className="h-16 w-16 text-primary/20 mb-6" />;
      case "image_text":
        return <ImageIcon className="h-16 w-16 text-primary/20 mb-6" />;
      case "chart":
        return <BarChart3 className="h-16 w-16 text-primary/20 mb-6" />;
      default:
        return <LayoutTemplate className="h-16 w-16 text-primary/20 mb-6" />;
    }
  };

  return (
    <div
      className={`flex-1 overflow-auto ${isPresentMode ? "bg-black p-0" : "bg-muted/20 p-8"} flex items-center justify-center`}
    >
      {/* Aspect Ratio 16:9 Presentation Canvas */}
      <div
        className={`w-full ${isPresentMode ? "h-full max-w-none rounded-none border-none aspect-video" : "max-w-4xl aspect-video border rounded-lg"} bg-background shadow-sm flex flex-col overflow-hidden relative transition-all duration-300`}
      >
        <div className="p-12 flex flex-col h-full items-center justify-center text-center">
          <LayoutIcon />

          <h1 className="text-4xl font-bold tracking-tight mb-8">
            {slide.title || "Untitled Slide"}
          </h1>

          {slide.layout === "bullet_points" && (
            <div className="text-left w-full max-w-2xl text-lg text-muted-foreground space-y-4">
              {slide.content
                .split("\n")
                .filter(Boolean)
                .map((line, i) => (
                  <div key={i} className="flex items-start gap-3">
                    <span className="text-primary mt-1">{"\u2022"}</span>
                    <span>{line.replace(/^[-\u2022]\s*/, "")}</span>
                  </div>
                ))}
            </div>
          )}

          {slide.layout === "chart" && slide.data_binding?.card_id && !slide.query_result && (
            <div className="w-full max-w-3xl h-64 border-2 border-dashed border-primary/20 rounded-xl flex items-center justify-center bg-primary/5 text-primary">
              <BarChart3 className="mr-2 h-5 w-5" />
              Live Data Bound: Dashboard {slide.data_binding.dashboard_id}, Card{" "}
              {slide.data_binding.card_id}
            </div>
          )}

          {slide.layout === "chart" && slide.query_result && slide.visualization_config && (
            <div className="w-full h-full max-w-4xl max-h-[600px] flex items-center justify-center pt-8">
              <ChartVisualization
                config={slide.visualization_config}
                data={slide.query_result.data || []}
                chartId={`slide-chart-${slide.data_binding?.card_id}`}
              />
            </div>
          )}

          {slide.layout === "chart" && slide.query_error && (
            <div className="w-full max-w-3xl h-64 border-2 border-dashed border-destructive/20 rounded-xl flex items-center justify-center bg-destructive/5 text-destructive">
              Failed to load live chart data: {slide.query_error}
            </div>
          )}

          {slide.layout === "chart" && !slide.data_binding?.card_id && (
            <div className="w-full max-w-3xl h-64 border-2 border-dashed border-muted rounded-xl flex flex-col items-center justify-center text-muted-foreground gap-2">
              <BarChart3 className="h-8 w-8 opacity-50" />
              <span>Select a Chart from the Properties sidebar to bind data.</span>
            </div>
          )}
        </div>

        {/* Speaker Notes Overlay indicator */}
        {slide.notes && (
          <div className="absolute bottom-0 left-0 right-0 bg-yellow-500/10 border-t border-yellow-500/20 p-3 text-xs text-yellow-700 dark:text-yellow-500/80">
            <strong>Notes:</strong> {slide.notes}
          </div>
        )}
      </div>
    </div>
  );
}
