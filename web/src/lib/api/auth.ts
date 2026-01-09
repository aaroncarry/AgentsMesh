import { request } from "./base";

// Auth API
export const authApi = {
  login: (email: string, password: string) =>
    request<{
      token: string;
      refresh_token: string;
      expires_in: number;
      user: { id: number; email: string; username: string; name?: string };
    }>(
      "/api/v1/auth/login",
      {
        method: "POST",
        body: { email, password },
        skipAuthRefresh: true, // Don't try to refresh on login failure
      }
    ),

  register: (data: { email: string; username: string; password: string; name?: string }) =>
    request<{
      token: string;
      refresh_token: string;
      expires_in: number;
      user: { id: number; email: string; username: string; name?: string };
    }>(
      "/api/v1/auth/register",
      {
        method: "POST",
        body: data,
      }
    ),

  logout: () => request("/api/v1/auth/logout", { method: "POST" }),

  // Email verification
  verifyEmail: (token: string) =>
    request<{
      message: string;
      token: string;
      refresh_token: string;
      expires_in: number;
      user: { id: number; email: string; username: string; name?: string; is_email_verified: boolean };
    }>("/api/v1/auth/verify-email", {
      method: "POST",
      body: { token },
    }),

  resendVerification: (email: string) =>
    request<{ message: string }>("/api/v1/auth/resend-verification", {
      method: "POST",
      body: { email },
    }),

  // Password reset
  forgotPassword: (email: string) =>
    request<{ message: string }>("/api/v1/auth/forgot-password", {
      method: "POST",
      body: { email },
    }),

  resetPassword: (token: string, newPassword: string) =>
    request<{ message: string }>("/api/v1/auth/reset-password", {
      method: "POST",
      body: { token, new_password: newPassword },
    }),

  // Token refresh
  refreshToken: (refreshToken: string) =>
    request<{
      token: string;
      refresh_token: string;
      expires_in: number;
    }>("/api/v1/auth/refresh", {
      method: "POST",
      body: { refresh_token: refreshToken },
    }),
};
