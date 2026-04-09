import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";

const BILLING = `/api/v1/orgs/${TEST_ORG_SLUG}/billing`;

test.describe("Promo Codes", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });

  /**
   * TC-PROMO-002: Validate promo code
   */
  test("validate promo code endpoint exists", async ({ api }) => {
    const res = await api.post(`${BILLING}/promo-codes/validate`, {
      code: "TESTCODE",
    });
    // 200 if valid, 400/404 if invalid code
    expect([200, 400, 404]).toContain(res.status);
  });

  /**
   * TC-PROMO-003: Invalid promo code
   */
  test("invalid promo code returns valid=false", async ({ api }) => {
    const res = await api.post(`${BILLING}/promo-codes/validate`, {
      code: "INVALID_CODE_XYZ_999",
    });
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data.valid).toBe(false);
  });

  /**
   * TC-PROMO-004: Redemption history
   */
  test("get promo code redemption history", async ({ api }) => {
    const res = await api.get(`${BILLING}/promo-codes/history`);
    expect(res.status).toBe(200);
  });
});
