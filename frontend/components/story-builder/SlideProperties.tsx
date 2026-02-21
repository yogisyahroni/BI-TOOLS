"use client";

import { Slide } from "@/types/story";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { LayoutTemplate, Database } from "lucide-react";
import { useDashboards } from "@/hooks/use-dashboards";

const LAYOUT_LABELS: Record<string, string> = {
  title: "Title Slide",
  bullet_points: "Bullet Points",
  image_text: "Image + Text",
  chart: "Chart",
};

export function SlideProperties({
  slide,
  onChange,
}: {
  slide: Slide;
  onChange: (field: keyof Slide, value: any) => void;
}) {
  // using autoFetch true to list dashboards
  const { dashboards } = useDashboards({ autoFetch: true });

  const handleDataBindingChange = (field: "dashboard_id" | "card_id", val: string) => {
    const newDataBinding = { ...slide.data_binding, [field]: val };
    onChange("data_binding", newDataBinding);
  };

  const selectedDashboard = dashboards.find((d) => d.id === slide.data_binding?.dashboard_id);
  const cards = selectedDashboard?.cards || [];

  if (!slide) return <div className="w-80 border-l bg-background shrink-0"></div>;

  return (
    <div className="w-80 border-l bg-background flex flex-col shrink-0 overflow-y-auto">
      <div className="p-4 border-b">
        <h3 className="font-semibold tracking-tight">Slide Properties</h3>
      </div>
      <div className="p-6 space-y-8">
        {/* Layout Configuration */}
        <div className="space-y-3">
          <Label className="text-xs font-semibold text-muted-foreground uppercase tracking-wider flex items-center gap-1.5 border-b pb-2">
            <LayoutTemplate className="h-3.5 w-3.5" /> Layout Selection
          </Label>
          <Select value={slide.layout} onValueChange={(val) => onChange("layout", val)}>
            <SelectTrigger className="h-9">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {Object.entries(LAYOUT_LABELS).map(([key, label]) => (
                <SelectItem key={key} value={key}>
                  {label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {/* Data Binding Configuration (Only for Chart) */}
        {slide.layout === "chart" && (
          <div className="space-y-4 p-5 rounded-lg bg-primary/5 border border-primary/10 transition-all duration-300">
            <Label className="text-xs font-bold text-primary uppercase tracking-wider flex items-center gap-1.5 border-b border-primary/10 pb-2">
              <Database className="h-4 w-4" /> Data Source
            </Label>
            <div className="space-y-4">
              <div className="space-y-1.5">
                <Label className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">
                  Dashboard
                </Label>
                <Select
                  value={slide.data_binding?.dashboard_id || ""}
                  onValueChange={(val) => {
                    onChange("data_binding", { dashboard_id: val, card_id: "" });
                  }}
                >
                  <SelectTrigger className="h-9">
                    <SelectValue placeholder="Select a dashboard" />
                  </SelectTrigger>
                  <SelectContent>
                    {dashboards.map((d) => (
                      <SelectItem key={d.id} value={d.id}>
                        {d.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              {selectedDashboard && (
                <div className="space-y-1.5">
                  <Label className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">
                    Chart / Metric
                  </Label>
                  <Select
                    value={slide.data_binding?.card_id || ""}
                    onValueChange={(val) => handleDataBindingChange("card_id", val)}
                  >
                    <SelectTrigger className="h-9">
                      <SelectValue placeholder="Select a card" />
                    </SelectTrigger>
                    <SelectContent>
                      {cards.map((c) => (
                        <SelectItem key={c.id} value={c.id}>
                          {c.title || `Card ${c.id}`}
                        </SelectItem>
                      ))}
                      {cards.length === 0 && (
                        <SelectItem value="none" disabled>
                          No cards available
                        </SelectItem>
                      )}
                    </SelectContent>
                  </Select>
                </div>
              )}
            </div>
          </div>
        )}

        <div className="space-y-6 pt-4 border-t">
          <div className="space-y-2">
            <Label className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
              Title
            </Label>
            <Input
              value={slide.title}
              onChange={(e) => onChange("title", e.target.value)}
              placeholder="Awesome Slide Title"
              className="h-9"
            />
          </div>

          {slide.layout !== "title" && slide.layout !== "chart" && (
            <div className="space-y-2">
              <Label className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                Content{" "}
                {slide.layout === "bullet_points" && (
                  <span className="text-[10px] lowercase normal-case opacity-60">
                    (one per line)
                  </span>
                )}
              </Label>
              <Textarea
                value={slide.content}
                onChange={(e) => onChange("content", e.target.value)}
                placeholder={
                  slide.layout === "bullet_points"
                    ? "\u2022 Great performance\n\u2022 Increased revenue"
                    : "Write your detailed explanation here..."
                }
                rows={6}
                className="resize-none"
              />
            </div>
          )}

          <div className="space-y-2">
            <Label className="text-xs font-semibold text-muted-foreground uppercase tracking-wider opacity-80">
              Speaker Notes
            </Label>
            <Textarea
              value={slide.notes || ""}
              onChange={(e) => onChange("notes", e.target.value)}
              placeholder="Don't forget to mention..."
              rows={4}
              className="resize-none italic bg-muted/30"
            />
          </div>
        </div>
      </div>
    </div>
  );
}
