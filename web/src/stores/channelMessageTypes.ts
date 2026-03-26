import type { ChannelMessage } from "@/lib/api";
import type { MentionPayload } from "@/lib/api";

export interface ChannelMessageCache {
  messages: ChannelMessage[];
  hasMore: boolean;
  loading: boolean;
  loadingMore: boolean;
  error: string | null;
}

export const EMPTY_CACHE: ChannelMessageCache = {
  messages: [],
  hasMore: false,
  loading: false,
  loadingMore: false,
  error: null,
};

export interface ChannelMessageState {
  cache: Record<number, ChannelMessageCache>;
  unreadCounts: Record<number, number>;

  // Message CRUD — all routed by channelId
  fetchMessages: (channelId: number, limit?: number, beforeId?: number) => Promise<void>;
  sendMessage: (
    channelId: number,
    content: string,
    podKey?: string,
    mentions?: MentionPayload[]
  ) => Promise<ChannelMessage>;
  addMessage: (channelId: number, message: ChannelMessage) => void;
  editMessage: (channelId: number, messageId: number, content: string) => Promise<void>;
  deleteMessage: (channelId: number, messageId: number) => Promise<void>;
  updateMessage: (channelId: number, data: { id: number; content: string; edited_at: string }) => void;
  removeMessage: (channelId: number, messageId: number) => void;

  // Unread / read state
  fetchUnreadCounts: () => Promise<void>;
  markRead: (channelId: number, messageId: number) => Promise<void>;
  muteChannel: (channelId: number, muted: boolean) => Promise<void>;
  incrementUnread: (channelId: number) => void;
  clearChannelUnread: (channelId: number) => void;
}

/** Get or create a channel cache entry */
export function getCache(state: ChannelMessageState, channelId: number): ChannelMessageCache {
  return state.cache[channelId] ?? EMPTY_CACHE;
}

/** Max number of channels to keep in cache (LRU eviction) */
const MAX_CACHED_CHANNELS = 20;

/** Immutably update a single channel's cache, evicting oldest entries beyond limit */
export function updateCache(
  state: ChannelMessageState,
  channelId: number,
  patch: Partial<ChannelMessageCache>
): { cache: Record<number, ChannelMessageCache> } {
  const newCache = {
    ...state.cache,
    [channelId]: { ...getCache(state, channelId), ...patch },
  };

  // Evict oldest entries if over limit (keep active channel + most recent)
  const keys = Object.keys(newCache).map(Number);
  if (keys.length > MAX_CACHED_CHANNELS) {
    const toEvict = keys
      .filter((k) => k !== channelId)
      .slice(0, keys.length - MAX_CACHED_CHANNELS);
    for (const k of toEvict) {
      delete newCache[k];
    }
  }

  return { cache: newCache };
}
