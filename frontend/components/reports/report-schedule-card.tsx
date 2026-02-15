'use client';

import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
} from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  Play,
  Pause,
  Edit,
  Trash2,
  History,
  MoreVertical,
  Clock,
  Calendar,
  FileText,
  CheckCircle2,
  XCircle,
  AlertCircle,
} from 'lucide-react';
import type { ScheduledReportResponse } from '@/types/scheduled-reports';

interface ReportScheduleCardProps {
  report: ScheduledReportResponse;
  onEdit: (report: ScheduledReportResponse) => void;
  onDelete: (report: ScheduledReportResponse) => void;
  onToggle: (report: ScheduledReportResponse) => void;
  onRunNow: (report: ScheduledReportResponse) => void;
  onViewHistory: (report: ScheduledReportResponse) => void;
}

export function ReportScheduleCard({
  report,
  onEdit,
  onDelete,
  onToggle,
  onRunNow,
  onViewHistory,
}: ReportScheduleCardProps) {
  const formatSchedule = () => {
    switch (report.scheduleType) {
      case 'daily':
        return `Daily at ${report.timeOfDay || '09:00'}`;
      case 'weekly':
        const days = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
        return `Weekly on ${days[report.dayOfWeek || 1]} at ${report.timeOfDay || '09:00'}`;
      case 'monthly':
        return `Monthly on day ${report.dayOfMonth || 1} at ${report.timeOfDay || '09:00'}`;
      case 'cron':
        return `Custom: ${report.cronExpr}`;
      default:
        return report.scheduleType;
    }
  };

  const formatNextRun = () => {
    if (!report.nextRunAt) return 'Not scheduled';
    const date = new Date(report.nextRunAt);
    return date.toLocaleString('en-US', {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const getStatusIcon = () => {
    if (!report.isActive) {
      return <Pause className="w-4 h-4" />;
    }
    if (report.lastRunStatus === 'failed') {
      return <XCircle className="w-4 h-4 text-destructive" />;
    }
    if (report.lastRunStatus === 'success') {
      return <CheckCircle2 className="w-4 h-4 text-green-500" />;
    }
    return <Clock className="w-4 h-4 text-primary" />;
  };

  return (
    <Card className="relative group">
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="flex-1 min-w-0 pr-2">
            <div className="flex items-center gap-2 mb-1">
              <h3 className="font-semibold truncate">{report.name}</h3>
              <Badge variant={report.isActive ? 'default' : 'secondary'} className="text-[10px]">
                {report.isActive ? 'Active' : 'Paused'}
              </Badge>
            </div>
            {report.description && (
              <p className="text-xs text-muted-foreground line-clamp-1">
                {report.description}
              </p>
            )}
          </div>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon" className="h-8 w-8">
                <MoreVertical className="w-4 h-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={() => onRunNow(report)}>
                <Play className="w-4 h-4 mr-2" />
                Run Now
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => onToggle(report)}>
                {report.isActive ? (
                  <>
                    <Pause className="w-4 h-4 mr-2" />
                    Pause
                  </>
                ) : (
                  <>
                    <Play className="w-4 h-4 mr-2" />
                    Activate
                  </>
                )}
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => onViewHistory(report)}>
                <History className="w-4 h-4 mr-2" />
                View History
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={() => onEdit(report)}>
                <Edit className="w-4 h-4 mr-2" />
                Edit
              </DropdownMenuItem>
              <DropdownMenuItem
                onClick={() => onDelete(report)}
                className="text-destructive focus:text-destructive"
              >
                <Trash2 className="w-4 h-4 mr-2" />
                Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </CardHeader>

      <CardContent className="pb-3">
        <div className="space-y-3">
          {/* Schedule Info */}
          <div className="flex items-center gap-2 text-sm">
            <Calendar className="w-4 h-4 text-muted-foreground" />
            <span className="text-muted-foreground">{formatSchedule()}</span>
          </div>

          {/* Next Run */}
          {report.isActive && report.nextRunAt && (
            <div className="flex items-center gap-2 text-sm">
              <Clock className="w-4 h-4 text-muted-foreground" />
              <span className="text-muted-foreground">Next run: {formatNextRun()}</span>
            </div>
          )}

          {/* Recipients */}
          <div className="flex items-center gap-2">
            <div className="flex -space-x-2">
              {report.recipients.slice(0, 3).map((recipient, i) => (
                <div
                  key={i}
                  className="w-6 h-6 rounded-full bg-primary/10 border-2 border-background flex items-center justify-center text-[10px] font-medium"
                  title={recipient.email}
                >
                  {recipient.email[0].toUpperCase()}
                </div>
              ))}
              {report.recipients.length > 3 && (
                <div className="w-6 h-6 rounded-full bg-muted border-2 border-background flex items-center justify-center text-[10px]">
                  +{report.recipients.length - 3}
                </div>
              )}
            </div>
            <span className="text-xs text-muted-foreground">
              {report.recipients.length} recipient{report.recipients.length !== 1 ? 's' : ''}
            </span>
          </div>

          {/* Format */}
          <div className="flex items-center gap-2">
            <FileText className="w-4 h-4 text-muted-foreground" />
            <Badge variant="outline" className="text-[10px] uppercase">
              {report.format}
            </Badge>
            <Badge variant="outline" className="text-[10px]">
              {report.resourceType}
            </Badge>
          </div>
        </div>
      </CardContent>

      <CardFooter className="pt-3 border-t">
        <div className="flex items-center justify-between w-full text-xs">
          <div className="flex items-center gap-2">
            {getStatusIcon()}
            <span className="text-muted-foreground">
              {report.lastRunAt
                ? `Last run: ${new Date(report.lastRunAt).toLocaleDateString()}`
                : 'Never run'}
            </span>
          </div>
          <div className="flex items-center gap-1 text-muted-foreground">
            <CheckCircle2 className="w-3 h-3 text-green-500" />
            {report.successCount}
            <span className="mx-1">/</span>
            <XCircle className="w-3 h-3 text-destructive" />
            {report.failureCount}
          </div>
        </div>
      </CardFooter>
    </Card>
  );
}
