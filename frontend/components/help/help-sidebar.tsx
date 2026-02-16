'use client';

import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetDescription } from '@/components/ui/sheet';
import ReactMarkdown from 'react-markdown';
import Link from 'next/link';
import { Button } from '@/components/ui/button';
import { ExternalLink } from 'lucide-react';
import { type HelpTopic } from '@/lib/help/help-config';

interface HelpSidebarProps {
    isOpen: boolean;
    onClose: () => void;
    topic: HelpTopic;
}

export function HelpSidebar({ isOpen, onClose, topic }: HelpSidebarProps) {
    return (
        <Sheet open={isOpen} onOpenChange={onClose}>
            <SheetContent className="w-[400px] sm:w-[540px] overflow-y-auto">
                <SheetHeader>
                    <SheetTitle>{topic.title}</SheetTitle>
                    <SheetDescription>
                        Contextual help for the current page.
                    </SheetDescription>
                </SheetHeader>
                <div className="mt-6 prose prose-slate dark:prose-invert max-w-none">
                    <ReactMarkdown>{topic.content}</ReactMarkdown>
                </div>
                {topic.links && topic.links.length > 0 && (
                    <div className="mt-8 border-t pt-4">
                        <h4 className="mb-2 text-sm font-semibold">Related Resources</h4>
                        <div className="flex flex-col gap-2">
                            {topic.links.map((link, i) => (
                                <Button key={i} variant="outline" size="sm" asChild className="justify-start">
                                    <Link href={link.href} onClick={onClose}>
                                        <ExternalLink className="mr-2 h-4 w-4" />
                                        {link.title}
                                    </Link>
                                </Button>
                            ))}
                        </div>
                    </div>
                )}
            </SheetContent>
        </Sheet>
    );
}
