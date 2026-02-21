"use client";

import { useState, useEffect } from "react";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Clock, Calendar, AlertCircle } from "lucide-react";
import type { ReportScheduleType, DayOfWeek } from "@/types/scheduled-reports";
import { SCHEDULE_TYPES, DAYS_OF_WEEK } from "@/types/scheduled-reports";

interface SchedulePickerProps {
  scheduleType: ReportScheduleType;
  cronExpr?: string;
  timeOfDay?: string;
  dayOfWeek?: number;
  dayOfMonth?: number;
  timezone?: string;
  onChange: (values: {
    scheduleType: ReportScheduleType;
    cronExpr?: string;
    timeOfDay?: string;
    dayOfWeek?: number;
    dayOfMonth?: number;
    timezone?: string;
  }) => void;
  disabled?: boolean;
  error?: string;
}

export function SchedulePicker({
  scheduleType,
  cronExpr = "0 9 * * *",
  timeOfDay = "09:00",
  dayOfWeek = 1,
  dayOfMonth = 1,
  timezone = "UTC",
  onChange,
  disabled = false,
  error,
}: SchedulePickerProps) {
  const [cronError, setCronError] = useState("");

  // Validate cron expression
  const validateCron = (expression: string): boolean => {
    // Basic cron validation (5 fields: minute hour day month dayOfWeek)
    const parts = expression.trim().split(/\s+/);
    if (parts.length !== 5) {
      setCronError("Cron expression must have 5 fields (minute hour day month dayOfWeek)");
      return false;
    }
    setCronError("");
    return true;
  };

  const handleTypeChange = (type: ReportScheduleType) => {
    onChange({
      scheduleType: type,
      cronExpr,
      timeOfDay,
      dayOfWeek,
      dayOfMonth,
      timezone,
    });
  };

  const handleCronChange = (value: string) => {
    validateCron(value);
    onChange({
      scheduleType,
      cronExpr: value,
      timeOfDay,
      dayOfWeek,
      dayOfMonth,
      timezone,
    });
  };

  const handleTimeChange = (value: string) => {
    onChange({
      scheduleType,
      cronExpr,
      timeOfDay: value,
      dayOfWeek,
      dayOfMonth,
      timezone,
    });
  };

  const handleDayOfWeekChange = (value: string) => {
    onChange({
      scheduleType,
      cronExpr,
      timeOfDay,
      dayOfWeek: parseInt(value, 10),
      dayOfMonth,
      timezone,
    });
  };

  const handleDayOfMonthChange = (value: string) => {
    const day = parseInt(value, 10);
    if (day >= 1 && day <= 31) {
      onChange({
        scheduleType,
        cronExpr,
        timeOfDay,
        dayOfWeek,
        dayOfMonth: day,
        timezone,
      });
    }
  };

  // Generate human-readable description
  const getScheduleDescription = () => {
    switch (scheduleType) {
      case "daily":
        return `Every day at ${timeOfDay}`;
      case "weekly":
        const dayName = DAYS_OF_WEEK.find((d) => d.value === dayOfWeek)?.label || "Monday";
        return `Every ${dayName} at ${timeOfDay}`;
      case "monthly":
        return `On the ${dayOfMonth}${getDaySuffix(dayOfMonth)} of every month at ${timeOfDay}`;
      case "cron":
        return `Custom schedule: ${cronExpr}`;
      default:
        return "";
    }
  };

  const getDaySuffix = (day: number): string => {
    if (day > 3 && day < 21) return "th";
    switch (day % 10) {
      case 1:
        return "st";
      case 2:
        return "nd";
      case 3:
        return "rd";
      default:
        return "th";
    }
  };

  return (
    <div className="space-y-6">
      {/* Schedule Type Selection */}
      <div className="space-y-3">
        <Label>Schedule Type</Label>
        <div className="grid grid-cols-2 gap-3">
          {SCHEDULE_TYPES.map((type) => (
            <button
              key={type.value}
              type="button"
              onClick={() => handleTypeChange(type.value)}
              disabled={disabled}
              className={`p-4 rounded-lg border-2 text-left transition-all ${
                scheduleType === type.value
                  ? "border-primary bg-primary/5"
                  : "border-muted hover:border-muted-foreground/20"
              }`}
            >
              <div className="font-medium">{type.label}</div>
              <div className="text-xs text-muted-foreground mt-1">{type.description}</div>
            </button>
          ))}
        </div>
      </div>

      {/* Schedule Configuration */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-base">
            <Clock className="w-4 h-4" />
            Configuration
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Cron Expression */}
          {scheduleType === "cron" && (
            <div className="space-y-2">
              <Label htmlFor="cronExpr">Cron Expression *</Label>
              <Input
                id="cronExpr"
                value={cronExpr}
                onChange={(e) => handleCronChange(e.target.value)}
                placeholder="0 9 * * *"
                disabled={disabled}
              />
              <p className="text-xs text-muted-foreground">
                Format: minute hour day month dayOfWeek (e.g., &quot;0 9 * * *&quot; for daily at 9
                AM)
              </p>
              {cronError && (
                <div className="flex items-center gap-2 text-sm text-destructive">
                  <AlertCircle className="w-4 h-4" />
                  {cronError}
                </div>
              )}
            </div>
          )}

          {/* Time of Day */}
          {scheduleType !== "cron" && (
            <div className="space-y-2">
              <Label htmlFor="timeOfDay">Time of Day *</Label>
              <Input
                id="timeOfDay"
                type="time"
                value={timeOfDay}
                onChange={(e) => handleTimeChange(e.target.value)}
                disabled={disabled}
              />
            </div>
          )}

          {/* Day of Week */}
          {scheduleType === "weekly" && (
            <div className="space-y-2">
              <Label>Day of Week *</Label>
              <Select
                value={String(dayOfWeek)}
                onValueChange={handleDayOfWeekChange}
                disabled={disabled}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {DAYS_OF_WEEK.map((day) => (
                    <SelectItem key={day.value} value={String(day.value)}>
                      {day.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          )}

          {/* Day of Month */}
          {scheduleType === "monthly" && (
            <div className="space-y-2">
              <Label htmlFor="dayOfMonth">Day of Month *</Label>
              <Input
                id="dayOfMonth"
                type="number"
                min={1}
                max={31}
                value={dayOfMonth}
                onChange={(e) => handleDayOfMonthChange(e.target.value)}
                disabled={disabled}
              />
              <p className="text-xs text-muted-foreground">Enter a day between 1 and 31</p>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Summary */}
      <div className="flex items-center gap-2 p-4 bg-muted rounded-lg">
        <Calendar className="w-5 h-5 text-muted-foreground" />
        <div className="flex-1">
          <div className="text-sm font-medium">Schedule Summary</div>
          <div className="text-sm text-muted-foreground">{getScheduleDescription()}</div>
        </div>
        <Badge variant="outline" className="text-[10px] uppercase">
          {scheduleType}
        </Badge>
      </div>

      {error && (
        <div className="flex items-center gap-2 text-sm text-destructive">
          <AlertCircle className="w-4 h-4" />
          {error}
        </div>
      )}
    </div>
  );
}
