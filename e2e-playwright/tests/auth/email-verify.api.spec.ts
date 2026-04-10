import { test, expect } from "../../fixtures/index";
import { CLEANUP, HASH_PASSWORD123 } from "../../helpers/test-data";
import { clearAuthRateLimit } from "../../helpers/redis";

test.describe("Email Verification", () => {
  test.beforeEach(async () => {
    clearAuthRateLimit();
  });
  /**
   * TC-VERIFY-001: Verify email with valid token
   * Maps to: e2e/account/auth/email-verify/TC-VERIFY-001-success.yaml
   */
  test("verify email with valid token succeeds", async ({ api, db }) => {
    const email = "verify-e2e@test.local";

    db.setup(`
      INSERT INTO users (email, username, password_hash, name, is_email_verified,
        email_verification_token, email_verification_expires_at)
      VALUES ('${email}', 'verifye2e', '${HASH_PASSWORD123}', 'Verify User', false,
        'test-verify-token-12345', NOW() + INTERVAL '1 hour')
      ON CONFLICT (email) DO UPDATE SET
        is_email_verified = false,
        email_verification_token = 'test-verify-token-12345',
        email_verification_expires_at = NOW() + INTERVAL '1 hour'
    `);

    const res = await api.postPublic("/api/v1/auth/verify-email", {
      token: "test-verify-token-12345",
    });
    expect(res.status).toBe(200);

    // DB verification
    const verified = db.queryValue(
      `SELECT is_email_verified::text FROM users WHERE email = '${email}'`
    );
    expect(verified).toBe("true");

    db.cleanup(CLEANUP.userByEmail(email));
  });

  /**
   * TC-VERIFY-002: Verify email with invalid token
   * Maps to: e2e/account/auth/email-verify/TC-VERIFY-002-invalid-token.yaml
   */
  test("verify email with invalid token fails", async ({ api }) => {
    const res = await api.postPublic("/api/v1/auth/verify-email", {
      token: "invalid-verification-token-xyz",
    });
    expect(res.status).toBe(400);
  });

  test("verify email with empty token fails", async ({ api }) => {
    const res = await api.postPublic("/api/v1/auth/verify-email", {
      token: "",
    });
    expect(res.status).toBe(400);
  });

  /**
   * TC-VERIFY-003: Resend verification email
   * Maps to: e2e/account/auth/email-verify/TC-VERIFY-003-resend.yaml
   */
  test("resend verification for valid email succeeds", async ({ api, db }) => {
    const email = "resend-e2e@test.local";

    db.setup(`
      INSERT INTO users (email, username, password_hash, name, is_email_verified)
      VALUES ('${email}', 'resende2e', '${HASH_PASSWORD123}', 'Resend User', false)
      ON CONFLICT (email) DO UPDATE SET is_email_verified = false
    `);

    const res = await api.postPublic("/api/v1/auth/resend-verification", {
      email,
    });
    expect(res.status).toBe(200);

    db.cleanup(CLEANUP.userByEmail(email));
  });

  test("resend verification for non-existent email returns 200 (no enumeration)", async ({ api }) => {
    const res = await api.postPublic("/api/v1/auth/resend-verification", {
      email: "nonexistent@test.local",
    });
    expect(res.status).toBe(200);
  });
});
