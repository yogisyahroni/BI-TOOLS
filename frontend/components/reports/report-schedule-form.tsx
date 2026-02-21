"use client";

import { useState, useEffect } from "react";
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
import { Switch } from "@/components/ui/switch";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Calendar, Clock, Mail, FileText, Settings, ChevronRight } from "lucide-react";
import type {
  ScheduleFormData,
  ReportResourceType,
  ReportFormat,
  CreateScheduledReportRequest,
} from "@/types/scheduled-reports";
import {
  REPORT_FORMATS,
  RESOURCE_TYPES,
  _SCHEDULE_TYPES,
  _DAYS_OF_WEEK,
} from "@/types/scheduled-reports";
import { RecipientManager } from "./recipient-manager";
import { SchedulePicker } from "./schedule-picker";
import { scheduledReportsApi } from "@/lib/api/scheduled-reports";

interface ReportScheduleFormProps {
  initialData?: Partial<ScheduleFormData>;
  onSubmit: (data: CreateScheduledReportRequest) => Promise<void>;
  onCancel: () => void;
  isSubmitting?: boolean;
}

const defaultFormData: ScheduleFormData = {
  name: "",
  description: "",
  resourceType: "dashboard",
  resourceId: "",
  resourceName: "",
  scheduleType: "daily",
  cronExpr: "0 9 * * *",
  timeOfDay: "09:00",
  dayOfWeek: 1,
  dayOfMonth: 1,
  timezone: "UTC",
  recipients: [],
  format: "pdf",
  includeFilters: false,
  subject: "",
  message: "",
};

export function ReportScheduleForm({
  initialData,
  onSubmit,
  onCancel,
  isSubmitting = false,
}: ReportScheduleFormProps) {
  const [formData, setFormData] = useState<ScheduleFormData>({
    ...defaultFormData,
    ...initialData,
  });
  const [timezones, setTimezones] = useState<{ value: string; label: string }[]>([]);
  const [currentStep, setCurrentStep] = useState(1);

  useEffect(() => {
    // Load timezones
    scheduledReportsApi.getTimezones().then((data) => {
      setTimezones(data.timezones);
    });
  }, []);

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const handleChange = (field: keyof ScheduleFormData, value: any) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    const request: CreateScheduledReportRequest = {
      name: formData.name,
      description: formData.description || undefined,
      resourceType: formData.resourceType,
      resourceId: formData.resourceId,
      scheduleType: formData.scheduleType,
      recipients: formData.recipients,
      format: formData.format,
      includeFilters: formData.includeFilters,
      subject: formData.subject || undefined,
      message: formData.message || undefined,
      timezone: formData.timezone,
    };

    // Add schedule-specific fields
    if (formData.scheduleType === "cron") {
      request.cronExpr = formData.cronExpr;
    } else {
      request.timeOfDay = formData.timeOfDay;
    }

    if (formData.scheduleType === "weekly") {
      request.dayOfWeek = formData.dayOfWeek;
    }

    if (formData.scheduleType === "monthly") {
      request.dayOfMonth = formData.dayOfMonth;
    }

    await onSubmit(request);
  };

  const isStepValid = () => {
    switch (currentStep) {
      case 1:
        return formData.name && formData.resourceId;
      case 2:
        return formData.timeOfDay || (formData.scheduleType === "cron" && formData.cronExpr);
      case 3:
        return formData.recipients.length > 0;
      default:
        return true;
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {/* Progress Steps */}
      <div className="flex items-center gap-2 mb-6">
        {[1, 2, 3, 4].map((step) => (
          <div
            key={step}
            className={`flex-1 h-2 rounded-full ${step <= currentStep ? "bg-primary" : "bg-muted"}`}
          />
        ))}
      </div>

      {/* Step 1: Resource Selection */}
      {currentStep === 1 && (
        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <FileText className="w-5 h-5" />
                Report Details
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="name">Report Name *</Label>
                <Input
                  id="name"
                  value={formData.name}
                  onChange={(e) => handleChange("name", e.target.value)}
                  placeholder="e.g., Weekly Sales Report"
                  required
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="description">Description</Label>
                <Textarea
                  id="description"
                  value={formData.description}
                  onChange={(e) => handleChange("description", e.target.value)}
                  placeholder="Brief description of this report"
                  rows={3}
                />
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Settings className="w-5 h-5" />
                Source
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label>Resource Type</Label>
                <div className="flex gap-2">
                  {RESOURCE_TYPES.map((type) => (
                    <Button
                      key={type.value}
                      type="button"
                      variant={formData.resourceType === type.value ? "default" : "outline"}
                      onClick={() => handleChange("resourceType", type.value)}
                      className="flex-1"
                    >
                      {type.label}
                    </Button>
                  ))}
                </div>
              </div>

              <div className="space-y-2">
                <Label htmlFor="resourceId">
                  {formData.resourceType === "dashboard" ? "Dashboard ID" : "Query ID"} *
                </Label>
                <Input
                  id="resourceId"
                  value={formData.resourceId}
                  onChange={(e) => handleChange("resourceId", e.target.value)}
                  placeholder={`Enter ${formData.resourceType} ID`}
                  required
                />
              </div>

              <div className="space-y-2">
                <Label>Format</Label>
                <div className="flex gap-2 flex-wrap">
                  {REPORT_FORMATS.map((fmt) => (
                    <Button
                      key={fmt.value}
                      type="button"
                      variant={formData.format === fmt.value ? "default" : "outline"}
                      onClick={() => handleChange("format", fmt.value)}
                      size="sm"
                    >
                      {fmt.label}
                    </Button>
                  ))}
                </div>
              </div>

              <div className="flex items-center gap-2">
                <Switch
                  id="includeFilters"
                  checked={formData.includeFilters}
                  onCheckedChange={(checked) => handleChange("includeFilters", checked)}
                />
                <Label htmlFor="includeFilters">Include current filters</Label>
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Step 2: Schedule */}
      {currentStep === 2 && (
        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Clock className="w-5 h-5" />
                Schedule
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-6">
              <SchedulePicker
                scheduleType={formData.scheduleType}
                cronExpr={formData.cronExpr}
                timeOfDay={formData.timeOfDay}
                dayOfWeek={formData.dayOfWeek}
                dayOfMonth={formData.dayOfMonth}
                timezone={formData.timezone}
                onChange={(values) => {
                  setFormData((prev) => ({ ...prev, ...values }));
                }}
              />

              <div className="space-y-2">
                <Label htmlFor="timezone">Timezone</Label>
                <Select
                  value={formData.timezone}
                  onValueChange={(value) => handleChange("timezone", value)}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select timezone" />
                  </SelectTrigger>
                  <SelectContent>
                    {timezones.map((tz) => (
                      <SelectItem key={tz.value} value={tz.value}>
                        {tz.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Step 3: Recipients */}
      {currentStep === 3 && (
        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Mail className="w-5 h-5" />
                Recipients
              </CardTitle>
            </CardHeader>
            <CardContent>
              <RecipientManager
                recipients={formData.recipients}
                onChange={(recipients) => handleChange("recipients", recipients)}
              />
            </CardContent>
          </Card>
        </div>
      )}

      {/* Step 4: Message & Review */}
      {currentStep === 4 && (
        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Mail className="w-5 h-5" />
                Email Content
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="subject">Subject (optional)</Label>
                <Input
                  id="subject"
                  value={formData.subject}
                  onChange={(e) => handleChange("subject", e.target.value)}
                  placeholder={`[Scheduled Report] ${formData.name}`}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="message">Message (optional)</Label>
                <Textarea
                  id="message"
                  value={formData.message}
                  onChange={(e) => handleChange("message", e.target.value)}
                  placeholder="Custom message to include in the email"
                  rows={4}
                />
              </div>
            </CardContent>
          </Card>

          {/* Review */}
          <Card>
            <CardHeader>
              <CardTitle>Review</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-2 gap-4 text-sm">
                <div>
                  <span className="text-muted-foreground">Name:</span>
                  <p className="font-medium">{formData.name}</p>
                </div>
                <div>
                  <span className="text-muted-foreground">Format:</span>
                  <p className="font-medium uppercase">{formData.format}</p>
                </div>
                <div>
                  <span className="text-muted-foreground">Schedule:</span>
                  <p className="font-medium capitalize">{formData.scheduleType}</p>
                </div>
                <div>
                  <span className="text-muted-foreground">Timezone:</span>
                  <p className="font-medium">{formData.timezone}</p>
                </div>
                <div className="col-span-2">
                  <span className="text-muted-foreground">Recipients:</span>
                  <div className="flex flex-wrap gap-2 mt-1">
                    {formData.recipients.map((r, i) => (
                      <Badge key={i} variant="secondary">
                        {r.email}
                      </Badge>
                    ))}
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Navigation */}
      <div className="flex justify-between pt-4 border-t">
        <Button
          type="button"
          variant="outline"
          onClick={currentStep === 1 ? onCancel : () => setCurrentStep(currentStep - 1)}
        >
          {currentStep === 1 ? "Cancel" : "Back"}
        </Button>

        {currentStep < 4 ? (
          <Button
            type="button"
            onClick={() => setCurrentStep(currentStep + 1)}
            disabled={!isStepValid()}
          >
            Next
            <ChevronRight className="w-4 h-4 ml-2" />
          </Button>
        ) : (
          <Button type="submit" disabled={isSubmitting || !isStepValid()}>
            {isSubmitting ? "Creating..." : initialData ? "Update Schedule" : "Create Schedule"}
          </Button>
        )}
      </div>
    </form>
  );
}
