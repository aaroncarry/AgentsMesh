import { request } from "./base";

// Channel types
export interface ChannelData {
  id: number;
  organization_id: number;
  name: string;
  description?: string;
  document?: string;
  repository_id?: number;
  ticket_id?: number;
  created_by_session?: string;
  created_by_user_id?: number;
  is_archived: boolean;
  created_at: string;
  updated_at: string;
}

export interface ChannelMessage {
  id: number;
  channel_id: number;
  sender_session?: string;
  sender_user_id?: number;
  message_type: "text" | "system" | "code" | "command";
  content: string;
  metadata?: Record<string, unknown>;
  created_at: string;
  session?: {
    session_key: string;
    agent_type?: {
      name: string;
    };
  };
  user?: {
    id: number;
    username: string;
    name?: string;
    avatar_url?: string;
  };
}

// Channels API
export const channelApi = {
  // List channels with optional filters
  list: (filters?: {
    repository_id?: number;
    ticket_id?: number;
    include_archived?: boolean;
  }) => {
    const params = new URLSearchParams();
    if (filters?.repository_id) params.append("repository_id", String(filters.repository_id));
    if (filters?.ticket_id) params.append("ticket_id", String(filters.ticket_id));
    if (filters?.include_archived) params.append("include_archived", "true");
    const query = params.toString() ? `?${params.toString()}` : "";
    return request<{ channels: ChannelData[]; total: number }>(`/api/v1/org/channels${query}`);
  },

  // Get a single channel
  get: (id: number) =>
    request<{ channel: ChannelData }>(`/api/v1/org/channels/${id}`),

  // Create a new channel
  create: (data: {
    name: string;
    description?: string;
    document?: string;
    repository_id?: number;
    ticket_id?: number;
  }) =>
    request<{ channel: ChannelData }>("/api/v1/org/channels", {
      method: "POST",
      body: data,
    }),

  // Update a channel
  update: (id: number, data: { name?: string; description?: string; document?: string }) =>
    request<{ channel: ChannelData }>(`/api/v1/org/channels/${id}`, {
      method: "PUT",
      body: data,
    }),

  // Archive a channel
  archive: (id: number) =>
    request<{ message: string }>(`/api/v1/org/channels/${id}/archive`, {
      method: "POST",
    }),

  // Unarchive a channel
  unarchive: (id: number) =>
    request<{ message: string }>(`/api/v1/org/channels/${id}/unarchive`, {
      method: "POST",
    }),

  // Get messages in a channel
  getMessages: (id: number, limit?: number, offset?: number) => {
    const params = new URLSearchParams();
    if (limit) params.append("limit", String(limit));
    if (offset) params.append("offset", String(offset));
    const query = params.toString() ? `?${params.toString()}` : "";
    return request<{ messages: ChannelMessage[] }>(`/api/v1/org/channels/${id}/messages${query}`);
  },

  // Send a message to a channel
  sendMessage: (id: number, content: string, sessionKey?: string, messageType?: string) =>
    request<{ message: ChannelMessage }>(`/api/v1/org/channels/${id}/messages`, {
      method: "POST",
      body: { content, session_key: sessionKey, message_type: messageType || "text" },
    }),

  // Get sessions joined to a channel
  getSessions: (id: number) =>
    request<{
      sessions: Array<{
        id: number;
        session_key: string;
        status: string;
        agent_status: string;
      }>;
      total: number;
    }>(`/api/v1/org/channels/${id}/sessions`),

  // Join a session to a channel
  joinSession: (id: number, sessionKey: string) =>
    request<{ message: string }>(`/api/v1/org/channels/${id}/sessions`, {
      method: "POST",
      body: { session_key: sessionKey },
    }),

  // Remove a session from a channel
  leaveSession: (id: number, sessionKey: string) =>
    request<{ message: string }>(`/api/v1/org/channels/${id}/sessions/${sessionKey}`, {
      method: "DELETE",
    }),
};
