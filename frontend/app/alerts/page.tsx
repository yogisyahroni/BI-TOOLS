"use client";

export const dynamic = "force-dynamic";

import { useState, useEffect } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Plus, Bell, Activity, CheckCircle2, VolumeX } from "lucide-react";
import { useToast } from "@/hooks/use-toast";
import { AlertList } from "@/components/alerts/alert-list";
import { TriggeredAlerts } from "@/components/alerts/triggered-alerts";
import { AlertCreateDialog } from "@/components/alerts/alert-create-dialog";
import { alertsApi } from "@/lib/api/alerts";
import type { Alert, AlertStats, TriggeredAlert, CreateAlertRequest } from "@/types/alerts";
import { SidebarLayout } from "@/components/sidebar-layout";

export default function AlertsPage() {
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [triggeredAlerts, setTriggeredAlerts] = useState<TriggeredAlert[]>([]);
  const [stats, setStats] = useState<AlertStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const { toast } = useToast();

  const fetchData = async () => {
    setLoading(true);
    try {
      const [alertsRes, triggeredRes, statsRes] = await Promise.all([
        alertsApi.list({ limit: 100 }),
        alertsApi.getTriggered(),
        alertsApi.getStats(),
      ]);

      setAlerts(alertsRes.alerts);
      setTriggeredAlerts(triggeredRes.alerts);
      setStats(statsRes);
    } catch (error) {
      console.error("Failed to fetch alerts data:", error);
      toast({
        title: "Error",
        description: "Failed to load alerts data",
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
    // eslint-disable-next-line react-hooks/exhaustive-deps
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleCreateAlert = async (data: CreateAlertRequest) => {
    try {
      await alertsApi.create(data);
      toast({
        title: "Success",
        description: "Alert created successfully",
      });
      fetchData();
      setIsCreateDialogOpen(false);
    } catch (error) {
      console.error("Failed to create alert:", error);
      toast({
        title: "Error",
        description: "Failed to create alert",
        variant: "destructive",
      });
    }
  };

  const handleDeleteAlert = async (alert: Alert) => {
    // eslint-disable-next-line no-alert
    if (!confirm("Are you sure you want to delete this alert?")) return;

    try {
      await alertsApi.delete(alert.id);
      toast({
        title: "Success",
        description: "Alert deleted successfully",
      });
      fetchData();
    } catch (error) {
      console.error("Failed to delete alert:", error);
      toast({
        title: "Error",
        description: "Failed to delete alert",
        variant: "destructive",
      });
    }
  };

  const handleAcknowledgeAlert = async (alert: Alert) => {
    try {
      await alertsApi.acknowledge(alert.id);
      toast({
        title: "Success",
        description: "Alert acknowledged",
      });
      fetchData();
    } catch (error) {
      console.error("Failed to acknowledge alert:", error);
      toast({
        title: "Error",
        description: "Failed to acknowledge alert",
        variant: "destructive",
      });
    }
  };

  const handleMuteAlert = async (alert: Alert, duration?: number) => {
    try {
      await alertsApi.mute(alert.id, { duration });
      toast({
        title: "Success",
        description: duration ? `Alert muted for ${duration} minutes` : "Alert muted indefinitely",
      });
      fetchData();
    } catch (error) {
      console.error("Failed to mute alert:", error);
      toast({
        title: "Error",
        description: "Failed to mute alert",
        variant: "destructive",
      });
    }
  };

  const handleUnmuteAlert = async (alert: Alert) => {
    try {
      await alertsApi.unmute(alert.id);
      toast({
        title: "Success",
        description: "Alert unmuted",
      });
      fetchData();
    } catch (error) {
      console.error("Failed to unmute alert:", error);
      toast({
        title: "Error",
        description: "Failed to unmute alert",
        variant: "destructive",
      });
    }
  };

  const handleTriggerCheck = async (alert: Alert) => {
    try {
      const result = await alertsApi.triggerCheck(alert.id);
      toast({
        title: "Check Complete",
        description: `Alert check completed: ${result.history.status}`,
      });
      fetchData();
    } catch (error) {
      console.error("Failed to trigger alert check:", error);
      toast({
        title: "Error",
        description: "Failed to run alert check",
        variant: "destructive",
      });
    }
  };

  const handleAcknowledgeAll = async () => {
    try {
      await Promise.all(
        triggeredAlerts
          .filter((ta) => !ta.acknowledged)
          .map((ta) => alertsApi.acknowledge(ta.alert.id)),
      );
      toast({
        title: "Success",
        description: "All triggered alerts acknowledged",
      });
      fetchData();
    } catch (error) {
      console.error("Failed to acknowledge all alerts:", error);
      toast({
        title: "Error",
        description: "Failed to acknowledge alerts",
        variant: "destructive",
      });
    }
  };

  return (
    <SidebarLayout>
      <div className="space-y-8">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Data Alerts</h1>
            <p className="text-muted-foreground mt-2">
              Monitor your data and receive notifications when metrics cross thresholds.
            </p>
          </div>
          <Button onClick={() => setIsCreateDialogOpen(true)}>
            <Plus className="mr-2 h-4 w-4" /> Create Alert
          </Button>
        </div>

        {/* Stats Cards */}
        {stats && (
          <div className="grid gap-6 md:grid-cols-4">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Total Alerts</CardTitle>
                <Bell className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{stats.total}</div>
                <p className="text-xs text-muted-foreground">{stats.active} active</p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Triggered</CardTitle>
                <Activity className="h-4 w-4 text-red-500" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-red-600">{stats.triggered}</div>
                <p className="text-xs text-muted-foreground">Require attention</p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Acknowledged</CardTitle>
                <CheckCircle2 className="h-4 w-4 text-indigo-500" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{stats.acknowledged}</div>
                <p className="text-xs text-muted-foreground">Being handled</p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Muted</CardTitle>
                <VolumeX className="h-4 w-4 text-gray-500" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{stats.muted}</div>
                <p className="text-xs text-muted-foreground">Temporarily silenced</p>
              </CardContent>
            </Card>
          </div>
        )}

        {/* Main Content */}
        <Tabs defaultValue="all" className="space-y-6">
          <TabsList>
            <TabsTrigger value="all">All Alerts</TabsTrigger>
            <TabsTrigger value="triggered">
              Triggered
              {triggeredAlerts.length > 0 && (
                <span className="ml-2 bg-red-100 text-red-800 text-xs px-2 py-0.5 rounded-full">
                  {triggeredAlerts.length}
                </span>
              )}
            </TabsTrigger>
          </TabsList>

          <TabsContent value="all" className="space-y-6">
            <Card>
              <CardHeader>
                <CardTitle>Your Alerts</CardTitle>
                <CardDescription>Manage your data monitoring rules</CardDescription>
              </CardHeader>
              <CardContent>
                <AlertList
                  alerts={alerts}
                  loading={loading}
                  onDelete={handleDeleteAlert}
                  onAcknowledge={handleAcknowledgeAlert}
                  onMute={handleMuteAlert}
                  onUnmute={handleUnmuteAlert}
                  onTriggerCheck={handleTriggerCheck}
                />
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="triggered">
            <TriggeredAlerts
              alerts={triggeredAlerts}
              onAcknowledge={(id) => {
                const alert = alerts.find((a) => a.id === id);
                // eslint-disable-next-line no-alert
                if (alert) handleAcknowledgeAlert(alert);
              }}
              onAcknowledgeAll={handleAcknowledgeAll}
              loading={loading}
            />
          </TabsContent>
        </Tabs>

        <AlertCreateDialog
          open={isCreateDialogOpen}
          onOpenChange={setIsCreateDialogOpen}
          onSubmit={handleCreateAlert}
        />
      </div>
    </SidebarLayout>
  );
}
