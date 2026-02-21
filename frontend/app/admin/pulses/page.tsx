"use client";

import { useState, useEffect, useCallback } from "react";
import { PageLayout } from "@/components/page-layout";
import { PageHeader, PageActions, PageContent } from "@/components/page-header";
import { Button } from "@/components/ui/button";
import { Activity, Plus, RefreshCw } from "lucide-react";
import { Pulse } from "@/types/pulses";
import { pulseService } from "@/services/pulse-service";
import { PulseDialog } from "@/components/pulse/pulse-dialog";
import { PulseList } from "@/components/pulse/pulse-list";
import { toast } from "sonner";

export default function PulsesPage() {
  const [pulses, setPulses] = useState<Pulse[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [pulseToEdit, setPulseToEdit] = useState<Pulse | undefined>(undefined);

  const fetchPulses = useCallback(async () => {
    setIsLoading(true);
    try {
      const data = await pulseService.getPulses();
      setPulses(data);
    } catch (error: any) {
      toast.error("Failed to load pulses");
      console.error(error);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchPulses();
  }, [fetchPulses]);

  const handleCreate = () => {
    setPulseToEdit(undefined);
    setIsDialogOpen(true);
  };

  const handleEdit = (pulse: Pulse) => {
    setPulseToEdit(pulse);
    setIsDialogOpen(true);
  };

  return (
    <PageLayout>
      <PageHeader
        title="Pulses"
        description="Schedule automated dashboard screenshots to Slack and Teams."
        icon={Activity}
        actions={
          <PageActions>
            <Button variant="outline" size="sm" onClick={fetchPulses} className="mr-2">
              <RefreshCw className="mr-2 h-4 w-4" /> Refresh
            </Button>
            <Button onClick={handleCreate}>
              <Plus className="mr-2 h-4 w-4" /> New Pulse
            </Button>
          </PageActions>
        }
      />

      <PageContent>
        <PulseList
          pulses={pulses}
          isLoading={isLoading}
          onEdit={handleEdit}
          onRefresh={fetchPulses}
        />
      </PageContent>

      <PulseDialog
        open={isDialogOpen}
        onOpenChange={setIsDialogOpen}
        onSuccess={fetchPulses}
        pulseToEdit={pulseToEdit}
      />
    </PageLayout>
  );
}
