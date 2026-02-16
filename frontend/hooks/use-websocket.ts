'use client';

import { useEffect, useRef, useState, useCallback } from 'react';
import type {
    WebSocketMessage,
    WebSocketState,
    NotificationWebSocketPayload,
    ActivityWebSocketPayload,
    SystemWebSocketPayload,
} from '@/lib/types/notifications';

interface UseWebSocketOptions {
    onNotification?: (payload: NotificationWebSocketPayload) => void;
    onActivity?: (payload: ActivityWebSocketPayload) => void;
    onSystem?: (payload: SystemWebSocketPayload) => void;
    onConnect?: () => void;
    onDisconnect?: () => void;
    onError?: (error: Event) => void;
    autoReconnect?: boolean;
    reconnectInterval?: number;
    /** Enable/disable WebSocket connection (useful for auth gating) */
    enabled?: boolean;
    /** Auth token for query param authentication */
    token?: string;
}

export function useWebSocket(options: UseWebSocketOptions = {}) {
    const {
        onNotification,
        onActivity,
        onSystem,
        onConnect,
        onDisconnect,
        onError,
        autoReconnect = true,
        reconnectInterval = 3000,
        enabled = true,
        token,
    } = options;

    const [state, setState] = useState<WebSocketState>({
        connected: false,
        connecting: false,
    });

    const wsRef = useRef<WebSocket | null>(null);
    const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
    const reconnectAttemptsRef = useRef(0);
    const enabledRef = useRef(enabled);

    // Keep enabled ref in sync
    useEffect(() => {
        enabledRef.current = enabled;
    }, [enabled]);

    // Use refs for handlers to prevent unnecessary re-connections when handlers change
    const onNotificationRef = useRef(onNotification);
    const onActivityRef = useRef(onActivity);
    const onSystemRef = useRef(onSystem);
    const onConnectRef = useRef(onConnect);
    const onDisconnectRef = useRef(onDisconnect);
    const onErrorRef = useRef(onError);

    // Update refs when props change
    useEffect(() => {
        onNotificationRef.current = onNotification;
        onActivityRef.current = onActivity;
        onSystemRef.current = onSystem;
        onConnectRef.current = onConnect;
        onDisconnectRef.current = onDisconnect;
        onErrorRef.current = onError;
    }, [onNotification, onActivity, onSystem, onConnect, onDisconnect, onError]);

    const connect = useCallback(() => {
        // Don't connect if disabled
        if (!enabledRef.current) {
            return;
        }

        // Prevent multiple connections
        if (wsRef.current) {
            if (wsRef.current.readyState === WebSocket.OPEN || wsRef.current.readyState === WebSocket.CONNECTING) {
                return;
            }
        }

        setState(prev => ({ ...prev, connecting: true, error: undefined }));

        try {
            // WebSocket must connect directly to Go backend (Next.js rewrites don't support WS)
            // Use environment variable or default to localhost:8080
            const backendHost = process.env.NEXT_PUBLIC_API_URL?.replace(/^https?:\/\//, '') || 'localhost:8080';
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            // Use the correct WebSocket endpoint with the environment variable
            let wsUrl = process.env.NEXT_PUBLIC_WS_URL || `${protocol}//${backendHost}/api/v1/ws`;

            if (token) {
                const separator = wsUrl.includes('?') ? '&' : '?';
                wsUrl = `${wsUrl}${separator}token=${token}`;
            }

            // Create WebSocket instance
            console.warn('[useWebSocket] Connecting to:', wsUrl);
            const ws = new WebSocket(wsUrl);

            ws.onopen = () => {
                setState({ connected: true, connecting: false });
                reconnectAttemptsRef.current = 0;
                onConnectRef.current?.();
            };

            ws.onmessage = (event) => {
                try {
                    const message: WebSocketMessage = JSON.parse(event.data);

                    switch (message.type) {
                        case 'notification':
                            onNotificationRef.current?.(message.payload as NotificationWebSocketPayload);
                            break;
                        case 'activity':
                            onActivityRef.current?.(message.payload as ActivityWebSocketPayload);
                            break;
                        case 'system':
                            onSystemRef.current?.(message.payload as SystemWebSocketPayload);
                            break;
                        default:
                        // Unknown message type - silently ignore
                    }
                } catch (error) {
                    // Silently handle parse errors
                }
            };

            ws.onerror = (error) => {
                setState(prev => ({
                    ...prev,
                    connecting: false,
                    error: 'WebSocket connection error',
                }));
                onErrorRef.current?.(error);
            };

            ws.onclose = () => {
                setState({ connected: false, connecting: false });
                wsRef.current = null;
                onDisconnectRef.current?.();

                // Auto-reconnect with exponential backoff, max 5 attempts
                // Only reconnect if still enabled and online
                if (autoReconnect && enabledRef.current && navigator.onLine) {
                    if (reconnectAttemptsRef.current >= 5) {
                        console.warn('WebSocket reconnection limit reached (5 attempts). Stopping.');
                        // Reset attempts after a long delay (optional) or just stop
                        return;
                    }

                    reconnectAttemptsRef.current++;
                    // 1s, 2s, 4s, 8s, 16s...
                    const delay = Math.min(1000 * Math.pow(2, reconnectAttemptsRef.current), 30000);

                    console.warn(`WebSocket reconnecting in ${delay}ms (Attempt ${reconnectAttemptsRef.current}/5)`);

                    reconnectTimeoutRef.current = setTimeout(() => {
                        connect();
                    }, delay);
                }
            };

            wsRef.current = ws;
        } catch (error) {
            setState({
                connected: false,
                connecting: false,
                error: error instanceof Error ? error.message : 'Failed to connect',
            });
            // Also trigger reconnect on initial connection fail
            if (autoReconnect && enabledRef.current && reconnectAttemptsRef.current < 5) {
                reconnectAttemptsRef.current++;
                const delay = Math.min(1000 * Math.pow(2, reconnectAttemptsRef.current), 30000);
                reconnectTimeoutRef.current = setTimeout(() => {
                    connect();
                }, delay);
            }
        }
    }, [autoReconnect, reconnectInterval, token]);

    const disconnect = useCallback(() => {
        if (reconnectTimeoutRef.current) {
            clearTimeout(reconnectTimeoutRef.current);
            reconnectTimeoutRef.current = null;
        }

        if (wsRef.current) {
            wsRef.current.close();
            wsRef.current = null;
        }

        setState({ connected: false, connecting: false });
    }, []);

    const send = useCallback((message: any) => {
        if (wsRef.current?.readyState === WebSocket.OPEN) {
            wsRef.current.send(JSON.stringify(message));
        }
    }, []);

    // Connect when enabled, disconnect when disabled
    useEffect(() => {
        if (enabled) {
            connect();
        } else {
            disconnect();
        }

        return () => {
            disconnect();
        };
    }, [enabled, connect, disconnect]);

    return {
        ...state,
        connect,
        disconnect,
        send,
    };
}
