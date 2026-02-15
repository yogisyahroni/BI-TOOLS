"use client";

import { CheckCircle2, AlertTriangle } from "lucide-react";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";

interface VerificationBadgeProps {
    status: "none" | "verified" | "deprecated";
    className?: string;
    showLabel?: boolean;
}

export function VerificationBadge({ status, className, showLabel = false }: VerificationBadgeProps) {
    if (!status || status === "none") return null;

    const config = {
        verified: {
            icon: CheckCircle2,
            color: "text-green-500",
            bg: "bg-green-500/10",
            border: "border-green-500/20",
            label: "Verified",
            tooltip: "This item has been verified by an administrator/editor as accurate and trusted.",
        },
        deprecated: {
            icon: AlertTriangle,
            color: "text-amber-500",
            bg: "bg-amber-500/10",
            border: "border-amber-500/20",
            label: "Deprecated",
            tooltip: "This item is deprecated and should not be used for new analysis.",
        },
    };

    const { icon: Icon, color, bg, border, label, tooltip } = config[status];

    return (
        <TooltipProvider>
            <Tooltip>
                <TooltipTrigger asChild>
                    <div
                        className={cn(
                            "inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full border text-xs font-medium cursor-help transition-colors",
                            bg,
                            border,
                            color,
                            className
                        )}
                    >
                        <Icon className="w-3.5 h-3.5" />
                        {showLabel && <span>{label}</span>}
                    </div>
                </TooltipTrigger>
                <TooltipContent>
                    <p>{tooltip}</p>
                </TooltipContent>
            </Tooltip>
        </TooltipProvider>
    );
}
