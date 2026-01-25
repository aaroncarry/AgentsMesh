import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { useBrowserNotification } from "../useBrowserNotification";

// Mock Notification API
const mockNotification = vi.fn();
const mockClose = vi.fn();

class MockNotification {
  static permission: NotificationPermission = "default";
  static requestPermission = vi.fn();

  title: string;
  options: NotificationOptions;
  onclick: ((event: Event) => void) | null = null;
  close = mockClose;

  constructor(title: string, options?: NotificationOptions) {
    this.title = title;
    this.options = options || {};
    mockNotification(title, options);
  }
}

describe("useBrowserNotification", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockClose.mockClear();
    mockNotification.mockClear();

    // Setup default Notification mock
    MockNotification.permission = "default";
    MockNotification.requestPermission = vi.fn().mockResolvedValue("granted");

    // @ts-expect-error - mocking global Notification
    global.Notification = MockNotification;
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe("initial state", () => {
    it("should return default permission when Notification API is supported", () => {
      MockNotification.permission = "default";

      const { result } = renderHook(() => useBrowserNotification());

      expect(result.current.permission).toBe("default");
      expect(result.current.isSupported).toBe(true);
    });

    it("should return granted permission when already granted", () => {
      MockNotification.permission = "granted";

      const { result } = renderHook(() => useBrowserNotification());

      expect(result.current.permission).toBe("granted");
    });

    it("should return denied permission when denied", () => {
      MockNotification.permission = "denied";

      const { result } = renderHook(() => useBrowserNotification());

      expect(result.current.permission).toBe("denied");
    });

    it("should return unsupported when Notification API is not available", () => {
      // @ts-expect-error - removing Notification from window
      delete global.Notification;

      const { result } = renderHook(() => useBrowserNotification());

      expect(result.current.permission).toBe("unsupported");
      expect(result.current.isSupported).toBe(false);
    });
  });

  describe("requestPermission", () => {
    it("should request permission and return true when granted", async () => {
      MockNotification.permission = "default";
      MockNotification.requestPermission = vi.fn().mockResolvedValue("granted");

      const { result } = renderHook(() => useBrowserNotification());

      let granted: boolean = false;
      await act(async () => {
        granted = await result.current.requestPermission();
      });

      expect(granted).toBe(true);
      expect(MockNotification.requestPermission).toHaveBeenCalled();
    });

    it("should return false when permission is denied", async () => {
      MockNotification.permission = "default";
      MockNotification.requestPermission = vi.fn().mockResolvedValue("denied");

      const { result } = renderHook(() => useBrowserNotification());

      let granted: boolean = true;
      await act(async () => {
        granted = await result.current.requestPermission();
      });

      expect(granted).toBe(false);
    });

    it("should return true immediately if already granted", async () => {
      MockNotification.permission = "granted";

      const { result } = renderHook(() => useBrowserNotification());

      let granted: boolean = false;
      await act(async () => {
        granted = await result.current.requestPermission();
      });

      expect(granted).toBe(true);
      expect(MockNotification.requestPermission).not.toHaveBeenCalled();
    });

    it("should return false when Notification API is not supported", async () => {
      // @ts-expect-error - removing Notification from window
      delete global.Notification;

      const { result } = renderHook(() => useBrowserNotification());

      let granted: boolean = true;
      await act(async () => {
        granted = await result.current.requestPermission();
      });

      expect(granted).toBe(false);
    });

    it("should handle request permission error gracefully", async () => {
      MockNotification.permission = "default";
      MockNotification.requestPermission = vi.fn().mockRejectedValue(new Error("Permission error"));

      const consoleSpy = vi.spyOn(console, "error").mockImplementation(() => {});
      const { result } = renderHook(() => useBrowserNotification());

      let granted: boolean = true;
      await act(async () => {
        granted = await result.current.requestPermission();
      });

      expect(granted).toBe(false);
      expect(consoleSpy).toHaveBeenCalled();
      consoleSpy.mockRestore();
    });
  });

  describe("showNotification", () => {
    it("should create notification when permission is granted", () => {
      MockNotification.permission = "granted";

      const { result } = renderHook(() => useBrowserNotification());

      act(() => {
        result.current.showNotification({
          title: "Test Title",
          body: "Test Body",
        });
      });

      expect(mockNotification).toHaveBeenCalledWith("Test Title", expect.objectContaining({
        body: "Test Body",
      }));
    });

    it("should return null when permission is not granted", () => {
      MockNotification.permission = "default";

      const consoleSpy = vi.spyOn(console, "warn").mockImplementation(() => {});
      const { result } = renderHook(() => useBrowserNotification());

      let notification: Notification | null;
      act(() => {
        notification = result.current.showNotification({
          title: "Test Title",
        });
      });

      expect(notification!).toBeNull();
      expect(mockNotification).not.toHaveBeenCalled();
      consoleSpy.mockRestore();
    });

    it("should return null when Notification API is not supported", () => {
      // @ts-expect-error - removing Notification from window
      delete global.Notification;

      const consoleSpy = vi.spyOn(console, "warn").mockImplementation(() => {});
      const { result } = renderHook(() => useBrowserNotification());

      let notification: Notification | null;
      act(() => {
        notification = result.current.showNotification({
          title: "Test Title",
        });
      });

      expect(notification!).toBeNull();
      consoleSpy.mockRestore();
    });

    it("should set notification options correctly", () => {
      MockNotification.permission = "granted";

      const { result } = renderHook(() => useBrowserNotification());

      act(() => {
        result.current.showNotification({
          title: "Test Title",
          body: "Test Body",
          icon: "/custom-icon.png",
          tag: "test-tag",
          data: { podKey: "pod-123" },
        });
      });

      expect(mockNotification).toHaveBeenCalledWith("Test Title", expect.objectContaining({
        body: "Test Body",
        icon: "/custom-icon.png",
        tag: "test-tag",
        data: { podKey: "pod-123" },
      }));
    });

    it("should use default icon when not specified", () => {
      MockNotification.permission = "granted";

      const { result } = renderHook(() => useBrowserNotification());

      act(() => {
        result.current.showNotification({
          title: "Test Title",
        });
      });

      expect(mockNotification).toHaveBeenCalledWith("Test Title", expect.objectContaining({
        icon: "/icons/icon-192x192.png",
      }));
    });

    it("should call onClick handler when notification is clicked", () => {
      MockNotification.permission = "granted";
      const onClick = vi.fn();
      const mockFocus = vi.fn();
      global.window.focus = mockFocus;

      const { result } = renderHook(() => useBrowserNotification());

      let notification: MockNotification | null = null;
      act(() => {
        notification = result.current.showNotification({
          title: "Test Title",
          onClick,
        }) as unknown as MockNotification;
      });

      expect(notification).not.toBeNull();

      // Simulate click
      const mockEvent = { preventDefault: vi.fn() } as unknown as Event;
      act(() => {
        notification!.onclick?.(mockEvent);
      });

      expect(mockEvent.preventDefault).toHaveBeenCalled();
      expect(mockFocus).toHaveBeenCalled();
      expect(onClick).toHaveBeenCalled();
      expect(mockClose).toHaveBeenCalled();
    });

    it("should auto-close notification after timeout", () => {
      vi.useFakeTimers();
      MockNotification.permission = "granted";

      const { result } = renderHook(() => useBrowserNotification());

      act(() => {
        result.current.showNotification({
          title: "Test Title",
        });
      });

      expect(mockClose).not.toHaveBeenCalled();

      act(() => {
        vi.advanceTimersByTime(5000);
      });

      expect(mockClose).toHaveBeenCalled();
      vi.useRealTimers();
    });
  });
});
