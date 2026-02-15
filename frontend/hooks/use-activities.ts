'use client';

import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useSession } from 'next-auth/react';
import { activityApi } from '@/lib/api/activities';
import { useWebSocket } from './use-websocket';
import type { ActivityWebSocketPayload } from '@/lib/types/notifications';
import { useCallback } from 'react';

export function useActivities(limit = 20, offset = 0) {
    const queryClient = useQueryClient();
    const { status } = useSession();
    const isAuthenticated = status === 'authenticated';

    // Query: Get user activity feed (only when authenticated)
    const { data, isLoading, error } = useQuery({
        queryKey: ['activities', 'user', limit, offset],
        queryFn: () => activityApi.getUserActivity(limit, offset),
        enabled: isAuthenticated,
        refetchOnWindowFocus: false,
    });

    // Handle real-time activity updates
    const handleActivityUpdate = useCallback((payload: ActivityWebSocketPayload) => {
        // Add new activity to the feed
        queryClient.setQueryData(['activities', 'user', limit, offset], (old: any) => {
            if (!old) return old;
            return {
                ...old,
                activities: [payload.activity, ...old.activities],
                total: old.total + 1,
            };
        });
    }, [queryClient, limit, offset]);

    // Connect to WebSocket for real-time updates (only when authenticated)
    useWebSocket({
        onActivity: handleActivityUpdate,
        enabled: isAuthenticated,
    });

    return {
        activities: data?.activities || [],
        total: data?.total || 0,
        isLoading: status === 'loading' || isLoading,
        error,
    };
}

export function useWorkspaceActivities(workspaceId: string, limit = 20, offset = 0) {
    const queryClient = useQueryClient();
    const { status } = useSession();
    const isAuthenticated = status === 'authenticated';

    // Query: Get workspace activity feed (only when authenticated and workspaceId provided)
    const { data, isLoading, error } = useQuery({
        queryKey: ['activities', 'workspace', workspaceId, limit, offset],
        queryFn: () => activityApi.getWorkspaceActivity(workspaceId, limit, offset),
        enabled: isAuthenticated && !!workspaceId,
        refetchOnWindowFocus: false,
    });

    // Handle real-time activity updates for workspace
    const handleActivityUpdate = useCallback((payload: ActivityWebSocketPayload) => {
        // Only update if activity belongs to this workspace
        if (payload.activity.workspaceId === workspaceId) {
            queryClient.setQueryData(['activities', 'workspace', workspaceId, limit, offset], (old: any) => {
                if (!old) return old;
                return {
                    ...old,
                    activities: [payload.activity, ...old.activities],
                    total: old.total + 1,
                };
            });
        }
    }, [queryClient, workspaceId, limit, offset]);

    useWebSocket({
        onActivity: handleActivityUpdate,
        enabled: isAuthenticated,
    });

    return {
        activities: data?.activities || [],
        total: data?.total || 0,
        isLoading: status === 'loading' || isLoading,
        error,
    };
}

export function useRecentActivities(limit = 50) {
    const { status } = useSession();
    const isAuthenticated = status === 'authenticated';

    // Query: Get recent activity (admin only, requires auth)
    const { data: activities = [], isLoading, error } = useQuery({
        queryKey: ['activities', 'recent', limit],
        queryFn: () => activityApi.getRecentActivity(limit),
        enabled: isAuthenticated,
        refetchOnWindowFocus: false,
    });

    return {
        activities,
        isLoading: status === 'loading' || isLoading,
        error,
    };
}
