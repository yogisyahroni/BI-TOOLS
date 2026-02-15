'use client';

import { useState, useCallback } from 'react';
import { type QueryHistory } from '@/lib/types';

export function useQueryHistory() {
  const [history, setHistory] = useState<QueryHistory[]>([]);

  const addToHistory = useCallback(
    (
      sql: string,
      connectionId: string,
      status: 'success' | 'error',
      executionTime: number,
      rowsReturned: number,
      error?: string,
      aiPrompt?: string
    ) => {
      const entry: QueryHistory = {
        id: `history_${Date.now()}`,
        userId: 'current_user', // TODO: Get from context
        connectionId,
        sql,
        aiPrompt,
        status,
        error,
        executionTime,
        rowsReturned,
        createdAt: new Date(),
      };

      setHistory((prev) => [entry, ...prev].slice(0, 100)); // Keep last 100

      // Save to localStorage for persistence
      try {
        const historyData = localStorage.getItem('insightengine_history');
        const parsed = historyData ? JSON.parse(historyData) : [];
        const updated = [entry, ...parsed].slice(0, 100);
        localStorage.setItem('insightengine_history', JSON.stringify(updated));
      } catch (e) {
        // Silently handle localStorage errors
      }
    },
    []
  );

  const deleteHistoryItem = useCallback((id: string) => {
    setHistory((prev) => {
      const updated = prev.filter((item) => item.id !== id);

      // Update localStorage
      try {
        localStorage.setItem('insightengine_history', JSON.stringify(updated));
      } catch (e) {
        // Silently handle localStorage errors
      }

      return updated;
    });
  }, []);

  const clearHistory = useCallback(() => {
    setHistory([]);
    localStorage.removeItem('insightengine_history');
  }, []);

  const loadHistory = useCallback(() => {
    try {
      const historyData = localStorage.getItem('insightengine_history');
      if (historyData) {
        const parsed = JSON.parse(historyData);
        setHistory(parsed);
        return parsed;
      }
    } catch (e) {
      // Silently handle localStorage errors
    }
    return [];
  }, []);

  const getRecentQueries = useCallback((limit: number = 10) => {
    return history.slice(0, limit);
  }, [history]);

  const getQueryStats = useCallback(() => {
    const successful = history.filter((h) => h.status === 'success').length;
    const failed = history.filter((h) => h.status === 'error').length;
    const avgExecutionTime =
      history.length > 0
        ? history.reduce((sum, h) => sum + h.executionTime, 0) / history.length
        : 0;

    return {
      total: history.length,
      successful,
      failed,
      avgExecutionTime: Math.round(avgExecutionTime),
    };
  }, [history]);

  return {
    history,
    addToHistory,
    clearHistory,
    deleteHistoryItem,
    loadHistory,
    getRecentQueries,
    getQueryStats,
  };
}
