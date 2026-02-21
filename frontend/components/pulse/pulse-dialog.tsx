"use client";

import { useState, useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { Pulse, CreatePulseRequest } from "@/types/pulses";
import { pulseService } from "@/services/pulse-service";
import { useDashboards } from "@/hooks/use-dashboards";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { toast } from "sonner";
import { Loader2 } from "lucide-react";

const pulseSchema = z.object({
  name: z.string().min(1, "Name is required"),
  dashboardId: z.string().min(1, "Dashboard is required"),
  schedule: z.string().min(1, "Schedule is required"), // Could add cron validation
  webhookUrl: z.string().url("Invalid URL").optional().or(z.literal("")),
  isActive: z.boolean(),
  width: z.coerce.number().min(800).max(3840),
  height: z.coerce.number().min(600).max(2160),
});

type PulseFormValues = z.infer<typeof pulseSchema>;

interface PulseDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess: () => void;
  pulseToEdit?: Pulse;
}

export function PulseDialog({ open, onOpenChange, onSuccess, pulseToEdit }: PulseDialogProps) {
  const [isLoading, setIsLoading] = useState(false);
  const { dashboards } = useDashboards({ autoFetch: true }); // Use existing hook

  const form = useForm<PulseFormValues>({
    resolver: zodResolver(pulseSchema),
    defaultValues: {
      name: "",
      dashboardId: "",
      schedule: "0 9 * * 1", // Weekly default
      webhookUrl: "",
      isActive: true,
      width: 1280,
      height: 720,
    },
  });

  useEffect(() => {
    if (pulseToEdit) {
      form.reset({
        name: pulseToEdit.name,
        dashboardId: pulseToEdit.dashboardId,
        schedule: pulseToEdit.schedule,
        webhookUrl: pulseToEdit.webhookUrl || "",
        isActive: pulseToEdit.isActive,
        width: pulseToEdit.config.width,
        height: pulseToEdit.config.height,
      });
    } else {
      form.reset({
        name: "",
        dashboardId: "",
        schedule: "0 9 * * 1",
        webhookUrl: "",
        isActive: true,
        width: 1280,
        height: 720,
      });
    }
  }, [pulseToEdit, form, open]);

  const onSubmit = async (values: PulseFormValues) => {
    setIsLoading(true);
    try {
      const payload: CreatePulseRequest = {
        name: values.name,
        dashboardId: values.dashboardId,
        schedule: values.schedule,
        webhookUrl: values.webhookUrl || "",
        isActive: values.isActive,
        config: {
          width: values.width,
          height: values.height,
          format: "png", // Default
        },
      };

      if (pulseToEdit) {
        await pulseService.updatePulse(pulseToEdit.id, payload);
        toast.success("Pulse updated successfully");
      } else {
        await pulseService.createPulse(payload);
        toast.success("Pulse created successfully");
      }
      onSuccess();
      onOpenChange(false);
    } catch (error: any) {
      toast.error(error.message || "Failed to save pulse");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[600px]">
        <DialogHeader>
          <DialogTitle>{pulseToEdit ? "Edit Pulse" : "Create New Pulse"}</DialogTitle>
          <DialogDescription>
            Configure automated dashboard screenshots delivered to Slack or Teams.
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Name</FormLabel>
                  <FormControl>
                    <Input placeholder="Weekly Sales Report" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="dashboardId"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Dashboard</FormLabel>
                  <Select
                    onValueChange={field.onChange}
                    defaultValue={field.value}
                    value={field.value}
                  >
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="Select a dashboard" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      {dashboards.map((dashboard: any) => (
                        <SelectItem key={dashboard.id} value={dashboard.id}>
                          {dashboard.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              )}
            />

            <div className="grid grid-cols-2 gap-4">
              <FormField
                control={form.control}
                name="schedule"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Cron Schedule</FormLabel>
                    <FormControl>
                      <Input placeholder="0 9 * * 1" {...field} />
                    </FormControl>
                    <FormDescription>
                      Standard cron expression (e.g., "0 9 * * 1" for Mon 9am)
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="isActive"
                render={({ field }) => (
                  <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm mt-8">
                    <div className="space-y-0.5">
                      <FormLabel>Active</FormLabel>
                    </div>
                    <FormControl>
                      <Switch checked={field.value} onCheckedChange={field.onChange} />
                    </FormControl>
                  </FormItem>
                )}
              />
            </div>

            <FormField
              control={form.control}
              name="webhookUrl"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Webhook URL (Slack/Teams)</FormLabel>
                  <FormControl>
                    <Input placeholder="https://hooks.slack.com/services/..." {...field} />
                  </FormControl>
                  <FormDescription>
                    Optional. If configured globally, this overrides it.
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <div className="grid grid-cols-2 gap-4">
              <FormField
                control={form.control}
                name="width"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Width (px)</FormLabel>
                    <FormControl>
                      <Input type="number" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="height"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Height (px)</FormLabel>
                    <FormControl>
                      <Input type="number" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
                Cancel
              </Button>
              <Button type="submit" disabled={isLoading}>
                {isLoading ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
                {pulseToEdit ? "Save Changes" : "Create Pulse"}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
