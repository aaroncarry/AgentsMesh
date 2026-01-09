import { request } from "./base";

// User API
export const userApi = {
  getMe: () =>
    request<{ user: { id: number; email: string; username: string; name?: string } }>(
      "/api/v1/users/me"
    ),

  getOrganizations: () =>
    request<{ organizations: Array<{ id: number; name: string; slug: string; role: string }> }>(
      "/api/v1/users/me/organizations"
    ),
};
