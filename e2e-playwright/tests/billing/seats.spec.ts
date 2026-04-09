import { test, expect } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";
import { clearAuthRateLimit } from "../../helpers/redis";

const BILLING = `/api/v1/orgs/${TEST_ORG_SLUG}/billing`;

test.describe("Billing Seats", () => {
  test.beforeEach(async () => { clearAuthRateLimit(); });

  /**
   * TC-SEAT-001: Get seat info
   */
  test("get seat information", async ({ api }) => {
    const res = await api.get(`${BILLING}/seats`);
    expect(res.status).toBe(200);
  });

  /**
   * TC-SEAT-004: Purchase seats (may fail due to payment requirement)
   */
  test("purchase seats returns appropriate status", async ({ api }) => {
    const res = await api.post(`${BILLING}/seats/purchase`, {
      quantity: 1,
    });
    // 200 if seats added, 400/402 if payment needed or limit
    expect([200, 400, 402]).toContain(res.status);
  });
});
