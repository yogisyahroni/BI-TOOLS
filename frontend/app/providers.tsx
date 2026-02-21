"use client";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import { SessionProvider, useSession } from "next-auth/react";
import { useState, useEffect } from "react";
import { HelpProvider } from "@/components/providers/help-provider";
import { Toaster } from "@/components/ui/toaster";
import { useWorkspaceStore } from "@/stores/useWorkspaceStore";
import { useDatabaseStore } from "@/stores/useDatabaseStore";
import { useDuckDBStore } from "@/lib/store/duckdb-store";

function WorkspaceInitializer() {
  const { status } = useSession();
  const initialize = useWorkspaceStore((state) => state.initialize);

  useEffect(() => {
    if (status !== "loading") {
      initialize(status === "authenticated");
    }
  }, [status, initialize]);

  return null;
}

function DatabaseInitializer() {
  const { status } = useSession();
  const initialize = useDatabaseStore((state) => state.initialize);
  const fetchDatabases = useDatabaseStore((state) => state.fetchDatabases);

  useEffect(() => {
    if (status === "authenticated") {
      initialize(true);
      fetchDatabases();
      // Polling interval
      const interval = setInterval(() => {
        fetchDatabases();
      }, 30000);
      return () => clearInterval(interval);
    } else if (status === "unauthenticated") {
      initialize(false);
    }
  }, [status, initialize, fetchDatabases]);

  return null;
}

function DuckDBInitializer() {
  const initialize = useDuckDBStore((state) => state.initialize);

  useEffect(() => {
    // Initialize DuckDB-Wasm globally on app mount
    initialize();
  }, [initialize]);

  return null;
}

export function Providers({ children }: { children: React.ReactNode }) {
  const [queryClient] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            staleTime: 60 * 1000, // 1 minute
            gcTime: 5 * 60 * 1000, // 5 minutes (formerly cacheTime)
            retry: 1,
            refetchOnWindowFocus: true,
            refetchOnReconnect: true,
          },
          mutations: {
            retry: 0,
          },
        },
      }),
  );

  return (
    <SessionProvider
      basePath="/api/auth"
      refetchInterval={0} // Disable auto-refetching to prevent session endpoint calls
      refetchOnWindowFocus={false} // Disable refetch on window focus to prevent session calls
    >
      <QueryClientProvider client={queryClient}>
        <HelpProvider>
          <WorkspaceInitializer />
          <DatabaseInitializer />
          <DuckDBInitializer />
          {children}
          <Toaster />
        </HelpProvider>
        <ReactQueryDevtools initialIsOpen={false} />
      </QueryClientProvider>
    </SessionProvider>
  );
}
