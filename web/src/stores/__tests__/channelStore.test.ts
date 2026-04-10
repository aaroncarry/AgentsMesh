import { describe, it, expect, beforeEach, vi } from "vitest";
import { act } from "@testing-library/react";
import { useChannelStore, useChannelMessageStore, Channel } from "../channel";
import type { ChannelMessage } from "@/lib/api";

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

const mockChannel: Channel = {
  id: 1,
  name: "general",
  description: "General discussion channel",
  is_archived: false,
  organization_id: 1,
  created_at: "2024-01-01T00:00:00Z",
  updated_at: "2024-01-01T00:00:00Z",
};

const mockChannel2: Channel = {
  id: 2,
  name: "dev-chat",
  description: "Development discussion",
  is_archived: false,
  organization_id: 1,
  created_at: "2024-01-02T00:00:00Z",
  updated_at: "2024-01-02T00:00:00Z",
};

const mockMessage: ChannelMessage = {
  id: 1,
  channel_id: 1,
  content: "Hello, world!",
  message_type: "text",
  created_at: "2024-01-01T00:00:00Z",
};

describe("Channel Store", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useChannelStore.setState({
      channels: [],
      currentChannel: null,
      loading: false,
      channelLoading: false,
      error: null,
      selectedChannelId: null,
      searchQuery: "",
      showArchived: false,
    });
    useChannelMessageStore.setState({ cache: {}, unreadCounts: {} });
  });

  describe("initial state", () => {
    it("should have default values", () => {
      const state = useChannelStore.getState();
      expect(state.channels).toEqual([]);
      expect(state.currentChannel).toBeNull();
      expect(state.loading).toBe(false);
      expect(state.error).toBeNull();
    });
  });

  describe("fetchChannels", () => {
    it("should fetch channels successfully", async () => {
      vi.mocked(channelApi.list).mockResolvedValue({ channels: [mockChannel, mockChannel2], total: 2 });
      await act(async () => { await useChannelStore.getState().fetchChannels(); });
      const state = useChannelStore.getState();
      expect(state.channels).toHaveLength(2);
      expect(state.channels[0].name).toBe("general");
    });

    it("should pass filters to API", async () => {
      vi.mocked(channelApi.list).mockResolvedValue({ channels: [], total: 0 });
      await act(async () => { await useChannelStore.getState().fetchChannels({ includeArchived: true }); });
      expect(channelApi.list).toHaveBeenCalledWith({ include_archived: true });
    });

    it("should handle empty response", async () => {
      vi.mocked(channelApi.list).mockResolvedValue({ channels: undefined as unknown as Channel[], total: 0 });
      await act(async () => { await useChannelStore.getState().fetchChannels(); });
      expect(useChannelStore.getState().channels).toEqual([]);
    });

    it("should handle fetch error", async () => {
      vi.mocked(channelApi.list).mockRejectedValue(new Error("Network error"));
      await act(async () => { await useChannelStore.getState().fetchChannels(); });
      expect(useChannelStore.getState().error).toBe("Network error");
    });
  });

  describe("fetchChannel", () => {
    it("should fetch single channel", async () => {
      vi.mocked(channelApi.get).mockResolvedValue({ channel: mockChannel });
      await act(async () => { await useChannelStore.getState().fetchChannel(1); });
      expect(useChannelStore.getState().currentChannel).toEqual(mockChannel);
    });

    it("should handle error", async () => {
      vi.mocked(channelApi.get).mockRejectedValue({ message: "Channel not found" });
      await act(async () => { await useChannelStore.getState().fetchChannel(999); });
      expect(useChannelStore.getState().error).toBe("Channel not found");
    });
  });

  describe("createChannel", () => {
    it("should create and add to list", async () => {
      vi.mocked(channelApi.create).mockResolvedValue({ channel: mockChannel });
      let result: Channel;
      await act(async () => { result = await useChannelStore.getState().createChannel({ name: "general", description: "General discussion channel" }); });
      expect(result!).toEqual(mockChannel);
      expect(useChannelStore.getState().channels).toContainEqual(mockChannel);
    });

    it("should convert camelCase to snake_case", async () => {
      vi.mocked(channelApi.create).mockResolvedValue({ channel: mockChannel });
      await act(async () => { await useChannelStore.getState().createChannel({ name: "test", repositoryId: 1, ticketSlug: "PROJ-2" }); });
      expect(channelApi.create).toHaveBeenCalledWith({ name: "test", description: undefined, document: undefined, repository_id: 1, ticket_slug: "PROJ-2" });
    });

    it("should handle error", async () => {
      vi.mocked(channelApi.create).mockRejectedValue(new Error("Create failed"));
      await expect(act(async () => { await useChannelStore.getState().createChannel({ name: "test" }); })).rejects.toThrow("Create failed");
    });
  });

  describe("updateChannel", () => {
    beforeEach(() => { useChannelStore.setState({ channels: [mockChannel], currentChannel: mockChannel }); });

    it("should update channel and currentChannel", async () => {
      vi.mocked(channelApi.update).mockResolvedValue({ channel: { ...mockChannel, name: "updated" } });
      await act(async () => { await useChannelStore.getState().updateChannel(1, { name: "updated" }); });
      expect(useChannelStore.getState().channels[0].name).toBe("updated");
      expect(useChannelStore.getState().currentChannel?.name).toBe("updated");
    });

    it("should not update currentChannel if different id", async () => {
      vi.mocked(channelApi.update).mockResolvedValue({ channel: { ...mockChannel2, name: "updated-dev" } });
      useChannelStore.setState({ channels: [mockChannel, mockChannel2] });
      await act(async () => { await useChannelStore.getState().updateChannel(2, { name: "updated-dev" }); });
      expect(useChannelStore.getState().currentChannel?.name).toBe("general");
    });
  });

  describe("archive/unarchive", () => {
    it("should archive", async () => {
      useChannelStore.setState({ channels: [mockChannel], currentChannel: mockChannel });
      vi.mocked(channelApi.archive).mockResolvedValue({ message: "ok" });
      await act(async () => { await useChannelStore.getState().archiveChannel(1); });
      expect(useChannelStore.getState().channels[0].is_archived).toBe(true);
    });

    it("should unarchive", async () => {
      const archived = { ...mockChannel, is_archived: true };
      useChannelStore.setState({ channels: [archived], currentChannel: archived });
      vi.mocked(channelApi.unarchive).mockResolvedValue({ message: "ok" });
      await act(async () => { await useChannelStore.getState().unarchiveChannel(1); });
      expect(useChannelStore.getState().channels[0].is_archived).toBe(false);
    });
  });

  describe("join/leave channel", () => {
    it("should join and refresh", async () => {
      useChannelStore.setState({ channels: [mockChannel], currentChannel: mockChannel });
      const updated = { ...mockChannel, pods: [{ pod_key: "pod-123", status: "running" }] };
      vi.mocked(channelApi.joinPod).mockResolvedValue({ message: "ok" });
      vi.mocked(channelApi.get).mockResolvedValue({ channel: updated });
      await act(async () => { await useChannelStore.getState().joinChannel(1, "pod-123"); });
      expect(useChannelStore.getState().channels[0].pods).toHaveLength(1);
    });

    it("should leave and refresh", async () => {
      useChannelStore.setState({ channels: [{ ...mockChannel, pods: [{ pod_key: "pod-123", status: "running" }] }], currentChannel: mockChannel });
      vi.mocked(channelApi.leavePod).mockResolvedValue({ message: "ok" });
      vi.mocked(channelApi.get).mockResolvedValue({ channel: { ...mockChannel, pods: [] } as never });
      await act(async () => { await useChannelStore.getState().leaveChannel(1, "pod-123"); });
      expect(useChannelStore.getState().channels[0].pods).toHaveLength(0);
    });
  });

  describe("setCurrentChannel", () => {
    it("should set and clear", () => {
      act(() => { useChannelStore.getState().setCurrentChannel(mockChannel); });
      expect(useChannelStore.getState().currentChannel).toEqual(mockChannel);
      act(() => { useChannelStore.getState().setCurrentChannel(null); });
      expect(useChannelStore.getState().currentChannel).toBeNull();
    });

    it("should not affect per-channel message cache", () => {
      useChannelMessageStore.setState({
        cache: { 1: { messages: [mockMessage], hasMore: false, loading: false, loadingMore: false, error: null } },
      });
      act(() => { useChannelStore.getState().setCurrentChannel(mockChannel); });
      expect(useChannelMessageStore.getState().cache[1]?.messages).toHaveLength(1);
    });
  });

  describe("clearError", () => {
    it("should clear", () => {
      useChannelStore.setState({ error: "err" });
      act(() => { useChannelStore.getState().clearError(); });
      expect(useChannelStore.getState().error).toBeNull();
    });
  });
});
