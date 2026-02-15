'use client';

import { createContext, useContext } from 'react';
import type {
    WebSocketState,
    NotificationWebSocketPayload,
    ActivityWebSocketPayload,
    SystemWebSocketPayload
} from '@/lib/types/notifications';

export type WebSocketEventType = 'notification' | 'activity' | 'system';

export type WebSocketEventHandler<T = any> = (payload: T) => void;

export interface WebSocketContextType extends WebSocketState {
    subscribe: (type: WebSocketEventType, handler: WebSocketEventHandler) => () => void;
}

export const WebSocketContext = createContext<WebSocketContextType | null>(null);

export function useWebSocketContext() {
    const context = useContext(WebSocketContext);
    if (!context) {
        throw new Error('useWebSocketContext must be used within a WebSocketProvider');
    }
    return context;
}
