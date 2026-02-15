"use client"

import { useState, useEffect } from "react"
import { 
  Calculator, 
  Sparkles, 
  Network, 
  TrendingUp, 
  AlertTriangle, 
  Brain,
  BarChart3,
  Activity,
  Lightbulb,
  ArrowUpRight,
  ArrowDownRight,
  Minus,
  Database,
  LineChart,
  FileText,
  AlertCircle
} from "lucide-react"

import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { useToast } from "@/components/ui/use-toast"
import { ForecastView } from "@/components/analytics/forecast-view"
import { AnomalyView } from "@/components/analytics/anomaly-view"
import { AutoInsights } from "@/components/analytics/auto-insights"
import { KeyDrivers } from "@/components/analytics/key-drivers"
import { AnomalyDataPoint } from "@/components/visualizations/anomaly-chart"
import { Insight, CorrelationResult } from "@/types/analytics"
import { cn } from "@/lib/utils"
import Link from "next/link"

interface MetricCardProps {
  title: string
  value: string | number
  change?: string
  trend?: 'up' | 'down' | 'neutral'
  icon: React.ReactNode
  description?: string
  isLoading?: boolean
}

function MetricCard({ title, value, change, trend, icon, description, isLoading }: MetricCardProps) {
  if (isLoading) {
    return (
      <Card className="relative overflow-hidden">
        <CardHeader className="flex flex-row items-center justify-between pb-2">
          <Skeleton className="h-4 w-24" />
          <Skeleton className="h-8 w-8 rounded-lg" />
        </CardHeader>
        <CardContent>
          <Skeleton className="h-8 w-20 mb-2" />
          <Skeleton className="h-4 w-32" />
        </CardContent>
      </Card>
    )
  }

  return (
    <Card className="relative overflow-hidden">
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <CardTitle className="text-sm font-medium text-muted-foreground">
          {title}
        </CardTitle>
        <div className="h-8 w-8 rounded-lg bg-primary/10 flex items-center justify-center text-primary">
          {icon}
        </div>
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold">{value}</div>
        {(change || description) && (
          <div className="flex items-center gap-2 mt-1">
            {change && trend && (
              <Badge 
                variant={trend === 'up' ? 'default' : trend === 'down' ? 'destructive' : 'secondary'}
                className="text-xs"
              >
                {trend === 'up' && <ArrowUpRight className="h-3 w-3 mr-1" />}
                {trend === 'down' && <ArrowDownRight className="h-3 w-3 mr-1" />}
                {trend === 'neutral' && <Minus className="h-3 w-3 mr-1" />}
                {change}
              </Badge>
            )}
            {description && (
              <span className="text-xs text-muted-foreground">{description}</span>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  )
}

interface Query {
  id: string
  name: string
  created_at: string
  updated_at: string
}

interface Connection {
  id: string
  name: string
  type: string
  created_at: string
}

export function AnalyticsView() {
    const { toast } = useToast()
    const [activeTab, setActiveTab] = useState("forecast")
    
    // Real data states
    const [queries, setQueries] = useState<Query[]>([])
    const [connections, setConnections] = useState<Connection[]>([])
    const [isLoadingData, setIsLoadingData] = useState(true)
    
    // Analytics results
    const [insights, setInsights] = useState<Insight[]>([])
    const [correlations, setCorrelations] = useState<CorrelationResult[]>([])
    const [isLoadingAnalytics, setIsLoadingAnalytics] = useState(false)
    
    // Fetch real data from backend
    useEffect(() => {
        const fetchData = async () => {
            setIsLoadingData(true)
            try {
                const [queriesRes, connectionsRes] = await Promise.all([
                    fetch('/api/queries'),
                    fetch('/api/connections')
                ])

                if (queriesRes.ok) {
                    const queriesData = await queriesRes.json()
                    setQueries(queriesData.data || queriesData || [])
                }

                if (connectionsRes.ok) {
                    const connectionsData = await connectionsRes.json()
                    setConnections(connectionsData.data || connectionsData || [])
                }
            } catch (error) {
                console.error("Failed to fetch data:", error)
                toast({
                    title: "Error",
                    description: "Failed to load data from server",
                    variant: "destructive"
                })
            } finally {
                setIsLoadingData(false)
            }
        }

        fetchData()
    }, [toast])

    // Calculate metrics from real data
    const totalQueries = queries.length
    const totalConnections = connections.length
    const recentQueries = queries.filter(q => {
        const daysSince = (Date.now() - new Date(q.updated_at || q.created_at).getTime()) / (1000 * 60 * 60 * 24)
        return daysSince <= 7
    }).length

    // Generate analytics from real data
    useEffect(() => {
        if (queries.length === 0) return

        const generateAnalytics = async () => {
            setIsLoadingAnalytics(true)
            try {
                // Convert queries to time-series data
                const timeSeriesData = queries
                    .filter(q => q.created_at)
                    .map(q => ({
                        timestamp: q.created_at,
                        value: 1,
                        queries: 1,
                        dayOfWeek: new Date(q.created_at).getDay(),
                        hour: new Date(q.created_at).getHours()
                    }))
                    .sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime())

                if (timeSeriesData.length < 3) {
                    // Not enough data for meaningful analytics
                    setInsights([{
                        id: 'insufficient-data',
                        type: 'statistic',
                        title: 'Insufficient Data',
                        description: `You have ${queries.length} queries. Create more queries to see AI-powered insights.`,
                        metric: 'queries',
                        value: queries.length,
                        confidence: 1,
                        createdAt: new Date().toISOString()
                    }])
                    setIsLoadingAnalytics(false)
                    return
                }

                // Generate insights
                const insightsRes = await fetch('/api/analytics/insights', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        data: timeSeriesData,
                        metricCol: 'queries',
                        timeCol: 'timestamp'
                    })
                })

                if (insightsRes.ok) {
                    const data = await insightsRes.json()
                    setInsights(data)
                }

                // Calculate correlations
                const corrRes = await fetch('/api/analytics/correlations', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        data: timeSeriesData,
                        cols: ['value', 'dayOfWeek', 'hour']
                    })
                })

                if (corrRes.ok) {
                    const data = await corrRes.json()
                    setCorrelations(data)
                }
            } catch (error) {
                console.error("Failed to generate analytics:", error)
            } finally {
                setIsLoadingAnalytics(false)
            }
        }

        generateAnalytics()
    }, [queries])

    const hasData = queries.length > 0 || connections.length > 0

    if (!hasData && !isLoadingData) {
        return (
            <div className="flex flex-col gap-8 p-6 max-w-[1600px] mx-auto">
                {/* Header */}
                <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                        <div className="h-10 w-10 rounded-xl bg-gradient-to-br from-primary to-primary/60 flex items-center justify-center text-primary-foreground shadow-lg">
                            <BarChart3 className="h-6 w-6" />
                        </div>
                        <div>
                            <h1 className="text-3xl font-bold tracking-tight">Advanced Analytics</h1>
                            <p className="text-muted-foreground mt-1">
                                AI-powered insights and predictive analytics for your business
                            </p>
                        </div>
                    </div>
                </div>

                {/* Empty State */}
                <Card className="flex flex-col items-center justify-center py-20 text-center">
                    <div className="h-20 w-20 rounded-full bg-muted/50 flex items-center justify-center mb-6">
                        <Database className="h-10 w-10 text-muted-foreground" />
                    </div>
                    <h2 className="text-2xl font-bold mb-2">No Data Available</h2>
                    <p className="text-muted-foreground max-w-md mb-8">
                        Analytics requires data to analyze. Start by creating connections and running queries to see AI-powered insights.
                    </p>
                    <div className="flex gap-4">
                        <Link href="/connections">
                            <Button className="gap-2">
                                <Database className="h-4 w-4" />
                                Add Connection
                            </Button>
                        </Link>
                        <Link href="/query-builder">
                            <Button variant="outline" className="gap-2">
                                <LineChart className="h-4 w-4" />
                                Create Query
                            </Button>
                        </Link>
                    </div>
                </Card>
            </div>
        )
    }

    return (
        <div className="flex flex-col gap-8 p-6 max-w-[1600px] mx-auto">
            {/* Header */}
            <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                    <div className="h-10 w-10 rounded-xl bg-gradient-to-br from-primary to-primary/60 flex items-center justify-center text-primary-foreground shadow-lg">
                        <BarChart3 className="h-6 w-6" />
                    </div>
                    <div>
                        <h1 className="text-3xl font-bold tracking-tight">Advanced Analytics</h1>
                        <p className="text-muted-foreground mt-1">
                            AI-powered insights and predictive analytics for your business
                        </p>
                    </div>
                </div>
                <div className="flex items-center gap-2">
                    <Badge variant="outline" className="gap-1">
                        <Brain className="h-3 w-3" />
                        AI Powered
                    </Badge>
                    <Badge variant="outline" className="gap-1">
                        <Activity className="h-3 w-3" />
                        Real-time
                    </Badge>
                </div>
            </div>

            {/* Key Metrics */}
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                <MetricCard
                    title="Total Queries"
                    value={totalQueries}
                    icon={<FileText className="h-4 w-4" />}
                    description="all time"
                    isLoading={isLoadingData}
                />
                <MetricCard
                    title="Recent Queries"
                    value={recentQueries}
                    change={`${totalQueries > 0 ? Math.round((recentQueries / totalQueries) * 100) : 0}%`}
                    trend={recentQueries > 0 ? 'up' : 'neutral'}
                    icon={<TrendingUp className="h-4 w-4" />}
                    description="last 7 days"
                    isLoading={isLoadingData}
                />
                <MetricCard
                    title="Connections"
                    value={totalConnections}
                    icon={<Database className="h-4 w-4" />}
                    description="data sources"
                    isLoading={isLoadingData}
                />
                <MetricCard
                    title="AI Insights"
                    value={isLoadingAnalytics ? "..." : insights.length}
                    change={insights.length > 0 ? "Active" : "None"}
                    trend={insights.length > 0 ? 'up' : 'neutral'}
                    icon={<Lightbulb className="h-4 w-4" />}
                    description="generated"
                    isLoading={isLoadingAnalytics}
                />
            </div>

            {/* Main Content */}
            <div className="grid gap-6 lg:grid-cols-3">
                {/* Left Column - Main Analytics */}
                <div className="lg:col-span-2">
                    <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-4">
                        <Card className="border-0 shadow-none bg-transparent">
                            <CardContent className="p-0">
                                <TabsList className="grid w-full grid-cols-3 lg:max-w-md">
                                    <TabsTrigger value="forecast" className="gap-2">
                                        <TrendingUp className="h-4 w-4" />
                                        <span className="hidden sm:inline">Trend Analysis</span>
                                        <span className="sm:hidden">Trends</span>
                                    </TabsTrigger>
                                    <TabsTrigger value="anomaly" className="gap-2">
                                        <AlertTriangle className="h-4 w-4" />
                                        <span className="hidden sm:inline">Anomaly Detection</span>
                                        <span className="sm:hidden">Anomaly</span>
                                    </TabsTrigger>
                                    <TabsTrigger value="correlations" className="gap-2">
                                        <Network className="h-4 w-4" />
                                        <span className="hidden sm:inline">Key Drivers</span>
                                        <span className="sm:hidden">Drivers</span>
                                    </TabsTrigger>
                                </TabsList>
                            </CardContent>
                        </Card>

                        <TabsContent value="forecast" className="mt-0 space-y-4">
                            <Card>
                                <CardHeader>
                                    <div className="flex items-center gap-2">
                                        <div className="h-8 w-8 rounded-lg bg-primary/10 flex items-center justify-center">
                                            <TrendingUp className="h-4 w-4 text-primary" />
                                        </div>
                                        <div>
                                            <CardTitle className="text-lg">Query Activity Trends</CardTitle>
                                            <CardDescription>
                                                Analyze your query creation patterns over time
                                            </CardDescription>
                                        </div>
                                    </div>
                                </CardHeader>
                                <CardContent>
                                    {queries.length < 3 ? (
                                        <div className="flex flex-col items-center justify-center py-12 text-center">
                                            <AlertCircle className="h-12 w-12 text-muted-foreground mb-4" />
                                            <p className="text-muted-foreground">
                                                Need at least 3 queries to generate trend analysis.
                                                <br />
                                                <Link href="/query-builder" className="text-primary hover:underline">
                                                    Create more queries
                                                </Link>
                                            </p>
                                        </div>
                                    ) : (
                                        <div className="h-[300px] flex items-center justify-center text-muted-foreground">
                                            <LineChart className="h-8 w-8 mr-2" />
                                            Trend visualization coming soon
                                        </div>
                                    )}
                                </CardContent>
                            </Card>
                        </TabsContent>
                        
                        <TabsContent value="anomaly" className="mt-0 space-y-4">
                            <Card>
                                <CardHeader>
                                    <div className="flex items-center gap-2">
                                        <div className="h-8 w-8 rounded-lg bg-amber-500/10 flex items-center justify-center">
                                            <AlertTriangle className="h-4 w-4 text-amber-500" />
                                        </div>
                                        <div>
                                            <CardTitle className="text-lg">Activity Anomalies</CardTitle>
                                            <CardDescription>
                                                Detect unusual patterns in your data activity
                                            </CardDescription>
                                        </div>
                                    </div>
                                </CardHeader>
                                <CardContent>
                                    {queries.length < 5 ? (
                                        <div className="flex flex-col items-center justify-center py-12 text-center">
                                            <AlertCircle className="h-12 w-12 text-muted-foreground mb-4" />
                                            <p className="text-muted-foreground">
                                                Need at least 5 queries for anomaly detection.
                                                <br />
                                                <Link href="/query-builder" className="text-primary hover:underline">
                                                    Create more queries
                                                </Link>
                                            </p>
                                        </div>
                                    ) : (
                                        <div className="h-[300px] flex items-center justify-center text-muted-foreground">
                                            <Activity className="h-8 w-8 mr-2" />
                                            Anomaly detection visualization coming soon
                                        </div>
                                    )}
                                </CardContent>
                            </Card>
                        </TabsContent>
                        
                        <TabsContent value="correlations" className="mt-0 space-y-4">
                            <Card>
                                <CardHeader className="pb-3">
                                    <div className="flex items-center justify-between">
                                        <div className="flex items-center gap-2">
                                            <div className="h-8 w-8 rounded-lg bg-primary/10 flex items-center justify-center">
                                                <Network className="h-4 w-4 text-primary" />
                                            </div>
                                            <div>
                                                <CardTitle className="text-lg">Correlation Analysis</CardTitle>
                                                <CardDescription>
                                                    Discover relationships between your metrics
                                                </CardDescription>
                                            </div>
                                        </div>
                                    </div>
                                </CardHeader>
                                <CardContent>
                                    <KeyDrivers correlations={correlations} isLoading={isLoadingAnalytics} />
                                </CardContent>
                            </Card>
                        </TabsContent>
                    </Tabs>
                </div>

                {/* Right Column - AI Insights */}
                <div className="lg:col-span-1">
                    <Card className="h-full border-l-4 border-l-primary">
                        <CardHeader className="pb-3">
                            <div className="flex items-center gap-2">
                                <div className="h-8 w-8 rounded-lg bg-gradient-to-br from-primary to-primary/60 flex items-center justify-center text-primary-foreground">
                                    <Sparkles className="h-4 w-4" />
                                </div>
                                <div>
                                    <CardTitle className="text-lg">AI Insights</CardTitle>
                                    <CardDescription>
                                        Automated discoveries from your data
                                    </CardDescription>
                                </div>
                            </div>
                        </CardHeader>
                        <CardContent>
                            <AutoInsights insights={insights} isLoading={isLoadingAnalytics} />
                        </CardContent>
                    </Card>
                </div>
            </div>
        </div>
    )
}
