"use client";

import { useEffect } from "react";

// This component is only active during development to clean up stale service workers.
// In production, service workers are managed by next-pwa automatically.
const IS_DEVELOPMENT = process.env.NODE_ENV === 'development';

export function ServiceWorkerReset() {
    useEffect(() => {
        // Only run in development to clean up mess
        if (!IS_DEVELOPMENT) return;

        if (typeof window !== "undefined" && "serviceWorker" in navigator) {
            // Silent cleanup - no console logs in development either
            navigator.serviceWorker.getRegistrations().then((registrations) => {
                for (const registration of registrations) {
                    registration.unregister().catch(() => {
                        // Silently ignore errors
                    });
                }
            }).catch(() => {
                // Silently ignore errors
            });

            // Also clear caches to prevent stale assets
            if ('caches' in window) {
                caches.keys().then((names) => {
                    for (const name of names) {
                        caches.delete(name).catch(() => {
                            // Silently ignore errors
                        });
                    }
                }).catch(() => {
                    // Silently ignore errors
                });
            }
        }
    }, []);

    return null;
}
