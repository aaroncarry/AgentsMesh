import { request } from "./base";

// Runner interface matching the store
export interface RunnerData {
  id: number;
  node_id: string;
  description?: string;
  status: "online" | "offline" | "maintenance" | "busy";
  last_heartbeat?: string;
  current_sessions: number;
  max_concurrent_sessions: number;
  runner_version?: string;
  is_enabled: boolean;
  host_info?: {
    os?: string;
    arch?: string;
    memory?: number;
    cpu_cores?: number;
    hostname?: string;
  };
  created_at: string;
  updated_at: string;
  active_sessions?: Array<{
    session_key: string;
    status: string;
    agent_status: string;
  }>;
}

export interface RegistrationToken {
  id: number;
  organization_id: number;
  description?: string;
  created_by_id: number;
  is_active: boolean;
  max_uses?: number;
  used_count: number;
  expires_at?: string;
  created_at: string;
}

export const runnerApi = {
  list: (status?: string) => {
    const params = status ? `?status=${status}` : "";
    return request<{ runners: RunnerData[] }>(`/api/v1/org/runners${params}`);
  },

  listAvailable: () =>
    request<{ runners: RunnerData[] }>("/api/v1/org/runners/available"),

  get: (id: number) =>
    request<{ runner: RunnerData }>(`/api/v1/org/runners/${id}`),

  update: (id: number, data: { description?: string; max_concurrent_sessions?: number; is_enabled?: boolean }) =>
    request<{ runner: RunnerData }>(`/api/v1/org/runners/${id}`, {
      method: "PUT",
      body: data,
    }),

  delete: (id: number) =>
    request<{ message: string }>(`/api/v1/org/runners/${id}`, {
      method: "DELETE",
    }),

  regenerateAuthToken: (id: number) =>
    request<{ auth_token: string; message: string }>(`/api/v1/org/runners/${id}/regenerate-token`, {
      method: "POST",
    }),

  // Registration token management
  listTokens: () =>
    request<{ tokens: RegistrationToken[] }>("/api/v1/org/runners/tokens"),

  createToken: (description?: string, maxUses?: number, expiresAt?: string) =>
    request<{ token: string; message: string }>("/api/v1/org/runners/tokens", {
      method: "POST",
      body: { description, max_uses: maxUses, expires_at: expiresAt },
    }),

  revokeToken: (id: number) =>
    request<{ message: string }>(`/api/v1/org/runners/tokens/${id}`, {
      method: "DELETE",
    }),
};
