"use client"

export const dynamic = 'force-dynamic';

import { useEffect, useState } from "react"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { LineageGraph } from "@/components/lineage/lineage-graph"
import { Loader2 } from "lucide-react"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"

export default function LineagePage() {
    const [data, setData] = useState<{ nodes: any[], edges: any[] } | null>(null)
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)

    useEffect(() => {
        const fetchData = async () => {
            try {
                // Use the Next.js proxy to handle authentication automatically
                const res = await fetch('/api/go/lineage')

                if (!res.ok) {
                    throw new Error(`Failed to fetch lineage data: ${res.statusText}`)
                }

                const jsonData = await res.json()
                setData(jsonData)
            } catch (err: any) {
                setError(err.message)
            } finally {
                setLoading(false)
            }
        }

        fetchData()
    }, [])

    if (loading) {
        return (
            <div className="flex h-[50vh] items-center justify-center">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
        )
    }

    if (error) {
        return (
            <div className="p-8">
                <Alert variant="destructive">
                    <AlertTitle>Error</AlertTitle>
                    <AlertDescription>{error}</AlertDescription>
                </Alert>
            </div>
        )
    }

    return (
        <div className="container mx-auto py-6 space-y-6">
            <div className="flex flex-col space-y-2">
                <h1 className="text-3xl font-bold tracking-tight">Data Lineage & Governance</h1>
                <p className="text-muted-foreground">
                    Visualize the flow of data from sources to dashboards.
                </p>
            </div>

            <Card>
                <CardHeader>
                    <CardTitle>System Lineage Graph</CardTitle>
                    <CardDescription>
                        Auto-generated map of data dependencies.
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    {data && <LineageGraph data={data} />}
                </CardContent>
            </Card>
        </div>
    )
}
