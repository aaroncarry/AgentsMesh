import { request } from "./base";

// Credential type constants
export const CredentialType = {
  RUNNER_LOCAL: "runner_local",
  OAUTH: "oauth",
  PAT: "pat",
  SSH_KEY: "ssh_key",
} as const;

export type CredentialTypeValue =
  (typeof CredentialType)[keyof typeof CredentialType];

// Git Credential types
export interface GitCredentialData {
  id: number;
  user_id: number;
  name: string;
  credential_type: CredentialTypeValue;
  repository_provider_id?: number;
  repository_provider?: {
    id: number;
    name: string;
    provider_type: string;
    base_url: string;
  };
  public_key?: string;
  fingerprint?: string;
  host_pattern?: string;
  is_default: boolean;
  created_at: string;
  updated_at: string;
}

// Runner local credential (virtual type, shown in UI but not stored in DB)
export interface RunnerLocalCredentialData {
  id: string; // "runner_local"
  name: string;
  credential_type: "runner_local";
  is_default: boolean;
}

export interface CreateGitCredentialRequest {
  name: string;
  credential_type: CredentialTypeValue;
  repository_provider_id?: number; // Required for oauth type
  pat?: string; // For pat type
  private_key?: string; // For ssh_key type
  host_pattern?: string;
}

export interface UpdateGitCredentialRequest {
  name?: string;
  pat?: string;
  private_key?: string;
  host_pattern?: string;
}

export interface SetDefaultRequest {
  credential_id?: number | null; // null means runner_local
}

// User Git Credential API
export const userGitCredentialApi = {
  // List all Git credentials for the current user
  list: () =>
    request<{
      credentials: GitCredentialData[];
      runner_local: RunnerLocalCredentialData;
    }>("/api/v1/users/git-credentials"),

  // Create a new Git credential
  create: (data: CreateGitCredentialRequest) =>
    request<{ credential: GitCredentialData }>("/api/v1/users/git-credentials", {
      method: "POST",
      body: data,
    }),

  // Get a single Git credential
  get: (id: number) =>
    request<{ credential: GitCredentialData }>(
      `/api/v1/users/git-credentials/${id}`
    ),

  // Update a Git credential
  update: (id: number, data: UpdateGitCredentialRequest) =>
    request<{ credential: GitCredentialData }>(
      `/api/v1/users/git-credentials/${id}`,
      {
        method: "PUT",
        body: data,
      }
    ),

  // Delete a Git credential
  delete: (id: number) =>
    request<{ message: string }>(`/api/v1/users/git-credentials/${id}`, {
      method: "DELETE",
    }),

  // Get the default Git credential
  getDefault: () =>
    request<{
      credential: GitCredentialData | RunnerLocalCredentialData;
      is_runner_local: boolean;
    }>("/api/v1/users/git-credentials/default"),

  // Set the default Git credential
  setDefault: (data: SetDefaultRequest) =>
    request<{ message: string; is_runner_local: boolean }>(
      "/api/v1/users/git-credentials/default",
      {
        method: "POST",
        body: data,
      }
    ),

  // Clear the default Git credential (falls back to runner_local)
  clearDefault: () =>
    request<{ message: string }>("/api/v1/users/git-credentials/default", {
      method: "DELETE",
    }),
};

// Helper function to get display name for credential type
export function getCredentialTypeLabel(
  type: CredentialTypeValue
): string {
  switch (type) {
    case CredentialType.RUNNER_LOCAL:
      return "Runner Local";
    case CredentialType.OAUTH:
      return "OAuth";
    case CredentialType.PAT:
      return "Personal Access Token";
    case CredentialType.SSH_KEY:
      return "SSH Key";
    default:
      return type;
  }
}

// Helper function to check if a credential is the virtual runner_local type
export function isRunnerLocalCredential(
  credential: GitCredentialData | RunnerLocalCredentialData
): credential is RunnerLocalCredentialData {
  return credential.credential_type === CredentialType.RUNNER_LOCAL;
}
