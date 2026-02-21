"use client";

import React, { useMemo } from "react";
import { EChartsWrapper } from "./echarts-wrapper";
import type { EChartsOption } from "echarts";

/**
 * TASK-CHART-009: Mekko / Marimekko Chart
 * Variable-width stacked bars where bar width encodes data magnitude.
 */

interface MekkoCategory {
  name: string;
  total: number;
  segments: { name: string; value: number; color?: string }[];
}

interface MekkoChartProps {
  data: MekkoCategory[];
  title?: string;
  showPercentages?: boolean;
  className?: string;
  onSegmentClick?: (category: string, segment: string, value: number) => void;
}

const MEKKO_COLORS = [
  "#3b82f6",
  "#10b981",
  "#f59e0b",
  "#ef4444",
  "#8b5cf6",
  "#ec4899",
  "#06b6d4",
  "#f97316",
];

export function MekkoChart({
  data,
  title,
  showPercentages = true,
  className = "h-full w-full min-h-[400px]",
  onSegmentClick,
}: MekkoChartProps) {
  const { seriesData, segmentNames } = useMemo(() => {
    const grandTotal = data.reduce((s, c) => s + c.total, 0);
    const segNameSet = new Set<string>();
    data.forEach((c) => c.segments.forEach((s) => segNameSet.add(s.name)));
    const segNames = Array.from(segNameSet);

    // Convert to echarts custom series
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const allRects: any[] = [];
    let xOffset = 0;

    data.forEach((cat) => {
      const catWidth = (cat.total / grandTotal) * 100;
      let yOffset = 0;

      cat.segments.forEach((seg, _si) => {
        const segHeight = (seg.value / cat.total) * 100;
        const segIdx = segNames.indexOf(seg.name);
        const color = seg.color ?? MEKKO_COLORS[segIdx % MEKKO_COLORS.length];

        allRects.push({
          type: "rect",
          shape: {
            x: 0,
            y: 0,
            width: 0,
            height: 0,
          },
          x: xOffset,
          y: yOffset,
          w: catWidth,
          h: segHeight,
          catName: cat.name,
          segName: seg.name,
          segValue: seg.value,
          color,
          pct: ((seg.value / cat.total) * 100).toFixed(1),
          catPct: ((cat.total / grandTotal) * 100).toFixed(1),
        });

        yOffset += segHeight;
      });

      xOffset += catWidth;
    });

    return { seriesData: allRects, segmentNames: segNames };
  }, [data]);

  // Use a pure SVG approach for Mekko since ECharts doesn't natively support variable-width bars
  const grandTotal = data.reduce((s, c) => s + c.total, 0);

  return (
    <div className={className}>
      {title && <h3 className="text-sm font-semibold text-center mb-3">{title}</h3>}
      <div className="relative w-full h-[calc(100%-60px)]">
        <svg viewBox="0 0 100 100" preserveAspectRatio="none" className="w-full h-full">
          {(() => {
            const rects: React.ReactElement[] = [];
            let xOff = 0;

            data.forEach((cat, ci) => {
              const catWidth = (cat.total / grandTotal) * 100;
              let yOff = 0;

              cat.segments.forEach((seg, si) => {
                const segHeight = (seg.value / cat.total) * 100;
                const segIdx = segmentNames.indexOf(seg.name);
                const color = seg.color ?? MEKKO_COLORS[segIdx % MEKKO_COLORS.length];

                rects.push(
                  <rect
                    key={`${ci}-${si}`}
                    x={xOff}
                    y={100 - yOff - segHeight}
                    width={catWidth}
                    height={segHeight}
                    fill={color}
                    stroke="white"
                    strokeWidth={0.3}
                    className="transition-opacity duration-200 hover:opacity-80 cursor-pointer"
                    onClick={() => onSegmentClick?.(cat.name, seg.name, seg.value)}
                  >
                    <title>
                      {`${cat.name} > ${seg.name}\nValue: ${seg.value.toLocaleString()}\n${((seg.value / cat.total) * 100).toFixed(1)}% of category\nCategory: ${((cat.total / grandTotal) * 100).toFixed(1)}% of total`}
                    </title>
                  </rect>,
                );

                // Label in center of rect if large enough
                if (showPercentages && catWidth > 8 && segHeight > 8) {
                  rects.push(
                    <text
                      key={`label-${ci}-${si}`}
                      x={xOff + catWidth / 2}
                      y={100 - yOff - segHeight / 2}
                      textAnchor="middle"
                      dominantBaseline="middle"
                      fontSize={Math.min(catWidth * 0.2, 3.5)}
                      fill="white"
                      fontWeight="600"
                      className="pointer-events-none select-none"
                    >
                      {((seg.value / cat.total) * 100).toFixed(0)}%
                    </text>,
                  );
                }

                yOff += segHeight;
              });

              xOff += catWidth;
            });

            return rects;
          })()}
        </svg>
      </div>

      {/* Category labels */}
      <div className="flex mt-1" style={{ height: 20 }}>
        {data.map((cat, i) => {
          const pct = (cat.total / grandTotal) * 100;
          return (
            <div key={i} className="text-center overflow-hidden" style={{ width: `${pct}%` }}>
              {pct > 5 && (
                <span className="text-[10px] text-muted-foreground truncate block">{cat.name}</span>
              )}
            </div>
          );
        })}
      </div>

      {/* Legend */}
      <div className="flex flex-wrap gap-3 justify-center mt-2">
        {segmentNames.map((name, i) => (
          <div key={name} className="flex items-center gap-1.5">
            <div
              className="w-3 h-3 rounded-sm flex-shrink-0"
              style={{ backgroundColor: MEKKO_COLORS[i % MEKKO_COLORS.length] }}
            />
            <span className="text-xs text-muted-foreground">{name}</span>
          </div>
        ))}
      </div>
    </div>
  );
}
