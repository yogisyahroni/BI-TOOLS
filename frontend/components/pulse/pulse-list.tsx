"use client";

import { useState } from "react";
import { Pulse } from "@/types/pulses";
import { pulseService } from "@/services/pulse-service";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Badge } from "@/components/ui/badge";
import { MoreHorizontal, Play, Edit, Trash2, Loader2 } from "lucide-react";
import { toast } from "sonner";
import { formatDistanceToNow } from "date-fns";

interface PulseListProps {
  pulses: Pulse[];
  isLoading: boolean;
  onEdit: (pulse: Pulse) => void;
  onRefresh: () => void;
}

export function PulseList({ pulses, isLoading, onEdit, onRefresh }: PulseListProps) {
  const [triggeringId, setTriggeringId] = useState<string | null>(null);
  const [deletingId, setDeletingId] = useState<string | null>(null);

  const handleTrigger = async (pulse: Pulse) => {
    setTriggeringId(pulse.id);
    try {
      await pulseService.triggerPulse(pulse.id);
      toast.success(`Pulse '${pulse.name}' triggered`);
      onRefresh();
    } catch (error: any) {
      toast.error(error.message || "Failed to trigger pulse");
    } finally {
      setTriggeringId(null);
    }
  };

  const handleDelete = async (pulse: Pulse) => {
    if (!confirm(`Are you sure you want to delete '${pulse.name}'?`)) return;
    setDeletingId(pulse.id);
    try {
      await pulseService.deletePulse(pulse.id);
      toast.success("Pulse deleted");
      onRefresh();
    } catch (error: any) {
      toast.error(error.message || "Failed to delete pulse");
    } finally {
      setDeletingId(null);
    }
  };

  if (isLoading) {
    return (
      <div className="space-y-4 mt-4">
        <Skeleton className="h-10 w-full" />
        <Skeleton className="h-16 w-full" />
        <Skeleton className="h-16 w-full" />
        <Skeleton className="h-16 w-full" />
      </div>
    );
  }

  if (pulses.length === 0) {
    return (
      <div className="text-center p-8 text-muted-foreground border rounded-lg bg-muted/10">
        No pulses found. Create one to get started.
      </div>
    );
  }

  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Name</TableHead>
            <TableHead>Schedule</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Last Run</TableHead>
            <TableHead className="text-right">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {pulses.map((pulse) => (
            <TableRow key={pulse.id}>
              <TableCell className="font-medium">
                <div>{pulse.name}</div>
                <div className="text-xs text-muted-foreground truncate max-w-[200px]">
                  {pulse.webhookUrl ? "Custom Webhook" : "Default Channel"}
                </div>
              </TableCell>
              <TableCell>
                <code className="bg-muted px-1 py-0.5 rounded text-xs">{pulse.schedule}</code>
              </TableCell>
              <TableCell>
                <Badge variant={pulse.isActive ? "default" : "secondary"}>
                  {pulse.isActive ? "Active" : "Paused"}
                </Badge>
              </TableCell>
              <TableCell>
                {pulse.lastRunAt ? (
                  <span title={new Date(pulse.lastRunAt).toLocaleString()}>
                    {formatDistanceToNow(new Date(pulse.lastRunAt), { addSuffix: true })}
                  </span>
                ) : (
                  <span className="text-muted-foreground">-</span>
                )}
              </TableCell>
              <TableCell className="text-right">
                <div className="flex justify-end gap-2">
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => handleTrigger(pulse)}
                    disabled={!!triggeringId}
                  >
                    {triggeringId === pulse.id ? (
                      <Loader2 className="h-4 w-4 animate-spin" />
                    ) : (
                      <Play className="h-4 w-4" />
                    )}
                  </Button>
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button variant="ghost" size="icon">
                        <MoreHorizontal className="h-4 w-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuItem onClick={() => onEdit(pulse)}>
                        <Edit className="mr-2 h-4 w-4" /> Edit
                      </DropdownMenuItem>
                      <DropdownMenuItem
                        className="text-destructive focus:text-destructive"
                        onClick={() => handleDelete(pulse)}
                      >
                        <Trash2 className="mr-2 h-4 w-4" /> Delete
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </div>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}
