"use client"

import { useState } from "react"
import { CalendarIcon, TrendingUp } from "lucide-react"

import { Button } from "@/components/ui/button"
import {
    Card,
    CardContent,
    CardDescription,
    CardFooter,
    CardHeader,
    CardTitle,
} from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select"

export interface ForecastConfigData {
    modelType: "linear" | "moving_average"
    horizon: number
}

interface ForecastConfigProps {
    onGenerate: (config: ForecastConfigData) => void
    isLoading?: boolean
}

export function ForecastConfig({ onGenerate, isLoading }: ForecastConfigProps) {
    const [modelType, setModelType] = useState<"linear" | "moving_average">("linear")
    const [horizon, setHorizon] = useState<number>(12)

    const handleGenerate = () => {
        onGenerate({ modelType, horizon })
    }

    return (
        <Card className="w-full max-w-sm">
            <CardHeader>
                <CardTitle>Forecast Settings</CardTitle>
                <CardDescription>Configure the prediction engine.</CardDescription>
            </CardHeader>
            <CardContent className="grid gap-4">
                <div className="grid gap-2">
                    <Label htmlFor="model">Model Type</Label>
                    <Select
                        value={modelType}
                        onValueChange={(val) => setModelType(val as "linear" | "moving_average")}
                    >
                        <SelectTrigger id="model">
                            <SelectValue placeholder="Select model" />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value="linear">Linear Regression (Trend)</SelectItem>
                            <SelectItem value="moving_average">Moving Average (Smoothing)</SelectItem>
                        </SelectContent>
                    </Select>
                </div>
                <div className="grid gap-2">
                    <Label htmlFor="horizon">Forecast Horizon</Label>
                    <div className="flex items-center gap-2">
                        <Input
                            id="horizon"
                            type="number"
                            min={1}
                            max={100}
                            value={horizon}
                            onChange={(e) => setHorizon(parseInt(e.target.value) || 12)}
                        />
                        <span className="text-muted-foreground text-sm">points</span>
                    </div>
                </div>
            </CardContent>
            <CardFooter>
                <Button className="w-full" onClick={handleGenerate} disabled={isLoading}>
                    {isLoading ? (
                        <>Generating...</>
                    ) : (
                        <>
                            <TrendingUp className="mr-2 h-4 w-4" />
                            Generate Forecast
                        </>
                    )}
                </Button>
            </CardFooter>
        </Card>
    )
}
