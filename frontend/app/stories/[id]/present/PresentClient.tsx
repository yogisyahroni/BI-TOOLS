"use client";

import { useEffect, useState } from "react";
import { SlideCanvas } from "@/components/story-builder/SlideCanvas";
import { storyService } from "@/services/storyService";
import { Story } from "@/types/story";
import { Loader2, X, ChevronLeft, ChevronRight } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useRouter } from "next/navigation";
import { ThemeProvider } from "@/components/theme-provider";

export function PresentClient({ id }: { id: string }) {
  const [story, setStory] = useState<Story | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [activeIndex, setActiveIndex] = useState(0);
  const router = useRouter();

  useEffect(() => {
    storyService
      .getStory(id)
      .then(setStory)
      .catch(() => setError("Failed to load story"))
      .finally(() => setLoading(false));
  }, [id]);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (!story) return;
      const max = (story.content?.slides?.length ?? 0) - 1;
      if (e.key === "ArrowRight" || e.key === "Space") {
        setActiveIndex((i) => Math.min(i + 1, max));
      } else if (e.key === "ArrowLeft") {
        setActiveIndex((i) => Math.max(i - 1, 0));
      } else if (e.key === "Escape") {
        router.push("/stories");
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [story, router]);

  if (loading) {
    return (
      <div className="flex h-screen items-center justify-center bg-background">
        <Loader2 className="w-8 h-8 animate-spin text-primary" />
      </div>
    );
  }

  if (error || !story) {
    return (
      <div className="flex h-screen items-center justify-center bg-background text-destructive">
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

        {/* Exit button */}
        <Button
          variant="ghost"
          size="icon"
          onClick={() => router.push("/stories")}
          className="absolute top-6 right-6 rounded-full bg-background/20 hover:bg-background/40 text-foreground/50 border border-white/10 shadow-sm opacity-0 hover:opacity-100 transition-opacity duration-300"
          title="Exit Fullscreen (Esc)"
        >
          <X className="w-5 h-5" />
        </Button>
      </div>
    </ThemeProvider>
  );
}
