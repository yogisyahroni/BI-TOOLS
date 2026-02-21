"use client";

import React, { useMemo, useState, useCallback } from "react";
import dynamic from "next/dynamic";
import { buildEChartsOptions } from "@/lib/visualizations/echarts-options";

const EChartsWrapper = dynamic(
  () => import("./visualizations/echarts-wrapper").then((mod) => mod.EChartsWrapper),
  {
    ssr: false,
    loading: () => (
      <div className="h-full w-full flex items-center justify-center bg-muted/20 animate-pulse rounded-lg">
        <span className="text-muted-foreground text-xs">Loading Chart Engine...</span>
      </div>
    ),
  },
);
import { useTheme } from "next-themes";
import { AlertCircle } from "lucide-react";
import { type VisualizationConfig } from "@/lib/types";
import { useWorkspaceTheme } from "@/components/theme/theme-provider";

// Import specialized components
import { MetricCard } from "./metric-card";
import { ProgressBar } from "./progress-bar";
import { GaugeChart } from "./gauge-chart";
import { SmallMultiples } from "./visualizations/small-multiples";

// Import annotation components
import { ChartAnnotations } from "./charts/chart-annotations";
import { AnnotationToolbar } from "./charts/annotation-toolbar";
import { CommentInput } from "./comments/comment-input";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";

import type { Comment, CreateAnnotationRequest, AnnotationPosition } from "@/types/comments";
import { type FilterCriteria } from "@/lib/cross-filter-context";

interface ChartVisualizationProps {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  data: Record<string, any>[];
  config: Partial<VisualizationConfig>;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  isLoading?: boolean;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  onDataClick?: (params: any) => void;

  // Annotation props (optional)
  chartId?: string;
  comments?: Comment[];
  currentUserId?: string;
  enableAnnotations?: boolean;
  onCreateAnnotation?: (data: CreateAnnotationRequest) => Promise<void>;
  onUpdateAnnotation?: (id: string, data: CreateAnnotationRequest) => Promise<void>;
  onDeleteAnnotation?: (id: string) => Promise<void>;
  activeFilters?: FilterCriteria[];
}

export function ChartVisualization({
  data,
  config,
  isLoading,
  onDataClick,

  // Annotation props
  chartId,
  comments = [],
  currentUserId,
  enableAnnotations = false,
  onCreateAnnotation,
  onUpdateAnnotation,
  onDeleteAnnotation,
  activeFilters = [],
}: ChartVisualizationProps) {
  const { theme } = useTheme();
  const { theme: workspaceTheme } = useWorkspaceTheme();

  // Annotation state
  const [isAnnotationMode, setIsAnnotationMode] = useState(false);
  const [annotationType, setAnnotationType] = useState<"point" | "range" | "text">("point");
  const [annotationColor, setAnnotationColor] = useState("#F59E0B");
  const [selectedAnnotationId, setSelectedAnnotationId] = useState<string | null>(null);
  const [pendingAnnotation, setPendingAnnotation] = useState<{
    position: AnnotationPosition;
    xValue?: number;
    yValue?: number;
  } | null>(null);
  const [showAnnotationDialog, setShowAnnotationDialog] = useState(false);
  const [editingComment, setEditingComment] = useState<Comment | null>(null);
  const [editContent, setEditContent] = useState("");

  // Calculate annotation count
  const annotationCount = useMemo(() => {
    if (!chartId) return 0;
    return comments.filter((c) => c.annotation && c.annotation.chartId === chartId).length;
  }, [comments, chartId]);

  // Strict Config with defaults
  const strictConfig = useMemo(() => {
    return {
      type: "bar",
      xAxis: "",
      yAxis: [],
      colors:
        config.colors && config.colors.length > 0
          ? config.colors
          : workspaceTheme?.chartPalette || [],
      ...config,
    } as VisualizationConfig;
  }, [config, workspaceTheme]);

  // Validate Config
  const isValid = useMemo(() => {
    // Metric only needs yAxis (value)
    if (strictConfig.type === "metric") {
      return strictConfig.yAxis && strictConfig.yAxis.length > 0;
    }
    // Gauge/Progress need yAxis (value)
    if (["gauge", "progress"].includes(strictConfig.type)) {
      return strictConfig.yAxis && strictConfig.yAxis.length > 0;
    }

    // Charts need X and Y
    if (!strictConfig.xAxis || !strictConfig.yAxis || strictConfig.yAxis.length === 0) return false;

    // Data check
    if (data.length > 0) {
      const keys = Object.keys(data[0]);
      if (strictConfig.xAxis && !keys.includes(strictConfig.xAxis)) return false;
    }
    return true;
  }, [strictConfig, data]);

  // Handle annotation click
  const handleAnnotationClick = useCallback(
    (position: AnnotationPosition, xValue?: number, yValue?: number) => {
      if (!isAnnotationMode) return;

      setPendingAnnotation({ position, xValue, yValue });
      setShowAnnotationDialog(true);
      setEditingComment(null);
      setEditContent("");
    },
    [isAnnotationMode],
  );

  // Handle annotation submit
  const handleAnnotationSubmit = useCallback(
    async (data: { content: string }) => {
      if (!pendingAnnotation || !chartId || !onCreateAnnotation) return;

      await onCreateAnnotation({
        chartId,
        content: data.content,
        position: pendingAnnotation.position,
        type: annotationType,
        color: annotationColor,
        xValue: pendingAnnotation.xValue,
        yValue: pendingAnnotation.yValue,
      });

      setShowAnnotationDialog(false);
      setPendingAnnotation(null);
    },
    [pendingAnnotation, chartId, annotationType, annotationColor, onCreateAnnotation],
  );

  // Handle edit annotation
  const handleEditAnnotation = useCallback((comment: Comment) => {
    setEditingComment(comment);
    setEditContent(comment.content);
    setShowAnnotationDialog(true);
  }, []);

  // Handle update annotation
  const handleUpdateAnnotation = useCallback(async () => {
    if (!editingComment || !editingComment.annotation || !onUpdateAnnotation) return;

    await onUpdateAnnotation(editingComment.annotation.id, {
      chartId: editingComment.annotation.chartId,
      content: editContent,
      position: editingComment.annotation.position,
      type: editingComment.annotation.type as "point" | "range" | "text",
      color: editingComment.annotation.color,
      xValue: editingComment.annotation.xValue || undefined,
      yValue: editingComment.annotation.yValue || undefined,
    });

    setShowAnnotationDialog(false);
    setEditingComment(null);
    setEditContent("");
  }, [editingComment, editContent, onUpdateAnnotation]);

  // Handle delete annotation
  const handleDeleteAnnotation = useCallback(
    async (annotationId: string) => {
      if (!onDeleteAnnotation) return;
      await onDeleteAnnotation(annotationId);
      setSelectedAnnotationId(null);
    },
    [onDeleteAnnotation],
  );

  // ---- Render Logic based on Type ----

  if (isLoading) {
    return (
      <div className="h-full w-full flex items-center justify-center bg-muted/20 animate-pulse rounded-lg">
        <span className="text-muted-foreground text-sm">Loading visualization...</span>
      </div>
    );
  }

  if (data.length === 0) {
    return (
      <div className="h-full w-full flex flex-col items-center justify-center border-2 border-dashed border-muted rounded-lg p-4">
        <AlertCircle className="h-8 w-8 text-muted-foreground mb-2" />
        <span className="text-muted-foreground font-medium">No data available to visualize</span>
      </div>
    );
  }

  if (!isValid) {
    return (
      <div className="h-full w-full flex flex-col items-center justify-center border-2 border-dashed border-destructive/30 bg-destructive/5 rounded-lg p-4">
        <AlertCircle className="h-8 w-8 text-destructive mb-2" />
        <span className="text-destructive font-medium">Invalid Configuration</span>
        <p className="text-xs text-muted-foreground mt-1 max-w-[250px] text-center">
          Please select valid axes from the configuration panel on the right.
        </p>
      </div>
    );
  }

  // 1. Metric Card Render
  if (strictConfig.type === "metric") {
    const valueCol = strictConfig.yAxis[0];
    const value = data[0]?.[valueCol]; // Take first row
    const previousValue = data.length > 1 ? data[1]?.[valueCol] : undefined;

    return (
      <div className="flex items-center justify-center h-full p-8">
        <div className="w-full max-w-sm">
          <MetricCard
            title={strictConfig.title || valueCol}
            value={typeof value === "number" ? value : String(value)}
            previousValue={
              strictConfig.showTrend && typeof previousValue === "number"
                ? previousValue
                : undefined
            }
            trendLabel={strictConfig.showTrend ? "vs previous" : undefined}
            size="lg"
          />
        </div>
      </div>
    );
  }

  // 2. Gauge Chart Render
  if (strictConfig.type === "gauge") {
    const valueCol = strictConfig.yAxis[0];
    const value = Number(data[0]?.[valueCol] || 0);

    return (
      <div className="flex items-center justify-center h-full p-8">
        <GaugeChart
          value={value}
          min={0}
          max={strictConfig.targetValue || 100}
          label={strictConfig.title || valueCol}
          size="lg"
        />
      </div>
    );
  }

  // 3. Progress Bar Render
  if (strictConfig.type === "progress") {
    const valueCol = strictConfig.yAxis[0];
    const value = Number(data[0]?.[valueCol] || 0);

    return (
      <div className="flex items-center justify-center h-full p-8">
        <div className="w-full max-w-md">
          <ProgressBar
            value={value}
            max={strictConfig.targetValue || 100}
            label={strictConfig.title || valueCol}
            size="lg"
            showPercentage
          />
        </div>
      </div>
    );
  }

  // 4. Default: ECharts (Bar, Line, Pie, Funnel, Combo, Scatter) with optional annotations
  const options = buildEChartsOptions(data, strictConfig, theme, activeFilters, chartId);

  const chartContent = (
    <div
      className={`h-full w-full min-h-[400px] border border-border rounded-lg bg-card p-4 shadow-sm overflow-hidden relative ${isAnnotationMode ? "ring-2 ring-primary ring-offset-2" : ""}`}
    >
      <EChartsWrapper
        options={options}
        isLoading={isLoading}
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        className="h-full w-full"
        onEvents={{
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
          click: (params: any) => {
            if (!isAnnotationMode && onDataClick) {
              onDataClick(params);
            }
          },
        }}
      />
    </div>
  );

  return (
    <div className="flex flex-col h-full">
      {/* Annotation Toolbar */}
      {enableAnnotations && chartId && currentUserId && (
        <div className="mb-3">
          <AnnotationToolbar
            isAnnotationMode={isAnnotationMode}
            onToggleMode={setIsAnnotationMode}
            selectedType={annotationType}
            onSelectType={setAnnotationType}
            selectedColor={annotationColor}
            onSelectColor={setAnnotationColor}
            annotationCount={annotationCount}
          />
        </div>
      )}

      {/* Chart with optional annotation layer */}
      {enableAnnotations && chartId && currentUserId ? (
        <ChartAnnotations
          chartId={chartId}
          comments={comments}
          currentUserId={currentUserId}
          isAnnotationMode={isAnnotationMode}
          selectedAnnotationType={annotationType}
          selectedColor={annotationColor}
          onAnnotationClick={handleAnnotationClick}
          onEditAnnotation={handleEditAnnotation}
          onDeleteAnnotation={handleDeleteAnnotation}
          selectedAnnotationId={selectedAnnotationId}
          onSelectAnnotation={setSelectedAnnotationId}
        >
          {chartContent}
        </ChartAnnotations>
      ) : (
        chartContent
      )}

      {/* Annotation Dialog */}
      <Dialog open={showAnnotationDialog} onOpenChange={setShowAnnotationDialog}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>{editingComment ? "Edit Annotation" : "Add Chart Annotation"}</DialogTitle>
          </DialogHeader>

          {editingComment ? (
            <div className="space-y-4">
              <textarea
                value={editContent}
                onChange={(e) => setEditContent(e.target.value)}
                className="w-full min-h-[100px] p-3 text-sm border rounded-md resize-none focus:outline-none focus:ring-2 focus:ring-primary"
                placeholder="Enter annotation text..."
                autoFocus
              />
              <div className="flex justify-end gap-2">
                <button
                  onClick={() => {
                    setShowAnnotationDialog(false);
                    setEditingComment(null);
                    setEditContent("");
                  }}
                  className="px-4 py-2 text-sm font-medium text-muted-foreground hover:text-foreground"
                >
                  Cancel
                </button>
                <button
                  onClick={handleUpdateAnnotation}
                  disabled={!editContent.trim()}
                  className="px-4 py-2 text-sm font-medium text-white bg-primary rounded-md hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  Update
                </button>
              </div>
            </div>
          ) : (
            <CommentInput
              onSubmit={async (data) => {
                await handleAnnotationSubmit({ content: data.content });
              }}
              entityType="chart"
              entityId={chartId || ""}
              placeholder="Add a comment about this data point..."
              currentUserId={currentUserId || ""}
              onCancel={() => {
                setShowAnnotationDialog(false);
                setPendingAnnotation(null);
              }}
              submitLabel="Add Annotation"
              autoFocus
            />
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}
