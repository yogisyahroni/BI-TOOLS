"use client";

import { useSession } from "next-auth/react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { fetchWithAuth } from "@/lib/utils";
import { type SavedQuery } from "@/lib/types";

interface UseSavedQueriesOptions {
  collectionId?: string;
  autoFetch?: boolean;
}

export function useSavedQueries(options: UseSavedQueriesOptions = {}) {
  const { status } = useSession();
  const isAuthenticated = status === "authenticated";
  const queryClient = useQueryClient();

  const queryKey = ["savedQueries", options.collectionId];

  const {
    data: queries = [],
    isLoading,
    error: queryError,
    refetch: fetchQueries,
  } = useQuery({
    queryKey,
    queryFn: async () => {
      const params = new URLSearchParams();
      if (options.collectionId) params.append("collectionId", options.collectionId);

      const response = await fetchWithAuth(`/api/go/queries?${params.toString()}`);
      if (!response.ok) {
        throw new Error(`Failed to fetch queries: ${response.status}`);
      }

      const data = (await response.json()) as { success: boolean; data: SavedQuery[] };
      if (!data.success) {
        throw new Error("Failed to fetch queries");
      }
      return data.data;
    },
    enabled: isAuthenticated && options.autoFetch !== false,
  });

  const saveMutation = useMutation({
    mutationFn: async (
      query: Omit<SavedQuery, "id" | "createdAt" | "updatedAt" | "views" | "pinned">,
    ) => {
      const response = await fetchWithAuth("/api/go/queries", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(query),
      });

      if (!response.ok) throw new Error(`Failed to save query: ${response.status}`);

      const data = (await response.json()) as { success: boolean; data: SavedQuery };
      if (!data.success) throw new Error("Failed to save query");
      return data.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["savedQueries"] });
    },
  });

  const deleteMutation = useMutation({
    mutationFn: async (queryId: string) => {
      const response = await fetchWithAuth(`/api/go/queries/${queryId}`, { method: "DELETE" });
      if (!response.ok) throw new Error(`Failed to delete query: ${response.status}`);
      return queryId;
    },
    onSuccess: (_, queryId) => {
      queryClient.setQueryData<SavedQuery[]>(queryKey, (old) =>
        old ? old.filter((q) => q.id !== queryId) : [],
      );
    },
  });

  const pinMutation = useMutation({
    mutationFn: async ({ queryId, pinned }: { queryId: string; pinned: boolean }) => {
      const response = await fetchWithAuth(`/api/go/queries/${queryId}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ pinned }),
      });
      if (!response.ok) throw new Error("Failed to update pin status");
      return { queryId, pinned };
    },
    onMutate: async ({ queryId, pinned }) => {
      await queryClient.cancelQueries({ queryKey });
      const previousQueries = queryClient.getQueryData<SavedQuery[]>(queryKey);
      if (previousQueries) {
        queryClient.setQueryData<SavedQuery[]>(queryKey, (old) =>
          old
            ? old.map((q) => (q.id === queryId ? { ...q, pinned, updatedAt: new Date() } : q))
            : [],
        );
      }
      return { previousQueries };
    },
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    onError: (err, newTodo, context) => {
      if (context?.previousQueries) {
        queryClient.setQueryData(queryKey, context.previousQueries);
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey });
    },
  });

  const updateMutation = useMutation({
    mutationFn: async ({ id, updates }: { id: string; updates: Partial<SavedQuery> }) => {
      const response = await fetchWithAuth(`/api/go/queries/${id}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(updates),
      });

      if (!response.ok) throw new Error(`Failed to update query: ${response.status}`);

      const data = (await response.json()) as { success: boolean; data: SavedQuery };
      if (!data.success) throw new Error("Failed to update query");
      return data.data;
    },
    onSuccess: (updatedQuery) => {
      queryClient.setQueryData<SavedQuery[]>(queryKey, (old) =>
        old ? old.map((q) => (q.id === updatedQuery.id ? updatedQuery : q)) : [],
      );
    },
  });

  // Preserve the original return signatures / interface
  const saveQuery = async (
    query: Omit<SavedQuery, "id" | "createdAt" | "updatedAt" | "views" | "pinned">,
  ) => {
    try {
      const data = await saveMutation.mutateAsync(query);
      return { success: true, data };
    } catch (err) {
      return { success: false, error: err instanceof Error ? err.message : "Unknown error" };
    }
  };

  const deleteQuery = async (queryId: string) => {
    try {
      await deleteMutation.mutateAsync(queryId);
      return { success: true };
    } catch (err) {
      return { success: false, error: err instanceof Error ? err.message : "Unknown error" };
    }
  };

  const pinQuery = async (queryId: string) => {
    try {
      const query = queries.find((q) => q.id === queryId);
      if (!query) return;
      await pinMutation.mutateAsync({ queryId, pinned: !query.pinned });
    } catch (err) {
      // optimistic update handled internally
    }
  };

  const updateQuery = async (id: string, updates: Partial<SavedQuery>) => {
    try {
      const data = await updateMutation.mutateAsync({ id, updates });
      return { success: true, data };
    } catch (err) {
      return { success: false, error: err instanceof Error ? err.message : "Unknown error" };
    }
  };

  return {
    queries,
    isLoading,
    error: queryError ? queryError.message : null,
    fetchQueries,
    saveQuery,
    deleteQuery,
    pinQuery,
    updateQuery,
  };
}
