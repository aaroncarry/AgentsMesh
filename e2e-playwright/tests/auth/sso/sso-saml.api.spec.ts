import { test, expect } from "../../../fixtures/index";
import { clearAuthRateLimit } from "../../../helpers/redis";

test.describe("SSO SAML API", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });

  /**
   * TC-SSO-SAML-002: SAML metadata for non-existent domain
   */
  test("SAML metadata for non-existent domain returns 404", async ({ api }) => {
    const res = await api.get(
      "/api/v1/auth/sso/nonexistent.example.com/saml/metadata"
    );
    expect(res.status).toBe(404);
  });
});
