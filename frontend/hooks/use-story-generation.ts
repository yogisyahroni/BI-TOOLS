import { useState, useCallback } from "react";
import { useMutation } from "@tanstack/react-query";
import { presentationApi } from "@/lib/api/presentation";
import { exportToPPT } from "@/lib/export-ppt";
import { Slide } from "@/types/presentation";
import { toast } from "sonner";

interface UseStoryGenerationProps {
  dashboardId?: string;
  initialSlides?: Slide[];
}

export function useStoryGeneration({ dashboardId, initialSlides = [] }: UseStoryGenerationProps) {
  const [slides, setSlides] = useState<Slide[]>(initialSlides);
  const [currentSlideIndex, setCurrentSlideIndex] = useState(0);
  const [storyTitle, setStoryTitle] = useState("New Data Story");

  const generateMutation = useMutation({
    mutationFn: async ({ prompt, providerId }: { prompt: string; providerId?: string }) => {
      if (!dashboardId) throw new Error("No dashboard context available.");
      if (!prompt.trim()) throw new Error("Please enter a prompt.");

      return await presentationApi.generate(dashboardId, prompt, providerId);
    },
    onSuccess: (data) => {
      setSlides(data.slides);
      setStoryTitle(data.title || "Generated Story");
      setCurrentSlideIndex(0);
      toast.success("Story generated successfully!");
    },
    onError: (error) => {
      console.error("Failed to generate story:", error);
      // Don't show generic error toast if it's our validation error
      if (
        error instanceof Error &&
        (error.message === "No dashboard context available." ||
          error.message === "Please enter a prompt.")
      ) {
        toast.error(error.message);
      } else {
        toast.error("Failed to generate story. Please try again.");
      }
    },
  });

  const generateStory = useCallback(
    async (prompt: string, providerId?: string) => {
      try {
        await generateMutation.mutateAsync({ prompt, providerId });
        return true;
      } catch {
        return false;
      }
    },
    [generateMutation],
  );

  const exportStory = useCallback(async () => {
    if (!slides.length) return;
    try {
      await exportToPPT({
        title: storyTitle,
        description: "Exported from Story Builder",
        slides,
      });
      toast.success("Export started!");
    } catch (error) {
      console.error("Export failed:", error);
      toast.error("Export failed. Please try again.");
    }
  }, [slides, storyTitle]);

  const addSlide = useCallback(() => {
    const newSlide: Slide = {
      title: "New Slide",
      layout: "title_and_body",
      bullet_points: ["Add your content here..."],
    };
    setSlides((prev) => {
      const updated = [...prev, newSlide];
      setCurrentSlideIndex(updated.length - 1);
      return updated;
    });
  }, []);

  const updateSlide = useCallback((index: number, updatedSlide: Partial<Slide>) => {
    setSlides((prev) => {
      const newSlides = [...prev];
      newSlides[index] = { ...newSlides[index], ...updatedSlide };
      return newSlides;
    });
  }, []);

  const deleteSlide = useCallback(
    (index: number) => {
      setSlides((prev) => {
        const newSlides = prev.filter((_, i) => i !== index);
        if (currentSlideIndex >= newSlides.length) {
          setCurrentSlideIndex(Math.max(0, newSlides.length - 1));
        }
        return newSlides;
      });
    },
    [currentSlideIndex],
  );

  return {
    slides,
    currentSlideIndex,
    isGenerating: generateMutation.isPending,
    storyTitle,
    setStoryTitle,
    setCurrentSlideIndex,
    generateStory,
    exportStory,
    addSlide,
    updateSlide,
    deleteSlide,
  };
}
