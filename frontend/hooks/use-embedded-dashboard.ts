'use client';

import { useState, useCallback, useEffect } from 'react';
import { Dashboard, DashboardCard, DashboardFilter } from '@/lib/types';

export function useEmbeddedDashboard(dashboardId: string, token: string | null) {
    const [dashboard, setDashboard] = useState<Dashboard | null>(null);
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    // Local state for layout (read-only mostly, but needed for grid)
    const [cards, setCards] = useState<DashboardCard[]>([]);
    const [filters, setFilters] = useState<DashboardFilter[]>([]);
    const [filterValues, setFilterValues] = useState<Record<string, any>>({});

    // Load dashboard
    const fetchDashboard = useCallback(async () => {
        if (!token) {
            setError("Missing embed token");
            setIsLoading(false);
            return;
        }

        setIsLoading(true);
        try {
            // Validate token and get dashboard config
            // The backend endpoint is /api/embed/token/validate?token=...
            const res = await fetch(`/api/embed/token/validate?token=${token}`);

            if (!res.ok) {
                if (res.status === 401) throw new Error('Invalid or expired token');
                throw new Error('Failed to load embedded dashboard');
            }

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
    }, [dashboardId, token]);

    useEffect(() => {
        if (dashboardId && token) {
            fetchDashboard();
        } else if (!token) {
            setIsLoading(false);
            setError("No token provided");
        }
    }, [dashboardId, token, fetchDashboard]);

    // Read-only handlers (no-ops or limited)
    const updateLayout = useCallback(() => { }, []);
    const removeCard = useCallback(() => { }, []);

    // Filter handlers (allowed)
    const setFilterValue = useCallback((key: string, value: any) => {
        setFilterValues(prev => ({ ...prev, [key]: value }));
    }, []);

    const addFilter = useCallback(() => { }, []); // No-op
    const removeFilter = useCallback(() => { }, []); // No-op

    return {
        dashboard: dashboard ? { ...dashboard, cards, filters } : null,
        filters,
        filterValues,
        isLoading,
        error,
        isEditing: false, // Always false for embed
        setIsEditing: () => { },
        updateLayout,
        addCard: () => { },
        removeCard,
        addFilter,
        removeFilter,
        setFilterValue,
        saveDashboard: async () => { }, // No-op
        togglePublic: async () => { } // No-op
    };
}
