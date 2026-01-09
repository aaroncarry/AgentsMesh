import { request } from "./base";

// Ticket types
export type TicketType = "task" | "bug" | "feature" | "epic" | "subtask" | "story";
export type TicketStatus = "backlog" | "todo" | "in_progress" | "in_review" | "done" | "cancelled";
export type TicketPriority = "none" | "low" | "medium" | "high" | "urgent";

export interface TicketData {
  id: number;
  number: number;
  identifier: string;
  type: TicketType;
  title: string;
  description?: string;
  content?: string;
  status: TicketStatus;
  priority: TicketPriority;
  severity?: string;
  estimate?: number;
  due_date?: string;
  started_at?: string;
  completed_at?: string;
  created_at: string;
  updated_at: string;
  reporter?: { id: number; username: string; name?: string; avatar_url?: string };
  assignees?: Array<{ id: number; username: string; name?: string; avatar_url?: string }>;
  labels?: Array<{ id: number; name: string; color: string }>;
  repository?: { id: number; name: string };
  parent_ticket?: { id: number; identifier: string; title: string };
}

export interface TicketRelation {
  id: number;
  source_ticket_id: number;
  target_ticket_id: number;
  relation_type: string;
  source_ticket?: { id: number; identifier: string; title: string };
  target_ticket?: { id: number; identifier: string; title: string };
  created_at: string;
}

export interface TicketCommit {
  id: number;
  ticket_id: number;
  commit_sha: string;
  commit_message?: string;
  commit_url?: string;
  author_name?: string;
  author_email?: string;
  committed_at?: string;
  created_at: string;
}

export interface BoardColumn {
  status: string;
  tickets: TicketData[];
  count: number;
}

// Tickets API
export const ticketApi = {
  list: (filters?: {
    status?: string;
    priority?: string;
    type?: string;
    assigneeId?: number;
    repositoryId?: number;
    search?: string;
    limit?: number;
    offset?: number;
  }) => {
    const params = new URLSearchParams();
    if (filters) {
      Object.entries(filters).forEach(([key, value]) => {
        if (value !== undefined && value !== null) {
          params.append(key, String(value));
        }
      });
    }
    const query = params.toString() ? `?${params.toString()}` : "";
    return request<{ tickets: TicketData[]; total: number }>(`/api/v1/org/tickets${query}`);
  },

  get: (identifier: string) =>
    request<TicketData>(`/api/v1/org/tickets/${identifier}`),

  create: (data: {
    repositoryId: number;
    type: string;
    title: string;
    description?: string;
    content?: string;
    priority?: string;
    severity?: string;
    estimate?: number;
    assigneeIds?: number[];
    labels?: string[];
    parentId?: number;
  }) =>
    request<TicketData>("/api/v1/org/tickets", {
      method: "POST",
      body: data,
    }),

  update: (identifier: string, data: {
    title?: string;
    description?: string;
    content?: string;
    type?: string;
    status?: string;
    priority?: string;
    severity?: string;
    estimate?: number;
    assigneeIds?: number[];
    labels?: string[];
  }) =>
    request<TicketData>(`/api/v1/org/tickets/${identifier}`, {
      method: "PUT",
      body: data,
    }),

  delete: (identifier: string) =>
    request<{ message: string }>(`/api/v1/org/tickets/${identifier}`, {
      method: "DELETE",
    }),

  updateStatus: (identifier: string, status: string) =>
    request<TicketData>(`/api/v1/org/tickets/${identifier}/status`, {
      method: "PATCH",
      body: { status },
    }),

  // Active tickets (in_progress or in_review)
  getActive: (limit?: number) => {
    const params = limit ? `?limit=${limit}` : "";
    return request<{ tickets: TicketData[] }>(`/api/v1/org/tickets/active${params}`);
  },

  // Board view
  getBoard: (repositoryId?: number) => {
    const params = repositoryId ? `?repository_id=${repositoryId}` : "";
    return request<{ columns: BoardColumn[] }>(`/api/v1/org/tickets/board${params}`);
  },

  // Sub-tickets
  getSubTickets: (identifier: string) =>
    request<{ tickets: TicketData[] }>(`/api/v1/org/tickets/${identifier}/sub-tickets`),

  // Relations
  listRelations: (identifier: string) =>
    request<{ relations: TicketRelation[] }>(`/api/v1/org/tickets/${identifier}/relations`),

  createRelation: (identifier: string, data: { target_ticket_id: number; relation_type: string }) =>
    request<{ relation: TicketRelation }>(`/api/v1/org/tickets/${identifier}/relations`, {
      method: "POST",
      body: data,
    }),

  deleteRelation: (identifier: string, relationId: number) =>
    request<{ message: string }>(`/api/v1/org/tickets/${identifier}/relations/${relationId}`, {
      method: "DELETE",
    }),

  // Commits
  listCommits: (identifier: string) =>
    request<{ commits: TicketCommit[] }>(`/api/v1/org/tickets/${identifier}/commits`),

  linkCommit: (identifier: string, data: {
    commit_sha: string;
    commit_message?: string;
    commit_url?: string;
    author_name?: string;
    author_email?: string;
    committed_at?: string;
  }) =>
    request<{ commit: TicketCommit }>(`/api/v1/org/tickets/${identifier}/commits`, {
      method: "POST",
      body: data,
    }),

  unlinkCommit: (identifier: string, commitId: number) =>
    request<{ message: string }>(`/api/v1/org/tickets/${identifier}/commits/${commitId}`, {
      method: "DELETE",
    }),

  // Merge Requests
  listMergeRequests: (identifier: string) =>
    request<{
      merge_requests: Array<{
        id: number;
        mr_iid: number;
        title: string;
        state: string;
        web_url: string;
        source_branch: string;
        target_branch: string;
      }>;
    }>(`/api/v1/org/tickets/${identifier}/merge-requests`),

  // Labels
  listLabels: (repositoryId?: number) => {
    const params = repositoryId ? `?repository_id=${repositoryId}` : "";
    return request<{ labels: Array<{ id: number; name: string; color: string }> }>(
      `/api/v1/org/labels${params}`
    );
  },

  createLabel: (name: string, color: string, repositoryId?: number) =>
    request<{ id: number; name: string; color: string }>("/api/v1/org/labels", {
      method: "POST",
      body: { name, color, repository_id: repositoryId },
    }),

  updateLabel: (id: number, data: { name?: string; color?: string }) =>
    request<{ id: number; name: string; color: string }>(`/api/v1/org/labels/${id}`, {
      method: "PUT",
      body: data,
    }),

  deleteLabel: (id: number) =>
    request<{ message: string }>(`/api/v1/org/labels/${id}`, {
      method: "DELETE",
    }),

  // Assignees
  addAssignee: (identifier: string, userId: number) =>
    request<{ message: string }>(`/api/v1/org/tickets/${identifier}/assignees`, {
      method: "POST",
      body: { user_id: userId },
    }),

  removeAssignee: (identifier: string, userId: number) =>
    request<{ message: string }>(`/api/v1/org/tickets/${identifier}/assignees/${userId}`, {
      method: "DELETE",
    }),

  // Ticket labels
  addLabel: (identifier: string, labelId: number) =>
    request<{ message: string }>(`/api/v1/org/tickets/${identifier}/labels`, {
      method: "POST",
      body: { label_id: labelId },
    }),

  removeLabel: (identifier: string, labelId: number) =>
    request<{ message: string }>(`/api/v1/org/tickets/${identifier}/labels/${labelId}`, {
      method: "DELETE",
    }),

  // Sessions (DevMesh integration)
  getSessions: (identifier: string, activeOnly?: boolean) => {
    const params = activeOnly ? "?active=true" : "";
    return request<{
      sessions: Array<{
        session_key: string;
        status: string;
        agent_status: string;
        model?: string;
        started_at?: string;
        runner_id: number;
        created_by_id: number;
      }>;
    }>(`/api/v1/org/tickets/${identifier}/sessions${params}`);
  },

  createSession: (identifier: string, data: {
    runner_id: number;
    initial_prompt?: string;
    model?: string;
    permission_mode?: string;
    think_level?: string;
  }) =>
    request<{
      message: string;
      session: {
        session_key: string;
        status: string;
      };
    }>(`/api/v1/org/tickets/${identifier}/sessions`, {
      method: "POST",
      body: data,
    }),

  // Batch sessions
  batchGetSessions: (ticketIds: number[]) =>
    request<{
      ticket_sessions: Record<number, Array<{
        session_key: string;
        status: string;
        agent_status: string;
      }>>;
    }>("/api/v1/org/tickets/batch-sessions", {
      method: "POST",
      body: { ticket_ids: ticketIds },
    }),
};
