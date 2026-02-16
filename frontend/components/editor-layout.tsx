'use client';

import React from 'react';
import { MainSidebar } from './main-sidebar';
import { useSidebar } from '@/contexts/sidebar-context';
import { cn } from '@/lib/utils';

interface EditorLayoutProps {
    children: React.ReactNode;
    className?: string;
}

// Layout khusus untuk Query Editor yang tidak pakai TopBar
export function EditorLayout({ children, className }: EditorLayoutProps) {
    const { isOpen, close, isCollapsed, toggleCollapse } = useSidebar();

    return (
        <div className="flex min-h-screen bg-background">
            {/* Sidebar */}
            <MainSidebar 
                isOpen={isOpen} 
                onClose={close} 
                isCollapsed={isCollapsed}
                onToggleCollapse={toggleCollapse}
            />

            {/* Main Content Area - Tanpa TopBar */}
            <div 
                className={cn(
                    'flex-1 flex flex-col min-w-0 transition-all duration-300',
                    isCollapsed ? 'lg:ml-16' : 'lg:ml-64'
                )}
            >
                <main 
                    className={cn(
                        'flex-1 overflow-hidden',
                        className
                    )}
                >
                    {children}
                </main>
            </div>
        </div>
    );
}
