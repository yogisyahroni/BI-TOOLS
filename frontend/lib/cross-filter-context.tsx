'use client';

import React, { createContext, useContext, useState, useCallback, useMemo } from 'react';

/**
 * Filter operators supported by the cross-filter system
 */
export type FilterOperator =
    | 'equals'
    | 'not_equals'
    | 'in'
    | 'not_in'
    | 'between'
    | 'greater_than'
    | 'less_than'
    | 'contains'
    | 'starts_with'
    | 'ends_with';

/**
 * Filter criteria structure for cross-filtering
 */
export interface FilterCriteria {
    /** Unique identifier for this filter */
    id: string;
    /** ID of the chart that created this filter */
    sourceChartId: string;
    /** Field/column name being filtered */
    fieldName: string;
    /** Filter operator */
    operator: FilterOperator;
    /** Filter value(s) */
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    value: any;
    /** Display label for the filter */
    label?: string;
    /** Timestamp when filter was created */
    timestamp: Date;
    /** Filter type (for rendering different UI) */
    type?: 'chart' | 'global';
}

/**
 * Cross-filter context value interface
 */
export interface CrossFilterContextValue {
    /** Map of active filters (key: filterId, value: FilterCriteria) */
    filters: Map<string, FilterCriteria>;

    /** Add a new filter */
    addFilter: (filter: Omit<FilterCriteria, 'id' | 'timestamp'>) => string;

    /** Update an existing filter */
    updateFilter: (filterId: string, updates: Partial<FilterCriteria>) => void;

    /** Remove a specific filter */
    removeFilter: (filterId: string) => void;

    /** Remove all filters from a specific chart */
    removeChartFilters: (chartId: string) => void;

    /** Clear all filters */
    clearFilters: () => void;

    /** Clear only global filters */
    clearGlobalFilters: () => void;

    /** Clear only chart filters */
    clearChartFilters: () => void;

    /** Get all active filters as array */
    getActiveFilters: () => FilterCriteria[];

    /** Get filters for a specific field */
    getFiltersForField: (fieldName: string) => FilterCriteria[];

    /** Get filters excluding those from a specific chart (avoid circular filtering) */
    getFiltersExcludingChart: (chartId: string) => FilterCriteria[];

    /** Check if a specific chart has active filters */
    isChartFiltered: (chartId: string) => boolean;

    /** Check if a specific field is filtered */
    isFieldFiltered: (fieldName: string) => boolean;

    /** Get count of active filters */
    getFilterCount: () => number;

    /** Get count of chart filters */
    getChartFilterCount: () => number;

    /** Get count of global filters */
    getGlobalFilterCount: () => number;
}

/**
 * Create the context
 */
const CrossFilterContext = createContext<CrossFilterContextValue | undefined>(undefined);

/**
 * Props for CrossFilterProvider
 */
export interface CrossFilterProviderProps {
    children: React.ReactNode;
    /** Optional initial filters */
    initialFilters?: FilterCriteria[];
    /** Optional callback when filters change */
    onFiltersChange?: (filters: FilterCriteria[]) => void;
}

/**
 * CrossFilterProvider component
 * Wraps the dashboard to provide cross-filtering capabilities
 */
export function CrossFilterProvider({
    children,
    initialFilters = [],
    onFiltersChange
}: CrossFilterProviderProps) {
    const [filters, setFilters] = useState<Map<string, FilterCriteria>>(() => {
        const map = new Map<string, FilterCriteria>();
        initialFilters.forEach(filter => {
            map.set(filter.id, filter);
        });
        return map;
    });

    /**
     * Add a new filter
     */
    const addFilter = useCallback((filterData: Omit<FilterCriteria, 'id' | 'timestamp'>): string => {
        const newFilter: FilterCriteria = {
            ...filterData,
            id: `filter-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
            timestamp: new Date(),
        };

        setFilters(prev => {
            const newFilters = new Map(prev);
            newFilters.set(newFilter.id, newFilter);

            // Trigger callback with array of filters
            if (onFiltersChange) {
                onFiltersChange(Array.from(newFilters.values()));
            }

            return newFilters;
        });

        return newFilter.id;
    }, [onFiltersChange]);

    /**
     * Update an existing filter
     */
    const updateFilter = useCallback((filterId: string, updates: Partial<FilterCriteria>) => {
        setFilters(prev => {
            const filter = prev.get(filterId);
            if (!filter) return prev;

            const newFilters = new Map(prev);
            newFilters.set(filterId, { ...filter, ...updates });

            if (onFiltersChange) {
                onFiltersChange(Array.from(newFilters.values()));
            }

            return newFilters;
        });
    }, [onFiltersChange]);

    /**
     * Remove a specific filter
     */
    const removeFilter = useCallback((filterId: string) => {
        setFilters(prev => {
            if (!prev.has(filterId)) return prev;

            const newFilters = new Map(prev);
            newFilters.delete(filterId);

            if (onFiltersChange) {
                onFiltersChange(Array.from(newFilters.values()));
            }

            return newFilters;
        });
    }, [onFiltersChange]);

    /**
     * Remove all filters from a specific chart
     */
    const removeChartFilters = useCallback((chartId: string) => {
        setFilters(prev => {
            const newFilters = new Map(prev);
            let hasChanges = false;

            for (const [id, filter] of newFilters.entries()) {
                if (filter.sourceChartId === chartId) {
                    newFilters.delete(id);
                    hasChanges = true;
                }
            }

            if (hasChanges) {
                if (onFiltersChange) {
                    onFiltersChange(Array.from(newFilters.values()));
                }
                return newFilters;
            }

            return prev;
        });
    }, [onFiltersChange]);

    /**
     * Clear all filters
     */
    const clearFilters = useCallback(() => {
        setFilters(prev => {
            if (prev.size === 0) return prev;

            if (onFiltersChange) {
                onFiltersChange([]);
            }

            return new Map();
        });
    }, [onFiltersChange]);

    /**
     * Clear only global filters
     */
    const clearGlobalFilters = useCallback(() => {
        setFilters(prev => {
            const newFilters = new Map(prev);
            let hasChanges = false;

            for (const [id, filter] of newFilters.entries()) {
                if (filter.type === 'global') {
                    newFilters.delete(id);
                    hasChanges = true;
                }
            }

            if (hasChanges) {
                if (onFiltersChange) {
                    onFiltersChange(Array.from(newFilters.values()));
                }
                return newFilters;
            }

            return prev;
        });
    }, [onFiltersChange]);

    /**
     * Clear only chart filters
     */
    const clearChartFilters = useCallback(() => {
        setFilters(prev => {
            const newFilters = new Map(prev);
            let hasChanges = false;

            for (const [id, filter] of newFilters.entries()) {
                if (filter.type === 'chart' || !filter.type) {
                    newFilters.delete(id);
                    hasChanges = true;
                }
            }

            if (hasChanges) {
                if (onFiltersChange) {
                    onFiltersChange(Array.from(newFilters.values()));
                }
                return newFilters;
            }

            return prev;
        });
    }, [onFiltersChange]);

    /**
     * Get all active filters as array
     */
    const getActiveFilters = useCallback(() => {
        return Array.from(filters.values());
    }, [filters]);

    /**
     * Get filters for a specific field
     */
    const getFiltersForField = useCallback((fieldName: string) => {
        return Array.from(filters.values()).filter(f => f.fieldName === fieldName);
    }, [filters]);

    /**
     * Get filters excluding those from a specific chart
     */
    const getFiltersExcludingChart = useCallback((chartId: string) => {
        return Array.from(filters.values()).filter(f => f.sourceChartId !== chartId);
    }, [filters]);

    /**
     * Check if a specific chart has active filters
     */
    const isChartFiltered = useCallback((chartId: string) => {
        for (const filter of filters.values()) {
            if (filter.sourceChartId === chartId) {
                return true;
            }
        }
        return false;
    }, [filters]);

    /**
     * Check if a specific field is filtered
     */
    const isFieldFiltered = useCallback((fieldName: string) => {
        for (const filter of filters.values()) {
            if (filter.fieldName === fieldName) {
                return true;
            }
        }
        return false;
    }, [filters]);

    /**
     * Get count of active filters
     */
    const getFilterCount = useCallback(() => {
        return filters.size;
    }, [filters]);

    /**
     * Get count of chart filters
     */
    const getChartFilterCount = useCallback(() => {
        let count = 0;
        for (const filter of filters.values()) {
            if (filter.type === 'chart' || !filter.type) {
                count++;
            }
        }
        return count;
    }, [filters]);

    /**
     * Get count of global filters
     */
    const getGlobalFilterCount = useCallback(() => {
        let count = 0;
        for (const filter of filters.values()) {
            if (filter.type === 'global') {
                count++;
            }
        }
        return count;
    }, [filters]);

    // Memoize context value to prevent unnecessary re-renders
    const contextValue = useMemo<CrossFilterContextValue>(() => ({
        filters,
        addFilter,
        updateFilter,
        removeFilter,
        removeChartFilters,
        clearFilters,
        clearGlobalFilters,
        clearChartFilters,
        getActiveFilters,
        getFiltersForField,
        getFiltersExcludingChart,
        isChartFiltered,
        isFieldFiltered,
        getFilterCount,
        getChartFilterCount,
        getGlobalFilterCount,
    }), [
        filters,
        addFilter,
        updateFilter,
        removeFilter,
        removeChartFilters,
        clearFilters,
        clearGlobalFilters,
        clearChartFilters,
        getActiveFilters,
        getFiltersForField,
        getFiltersExcludingChart,
        isChartFiltered,
        isFieldFiltered,
        getFilterCount,
        getChartFilterCount,
        getGlobalFilterCount,
    ]);

    return (
        <CrossFilterContext.Provider value={contextValue}>
            {children}
        </CrossFilterContext.Provider>
    );
}

/**
 * Custom hook to use cross-filter context
 * @throws Error if used outside of CrossFilterProvider
 */
export function useCrossFilter(): CrossFilterContextValue {
    const context = useContext(CrossFilterContext);

    if (!context) {
        throw new Error('useCrossFilter must be used within a CrossFilterProvider');
    }

    return context;
}

/**
 * Helper hook to apply filters to data
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
 */
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
export function useFilteredData<T extends Record<string, any>>(
    data: T[] | undefined,
    chartId: string,
    fieldMapping?: Record<string, string>
): T[] {
    const { getFiltersExcludingChart } = useCrossFilter();

    return useMemo(() => {
        if (!data || data.length === 0) return [];

        const applicableFilters = getFiltersExcludingChart(chartId);
        if (applicableFilters.length === 0) return data;

        return data.filter(row => {
            // Apply ALL filters (AND logic)
            return applicableFilters.every(filter => {
                const fieldName = fieldMapping?.[filter.fieldName] || filter.fieldName;
                const value = row[fieldName];

                if (value === undefined || value === null) return false;

                switch (filter.operator) {
                    case 'equals':
                        return value === filter.value;

                    case 'not_equals':
                        return value !== filter.value;

                    case 'in':
                        return Array.isArray(filter.value) && filter.value.includes(value);

                    case 'not_in':
                        return Array.isArray(filter.value) && !filter.value.includes(value);

                    case 'between':
                        if (!Array.isArray(filter.value) || filter.value.length !== 2) return false;
                        return value >= filter.value[0] && value <= filter.value[1];

                    case 'greater_than':
                        return value > filter.value;

                    case 'less_than':
                        return value < filter.value;

                    case 'contains':
                        return String(value).toLowerCase().includes(String(filter.value).toLowerCase());

                    case 'starts_with':
                        return String(value).toLowerCase().startsWith(String(filter.value).toLowerCase());

                    case 'ends_with':
                        return String(value).toLowerCase().endsWith(String(filter.value).toLowerCase());

                    default:
                        return true;
                }
            });
        });
    }, [data, chartId, fieldMapping, getFiltersExcludingChart]);
}
