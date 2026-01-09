import { request } from "./base";

// SSH Key types
export interface SSHKeyData {
  id: number;
  organization_id: number;
  name: string;
  public_key: string;
  fingerprint: string;
  created_at: string;
  updated_at: string;
}

// SSH Key API
export const sshKeyApi = {
  list: () =>
    request<{ ssh_keys: SSHKeyData[] }>("/api/v1/org/ssh-keys"),

  get: (id: number) =>
    request<{ ssh_key: SSHKeyData }>(`/api/v1/org/ssh-keys/${id}`),

  create: (data: {
    name: string;
    private_key?: string; // Optional: if nil, generate a new key pair
  }) =>
    request<{ ssh_key: SSHKeyData }>("/api/v1/org/ssh-keys", {
      method: "POST",
      body: data,
    }),

  update: (id: number, name: string) =>
    request<{ ssh_key: SSHKeyData }>(`/api/v1/org/ssh-keys/${id}`, {
      method: "PUT",
      body: { name },
    }),

  delete: (id: number) =>
    request<{ message: string }>(`/api/v1/org/ssh-keys/${id}`, {
      method: "DELETE",
    }),
};
