"use client";

import React, { useEffect, useState, useCallback } from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { Plus, Search, Filter, LayoutGrid, List, GitBranch } from "lucide-react";
import { usePipelines } from "@/hooks/use-pipelines";
import { pipelineApi } from "@/lib/api/pipelines";
import type { Pipeline, PipelineStats } from "@/lib/types/batch2";
import { PipelineCard } from "./pipeline-card";
import { PipelineStatsOverview } from "./pipeline-stats-overview";

interface PipelineListProps {
  workspaceId: string;
}

type StatusFilter = "all" | "active" | "failed" | "idle";
type ViewMode = "grid" | "table";

export function PipelineList({ workspaceId }: PipelineListProps) {
  const router = useRouter();
  const { pipelines, isLoading, error, fetchPipelines, deletePipeline } = usePipelines({
    workspaceId,
  });

  const [stats, setStats] = useState<PipelineStats | null>(null);
  const [statsLoading, setStatsLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState<StatusFilter>("all");
  const [viewMode, setViewMode] = useState<ViewMode>("grid");

  // Fetch data
  useEffect(() => {
    fetchPipelines();
  }, [fetchPipelines]);

  useEffect(() => {
    (async () => {
      try {
        setStatsLoading(true);
        const result = await pipelineApi.stats(workspaceId);
        setStats(result);
      } catch {
        // Stats are non-critical — fail silently
      } finally {
        setStatsLoading(false);
      }
    })();
  }, [workspaceId]);

  useEffect(() => {
    if (error) {
      toast.error(error);
    }
  }, [error]);

  // Handlers
  const handleEdit = useCallback(
    (pipeline: Pipeline) => {
      router.push(`/workspace/${workspaceId}/pipelines/${pipeline.id}`);
    },
    [router, workspaceId],
  );

  const handleDelete = useCallback(
    async (id: string) => {
      const confirmed = window.confirm("Delete this pipeline? This action cannot be undone.");
      if (!confirmed) return;
      const result = await deletePipeline(id);
      if (result.success) {
        toast.success("Pipeline deleted");
      } else {
        toast.error(result.error || "Failed to delete");
      }
    },
    [deletePipeline],
  );

  const handleRun = useCallback(() => {
    // Refresh after execution completes (card handles its own run)
    fetchPipelines();
  }, [fetchPipelines]);

  const handleViewHistory = useCallback(
    (pipeline: Pipeline) => {
      router.push(`/workspace/${workspaceId}/pipelines/${pipeline.id}?tab=history`);
    },
    [router, workspaceId],
  );

  // Filtering
  const filteredPipelines = pipelines.filter((p) => {
    if (searchQuery) {
      const q = searchQuery.toLowerCase();
      const nameMatch = p.name.toLowerCase().includes(q);
      const descMatch = (p.description || "").toLowerCase().includes(q);
      const typeMatch = p.sourceType.toLowerCase().includes(q);
      if (!nameMatch && !descMatch && !typeMatch) return false;
    }
    if (statusFilter === "active") return p.isActive && p.lastStatus !== "FAILED";
    if (statusFilter === "failed") return p.lastStatus === "FAILED";
    if (statusFilter === "idle") return !p.lastRunAt;
    return true;
  });

  const STATUS_BUTTONS: { value: StatusFilter; label: string }[] = [
    { value: "all", label: "All" },
    { value: "active", label: "Active" },
    { value: "failed", label: "Failed" },
    { value: "idle", label: "Never Run" },
  ];

  // Loading skeleton
  if (isLoading) {
    return (
      <div className="space-y-6 p-6">
        <PipelineStatsOverview stats={null} isLoading={true} />
        <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
          {Array.from({ length: 6 }).map((_, i) => (
            <div
              key={i}
              className="h-48 rounded-xl bg-zinc-900/50 border border-white/[0.06] animate-pulse"
            />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6 p-6">
      {/* Stats Overview */}
      <PipelineStatsOverview stats={stats} isLoading={statsLoading} />

      {/* Toolbar */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div className="flex items-center gap-3 flex-1 w-full sm:w-auto">
          {/* Search */}
          <div className="relative flex-1 sm:max-w-xs">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-zinc-500" />
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search pipelines..."
              className="w-full pl-9 pr-4 py-2 rounded-lg bg-black/30 border border-white/[0.06] text-sm text-white
                                placeholder:text-zinc-600 focus:border-white/[0.15] focus:outline-none focus:ring-1 focus:ring-white/[0.08]
                                transition-all duration-200"
            />
          </div>

          {/* Status Filter */}
          <div className="flex items-center rounded-lg bg-black/20 border border-white/[0.06] p-0.5">
            {STATUS_BUTTONS.map((btn) => (
              <button
                key={btn.value}
                onClick={() => setStatusFilter(btn.value)}
                className={`px-3 py-1.5 rounded-md text-[11px] font-medium transition-all duration-200
                                    ${
                                      statusFilter === btn.value
                                        ? "bg-white/[0.08] text-white shadow-sm"
                                        : "text-zinc-500 hover:text-zinc-300"
                                    }`}
              >
                {btn.label}
              </button>
            ))}
          </div>

          {/* View Mode Toggle */}
          <div className="hidden sm:flex items-center rounded-lg bg-black/20 border border-white/[0.06] p-0.5">
            <button
              onClick={() => setViewMode("grid")}
              className={`p-1.5 rounded-md transition-all duration-200
                                ${viewMode === "grid" ? "bg-white/[0.08] text-white" : "text-zinc-500 hover:text-zinc-300"}`}
            >
              <LayoutGrid className="w-4 h-4" />
            </button>
            <button
              onClick={() => setViewMode("table")}
              className={`p-1.5 rounded-md transition-all duration-200
                                ${viewMode === "table" ? "bg-white/[0.08] text-white" : "text-zinc-500 hover:text-zinc-300"}`}
            >
              <List className="w-4 h-4" />
            </button>
          </div>
        </div>

        {/* New Pipeline Button */}
        <button
          onClick={() => router.push(`/workspace/${workspaceId}/pipelines/new`)}
          className="flex items-center gap-2 px-4 py-2.5 rounded-lg 
                        bg-gradient-to-r from-violet-600 to-blue-600 
                        hover:from-violet-500 hover:to-blue-500
                        text-white text-xs font-semibold shadow-lg shadow-violet-600/20
                        active:scale-[0.97] transition-all duration-200"
        >
          <Plus className="w-4 h-4" />
          New Pipeline
        </button>
      </div>

      {/* Pipeline Grid / Empty State */}
      {filteredPipelines.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-20 px-6 rounded-xl border border-dashed border-white/[0.06]">
          <div className="w-14 h-14 rounded-2xl bg-zinc-900 border border-white/[0.06] flex items-center justify-center mb-4">
            <GitBranch className="w-6 h-6 text-zinc-600" />
          </div>
          <h3 className="text-sm font-semibold text-zinc-300 mb-1">
            {searchQuery || statusFilter !== "all" ? "No matching pipelines" : "No pipelines yet"}
          </h3>
          <p className="text-xs text-zinc-600 text-center max-w-sm mb-5">
            {searchQuery || statusFilter !== "all"
              ? "Try adjusting your search or filter criteria."
              : "Create your first data pipeline to start moving data between sources and destinations."}
          </p>
          {!searchQuery && statusFilter === "all" && (
            <button
              onClick={() => router.push(`/workspace/${workspaceId}/pipelines/new`)}
              className="flex items-center gap-2 px-4 py-2 rounded-lg bg-white/[0.06] border border-white/[0.08]
                                text-xs text-zinc-300 hover:bg-white/[0.1] transition-all duration-200"
            >
              <Plus className="w-3.5 h-3.5" />
              Create Pipeline
            </button>
          )}
        </div>
      ) : (
        <div
          className={
            viewMode === "grid"
              ? "grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4"
              : "grid grid-cols-1 gap-3"
          }
        >
          {filteredPipelines.map((pipeline) => (
            <PipelineCard
              key={pipeline.id}
              pipeline={pipeline}
              onEdit={handleEdit}
              onDelete={handleDelete}
              onRun={handleRun}
              onViewHistory={handleViewHistory}
            />
          ))}
        </div>
      )}

      {/* Count */}
      {filteredPipelines.length > 0 && (
        <div className="text-[11px] text-zinc-600 text-right">
          Showing {filteredPipelines.length} of {pipelines.length} pipeline
          {pipelines.length !== 1 ? "s" : ""}
        </div>
      )}
    </div>
  );
}
