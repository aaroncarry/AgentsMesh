import { test, expect } from "../../fixtures/index";
import { CLEANUP, HASH_PASSWORD123 } from "../../helpers/test-data";
import { clearAuthRateLimit } from "../../helpers/redis";

test.describe("Password Reset Flow", () => {
  test.beforeEach(async () => {
    clearAuthRateLimit();
  });
  /**
   * TC-PWRST-001: Request password reset
   * Maps to: e2e/account/auth/password-reset/TC-PWRST-001-forgot-password.yaml
   */
  test("forgot password returns success for valid email", async ({ api }) => {
    const res = await api.postPublic("/api/v1/auth/forgot-password", {
      email: "dev@agentsmesh.local",
    });
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data.message).toBeTruthy();
  });

  test("forgot password returns success for non-existent email (no enumeration)", async ({ api }) => {
    const res = await api.postPublic("/api/v1/auth/forgot-password", {
      email: "nonexistent@test.local",
    });
    // Security: same response for non-existent emails
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data.message).toBeTruthy();
  });

  /**
   * TC-PWRST-002: Reset password with valid token
   * Maps to: e2e/account/auth/password-reset/TC-PWRST-002-reset-success.yaml
   */
  test("reset password with valid token succeeds", async ({ api, db }) => {
    const email = "pwreset-e2e@test.local";

    // Setup: create user with reset token
    db.setup(`
      INSERT INTO users (email, username, password_hash, name, is_email_verified, password_reset_token, password_reset_expires_at)
      VALUES ('${email}', 'pwresete2e', '${HASH_PASSWORD123}', 'PW Reset User', true, 'test-reset-token-12345', NOW() + INTERVAL '1 hour')
      ON CONFLICT (email) DO UPDATE SET
        password_reset_token = 'test-reset-token-12345',
        password_reset_expires_at = NOW() + INTERVAL '1 hour'
    `);

    const res = await api.postPublic("/api/v1/auth/reset-password", {
      token: "test-reset-token-12345",
      new_password: "NewTestPass456!",
    });
    expect(res.status).toBe(200);

    // Verify login with new password works
    const loginRes = await api.postPublic("/api/v1/auth/login", {
      email,
      password: "NewTestPass456!",
    });
    expect(loginRes.status).toBe(200);

    db.cleanup(CLEANUP.userByEmail(email));
  });

  /**
   * TC-PWRST-003: Reset password with invalid token
   * Maps to: e2e/account/auth/password-reset/TC-PWRST-003-invalid-token.yaml
   */
  test("reset password with invalid token fails", async ({ api }) => {
    const res = await api.postPublic("/api/v1/auth/reset-password", {
      token: "invalid-reset-token-xyz",
      new_password: "NewTestPass456!",
    });
    expect(res.status).toBe(400);
  });

  test("reset password with empty token fails", async ({ api }) => {
    const res = await api.postPublic("/api/v1/auth/reset-password", {
      token: "",
      new_password: "NewTestPass456!",
    });
    expect(res.status).toBe(400);
  });

  test("reset password with weak new password fails", async ({ api }) => {
    const res = await api.postPublic("/api/v1/auth/reset-password", {
      token: "some-token",
      new_password: "123",
    });
    expect(res.status).toBe(400);
  });
});
