"use client";

import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  AlertCircle,
  AlertTriangle,
  Info,
  Bell,
  CheckCircle2,
  VolumeX,
  Volume2,
  MoreHorizontal,
  Search,
  Clock,
  Play,
  Edit,
  Trash2,
} from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { formatDistanceToNow } from "date-fns";
import type { Alert, AlertSeverity, AlertState } from "@/types/alerts";
import { ALERT_SEVERITIES, ALERT_STATES } from "@/types/alerts";

interface AlertListProps {
  alerts: Alert[];
  loading?: boolean;
  onEdit?: (alert: Alert) => void;
  onDelete?: (alert: Alert) => void;
  onAcknowledge?: (alert: Alert) => void;
  onMute?: (alert: Alert, duration?: number) => void;
  onUnmute?: (alert: Alert) => void;
  onTriggerCheck?: (alert: Alert) => void;
}

export function AlertList({
  alerts,
  loading,
  onEdit,
  onDelete,
  onAcknowledge,
  onMute,
  onUnmute,
  onTriggerCheck,
}: AlertListProps) {
  const [searchQuery, setSearchQuery] = useState("");
  const [severityFilter, setSeverityFilter] = useState<AlertSeverity | "all">("all");
  const [stateFilter, setStateFilter] = useState<AlertState | "all">("all");

  const filteredAlerts = alerts.filter((alert) => {
    const matchesSearch =
      alert.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      alert.description?.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesSeverity = severityFilter === "all" || alert.severity === severityFilter;
    const matchesState = stateFilter === "all" || alert.state === stateFilter;
    return matchesSearch && matchesSeverity && matchesState;
  });

  const getSeverityIcon = (severity: AlertSeverity) => {
    switch (severity) {
      case "critical":
        return <AlertCircle className="h-4 w-4 text-red-500" />;
      case "warning":
        return <AlertTriangle className="h-4 w-4 text-amber-500" />;
      case "info":
        return <Info className="h-4 w-4 text-blue-500" />;
      default:
        return <Bell className="h-4 w-4" />;
    }
  };

  const getSeverityColor = (severity: AlertSeverity) => {
    switch (severity) {
      case "critical":
        return "bg-red-100 text-red-800 border-red-200";
      case "warning":
        return "bg-amber-100 text-amber-800 border-amber-200";
      case "info":
        return "bg-blue-100 text-blue-800 border-blue-200";
      default:
        return "bg-gray-100 text-gray-800 border-gray-200";
    }
  };

  const getStateColor = (state: AlertState) => {
    switch (state) {
      case "ok":
        return "bg-green-100 text-green-800 border-green-200";
      case "triggered":
        return "bg-red-100 text-red-800 border-red-200";
      case "acknowledged":
        return "bg-indigo-100 text-indigo-800 border-indigo-200";
      case "muted":
        return "bg-gray-100 text-gray-800 border-gray-200";
      case "error":
        return "bg-purple-100 text-purple-800 border-purple-200";
      default:
        return "bg-gray-100 text-gray-800 border-gray-200";
    }
  };

  const getStateIcon = (state: AlertState) => {
    switch (state) {
      case "ok":
        return <CheckCircle2 className="h-4 w-4 text-green-500" />;
      case "triggered":
        return <AlertCircle className="h-4 w-4 text-red-500" />;
      case "muted":
        return <VolumeX className="h-4 w-4 text-gray-500" />;
      default:
        return <Bell className="h-4 w-4" />;
    }
  };

  if (loading) {
    return (
      <div className="space-y-4">
        {[1, 2, 3].map((i) => (
          <Card key={i} className="animate-pulse">
            <CardContent className="p-6">
              <div className="h-4 bg-gray-200 rounded w-1/4 mb-4" />
              <div className="h-3 bg-gray-200 rounded w-3/4" />
            </CardContent>
          </Card>
        ))}
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Filters */}
      <div className="flex flex-wrap gap-4">
        <div className="relative flex-1 min-w-[200px]">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
          <Input
            placeholder="Search alerts..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-10"
          />
        </div>
        <Select
          value={severityFilter}
          onValueChange={(v) => setSeverityFilter(v as AlertSeverity | "all")}
        >
          <SelectTrigger className="w-[150px]">
            <SelectValue placeholder="Severity" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Severities</SelectItem>
            {ALERT_SEVERITIES.map((s) => (
              <SelectItem key={s.value} value={s.value}>
                {s.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Select value={stateFilter} onValueChange={(v) => setStateFilter(v as AlertState | "all")}>
          <SelectTrigger className="w-[150px]">
            <SelectValue placeholder="State" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All States</SelectItem>
            {ALERT_STATES.map((s) => (
              <SelectItem key={s.value} value={s.value}>
                {s.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {/* Alert Cards */}
      {filteredAlerts.length === 0 ? (
        <Card>
          <CardContent className="py-12 text-center">
            <Bell className="h-12 w-12 text-gray-300 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-900">No alerts found</h3>
            <p className="text-gray-500 mt-1">
              {alerts.length === 0
                ? "Create your first alert to start monitoring your data."
                : "No alerts match your current filters."}
            </p>
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-3">
          {filteredAlerts.map((alert) => (
            <Card
              key={alert.id}
              className={`hover:shadow-md transition-shadow ${
                alert.state === "triggered" ? "border-red-300" : ""
              }`}
            >
              <CardContent className="p-4">
                <div className="flex items-start justify-between gap-4">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 flex-wrap">
                      <h3 className="font-semibold text-gray-900 truncate">{alert.name}</h3>
                      <Badge variant="outline" className={getSeverityColor(alert.severity)}>
                        <span className="flex items-center gap-1">
                          {getSeverityIcon(alert.severity)}
                          {alert.severity}
                        </span>
                      </Badge>
                      <Badge variant="outline" className={getStateColor(alert.state)}>
                        <span className="flex items-center gap-1">
                          {getStateIcon(alert.state)}
                          {alert.state}
                        </span>
                      </Badge>
                      {alert.isMuted && (
                        <Badge variant="outline" className="bg-gray-100">
                          <VolumeX className="h-3 w-3 mr-1" />
                          Muted
                        </Badge>
                      )}
                    </div>
                    {alert.description && (
                      <p className="text-sm text-gray-500 mt-1 truncate">{alert.description}</p>
                    )}
                    <div className="flex items-center gap-4 mt-2 text-sm text-gray-500">
                      <span className="flex items-center gap-1">
                        <Clock className="h-3 w-3" />
                        {alert.lastRunAt
                          ? `Last run ${formatDistanceToNow(new Date(alert.lastRunAt), { addSuffix: true })}`
                          : "Never run"}
                      </span>
                      {alert.query && (
                        <span className="text-gray-400">Query: {alert.query.name}</span>
                      )}
                      <span className="font-mono text-xs">
                        {alert.column} {alert.operator} {alert.threshold}
                      </span>
                    </div>
                  </div>

                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button variant="ghost" size="sm">
                        <MoreHorizontal className="h-4 w-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      {alert.state === "triggered" && (
                        <DropdownMenuItem onClick={() => onAcknowledge?.(alert)}>
                          <CheckCircle2 className="h-4 w-4 mr-2" />
                          Acknowledge
                        </DropdownMenuItem>
                      )}
                      {alert.isMuted ? (
                        <DropdownMenuItem onClick={() => onUnmute?.(alert)}>
                          <Volume2 className="h-4 w-4 mr-2" />
                          Unmute
                        </DropdownMenuItem>
                      ) : (
                        <DropdownMenuItem onClick={() => onMute?.(alert)}>
                          <VolumeX className="h-4 w-4 mr-2" />
                          Mute
                        </DropdownMenuItem>
                      )}
                      <DropdownMenuItem onClick={() => onTriggerCheck?.(alert)}>
                        <Play className="h-4 w-4 mr-2" />
                        Run Check Now
                      </DropdownMenuItem>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem onClick={() => onEdit?.(alert)}>
                        <Edit className="h-4 w-4 mr-2" />
                        Edit
                      </DropdownMenuItem>
                      <DropdownMenuItem onClick={() => onDelete?.(alert)} className="text-red-600">
                        <Trash2 className="h-4 w-4 mr-2" />
                        Delete
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
