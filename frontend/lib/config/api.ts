// API Configuration
// This file contains centralized API configuration for the frontend
// It defines base URLs and common headers for API requests

// Base URL for API requests - using environment variable or default to localhost
export const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

// WebSocket URL for real-time updates
export const WS_BASE_URL = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080/api/v1/ws';

// Common headers for API requests
export const API_HEADERS = {
  'Content-Type': 'application/json',
};

// API endpoints configuration
export const API_ENDPOINTS = {
  // Authentication endpoints
  AUTH: {
    LOGIN: '/api/auth/login',
    LOGOUT: '/api/auth/logout',
    ME: '/api/auth/me',
  },
  
  // Data connection endpoints
  CONNECTIONS: {
    BASE: '/api/go/connections',
    TEST: (id: string) => `/api/go/connections/${id}/test`,
    SCHEMA: (id: string) => `/api/go/connections/${id}/schema`,
  },
  
  // Dashboard endpoints
  DASHBOARDS: {
    BASE: '/api/go/dashboards',
    BY_ID: (id: string) => `/api/go/dashboards/${id}`,
  },
  
  // Notification endpoints
  NOTIFICATIONS: {
    BASE: '/api/go/notifications',
    UNREAD: '/api/go/notifications/unread',
    UNREAD_COUNT: '/api/go/notifications/unread-count',
  },
  
  // Query endpoints
  QUERIES: {
    BASE: '/api/go/queries',
    EXECUTE: '/api/go/queries/execute',
    SAVED: '/api/go/queries/saved',
  },
  
  // WebSocket endpoint
  WEBSOCKET: {
    BASE: '/api/v1/ws',
  },
};