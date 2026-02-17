/**
 * Visualization Components Barrel Export
 * TASK-CHART-001 through TASK-CHART-024
 *
 * All chart components and utilities for the Insight Engine chart library.
 */

// ─── Wrapper ─────────────────────────────────────────────────────────────────
export { EChartsWrapper } from './echarts-wrapper';

// ─── Chart Components ────────────────────────────────────────────────────────
// CHART-002: KPI Card / Big Number
export { KPICard } from './kpi-card';

// CHART-003: Bullet Chart
export { BulletChart } from './bullet-chart';

// CHART-004: Histogram
export { HistogramChart } from './histogram-chart';

// CHART-005: Box Plot
export { BoxPlotChart } from './boxplot-chart';

// CHART-006: Radar / Spider
export { RadarChart } from './radar-chart';

// CHART-007: Donut Chart
export { DonutChart } from './donut-chart';

// CHART-008: Sunburst Chart
export { SunburstChart } from './sunburst-chart';

// CHART-009: Mekko / Marimekko
export { MekkoChart } from './mekko-chart';

// CHART-010: Ribbon Chart
export { RibbonChart } from './ribbon-chart';

// CHART-011: Stream Graph
export { StreamGraph } from './stream-graph';

// CHART-012: Parallel Coordinates
export { ParallelCoordinates } from './parallel-coordinates';

// CHART-013: Polar Area / Nightingale Rose
export { PolarAreaChart } from './polar-area-chart';

// CHART-014: Treemap Enhanced
export { TreemapEnhanced } from './treemap-enhanced';

// CHART-015: Word Cloud
export { WordCloudChart } from './word-cloud-chart';

// CHART-016: Chord Diagram
export { ChordDiagram } from './chord-diagram';

// CHART-017: Calendar Heatmap
export { CalendarHeatmap } from './calendar-heatmap';

// CHART-018: Network Graph
export { NetworkGraph } from './network-graph';

// CHART-019: Diverging Bar Chart
export { DivergingBarChart } from './diverging-bar-chart';

// CHART-020: Nested Pie / Double Doughnut
export { NestedPieChart } from './nested-pie-chart';

// ─── Existing Charts ─────────────────────────────────────────────────────────
export { AnomalyChart } from './anomaly-chart';
export { ChoroplethMap } from './choropleth-map';
export { ForecastChart } from './forecast-chart';
export { FunnelChart } from './funnel-chart';
export { GanttChart } from './gantt-chart';
export { HeatmapChart } from './heatmap-chart';
export { SankeyChart } from './sankey-chart';
export { SmallMultiples } from './small-multiples';
export { TreemapChart } from './treemap-chart';
export { WaterfallChart } from './waterfall-chart';

// ─── Utilities & Types ──────────────────────────────────────────────────────
// CHART-022: Advanced Tooltips
export {
    TooltipProvider,
    useAdvancedTooltip,
    buildEChartsTooltip,
    tooltipFormatters,
} from './advanced-tooltips';

// CHART-023: Animated Transitions
export {
    animationPresets,
    getAnimationConfig,
    applyAnimation,
    staggerDelay,
    progressiveAnimation,
} from './animated-transitions';

// CHART-024: Accessibility
export {
    accessibleColorPalettes,
    chartPatterns,
    createPatternFill,
    chartAriaAttributes,
    generateChartSummary,
    chartKeyboardHandler,
    createLiveRegion,
} from './chart-accessibility';

// Chart Type Definitions
export type {
    SankeyNode,
    SankeyLink,
    SankeyData,
    SankeyChartProps,
    GanttTask,
    GanttChartProps,
    HeatmapDataPoint,
    HeatmapChartProps,
    TreemapNode,
    TreemapChartProps,
    WaterfallDataPoint,
    WaterfallChartProps,
    FunnelDataPoint,
    FunnelChartProps,
    CommonChartConfig,
} from './advanced-chart-types';

export { CHART_COLOR_PALETTES } from './advanced-chart-types';
