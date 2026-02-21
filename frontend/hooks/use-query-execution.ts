"use client";

import { useState, useCallback } from "react";
import { useMutation } from "@tanstack/react-query";
import { fetchWithAuth } from "@/lib/utils";
// eslint-disable-next-line @typescript-eslint/no-unused-vars
import { type QueryResult } from "@/lib/types";
import { useDuckDBStore } from "@/lib/store/duckdb-store";

interface ExecuteOptions {
  sql: string;
  connectionId: string;
  aiPrompt?: string;
  limit?: number;
  page?: number; // 1-indexed
  pageSize?: number; // default 50
}

interface PaginationState {
  page: number;
  pageSize: number;
  totalRows: number;
}

interface QueryState {
  data: Record<string, any>[] | null;
  columns: string[] | null;
  rowCount: number;
  executionTime: number;
  pagination: PaginationState;
}

export function useQueryExecution() {
  const [lastOptions, setLastOptions] = useState<ExecuteOptions | null>(null);
  const [queryState, setQueryState] = useState<QueryState>({
    data: null,
    columns: null,
    rowCount: 0,
    executionTime: 0,
    pagination: {
      page: 1,
      pageSize: 50,
      totalRows: 0,
    },
  });

  const executeMutation = useMutation({
    mutationFn: async (options: ExecuteOptions) => {
      const startTime = performance.now();
      const response = await fetchWithAuth("/api/go/queries/execute?format=arrow", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          sql: options.sql,
          connectionId: options.connectionId,
          aiPrompt: options.aiPrompt,
          limit: options.limit || 1000,
          page: options.page || 1,
          pageSize: options.pageSize || 50,
        }),
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const contentType = response.headers.get("Content-Type");

      let finalData: Record<string, any>[] = [];
      let finalColumns: string[] = [];
      let rowCount = 0;
      let totalRows = 0;

      if (contentType?.includes("application/vnd.apache.arrow.stream")) {
        // Handle Arrow IPC Stream
        const arrayBuffer = await response.arrayBuffer();
        const buffer = new Uint8Array(arrayBuffer);

        // Ingest into DuckDB Wasm
        const tableName = `query_${Date.now()}`;
        const duckdbStore = useDuckDBStore.getState();
        await duckdbStore.ingestArrowBuffer(tableName, buffer);

        // Query it out to update the UI State
        finalData = await duckdbStore.query(`SELECT * FROM "${tableName}"`);
        rowCount = finalData.length;
        totalRows = rowCount; // For ad-hoc, total is usually what's returned

        if (finalData.length > 0) {
          finalColumns = Object.keys(finalData[0]);
        }
      } else {
        // Fallback to standard JSON JSON
        const result = (await response.json()) as {
          success: boolean;
          data?: Record<string, any>[];
          columns?: string[];
          rowCount: number;
          error?: string;
          totalRows?: number;
        };

        if (!result.success) {
          throw new Error(result.error || "Query execution failed");
        }

        finalData = result.data || [];
        finalColumns = result.columns || [];
        rowCount = result.rowCount;
        totalRows = result.totalRows || rowCount;
      }

      const executionTime = Math.round(performance.now() - startTime);

      return {
        options,
        result: {
          data: finalData,
          columns: finalColumns,
          rowCount,
          totalRows,
          executionTime,
        },
      };
    },
    onSuccess: ({ options, result }) => {
      setQueryState({
        data: result.data || null,
        columns: result.columns || null,
        rowCount: result.rowCount,
        executionTime: result.executionTime,
        pagination: {
          page: options.page || 1,
          pageSize: options.pageSize || 50,
          totalRows: result.totalRows || result.rowCount,
        },
      });
      setLastOptions({ ...options, page: options.page || 1, pageSize: options.pageSize || 50 });
    },
  });

  const clearResults = useCallback(() => {
    executeMutation.reset();
    setQueryState({
      data: null,
      columns: null,
      rowCount: 0,
      executionTime: 0,
      pagination: {
        page: 1,
        pageSize: 50,
        totalRows: 0,
      },
    });
  }, [executeMutation]);

  const setPage = useCallback(
    (newPage: number) => {
      if (lastOptions) {
        executeMutation.mutate({ ...lastOptions, page: newPage });
      }
    },
    [executeMutation, lastOptions],
  );

  const setPageSize = useCallback(
    (newSize: number) => {
      if (lastOptions) {
        executeMutation.mutate({ ...lastOptions, pageSize: newSize, page: 1 });
      }
    },
    [executeMutation, lastOptions],
  );

  const execute = useCallback(
    async (options: ExecuteOptions) => {
      try {
        const { result } = await executeMutation.mutateAsync(options);
        return {
          success: true,
          data: result.data,
          columns: result.columns,
          rowCount: result.rowCount,
          executionTime: result.executionTime,
        };
      } catch (error) {
        return {
          success: false,
          error: error instanceof Error ? error.message : "Unknown error",
        };
      }
    },
    [executeMutation],
  );

  return {
    isLoading: executeMutation.isPending,
    isExecuting: executeMutation.isPending,
    error: executeMutation.error ? executeMutation.error.message : null,
    data: queryState.data,
    columns: queryState.columns,
    rowCount: queryState.rowCount,
    executionTime: queryState.executionTime,
    pagination: queryState.pagination,
    execute,
    clearResults,
    setPage,
    setPageSize,
  };
}
