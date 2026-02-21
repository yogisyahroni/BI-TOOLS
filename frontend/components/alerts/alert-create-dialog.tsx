"use client";

import { useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ConditionBuilder } from "./condition-builder";
import { NotificationConfig } from "./notification-config";
import { AlertReview } from "./alert-review";
import { ChevronLeft, ChevronRight, Save } from "lucide-react";
import type {
  CreateAlertRequest,
  AlertSeverity,
  AlertOperator,
  AlertChannelInput,
} from "@/types/alerts";
import { ALERT_SEVERITIES, SCHEDULE_OPTIONS, COOLDOWN_OPTIONS } from "@/types/alerts";
import { alertsApi } from "@/lib/api/alerts";

interface AlertCreateDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (data: CreateAlertRequest) => void;
}

const STEPS = [
  { id: "basic", label: "Basic Info" },
  { id: "query", label: "Data Source" },
  { id: "condition", label: "Condition" },
  { id: "schedule", label: "Schedule" },
  { id: "notifications", label: "Notifications" },
  { id: "review", label: "Review" },
];

export function AlertCreateDialog({ open, onOpenChange, onSubmit }: AlertCreateDialogProps) {
  const [currentStep, setCurrentStep] = useState(0);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [_testResult, _setTestResult] = useState<{
    success: boolean;
    message: string;
    value?: number;
  } | null>(null);

  // Form state
  const [formData, setFormData] = useState({
    name: "",
    description: "",
    queryId: "",
    queryName: "",
    column: "",
    operator: ">" as AlertOperator,
    threshold: 0,
    schedule: "5m",
    timezone: "UTC",
    severity: "warning" as AlertSeverity,
    cooldownMinutes: 5,
    channels: [] as AlertChannelInput[],
  });

  const [_availableQueries, _setAvailableQueries] = useState<Array<{ id: string; name: string }>>(
    [],
  );
  const [_availableColumns, _setAvailableColumns] = useState<string[]>([]);

  const handleNext = () => {
    if (currentStep < STEPS.length - 1) {
      setCurrentStep(currentStep + 1);
    }
  };

  const handleBack = () => {
    if (currentStep > 0) {
      setCurrentStep(currentStep - 1);
    }
  };

  const handleSubmit = async () => {
    setIsSubmitting(true);
    try {
      await onSubmit({
        name: formData.name,
        description: formData.description || undefined,
        queryId: formData.queryId,
        column: formData.column,
        operator: formData.operator,
        threshold: formData.threshold,
        schedule: formData.schedule,
        timezone: formData.timezone,
        severity: formData.severity,
        cooldownMinutes: formData.cooldownMinutes,
        channels: formData.channels.length > 0 ? formData.channels : undefined,
      });
      // Reset form
      setFormData({
        name: "",
        description: "",
        queryId: "",
        queryName: "",
        column: "",
        operator: ">",
        threshold: 0,
        schedule: "5m",
        timezone: "UTC",
        severity: "warning",
        cooldownMinutes: 5,
        channels: [],
      });
      setCurrentStep(0);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleTestCondition = async () => {
    if (!formData.queryId || !formData.column) {
      return { success: false, message: "Please select a query and column first" };
    }

    try {
      const result = await alertsApi.test({
        name: formData.name,
        description: formData.description,
        queryId: formData.queryId,
        column: formData.column,
        operator: formData.operator,
        threshold: formData.threshold,
      });
      return {
        success: true,
        message: result.triggered ? "Condition would trigger!" : "Condition would not trigger",
        value: result.value,
      };
    } catch (error) {
      return {
        success: false,
        message: error instanceof Error ? error.message : "Test failed",
      };
    }
  };

  const canProceed = () => {
    switch (currentStep) {
      case 0:
        return formData.name.trim() !== "";
      case 1:
        return formData.queryId !== "";
      case 2:
        return formData.column !== "";
      case 3:
        return formData.schedule !== "";
      default:
        return true;
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-3xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Create New Alert</DialogTitle>
          <DialogDescription>
            Step {currentStep + 1} of {STEPS.length}: {STEPS[currentStep].label}
          </DialogDescription>
        </DialogHeader>

        {/* Progress indicator */}
        <div className="flex items-center gap-2 mb-4">
          {STEPS.map((step, index) => (
            <div
              key={step.id}
              className={`flex-1 h-2 rounded-full ${
                index <= currentStep ? "bg-indigo-600" : "bg-gray-200"
              }`}
            />
          ))}
        </div>

        {/* Step content */}
        <div className="py-4">
          {currentStep === 0 && (
            <div className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="name">Alert Name *</Label>
                <Input
                  id="name"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  placeholder="e.g., High CPU Usage"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="description">Description</Label>
                <Textarea
                  id="description"
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                  placeholder="Describe what this alert monitors..."
                  rows={3}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="severity">Severity</Label>
                <Select
                  value={formData.severity}
                  onValueChange={(v) => setFormData({ ...formData, severity: v as AlertSeverity })}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {ALERT_SEVERITIES.map((s) => (
                      <SelectItem key={s.value} value={s.value}>
                        {s.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
          )}

          {currentStep === 1 && (
            <div className="space-y-4">
              <div className="space-y-2">
                <Label>Select Query *</Label>
                <Select
                  value={formData.queryId}
                  onValueChange={(v) => setFormData({ ...formData, queryId: v })}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Choose a saved query..." />
                  </SelectTrigger>
                  <SelectContent>
                    {/* This would be populated from API */}
                    <SelectItem value="query-1">Sales Dashboard Query</SelectItem>
                    <SelectItem value="query-2">User Metrics</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
          )}

          {currentStep === 2 && (
            <ConditionBuilder
              column={formData.column}
              operator={formData.operator}
              threshold={formData.threshold}
              availableColumns={["cpu_usage", "memory_usage", "active_users"]}
              onChange={(values) => setFormData({ ...formData, ...values })}
              onTest={handleTestCondition}
            />
          )}

          {currentStep === 3 && (
            <div className="space-y-4">
              <div className="space-y-2">
                <Label>Check Frequency</Label>
                <Select
                  value={formData.schedule}
                  onValueChange={(v) => setFormData({ ...formData, schedule: v })}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {SCHEDULE_OPTIONS.map((opt) => (
                      <SelectItem key={opt.value} value={opt.value}>
                        {opt.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label>Cooldown Period</Label>
                <Select
                  value={formData.cooldownMinutes.toString()}
                  onValueChange={(v) => setFormData({ ...formData, cooldownMinutes: parseInt(v) })}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {COOLDOWN_OPTIONS.map((opt) => (
                      <SelectItem key={opt.value} value={opt.value.toString()}>
                        {opt.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <p className="text-sm text-gray-500">
                  Time to wait before sending another notification
                </p>
              </div>
            </div>
          )}

          {currentStep === 4 && (
            <NotificationConfig
              channels={formData.channels}
              onChange={(channels) => setFormData({ ...formData, channels })}
            />
          )}

          {currentStep === 5 && <AlertReview formData={formData} />}
        </div>

        <DialogFooter className="flex justify-between">
          <Button variant="outline" onClick={handleBack} disabled={currentStep === 0}>
            <ChevronLeft className="h-4 w-4 mr-2" />
            Back
          </Button>
          {currentStep < STEPS.length - 1 ? (
            <Button onClick={handleNext} disabled={!canProceed()}>
              Next
              <ChevronRight className="h-4 w-4 ml-2" />
            </Button>
          ) : (
            <Button onClick={handleSubmit} disabled={isSubmitting}>
              <Save className="h-4 w-4 mr-2" />
              {isSubmitting ? "Creating..." : "Create Alert"}
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
