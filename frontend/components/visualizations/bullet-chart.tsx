"use client";

import React from "react";
import { EChartsWrapper } from "./echarts-wrapper";
import type { EChartsOption } from "echarts";

/**
 * TASK-CHART-003: Bullet Chart
 * Stephen Few-style bullet chart for performance vs target comparison.
 */

interface BulletItem {
  label: string;
  actual: number;
  target: number;
  ranges: [number, number, number]; // [poor, acceptable, good] thresholds
  max?: number;
}

interface BulletChartProps {
  data: BulletItem[];
  title?: string;
  orientation?: "horizontal" | "vertical";
  rangeColors?: [string, string, string];
  actualColor?: string;
  targetColor?: string;
  className?: string;
  height?: number | string;
  onBarClick?: (item: BulletItem, index: number) => void;
}

export function BulletChart({
  data,
  title,
  orientation = "horizontal",
  rangeColors = ["#e5e7eb", "#d1d5db", "#9ca3af"],
  actualColor = "#3b82f6",
  targetColor = "#111827",
  className = "h-full w-full min-h-[300px]",
  height,
  onBarClick,
}: BulletChartProps) {
  const isHorizontal = orientation === "horizontal";
  const categoryData = data.map((d) => d.label);

  const _maxValues = data.map((d) => d.max ?? Math.max(d.ranges[2], d.actual, d.target) * 1.1);

  // Build 3 stacked bars for qualitative ranges (poor, acceptable, good)
  const rangeSeries = [0, 1, 2].map((rangeIdx) => ({
    name: ["Poor", "Acceptable", "Good"][rangeIdx],
    type: "bar" as const,
    stack: "ranges",
    barWidth: "60%",
    itemStyle: {
      color: rangeColors[rangeIdx],
      borderRadius: rangeIdx === 2 ? (isHorizontal ? [0, 4, 4, 0] : [4, 4, 0, 0]) : 0,
    },
    data: data.map((d) => {
      const prev = rangeIdx === 0 ? 0 : d.ranges[rangeIdx - 1];
      return d.ranges[rangeIdx] - prev;
    }),
    silent: true,
    z: 1,
  }));

  // Actual value bar (narrower, overlaid)
  const actualSeries = {
    name: "Actual",
    type: "bar" as const,
    barWidth: "25%",
    barGap: "-100%",
    itemStyle: {
      color: actualColor,
      borderRadius: isHorizontal ? [0, 3, 3, 0] : [3, 3, 0, 0],
    },
    data: data.map((d) => d.actual),
    z: 2,
  };

  // Target marker line (scatter with custom symbol)
  const targetSeries = {
    name: "Target",
    type: "scatter" as const,
    symbol: isHorizontal ? "rect" : "rect",
    symbolSize: isHorizontal ? [3, 20] : [20, 3],
    itemStyle: { color: targetColor },
    data: data.map((d) => d.target),
    z: 3,
  };

  const option: EChartsOption = {
    title: title
      ? { text: title, left: "center", textStyle: { fontSize: 14, fontWeight: 600 } }
      : undefined,
    tooltip: {
      trigger: "axis",
      axisPointer: { type: "shadow" },
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      formatter: (params: any) => {
        if (!Array.isArray(params) || params.length === 0) return "";
        const idx = params[0].dataIndex;
        const item = data[idx];
        return `<strong>${item.label}</strong><br/>
                    Actual: <strong>${item.actual.toLocaleString()}</strong><br/>
                    Target: ${item.target.toLocaleString()}<br/>
                    Poor: 0–${item.ranges[0]}<br/>
                    Acceptable: ${item.ranges[0]}–${item.ranges[1]}<br/>
                    Good: ${item.ranges[1]}–${item.ranges[2]}`;
      },
    },
    legend: {
      bottom: 0,
      itemWidth: 12,
      itemHeight: 12,
      textStyle: { fontSize: 11 },
      data: ["Poor", "Acceptable", "Good", "Actual", "Target"],
    },
    grid: {
      left: isHorizontal ? "15%" : "10%",
      right: "8%",
      top: title ? "15%" : "8%",
      bottom: "18%",
    },
    [isHorizontal ? "yAxis" : "xAxis"]: {
      type: "category",
      data: categoryData,
      axisLine: { show: false },
      axisTick: { show: false },
      axisLabel: { fontSize: 11, fontWeight: 500 },
    },
    [isHorizontal ? "xAxis" : "yAxis"]: {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      type: "value",
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      max: (val: any) => Math.ceil(val.max * 1.05),
      axisLabel: { fontSize: 10 },
      splitLine: { lineStyle: { opacity: 0.15 } },
    },
    series: [...rangeSeries, actualSeries, targetSeries],
    animationDuration: 600,
    animationEasing: "cubicOut",
  };

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const events = onBarClick
    ? {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        click: (params: any) => {
          if (params.seriesName === "Actual") {
            onBarClick(data[params.dataIndex], params.dataIndex);
          }
        },
      }
    : undefined;

  return <EChartsWrapper options={option} className={className} onEvents={events} />;
}
