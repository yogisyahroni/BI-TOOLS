'use client';

import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';
import { SessionProvider } from 'next-auth/react';
import { useState } from 'react';
import { HelpProvider } from '@/components/providers/help-provider';

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
            })
    );

    return (
        <SessionProvider 
            basePath="/api/auth"
            refetchInterval={0}  // Disable auto-refetching to prevent session endpoint calls
            refetchOnWindowFocus={false}  // Disable refetch on window focus to prevent session calls
        >
            <QueryClientProvider client={queryClient}>
                <HelpProvider>
                    {children}
                </HelpProvider>
                <ReactQueryDevtools initialIsOpen={false} />
            </QueryClientProvider>
        </SessionProvider>
    );
}


