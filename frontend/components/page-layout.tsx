'use client';

import React from 'react';
import { MainSidebar } from './main-sidebar';
import { TopBar } from './top-bar';
import { useSidebar } from '@/contexts/sidebar-context';
import { cn } from '@/lib/utils';

interface PageLayoutProps {
    children: React.ReactNode;
    className?: string;
}

export function PageLayout({ children, className }: PageLayoutProps) {
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

            {/* Main Content Area */}
            <div 
                className={cn(
                    'flex-1 flex flex-col min-w-0 transition-all duration-300',
                    isCollapsed ? 'lg:ml-16' : 'lg:ml-64'
                )}
            >
                <TopBar />
                
                <main 
                    className={cn(
                        'flex-1 overflow-auto p-4 lg:p-8',
                        className
                    )}
                >
                    <div className="max-w-7xl mx-auto">
                        {children}
                    </div>
                </main>
            </div>
        </div>
    );
}
