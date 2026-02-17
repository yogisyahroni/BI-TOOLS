import React from "react";
import {
    Tooltip,
    TooltipContent,
    TooltipProvider,
    TooltipTrigger,
} from "@/components/ui/tooltip";
import { HelpCircle } from "lucide-react";

interface ContextualHelpProps {
    content: string;
    side?: "top" | "right" | "bottom" | "left";
    className?: string;
}

export function ContextualHelp({
    content,
    side = "right",
    className,
}: ContextualHelpProps) {
    return (
        <TooltipProvider>
            <Tooltip delayDuration={300}>
                <TooltipTrigger asChild>
                    <HelpCircle
                        className={`h-4 w-4 text-muted-foreground hover:text-foreground cursor-help transition-colors ${className}`}
                    />
                </TooltipTrigger>
                <TooltipContent side={side} className="max-w-xs text-sm">
                    <p>{content}</p>
                </TooltipContent>
            </Tooltip>
        </TooltipProvider>
    );
}
