import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";

const BILLING = `/api/v1/orgs/${TEST_ORG_SLUG}/billing`;

test.describe("Billing Cycle", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });

  /**
   * TC-CYCLE-001: Display current billing cycle
   */
  test("billing overview shows cycle info", async ({ api }) => {
    const res = await api.get(`${BILLING}/overview`);
    expect(res.status).toBe(200);
  });

  /**
   * TC-CYCLE-002/003: Change billing cycle
   */
  test("change billing cycle returns appropriate status", async ({ api }) => {
    const res = await api.post(`${BILLING}/subscription/change-cycle`, {
      billing_cycle: "yearly",
    });
    // 200 if changed, 400 if same cycle or not allowed
    expect([200, 400]).toContain(res.status);
  });
});
