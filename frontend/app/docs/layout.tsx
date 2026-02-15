'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { cn } from '@/lib/utils';
import { Book, Code, Home, Menu, StickyNote } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Sheet, SheetContent, SheetTrigger } from '@/components/ui/sheet';
import { useState } from 'react';

const docsConfig = [
    {
        title: 'Getting Started',
        items: [
            { title: 'Introduction', href: '/docs' },
            { title: 'Quick Start', href: '/docs/quick-start' },
        ],
    },
    {
        title: 'Core Concepts',
        items: [
            { title: 'Dashboards', href: '/docs/dashboards' },
            { title: 'Data Sources', href: '/docs/data-sources' },
            { title: 'Alerts & Notifications', href: '/docs/alerts' },
        ],
    },
    {
        title: 'Advanced',
        items: [
            { title: 'API Reference', href: '/docs/api', icon: Code },
        ],
    },
];

interface DocsLayoutProps {
    children: React.ReactNode;
}

export default function DocsLayout({ children }: DocsLayoutProps) {
    const pathname = usePathname();
    const [isOpen, setIsOpen] = useState(false);

    return (
        <div className="flex min-h-screen flex-col lg:flex-row">
            {/* Mobile Header */}
            <header className="sticky top-0 z-50 flex h-14 items-center gap-4 border-b bg-background px-6 lg:hidden">
                <Sheet open={isOpen} onOpenChange={setIsOpen}>
                    <SheetTrigger asChild>
                        <Button variant="ghost" size="icon" className="lg:hidden">
                            <Menu className="h-6 w-6" />
                            <span className="sr-only">Toggle navigation menu</span>
                        </Button>
                    </SheetTrigger>
                    <SheetContent side="left" className="w-64 p-0">
                        <DocsSidebar pathname={pathname} onNavigate={() => setIsOpen(false)} />
                    </SheetContent>
                </Sheet>
                <div className="font-semibold">Documentation</div>
            </header>

            {/* Desktop Sidebar */}
            <aside className="hidden w-64 border-r bg-background lg:block">
                <DocsSidebar pathname={pathname} />
            </aside>

            {/* Main Content */}
            <main className="flex-1 overflow-y-auto">
                <div className="container max-w-4xl py-6 lg:py-10">
                    {children}
                </div>
            </main>
        </div>
    );
}

function DocsSidebar({ pathname, onNavigate }: { pathname: string; onNavigate?: () => void }) {
    return (
        <ScrollArea className="h-full py-6 pr-6 pl-4">
            <div className="mb-6 flex items-center px-2 font-bold text-lg">
                <Book className="mr-2 h-5 w-5" />
                <span>InsightEngine Docs</span>
            </div>
            {docsConfig.map((group, index) => (
                <div key={index} className="mb-6">
                    <h4 className="mb-2 rounded-md px-2 py-1 text-sm font-semibold">{group.title}</h4>
                    {group.items.length > 0 && (
                        <div className="grid grid-flow-row auto-rows-max text-sm">
                            {group.items.map((item, i) => (
                                <Link
                                    key={i}
                                    href={item.href}
                                    onClick={onNavigate}
                                    className={cn(
                                        'group flex w-full items-center rounded-md border border-transparent px-2 py-1.5 hover:underline',
                                        pathname === item.href
                                            ? 'font-medium text-foreground'
                                            : 'text-muted-foreground'
                                    )}
                                >
                                    {item.icon && <item.icon className="mr-2 h-4 w-4" />}
                                    {item.title}
                                </Link>
                            ))}
                        </div>
                    )}
                </div>
            ))}
        </ScrollArea>
    );
}
