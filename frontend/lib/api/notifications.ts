import type {
    Notification,
    NotificationResponse,
    UnreadCountResponse,
    CreateNotificationInput,
} from '@/lib/types/notifications';
import { fetchWithAuth } from '@/lib/utils';

export const notificationApi = {
    // Get user notifications (paginated)
    getNotifications: async (limit = 20, offset = 0): Promise<NotificationResponse> => {
        const res = await fetchWithAuth(`/api/go/notifications?limit=${limit}&offset=${offset}`, {
            credentials: 'include',
        });
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to fetch notifications');
        }
        return res.json();
    },

    // Get unread notifications
    getUnreadNotifications: async (): Promise<Notification[]> => {
        const res = await fetchWithAuth(`/api/go/notifications/unread`, {
            credentials: 'include',
        });
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to fetch unread notifications');
        }
        return res.json();
    },

    // Get unread count
    getUnreadCount: async (): Promise<UnreadCountResponse> => {
        const res = await fetchWithAuth(`/api/go/notifications/unread-count`, {
            credentials: 'include',
        });
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to fetch unread count');
        }
        return res.json();
    },

    // Mark notification as read
    markAsRead: async (id: string): Promise<{ message: string }> => {
        const res = await fetchWithAuth(`/api/go/notifications/${id}/read`, {
            method: 'PUT',
            credentials: 'include',
        });
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to mark notification as read');
        }
        return res.json();
    },

    // Mark all as read
    markAllAsRead: async (): Promise<{ message: string }> => {
        const res = await fetchWithAuth(`/api/go/notifications/read-all`, {
            method: 'PUT',
            credentials: 'include',
        });
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to mark all as read');
        }
        return res.json();
    },

    // Delete notification
    deleteNotification: async (id: string): Promise<{ message: string }> => {
        const res = await fetchWithAuth(`/api/go/notifications/${id}`, {
            method: 'DELETE',
            credentials: 'include',
        });
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to delete notification');
        }
        return res.json();
    },

    // Delete all read notifications
    deleteReadNotifications: async (): Promise<{ message: string }> => {
        const res = await fetchWithAuth(`/api/go/notifications/read`, {
            method: 'DELETE',
            credentials: 'include',
        });
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to delete read notifications');
        }
        return res.json();
    },

    // Create notification (admin only)
    createNotification: async (data: CreateNotificationInput): Promise<Notification> => {
        const res = await fetchWithAuth(`/api/go/notifications`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            credentials: 'include',
            body: JSON.stringify(data),
        });
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to create notification');
        }
        return res.json();
    },

    // Broadcast system notification (admin only)
    broadcastSystemNotification: async (title: string, message: string): Promise<{ message: string }> => {
        const res = await fetchWithAuth(`/api/go/notifications/broadcast`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            credentials: 'include',
            body: JSON.stringify({ title, message }),
        });
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to broadcast notification');
        }
        return res.json();
    },
};
