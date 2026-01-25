"use client";

import { useCallback, useState, useSyncExternalStore } from "react";

export interface BrowserNotificationOptions {
  title: string;
  body?: string;
  icon?: string;
  tag?: string;
  data?: Record<string, unknown>;
  onClick?: () => void;
}

interface UseBrowserNotificationReturn {
  permission: NotificationPermission | "unsupported";
  isSupported: boolean;
  requestPermission: () => Promise<boolean>;
  showNotification: (options: BrowserNotificationOptions) => Notification | null;
}

// Check if Notification API is supported
function getIsSupported(): boolean {
  if (typeof window === "undefined") return false;
  return "Notification" in window;
}

// Get current permission status
function getPermission(): NotificationPermission | "unsupported" {
  if (typeof window === "undefined" || !("Notification" in window)) {
    return "unsupported";
  }
  return Notification.permission;
}

// Server-side snapshot
function getServerSnapshot(): NotificationPermission | "unsupported" {
  return "default";
}

// Subscribe to permission changes (via visibility change as a proxy)
function subscribeToPermissionChanges(callback: () => void): () => void {
  // There's no direct API to listen for permission changes,
  // but we can use visibility change as a proxy since permission
  // dialogs typically happen when the page is visible
  document.addEventListener("visibilitychange", callback);
  return () => document.removeEventListener("visibilitychange", callback);
}

/**
 * Hook for browser native notifications (Notification API)
 *
 * Unlike push notifications that require a service worker,
 * this uses the Notification API directly for local notifications.
 */
export function useBrowserNotification(): UseBrowserNotificationReturn {
  // Use useSyncExternalStore for permission to avoid hydration issues
  const permission = useSyncExternalStore(
    subscribeToPermissionChanges,
    getPermission,
    getServerSnapshot
  );

  // Track support status with lazy initialization
  const [isSupported] = useState(getIsSupported);

  // Request notification permission
  const requestPermission = useCallback(async (): Promise<boolean> => {
    if (!getIsSupported()) {
      console.warn("[BrowserNotification] Notifications not supported");
      return false;
    }

    if (Notification.permission === "granted") {
      return true;
    }

    try {
      const result = await Notification.requestPermission();
      return result === "granted";
    } catch (error) {
      console.error("[BrowserNotification] Failed to request permission:", error);
      return false;
    }
  }, []);

  // Show a browser notification
  const showNotification = useCallback(
    (options: BrowserNotificationOptions): Notification | null => {
      if (!getIsSupported()) {
        console.warn("[BrowserNotification] Notifications not supported");
        return null;
      }

      if (Notification.permission !== "granted") {
        console.warn("[BrowserNotification] Permission not granted");
        return null;
      }

      try {
        const notification = new Notification(options.title, {
          body: options.body,
          icon: options.icon || "/icons/icon-192x192.png",
          tag: options.tag,
          data: options.data,
          requireInteraction: false,
        });

        if (options.onClick) {
          notification.onclick = (event) => {
            event.preventDefault();
            // Focus the window when notification is clicked
            window.focus();
            options.onClick?.();
            notification.close();
          };
        }

        // Auto-close after 5 seconds if no interaction
        setTimeout(() => {
          notification.close();
        }, 5000);

        return notification;
      } catch (error) {
        console.error("[BrowserNotification] Failed to show notification:", error);
        return null;
      }
    },
    []
  );

  return {
    permission,
    isSupported,
    requestPermission,
    showNotification,
  };
}
