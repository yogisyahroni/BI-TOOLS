"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { type BusinessMetric } from "@/lib/types";
import { fetchWithAuth } from "@/lib/utils";

interface UseBusinessMetricsOptions {
  status?: string;
  autoFetch?: boolean;
}

export function useBusinessMetrics(options: UseBusinessMetricsOptions = {}) {
  const queryClient = useQueryClient();

  const {
    data: metrics = [],
    isLoading,
    error: queryError,
    refetch,
  } = useQuery({
    queryKey: ["metrics", options.status],
    queryFn: async () => {
      const params = new URLSearchParams();
      if (options.status) params.append("status", options.status);

      const response = await fetchWithAuth(`/api/go/metrics?${params.toString()}`);

      if (!response.ok) {
        throw new Error(`Failed to fetch metrics: ${response.status}`);
      }

      const data = (await response.json()) as { success: boolean; data: BusinessMetric[] };

      if (!data.success) {
        throw new Error("Failed to fetch metrics");
      }
      return data.data;
    },
    enabled: options.autoFetch !== false,
  });

  const createMutation = useMutation({
    mutationFn: async (metric: Omit<BusinessMetric, "id" | "createdAt" | "updatedAt">) => {
      const response = await fetchWithAuth("/api/go/metrics", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(metric),
      });

      if (!response.ok) throw new Error("Failed to save metric");

      const data = await response.json();
      if (!data.success) throw new Error(data.error || "Failed to save metric");
      return data.data as BusinessMetric;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["metrics"] });
    },
  });

  const saveMetric = async (metric: Omit<BusinessMetric, "id" | "createdAt" | "updatedAt">) => {
    try {
      const data = await createMutation.mutateAsync(metric);
      return { success: true, data };
    } catch (err) {
      return { success: false, error: err instanceof Error ? err.message : "Unknown error" };
    }
  };

  return {
    metrics,
    isLoading,
    error: queryError ? queryError.message : null,
    fetchMetrics: refetch,
    saveMetric,
  };
}
