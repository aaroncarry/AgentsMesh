import { describe, it, expect, beforeEach, vi } from "vitest";
import { act } from "@testing-library/react";
import { useChannelStore, useChannelMessageStore } from "../channel";
import { EMPTY_CACHE } from "../channelMessageStore";
import type { ChannelMessage } from "@/lib/api";

type Message = ChannelMessage;

vi.mock("@/lib/api", () => ({
  channelApi: {
    list: vi.fn(),
    get: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    archive: vi.fn(),
    unarchive: vi.fn(),
    getMessages: vi.fn(),
    sendMessage: vi.fn(),
    joinPod: vi.fn(),
    leavePod: vi.fn(),
    markRead: vi.fn(),
  },
}));

import { channelApi } from "@/lib/api";

const mockMessage: Message = {
  id: 1, channel_id: 1, content: "Hello, world!", message_type: "text", created_at: "2024-01-01T00:00:00Z",
};

describe("Channel Message Store", () => {
  const CH = 1;

  beforeEach(() => {
    vi.clearAllMocks();
    useChannelMessageStore.setState({ cache: {}, unreadCounts: {} });
    useChannelStore.setState({ currentChannel: null });
  });

  describe("initial state", () => {
    it("should have default values", () => {
      const s = useChannelMessageStore.getState();
      expect(s.cache).toEqual({});
      expect(s.unreadCounts).toEqual({});
    });

    it("should return EMPTY_CACHE for unvisited channels", () => {
      const cache = useChannelMessageStore.getState().cache[999] ?? EMPTY_CACHE;
      expect(cache.messages).toEqual([]);
      expect(cache.hasMore).toBe(false);
    });
  });

  describe("fetchMessages", () => {
    it("should populate cache", async () => {
      vi.mocked(channelApi.getMessages).mockResolvedValue({ messages: [mockMessage], has_more: true });
      await act(async () => { await useChannelMessageStore.getState().fetchMessages(CH); });
      const c = useChannelMessageStore.getState().cache[CH];
      expect(c.messages).toHaveLength(1);
      expect(c.hasMore).toBe(true);
      expect(c.loading).toBe(false);
    });

    it("should prepend on load-more (beforeId)", async () => {
      const existing = { ...mockMessage, id: 10, content: "Existing" };
      useChannelMessageStore.setState({
        cache: { [CH]: { messages: [existing], hasMore: true, loading: false, loadingMore: false, error: null } },
      });
      vi.mocked(channelApi.getMessages).mockResolvedValue({ messages: [{ ...mockMessage, id: 5, content: "Older" }], has_more: false });
      await act(async () => { await useChannelMessageStore.getState().fetchMessages(CH, 50, 10); });
      const c = useChannelMessageStore.getState().cache[CH];
      expect(c.messages).toHaveLength(2);
      expect(c.messages[0].content).toBe("Older");
      expect(c.messages[1].content).toBe("Existing");
      expect(c.hasMore).toBe(false);
    });

    it("should replace on refresh (no beforeId)", async () => {
      useChannelMessageStore.setState({
        cache: { [CH]: { messages: [{ ...mockMessage, content: "Old" }], hasMore: false, loading: false, loadingMore: false, error: null } },
      });
      vi.mocked(channelApi.getMessages).mockResolvedValue({ messages: [mockMessage], has_more: false });
      await act(async () => { await useChannelMessageStore.getState().fetchMessages(CH); });
      expect(useChannelMessageStore.getState().cache[CH].messages[0].content).toBe("Hello, world!");
    });

    it("should isolate channels", async () => {
      useChannelMessageStore.setState({
        cache: { 2: { messages: [{ ...mockMessage, id: 99, channel_id: 2, content: "Ch2" }], hasMore: false, loading: false, loadingMore: false, error: null } },
      });
      vi.mocked(channelApi.getMessages).mockResolvedValue({ messages: [mockMessage], has_more: false });
      await act(async () => { await useChannelMessageStore.getState().fetchMessages(CH); });
      expect(useChannelMessageStore.getState().cache[CH].messages).toHaveLength(1);
      expect(useChannelMessageStore.getState().cache[2].messages[0].content).toBe("Ch2");
    });

    it("should handle error", async () => {
      vi.mocked(channelApi.getMessages).mockRejectedValue({ message: "Fail" });
      await act(async () => { await useChannelMessageStore.getState().fetchMessages(CH); });
      expect((useChannelMessageStore.getState().cache[CH] ?? EMPTY_CACHE).loading).toBe(false);
    });
  });

  describe("sendMessage", () => {
    it("should send and add to cache", async () => {
      vi.mocked(channelApi.sendMessage).mockResolvedValue({ message: mockMessage });
      let result: Message;
      await act(async () => { result = await useChannelMessageStore.getState().sendMessage(CH, "Hello, world!"); });
      expect(result!).toEqual(mockMessage);
      expect(useChannelMessageStore.getState().cache[CH].messages).toContainEqual(mockMessage);
    });

    it("should pass podKey", async () => {
      vi.mocked(channelApi.sendMessage).mockResolvedValue({ message: mockMessage });
      await act(async () => { await useChannelMessageStore.getState().sendMessage(CH, "Hello", "pod-123"); });
      expect(channelApi.sendMessage).toHaveBeenCalledWith(CH, "Hello", "pod-123", undefined, undefined);
    });

    it("should handle error", async () => {
      vi.mocked(channelApi.sendMessage).mockRejectedValue({ message: "Send failed" });
      await expect(act(async () => { await useChannelMessageStore.getState().sendMessage(CH, "Hello"); })).rejects.toEqual({ message: "Send failed" });
    });
  });

  describe("addMessage", () => {
    it("should add to channel cache", () => {
      act(() => { useChannelMessageStore.getState().addMessage(CH, mockMessage); });
      expect(useChannelMessageStore.getState().cache[CH].messages).toHaveLength(1);
    });

    it("should append to existing", () => {
      useChannelMessageStore.setState({
        cache: { [CH]: { messages: [{ ...mockMessage, id: 0, content: "First" }], hasMore: false, loading: false, loadingMore: false, error: null } },
      });
      act(() => { useChannelMessageStore.getState().addMessage(CH, mockMessage); });
      expect(useChannelMessageStore.getState().cache[CH].messages).toHaveLength(2);
    });

    it("should deduplicate", () => {
      useChannelMessageStore.setState({
        cache: { [CH]: { messages: [mockMessage], hasMore: false, loading: false, loadingMore: false, error: null } },
      });
      act(() => { useChannelMessageStore.getState().addMessage(CH, { ...mockMessage }); });
      expect(useChannelMessageStore.getState().cache[CH].messages).toHaveLength(1);
    });

    it("should isolate channels", () => {
      useChannelMessageStore.setState({
        cache: { 2: { messages: [], hasMore: false, loading: false, loadingMore: false, error: null } },
      });
      act(() => { useChannelMessageStore.getState().addMessage(CH, mockMessage); });
      expect(useChannelMessageStore.getState().cache[CH].messages).toHaveLength(1);
      expect(useChannelMessageStore.getState().cache[2].messages).toHaveLength(0);
    });
  });

  describe("updateMessage", () => {
    it("should update content", () => {
      useChannelMessageStore.setState({
        cache: { [CH]: { messages: [mockMessage], hasMore: false, loading: false, loadingMore: false, error: null } },
      });
      act(() => { useChannelMessageStore.getState().updateMessage(CH, { id: 1, content: "Updated", edited_at: "2024-01-02T00:00:00Z" }); });
      expect(useChannelMessageStore.getState().cache[CH].messages[0].content).toBe("Updated");
    });
  });

  describe("removeMessage", () => {
    it("should remove from cache", () => {
      useChannelMessageStore.setState({
        cache: { [CH]: { messages: [mockMessage], hasMore: false, loading: false, loadingMore: false, error: null } },
      });
      act(() => { useChannelMessageStore.getState().removeMessage(CH, 1); });
      expect(useChannelMessageStore.getState().cache[CH].messages).toHaveLength(0);
    });
  });

  describe("unread counts", () => {
    it("should increment", () => {
      act(() => { useChannelMessageStore.getState().incrementUnread(CH); useChannelMessageStore.getState().incrementUnread(CH); });
      expect(useChannelMessageStore.getState().unreadCounts[CH]).toBe(2);
    });

    it("should clear via clearChannelUnread", () => {
      act(() => {
        useChannelMessageStore.getState().incrementUnread(CH);
        useChannelMessageStore.getState().incrementUnread(CH);
      });
      expect(useChannelMessageStore.getState().unreadCounts[CH]).toBe(2);
      act(() => { useChannelMessageStore.getState().clearChannelUnread(CH); });
      expect(useChannelMessageStore.getState().unreadCounts[CH]).toBeUndefined();
    });

    it("clearChannelUnread should no-op for unknown channel", () => {
      act(() => { useChannelMessageStore.getState().clearChannelUnread(999); });
      expect(useChannelMessageStore.getState().unreadCounts).toEqual({});
    });

    it("totalUnreadCount should sum all channels", () => {
      act(() => {
        useChannelMessageStore.getState().incrementUnread(1);
        useChannelMessageStore.getState().incrementUnread(1);
        useChannelMessageStore.getState().incrementUnread(2);
        useChannelMessageStore.getState().incrementUnread(3);
        useChannelMessageStore.getState().incrementUnread(3);
        useChannelMessageStore.getState().incrementUnread(3);
      });
      expect(useChannelMessageStore.getState().totalUnreadCount()).toBe(6);
    });

    it("totalUnreadCount should return 0 when no unread", () => {
      expect(useChannelMessageStore.getState().totalUnreadCount()).toBe(0);
    });

    it("totalUnreadCount should update when channel is cleared", () => {
      act(() => {
        useChannelMessageStore.getState().incrementUnread(1);
        useChannelMessageStore.getState().incrementUnread(2);
        useChannelMessageStore.getState().incrementUnread(2);
      });
      expect(useChannelMessageStore.getState().totalUnreadCount()).toBe(3);
      act(() => { useChannelMessageStore.getState().clearChannelUnread(2); });
      expect(useChannelMessageStore.getState().totalUnreadCount()).toBe(1);
    });
  });

  describe("LRU eviction", () => {
    it("should evict oldest channels when cache exceeds 20", async () => {
      // Populate 21 channels
      for (let i = 1; i <= 21; i++) {
        vi.mocked(channelApi.getMessages).mockResolvedValue({ messages: [{ ...mockMessage, id: i, channel_id: i }], has_more: false });
        await act(async () => { await useChannelMessageStore.getState().fetchMessages(i); });
      }
      const keys = Object.keys(useChannelMessageStore.getState().cache).map(Number);
      expect(keys.length).toBeLessThanOrEqual(20);
      // Channel 21 (most recent) should be present
      expect(useChannelMessageStore.getState().cache[21]).toBeDefined();
    });
  });
});
