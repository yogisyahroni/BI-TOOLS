'use client';

import { useState, useCallback, useEffect } from 'react';
import { Dashboard, DashboardCard, VisualizationConfig, DashboardFilter } from '@/lib/types';
import { useQueryExecution } from '@/hooks/use-query-execution';

export function useDashboard(dashboardId: string) {
    const [dashboard, setDashboard] = useState<Dashboard | null>(null);
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [isEditing, setIsEditing] = useState(false);

    // Local state for layout changes before saving
    const [cards, setCards] = useState<DashboardCard[]>([]);

    // Filter Configuration State
    const [filters, setFilters] = useState<DashboardFilter[]>([]);

    // Runtime Filter Values
    const [filterValues, setFilterValues] = useState<Record<string, any>>({});

    // Load dashboard
    const fetchDashboard = useCallback(async () => {
        setIsLoading(true);
        try {
            const res = await fetch(`/api/go/dashboards/${dashboardId}`);
            if (!res.ok) throw new Error('Failed to fetch dashboard');
            const data = await res.json();

            if (data.success && data.data) {
                setDashboard(data.data);
                setCards(data.data.cards || []);
                setFilters(data.data.filters || []);

                // Initialize default values
                const initialValues: Record<string, any> = {};
                if (data.data.filters) {
                    data.data.filters.forEach((f: DashboardFilter) => {
                        if (f.defaultValue) initialValues[f.key] = f.defaultValue;
                    });
                }
                setFilterValues(initialValues);
            } else {
                setError(data.error || 'Unknown error');
            }
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to load dashboard');
        } finally {
            setIsLoading(false);
        }
    }, [dashboardId]);

    useEffect(() => {
        if (dashboardId) {
            fetchDashboard();
        }
    }, [dashboardId, fetchDashboard]);

    const updateLayout = useCallback((layout: any[]) => {
        setCards(prevCards => {
            return prevCards.map(card => {
                const layoutItem = layout.find((l: any) => l.i === card.id);
                if (layoutItem) {
                    return {
                        ...card,
                        position: {
                            x: layoutItem.x,
                            y: layoutItem.y,
                            w: layoutItem.w,
                            h: layoutItem.h
                        }
                    };
                }
                return card;
            });
        });
    }, []);

    const addCard = useCallback((card: Omit<DashboardCard, 'id' | 'dashboardId' | 'position'>) => {
        const newCard: DashboardCard = {
            ...card,
            id: `temp-${Date.now()}`, // Temporary ID until saved
            dashboardId,
            position: { x: 0, y: Infinity, w: 4, h: 6 } // Place at bottom
        };
        setCards(prev => [...prev, newCard]);
    }, [dashboardId]);

    const removeCard = useCallback((cardId: string) => {
        setCards(prev => prev.filter(c => c.id !== cardId));
    }, []);

    const updateCard = useCallback((cardId: string, updates: Partial<DashboardCard>) => {
        setCards(prev => prev.map(c => c.id === cardId ? { ...c, ...updates } : c));
    }, []);

    // Filter Management
    const addFilter = useCallback((filter: DashboardFilter) => {
        setFilters(prev => [...prev, filter]);
    }, []);

    const removeFilter = useCallback((filterId: string) => {
        setFilters(prev => prev.filter(f => f.id !== filterId));
    }, []);

    const setFilterValue = useCallback((key: string, value: any) => {
        setFilterValues(prev => ({ ...prev, [key]: value }));
    }, []);

    const saveDashboard = useCallback(async () => {
        try {
            const res = await fetch(`/api/go/dashboards/${dashboardId}`, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    filters, // Save filter config
                    cards: cards.map(c => ({
                        queryId: c.queryId,
                        type: c.type,
                        title: c.title,
                        textContent: c.textContent,
                        position: c.position,
                        visualizationConfig: c.visualizationConfig
                    }))
                })
            });

            if (!res.ok) throw new Error('Failed to save dashboard');

            const data = await res.json();
            if (data.success && data.data) {
                setDashboard(data.data);
                setCards(data.data.cards);
                setFilters(data.data.filters || []);
                setIsEditing(false);
            }
        } catch (err) {
            // Error handled via UI toast
        }
    }, [dashboardId, cards, filters]);

    const togglePublic = useCallback(async (isPublic: boolean) => {
        try {
            const res = await fetch(`/api/go/dashboards/${dashboardId}`, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ isPublic })
            });

            if (!res.ok) throw new Error('Failed to update visibility');

            const data = await res.json();
            if (data.success && data.data) {
                setDashboard(prev => prev ? { ...prev, isPublic: data.data.isPublic } : null);
            }
        } catch (err) {
            throw err;
        }
    }, [dashboardId]);

    const certifyDashboard = useCallback(async (status: 'verified' | 'deprecated' | 'none') => {
        try {
            const res = await fetch(`/api/go/dashboards/${dashboardId}/certify`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ status })
            });

            if (!res.ok) throw new Error('Failed to update certification status');

            const data = await res.json();
            if (data.success && data.data) {
                setDashboard(prev => prev ? { ...prev, certificationStatus: status, certifiedBy: data.data.certifiedBy, certifiedAt: data.data.certifiedAt } : null);
                return { success: true };
            }
            return { success: false, error: data.message };
        } catch (err) {
            return { success: false, error: err instanceof Error ? err.message : 'Unknown error' };
        }
    }, [dashboardId]);

    return {
        dashboard: dashboard ? { ...dashboard, cards, filters } : null,
        filters,
        filterValues,
        isLoading,
        error,
        isEditing,
        setIsEditing,
        updateLayout,
        addCard,
        removeCard,
        updateCard,
        addFilter,
        removeFilter,
        setFilterValue,
        saveDashboard,
        togglePublic,
        certifyDashboard
    };
}
