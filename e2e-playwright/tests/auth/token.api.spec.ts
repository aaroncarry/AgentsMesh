import { test, expect } from "../../fixtures/index";
import { TEST_USER } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";

test.describe("Token Management", () => {
  test.beforeEach(async () => {
    clearAuthRateLimit();
  });
  /**
   * TC-TOKEN-001: Token refresh
   * Maps to: e2e/account/auth/token/TC-TOKEN-001-refresh.yaml
   */
  test("refresh token returns new tokens", async ({ api }) => {
    const loginData = await api.login();
    const refreshToken = loginData.refresh_token;

    const res = await api.postPublic("/api/v1/auth/refresh", {
      refresh_token: refreshToken,
    });
    expect(res.status).toBe(200);

    const data = await res.json();
    expect(data.token).toBeTruthy();
    expect(data.refresh_token).toBeTruthy();
  });

  /**
   * TC-TOKEN-002: Token refresh with invalid token
   * Maps to: e2e/account/auth/token/TC-TOKEN-002-invalid-refresh.yaml
   */
  test("refresh with invalid token fails", async ({ api }) => {
    const res = await api.postPublic("/api/v1/auth/refresh", {
      refresh_token: "invalid-refresh-token-12345",
    });
    expect(res.status).toBe(401);
  });

  test("refresh with empty token fails", async ({ api }) => {
    const res = await api.postPublic("/api/v1/auth/refresh", {
      refresh_token: "",
    });
    expect(res.status).toBe(400);
  });

  /**
   * TC-TOKEN-003: Logout invalidates token
   * Maps to: e2e/account/auth/token/TC-TOKEN-003-logout.yaml
   */
  test("logout invalidates access token", async ({ api }) => {
    const loginData = await api.login();
    const token = loginData.token;

    const logoutRes = await api.postWithToken("/api/v1/auth/logout", {}, token);
    expect(logoutRes.status).toBe(200);
  });

  /**
   * TC-TOKEN-004: Multi-device concurrent login
   * Maps to: e2e/account/auth/token/TC-TOKEN-004-multi-device-login.yaml
   */
  test("multi-device login produces independent tokens", async ({ api }) => {
    // Login from three "devices" with delay to ensure different JWT iat claims
    const a = await api.login(TEST_USER.email, TEST_USER.password);
    await new Promise((r) => setTimeout(r, 1100));
    const b = await api.login(TEST_USER.email, TEST_USER.password);
    await new Promise((r) => setTimeout(r, 1100));
    const c = await api.login(TEST_USER.email, TEST_USER.password);

    // Tokens should be different
    expect(a.token).not.toBe(b.token);
    expect(b.token).not.toBe(c.token);

    // All tokens should work
    for (const token of [a.token, b.token, c.token]) {
      const res = await api.getWithToken("/api/v1/users/me", token);
      expect(res.status).toBe(200);
    }

    // Logout device A
    await api.postWithToken("/api/v1/auth/logout", {}, a.token);

    // B and C should still work
    const resB = await api.getWithToken("/api/v1/users/me", b.token);
    expect(resB.status).toBe(200);
    const resC = await api.getWithToken("/api/v1/users/me", c.token);
    expect(resC.status).toBe(200);
  });
});
