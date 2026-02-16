'use client';

import React, { createContext, useContext, useState, ReactNode } from 'react';

interface SidebarContextType {
    isOpen: boolean;
    toggle: () => void;
    open: () => void;
    close: () => void;
    isCollapsed: boolean;
    toggleCollapse: () => void;
}

const SidebarContext = createContext<SidebarContextType | undefined>(undefined);

export function SidebarProvider({ children }: { children: ReactNode }) {
    const [isOpen, setIsOpen] = useState(false);
    const [isCollapsed, setIsCollapsed] = useState(false);

    const toggle = () => setIsOpen((prev) => !prev);
    const open = () => setIsOpen(true);
    const close = () => setIsOpen(false);
    const toggleCollapse = () => setIsCollapsed((prev) => !prev);

    return (
        <SidebarContext.Provider value={{ isOpen, toggle, open, close, isCollapsed, toggleCollapse }}>
            {children}
        </SidebarContext.Provider>
    );
}

export function useSidebar() {
    const context = useContext(SidebarContext);
    if (context === undefined) {
        throw new Error('useSidebar must be used within a SidebarProvider');
    }
    return context;
}
