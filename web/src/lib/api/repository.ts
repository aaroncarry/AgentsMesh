import { request } from "./base";
import { GitProviderData } from "./git-provider";

// Repository types
export interface RepositoryData {
  id: number;
  organization_id: number;
  git_provider_id: number;
  external_id: string;
  name: string;
  full_path: string;
  default_branch: string;
  ticket_prefix?: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  git_provider?: GitProviderData;
}

// Repository API
export const repositoryApi = {
  list: () => {
    return request<{ repositories: RepositoryData[] }>(`/api/v1/org/repositories`);
  },

  get: (id: number) =>
    request<{ repository: RepositoryData }>(`/api/v1/org/repositories/${id}`),

  create: (data: {
    git_provider_id: number;
    external_id: string;
    name: string;
    full_path: string;
    default_branch?: string;
    ticket_prefix?: string;
  }) =>
    request<{ repository: RepositoryData }>("/api/v1/org/repositories", {
      method: "POST",
      body: data,
    }),

  update: (id: number, data: {
    name?: string;
    default_branch?: string;
    ticket_prefix?: string;
    is_active?: boolean;
  }) =>
    request<{ repository: RepositoryData }>(`/api/v1/org/repositories/${id}`, {
      method: "PUT",
      body: data,
    }),

  delete: (id: number) =>
    request<{ message: string }>(`/api/v1/org/repositories/${id}`, {
      method: "DELETE",
    }),

  listBranches: (id: number) =>
    request<{ branches: string[] }>(`/api/v1/org/repositories/${id}/branches`),

  syncBranches: (id: number) =>
    request<{ branches: string[]; message: string }>(`/api/v1/org/repositories/${id}/sync-branches`, {
      method: "POST",
    }),

  setupWebhook: (id: number) =>
    request<{ message: string; webhook_url?: string }>(`/api/v1/org/repositories/${id}/webhook`, {
      method: "POST",
    }),
};
