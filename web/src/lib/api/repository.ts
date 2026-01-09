import { request } from "./base";

// Repository types (self-contained, no git_provider_id)
export interface RepositoryData {
  id: number;
  organization_id: number;
  provider_type: string; // github, gitlab, gitee, generic
  provider_base_url: string; // https://github.com
  clone_url: string;
  external_id: string;
  name: string;
  full_path: string;
  default_branch: string;
  ticket_prefix?: string;
  visibility: string; // "organization" or "private"
  imported_by_user_id?: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateRepositoryRequest {
  provider_type: string;
  provider_base_url: string;
  clone_url?: string;
  external_id: string;
  name: string;
  full_path: string;
  default_branch?: string;
  ticket_prefix?: string;
  visibility?: string;
}

export interface UpdateRepositoryRequest {
  name?: string;
  default_branch?: string;
  ticket_prefix?: string;
  is_active?: boolean;
}

// Repository API
export const repositoryApi = {
  list: () => {
    return request<{ repositories: RepositoryData[] }>(`/api/v1/org/repositories`);
  },

  get: (id: number) =>
    request<{ repository: RepositoryData }>(`/api/v1/org/repositories/${id}`),

  create: (data: CreateRepositoryRequest) =>
    request<{ repository: RepositoryData }>("/api/v1/org/repositories", {
      method: "POST",
      body: data,
    }),

  update: (id: number, data: UpdateRepositoryRequest) =>
    request<{ repository: RepositoryData }>(`/api/v1/org/repositories/${id}`, {
      method: "PUT",
      body: data,
    }),

  delete: (id: number) =>
    request<{ message: string }>(`/api/v1/org/repositories/${id}`, {
      method: "DELETE",
    }),

  listBranches: (id: number, accessToken: string) =>
    request<{ branches: string[] }>(`/api/v1/org/repositories/${id}/branches`, {
      headers: { "X-Git-Access-Token": accessToken },
    }),

  syncBranches: (id: number, accessToken: string) =>
    request<{ branches: string[] }>(`/api/v1/org/repositories/${id}/sync-branches`, {
      method: "POST",
      body: { access_token: accessToken },
    }),

  setupWebhook: (id: number) =>
    request<{ message: string; webhook_url?: string }>(`/api/v1/org/repositories/${id}/webhook`, {
      method: "POST",
    }),
};
