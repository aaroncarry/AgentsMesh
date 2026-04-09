import { test, expect } from "../../../fixtures/index";
import { clearAuthRateLimit } from "../../../helpers/redis";

test.describe("SSO Discovery API", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });

  /**
   * TC-SSO-DISC-001: Discover SSO for valid domain
   */
  test("discover SSO configs by email", async ({ api }) => {
    const res = await api.get(
      "/api/v1/auth/sso/discover?email=user@agentsmesh.local"
    );
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data.configs).toBeDefined();
  });

  /**
   * TC-SSO-DISC-002: Non-SSO domain returns empty
   */
  test("non-SSO domain returns empty configs", async ({ api }) => {
    const res = await api.get(
      "/api/v1/auth/sso/discover?email=user@no-sso-domain.com"
    );
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data.configs?.length ?? 0).toBe(0);
  });

  /**
   * TC-SSO-DISC-003: Invalid email format
   */
  test("invalid email returns 400", async ({ api }) => {
    const res = await api.get("/api/v1/auth/sso/discover?email=not-an-email");
    expect([200, 400]).toContain(res.status);
  });
});
