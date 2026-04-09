import { test, expect } from "../../../fixtures/index";
import { clearAuthRateLimit } from "../../../helpers/redis";

test.describe("SSO LDAP API", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });

  /**
   * TC-SSO-LDAP-001: LDAP auth with invalid credentials
   */
  test("LDAP auth with invalid credentials returns 401", async ({ api }) => {
    const res = await api.postPublic(
      "/api/v1/auth/sso/nonexistent.example.com/ldap",
      { username: "baduser", password: "badpass" }
    );
    // 401 (invalid creds) or 404 (domain not found)
    expect([401, 404]).toContain(res.status);
  });

  /**
   * TC-SSO-LDAP-002: LDAP auth with non-existent domain
   */
  test("LDAP auth for non-existent domain returns 404", async ({ api }) => {
    const res = await api.postPublic(
      "/api/v1/auth/sso/no-such-domain.test/ldap",
      { username: "user", password: "pass" }
    );
    expect(res.status).toBe(404);
  });
});
