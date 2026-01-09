import { request } from "./base";

// Git Connection types
export interface GitConnectionData {
  id: string; // Format: "connection:123" or "oauth:github"
  type: "oauth" | "personal";
  provider_type: string; // github, gitlab, gitee
  provider_name: string; // Display name
  base_url: string; // https://github.com
  username: string; // Username on the platform
  avatar_url?: string;
  auth_type?: string; // pat, ssh (only for personal)
  is_active: boolean;
  created_at: string;
}

export interface RemoteRepositoryData {
  id: string;
  name: string;
  full_path: string;
  description: string;
  default_branch: string;
  visibility: string;
  clone_url: string;
  ssh_clone_url: string;
  web_url: string;
}

export interface CreateConnectionRequest {
  provider_type: string;
  provider_name: string;
  base_url: string;
  auth_type?: "pat" | "ssh";
  access_token?: string;
  ssh_private_key?: string;
}

export interface UpdateConnectionRequest {
  provider_name?: string;
  access_token?: string;
  ssh_private_key?: string;
  is_active?: boolean;
}

// Git Connection API
export const gitConnectionApi = {
  // List all Git connections (OAuth + Personal)
  list: () =>
    request<{ connections: GitConnectionData[] }>("/api/v1/user/git-connections"),

  // Create a new Git connection (PAT/SSH)
  create: (data: CreateConnectionRequest) =>
    request<{ connection: GitConnectionData }>("/api/v1/user/git-connections", {
      method: "POST",
      body: data,
    }),

  // Get a single Git connection
  get: (id: string) =>
    request<{ connection: GitConnectionData }>(`/api/v1/user/git-connections/${id}`),

  // Update a Git connection
  update: (id: string, data: UpdateConnectionRequest) =>
    request<{ connection: GitConnectionData }>(`/api/v1/user/git-connections/${id}`, {
      method: "PUT",
      body: data,
    }),

  // Delete a Git connection
  delete: (id: string) =>
    request<{ message: string }>(`/api/v1/user/git-connections/${id}`, {
      method: "DELETE",
    }),

  // List repositories accessible through a connection
  listRepositories: (
    connectionId: string,
    options?: { page?: number; perPage?: number; search?: string }
  ) => {
    const params = new URLSearchParams();
    if (options?.page) params.append("page", String(options.page));
    if (options?.perPage) params.append("per_page", String(options.perPage));
    if (options?.search) params.append("search", options.search);
    const query = params.toString();
    return request<{
      repositories: RemoteRepositoryData[];
      page: number;
      per_page: number;
    }>(`/api/v1/user/git-connections/${connectionId}/repositories${query ? `?${query}` : ""}`);
  },
};
