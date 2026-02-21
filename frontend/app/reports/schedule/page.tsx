"use client";

export const dynamic = "force-dynamic";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Plus, Search, RefreshCw, Clock, Calendar } from "lucide-react";
import { scheduledReportsApi } from "@/lib/api/scheduled-reports";
import type { ScheduledReportResponse, ScheduledReportFilter } from "@/types/scheduled-reports";
import { ReportScheduleCard } from "@/components/reports/report-schedule-card";
import { ReportScheduleForm } from "@/components/reports/report-schedule-form";
import { ReportHistory } from "@/components/reports/report-history";
import { toast } from "sonner";

export default function ScheduledReportsPage() {
  const queryClient = useQueryClient();

  const [searchQuery, setSearchQuery] = useState("");
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [isEditDialogOpen, setIsEditDialogOpen] = useState(false);
  const [isHistoryDialogOpen, setIsHistoryDialogOpen] = useState(false);
  const [selectedReport, setSelectedReport] = useState<ScheduledReportResponse | null>(null);
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const [filter, _setFilter] = useState<ScheduledReportFilter>({
    page: 1,
    limit: 20,
  });

  const {
    data: response,
    isLoading: loading,
    refetch: fetchReports,
  } = useQuery({
    queryKey: ["scheduledReports", filter, searchQuery],
    queryFn: () =>
      scheduledReportsApi.list({
        ...filter,
        search: searchQuery || undefined,
      }),
  });

  const reports = response?.reports || [];

  const createMutation = useMutation({
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    mutationFn: (data: any) => scheduledReportsApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["scheduledReports"] });
      toast.success("Scheduled report created successfully");
      setIsCreateDialogOpen(false);
    },
    onError: (error) => {
      toast.error("Failed to create scheduled report");
      console.error(error);
    },
  });

  const updateMutation = useMutation({
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    mutationFn: ({ id, data }: { id: string; data: any }) => scheduledReportsApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["scheduledReports"] });
      toast.success("Scheduled report updated successfully");
      setIsEditDialogOpen(false);
      setSelectedReport(null);
    },
    onError: (error) => {
      toast.error("Failed to update scheduled report");
      console.error(error);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => scheduledReportsApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["scheduledReports"] });
      toast.success("Scheduled report deleted");
    },
    onError: (error) => {
      toast.error("Failed to delete scheduled report");
      console.error(error);
    },
  });

  const toggleMutation = useMutation({
    mutationFn: (id: string) => scheduledReportsApi.toggleActive(id),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ["scheduledReports"] });
      const report = reports.find((r) => r.id === variables);
      if (report) {
        toast.success(`Report ${report.isActive ? "paused" : "activated"}`);
      } else {
        toast.success(`Report status updated`);
      }
    },
    onError: (error) => {
      toast.error("Failed to update report status");
      console.error(error);
    },
  });

  const runNowMutation = useMutation({
    mutationFn: (id: string) => scheduledReportsApi.trigger(id),
    onSuccess: (res) => {
      toast.success(`Report generation started. Run ID: ${res.runId}`);
    },
    onError: (error) => {
      toast.error("Failed to trigger report");
      console.error(error);
    },
  });

  // Handlers
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const handleCreate = async (data: any) => {
    createMutation.mutate(data);
  };
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const handleEdit = async (data: any) => {
    if (selectedReport) updateMutation.mutate({ id: selectedReport.id, data });
  };
  const handleDelete = (report: ScheduledReportResponse) => {
    // eslint-disable-next-line no-alert
    if (!confirm("Are you sure you want to delete this scheduled report?")) return;
    deleteMutation.mutate(report.id);
  };
  const handleToggle = (report: ScheduledReportResponse) => toggleMutation.mutate(report.id);
  const handleRunNow = (report: ScheduledReportResponse) => runNowMutation.mutate(report.id);

  const handleViewHistory = (report: ScheduledReportResponse) => {
    setSelectedReport(report);
    setIsHistoryDialogOpen(true);
  };

  return (
    <div className="container mx-auto py-6 px-4 max-w-7xl">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4 mb-8">
        <div>
          <h1 className="text-3xl font-bold tracking-tight flex items-center gap-3">
            <Clock className="w-8 h-8 text-primary" />
            Scheduled Reports
          </h1>
          <p className="text-muted-foreground mt-2">
            Automate report generation and delivery to your team
          </p>
        </div>
        <div className="flex items-center gap-3">
          <Button variant="outline" onClick={() => fetchReports()} disabled={loading}>
            <RefreshCw className={`w-4 h-4 mr-2 ${loading ? "animate-spin" : ""}`} />
            Refresh
          </Button>
          <Button onClick={() => setIsCreateDialogOpen(true)}>
            <Plus className="w-4 h-4 mr-2" />
            New Schedule
          </Button>
        </div>
      </div>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-4 mb-6">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <Input
            placeholder="Search scheduled reports..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-10"
          />
        </div>
      </div>

      {/* Reports Grid */}
      {loading ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {[...Array(6)].map((_, i) => (
            <Skeleton key={i} className="h-48 w-full rounded-lg" />
          ))}
        </div>
      ) : reports.length === 0 ? (
        <div className="text-center py-20">
          <div className="bg-muted/50 w-20 h-20 rounded-full flex items-center justify-center mx-auto mb-6">
            <Calendar className="w-10 h-10 text-muted-foreground" />
          </div>
          <h3 className="text-xl font-semibold mb-2">No Scheduled Reports</h3>
          <p className="text-muted-foreground max-w-md mx-auto mb-6">
            Create automated reports to deliver insights to your team on a schedule
          </p>
          <Button onClick={() => setIsCreateDialogOpen(true)}>
            <Plus className="w-4 h-4 mr-2" />
            Create Your First Schedule
          </Button>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {reports.map((report) => (
            <ReportScheduleCard
              key={report.id}
              report={report}
              onEdit={(r) => {
                setSelectedReport(r);
                setIsEditDialogOpen(true);
              }}
              onDelete={handleDelete}
              onToggle={handleToggle}
              onRunNow={handleRunNow}
              onViewHistory={handleViewHistory}
            />
          ))}
        </div>
      )}

      {/* Create Dialog */}
      <Dialog open={isCreateDialogOpen} onOpenChange={setIsCreateDialogOpen}>
        <DialogContent className="max-w-4xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Create Scheduled Report</DialogTitle>
            <DialogDescription>Set up automated report generation and delivery</DialogDescription>
          </DialogHeader>
          <ReportScheduleForm
            onSubmit={handleCreate}
            onCancel={() => setIsCreateDialogOpen(false)}
          />
        </DialogContent>
      </Dialog>

      {/* Edit Dialog */}
      <Dialog open={isEditDialogOpen} onOpenChange={setIsEditDialogOpen}>
        <DialogContent className="max-w-4xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Edit Scheduled Report</DialogTitle>
            <DialogDescription>Update your scheduled report configuration</DialogDescription>
          </DialogHeader>
          {selectedReport && (
            <ReportScheduleForm
              initialData={{
                name: selectedReport.name,
                description: selectedReport.description || "",
                resourceType: selectedReport.resourceType,
                resourceId: selectedReport.resourceId,
                scheduleType: selectedReport.scheduleType,
                cronExpr: selectedReport.cronExpr || "",
                timeOfDay: selectedReport.timeOfDay || "09:00",
                dayOfWeek: selectedReport.dayOfWeek ?? 1,
                dayOfMonth: selectedReport.dayOfMonth ?? 1,
                // eslint-disable-next-line @typescript-eslint/no-explicit-any
                timezone: selectedReport.timezone,
                recipients: selectedReport.recipients.map((r) => ({
                  email: r.email,
                  // eslint-disable-next-line @typescript-eslint/no-explicit-any
                  type: r.type as any,
                })),
                format: selectedReport.format,
                includeFilters: selectedReport.includeFilters,
                subject: selectedReport.subject,
                message: selectedReport.message,
              }}
              onSubmit={handleEdit}
              onCancel={() => {
                setIsEditDialogOpen(false);
                setSelectedReport(null);
              }}
            />
          )}
        </DialogContent>
      </Dialog>

      {/* History Dialog */}
      <Dialog open={isHistoryDialogOpen} onOpenChange={setIsHistoryDialogOpen}>
        <DialogContent className="max-w-3xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Report History</DialogTitle>
            <DialogDescription>View past runs for {selectedReport?.name}</DialogDescription>
          </DialogHeader>
          {selectedReport && <ReportHistory reportId={selectedReport.id} />}
        </DialogContent>
      </Dialog>
    </div>
  );
}
