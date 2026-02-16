'use client';

import React from 'react';
import { PageLayout } from './page-layout';
import { EditorLayout } from './editor-layout';
import { usePathname } from 'next/navigation';

interface SidebarLayoutProps {
    children: React.ReactNode;
}

// Pages that need full editor layout (no TopBar)
const EDITOR_PAGES = ['/query', '/query-builder'];

export function SidebarLayout({ children }: SidebarLayoutProps) {
    const pathname = usePathname();
    
    // Check if current page is an editor page
    const isEditorPage = EDITOR_PAGES.some(page => pathname.startsWith(page));
    
    if (isEditorPage) {
        // Use EditorLayout untuk query editor (tanpa TopBar)
        return <EditorLayout>{children}</EditorLayout>;
    }
    
    // Use PageLayout untuk pages lainnya (dengan TopBar)
    return <PageLayout>{children}</PageLayout>;
}
