"use client";

import { useState } from "react";
import { format } from "date-fns";
import { AlertTriangle, RotateCcw, X, Loader2, CheckCircle2, Eye } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";
import type {
  DashboardVersion,
  QueryVersion,
  VersionResourceType,
  RestoreVersionResponse,
} from "@/types/versions";
import { restoreDashboardVersion, restoreQueryVersion } from "@/lib/api/versions";

interface VersionRestoreDialogProps {
  isOpen: boolean;
  onClose: () => void;
  version: DashboardVersion | QueryVersion | null;
  resourceType: VersionResourceType;
  resourceName: string;
  onRestored?: () => void;
}

export function VersionRestoreDialog({
  isOpen,
  onClose,
  version,
  resourceType,
  resourceName,
  onRestored,
}: VersionRestoreDialogProps) {
  const [isRestoring, setIsRestoring] = useState(false);
  const [isRestored, setIsRestored] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [restoreResult, setRestoreResult] = useState<RestoreVersionResponse | null>(null);

  if (!version) return null;

  const isDashboard = resourceType === "dashboard";
  const createdAt = new Date(version.createdAt);

  // Get initials for avatar
  const getInitials = (name: string) => {
    return name
      .split(" ")
      .map((n) => n[0])
      .join("")
      .toUpperCase()
      .slice(0, 2);
  };

  const handleRestore = async () => {
    setIsRestoring(true);
    setError(null);

    try {
      let result: RestoreVersionResponse;

      if (isDashboard) {
        result = await restoreDashboardVersion(version.id);
      } else {
        result = await restoreQueryVersion(version.id);
      }

      setRestoreResult(result);
      setIsRestored(true);
      onRestored?.();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to restore version");
    } finally {
      setIsRestoring(false);
    }
  };

  const handleClose = () => {
    // Reset state when closing
    setIsRestored(false);
    setRestoreResult(null);
    setError(null);
    onClose();
  };

  // Success state
  if (isRestored && restoreResult) {
    return (
      <Dialog open={isOpen} onOpenChange={handleClose}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2 text-green-600">
              <CheckCircle2 className="h-6 w-6" />
              Version Restored Successfully
            </DialogTitle>
          </DialogHeader>

          <div className="py-6 text-center space-y-4">
            <div className="w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mx-auto">
              <RotateCcw className="h-8 w-8 text-green-600" />
            </div>

            <div>
              <p className="text-lg font-medium text-foreground">{resourceName}</p>
              <p className="text-sm text-muted-foreground mt-1">
                Restored to Version {version.version}
              </p>
            </div>

            <div className="text-sm text-muted-foreground bg-muted p-3 rounded-lg">
              <p>Created by: {version.createdByUser?.name || "Unknown"}</p>
              <p>Original date: {format(createdAt, "PPp")}</p>
            </div>
          </div>

          <DialogFooter>
            <Button onClick={handleClose} className="w-full">
              Done
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    );
  }

  return (
    <Dialog open={isOpen} onOpenChange={handleClose}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <AlertTriangle className="h-5 w-5 text-yellow-600" />
            Restore Version
          </DialogTitle>
          <DialogDescription>
            Are you sure you want to restore <strong>{resourceName}</strong> to this version?
          </DialogDescription>
        </DialogHeader>

        {/* Warning */}
        <Alert variant="destructive" className="bg-red-50 border-red-200">
          <AlertTriangle className="h-4 w-4 text-red-600" />
          <AlertTitle className="text-red-800">Warning</AlertTitle>
          <AlertDescription className="text-red-700">
            This will replace your current {resourceType} with the selected version. Any unsaved
            changes will be lost. This action cannot be undone.
          </AlertDescription>
        </Alert>

        {/* Version preview */}
        <div className="border rounded-lg p-4 space-y-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <Avatar className="h-10 w-10">
                <AvatarImage src={version.createdByUser?.avatar || version.createdByUser?.image} />
                <AvatarFallback>
                  {version.createdByUser?.name ? getInitials(version.createdByUser.name) : "U"}
                </AvatarFallback>
              </Avatar>
              <div>
                <p className="font-medium">Version {version.version}</p>
                <p className="text-sm text-muted-foreground">{format(createdAt, "PPp")}</p>
              </div>
            </div>

            {version.isAutoSave && <Badge variant="secondary">Auto-save</Badge>}
          </div>

          <Separator />

          {/* Change summary */}
          {version.changeSummary && (
            <div>
              <p className="text-sm font-medium text-muted-foreground mb-1">Change Summary</p>
              <p className="text-sm">{version.changeSummary}</p>
            </div>
          )}

          {/* Version details */}
          <div className="grid grid-cols-2 gap-4 text-sm">
            <div>
              <p className="text-muted-foreground">Name</p>
              <p className="font-medium truncate">{version.name}</p>
            </div>
            <div>
              <p className="text-muted-foreground">Created by</p>
              <p className="font-medium">{version.createdByUser?.name || "Unknown"}</p>
            </div>
          </div>

          {/* Preview button */}
          <Button variant="outline" className="w-full" onClick={() => {}}>
            <Eye className="h-4 w-4 mr-2" />
            Preview This Version
          </Button>
        </div>

        {/* Error */}
        {error && (
          <Alert variant="destructive">
            <AlertTriangle className="h-4 w-4" />
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        <DialogFooter className="gap-2">
          <Button variant="outline" onClick={handleClose} disabled={isRestoring}>
            Cancel
          </Button>
          <Button onClick={handleRestore} disabled={isRestoring} variant="destructive">
            {isRestoring ? (
              <>
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                Restoring...
              </>
            ) : (
              <>
                <RotateCcw className="h-4 w-4 mr-2" />
                Restore Version {version.version}
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
