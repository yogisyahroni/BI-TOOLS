// Chart Templates Library - TASK-048
// Pre-configured chart settings untuk quick start

import type { EChartsOption } from "echarts";
import { getPaletteById } from "./chart-palettes";

/**
 * Chart Template Category
 */
export type ChartTemplateCategory =
  | "business"
  | "financial"
  | "analytics"
  | "marketing"
  | "operations"
  | "custom";

/**
 * Chart Template Definition
 */
export interface ChartTemplate {
  id: string;
  name: string;
  description: string;
  category: ChartTemplateCategory;
  chartType: string;
  thumbnail?: string;
  config: Partial<EChartsOption>;
  requiredFields?: string[];
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  exampleData?: any;
}

/**
 * Business Chart Templates
 */
export const BUSINESS_TEMPLATES: ChartTemplate[] = [
  {
    id: "sales-comparison",
    name: "Sales Comparison",
    description: "Compare sales across categories or time periods",
    category: "business",
    chartType: "bar",
    requiredFields: ["category", "value"],
    config: {
      title: {
        text: "Sales by Category",
        left: "center",
        textStyle: { fontSize: 18, fontWeight: 600 },
      },
      tooltip: {
        trigger: "axis",
        axisPointer: { type: "shadow" },
        formatter: "{b}: ${c}",
      },
      grid: { left: "10%", right: "10%", bottom: "15%", top: "20%" },
      xAxis: { type: "category", axisLabel: { rotate: 45 } },
      yAxis: { type: "value", name: "Sales ($)", axisLabel: { formatter: "${value}" } },
      series: [
        {
          type: "bar",
          itemStyle: { color: "#3b82f6" },
          label: { show: true, position: "top", formatter: "${c}" },
        },
      ],
    },
  },
  {
    id: "revenue-trend",
    name: "Revenue Trend",
    description: "Track revenue over time with trend line",
    category: "business",
    chartType: "line",
    requiredFields: ["date", "revenue"],
    config: {
      title: {
        text: "Monthly Revenue Trend",
        left: "center",
        textStyle: { fontSize: 18, fontWeight: 600 },
      },
      tooltip: {
        trigger: "axis",
        formatter: "{b}: ${c}",
      },
      grid: { left: "10%", right: "10%", bottom: "15%", top: "20%" },
      xAxis: { type: "category", boundaryGap: false },
      yAxis: { type: "value", name: "Revenue ($)", axisLabel: { formatter: "${value}K" } },
      series: [
        {
          type: "line",
          smooth: true,
          lineStyle: { width: 3 },
          itemStyle: { color: "#10b981" },
          areaStyle: { opacity: 0.3 },
        },
      ],
    },
  },
  {
    id: "market-share",
    name: "Market Share",
    description: "Visualize market share distribution",
    category: "business",
    chartType: "pie",
    requiredFields: ["category", "share"],
    config: {
      title: {
        text: "Market Share Distribution",
        left: "center",
        textStyle: { fontSize: 18, fontWeight: 600 },
      },
      tooltip: {
        trigger: "item",
        formatter: "{b}: {d}%",
      },
      legend: { bottom: "5%", left: "center" },
      series: [
        {
          type: "pie",
          radius: ["40%", "70%"],
          avoidLabelOverlap: true,
          label: {
            show: true,
            formatter: "{b}: {d}%",
          },
          emphasis: {
            label: { show: true, fontSize: 14, fontWeight: "bold" },
          },
        },
      ],
    },
  },
];

/**
 * Financial Chart Templates
 */
export const FINANCIAL_TEMPLATES: ChartTemplate[] = [
  {
    id: "profit-loss",
    name: "Profit & Loss Waterfall",
    description: "P&L breakdown with waterfall visualization",
    category: "financial",
    chartType: "waterfall",
    requiredFields: ["category", "value"],
    config: {
      title: {
        text: "Profit & Loss Statement",
        left: "center",
        textStyle: { fontSize: 18, fontWeight: 600 },
      },
      tooltip: {
        trigger: "axis",
        axisPointer: { type: "shadow" },
        formatter: "{b}: ${c}",
      },
      grid: { left: "15%", right: "10%", bottom: "20%", top: "20%" },
      legend: { top: "10%", data: ["Increase", "Decrease", "Total"] },
      xAxis: { type: "category", axisLabel: { rotate: 45 } },
      yAxis: { type: "value", name: "Amount ($)", axisLabel: { formatter: "${value}K" } },
    },
  },
  {
    id: "budget-actual",
    name: "Budget vs Actual",
    description: "Compare budgeted vs actual spending",
    category: "financial",
    chartType: "bar",
    requiredFields: ["category", "budget", "actual"],
    config: {
      title: {
        text: "Budget vs Actual Comparison",
        left: "center",
        textStyle: { fontSize: 18, fontWeight: 600 },
      },
      tooltip: {
        trigger: "axis",
        axisPointer: { type: "shadow" },
      },
      legend: { top: "10%", data: ["Budget", "Actual"] },
      grid: { left: "10%", right: "10%", bottom: "15%", top: "20%" },
      xAxis: { type: "category" },
      yAxis: { type: "value", name: "Amount ($)", axisLabel: { formatter: "${value}" } },
      series: [
        {
          name: "Budget",
          type: "bar",
          itemStyle: { color: "#94a3b8" },
          data: [],
        },
        {
          name: "Actual",
          type: "bar",
          itemStyle: { color: "#3b82f6" },
          data: [],
        },
      ],
    },
  },
  {
    id: "cash-flow",
    name: "Cash Flow",
    description: "Track cash flow over time",
    category: "financial",
    chartType: "line",
    requiredFields: ["date", "cashFlow"],
    config: {
      title: {
        text: "Cash Flow Trend",
        left: "center",
        textStyle: { fontSize: 18, fontWeight: 600 },
      },
      tooltip: {
        trigger: "axis",
        formatter: "{b}: ${c}",
      },
      grid: { left: "10%", right: "10%", bottom: "15%", top: "20%" },
      xAxis: { type: "category", boundaryGap: false },
      yAxis: { type: "value", name: "Cash Flow ($)", axisLabel: { formatter: "${value}K" } },
      visualMap: {
        show: false,
        dimension: 1,
        pieces: [
          { lte: 0, color: "#ef4444" },
          { gt: 0, color: "#10b981" },
        ],
      },
      series: [
        {
          type: "line",
          smooth: true,
          lineStyle: { width: 3 },
          areaStyle: { opacity: 0.3 },
        },
      ],
    },
  },
];

/**
 * Analytics Chart Templates
 */
export const ANALYTICS_TEMPLATES: ChartTemplate[] = [
  {
    id: "correlation-heatmap",
    name: "Correlation Matrix",
    description: "Show correlations between variables",
    category: "analytics",
    chartType: "heatmap",
    requiredFields: ["xAxis", "yAxis", "value"],
    config: {
      title: {
        text: "Correlation Matrix",
        left: "center",
        textStyle: { fontSize: 18, fontWeight: 600 },
      },
      tooltip: {
        position: "top",
        formatter: "{b}: {c}",
      },
      grid: { left: "15%", right: "15%", bottom: "15%", top: "20%" },
      xAxis: { type: "category", splitArea: { show: true } },
      yAxis: { type: "category", splitArea: { show: true } },
      visualMap: {
        min: -1,
        max: 1,
        calculable: true,
        orient: "horizontal",
        left: "center",
        bottom: "0%",
        inRange: {
          color: [
            "#313695",
            "#4575b4",
            "#74add1",
            "#abd9e9",
            "#e0f3f8",
            "#ffffbf",
            "#fee090",
            "#fdae61",
            "#f46d43",
            "#d73027",
            "#a50026",
          ],
        },
      },
    },
  },
  {
    id: "distribution-histogram",
    name: "Distribution Histogram",
    description: "Show value distribution",
    category: "analytics",
    chartType: "bar",
    requiredFields: ["range", "frequency"],
    config: {
      title: {
        text: "Distribution Analysis",
        left: "center",
        textStyle: { fontSize: 18, fontWeight: 600 },
      },
      tooltip: {
        trigger: "axis",
        axisPointer: { type: "shadow" },
      },
      grid: { left: "10%", right: "10%", bottom: "15%", top: "20%" },
      xAxis: { type: "category", name: "Range" },
      yAxis: { type: "value", name: "Frequency" },
      series: [
        {
          type: "bar",
          itemStyle: { color: "#8b5cf6" },
          barCategoryGap: "1%",
        },
      ],
    },
  },
  {
    id: "scatter-regression",
    name: "Scatter with Regression",
    description: "Scatter plot with trend line",
    category: "analytics",
    chartType: "scatter",
    requiredFields: ["x", "y"],
    config: {
      title: {
        text: "Scatter Plot Analysis",
        left: "center",
        textStyle: { fontSize: 18, fontWeight: 600 },
      },
      tooltip: {
        trigger: "item",
        formatter: "X: {c[0]}<br/>Y: {c[1]}",
      },
      grid: { left: "10%", right: "10%", bottom: "15%", top: "20%" },
      xAxis: { type: "value", name: "X Axis", nameLocation: "middle", nameGap: 30 },
      yAxis: { type: "value", name: "Y Axis", nameLocation: "middle", nameGap: 40 },
      series: [
        {
          type: "scatter",
          symbolSize: 10,
          itemStyle: { color: "#3b82f6", opacity: 0.7 },
        },
      ],
    },
  },
];

/**
 * Marketing Chart Templates
 */
export const MARKETING_TEMPLATES: ChartTemplate[] = [
  {
    id: "conversion-funnel",
    name: "Conversion Funnel",
    description: "Track conversion through stages",
    category: "marketing",
    chartType: "funnel",
    requiredFields: ["stage", "value"],
    config: {
      title: {
        text: "Conversion Funnel",
        left: "center",
        textStyle: { fontSize: 18, fontWeight: 600 },
      },
      tooltip: {
        trigger: "item",
        formatter: "{b}: {c} ({d}%)",
      },
      series: [
        {
          type: "funnel",
          left: "10%",
          right: "10%",
          top: "15%",
          bottom: "15%",
          sort: "descending",
          gap: 2,
          label: {
            show: true,
            position: "inside",
            formatter: "{b}\n{c} ({d}%)",
          },
        },
      ],
    },
  },
  {
    id: "channel-performance",
    name: "Channel Performance",
    description: "Compare marketing channel effectiveness",
    category: "marketing",
    chartType: "bar",
    requiredFields: ["channel", "conversions", "cost"],
    config: {
      title: {
        text: "Marketing Channel Performance",
        left: "center",
        textStyle: { fontSize: 18, fontWeight: 600 },
      },
      tooltip: {
        trigger: "axis",
        axisPointer: { type: "shadow" },
      },
      legend: { top: "10%", data: ["Conversions", "Cost"] },
      grid: { left: "10%", right: "10%", bottom: "15%", top: "25%" },
      xAxis: { type: "category" },
      yAxis: [
        { type: "value", name: "Conversions", position: "left" },
        {
          type: "value",
          name: "Cost ($)",
          position: "right",
          axisLabel: { formatter: "${value}" },
        },
      ],
      series: [
        {
          name: "Conversions",
          type: "bar",
          itemStyle: { color: "#10b981" },
          data: [],
        },
        {
          name: "Cost",
          type: "line",
          yAxisIndex: 1,
          itemStyle: { color: "#ef4444" },
          data: [],
        },
      ],
    },
  },
];

/**
 * Operations Chart Templates
 */
export const OPERATIONS_TEMPLATES: ChartTemplate[] = [
  {
    id: "gantt-timeline",
    name: "Project Timeline",
    description: "Track project tasks and milestones",
    category: "operations",
    chartType: "gantt",
    requiredFields: ["task", "start", "end"],
    config: {
      title: {
        text: "Project Timeline",
        left: "center",
        textStyle: { fontSize: 18, fontWeight: 600 },
      },
      tooltip: {
        trigger: "item",
        formatter: "{b}<br/>Start: {c[0]}<br/>End: {c[1]}",
      },
      grid: { left: "20%", right: "10%", bottom: "15%", top: "20%" },
    },
  },
  {
    id: "capacity-utilization",
    name: "Capacity Utilization",
    description: "Monitor resource capacity usage",
    category: "operations",
    chartType: "bar",
    requiredFields: ["resource", "used", "available"],
    config: {
      title: {
        text: "Resource Capacity Utilization",
        left: "center",
        textStyle: { fontSize: 18, fontWeight: 600 },
      },
      tooltip: {
        trigger: "axis",
        axisPointer: { type: "shadow" },
        formatter: "{b}<br/>Used: {c[0]}<br/>Available: {c[1]}",
      },
      legend: { top: "10%", data: ["Used", "Available"] },
      grid: { left: "15%", right: "10%", bottom: "15%", top: "25%" },
      xAxis: { type: "value", max: 100, axisLabel: { formatter: "{value}%" } },
      yAxis: { type: "category" },
      series: [
        {
          name: "Used",
          type: "bar",
          stack: "total",
          itemStyle: { color: "#f59e0b" },
          data: [],
        },
        {
          name: "Available",
          type: "bar",
          stack: "total",
          itemStyle: { color: "#e5e7eb" },
          data: [],
        },
      ],
    },
  },
];

/**
 * All Templates Combined
 */
export const ALL_CHART_TEMPLATES: ChartTemplate[] = [
  ...BUSINESS_TEMPLATES,
  ...FINANCIAL_TEMPLATES,
  ...ANALYTICS_TEMPLATES,
  ...MARKETING_TEMPLATES,
  ...OPERATIONS_TEMPLATES,
];

/**
 * Get template by ID
 */
export function getTemplateById(id: string): ChartTemplate | undefined {
  return ALL_CHART_TEMPLATES.find((t) => t.id === id);
}

/**
 * Get templates by category
 */
export function getTemplatesByCategory(category: ChartTemplateCategory): ChartTemplate[] {
  return ALL_CHART_TEMPLATES.filter((t) => t.category === category);
}

/**
 * Get templates by chart type
 */
export function getTemplatesByChartType(chartType: string): ChartTemplate[] {
  return ALL_CHART_TEMPLATES.filter((t) => t.chartType === chartType);
}

/**
 * Apply template to data
 */
export function applyTemplate(
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  template: ChartTemplate,
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  data: any[],
  fieldMapping?: Record<string, string>,
): Partial<EChartsOption> {
  const config = JSON.parse(JSON.stringify(template.config)); // Deep clone

  // Apply data transformations based on chart type
  if (template.chartType === "bar" && config.series && Array.isArray(config.series)) {
    if (fieldMapping?.category && fieldMapping?.value) {
      config.xAxis = { ...config.xAxis, data: data.map((d) => d[fieldMapping.category]) };
      config.series[0].data = data.map((d) => d[fieldMapping.value]);
    }
  }

  if (template.chartType === "line" && config.series && Array.isArray(config.series)) {
    if (fieldMapping?.x && fieldMapping?.y) {
      config.xAxis = { ...config.xAxis, data: data.map((d) => d[fieldMapping.x]) };
      config.series[0].data = data.map((d) => d[fieldMapping.y]);
    }
  }

  if (template.chartType === "pie" && config.series && Array.isArray(config.series)) {
    if (fieldMapping?.name && fieldMapping?.value) {
      config.series[0].data = data.map((d) => ({
        name: d[fieldMapping.name],
        value: d[fieldMapping.value],
      }));
    }
  }

  return config;
}

/**
 * Save custom template
 */
export interface CustomTemplate {
  id: string;
  name: string;
  description: string;
  category: ChartTemplateCategory;
  config: Partial<EChartsOption>;
  createdAt: string;
  updatedAt: string;
}

/**
 * Save template to localStorage
 */
export function saveCustomTemplate(
  template: Omit<CustomTemplate, "id" | "createdAt" | "updatedAt">,
): CustomTemplate {
  const customTemplate: CustomTemplate = {
    ...template,
    id: `custom-${Date.now()}`,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  };

  const existing = loadCustomTemplates();
  const updated = [...existing, customTemplate];

  if (typeof window !== "undefined") {
    localStorage.setItem("custom-chart-templates", JSON.stringify(updated));
  }

  return customTemplate;
}

/**
 * Load custom templates from localStorage
 */
export function loadCustomTemplates(): CustomTemplate[] {
  if (typeof window === "undefined") return [];

  const stored = localStorage.getItem("custom-chart-templates");
  return stored ? JSON.parse(stored) : [];
}

/**
 * Delete custom template
 */
export function deleteCustomTemplate(id: string): void {
  const existing = loadCustomTemplates();
  const updated = existing.filter((t) => t.id !== id);

  if (typeof window !== "undefined") {
    localStorage.setItem("custom-chart-templates", JSON.stringify(updated));
  }
}

/**
 * Update custom template
 */
export function updateCustomTemplate(
  id: string,
  updates: Partial<CustomTemplate>,
): CustomTemplate | null {
  const existing = loadCustomTemplates();
  const index = existing.findIndex((t) => t.id === id);

  if (index === -1) return null;

  const updated = {
    ...existing[index],
    ...updates,
    updatedAt: new Date().toISOString(),
  };

  existing[index] = updated;

  if (typeof window !== "undefined") {
    localStorage.setItem("custom-chart-templates", JSON.stringify(existing));
  }

  return updated;
}
