'use client';

import React, { createContext, useContext, useState, useEffect } from 'react';
import { usePathname } from 'next/navigation';
import { getHelpTopic, type HelpTopic } from '@/lib/help/help-config';

interface HelpContextType {
    isOpen: boolean;
    toggle: () => void;
    open: () => void;
    close: () => void;
    topic: HelpTopic;
}

const HelpContext = createContext<HelpContextType | undefined>(undefined);

export function HelpProvider({ children }: { children: React.ReactNode }) {
    const [isOpen, setIsOpen] = useState(false);
    const pathname = usePathname();
    const [topic, setTopic] = useState<HelpTopic>(getHelpTopic(pathname));

    useEffect(() => {
        setTopic(getHelpTopic(pathname));
    }, [pathname]);

    const toggle = () => setIsOpen((prev) => !prev);
    const open = () => setIsOpen(true);
    const close = () => setIsOpen(false);

    return (
        <HelpContext.Provider value={{ isOpen, toggle, open, close, topic }}>
            {children}
            <HelpSidebar isOpen={isOpen} onClose={close} topic={topic} />
        </HelpContext.Provider>
    );
}

export function useHelp() {
    const context = useContext(HelpContext);
    if (!context) {
        throw new Error('useHelp must be used within a HelpProvider');
    }
    return context;
}

// Circular dependency avoidance: We'll import HelpSidebar dynamically or defined in a separate file.
// For simplicity in this file-write, I'll import it from components.
import { HelpSidebar } from '@/components/help/help-sidebar';
