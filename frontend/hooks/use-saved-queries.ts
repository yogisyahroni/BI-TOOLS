'use client';

import { useState, useCallback, useEffect } from 'react';
import { useSession } from 'next-auth/react';
import { fetchWithAuth } from '@/lib/utils';
import { type SavedQuery } from '@/lib/types';

interface UseSavedQueriesOptions {
  collectionId?: string;
  autoFetch?: boolean;
}

export function useSavedQueries(options: UseSavedQueriesOptions = {}) {
  const { status } = useSession();
  const isAuthenticated = status === 'authenticated';

  const [queries, setQueries] = useState<SavedQuery[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchQueries = useCallback(async () => {
    // Skip fetch if not authenticated
    if (!isAuthenticated) {
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const params = new URLSearchParams();
      if (options.collectionId) params.append('collectionId', options.collectionId);

      const response = await fetchWithAuth(`/api/go/queries?${params.toString()}`);

      if (!response.ok) {
        throw new Error(`Failed to fetch queries: ${response.status}`);
      }

      const data = (await response.json()) as { success: boolean; data: SavedQuery[] };

      if (data.success) {
        setQueries(data.data);
      } else {
        throw new Error('Failed to fetch queries');
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error';
      setError(errorMessage);
    } finally {
      setIsLoading(false);
    }
  }, [isAuthenticated, options.collectionId]);

  const saveQuery = useCallback(
    async (query: Omit<SavedQuery, 'id' | 'createdAt' | 'updatedAt' | 'views' | 'pinned'>) => {
      try {
        const response = await fetchWithAuth('/api/go/queries', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(query),
        });

        if (!response.ok) {
          throw new Error(`Failed to save query: ${response.status}`);
        }

        const data = (await response.json()) as { success: boolean; data: SavedQuery };

        if (data.success) {
          setQueries((prev) => [data.data, ...prev]);
          return { success: true, data: data.data };
        } else {
          throw new Error('Failed to save query');
        }
      } catch (err) {
        const errorMessage = err instanceof Error ? err.message : 'Unknown error';
        return { success: false, error: errorMessage };
      }
    },
    []
  );

  const deleteQuery = useCallback(async (queryId: string) => {
    try {
      const response = await fetchWithAuth(`/api/go/queries/${queryId}`, {
        method: 'DELETE',
      });

      if (!response.ok) {
        throw new Error(`Failed to delete query: ${response.status}`);
      }

      setQueries((prev) => prev.filter((q) => q.id !== queryId));
      return { success: true };
    } catch (err) {
      return { success: false, error: err instanceof Error ? err.message : 'Unknown error' };
    }
  }, []);

  const pinQuery = useCallback(async (queryId: string) => {
    try {
      // Optimistic update
      setQueries((prev) =>
        prev.map((q) =>
          q.id === queryId
            ? { ...q, pinned: !q.pinned, updatedAt: new Date() }
            : q
        )
      );

      const query = queries.find((q) => q.id === queryId);
      if (!query) return;

      const response = await fetchWithAuth(`/api/go/queries/${queryId}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ pinned: !query.pinned }),
      });

      if (!response.ok) {
        // Revert on error
        setQueries((prev) =>
          prev.map((q) =>
            q.id === queryId
              ? { ...q, pinned: !q.pinned } // Revert
              : q
          )
        );
        throw new Error('Failed to update pin status');
      }
    } catch (err) {
      // Error already handled via optimistic revert
    }
  }, [queries]);

  const updateQuery = useCallback(async (id: string, updates: Partial<SavedQuery>) => {
    try {
      const response = await fetchWithAuth(`/api/go/queries/${id}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(updates),
      });

      if (!response.ok) {
        throw new Error(`Failed to update query: ${response.status}`);
      }

      const data = (await response.json()) as { success: boolean; data: SavedQuery };

      if (data.success) {
        setQueries((prev) =>
          prev.map((q) => (q.id === id ? data.data : q))
        );
        return { success: true, data: data.data };
      } else {
        throw new Error('Failed to update query');
      }
    } catch (err) {
      return { success: false, error: err instanceof Error ? err.message : 'Unknown error' };
    }
  }, []);

  useEffect(() => {
    // Only auto-fetch if authenticated
    if (options.autoFetch !== false && isAuthenticated) {
      fetchQueries();
    }
  }, [options.collectionId, fetchQueries, options.autoFetch, isAuthenticated]);

  return {
    queries,
    isLoading,
    error,
    fetchQueries,
    saveQuery,
    deleteQuery,
    pinQuery,
    updateQuery,
  };
}
