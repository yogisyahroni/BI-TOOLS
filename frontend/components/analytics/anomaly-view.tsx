"use client";

import { Loader2, AlertTriangle } from "lucide-react";
import { fetchWithAuth } from "@/lib/utils";
import { useState } from "react";
import { AnomalyChart, type AnomalyDataPoint } from "@/components/visualizations/anomaly-chart";
import { AnomalyConfig, type AnomalyConfigData } from "@/components/analytics/anomaly-config";
import { useToast } from "@/components/ui/use-toast";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { ReportGenerator } from "@/components/analytics/report-generator";

interface AnomalyViewProps {
  history: AnomalyDataPoint[]; // Base data to analyze
}

export function AnomalyView({ history }: AnomalyViewProps) {
  const [anomalies, setAnomalies] = useState<AnomalyDataPoint[]>([]);
  const [loading, setLoading] = useState(false);
  const { toast } = useToast();

  const handleDetectAnomalies = async (config: AnomalyConfigData) => {
    setLoading(true);
    try {
      const res = await fetchWithAuth("/api/go/anomalies", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          data: history,
          method: config.method,
          sensitivity: config.sensitivity,
        }),
      });

      if (!res.ok) {
        const errorData = await res.json();
        throw new Error(errorData.error || "Failed to detect anomalies");
      }

      const result = await res.json();
      // Backend returns { anomalies: [...], summary: ... }
      // anomalies structure: { timestamp, value, score, severity }
      setAnomalies(result.anomalies || []);

      toast({
        title: "Analysis Complete",
        description: `Detected ${result.anomalies?.length || 0} anomalies using ${config.method}.`,
      });
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } catch (error: any) {
      console.error(error);
      toast({
        title: "Error",
        description: error.message || "Failed to detect anomalies.",
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-end">
        <ReportGenerator data={anomalies} filename="anomalies_report" title="Export Anomalies" />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
        {/* Configuration Panel */}
        <div className="lg:col-span-1">
          <AnomalyConfig onDetect={handleDetectAnomalies} isLoading={loading} />

          {/* Anomaly Stats Summary */}
          {anomalies.length > 0 && (
            <Card className="mt-4">
              <CardHeader>
                <CardTitle className="text-lg">Summary</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-2">
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">Total Detected:</span>
                    <span className="font-bold">{anomalies.length}</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">High Severity:</span>
                    <span className="font-bold text-destructive">
                      {anomalies.filter((a) => a.severity === "high").length}
                    </span>
                  </div>
                </div>
              </CardContent>
            </Card>
          )}
        </div>

        {/* Chart View */}
        <div className="lg:col-span-3">
          <Card className="h-full min-h-[500px] flex flex-col">
            <CardHeader>
              <CardTitle>Anomaly Analysis</CardTitle>
              <CardDescription>
                Visualizing data points with detected irregularities.
              </CardDescription>
            </CardHeader>
            <CardContent className="flex-1">
              <AnomalyChart data={history} anomalies={anomalies} className="h-[400px] w-full" />
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
