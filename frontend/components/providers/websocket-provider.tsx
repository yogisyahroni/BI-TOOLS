"use client";

import { useSession } from "next-auth/react";
import { useState, useEffect, useRef, useCallback } from "react";
import { useWebSocket } from "@/hooks/use-websocket";
import {
  WebSocketContext,
  type WebSocketEventType,
  type WebSocketEventHandler,
} from "@/lib/websocket-context";
import type {
  NotificationWebSocketPayload,
  ActivityWebSocketPayload,
  SystemWebSocketPayload,
} from "@/lib/types/notifications";

/**
 * WebSocketProvider - Initializes global WebSocket connection
 *
 * This component establishes a WebSocket connection ONLY after user is authenticated.
 * It handles:
 * - Auto-connect on mount (only if session exists)
 * - Auto-reconnect on connection loss
 * - Real-time message routing (notifications, activity, system) via Event Bus
 */
export function WebSocketProvider({ children }: { children: React.ReactNode }) {
  const { status } = useSession();
  const [token, setToken] = useState<string | undefined>(undefined);

  // Event Bus: Map of event types to Set of handlers
  const listenersRef = useRef<Map<WebSocketEventType, Set<WebSocketEventHandler>>>(new Map());

  // Initialize listeners map
  if (listenersRef.current.size === 0) {
    listenersRef.current.set("notification", new Set());
    listenersRef.current.set("activity", new Set());
    listenersRef.current.set("system", new Set());
  }

  const subscribe = useCallback((type: WebSocketEventType, handler: WebSocketEventHandler) => {
    const listeners = listenersRef.current.get(type);
    if (listeners) {
      listeners.add(handler);
    }

    // Return unsubscribe function
    return () => {
      const listeners = listenersRef.current.get(type);
      if (listeners) {
        listeners.delete(handler);
      }
    };
  }, []);

  // Dispatch events to listeners
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const dispatch = useCallback((type: WebSocketEventType, payload: any) => {
    const listeners = listenersRef.current.get(type);
    if (listeners) {
      listeners.forEach((handler) => {
        try {
          handler(payload);
        } catch (err) {
          console.error(`Error in WebSocket ${type} handler:`, err);
        }
      });
    }
  }, []);

  useEffect(() => {
    if (status === "authenticated") {
      fetch("/api/auth/token")
        .then((res) => res.json())
        .then((data) => {
          if (data.token) {
            setToken(data.token);
          }
        })
        .catch((err) => console.error("Failed to fetch WS token:", err));
    } else {
      setToken(undefined);
    }
  }, [status]);

  // Only initialize WebSocket when token is available
  const isAuthenticated = status === "authenticated" && !!token;

  const wsState = useWebSocket({
    enabled: isAuthenticated,
    autoReconnect: true,
    reconnectInterval: 3000,
    token: token,
    onNotification: (payload) => dispatch("notification", payload),
    onActivity: (payload) => dispatch("activity", payload),
    onSystem: (payload) => dispatch("system", payload),
  });

  return (
    <WebSocketContext.Provider value={{ ...wsState, subscribe }}>
      {children}
    </WebSocketContext.Provider>
  );
}
