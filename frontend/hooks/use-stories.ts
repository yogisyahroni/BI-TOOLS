"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { storyService } from "@/services/storyService";
import { toast } from "sonner";
import type { Story, Slide } from "@/types/story";

export function useStories(dashboardId?: string) {
  const queryClient = useQueryClient();
  const queryKey = ["stories", dashboardId];

  // Query: List stories
  const {
    data: stories = [],
    isLoading,
    error,
  } = useQuery({
    queryKey,
    queryFn: () => storyService.getStories(), // Assuming getStories gets all for now
  });

  // Mutation: Create Story (Manual)
  const createManualMutation = useMutation({
    mutationFn: ({ title, description }: { title: string; description?: string }) =>
      storyService.createManualStory({ title, description }),
    onMutate: async (newStoryVars) => {
      await queryClient.cancelQueries({ queryKey });
      const previousStories = queryClient.getQueryData(queryKey);

      const tempId = `temp-${Date.now()}`;
      const optimisticStory: any = {
        id: tempId,
        title: newStoryVars.title,
        description: newStoryVars.description || "",
        content: { slides: [] },
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        _optimistic: true,
        _status: "pending",
      };

      queryClient.setQueryData(queryKey, (old: any) => [optimisticStory, ...(old || [])]);

      return { previousStories, tempId };
    },
    onSuccess: (realStory, variables, context) => {
      // Normalize
      realStory.content = realStory.content || { slides: [] };
      queryClient.setQueryData(queryKey, (old: any) =>
        (old || []).map((item: any) =>
          item.id === context?.tempId ? { ...realStory, _optimistic: false } : item,
        ),
      );
      toast.success("Story created successfully");
    },
    onError: (error: Error, variables, context) => {
      if (context?.previousStories) {
        queryClient.setQueryData(queryKey, context.previousStories);
      }
      toast.error(`Failed to create story: ${error.message}`);
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey });
    },
  });

  // Mutation: Create Story (AI)
  const createAIMutation = useMutation({
    mutationFn: ({ prompt }: { prompt: string }) =>
      storyService.createStory({ dashboard_id: "dashboard-123", prompt }), // Example
    onMutate: async (newStoryVars) => {
      await queryClient.cancelQueries({ queryKey });
      const previousStories = queryClient.getQueryData(queryKey);

      const tempId = `temp-${Date.now()}`;
      const optimisticStory: any = {
        id: tempId,
        title: "Generating AI Story...",
        description: "This might take a moment.",
        content: { slides: [] },
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        _optimistic: true,
        _status: "pending",
      };

      queryClient.setQueryData(queryKey, (old: any) => [optimisticStory, ...(old || [])]);

      return { previousStories, tempId };
    },
    onSuccess: (realStory, variables, context) => {
      realStory.content = realStory.content || { slides: [] };
      queryClient.setQueryData(queryKey, (old: any) =>
        (old || []).map((item: any) =>
          item.id === context?.tempId ? { ...realStory, _optimistic: false } : item,
        ),
      );
      toast.success("AI Story generated");
    },
    onError: (error: Error, variables, context) => {
      if (context?.previousStories) {
        queryClient.setQueryData(queryKey, context.previousStories);
      }
      toast.error(`Failed to generate story: ${error.message}`);
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey });
    },
  });

  // Mutation: Update Story
  const updateMutation = useMutation({
    mutationFn: ({ id, updates }: { id: string; updates: Partial<Story> }) =>
      storyService.updateStory(id, updates),
    onMutate: async ({ id, updates }) => {
      await queryClient.cancelQueries({ queryKey });
      const previousStories = queryClient.getQueryData(queryKey);

      queryClient.setQueryData(queryKey, (old: any) =>
        (old || []).map((item: any) =>
          item.id === id ? { ...item, ...updates, _optimistic: true } : item,
        ),
      );

      return { previousStories };
    },
    onSuccess: (updatedStory, { id }) => {
      queryClient.setQueryData(queryKey, (old: any) =>
        (old || []).map((item: any) =>
          item.id === id ? { ...updatedStory, _optimistic: false } : item,
        ),
      );
      toast.success("Story updated successfully");
    },
    onError: (error: Error, variables, context) => {
      if (context?.previousStories) {
        queryClient.setQueryData(queryKey, context.previousStories);
      }
      toast.error(`Failed to update story: ${error.message}`);
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey });
    },
  });

  // Mutation: Delete Story
  const deleteMutation = useMutation({
    mutationFn: storyService.deleteStory,
    onMutate: async (id: string) => {
      await queryClient.cancelQueries({ queryKey });
      const previousStories = queryClient.getQueryData(queryKey);

      queryClient.setQueryData(queryKey, (old: any) =>
        (old || []).filter((item: any) => item.id !== id),
      );

      return { previousStories };
    },
    onSuccess: () => {
      toast.success("Story deleted successfully");
    },
    onError: (error: Error, variables, context) => {
      if (context?.previousStories) {
        queryClient.setQueryData(queryKey, context.previousStories);
      }
      toast.error(`Failed to delete story: ${error.message}`);
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey });
    },
  });

  return {
    stories,
    isLoading,
    error: error?.message || null,

    // Mutations
    createAIStory: async (prompt: string) => {
      try {
        const res = await createAIMutation.mutateAsync({ prompt });
        return { success: true, data: res };
      } catch (error: any) {
        return { success: false, error: error.message };
      }
    },
    createManualStory: async (title: string, description?: string) => {
      try {
        const res = await createManualMutation.mutateAsync({ title, description });
        return { success: true, data: res };
      } catch (error: any) {
        return { success: false, error: error.message };
      }
    },
    updateStory: async (id: string, updates: Partial<Story>) => {
      try {
        await updateMutation.mutateAsync({ id, updates });
        return { success: true };
      } catch (error: any) {
        return { success: false, error: error.message };
      }
    },
    deleteStory: async (id: string) => {
      try {
        await deleteMutation.mutateAsync(id);
        return { success: true };
      } catch (error: any) {
        return { success: false, error: error.message };
      }
    },

    // Loading states
    isCreating: createManualMutation.isPending,
    isUpdating: updateMutation.isPending,
    isDeleting: deleteMutation.isPending,
  };
}
