import { request } from "./base";

// Binding types
export interface SessionBinding {
  id: number;
  organization_id: number;
  initiator_session: string;
  target_session: string;
  granted_scopes: string[];
  pending_scopes: string[];
  status: "pending" | "active" | "rejected" | "inactive" | "expired";
  requested_at?: string;
  responded_at?: string;
  expires_at?: string;
  rejection_reason?: string;
  created_at: string;
  updated_at: string;
}

// Binding API
export const bindingApi = {
  // Request a new binding
  requestBinding: (targetSession: string, scopes: string[], policy?: string, sessionKey?: string) =>
    request<{ binding: SessionBinding }>("/api/v1/org/bindings", {
      method: "POST",
      body: { target_session: targetSession, scopes, policy },
      headers: sessionKey ? { "X-Session-Key": sessionKey } : undefined,
    }),

  // Accept a pending binding
  acceptBinding: (bindingId: number, sessionKey?: string) =>
    request<{ binding: SessionBinding }>("/api/v1/org/bindings/accept", {
      method: "POST",
      body: { binding_id: bindingId },
      headers: sessionKey ? { "X-Session-Key": sessionKey } : undefined,
    }),

  // Reject a pending binding
  rejectBinding: (bindingId: number, reason?: string, sessionKey?: string) =>
    request<{ binding: SessionBinding }>("/api/v1/org/bindings/reject", {
      method: "POST",
      body: { binding_id: bindingId, reason },
      headers: sessionKey ? { "X-Session-Key": sessionKey } : undefined,
    }),

  // Request additional scopes
  requestScopes: (bindingId: number, scopes: string[], sessionKey?: string) =>
    request<{ binding: SessionBinding }>(`/api/v1/org/bindings/${bindingId}/scopes`, {
      method: "POST",
      body: { scopes },
      headers: sessionKey ? { "X-Session-Key": sessionKey } : undefined,
    }),

  // Approve pending scopes
  approveScopes: (bindingId: number, scopes: string[], sessionKey?: string) =>
    request<{ binding: SessionBinding }>(`/api/v1/org/bindings/${bindingId}/scopes/approve`, {
      method: "POST",
      body: { scopes },
      headers: sessionKey ? { "X-Session-Key": sessionKey } : undefined,
    }),

  // Unbind from a session
  unbind: (targetSession: string, sessionKey?: string) =>
    request<void>("/api/v1/org/bindings/unbind", {
      method: "POST",
      body: { target_session: targetSession },
      headers: sessionKey ? { "X-Session-Key": sessionKey } : undefined,
    }),

  // List all bindings
  listBindings: (status?: string, sessionKey?: string) => {
    const params = status ? `?status=${status}` : "";
    return request<{ bindings: SessionBinding[]; total: number }>(`/api/v1/org/bindings${params}`, {
      headers: sessionKey ? { "X-Session-Key": sessionKey } : undefined,
    });
  },

  // Get pending binding requests
  getPendingBindings: (sessionKey?: string) =>
    request<{ pending: SessionBinding[]; count: number }>("/api/v1/org/bindings/pending", {
      headers: sessionKey ? { "X-Session-Key": sessionKey } : undefined,
    }),

  // Get bound sessions
  getBoundSessions: (sessionKey?: string) =>
    request<{ sessions: string[]; count: number }>("/api/v1/org/bindings/sessions", {
      headers: sessionKey ? { "X-Session-Key": sessionKey } : undefined,
    }),

  // Check binding status with a specific session
  checkBinding: (targetSession: string, sessionKey?: string) =>
    request<{ is_bound: boolean; binding?: SessionBinding }>(`/api/v1/org/bindings/check/${targetSession}`, {
      headers: sessionKey ? { "X-Session-Key": sessionKey } : undefined,
    }),
};
