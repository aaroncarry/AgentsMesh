import { request, orgPath } from "./base";

export interface APIKeyData {
  id: number;
  organization_id: number;
  name: string;
  description?: string;
  key_prefix: string;
  scopes: string[];
  is_enabled: boolean;
  expires_at?: string;
  last_used_at?: string;
  created_by: number;
  created_at: string;
  updated_at: string;
}

export interface CreateAPIKeyRequest {
  name: string;
  description?: string;
  scopes: string[];
  expires_in?: number; // seconds, null = never
}

export interface UpdateAPIKeyRequest {
  name?: string;
  description?: string;
  scopes?: string[];
  is_enabled?: boolean;
}

export const apiKeyApi = {
  list: () =>
    request<{ api_keys: APIKeyData[]; total: number }>(orgPath("/api-keys")),

  get: (id: number) =>
    request<{ api_key: APIKeyData }>(`${orgPath("/api-keys")}/${id}`),

  create: (data: CreateAPIKeyRequest) =>
    request<{ api_key: APIKeyData; raw_key: string }>(orgPath("/api-keys"), {
      method: "POST",
      body: data,
    }),

  update: (id: number, data: UpdateAPIKeyRequest) =>
    request<{ api_key: APIKeyData }>(`${orgPath("/api-keys")}/${id}`, {
      method: "PUT",
      body: data,
    }),

  delete: (id: number) =>
    request<{ message: string }>(`${orgPath("/api-keys")}/${id}`, {
      method: "DELETE",
    }),

  revoke: (id: number) =>
    request<{ message: string }>(`${orgPath("/api-keys")}/${id}/revoke`, {
      method: "POST",
    }),
};
