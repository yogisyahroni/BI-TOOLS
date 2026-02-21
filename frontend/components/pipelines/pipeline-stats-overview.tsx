"use client";

import React from "react";
import type { PipelineStats } from "@/lib/types/batch2";
import { GitBranch, Play, TrendingUp, AlertTriangle, Database, CheckCircle2 } from "lucide-react";

interface PipelineStatsOverviewProps {
  stats: PipelineStats | null;
  isLoading: boolean;
}

interface StatCardProps {
  icon: React.ReactNode;
  label: string;
  value: string | number;
  subValue?: string;
  gradient: string;
  iconBg: string;
}

function StatCard({ icon, label, value, subValue, gradient, iconBg }: StatCardProps) {
  return (
    <div
      className={`relative rounded-xl border border-white/[0.06] overflow-hidden
            bg-gradient-to-br ${gradient} backdrop-blur-xl
            transition-all duration-300 hover:border-white/[0.12] hover:translate-y-[-1px]`}
    >
      <div className="p-5">
        <div className="flex items-start justify-between">
          <div>
            <p className="text-[11px] font-medium text-zinc-400 uppercase tracking-wider mb-1.5">
              {label}
            </p>
            <p className="text-2xl font-bold text-white tracking-tight">{value}</p>
            {subValue && <p className="text-[11px] text-zinc-500 mt-1">{subValue}</p>}
          </div>
          <div
            className={`flex-shrink-0 w-10 h-10 rounded-lg ${iconBg} flex items-center justify-center`}
          >
            {icon}
          </div>
        </div>
      </div>
    </div>
  );
}

function SkeletonCard() {
  return (
    <div className="rounded-xl border border-white/[0.06] bg-zinc-900/50 p-5 animate-pulse">
      <div className="flex items-start justify-between">
        <div className="space-y-2">
          <div className="h-3 w-20 bg-zinc-800 rounded" />
          <div className="h-7 w-16 bg-zinc-800 rounded" />
          <div className="h-3 w-24 bg-zinc-800 rounded" />
        </div>
        <div className="w-10 h-10 bg-zinc-800 rounded-lg" />
      </div>
    </div>
  );
}

export function PipelineStatsOverview({ stats, isLoading }: PipelineStatsOverviewProps) {
  if (isLoading) {
    return (
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <SkeletonCard />
        <SkeletonCard />
        <SkeletonCard />
        <SkeletonCard />
      </div>
    );
  }

  if (!stats) return null;

  const successRateFormatted = `${stats.successRate?.toFixed(1) || 0}%`;
  const failedCount = stats.recentFailures?.length || 0;

  return (
    <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <StatCard
        icon={<GitBranch className="w-5 h-5 text-violet-400" />}
        label="Total Pipelines"
        value={stats.totalPipelines}
        subValue={`${stats.activePipelines} active`}
        gradient="from-violet-950/30 to-zinc-950"
        iconBg="bg-violet-500/10"
      />
      <StatCard
        icon={<Play className="w-5 h-5 text-blue-400" />}
        label="Total Runs"
        value={stats.totalExecutions}
        subValue={`${stats.totalRowsProcessed?.toLocaleString() || 0} rows processed`}
        gradient="from-blue-950/30 to-zinc-950"
        iconBg="bg-blue-500/10"
      />
      <StatCard
        icon={<CheckCircle2 className="w-5 h-5 text-emerald-400" />}
        label="Success Rate"
        value={successRateFormatted}
        subValue="Last 30 days"
        gradient="from-emerald-950/30 to-zinc-950"
        iconBg="bg-emerald-500/10"
      />
      <StatCard
        icon={<AlertTriangle className="w-5 h-5 text-amber-400" />}
        label="Recent Failures"
        value={failedCount}
        subValue={failedCount > 0 ? "Requires attention" : "All clear"}
        gradient="from-amber-950/30 to-zinc-950"
        iconBg="bg-amber-500/10"
      />
    </div>
  );
}
