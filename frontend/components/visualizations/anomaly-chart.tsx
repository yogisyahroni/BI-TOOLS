"use client"

import { useMemo } from "react"
import { CartesianGrid, ComposedChart, Line, Scatter, XAxis, YAxis, Tooltip, Legend } from "recharts"
import { format } from "date-fns"

import {
    ChartConfig,
    ChartContainer,
    ChartLegend,
    ChartLegendContent,
    ChartTooltip,
    ChartTooltipContent,
} from "@/components/ui/chart"

export interface AnomalyDataPoint {
    timestamp: string // ISO string
    value: number
    score?: number
    severity?: "low" | "medium" | "high"
}

interface AnomalyChartProps {
    data: AnomalyDataPoint[]
    anomalies: AnomalyDataPoint[]
    className?: string
}

const chartConfig = {
    value: {
        label: "Value",
        color: "hsl(var(--primary))",
    },
    anomaly: {
        label: "Anomaly",
        color: "hsl(var(--destructive))",
    },
} satisfies ChartConfig

export function AnomalyChart({ data, anomalies, className }: AnomalyChartProps) {
    const chartData = useMemo(() => {
        // Merge data and anomalies
        // We want to show all data points as a line, and overlay anomalies as points

        // Crietate a map for quick lookup of anomalies
        const anomalyMap = new Map(anomalies.map(a => [a.timestamp, a]))

        return data.map(point => {
            const anomaly = anomalyMap.get(point.timestamp)
            return {
                ...point,
                formattedDate: format(new Date(point.timestamp), "MMM dd HH:mm"),
                anomalyValue: anomaly ? point.value : null, // Only present if anomaly
                score: anomaly?.score,
                severity: anomaly?.severity,
            }
        })
    }, [data, anomalies])

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
                    content={({ active, payload, label }) => {
                        if (active && payload && payload.length) {
                            const dataPoint = payload[0].payload;
                            return (
                                <div className="rounded-lg border bg-background p-2 shadow-sm">
                                    <div className="grid grid-cols-2 gap-2">
                                        <div className="flex flex-col">
                                            <span className="text-[0.70rem] uppercase text-muted-foreground">
                                                Time
                                            </span>
                                            <span className="font-bold text-muted-foreground">
                                                {dataPoint.formattedDate}
                                            </span>
                                        </div>
                                        <div className="flex flex-col">
                                            <span className="text-[0.70rem] uppercase text-muted-foreground">
                                                Value
                                            </span>
                                            <span className="font-bold">
                                                {dataPoint.value}
                                            </span>
                                        </div>
                                        {dataPoint.score && (
                                            <>
                                                <div className="flex flex-col">
                                                    <span className="text-[0.70rem] uppercase text-muted-foreground">
                                                        Score
                                                    </span>
                                                    <span className="font-bold text-destructive">
                                                        {dataPoint.score.toFixed(2)}
                                                    </span>
                                                </div>
                                                <div className="flex flex-col">
                                                    <span className="text-[0.70rem] uppercase text-muted-foreground">
                                                        Severity
                                                    </span>
                                                    <span className="font-bold text-destructive capitalize">
                                                        {dataPoint.severity}
                                                    </span>
                                                </div>
                                            </>
                                        )}
                                    </div>
                                </div>
                            )
                        }
                        return null;
                    }}
                />

                <ChartLegend content={<ChartLegendContent />} />

                {/* Main Data Line */}
                <Line
                    dataKey="value"
                    type="monotone"
                    stroke="var(--color-value)"
                    strokeWidth={2}
                    dot={false}
                    activeDot={{ r: 4 }}
                />

                {/* Anomalies as Scatter points */}
                <Scatter
                    dataKey="anomalyValue"
                    fill="var(--color-anomaly)"
                    r={6}
                />

            </ComposedChart>
        </ChartContainer>
    )
}
