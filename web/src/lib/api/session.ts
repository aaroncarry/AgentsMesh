import { request } from "./base";

// Session interface matching the store
export interface SessionData {
  id: number;
  session_key: string;
  status: "initializing" | "running" | "paused" | "terminated" | "failed";
  agent_status: string;
  initial_prompt?: string;
  branch_name?: string;
  worktree_path?: string;
  started_at?: string;
  finished_at?: string;
  last_activity?: string;
  created_at: string;
  runner?: {
    id: number;
    node_id: string;
    status: string;
  };
  agent_type?: {
    id: number;
    name: string;
    slug: string;
  };
  repository?: {
    id: number;
    name: string;
    full_path: string;
  };
  ticket?: {
    id: number;
    identifier: string;
    title: string;
  };
  created_by?: {
    id: number;
    username: string;
    name?: string;
  };
}

// Sessions API
export const sessionApi = {
  list: (filters?: { status?: string; runnerId?: number }) => {
    const params = new URLSearchParams();
    if (filters?.status) params.append("status", filters.status);
    if (filters?.runnerId) params.append("runner_id", String(filters.runnerId));
    const query = params.toString() ? `?${params.toString()}` : "";
    return request<{ sessions: SessionData[]; total: number }>(`/api/v1/org/sessions${query}`);
  },

  get: (key: string) =>
    request<{ session: SessionData }>(`/api/v1/org/sessions/${key}`),

  create: (data: {
    agent_type_id: number;
    runner_id?: number;
    repository_id?: number;
    ticket_id?: number;
    initial_prompt?: string;
    branch_name?: string;
  }) =>
    request<{ message: string; session: SessionData }>(
      "/api/v1/org/sessions",
      {
        method: "POST",
        body: data,
      }
    ),

  terminate: (key: string) =>
    request<{ message: string }>(`/api/v1/org/sessions/${key}/terminate`, {
      method: "POST",
    }),

  // Get connection info for WebSocket terminal
  getConnectionInfo: (key: string) =>
    request<{ session_key: string; ws_url: string; status: string }>(
      `/api/v1/org/sessions/${key}/connect`
    ),

  // Terminal control - observe terminal output
  observeTerminal: (key: string, lines?: number) => {
    const params = lines ? `?lines=${lines}` : "";
    return request<{
      session_key: string;
      output: string;
      status: string;
      agent_status: string;
    }>(`/api/v1/org/sessions/${key}/terminal/observe${params}`);
  },

  // Terminal control - send input
  sendTerminalInput: (key: string, input: string) =>
    request<{ message: string }>(`/api/v1/org/sessions/${key}/terminal/input`, {
      method: "POST",
      body: { input },
    }),

  // Terminal control - resize terminal
  resizeTerminal: (key: string, cols: number, rows: number) =>
    request<{ message: string }>(`/api/v1/org/sessions/${key}/terminal/resize`, {
      method: "POST",
      body: { cols, rows },
    }),

  // Send prompt to session
  sendPrompt: (key: string, prompt: string) =>
    request<{ message: string }>(`/api/v1/org/sessions/${key}/send-prompt`, {
      method: "POST",
      body: { prompt },
    }),
};
