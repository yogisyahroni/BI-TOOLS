"use client"

import { _Loader2 } from 'lucide-react';
import { fetchWithAuth } from '@/lib/utils';
import { useState } from "react"
import { ForecastChart, type DataPoint } from "@/components/visualizations/forecast-chart"
import { ForecastConfig, type ForecastConfigData } from "@/components/analytics/forecast-config"
import { useToast } from "@/components/ui/use-toast"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { ReportGenerator } from "@/components/analytics/report-generator"

interface ForecastViewProps {
    history: DataPoint[]
}

export function ForecastView({ history }: ForecastViewProps) {
    const [forecast, setForecast] = useState<DataPoint[]>([])
    const [loading, setLoading] = useState(false)
    const { toast } = useToast()

    const handleGenerateForecast = async (config: ForecastConfigData) => {
        setLoading(true)
        try {
            const res = await fetchWithAuth('/api/go/forecast', {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({
                    series: history,
                    horizon: config.horizon,
                    model_type: config.modelType,
                }),
            })

            if (!res.ok) {
                throw new Error("Failed to generate forecast")
            }

            const result = await res.json()
            setForecast(result.forecast)

            toast({
                title: "Forecast Generated",
                description: `Successfully generated ${result.forecast.length} data points using ${result.model_used}.`,
            })
        } catch (error) {
            console.error(error)
            toast({
                title: "Error",
                description: "Failed to generate forecast. Please try again.",
                variant: "destructive",
            })
        } finally {
            setLoading(false)
        }
    }

    return (
        <div className="flex flex-col gap-6">
            <div className="flex items-center justify-end">
                <ReportGenerator
                    data={forecast.length > 0 ? [...history, ...forecast] : history}
                    filename="forecast_export"
                    title="Export Forecast"
                />
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
                {/* Configuration Panel */}
                <div className="lg:col-span-1">
                    <ForecastConfig onGenerate={handleGenerateForecast} isLoading={loading} />
                </div>

                {/* Chart View */}
                <div className="lg:col-span-3">
                    <Card className="h-full min-h-[500px] flex flex-col">
                        <CardHeader>
                            <CardTitle>Sales Forecast</CardTitle>
                            <CardDescription>
                                Historical data vs. AI-generated predictions.
                            </CardDescription>
                        </CardHeader>
                        <CardContent className="flex-1">
                            <ForecastChart
                                history={history}
                                forecast={forecast}
                                className="h-[400px] w-full"
                            />
                        </CardContent>
                    </Card>
                </div>
            </div>
        </div>
    )
}
