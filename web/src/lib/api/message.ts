import { request } from "./base";

// Agent Message types
export interface AgentMessage {
  id: number;
  sender_session: string;
  receiver_session: string;
  message_type: string;
  content: Record<string, unknown>;
  status: "pending" | "delivered" | "read" | "failed" | "dead_letter";
  correlation_id?: string;
  parent_message_id?: number;
  delivery_attempts: number;
  max_retries: number;
  delivered_at?: string;
  read_at?: string;
  created_at: string;
  updated_at: string;
}

export interface DeadLetterEntry {
  id: number;
  original_message_id: number;
  original_message?: AgentMessage;
  reason: string;
  final_attempt: number;
  moved_at: string;
  replayed_at?: string;
  replay_result?: string;
}

// Message API
export const messageApi = {
  // Send a message to another session
  sendMessage: (data: {
    receiver_session: string;
    message_type: string;
    content: Record<string, unknown>;
    correlation_id?: string;
    reply_to_id?: number;
  }, sessionKey?: string) =>
    request<{ message: AgentMessage }>("/api/v1/org/messages", {
      method: "POST",
      body: data,
      headers: sessionKey ? { "X-Session-Key": sessionKey } : undefined,
    }),

  // Get messages for the current session
  getMessages: (params?: {
    unread_only?: boolean;
    message_types?: string[];
    limit?: number;
    offset?: number;
  }, sessionKey?: string) => {
    const searchParams = new URLSearchParams();
    if (params?.unread_only) searchParams.append("unread_only", "true");
    if (params?.message_types) {
      params.message_types.forEach(t => searchParams.append("message_types", t));
    }
    if (params?.limit) searchParams.append("limit", String(params.limit));
    if (params?.offset) searchParams.append("offset", String(params.offset));
    const query = searchParams.toString() ? `?${searchParams.toString()}` : "";
    return request<{ messages: AgentMessage[]; total: number; unread_count: number }>(
      `/api/v1/org/messages${query}`,
      { headers: sessionKey ? { "X-Session-Key": sessionKey } : undefined }
    );
  },

  // Get count of unread messages
  getUnreadCount: (sessionKey?: string) =>
    request<{ count: number }>("/api/v1/org/messages/unread-count", {
      headers: sessionKey ? { "X-Session-Key": sessionKey } : undefined,
    }),

  // Get a specific message by ID
  getMessage: (id: number, sessionKey?: string) =>
    request<{ message: AgentMessage }>(`/api/v1/org/messages/${id}`, {
      headers: sessionKey ? { "X-Session-Key": sessionKey } : undefined,
    }),

  // Mark messages as read
  markRead: (messageIds: number[], sessionKey?: string) =>
    request<{ marked_count: number }>("/api/v1/org/messages/mark-read", {
      method: "POST",
      body: { message_ids: messageIds },
      headers: sessionKey ? { "X-Session-Key": sessionKey } : undefined,
    }),

  // Mark all messages as read
  markAllRead: (sessionKey?: string) =>
    request<{ marked_count: number }>("/api/v1/org/messages/mark-all-read", {
      method: "POST",
      headers: sessionKey ? { "X-Session-Key": sessionKey } : undefined,
    }),

  // Get conversation by correlation ID
  getConversation: (correlationId: string, limit?: number, sessionKey?: string) => {
    const params = limit ? `?limit=${limit}` : "";
    return request<{ messages: AgentMessage[]; total: number }>(
      `/api/v1/org/messages/conversation/${correlationId}${params}`,
      { headers: sessionKey ? { "X-Session-Key": sessionKey } : undefined }
    );
  },

  // Get sent messages
  getSentMessages: (params?: { limit?: number; offset?: number }, sessionKey?: string) => {
    const searchParams = new URLSearchParams();
    if (params?.limit) searchParams.append("limit", String(params.limit));
    if (params?.offset) searchParams.append("offset", String(params.offset));
    const query = searchParams.toString() ? `?${searchParams.toString()}` : "";
    return request<{ messages: AgentMessage[]; total: number }>(
      `/api/v1/org/messages/sent${query}`,
      { headers: sessionKey ? { "X-Session-Key": sessionKey } : undefined }
    );
  },

  // Get dead letter queue entries
  getDeadLetters: (params?: { limit?: number; offset?: number }) => {
    const searchParams = new URLSearchParams();
    if (params?.limit) searchParams.append("limit", String(params.limit));
    if (params?.offset) searchParams.append("offset", String(params.offset));
    const query = searchParams.toString() ? `?${searchParams.toString()}` : "";
    return request<{ entries: DeadLetterEntry[]; total: number }>(`/api/v1/org/messages/dlq${query}`);
  },

  // Replay a dead letter message
  replayDeadLetter: (entryId: number) =>
    request<{ message: string; replayed_message: AgentMessage }>(
      `/api/v1/org/messages/dlq/${entryId}/replay`,
      { method: "POST" }
    ),
};
