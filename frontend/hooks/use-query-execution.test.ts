import { renderHook, act, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { useQueryExecution } from "./use-query-execution";
import * as utils from "@/lib/utils";

// Mock fetchWithAuth
vi.mock("@/lib/utils", () => ({
  fetchWithAuth: vi.fn(),
}));

describe("useQueryExecution", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("should initialize with default state", () => {
    const { result } = renderHook(() => useQueryExecution());

    expect(result.current.isLoading).toBe(false);
    expect(result.current.isExecuting).toBe(false);
    expect(result.current.data).toBe(null);
    expect(result.current.columns).toBe(null);
    expect(result.current.rowCount).toBe(0);
    expect(result.current.error).toBe(null);
  });

  it("should handle successful query execution", async () => {
    const mockData = [{ id: 1, name: "Test" }];
    const mockColumns = ["id", "name"];
    const mockResponse = {
      ok: true,
      json: async () => ({
        success: true,
        data: mockData,
        columns: mockColumns,
        rowCount: 1,
        executionTime: 100,
        totalRows: 10,
      }),
    };

    vi.spyOn(utils, "fetchWithAuth").mockResolvedValue(mockResponse as Response);

    const { result } = renderHook(() => useQueryExecution());

    await act(async () => {
      await result.current.execute({
        sql: "SELECT * FROM users",
        connectionId: "db1",
      });
    });

    expect(result.current.isLoading).toBe(false);
    expect(result.current.isExecuting).toBe(false);
    expect(result.current.data).toEqual(mockData);
    expect(result.current.columns).toEqual(mockColumns);
    expect(result.current.rowCount).toBe(1);
    expect(result.current.error).toBe(null);
  });

  it("should handle API errors", async () => {
    const mockResponse = {
      ok: false,
      status: 500,
      statusText: "Internal Server Error",
    };

    vi.spyOn(utils, "fetchWithAuth").mockResolvedValue(mockResponse as Response);

    const { result } = renderHook(() => useQueryExecution());

    await act(async () => {
      await result.current.execute({
        sql: "SELECT * FROM users",
        connectionId: "db1",
      });
    });

    expect(result.current.isLoading).toBe(false);
    expect(result.current.error).toContain("HTTP error! status: 500");
  });

  it("should handle manual error throwing from API success:false", async () => {
    const mockResponse = {
      ok: true,
      json: async () => ({
        success: false,
        error: "SQL Syntax Error",
      }),
    };

    vi.spyOn(utils, "fetchWithAuth").mockResolvedValue(mockResponse as Response);

    const { result } = renderHook(() => useQueryExecution());

    await act(async () => {
      await result.current.execute({
        sql: "SELECT * FROM users",
        connectionId: "db1",
      });
    });

    expect(result.current.error).toBe("SQL Syntax Error");
  });

  it("should update pagination state on execution", async () => {
    const mockResponse = {
      ok: true,
      json: async () => ({
        success: true,
        data: [],
        columns: [],
        rowCount: 0,
        executionTime: 10,
        totalRows: 100,
      }),
    };

    vi.spyOn(utils, "fetchWithAuth").mockResolvedValue(mockResponse as Response);

    const { result } = renderHook(() => useQueryExecution());

    await act(async () => {
      await result.current.execute({
        sql: "SELECT * FROM users",
        connectionId: "db1",
        page: 2,
        pageSize: 25,
      });
    });

    expect(result.current.pagination.page).toBe(2);
    expect(result.current.pagination.pageSize).toBe(25);
    expect(result.current.pagination.totalRows).toBe(100);
  });
});
