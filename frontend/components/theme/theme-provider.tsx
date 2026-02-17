'use client';

import { useEffect, useCallback, createContext, useContext, useState } from 'react';
import { useParams } from 'next/navigation';
import { useTheme } from 'next-themes';
import { _fetchWithAuth } from '@/lib/utils';

interface ThemeConfig {
    primaryColor: string;
    radius: number;
    fontFamily: string;
    chartPalette?: string[];
    darkMode: boolean;
}

interface ThemeContextType {
    theme: ThemeConfig;
    isLoading: boolean;
}

const DEFAULT_THEME: ThemeConfig = {
    primaryColor: '#7c3aed',
    radius: 0.5,
    fontFamily: 'Inter',
    chartPalette: undefined,
    darkMode: false
};

const ThemeContext = createContext<ThemeContextType>({
    theme: DEFAULT_THEME,
    isLoading: true
});

export function useWorkspaceTheme() {
    return useContext(ThemeContext);
}

export function WorkspaceThemeProvider({ children }: { children: React.ReactNode }) {
    const params = useParams();
    const workspaceId = params?.workspaceId as string;
    const { _setTheme } = useTheme();
    const [config, _setConfig] = useState<ThemeConfig>(DEFAULT_THEME);
    const [isLoading, setIsLoading] = useState(true);

    const applyTheme = useCallback((newConfig: ThemeConfig) => {
        const root = document.documentElement;

        // Apply Radius
        root.style.setProperty('--radius', `${newConfig.radius}rem`);

        // Apply Primary Color
        root.style.setProperty('--primary-brand', newConfig.primaryColor);

        // Apply Font
        if (newConfig.fontFamily) {
            root.style.setProperty('--font-sans', newConfig.fontFamily);
        }

        // Apply Mode (sync with next-themes if enforced)
        // if (newConfig.darkMode) setTheme('dark');
    }, []);

    useEffect(() => {
        if (!workspaceId) return;

        // TASK: Implement backend theme endpoint
        // For now, use default theme to avoid 404s
        /*
        async function fetchTheme() {
            try {
                const res = await fetchWithAuth(`/api/go/workspaces/${workspaceId}/theme`);
                if (res.ok) {
                    const data = await res.json();
                    setConfig(data);
                    applyTheme(data);
                }
            } catch (error) {
                console.error("Failed to load theme:", error);
            } finally {
                setIsLoading(false);
            }
        }

        fetchTheme(); 
        */
        setIsLoading(false);
    }, [workspaceId, applyTheme]);

    return (
        <ThemeContext.Provider value={{ theme: config, isLoading }}>
            {children}
        </ThemeContext.Provider>
    );
}
