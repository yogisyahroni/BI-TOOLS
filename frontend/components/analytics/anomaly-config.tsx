"use client"

import { useState } from "react"
import { Activity, _Settings2 } from "lucide-react"

import { Button } from "@/components/ui/button"
import {
    Card,
    CardContent,
    CardDescription,
    CardFooter,
    CardHeader,
    CardTitle,
} from "@/components/ui/card"
import { _Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select"
import { Slider } from "@/components/ui/slider"

export interface AnomalyConfigData {
    method: "z-score" | "iqr"
    sensitivity: number
}

interface AnomalyConfigProps {
    onDetect: (config: AnomalyConfigData) => void
    isLoading?: boolean
}

export function AnomalyConfig({ onDetect, isLoading }: AnomalyConfigProps) {
    const [method, setMethod] = useState<"z-score" | "iqr">("z-score")
    const [sensitivity, setSensitivity] = useState<number>(3.0)

    const handleDetect = () => {
        onDetect({ method, sensitivity })
    }

    return (
        <Card className="w-full max-w-sm">
            <CardHeader>
                <CardTitle>Anomaly Detection</CardTitle>
                <CardDescription>Configure detection algorithms.</CardDescription>
            </CardHeader>
            <CardContent className="grid gap-4">
                <div className="grid gap-2">
                    <Label htmlFor="method">Method</Label>
                    <Select
                        value={method}
                        onValueChange={(val) => {
                            setMethod(val as "z-score" | "iqr")
                            // Reset sensitivity to defaults when changing method
                            if (val === "z-score") setSensitivity(3.0)
                            else if (val === "iqr") setSensitivity(1.5)
                        }}
                    >
                        <SelectTrigger id="method">
                            <SelectValue placeholder="Select method" />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value="z-score">Z-Score (Standard Deviation)</SelectItem>
                            <SelectItem value="iqr">Interquartile Range (Robust)</SelectItem>
                        </SelectContent>
                    </Select>
                </div>
                <div className="grid gap-2">
                    <div className="flex items-center justify-between">
                        <Label htmlFor="sensitivity">Sensitivity (Threshold)</Label>
                        <span className="text-sm font-medium text-muted-foreground">{sensitivity.toFixed(1)}</span>
                    </div>
                    <Slider
                        id="sensitivity"
                        min={1.0}
                        max={5.0}
                        step={0.1}
                        value={[sensitivity]}
                        onValueChange={(vals) => setSensitivity(vals[0])}
                    />
                    <p className="text-xs text-muted-foreground">
                        {method === "z-score"
                            ? "Standard deviations from mean (Default: 3.0)"
                            : "Multiplier for IQR (Default: 1.5)"}
                    </p>
                </div>
            </CardContent>
            <CardFooter>
                <Button className="w-full" onClick={handleDetect} disabled={isLoading}>
                    {isLoading ? (
                        <>Detecting...</>
                    ) : (
                        <>
                            <Activity className="mr-2 h-4 w-4" />
                            Detect Anomalies
                        </>
                    )}
                </Button>
            </CardFooter>
        </Card>
    )
}
