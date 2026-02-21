"use client";

import { useEffect, useState } from "react";
import { SlideCanvas } from "@/components/story-builder/SlideCanvas";
import { Story } from "@/types/story";
import { Loader2, ChevronLeft, ChevronRight } from "lucide-react";
import { Button } from "@/components/ui/button";
import { ThemeProvider } from "@/components/theme-provider";

export default function PresentClient({ token }: { token: string }) {
  const [story, setStory] = useState<Story | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [activeIndex, setActiveIndex] = useState(0);

  useEffect(() => {
    fetch(`/api/public/stories/${token}`)
      .then((res) => {
        if (!res.ok) throw new Error("Failed to load story");
        return res.json();
      })
      .then(setStory)
      .catch(() => setError("Failed to load story. The link may be invalid or no longer public."))
      .finally(() => setLoading(false));
  }, [token]);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (!story) return;
      const max = (story.content?.slides?.length ?? 0) - 1;
      if (e.key === "ArrowRight" || e.key === "Space") {
        setActiveIndex((i) => Math.min(i + 1, max));
      } else if (e.key === "ArrowLeft") {
        setActiveIndex((i) => Math.max(i - 1, 0));
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [story]);

  if (loading) {
    return (
      <div className="flex h-screen items-center justify-center bg-background">
        <Loader2 className="w-8 h-8 animate-spin text-primary" />
      </div>
    );
  }

  if (error || !story) {
    return (
      <div className="flex h-screen items-center justify-center bg-background text-destructive text-center p-6">
        {error || "Story not found"}
      </div>
    );
  }

  const slides = story.content?.slides ?? [];
  if (slides.length === 0) {
    return (
      <div className="flex h-screen items-center justify-center bg-background text-muted-foreground">
        No slides in this story.
      </div>
    );
  }

  const activeSlide = slides[activeIndex];

  return (
    <ThemeProvider attribute="class" defaultTheme="system" enableSystem>
      <div className="relative h-screen w-screen bg-black overflow-hidden flex flex-col items-center justify-center">
        <div className="w-full h-full max-w-[1920px] max-h-[1080px] p-0 md:p-4 flex items-center justify-center bg-black">
          <SlideCanvas slide={activeSlide} isPresentMode={true} />
        </div>

        {/* Presentation Controls overlay */}
        <div className="absolute bottom-6 left-1/2 -translate-x-1/2 bg-background/80 backdrop-blur-md border border-border/40 px-6 py-3 rounded-full flex items-center gap-6 shadow-2xl opacity-0 hover:opacity-100 transition-opacity duration-300">
          <Button
            variant="ghost"
            size="icon"
            onClick={() => setActiveIndex((i) => Math.max(i - 1, 0))}
            disabled={activeIndex === 0}
            className="rounded-full"
          >
            <ChevronLeft className="w-5 h-5" />
          </Button>
          <span className="text-sm font-semibold tabular-nums text-foreground">
            {activeIndex + 1} / {slides.length}
          </span>
          <Button
            variant="ghost"
            size="icon"
            onClick={() => setActiveIndex((i) => Math.min(i + 1, slides.length - 1))}
            disabled={activeIndex === slides.length - 1}
            className="rounded-full"
          >
            <ChevronRight className="w-5 h-5" />
          </Button>
        </div>
      </div>
    </ThemeProvider>
  );
}
