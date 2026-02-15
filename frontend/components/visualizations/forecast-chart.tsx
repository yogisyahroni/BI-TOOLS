"use client"

import { useMemo } from "react"
import { Area, CartesianGrid, ComposedChart, Line, XAxis, YAxis } from "recharts"
import { format } from "date-fns"

import {
    ChartConfig,
    ChartContainer,
    ChartLegend,
    ChartLegendContent,
    ChartTooltip,
    ChartTooltipContent,
} from "@/components/ui/chart"

export interface DataPoint {
    timestamp: string // ISO string
    value: number
}

interface ForecastChartProps {
    history: DataPoint[]
    forecast: DataPoint[]
    className?: string
}

const chartConfig = {
    history: {
        label: "Historical Data",
        color: "hsl(var(--primary))",
    },
    forecast: {
        label: "Forecast",
        color: "hsl(var(--destructive))", // Accent color for forecast
    },
} satisfies ChartConfig

export function ForecastChart({ history, forecast, className }: ForecastChartProps) {
    const chartData = useMemo(() => {
        // Combine history and forecast into a unified structure
        // We want the lines to connect, so we might need the last history point to be the start of forecast
        // or just rely on visual proximity.

        const combined = []

        // Map history
        history.forEach((h) => {
            combined.push({
                timestamp: h.timestamp,
                formattedDate: format(new Date(h.timestamp), "MMM dd HH:mm"),
                history: h.value,
                forecast: null, // No forecast for this point
            })
        })

        // Bridge the gap: If we have history, adding the last history point as the start of forecast
        // creates a continuous line visually in most cases, or we can just let them be separate.
        // Let's add the last history point to the forecast line as a starting point if available.
        if (history.length > 0 && forecast.length > 0) {
            const lastHistory = history[history.length - 1]
            // Push a point that has BOTH history and forecast values (equal) to connect lines?
            // Or just a duplicate point with forecast value = history value.
            // Actually, easiest is to have the forecast line start exactly where history ends.
            // But the forecast returned by backend starts at t+1.
            // So visually there is a gap of 1 interval.
            // Recharts `connectNulls` might help if we interleave data, but here we have distinct series.
            // Let's just push forecast data.
        }

        // Map forecast
        forecast.forEach((f) => {
            combined.push({
                timestamp: f.timestamp,
                formattedDate: format(new Date(f.timestamp), "MMM dd HH:mm"),
                history: null,
                forecast: f.value,
            })
        })

        return combined
    }, [history, forecast])

    return (
        <ChartContainer config={chartConfig} className={className}>
            <ComposedChart data={chartData}>
                <CartesianGrid vertical={false} strokeDasharray="3 3" />
                <XAxis
                    dataKey="formattedDate"
                    tickLine={false}
                    axisLine={false}
                    tickMargin={8}
                    minTickGap={32}
                />
                <YAxis
                    tickLine={false}
                    axisLine={false}
                    tickMargin={8}
                    width={40}
                />
                <ChartTooltip
                    cursor={false}
                    content={<ChartTooltipContent indicator="dot" />}
                />
                <ChartLegend content={<ChartLegendContent />} />

                {/* Historical Line */}
                <Line
                    dataKey="history"
                    type="monotone"
                    stroke="var(--color-history)"
                    strokeWidth={2}
                    dot={false}
                    activeDot={{ r: 4 }}
                />

                {/* Forecast Line (Dashed) */}
                <Line
                    dataKey="forecast"
                    type="monotone"
                    stroke="var(--color-forecast)"
                    strokeWidth={2}
                    strokeDasharray="5 5"
                    dot={false}
                    activeDot={{ r: 4 }}
                    connectNulls // This helps if we bridge data logic is slightly off, but here data is disjoint
                />

                {/* Optional: Area under forecast to highlight it */}
                <Area
                    type="monotone"
                    dataKey="forecast"
                    fill="var(--color-forecast)"
                    fillOpacity={0.1}
                    stroke="none"
                />

            </ComposedChart>
        </ChartContainer>
    )
}
