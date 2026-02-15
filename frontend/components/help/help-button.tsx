'use client';

import { Button } from '@/components/ui/button';
import { HelpCircle } from 'lucide-react';
import { useHelp } from '@/components/providers/help-provider';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';

export function HelpButton() {
    const { toggle } = useHelp();

    return (
        <TooltipProvider>
            <Tooltip>
                <TooltipTrigger asChild>
                    <Button variant="ghost" size="icon" onClick={toggle} className="text-muted-foreground hover:text-foreground">
                        <HelpCircle className="h-5 w-5" />
                        <span className="sr-only">Help</span>
                    </Button>
                </TooltipTrigger>
                <TooltipContent>
                    <p>Get Help</p>
                </TooltipContent>
            </Tooltip>
        </TooltipProvider>
    );
}
