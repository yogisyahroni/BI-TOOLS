'use client';

import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import {
  CheckCircle2,
  XCircle,
  Loader2,
  Clock,
  Download,
  AlertCircle,
} from 'lucide-react';
import { scheduledReportsApi } from '@/lib/api/scheduled-reports';
import type { ScheduledReportRun } from '@/types/scheduled-reports';
import { toast } from 'sonner';

interface ReportHistoryProps {
  reportId: string;
}

export function ReportHistory({ reportId }: ReportHistoryProps) {
  const [runs, setRuns] = useState<ScheduledReportRun[]>([]);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(false);

  const fetchHistory = async (pageNum: number = 1) => {
    try {
      setLoading(true);
      const response = await scheduledReportsApi.getHistory(reportId, {
        page: pageNum,
        limit: 10,
        orderBy: 'started_at DESC',
      });

      if (pageNum === 1) {
        setRuns(response.runs);
      } else {
        setRuns((prev) => [...prev, ...response.runs]);
      }

      setHasMore(response.runs.length === 10 && response.total > pageNum * 10);
    } catch (error) {
      toast.error('Failed to load report history');
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchHistory(1);
        // eslint-disable-next-line react-hooks/exhaustive-deps
        // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [reportId]);

  const handleLoadMore = () => {
    const nextPage = page + 1;
    setPage(nextPage);
    fetchHistory(nextPage);
  };

  const handleDownload = async (run: ScheduledReportRun) => {
    if (!run.fileUrl) {
      toast.error('No file available for download');
      return;
    }

    try {
      const { downloadUrl } = await scheduledReportsApi.getDownloadUrl(run.id);
      window.open(downloadUrl, '_blank');
    } catch (error) {
      toast.error('Failed to get download URL');
      console.error(error);
    }
  };

  const formatDuration = (ms?: number) => {
    if (!ms) return '-';
    if (ms < 1000) return `${ms}ms`;
    if (ms < 60000) return `${Math.round(ms / 1000)}s`;
    return `${Math.round(ms / 60000)}m ${Math.round((ms % 60000) / 1000)}s`;
  };

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr);
    return date.toLocaleString('en-US', {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'success':
        return (
          <Badge variant="outline" className="text-green-500 border-green-200 bg-green-50">
            <CheckCircle2 className="w-3 h-3 mr-1" />
            Success
          </Badge>
        );
      case 'failed':
        return (
          <Badge variant="outline" className="text-destructive border-destructive/20 bg-destructive/10">
            <XCircle className="w-3 h-3 mr-1" />
            Failed
          </Badge>
        );
      case 'running':
        return (
          <Badge variant="outline" className="text-blue-500 border-blue-200 bg-blue-50">
            <Loader2 className="w-3 h-3 mr-1 animate-spin" />
            Running
          </Badge>
        );
      case 'pending':
        return (
          <Badge variant="outline" className="text-amber-500 border-amber-200 bg-amber-50">
            <Clock className="w-3 h-3 mr-1" />
            Pending
          </Badge>
        );
      default:
        return (
          <Badge variant="outline">
            {status}
          </Badge>
        );
    }
  };

  if (loading && runs.length === 0) {
    return (
      <div className="space-y-3">
        {[...Array(5)].map((_, i) => (
          <Skeleton key={i} className="h-12 w-full" />
        ))}
      </div>
    );
  }

  if (runs.length === 0) {
    return (
      <div className="text-center py-12">
        <Clock className="w-12 h-12 text-muted-foreground mx-auto mb-4" />
        <h3 className="text-lg font-medium mb-2">No History Yet</h3>
        <p className="text-muted-foreground text-sm">
          This report hasn&apos;t been run yet. Trigger it manually or wait for the next scheduled run.
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Status</TableHead>
            <TableHead>Started</TableHead>
            <TableHead>Duration</TableHead>
            <TableHead>Triggered By</TableHead>
            <TableHead className="text-right">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {runs.map((run) => (
            <TableRow key={run.id}>
              <TableCell>
                <div className="space-y-1">
                  {getStatusBadge(run.status)}
                  {run.errorMessage && (
                    <p className="text-xs text-destructive flex items-center gap-1">
                      <AlertCircle className="w-3 h-3" />
                      {run.errorMessage}
                    </p>
                  )}
                </div>
              </TableCell>
              <TableCell className="text-sm">
                {formatDate(run.startedAt)}
              </TableCell>
              <TableCell className="text-sm">
                {formatDuration(run.durationMs)}
              </TableCell>
              <TableCell className="text-sm">
                <Badge variant="secondary" className="text-[10px]">
                  {run.triggeredBy || 'schedule'}
                </Badge>
              </TableCell>
              <TableCell className="text-right">
                {run.status === 'success' && run.fileUrl && (
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => handleDownload(run)}
                  >
                    <Download className="w-4 h-4" />
                  </Button>
                )}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>

      {hasMore && (
        <div className="flex justify-center pt-4">
          <Button
            variant="outline"
            onClick={handleLoadMore}
            disabled={loading}
          >
            {loading ? (
              <>
                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                Loading...
              </>
            ) : (
              'Load More'
            )}
          </Button>
        </div>
      )}
    </div>
  );
}
