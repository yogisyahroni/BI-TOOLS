import React from "react";
import { type Slide } from "@/types/presentation";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

interface SlideProps {
  slide: Slide;
  chartComponent?: React.ReactNode;
}

export function SlideView({ slide, chartComponent }: SlideProps) {
  switch (slide.layout) {
    case "title_only":
      return (
        <div className="h-full flex flex-col items-center justify-center p-8 bg-background border rounded-lg shadow-sm">
          <h1 className="text-4xl font-bold text-center text-primary">{slide.title}</h1>
        </div>
      );
    case "bullet_points":
      return (
        <div className="h-full flex flex-col p-8 bg-background border rounded-lg shadow-sm">
          <h2 className="text-3xl font-bold mb-6 text-primary">{slide.title}</h2>
          <ul className="list-disc pl-8 space-y-4">
            {slide.bullet_points?.map((point, index) => (
              <li key={index} className="text-xl text-muted-foreground">
                {point}
              </li>
            ))}
          </ul>
        </div>
      );
    case "chart_focus":
      return (
        <div className="h-full flex flex-col p-8 bg-background border rounded-lg shadow-sm">
          <h2 className="text-2xl font-bold mb-4 text-primary">{slide.title}</h2>
          <div className="flex-grow flex items-center justify-center bg-muted/20 rounded-md p-4">
            {chartComponent ? (
              <div className="w-full h-full min-h-[400px]">{chartComponent}</div>
            ) : (
              <div className="text-muted-foreground italic">Chart Placeholder</div>
            )}
          </div>
        </div>
      );
    case "title_and_body":
    default:
      return (
        <div className="h-full flex flex-col p-8 bg-background border rounded-lg shadow-sm">
          <h2 className="text-3xl font-bold mb-6 text-primary">{slide.title}</h2>
          <div className="flex-grow">
            {slide.bullet_points && (
              <ul className="list-disc pl-8 space-y-4 mb-4">
                {slide.bullet_points.map((point, index) => (
                  <li key={index} className="text-lg text-muted-foreground">
                    {point}
                  </li>
                ))}
              </ul>
            )}
            {chartComponent && <div className="mt-4 h-[300px]">{chartComponent}</div>}
          </div>
        </div>
      );
  }
}
